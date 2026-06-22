package eino

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunartifact"
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

	runRequest := rez.AgentRunRequest{
		WorkflowKind:   agentrun.WorkflowKindIncidentContextPack,
		IdempotencyKey: "incident-context-pack:test",
		SubjectKind:    "incident",
		SubjectID:      uuid.New(),
		Metadata:       map[string]any{"trigger": "test"},
	}
	run, err := svc.RequestRun(ctx, runRequest)
	s.Require().NoError(err)
	s.Equal(agentrun.StatusQueued, run.Status)
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
	params := rez.AgentRunRequest{
		WorkflowKind:   agentrun.WorkflowKindAlertInvestigation,
		IdempotencyKey: "alert-investigation:test",
		SubjectKind:    "alert",
		SubjectID:      uuid.New(),
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

	createRun := dbc.AgentRun.Create().
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

	createArtifacts := dbc.AgentRunArtifact.Query().
		Where(agentrunartifact.AgentRunID(run.ID))
	artifacts, artifactsErr := createArtifacts.All(ctx)
	s.Require().NoError(artifactsErr)
	s.Len(artifacts, 2)
	s.Contains([]agentrunartifact.Kind{artifacts[0].Kind, artifacts[1].Kind}, agentrunartifact.KindResult)
	s.Contains([]agentrunartifact.Kind{artifacts[0].Kind, artifacts[1].Kind}, agentrunartifact.KindModel)
}

func (s *AgentServiceSuite) TestRunWorkflowMarksUnknownWorkflowFailed() {
	ctx := s.SeedTenantContext()
	jobSvc := mocks.NewMockJobService(s.T())
	msgSvc := mocks.NewMockMessageService(s.T())
	msgSvc.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	dbc := s.Client(ctx)

	createRun := dbc.AgentRun.Create().
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
}
