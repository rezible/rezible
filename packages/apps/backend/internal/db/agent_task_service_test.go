package db

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/agents"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunfinding"
	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
)

type AgentTaskServiceSuite struct {
	testkit.Suite
}

func TestAgentTaskServiceSuite(t *testing.T) {
	suite.Run(t, &AgentTaskServiceSuite{Suite: testkit.NewSuite()})
}

func (s *AgentTaskServiceSuite) newAgentTaskService(jobSvc *mocks.MockJobService, msgSvc *mocks.MockMessageService) *AgentTaskService {
	return s.newAgentTaskServiceWithRegistry(jobSvc, msgSvc, newTestWorkflowRegistry())
}

func (s *AgentTaskServiceSuite) newAgentTaskServiceWithRegistry(jobSvc *mocks.MockJobService, msgSvc *mocks.MockMessageService, reg *agents.WorkflowRegistry) *AgentTaskService {
	if msgSvc == nil {
		msgSvc = mocks.NewMockMessageService(s.T())
		msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	}
	return &AgentTaskService{
		logger: slog.Default(),
		db:     s.Database(),
		jobs:   jobSvc,
		msgs:   msgSvc,
		reg:    reg,
	}
}

func alertWorkflowInput(alertID uuid.UUID) map[string]any {
	input, err := agents.EncodeInput(agents.AlertInvestigationInput{
		Subjects: []agents.SubjectRef{{
			Type: "alert",
			ID:   alertID,
		}},
	})
	if err != nil {
		panic(err)
	}
	return input
}

func newTestWorkflowRegistry() *agents.WorkflowRegistry {
	reg := agents.NewWorkflowRegistry()
	reg.Register(agents.WorkflowKindAlertInvestigation, testAgentWorkflow)
	return reg
}

func testAgentWorkflow(context.Context, *ent.AgentTask, *ent.AgentRun) (*agents.RunResult, error) {
	return &agents.RunResult{
		Content: "done",
		Data:    map[string]any{"source": "test"},
		Citations: []agents.RunCitationInput{{
			CitationKind:     "primary_subject",
			DomainEntityType: "alert",
			DomainEntityID:   uuid.New(),
			Summary:          "Alert",
		}},
		Findings: []agents.RunFindingInput{{
			FindingKind: "observation",
			Content:     "Observed alert context.",
			Citations: []agents.RunFindingCitationInput{{
				CitationIndex: 1,
				SupportKind:   "source",
			}},
		}},
	}, nil
}

func (s *AgentTaskServiceSuite) createAgentOwner(ctx context.Context) uuid.UUID {
	user, err := s.Client(ctx).User.Create().
		SetEmail("agent-owner+" + uuid.NewString() + "@example.com").
		SetName("Agent Owner").
		Save(ctx)
	s.Require().NoError(err)
	return user.ID
}

func (s *AgentTaskServiceSuite) TestCreateTaskPersistsTaskAndEnqueuesRunRequest() {
	ctx := s.SeedTenantContext()
	ownerID := s.createAgentOwner(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	jobSvc.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("jobs.RequestAgentTaskRun"), mock.Anything).
		Return(nil, nil).
		Once()

	svc := s.newAgentTaskService(jobSvc, msgSvc)
	task, err := svc.CreateTask(ctx, rez.CreateAgentTaskRequest{
		OwnerUserID:    ownerID,
		WorkflowKind:   agents.WorkflowKindAlertInvestigation,
		WorkflowInput:  alertWorkflowInput(uuid.New()),
		TriggerKind:    "manual",
		TriggerPayload: map[string]any{"source": "test"},
	})

	s.Require().NoError(err)
	s.Equal(agents.WorkflowKindAlertInvestigation.String(), task.WorkflowKind)
	runs, runErr := s.Client(ctx).AgentRun.Query().
		Where(agentrun.AgentTaskID(task.ID)).
		All(ctx)
	s.Require().NoError(runErr)
	s.Empty(runs)
}

func (s *AgentTaskServiceSuite) TestCreateRequestedRunAllocatesAttemptAndEnqueuesWorkflow() {
	ctx := s.SeedTenantContext()
	ownerID := s.createAgentOwner(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	jobSvc.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("jobs.RunAgentWorkflow"), mock.Anything).
		Return(nil, nil).
		Once()
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.MatchedBy(func(event any) bool {
			_, ok := event.(rez.AgentRunQueuedEvent)
			return ok
		})).
		Return(nil).
		Once()

	task, taskErr := s.Client(ctx).AgentTask.Create().
		SetOwnerUserID(ownerID).
		SetWorkflowKind(agents.WorkflowKindAlertInvestigation.String()).
		SetWorkflowInput(alertWorkflowInput(uuid.New())).
		SetTriggerKind("manual").
		SetTriggerPayload(map[string]any{}).
		Save(ctx)
	s.Require().NoError(taskErr)

	svc := s.newAgentTaskService(jobSvc, msgSvc)
	run, err := svc.CreateRequestedRun(ctx, task.ID)

	s.Require().NoError(err)
	s.Equal(1, run.Attempt)
	s.Equal(agentrun.StatusQueued, run.Status)
}

func (s *AgentTaskServiceSuite) TestRunWorkflowUsesRegistryRunnerAndPersistsResult() {
	ctx := s.SeedTenantContext()
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	task, run := s.createQueuedAgentRun(ctx)
	svc := s.newAgentTaskServiceWithRegistry(nil, msgSvc, newTestWorkflowRegistry())

	err := svc.RunWorkflow(ctx, run.ID)

	s.Require().NoError(err)
	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusSucceeded, updated.Status)
	result, resultErr := svc.GetRunResult(ctx, run.ID)
	s.Require().NoError(resultErr)
	s.Equal("done", result.Content)
	s.Equal(task.ID, updated.AgentTaskID)
}

func (s *AgentTaskServiceSuite) TestUpdateRunPersistsResultFindingsAndCitations() {
	ctx := s.SeedTenantContext()
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	task, run := s.createQueuedAgentRun(ctx)

	svc := s.newAgentTaskService(nil, msgSvc)
	_, err := svc.UpdateRun(ctx, task, run, UpdateAgentRunParams{
		Status: agentrun.StatusSucceeded,
		Result: &agents.RunResult{
			Content: "done",
			Data:    map[string]any{"source": "test"},
			Citations: []agents.RunCitationInput{{
				CitationKind:     "primary_subject",
				DomainEntityType: "alert",
				DomainEntityID:   uuid.New(),
				Summary:          "Alert",
			}},
			Findings: []agents.RunFindingInput{{
				FindingKind: "observation",
				Content:     "Observed alert context.",
				Citations: []agents.RunFindingCitationInput{{
					CitationIndex: 1,
					SupportKind:   "source",
				}},
			}},
		},
	})

	s.Require().NoError(err)
	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusSucceeded, updated.Status)
	result, resultErr := svc.GetRunResult(ctx, run.ID)
	s.Require().NoError(resultErr)
	s.Equal("done", result.Content)
	findings, findingErr := s.Client(ctx).AgentRunFinding.Query().
		Where(agentrunfinding.AgentRunID(run.ID)).
		All(ctx)
	s.Require().NoError(findingErr)
	s.Len(findings, 1)
}

func (s *AgentTaskServiceSuite) TestUpdateRunRejectsBlankResultContent() {
	ctx := s.SeedTenantContext()
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	task, run := s.createQueuedAgentRun(ctx)

	svc := s.newAgentTaskService(nil, msgSvc)
	_, err := svc.UpdateRun(ctx, task, run, UpdateAgentRunParams{
		Status: agentrun.StatusSucceeded,
		Result: &agents.RunResult{},
	})

	s.Require().Error(err)
	s.ErrorContains(err, "result content is required")
	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusQueued, updated.Status)
}

func (s *AgentTaskServiceSuite) TestUpdateRunRejectsCitationWithoutSummary() {
	ctx := s.SeedTenantContext()
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	task, run := s.createQueuedAgentRun(ctx)

	svc := s.newAgentTaskService(nil, msgSvc)
	_, err := svc.UpdateRun(ctx, task, run, UpdateAgentRunParams{
		Status: agentrun.StatusSucceeded,
		Result: &agents.RunResult{
			Content: "done",
			Citations: []agents.RunCitationInput{{
				CitationKind:     "primary_subject",
				DomainEntityType: "alert",
				DomainEntityID:   uuid.New(),
			}},
		},
	})

	s.Require().Error(err)
	s.ErrorContains(err, "summary")
	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusQueued, updated.Status)
}

func (s *AgentTaskServiceSuite) TestUpdateRunRejectsFindingWithInvalidCitationIndex() {
	ctx := s.SeedTenantContext()
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	task, run := s.createQueuedAgentRun(ctx)

	svc := s.newAgentTaskService(nil, msgSvc)
	_, err := svc.UpdateRun(ctx, task, run, UpdateAgentRunParams{
		Status: agentrun.StatusSucceeded,
		Result: &agents.RunResult{
			Content: "done",
			Citations: []agents.RunCitationInput{{
				CitationKind:     "primary_subject",
				DomainEntityType: "alert",
				DomainEntityID:   uuid.New(),
				Summary:          "Alert",
			}},
			Findings: []agents.RunFindingInput{{
				FindingKind: "observation",
				Content:     "Observed alert context.",
				Citations: []agents.RunFindingCitationInput{{
					CitationIndex: 2,
					SupportKind:   "source",
				}},
			}},
		},
	})

	s.Require().Error(err)
	s.ErrorContains(err, "out of range")
	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusQueued, updated.Status)
}

func (s *AgentTaskServiceSuite) TestUpdateRunRejectsFindingWithoutKind() {
	ctx := s.SeedTenantContext()
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()
	task, run := s.createQueuedAgentRun(ctx)

	svc := s.newAgentTaskService(nil, msgSvc)
	_, err := svc.UpdateRun(ctx, task, run, UpdateAgentRunParams{
		Status: agentrun.StatusSucceeded,
		Result: &agents.RunResult{
			Content: "done",
			Findings: []agents.RunFindingInput{{
				Content: "Observed alert context.",
			}},
		},
	})

	s.Require().Error(err)
	s.ErrorContains(err, "finding kind")
	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusQueued, updated.Status)
}

func (s *AgentTaskServiceSuite) createQueuedAgentRun(ctx context.Context) (*ent.AgentTask, *ent.AgentRun) {
	ownerID := s.createAgentOwner(ctx)
	task, taskErr := s.Client(ctx).AgentTask.Create().
		SetOwnerUserID(ownerID).
		SetWorkflowKind(agents.WorkflowKindAlertInvestigation.String()).
		SetWorkflowInput(alertWorkflowInput(uuid.New())).
		SetTriggerKind("manual").
		SetTriggerPayload(map[string]any{}).
		Save(ctx)
	s.Require().NoError(taskErr)
	run, runErr := s.Client(ctx).AgentRun.Create().
		SetAgentTaskID(task.ID).
		SetAttempt(1).
		SetStatus(agentrun.StatusQueued).
		Save(ctx)
	s.Require().NoError(runErr)
	return task, run
}
