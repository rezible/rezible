package db

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alertinvestigation"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	saentity "github.com/rezible/rezible/ent/systemanalysisentity"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type InvestigationServiceSuite struct {
	test.Suite
}

func TestInvestigationServiceSuite(t *testing.T) {
	suite.Run(t, &InvestigationServiceSuite{Suite: test.NewSuite()})
}

type investigationServiceHarness struct {
	tdb     rez.Database
	jobs    *mocks.MockJobService
	agents  *AgentSessionService
	service *InvestigationService
}

func (s *InvestigationServiceSuite) newHarness(tdb rez.Database) *investigationServiceHarness {
	jobs := mocks.NewMockJobService(s.T())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	agents := &AgentSessionService{
		logger: logger,
		db:     tdb,
		jobs:   jobs,
	}
	alerts := &AlertService{db: tdb}
	return &investigationServiceHarness{
		tdb:    tdb,
		jobs:   jobs,
		agents: agents,
		service: &InvestigationService{
			db:     tdb,
			alerts: alerts,
			agents: agents,
		},
	}
}

func (s *InvestigationServiceSuite) createAlertInstance(ctx context.Context, client *ent.Client, knowledgeEntityID *uuid.UUID) *ent.AlertInstance {
	createAlert := client.Alert.Create().
		SetTitle("Checkout alert").
		SetDescription("High error rate").
		SetDefinition("sum(rate(errors[5m])) > 1").
		SetNillableKnowledgeEntityID(knowledgeEntityID)
	alert := createAlert.SaveX(ctx)
	return client.AlertInstance.Create().
		SetAlertID(alert.ID).
		SaveX(ctx)
}

func (s *InvestigationServiceSuite) createAgentSession(ctx context.Context, client *ent.Client) *ent.AgentSession {
	inputJSON, jsonErr := json.Marshal(rezai.AlertAgentInput{AlertInstanceID: uuid.New()})
	s.Require().NoError(jsonErr)
	createSession := client.AgentSession.Create().
		SetAgentName(rezai.AlertsAgent.Name).
		SetInput(inputJSON)
	return createSession.SaveX(ctx)
}

func (s *InvestigationServiceSuite) TestCreateAlertInvestigationCreatesAnalysisAndSeedsAlertEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	createEntity := client.KnowledgeEntity.Create().
		SetKind(kne.KindSignal).
		SetSubkind("alert")
	entity := createEntity.SaveX(ctx)
	instance := s.createAlertInstance(ctx, client, &entity.ID)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.MatchedBy(func(args river.JobArgs) bool {
			_, ok := args.(jobs.StartAgentSession)
			return ok
		}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil).
		Once()

	investigation, createErr := h.service.CreateAlertInvestigation(ctx, instance.ID)
	s.Require().NoError(createErr)
	s.Require().NotNil(investigation)
	s.Equal(instance.ID, investigation.AlertInstanceID)

	session := client.AgentSession.GetX(ctx, investigation.AgentSessionID)
	s.Require().NotNil(session.SystemAnalysisID)
	analysis := client.SystemAnalysis.GetX(ctx, *session.SystemAnalysisID)
	s.Require().NotNil(analysis.SubjectEntityID)
	s.Equal(entity.ID, *analysis.SubjectEntityID)

	queryAnalysisEntities := client.SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID))
	analysisEntities := queryAnalysisEntities.AllX(ctx)
	s.Require().Len(analysisEntities, 1)
	s.Equal(entity.ID, analysisEntities[0].KnowledgeEntityID)
}

func (s *InvestigationServiceSuite) TestCreateAlertInvestigationRejectsAlertWithoutKnowledgeEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	instance := s.createAlertInstance(ctx, client, nil)

	sessionsBefore := client.AgentSession.Query().CountX(ctx)
	analysesBefore := client.SystemAnalysis.Query().CountX(ctx)
	investigationsBefore := client.AlertInvestigation.Query().CountX(ctx)
	investigation, createErr := h.service.CreateAlertInvestigation(ctx, instance.ID)
	s.Nil(investigation)
	s.ErrorIs(createErr, rez.ErrInvalidInput)
	s.Equal(sessionsBefore, client.AgentSession.Query().CountX(ctx))
	s.Equal(analysesBefore, client.SystemAnalysis.Query().CountX(ctx))
	s.Equal(investigationsBefore, client.AlertInvestigation.Query().CountX(ctx))
}

func (s *InvestigationServiceSuite) TestCreateAlertInvestigationRollsBackWhenStartJobFails() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	createEntity := client.KnowledgeEntity.Create().
		SetKind(kne.KindSignal).
		SetSubkind("alert")
	entity := createEntity.SaveX(ctx)
	instance := s.createAlertInstance(ctx, client, &entity.ID)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("job insert failed")).
		Once()

	querySessionsBefore := client.AgentSession.Query()
	sessionsBefore := querySessionsBefore.CountX(ctx)
	queryAnalysesBefore := client.SystemAnalysis.Query()
	analysesBefore := queryAnalysesBefore.CountX(ctx)
	queryInvestigationsBefore := client.AlertInvestigation.Query()
	investigationsBefore := queryInvestigationsBefore.CountX(ctx)

	investigation, createErr := h.service.CreateAlertInvestigation(ctx, instance.ID)
	s.Nil(investigation)
	s.Require().Error(createErr)
	querySessionsAfter := client.AgentSession.Query()
	s.Equal(sessionsBefore, querySessionsAfter.CountX(ctx))
	queryAnalysesAfter := client.SystemAnalysis.Query()
	s.Equal(analysesBefore, queryAnalysesAfter.CountX(ctx))
	queryInvestigationsAfter := client.AlertInvestigation.Query()
	s.Equal(investigationsBefore, queryInvestigationsAfter.CountX(ctx))
}

func (s *InvestigationServiceSuite) TestOnAgentTurnFinishedPersistsReportArtifact() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	instance := s.createAlertInstance(ctx, client, nil)
	session := s.createAgentSession(ctx, client)
	createInvestigation := client.AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(session.ID)
	investigation := createInvestigation.SaveX(ctx)
	report := schematypes.AlertInvestigationReport{
		Text:               "Checkout errors likely come from database saturation.",
		LikelyCause:        "database saturation",
		RecommendedActions: []string{"check database CPU"},
	}
	reportJSON, jsonErr := json.Marshal(report)
	s.Require().NoError(jsonErr)
	client.AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart(string(reportJSON))}).
		SaveX(ctx)

	event := &rezai.EventOnAgentTurnFinished{AgentSessionId: session.ID}
	s.Require().NoError(h.service.onAgentTurnFinished(ctx, event))

	updated := client.AlertInvestigation.GetX(ctx, investigation.ID)
	s.Equal(report.Text, updated.Report.Text)
	s.Equal(report.LikelyCause, updated.Report.LikelyCause)
	s.Equal(report.RecommendedActions, updated.Report.RecommendedActions)
}

func (s *InvestigationServiceSuite) TestOnAgentTurnFinishedIgnoresNonInvestigationSessionsAndMissingArtifacts() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	nonInvestigationSession := s.createAgentSession(ctx, client)
	client.AgentArtifact.Create().
		SetAgentSessionID(nonInvestigationSession.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart("{")}).
		SaveX(ctx)
	s.NoError(h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: nonInvestigationSession.ID}))

	instance := s.createAlertInstance(ctx, client, nil)
	investigationSession := s.createAgentSession(ctx, client)
	client.AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(investigationSession.ID).
		SaveX(ctx)
	s.NoError(h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: investigationSession.ID}))
}

func (s *InvestigationServiceSuite) TestOnAgentTurnFinishedRejectsMalformedReportArtifact() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	instance := s.createAlertInstance(ctx, client, nil)
	session := s.createAgentSession(ctx, client)
	client.AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(session.ID).
		SaveX(ctx)
	client.AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart("{")}).
		SaveX(ctx)

	err := h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: session.ID})
	s.Require().Error(err)
}

func (s *InvestigationServiceSuite) TestReportArtifactRequiresJSONPart() {
	//tdb := s.CreateTestDatabase()
	h := s.newHarness(nil)
	_, err := h.service.alertInvestigationReportFromArtifact([]*ai.Part{ai.NewTextPart("not json")})
	s.Error(err)
}

func (s *InvestigationServiceSuite) TestSetAlertInvestigationReportTrimsAndValidatesText() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	instance := s.createAlertInstance(ctx, client, nil)
	session := s.createAgentSession(ctx, client)
	client.AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(session.ID).
		SaveX(ctx)

	_, emptyErr := h.service.SetAlertInvestigationReport(ctx, session.ID, schematypes.AlertInvestigationReport{Text: "  "})
	s.Error(emptyErr)

	updated, updateErr := h.service.SetAlertInvestigationReport(ctx, session.ID, schematypes.AlertInvestigationReport{Text: "  ok  "})
	s.Require().NoError(updateErr)
	s.Equal("ok", updated.Report.Text)
}

func (s *InvestigationServiceSuite) TestInvestigationReportArtifactQueryUsesReportName() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	h := s.newHarness(tdb)
	instance := s.createAlertInstance(ctx, client, nil)
	session := s.createAgentSession(ctx, client)
	createInvestigation := client.AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(session.ID)
	investigation := createInvestigation.SaveX(ctx)
	report := schematypes.AlertInvestigationReport{Text: "real report"}
	reportJSON, jsonErr := json.Marshal(report)
	s.Require().NoError(jsonErr)
	client.AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("other_report").
		SetParts([]*ai.Part{ai.NewJSONPart(`{"text":"wrong"}`)}).
		SaveX(ctx)
	client.AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart(string(reportJSON))}).
		SaveX(ctx)

	s.Require().NoError(h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: session.ID}))
	queryUpdated := client.AlertInvestigation.Query().
		Where(alertinvestigation.ID(investigation.ID)).
		WithAgentSession()
	updated := queryUpdated.OnlyX(ctx)
	s.Equal("real report", updated.Report.Text)
}
