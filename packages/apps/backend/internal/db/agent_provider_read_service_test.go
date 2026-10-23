package db

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	apr "github.com/rezible/rezible/ent/agentproviderread"
	"github.com/rezible/rezible/ent/agentturn"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
)

type AgentProviderReadServiceSuite struct {
	test.Suite
}

func TestAgentProviderReadServiceSuite(t *testing.T) {
	suite.Run(t, &AgentProviderReadServiceSuite{Suite: test.NewSuite()})
}

// runningTurn creates an agent session with one running turn, and a service its reads are of.
func (s *AgentProviderReadServiceSuite) runningTurn(ctx context.Context, tdb rez.Database) (*ent.AgentTurn, uuid.UUID, uuid.UUID) {
	client := tdb.Client(ctx)
	createSession := client.AgentSession.Create().
		SetAgentName("investigation").
		SetInput([]byte(`{"query":"Why is checkout failing?"}`))
	session, sessionErr := createSession.Save(ctx)
	s.Require().NoError(sessionErr)
	createTurn := client.AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(session.ID).
		SetSequence(1).
		SetRiverJobID(1).
		SetStatus(agentturn.StatusRunning)
	turn, turnErr := createTurn.Save(ctx)
	s.Require().NoError(turnErr)
	createService := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	service, serviceErr := createService.Save(ctx)
	s.Require().NoError(serviceErr)
	return turn, service.ID, createTestLogIntegration(ctx, s.Require(), tdb)
}

// createTestLogIntegration creates an installation for provider reads to be made through.
func createTestLogIntegration(ctx context.Context, require *require.Assertions, tdb rez.Database) uuid.UUID {
	createIntegration := tdb.Client(ctx).Integration.Create().
		SetProvider("provider-a").
		SetName("provider-a").
		SetDisplayName("Provider A").
		SetProviderInstallationRef("installation-1").
		SetInstallationConfig([]byte(`{}`))
	integration, integrationErr := createIntegration.Save(ctx)
	require.NoError(integrationErr)
	return integration.ID
}

func (s *AgentProviderReadServiceSuite) startParams(turn *ent.AgentTurn, serviceID, integrationID uuid.UUID) rez.StartAgentProviderReadParams {
	return rez.StartAgentProviderReadParams{
		AgentTurnID:     turn.ID,
		ToolName:        "search_logs",
		ServiceEntityID: serviceID,
		IntegrationID:   integrationID,
		Input:           json.RawMessage(`{"start":"2026-06-01T10:00:00Z","end":"2026-06-01T11:00:00Z"}`),
	}
}

func (s *AgentProviderReadServiceSuite) TestProviderReadBudgetCountsAcrossAttempts() {
	ctx, tdb := s.SetupTestDatabase()
	service := NewAgentProviderReadService(tdb, test.NewClock(time.Now()))
	turn, serviceID, integrationID := s.runningTurn(ctx, tdb)
	params := s.startParams(turn, serviceID, integrationID)

	// Reads up to the budget start, and the next is refused.
	for i := range rezai.MaxProviderReadsPerTurn {
		_, used, startErr := service.StartProviderRead(ctx, params)
		s.Require().NoError(startErr)
		s.Equal(i+1, used)
	}
	_, _, overErr := service.StartProviderRead(ctx, params)
	s.Require().ErrorIs(overErr, rezai.ErrToolRefused)

	// A resumed attempt of the same turn sees the reads its earlier attempts made, including those left running.
	updateTurn := tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).
		SetStatus(agentturn.StatusFailed)
	s.Require().NoError(updateTurn.Exec(ctx))
	resumeTurn := tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).
		SetStatus(agentturn.StatusRunning)
	s.Require().NoError(resumeTurn.Exec(ctx))
	used, countErr := service.CountTurnProviderReads(ctx, turn.ID)
	s.Require().NoError(countErr)
	s.Equal(rezai.MaxProviderReadsPerTurn, used)
	_, _, resumedErr := service.StartProviderRead(ctx, params)
	s.ErrorIs(resumedErr, rezai.ErrToolRefused)
}

func (s *AgentProviderReadServiceSuite) TestProviderReadsAreRecorded() {
	ctx, tdb := s.SetupTestDatabase()
	finishedAt := time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC)
	service := NewAgentProviderReadService(tdb, test.NewClock(finishedAt))
	turn, serviceID, integrationID := s.runningTurn(ctx, tdb)

	succeeded, used, startErr := service.StartProviderRead(ctx, s.startParams(turn, serviceID, integrationID))
	s.Require().NoError(startErr)
	s.Equal(1, used)
	s.Equal(apr.StatusRunning, succeeded.Status)
	s.Equal(turn.AgentSessionID, succeeded.AgentSessionID)
	s.Equal(&integrationID, succeeded.IntegrationID)
	record := json.RawMessage(`{"query":"{service=\"checkout-api\"}","source_url":"https://example.com/explore"}`)
	finishParams := rez.FinishAgentProviderReadParams{ID: succeeded.ID, Read: record}
	s.Require().NoError(service.FinishProviderRead(ctx, finishParams))

	failed, used, failedStartErr := service.StartProviderRead(ctx, s.startParams(turn, serviceID, integrationID))
	s.Require().NoError(failedStartErr)
	s.Equal(2, used)
	failParams := rez.FinishAgentProviderReadParams{ID: failed.ID, Error: "logs: 503 Service Unavailable"}
	s.Require().NoError(service.FinishProviderRead(ctx, failParams))

	storedSuccess, successErr := tdb.Client(ctx).AgentProviderRead.Get(ctx, succeeded.ID)
	s.Require().NoError(successErr)
	s.Equal(apr.StatusSucceeded, storedSuccess.Status)
	s.JSONEq(string(record), string(storedSuccess.Read))
	s.JSONEq(string(s.startParams(turn, serviceID, integrationID).Input), string(storedSuccess.Input))
	s.Nil(storedSuccess.Error)
	s.Require().NotNil(storedSuccess.FinishedAt)
	s.True(finishedAt.Equal(*storedSuccess.FinishedAt))

	storedFailure, failureErr := tdb.Client(ctx).AgentProviderRead.Get(ctx, failed.ID)
	s.Require().NoError(failureErr)
	s.Equal(apr.StatusFailed, storedFailure.Status)
	s.Require().NotNil(storedFailure.Error)
	s.Equal("logs: 503 Service Unavailable", *storedFailure.Error)
}
