package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/test"
)

type KnowledgeGraphServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphServiceSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeGraphServiceSuite) knowledgeService(tdb rez.Database) *KnowledgeGraphService {
	return &KnowledgeGraphService{db: tdb}
}

func (s *KnowledgeGraphServiceSuite) createEvent(tdb rez.Database, resourceRef string, occurredAt time.Time) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	event, err := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("knowledge-graph-tests").
		SetProviderResourceRef(resourceRef).
		SetProviderEventSource("knowledge-graph-tests").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetKind("system_component").
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes([]byte(`{}`)).
		Save(ctx)
	s.Require().NoError(err)
	return event
}

func (s *KnowledgeGraphServiceSuite) testProviderRef(resourceRef string) ent.ProviderResourceRef {
	return ent.ProviderResourceRef{
		Provider:          "test",
		ProviderNamespace: "knowledge-graph-tests",
		ResourceRef:       resourceRef,
	}
}

func (s *KnowledgeGraphServiceSuite) TestCurrentStateAndBoundedView() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()

	apiRef := s.testProviderRef("service:api")
	apiEntityRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: apiRef,
	}
	apiEvent := s.createEvent(tdb, apiRef.ResourceRef, now.Add(-2*time.Hour))
	oldEvidence := ent.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now.Add(-2 * time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Old API name",
		},
		SubjectEntity: &apiEntityRef,
	}
	s.Require().NoError(service.IngestEvidence(ctx, apiEvent, oldEvidence))
	apiAliasQuery := tdb.Client(ctx).KnowledgeSubjectAlias.Query()
	apiAliasQuery.Where(ksa.ProviderResourceRef(apiRef.ResourceRef))
	apiAlias := apiAliasQuery.OnlyX(ctx)
	apiEntityID := *apiAlias.EntityID

	latestEvent := s.createEvent(tdb, apiRef.ResourceRef, now.Add(-time.Hour))
	latestEvidence := oldEvidence
	latestEvidence.EffectiveAt = now.Add(-time.Hour)
	latestEvidence.SubjectState.DisplayName = "API"
	s.Require().NoError(service.IngestEvidence(ctx, latestEvent, latestEvidence))

	databaseRef := s.testProviderRef("service:database")
	databaseEntityRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "database",
		ProviderResourceRef: databaseRef,
	}
	databaseEvidence := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "component_exists",
		EffectiveAt:   now.Add(-time.Hour),
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "Database"},
		SubjectEntity: &databaseEntityRef,
	}
	databaseEvent := s.createEvent(tdb, databaseRef.ResourceRef, now.Add(-time.Hour))
	s.Require().NoError(service.IngestEvidence(ctx, databaseEvent, databaseEvidence))

	relRef := s.testProviderRef("api-uses-database")
	relationshipRef := ent.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: relRef,
		Source:              apiEntityRef,
		Target:              databaseEntityRef,
	}
	relEvidence := ent.KnowledgeEvidenceRef{
		Kind:                ke.KindObserved,
		Assertion:           "relationship_exists",
		EffectiveAt:         now,
		SubjectState:        schematypes.KnowledgeGraphSubjectState{DisplayName: "uses"},
		SubjectRelationship: &relationshipRef,
	}
	relEvent := s.createEvent(tdb, relRef.ResourceRef, now)
	s.Require().NoError(service.IngestEvidence(ctx, relEvent, relEvidence))

	apiEntity, err := service.GetEntity(ctx, apiEntityID)
	s.Require().NoError(err)
	current := apiEntity.LatestEvidence()
	s.Require().NotNil(current)
	s.Equal("API", current.SubjectState.DisplayName)

	viewParams := rez.GetKnowledgeGraphViewParams{Depth: 1, EntityID: apiEntity.ID}
	view, err := service.GetView(ctx, viewParams)
	s.Require().NoError(err)
	s.Len(view.Entities, 2)
	s.Len(view.Relationships, 1)
	s.Equal(apiEntity.ID, view.RootID)
}

func (s *KnowledgeGraphServiceSuite) TestBoundedViewIncludesIsolatedRootEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()
	ref := s.testProviderRef("service:api")
	entityRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: ref,
	}
	evidence := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "component_exists",
		EffectiveAt:   now,
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		SubjectEntity: &entityRef,
	}
	event := s.createEvent(tdb, ref.ResourceRef, now)
	s.Require().NoError(service.IngestEvidence(ctx, event, evidence))
	aliasQuery := tdb.Client(ctx).KnowledgeSubjectAlias.Query()
	aliasQuery.Where(ksa.ProviderResourceRef(ref.ResourceRef))
	alias := aliasQuery.OnlyX(ctx)

	viewParams := rez.GetKnowledgeGraphViewParams{Depth: 1, EntityID: *alias.EntityID}
	view, err := service.GetView(ctx, viewParams)
	s.Require().NoError(err)
	s.Equal(*alias.EntityID, view.RootID)
	s.Len(view.Entities, 1)
	s.Empty(view.Relationships)
}

func (s *KnowledgeGraphServiceSuite) TestKnowledgeSubjectAliasCannotMapOneResourceToDifferentSubjects() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	service := s.knowledgeService(tdb)

	createFirst := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryActor).
		SetKind("user")
	first := createFirst.SaveX(ctx)
	createSecond := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryActor).
		SetKind("user")
	second := createSecond.SaveX(ctx)
	createFirstAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderNamespace("knowledge-graph-tests").
		SetProviderResourceRef("shared-resource").
		SetEntityID(first.ID)
	createFirstAlias.SaveX(ctx)
	createSecondAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderNamespace("knowledge-graph-tests").
		SetProviderResourceRef("shared-resource").
		SetEntityID(second.ID)
	_, secondAliasErr := createSecondAlias.Save(ctx)

	sharedRef := s.testProviderRef("shared-resource")
	entityRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                "user",
		ProviderResourceRef: sharedRef,
	}
	evidence := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "user_observed",
		EffectiveAt:   time.Now().UTC(),
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "Second"},
		SubjectEntity: &entityRef,
	}
	event := s.createEvent(tdb, "event:shared-resource", evidence.EffectiveAt)
	s.Require().Error(secondAliasErr)
	err := service.IngestEvidence(ctx, event, evidence)
	s.Require().NoError(err)
	evidenceQuery := client.KnowledgeEvidence.Query()
	s.Equal(1, evidenceQuery.CountX(ctx))
}

func (s *KnowledgeGraphServiceSuite) TestRelationshipIngestionRollsBackWhenEndpointHasNoEvidence() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()

	sourceRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef("service:api"),
	}
	targetRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "database",
		ProviderResourceRef: s.testProviderRef("database:primary"),
	}
	relationshipRef := ent.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: s.testProviderRef("api-uses-primary"),
		Source:              sourceRef,
		Target:              targetRef,
	}
	sourceEvidence := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "source_observed",
		EffectiveAt:   now,
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		SubjectEntity: &sourceRef,
	}
	relationshipEvidence := ent.KnowledgeEvidenceRef{
		Kind:                ke.KindObserved,
		Assertion:           "relationship_observed",
		EffectiveAt:         now,
		SubjectRelationship: &relationshipRef,
	}
	event := s.createEvent(tdb, relationshipRef.ProviderResourceRef.ResourceRef, now)

	ingestErr := service.IngestEvidence(ctx, event, relationshipEvidence, sourceEvidence)
	s.Require().Error(ingestErr)
	s.ErrorContains(ingestErr, errNoExistingEndpointEntityAlias.Error())
	s.Zero(client.KnowledgeEntity.Query().CountX(ctx))
	s.Zero(client.KnowledgeRelationship.Query().CountX(ctx))
	s.Zero(client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Zero(client.KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphServiceSuite) TestRelationshipIngestionEvidenceBacksEndpointsRegardlessOfInputOrder() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()

	sourceRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef("service:api"),
	}
	targetRef := ent.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "database",
		ProviderResourceRef: s.testProviderRef("database:primary"),
	}
	relationshipRef := ent.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: s.testProviderRef("api-uses-primary"),
		Source:              sourceRef,
		Target:              targetRef,
	}
	sourceEvidence := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "source_observed",
		EffectiveAt:   now,
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		SubjectEntity: &sourceRef,
	}
	targetEvidence := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "target_observed",
		EffectiveAt:   now,
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "Primary database"},
		SubjectEntity: &targetRef,
	}
	relationshipEvidence := ent.KnowledgeEvidenceRef{
		Kind:                ke.KindObserved,
		Assertion:           "relationship_observed",
		EffectiveAt:         now,
		SubjectRelationship: &relationshipRef,
	}
	event := s.createEvent(tdb, relationshipRef.ProviderResourceRef.ResourceRef, now)

	ingestErr := service.IngestEvidence(ctx, event, relationshipEvidence, targetEvidence, sourceEvidence)
	s.Require().NoError(ingestErr)
	s.Equal(2, client.KnowledgeEntity.Query().CountX(ctx))
	s.Equal(1, client.KnowledgeRelationship.Query().CountX(ctx))
	s.Equal(3, client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(3, client.KnowledgeEvidence.Query().CountX(ctx))
}
