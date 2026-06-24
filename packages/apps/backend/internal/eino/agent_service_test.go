package eino

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentcase"
	"github.com/rezible/rezible/ent/agentcaseartifact"
	"github.com/rezible/rezible/ent/agentcasestep"
	"github.com/rezible/rezible/ent/agentrun"
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

func (s *AgentServiceSuite) newService(jobSvc rez.JobService, msgSvc rez.MessageService) *AgentService {
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
		workflows: map[agentrun.WorkflowKind]agentWorkflow{
			agentrun.WorkflowKindIncidentContextPack:   stubWorkflow{kind: agentrun.WorkflowKindIncidentContextPack},
			agentrun.WorkflowKindAlertInvestigation:    stubWorkflow{kind: agentrun.WorkflowKindAlertInvestigation},
			agentrun.WorkflowKindRetrospectiveAnalysis: stubWorkflow{kind: agentrun.WorkflowKindRetrospectiveAnalysis},
		},
	}
}

func (s *AgentServiceSuite) TestRequestRunCreatesRunAndPreservesContextForEnqueue() {
	ctx := s.SeedTenantContext()
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

	svc := s.newService(jobSvc, msgSvc)
	agentCase, caseErr := s.Client(ctx).AgentCase.Create().
		SetStatus(agentcase.StatusOpen).
		SetTitle("Incident context pack").
		SetWorkflowKind(agentcase.WorkflowKindIncidentContextPack).
		SetSubjectKind("incident").
		SetSubjectID(uuid.New()).
		SetTriggerMetadata(map[string]any{"trigger": "test"}).
		Save(ctx)
	s.Require().NoError(caseErr)

	runRequest := rez.AgentRunRequest{
		AgentCaseID:    agentCase.ID,
		WorkflowKind:   agentrun.WorkflowKindIncidentContextPack,
		IdempotencyKey: "incident-context-pack:test",
		SubjectKind:    "incident",
		SubjectID:      *agentCase.SubjectID,
		Metadata:       map[string]any{"trigger": "test"},
	}
	run, err := svc.RequestRun(ctx, runRequest)
	s.Require().NoError(err)
	s.Equal(agentrun.StatusQueued, run.Status)
	s.Equal(agentCase.ID, *run.AgentCaseID)
	s.Equal(agentrun.WorkflowKindIncidentContextPack, run.WorkflowKind)
	s.Equal(runRequest.SubjectID, *run.SubjectID)
}

func (s *AgentServiceSuite) TestRequestRunReusesExistingRunForIdempotencyKey() {
	ctx := s.SeedTenantContext()
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	jobSvc.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Twice()
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc := s.newService(jobSvc, msgSvc)
	agentCase, caseErr := s.Client(ctx).AgentCase.Create().
		SetStatus(agentcase.StatusOpen).
		SetTitle("Alert investigation").
		SetWorkflowKind(agentcase.WorkflowKindAlertInvestigation).
		SetSubjectKind("alert").
		SetSubjectID(uuid.New()).
		SetTriggerMetadata(map[string]any{"trigger": "test"}).
		Save(ctx)
	s.Require().NoError(caseErr)

	params := rez.AgentRunRequest{
		AgentCaseID:    agentCase.ID,
		WorkflowKind:   agentrun.WorkflowKindAlertInvestigation,
		IdempotencyKey: "alert-investigation:test",
		SubjectKind:    "alert",
		SubjectID:      *agentCase.SubjectID,
	}

	first, firstErr := svc.RequestRun(ctx, params)
	s.Require().NoError(firstErr)

	second, secondErr := svc.RequestRun(ctx, params)
	s.Require().NoError(secondErr)

	s.Equal(first.ID, second.ID)
}

func (s *AgentServiceSuite) TestRunWorkflowCompletesStubAndPersistsArtifacts() {
	ctx := s.SeedTenantContext()
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc := s.newService(jobSvc, msgSvc)
	dbc := s.Client(ctx)

	agentCase, caseErr := dbc.AgentCase.Create().
		SetStatus(agentcase.StatusOpen).
		SetTitle("Retrospective analysis").
		SetWorkflowKind(agentcase.WorkflowKindRetrospectiveAnalysis).
		SetSubjectKind("incident").
		SetSubjectID(uuid.New()).
		SetTriggerMetadata(map[string]any{"trigger": "test"}).
		Save(ctx)
	s.Require().NoError(caseErr)

	createRun := dbc.AgentRun.Create().
		SetAgentCaseID(agentCase.ID).
		SetWorkflowKind(agentrun.WorkflowKindRetrospectiveAnalysis).
		SetStatus(agentrun.StatusQueued).
		SetIdempotencyKey("retrospective:test").
		SetModelMetadata(map[string]any{"provider": "gemini"})

	run, createErr := createRun.Save(ctx)
	s.Require().NoError(createErr)

	runErr := svc.RunWorkflow(ctx, run.ID)
	s.Require().NoError(runErr)

	updated, updateErr := dbc.AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updateErr)
	s.Equal(agentrun.StatusSucceeded, updated.Status)
	s.NotNil(updated.StartedAt)
	s.NotNil(updated.CompletedAt)

	updatedCase, caseUpdateErr := dbc.AgentCase.Get(ctx, agentCase.ID)
	s.Require().NoError(caseUpdateErr)
	s.Equal(agentcase.StatusCompleted, updatedCase.Status)

	steps, stepsErr := dbc.AgentCaseStep.Query().
		Where(agentcasestep.AgentCaseID(agentCase.ID)).
		Order(agentcasestep.BySequence()).
		All(ctx)
	s.Require().NoError(stepsErr)
	s.Len(steps, 2)
	s.Equal(agentcasestep.KindSystem, steps[0].Kind)
	s.Equal(agentcasestep.KindConclusion, steps[1].Kind)

	artifacts, artifactsErr := dbc.AgentCaseArtifact.Query().
		Where(agentcaseartifact.AgentCaseID(agentCase.ID)).
		All(ctx)
	s.Require().NoError(artifactsErr)
	s.Len(artifacts, 1)
	s.Equal(agentcaseartifact.KindModel, artifacts[0].Kind)
}

func (s *AgentServiceSuite) TestRunWorkflowMarksUnknownWorkflowFailed() {
	ctx := s.SeedTenantContext()
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	dbc := s.Client(ctx)
	agentCase, caseErr := dbc.AgentCase.Create().
		SetStatus(agentcase.StatusOpen).
		SetTitle("Alert investigation").
		SetWorkflowKind(agentcase.WorkflowKindAlertInvestigation).
		SetSubjectKind("alert").
		SetSubjectID(uuid.New()).
		SetTriggerMetadata(map[string]any{"trigger": "test"}).
		Save(ctx)
	s.Require().NoError(caseErr)

	createRun := dbc.AgentRun.Create().
		SetAgentCaseID(agentCase.ID).
		SetWorkflowKind(agentrun.WorkflowKindAlertInvestigation).
		SetStatus(agentrun.StatusQueued).
		SetIdempotencyKey("missing:test")
	run, createErr := createRun.Save(ctx)
	s.Require().NoError(createErr)

	svc := s.newService(jobSvc, msgSvc)
	delete(svc.workflows, agentrun.WorkflowKindAlertInvestigation)

	err := svc.RunWorkflow(ctx, run.ID)
	s.Require().Error(err)

	updated, updatedErr := dbc.AgentRun.Get(ctx, run.ID)
	s.Require().NoError(updatedErr)
	s.Equal(agentrun.StatusFailed, updated.Status)
	s.Equal("unknown_workflow", updated.ErrorCode)
	s.NotEmpty(updated.ErrorMessage)
	s.NotNil(updated.FailedAt)

	updatedCase, caseUpdateErr := dbc.AgentCase.Get(ctx, agentCase.ID)
	s.Require().NoError(caseUpdateErr)
	s.Equal(agentcase.StatusFailed, updatedCase.Status)
	s.Equal("unknown_workflow", updatedCase.ErrorCode)
}
