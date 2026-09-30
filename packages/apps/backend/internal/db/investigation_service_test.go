package db

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sae "github.com/rezible/rezible/ent/systemanalysisentity"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sar "github.com/rezible/rezible/ent/systemanalysisrelationship"
)

type InvestigationServiceSuite struct {
	test.Suite
}

func TestInvestigationServiceSuite(t *testing.T) {
	suite.Run(t, &InvestigationServiceSuite{
		Suite: test.NewSuite(),
	})
}

func (s *InvestigationServiceSuite) newService(tdb rez.Database, jobService *mocks.MockJobService) *InvestigationService {
	agentService := &AiAgentSessionService{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		db:     tdb,
		jobs:   jobService,
	}
	return NewInvestigationService(tdb, agentService, jobService)
}

func (s *InvestigationServiceSuite) expectStartJob(jobService *mocks.MockJobService, result *rivertype.JobInsertResult, insertErr error) {
	jobService.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), (*river.InsertOpts)(nil)).
		Return(result, insertErr).
		Once()
}

func (s *InvestigationServiceSuite) createAnalysis(tdb rez.Database, ctx context.Context) *ent.SystemAnalysis {
	analysis, analysisErr := tdb.Client(ctx).SystemAnalysis.Create().Save(ctx)
	s.Require().NoError(analysisErr)

	return analysis
}

func (s *InvestigationServiceSuite) TestCreateInvestigationUsesPreparedAnalysisAndPreservesContext() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	now := time.Now().UTC()

	createSource := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	source, sourceErr := createSource.Save(ctx)
	s.Require().NoError(sourceErr)

	createTarget := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("database")
	target, targetErr := createTarget.Save(ctx)
	s.Require().NoError(targetErr)

	createRelationship := client.KnowledgeRelationship.Create().
		SetPredicate(knr.PredicateUses).
		SetSourceEntityID(source.ID).
		SetTargetEntityID(target.ID)
	relationship, relationshipErr := createRelationship.Save(ctx)
	s.Require().NoError(relationshipErr)

	encodedAttributes, encodeErr := projections.EncodeAttributes(struct{}{})
	s.Require().NoError(encodeErr)

	providerResourceRef := "service:" + uuid.NewString()
	createEvent := client.NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(providerResourceRef).
		SetProviderEventSource("investigation-tests").
		SetProviderEventRef(uuid.NewString()).
		SetKind("deployment").
		SetAttributes(encodedAttributes).
		SetOccurredAt(now).
		SetReceivedAt(now)
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)

	createAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(providerResourceRef).
		SetEntityID(source.ID)
	alias, aliasErr := createAlias.Save(ctx)
	s.Require().NoError(aliasErr)

	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(alias.ID).
		SetKind(kev.KindObserved).
		SetAssertion("deployment_observed").
		SetEffectiveAt(now).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Checkout API",
		})
	evidence, evidenceErr := createEvidence.Save(ctx)
	s.Require().NoError(evidenceErr)

	queryService, queryServiceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queryServiceErr)

	analysisService, analysisServiceErr := NewSystemAnalysisService(tdb, queryService)
	s.Require().NoError(analysisServiceErr)

	analysis := s.createAnalysis(tdb, ctx)
	includeSystemAnalysisSubjectsParams := rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      analysis.ID,
		RelationshipIds: []uuid.UUID{relationship.ID},
	}

	s.Require().NoError(analysisService.IncludeSystemAnalysisSubjects(ctx, includeSystemAnalysisSubjectsParams))

	entryParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       "observation",
		Title:      "Payment error rate increased",
		Body:       "A deployment preceded the increase.",
		SetSubjects: []rez.SetSystemAnalysisEntrySubjectParams{
			{
				Role:              "primary",
				KnowledgeEntityID: &source.ID,
			},
			{
				Role:                    "affected",
				KnowledgeRelationshipID: &relationship.ID,
			},
			{
				Role:                "evidence_for",
				KnowledgeEvidenceID: &evidence.ID,
			},
		},
	}
	entry, entryErr := analysisService.SetSystemAnalysisEntry(ctx, uuid.Nil, entryParams)
	s.Require().NoError(entryErr)

	jobService := mocks.NewMockJobService(s.T())
	s.expectStartJob(jobService, &rivertype.JobInsertResult{
		Job: &rivertype.JobRow{
			ID: 1001,
		},
	}, nil)
	service := s.newService(tdb, jobService)
	investigationParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      "  Why did the payment error rate increase?  ",
	}

	investigation, createErr := service.CreateInvestigation(ctx, investigationParams)
	s.Require().NoError(createErr)
	s.NotEqual(uuid.Nil, investigation.ID)
	s.Equal(analysis.ID, investigation.SystemAnalysisID)
	s.Require().NotNil(investigation.Edges.SystemAnalysis)
	s.Require().NotNil(investigation.Edges.AgentSession)

	var sessionInput rezai.InvestigationAgentSessionInput
	s.Require().NoError(json.Unmarshal(investigation.Edges.AgentSession.Input, &sessionInput))
	s.Equal("Why did the payment error rate increase?", sessionInput.Query)
	querySystemAnalysisEntity := client.SystemAnalysisEntity.Query().
		Where(sae.AnalysisID(analysis.ID))
	systemAnalysisEntityCount, systemAnalysisEntityCountErr := querySystemAnalysisEntity.Count(ctx)
	s.Require().NoError(systemAnalysisEntityCountErr)

	s.Equal(2, systemAnalysisEntityCount)
	querySystemAnalysisRelationship := client.SystemAnalysisRelationship.Query().
		Where(sar.AnalysisID(analysis.ID))
	systemAnalysisRelationshipCount, systemAnalysisRelationshipCountErr := querySystemAnalysisRelationship.Count(ctx)
	s.Require().NoError(systemAnalysisRelationshipCountErr)

	s.Equal(1, systemAnalysisRelationshipCount)

	queryEntrySubjects := client.SystemAnalysisEntrySubject.Query().
		Where(saes.EntryID(entry.ID))
	entrySubjects, entrySubjectsErr := queryEntrySubjects.All(ctx)
	s.Require().NoError(entrySubjectsErr)
	s.Require().Len(entrySubjects, 3)
	var attachedEvidence bool
	for _, subject := range entrySubjects {
		attachedEvidence = attachedEvidence || (subject.KnowledgeEvidenceID != nil && *subject.KnowledgeEvidenceID == evidence.ID)
	}
	s.True(attachedEvidence)
}

func (s *InvestigationServiceSuite) TestCreateInvestigationValidatesOwnershipQuestionAndEmptyAnalysis() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	analysis := s.createAnalysis(tdb, ctx)
	jobService := mocks.NewMockJobService(s.T())
	s.expectStartJob(jobService, &rivertype.JobInsertResult{
		Job: &rivertype.JobRow{
			ID: 1001,
		},
	}, nil)
	service := s.newService(tdb, jobService)

	missingAnalysisParams := rez.CreateInvestigationParams{
		Query: "question",
	}

	_, missingAnalysisErr := service.CreateInvestigation(ctx, missingAnalysisParams)
	s.ErrorIs(missingAnalysisErr, rez.ErrInvalidInput)
	missingQuestionParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      " \t ",
	}

	_, missingQuestionErr := service.CreateInvestigation(ctx, missingQuestionParams)
	s.ErrorIs(missingQuestionErr, rez.ErrInvalidInput)

	investigationParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      "Question about empty context",
	}

	investigation, createErr := service.CreateInvestigation(ctx, investigationParams)
	s.Require().NoError(createErr)
	s.NotEqual(uuid.Nil, investigation.ID)
	querySystemAnalysisEntity := tdb.Client(ctx).SystemAnalysisEntity.Query().
		Where(sae.AnalysisID(analysis.ID))
	systemAnalysisEntityCount, systemAnalysisEntityCountErr := querySystemAnalysisEntity.Count(ctx)
	s.Require().NoError(systemAnalysisEntityCountErr)

	s.Equal(0, systemAnalysisEntityCount)
	querySystemAnalysisRelationship := tdb.Client(ctx).SystemAnalysisRelationship.Query().
		Where(sar.AnalysisID(analysis.ID))
	systemAnalysisRelationshipCount, systemAnalysisRelationshipCountErr := querySystemAnalysisRelationship.Count(ctx)
	s.Require().NoError(systemAnalysisRelationshipCountErr)

	s.Equal(0, systemAnalysisRelationshipCount)
	tenantInputParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "follow-up question",
		SubmissionKey:   "tenant-context-is-not-a-user",
	}

	_, tenantInputErr := service.SubmitInvestigationUserInput(ctx, tenantInputParams)
	s.ErrorIs(tenantInputErr, rez.ErrInvalidInput)

	inUseParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      "second question",
	}

	_, inUseErr := service.CreateInvestigation(ctx, inUseParams)
	s.ErrorIs(inUseErr, rez.ErrConflict)
	investigationCount, investigationCountErr := tdb.Client(ctx).Investigation.Query().Count(ctx)
	s.Require().NoError(investigationCountErr)

	s.Equal(1, investigationCount)
}

func (s *InvestigationServiceSuite) TestCreateInvestigationRollsBackWhenSessionStartupCannotBeQueued() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	analysis := s.createAnalysis(tdb, ctx)
	jobService := mocks.NewMockJobService(s.T())
	s.expectStartJob(jobService, nil, errors.New("queue unavailable"))
	service := s.newService(tdb, jobService)

	createParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      "Question requiring a queued agent",
	}

	_, createErr := service.CreateInvestigation(ctx, createParams)
	s.Error(createErr)
	investigationCount, investigationCountErr := tdb.Client(ctx).Investigation.Query().Count(ctx)
	s.Require().NoError(investigationCountErr)

	s.Equal(0, investigationCount)
	agentSessionCount, agentSessionCountErr := tdb.Client(ctx).AgentSession.Query().Count(ctx)
	s.Require().NoError(agentSessionCountErr)

	s.Equal(0, agentSessionCount)
}

func (s *InvestigationServiceSuite) TestConcurrentInvestigationsClaimAnalysisOnlyOnce() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	analysis := s.createAnalysis(tdb, ctx)
	jobService := mocks.NewMockJobService(s.T())
	enteredSessionStart := make(chan struct{}, 2)
	releaseSessionStart := make(chan struct{})
	createResults := make(chan error, 2)
	jobService.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), (*river.InsertOpts)(nil)).
		Run(func(context.Context, river.JobArgs, *river.InsertOpts) {
			enteredSessionStart <- struct{}{}
			<-releaseSessionStart
		}).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 1001,
			},
		}, nil).
		Twice()
	service := s.newService(tdb, jobService)
	for range 2 {
		go func() {
			createParams := rez.CreateInvestigationParams{
				AnalysisID: analysis.ID,
				Query:      "Concurrent analysis question",
			}

			_, createErr := service.CreateInvestigation(ctx, createParams)
			createResults <- createErr
		}()
	}

	for range 2 {
		select {
		case <-enteredSessionStart:
		case <-time.After(5 * time.Second):
			close(releaseSessionStart)
			s.FailNow("both requests did not reach session startup")
		}
	}
	close(releaseSessionStart)

	var conflictCount int
	for range 2 {
		createErr := <-createResults
		if createErr != nil {
			s.ErrorIs(createErr, rez.ErrConflict)
			conflictCount++
		}
	}
	s.Equal(1, conflictCount)
	investigationCount, investigationCountErr := tdb.Client(ctx).Investigation.Query().Count(ctx)
	s.Require().NoError(investigationCountErr)

	s.Equal(1, investigationCount)
	agentSessionCount, agentSessionCountErr := tdb.Client(ctx).AgentSession.Query().Count(ctx)
	s.Require().NoError(agentSessionCountErr)

	s.Equal(1, agentSessionCount)
}
