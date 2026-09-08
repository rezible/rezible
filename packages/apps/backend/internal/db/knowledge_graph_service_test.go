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

type testEntityLinkingAttributes map[string]string

func (a testEntityLinkingAttributes) Values() map[string]string {
	return a
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

func (s *KnowledgeGraphServiceSuite) testProviderRef(resourceRef string) rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
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
	apiEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: apiRef,
	}
	apiEvent := s.createEvent(tdb, apiRef.ResourceRef, now.Add(-2*time.Hour))
	oldEvidence := rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now.Add(-2 * time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Old API name",
		},
		Subject: rez.KnowledgeSubjectRef{Entity: &apiEntityRef},
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
	databaseEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "database",
		ProviderResourceRef: databaseRef,
	}
	databaseEvidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "component_exists",
		EffectiveAt:  now.Add(-time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Database"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &databaseEntityRef},
	}
	databaseEvent := s.createEvent(tdb, databaseRef.ResourceRef, now.Add(-time.Hour))
	s.Require().NoError(service.IngestEvidence(ctx, databaseEvent, databaseEvidence))

	relRef := s.testProviderRef("api-uses-database")
	relationshipRef := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: relRef,
		Source:              apiEntityRef,
		Target:              databaseEntityRef,
	}
	relEvidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "relationship_exists",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "uses"},
		Subject:      rez.KnowledgeSubjectRef{Relationship: &relationshipRef},
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
	entityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: ref,
	}
	evidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "component_exists",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &entityRef},
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
	entityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                "user",
		ProviderResourceRef: sharedRef,
	}
	evidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "user_observed",
		EffectiveAt:  time.Now().UTC(),
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Second"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &entityRef},
	}
	event := s.createEvent(tdb, "event:shared-resource", evidence.EffectiveAt)
	s.Require().Error(secondAliasErr)
	err := service.IngestEvidence(ctx, event, evidence)
	s.Require().NoError(err)
	evidenceQuery := client.KnowledgeEvidence.Query()
	s.Equal(1, evidenceQuery.CountX(ctx))
}

func (s *KnowledgeGraphServiceSuite) TestEntityAliasesMatchLinkingAttributes() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	svc := s.knowledgeService(tdb)
	now := time.Now().UTC()

	firstRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                "user",
		ProviderResourceRef: s.testProviderRef("slack-user"),
		LinkingAttributes: testEntityLinkingAttributes{
			"user.email": "alice@example.com",
		},
	}
	secondRef := rez.KnowledgeEntityRef{
		Category: kne.CategoryActor,
		Kind:     "user",
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "github",
			ProviderNamespace: "acme",
			ResourceRef:       "42",
		},
		LinkingAttributes: testEntityLinkingAttributes{
			"user.email": "alice@example.com",
		},
	}
	firstEvidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "user_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Alice"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &firstRef},
	}
	secondEvidence := firstEvidence
	secondEvidence.Subject.Entity = &secondRef
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(tdb, "event:first", now), firstEvidence))
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(tdb, "event:second", now.Add(time.Minute)), secondEvidence))

	client := tdb.Client(ctx)
	s.Equal(1, client.KnowledgeEntity.Query().Where(kne.CategoryEQ(kne.CategoryActor), kne.KindEQ("user")).CountX(ctx))
	s.Equal(2, client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(1, client.KnowledgeEntityLinkingAttribute.Query().CountX(ctx))
	s.Equal(2, client.KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphServiceSuite) TestEntityLinkingAttributeConflictRollsBackEvidence() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	svc := s.knowledgeService(tdb)
	now := time.Now().UTC()
	makeRef := func(resourceRef string, values map[string]string) rez.KnowledgeEntityRef {
		return rez.KnowledgeEntityRef{
			Category:            kne.CategoryActor,
			Kind:                "user",
			ProviderResourceRef: s.testProviderRef(resourceRef),
			LinkingAttributes:   testEntityLinkingAttributes(values),
		}
	}
	makeEvidence := func(ref *rez.KnowledgeEntityRef) rez.KnowledgeEvidenceRef {
		return rez.KnowledgeEvidenceRef{
			Kind:         ke.KindObserved,
			Assertion:    "user_observed",
			EffectiveAt:  now,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "User"},
			Subject:      rez.KnowledgeSubjectRef{Entity: ref},
		}
	}
	firstRef := makeRef("first", map[string]string{"user.email": "first@example.com"})
	secondRef := makeRef("second", map[string]string{"user.employee_id": "second"})
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(tdb, "event:first", now), makeEvidence(&firstRef)))
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(tdb, "event:second", now), makeEvidence(&secondRef)))

	conflictRef := makeRef("third", map[string]string{
		"user.email":       "first@example.com",
		"user.employee_id": "second",
	})
	conflictEvent := s.createEvent(tdb, "event:third", now.Add(time.Minute))
	conflictErr := svc.IngestEvidence(ctx, conflictEvent, makeEvidence(&conflictRef))
	s.Require().ErrorIs(conflictErr, rez.ErrConflict)
	s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))
	s.Equal(2, tdb.Client(ctx).KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(2, tdb.Client(ctx).KnowledgeEntityLinkingAttribute.Query().CountX(ctx))
	s.Equal(2, tdb.Client(ctx).KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphServiceSuite) TestRelationshipAliasesConvergeThroughLinkedEndpointEntities() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()
	targetRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef("service:api"),
	}
	firstSourceRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryActor,
		Kind:                "user",
		ProviderResourceRef: s.testProviderRef("slack:alice"),
		LinkingAttributes:   testEntityLinkingAttributes{"user.email": "alice@example.com"},
	}
	firstRelationshipRef := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: s.testProviderRef("relationship:one"),
		Source:              firstSourceRef,
		Target:              targetRef,
	}
	entityEvidence := func(ref *rez.KnowledgeEntityRef, name string) rez.KnowledgeEvidenceRef {
		return rez.KnowledgeEvidenceRef{
			Kind:         ke.KindObserved,
			Assertion:    "entity_observed",
			EffectiveAt:  now,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: name},
			Subject:      rez.KnowledgeSubjectRef{Entity: ref},
		}
	}
	relationshipEvidence := func(ref *rez.KnowledgeRelationshipRef) rez.KnowledgeEvidenceRef {
		return rez.KnowledgeEvidenceRef{
			Kind:        ke.KindObserved,
			Assertion:   "relationship_observed",
			EffectiveAt: now,
			Subject:     rez.KnowledgeSubjectRef{Relationship: ref},
		}
	}
	s.Require().NoError(service.IngestEvidence(ctx, s.createEvent(tdb, "event:relationship:one", now),
		entityEvidence(&firstSourceRef, "Alice"),
		entityEvidence(&targetRef, "API"),
		relationshipEvidence(&firstRelationshipRef),
	))

	secondSourceRef := firstSourceRef
	secondSourceRef.ProviderResourceRef = rez.ProviderResourceRef{
		Provider:          "github",
		ProviderNamespace: "acme",
		ResourceRef:       "42",
	}
	secondRelationshipRef := firstRelationshipRef
	secondRelationshipRef.ProviderResourceRef = s.testProviderRef("relationship:two")
	secondRelationshipRef.Source = secondSourceRef
	s.Require().NoError(service.IngestEvidence(ctx, s.createEvent(tdb, "event:relationship:two", now.Add(time.Minute)),
		entityEvidence(&secondSourceRef, "Alice"),
		relationshipEvidence(&secondRelationshipRef),
	))

	s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).KnowledgeRelationship.Query().CountX(ctx))
	s.Equal(5, tdb.Client(ctx).KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(5, tdb.Client(ctx).KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphServiceSuite) TestRelationshipIngestionRollsBackWhenEndpointHasNoEvidence() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()

	sourceRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef("service:api"),
	}
	targetRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "database",
		ProviderResourceRef: s.testProviderRef("database:primary"),
	}
	relationshipRef := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: s.testProviderRef("api-uses-primary"),
		Source:              sourceRef,
		Target:              targetRef,
	}
	sourceEvidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "source_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &sourceRef},
	}
	relationshipEvidence := rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "relationship_observed",
		EffectiveAt: now,
		Subject:     rez.KnowledgeSubjectRef{Relationship: &relationshipRef},
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

	sourceRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef("service:api"),
	}
	targetRef := rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "database",
		ProviderResourceRef: s.testProviderRef("database:primary"),
	}
	relationshipRef := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateUses,
		ProviderResourceRef: s.testProviderRef("api-uses-primary"),
		Source:              sourceRef,
		Target:              targetRef,
	}
	sourceEvidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "source_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &sourceRef},
	}
	targetEvidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "target_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Primary database"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &targetRef},
	}
	relationshipEvidence := rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "relationship_observed",
		EffectiveAt: now,
		Subject:     rez.KnowledgeSubjectRef{Relationship: &relationshipRef},
	}
	event := s.createEvent(tdb, relationshipRef.ProviderResourceRef.ResourceRef, now)

	ingestErr := service.IngestEvidence(ctx, event, relationshipEvidence, targetEvidence, sourceEvidence)
	s.Require().NoError(ingestErr)
	s.Equal(2, client.KnowledgeEntity.Query().CountX(ctx))
	s.Equal(1, client.KnowledgeRelationship.Query().CountX(ctx))
	s.Equal(3, client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(3, client.KnowledgeEvidence.Query().CountX(ctx))
}
