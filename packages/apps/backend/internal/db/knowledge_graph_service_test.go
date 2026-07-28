package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

type KnowledgeGraphServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphServiceSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeGraphServiceSuite) knowledgeService() *KnowledgeGraphService {
	return &KnowledgeGraphService{db: s.Database()}
}

func (s *KnowledgeGraphServiceSuite) createEvent(kind projections.SubjectKind, eventKind ne.Kind, subjectRef string, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)
	event, createErr := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("knowledge-graph-tests").
		SetProviderEventRef(uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetKind(eventKind).
		SetSubjectKind(kind.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encodedAttributes).
		Save(ctx)
	s.Require().NoError(createErr)
	return event
}

func (s *KnowledgeGraphServiceSuite) TestCurrentStateAndBoundedView() {
	ctx := s.SeedTenantContext()
	service := s.knowledgeService()
	now := time.Now().UTC()
	sourceAlias := ent.KnowledgeAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: "service:api",
	}
	targetAlias := ent.KnowledgeAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: "service:database",
	}

	sourceEvent := s.createEvent(projections.SubjectKindSystemComponent, ne.KindObserved, sourceAlias.ProviderSubjectRef, now.Add(-2*time.Hour), struct{}{})
	source, ingestErr := service.IngestEntityEvidence(ctx, sourceEvent, ent.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now.Add(-2 * time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Old API name",
		},
		SubjectEntity: &ent.KnowledgeEntityRef{Kind: "system_component", Alias: sourceAlias},
	})
	s.Require().NoError(ingestErr)

	latestEvent := s.createEvent(projections.SubjectKindSystemComponent, ne.KindObserved, sourceAlias.ProviderSubjectRef, now.Add(-time.Hour), struct{}{})
	_, ingestErr = service.IngestEntityEvidence(ctx, latestEvent, ent.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now.Add(-time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "API",
		},
		SubjectEntity: &ent.KnowledgeEntityRef{Kind: "system_component", Alias: sourceAlias},
	})
	s.Require().NoError(ingestErr)

	targetEvent := s.createEvent(projections.SubjectKindSystemComponent, ne.KindObserved, targetAlias.ProviderSubjectRef, now.Add(-time.Hour), struct{}{})
	_, ingestErr = service.IngestEntityEvidence(ctx, targetEvent, ent.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "component_exists",
		EffectiveAt:  now.Add(-time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Database"},
		SubjectEntity: &ent.KnowledgeEntityRef{
			Kind:  "system_component",
			Alias: targetAlias,
		},
	})
	s.Require().NoError(ingestErr)

	relationshipAlias := ent.KnowledgeAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: "api-uses-database",
	}
	relationshipEvent := s.createEvent(projections.SubjectKindSystemRelationship, ne.KindObserved, relationshipAlias.ProviderSubjectRef, now, struct{}{})
	_, ingestErr = service.IngestEvidenceBulk(ctx, relationshipEvent, ent.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "relationship_exists",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "uses"},
		SubjectRelationship: &ent.KnowledgeRelationshipRef{
			Kind:   "uses",
			Alias:  relationshipAlias,
			Source: ent.KnowledgeEntityRef{Kind: "system_component", Alias: sourceAlias},
			Target: ent.KnowledgeEntityRef{Kind: "system_component", Alias: targetAlias},
		},
	})
	s.Require().NoError(ingestErr)

	current, getErr := service.GetEntity(ctx, source.ID)
	s.Require().NoError(getErr)

	currEv := current.LatestEvidence()
	s.Require().NotNil(currEv)
	s.Require().NotNil(currEv.SubjectState)
	s.Equal("API", currEv.SubjectState.DisplayName)

	viewParams := rez.GetKnowledgeGraphViewParams{Depth: 1, EntityID: source.ID}
	view, viewErr := service.GetView(ctx, viewParams)
	s.Require().NoError(viewErr)
	s.Len(view.Entities, 2)
	s.Len(view.Relationships, 1)
	s.Equal(source.ID, view.RootID)
}
