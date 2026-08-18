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
	jobs    *mocks.MockJobService
	agents  *AgentSessionService
	service *InvestigationService
}

func (s *InvestigationServiceSuite) newHarness() *investigationServiceHarness {
	jobs := mocks.NewMockJobService(s.T())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	agents := &AgentSessionService{
		logger: logger,
		db:     s.Database(),
		jobs:   jobs,
	}
	alerts := &AlertService{db: s.Database()}
	return &investigationServiceHarness{
		jobs:   jobs,
		agents: agents,
		service: &InvestigationService{
			db:     s.Database(),
			alerts: alerts,
			agents: agents,
		},
	}
}

func (s *InvestigationServiceSuite) createAlertInstance(ctx context.Context, knowledgeEntityID *uuid.UUID) *ent.AlertInstance {
	createAlert := s.Client(ctx).Alert.Create().
		SetTitle("Checkout alert").
		SetDescription("High error rate").
		SetDefinition("sum(rate(errors[5m])) > 1").
		SetNillableKnowledgeEntityID(knowledgeEntityID)
	alert := createAlert.SaveX(ctx)
	return s.Client(ctx).AlertInstance.Create().
		SetAlertID(alert.ID).
		SaveX(ctx)
}

func (s *InvestigationServiceSuite) createAgentSession(ctx context.Context) *ent.AgentSession {
	inputJSON, jsonErr := json.Marshal(rezai.AlertAgentInput{AlertInstanceID: uuid.New()})
	s.Require().NoError(jsonErr)
	createSession := s.Client(ctx).AgentSession.Create().
		SetAgentName(rezai.AlertsAgent.Name).
		SetInput(inputJSON)
	return createSession.SaveX(ctx)
}

func (s *InvestigationServiceSuite) TestCreateAlertInvestigationCreatesAnalysisAndSeedsAlertEntity() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	createEntity := s.Client(ctx).KnowledgeEntity.Create().
		SetKind(kne.KindSignal).
		SetSubkind("alert")
	entity := createEntity.SaveX(ctx)
	instance := s.createAlertInstance(ctx, &entity.ID)

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

	session := s.Client(ctx).AgentSession.GetX(ctx, investigation.AgentSessionID)
	s.Require().NotNil(session.SystemAnalysisID)
	analysis := s.Client(ctx).SystemAnalysis.GetX(ctx, *session.SystemAnalysisID)
	s.Require().NotNil(analysis.SubjectEntityID)
	s.Equal(entity.ID, *analysis.SubjectEntityID)

	queryAnalysisEntities := s.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID))
	analysisEntities := queryAnalysisEntities.AllX(ctx)
	s.Require().Len(analysisEntities, 1)
	s.Equal(entity.ID, analysisEntities[0].KnowledgeEntityID)
}

func (s *InvestigationServiceSuite) TestCreateAlertInvestigationWorksWithoutAlertEntity() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	instance := s.createAlertInstance(ctx, nil)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1002}}, nil).
		Once()

	investigation, createErr := h.service.CreateAlertInvestigation(ctx, instance.ID)
	s.Require().NoError(createErr)
	session := s.Client(ctx).AgentSession.GetX(ctx, investigation.AgentSessionID)
	s.Require().NotNil(session.SystemAnalysisID)
	analysis := s.Client(ctx).SystemAnalysis.GetX(ctx, *session.SystemAnalysisID)
	s.Nil(analysis.SubjectEntityID)
	queryAnalysisEntities := s.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID))
	s.Equal(0, queryAnalysisEntities.CountX(ctx))
}

func (s *InvestigationServiceSuite) TestCreateAlertInvestigationRollsBackWhenStartJobFails() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	createEntity := s.Client(ctx).KnowledgeEntity.Create().
		SetKind(kne.KindSignal).
		SetSubkind("alert")
	entity := createEntity.SaveX(ctx)
	instance := s.createAlertInstance(ctx, &entity.ID)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("job insert failed")).
		Once()

	querySessionsBefore := s.Client(ctx).AgentSession.Query()
	sessionsBefore := querySessionsBefore.CountX(ctx)
	queryAnalysesBefore := s.Client(ctx).SystemAnalysis.Query()
	analysesBefore := queryAnalysesBefore.CountX(ctx)
	queryInvestigationsBefore := s.Client(ctx).AlertInvestigation.Query()
	investigationsBefore := queryInvestigationsBefore.CountX(ctx)

	investigation, createErr := h.service.CreateAlertInvestigation(ctx, instance.ID)
	s.Nil(investigation)
	s.Require().Error(createErr)
	querySessionsAfter := s.Client(ctx).AgentSession.Query()
	s.Equal(sessionsBefore, querySessionsAfter.CountX(ctx))
	queryAnalysesAfter := s.Client(ctx).SystemAnalysis.Query()
	s.Equal(analysesBefore, queryAnalysesAfter.CountX(ctx))
	queryInvestigationsAfter := s.Client(ctx).AlertInvestigation.Query()
	s.Equal(investigationsBefore, queryInvestigationsAfter.CountX(ctx))
}

func (s *InvestigationServiceSuite) TestOnAgentTurnFinishedPersistsReportArtifact() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	instance := s.createAlertInstance(ctx, nil)
	session := s.createAgentSession(ctx)
	createInvestigation := s.Client(ctx).AlertInvestigation.Create().
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
	s.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart(string(reportJSON))}).
		SaveX(ctx)

	event := &rezai.EventOnAgentTurnFinished{AgentSessionId: session.ID}
	s.Require().NoError(h.service.onAgentTurnFinished(ctx, event))

	updated := s.Client(ctx).AlertInvestigation.GetX(ctx, investigation.ID)
	s.Equal(report.Text, updated.Report.Text)
	s.Equal(report.LikelyCause, updated.Report.LikelyCause)
	s.Equal(report.RecommendedActions, updated.Report.RecommendedActions)
}

func (s *InvestigationServiceSuite) TestOnAgentTurnFinishedIgnoresNonInvestigationSessionsAndMissingArtifacts() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	nonInvestigationSession := s.createAgentSession(ctx)
	s.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(nonInvestigationSession.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart("{")}).
		SaveX(ctx)
	s.NoError(h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: nonInvestigationSession.ID}))

	instance := s.createAlertInstance(ctx, nil)
	investigationSession := s.createAgentSession(ctx)
	s.Client(ctx).AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(investigationSession.ID).
		SaveX(ctx)
	s.NoError(h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: investigationSession.ID}))
}

func (s *InvestigationServiceSuite) TestOnAgentTurnFinishedRejectsMalformedReportArtifact() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	instance := s.createAlertInstance(ctx, nil)
	session := s.createAgentSession(ctx)
	s.Client(ctx).AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(session.ID).
		SaveX(ctx)
	s.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart("{")}).
		SaveX(ctx)

	err := h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: session.ID})
	s.Require().Error(err)
}

func (s *InvestigationServiceSuite) TestReportArtifactRequiresJSONPart() {
	h := s.newHarness()
	_, err := h.service.alertInvestigationReportFromArtifact([]*ai.Part{ai.NewTextPart("not json")})
	s.Error(err)
}

func (s *InvestigationServiceSuite) TestSetAlertInvestigationReportTrimsAndValidatesText() {
	ctx := s.SeedTenantContext()
	h := s.newHarness()
	instance := s.createAlertInstance(ctx, nil)
	session := s.createAgentSession(ctx)
	s.Client(ctx).AlertInvestigation.Create().
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
	h := s.newHarness()
	instance := s.createAlertInstance(ctx, nil)
	session := s.createAgentSession(ctx)
	createInvestigation := s.Client(ctx).AlertInvestigation.Create().
		SetAlertInstanceID(instance.ID).
		SetAgentSessionID(session.ID)
	investigation := createInvestigation.SaveX(ctx)
	report := schematypes.AlertInvestigationReport{Text: "real report"}
	reportJSON, jsonErr := json.Marshal(report)
	s.Require().NoError(jsonErr)
	s.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("other_report").
		SetParts([]*ai.Part{ai.NewJSONPart(`{"text":"wrong"}`)}).
		SaveX(ctx)
	s.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetName("investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart(string(reportJSON))}).
		SaveX(ctx)

	s.Require().NoError(h.service.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: session.ID}))
	queryUpdated := s.Client(ctx).AlertInvestigation.Query().
		Where(alertinvestigation.ID(investigation.ID)).
		WithAgentSession()
	updated := queryUpdated.OnlyX(ctx)
	s.Equal("real report", updated.Report.Text)
}
