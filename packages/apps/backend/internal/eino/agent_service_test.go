package eino

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunfinding"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
)

type AgentServiceSuite struct {
	testkit.Suite
}

func TestAgentServiceSuite(t *testing.T) {
	suite.Run(t, &AgentServiceSuite{Suite: testkit.NewSuite()})
}

type testWorkflow struct {
	fail bool
}

func (w testWorkflow) Kind() rez.AgentWorkflowKind {
	return rez.AgentWorkflowKindAlertInvestigation
}

func (w testWorkflow) Validate(context.Context, *ent.AgentTask, *ent.AgentRun) error {
	return nil
}

func (w testWorkflow) Run(_ context.Context, _ *WorkflowRecorder, _ *ent.AgentTask, _ *ent.AgentRun) (*AgentWorkflowResult, error) {
	if w.fail {
		return nil, fmt.Errorf("workflow failed")
	}
	return &AgentWorkflowResult{
		Content: "done",
		Data: map[string]any{
			"schema": "test.v1",
		},
		Citations: []rez.AgentRunCitationInput{{
			CitationKind:     "primary_subject",
			DomainEntityType: "alert",
			DomainEntityID:   uuid.New(),
			Summary:          "Alert",
		}},
		Findings: []rez.AgentRunFindingInput{{
			FindingKind: "observation",
			Content:     "Observed alert context.",
			Citations: []rez.AgentRunFindingCitationInput{{
				CitationIndex: 1,
				SupportKind:   "source",
			}},
		}},
	}, nil
}

func (s *AgentServiceSuite) newService(jobSvc rez.JobService, msgSvc rez.MessageService, workflow agentWorkflow) *AgentService {
	cfg := s.Config()
	cfg.AI.Provider = "gemini"
	cfg.AI.Model = "gemini-2.5-flash"
	cfg.AI.StoreRawModelPayloads = false
	return &AgentService{
		cfg:       cfg.AI,
		logger:    slog.Default(),
		db:        s.Database(),
		jobs:      jobSvc,
		msgs:      msgSvc,
		modelProv: newChatModelProvider(cfg.AI),
		workflows: map[rez.AgentWorkflowKind]agentWorkflow{
			workflow.Kind(): workflow,
		},
	}
}

func (s *AgentServiceSuite) createOwner(ctx context.Context) uuid.UUID {
	user, err := s.Client(ctx).User.Create().
		SetEmail("agent-owner+" + uuid.NewString() + "@example.com").
		SetName("Agent Owner").
		Save(ctx)
	s.Require().NoError(err)
	return user.ID
}

func alertTaskRequest(ownerID uuid.UUID) rez.CreateAgentTaskRequest {
	alertID := uuid.New()
	return rez.CreateAgentTaskRequest{
		OwnerUserID:  ownerID,
		WorkflowKind: rez.AgentWorkflowKindAlertInvestigation,
		WorkflowInput: map[string]any{
			"schema": "alert_investigation.v1",
			"subjects": []map[string]any{{
				"type": "alert",
				"id":   alertID.String(),
			}},
		},
		TriggerKind:    "manual",
		TriggerPayload: map[string]any{"source": "test"},
	}
}

func (s *AgentServiceSuite) TestCreateTaskCreatesTaskAndQueuesFirstRun() {
	ctx := s.SeedTenantContext()
	ownerID := s.createOwner(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())

	jobSvc.EXPECT().
		Insert(mock.MatchedBy(func(ctx context.Context) bool {
			tenantID, ok := execution.GetContext(ctx).TenantID()
			return ok && tenantID == s.SeedTenant.ID
		}), mock.AnythingOfType("jobs.RunAgentWorkflow"), mock.Anything).
		Return(nil, nil).
		Once()
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc := s.newService(jobSvc, msgSvc, testWorkflow{})
	task, err := svc.CreateTask(ctx, alertTaskRequest(ownerID))
	s.Require().NoError(err)
	s.Equal(ownerID, task.OwnerUserID)

	runs, runErr := s.Client(ctx).AgentRun.Query().
		Where(agentrun.AgentTaskID(task.ID)).
		All(ctx)
	s.Require().NoError(runErr)
	s.Require().Len(runs, 1)
	s.Equal(1, runs[0].Attempt)
	s.Equal(agentrun.StatusQueued, runs[0].Status)
}

func (s *AgentServiceSuite) TestRequestRunAllocatesNextAttempt() {
	ctx := s.SeedTenantContext()
	ownerID := s.createOwner(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	jobSvc.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Times(2)
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc := s.newService(jobSvc, msgSvc, testWorkflow{})
	task, taskErr := svc.CreateTask(ctx, alertTaskRequest(ownerID))
	s.Require().NoError(taskErr)

	run, runErr := svc.RequestTaskRun(ctx, task.ID)
	s.Require().NoError(runErr)
	s.Equal(2, run.Attempt)
}

func (s *AgentServiceSuite) TestRunWorkflowPersistsResultFindingsAndCitations() {
	ctx := s.SeedTenantContext()
	ownerID := s.createOwner(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc := s.newService(jobSvc, msgSvc, testWorkflow{})
	task, taskErr := s.Client(ctx).AgentTask.Create().
		SetOwnerUserID(ownerID).
		SetWorkflowKind(string(rez.AgentWorkflowKindAlertInvestigation)).
		SetWorkflowInput(alertTaskRequest(ownerID).WorkflowInput).
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

	err := svc.RunWorkflow(ctx, run.ID)
	s.Require().NoError(err)

	updated, updateErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusSucceeded, updated.Status)
	s.NotNil(updated.StartedAt)
	s.NotNil(updated.FinishedAt)

	result, resultErr := svc.GetRunResult(ctx, run.ID)
	s.Require().NoError(resultErr)
	s.Equal("done", result.Content)

	findings, findingErr := s.Client(ctx).AgentRunFinding.Query().
		Where(agentrunfinding.AgentRunID(run.ID)).
		All(ctx)
	s.Require().NoError(findingErr)
	s.Require().Len(findings, 1)
	s.Equal("observation", findings[0].FindingKind)
}

func (s *AgentServiceSuite) TestRunWorkflowMarksUnknownWorkflowFailed() {
	ctx := s.SeedTenantContext()
	ownerID := s.createOwner(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	task, taskErr := s.Client(ctx).AgentTask.Create().
		SetOwnerUserID(ownerID).
		SetWorkflowKind(string(rez.AgentWorkflowKindAlertInvestigation)).
		SetWorkflowInput(alertTaskRequest(ownerID).WorkflowInput).
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

	svc := s.newService(jobSvc, msgSvc, testWorkflow{})
	delete(svc.workflows, rez.AgentWorkflowKindAlertInvestigation)

	err := svc.RunWorkflow(ctx, run.ID)
	s.Require().Error(err)

	updated, updatedErr := s.Client(ctx).AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updatedErr)
	s.Equal(agentrun.StatusFailed, updated.Status)
	s.NotEmpty(updated.ErrorMessage)
	s.NotNil(updated.FinishedAt)
}
