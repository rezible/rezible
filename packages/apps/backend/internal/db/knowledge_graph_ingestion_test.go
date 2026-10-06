package db

import (
	"context"
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

type KnowledgeGraphIngestionServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphIngestionServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphIngestionServiceSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeGraphIngestionServiceSuite) makeService(tdb rez.Database) *KnowledgeGraphIngestionService {
	return &KnowledgeGraphIngestionService{db: tdb}
}

func (s *KnowledgeGraphIngestionServiceSuite) createEvent(ctx context.Context, tdb rez.Database, resourceRef string, occurredAt time.Time) *ent.NormalizedEvent {
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

func (s *KnowledgeGraphIngestionServiceSuite) testProviderRef(resourceRef string) rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          "test",
		ProviderNamespace: "knowledge-graph-tests",
		ResourceRef:       resourceRef,
	}
}

func (s *KnowledgeGraphIngestionServiceSuite) TestKnowledgeSubjectAliasCannotMapOneResourceToDifferentSubjects() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	service := s.makeService(tdb)

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
	event := s.createEvent(ctx, tdb, "event:shared-resource", evidence.EffectiveAt)
	s.Require().Error(secondAliasErr)
	err := service.IngestEvidence(ctx, event, evidence)
	s.Require().NoError(err)
	evidenceQuery := client.KnowledgeEvidence.Query()
	s.Equal(1, evidenceQuery.CountX(ctx))
}

type testEntityLinkingAttributes map[string]string

func (a testEntityLinkingAttributes) Values() map[string]string {
	return a
}

func (s *KnowledgeGraphIngestionServiceSuite) TestEntityAliasesMatchLinkingAttributes() {
	ctx, tdb := s.SetupTestDatabase()
	svc := s.makeService(tdb)
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
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:first", now), firstEvidence))
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:second", now.Add(time.Minute)), secondEvidence))

	client := tdb.Client(ctx)
	s.Equal(1, client.KnowledgeEntity.Query().Where(kne.CategoryEQ(kne.CategoryActor), kne.KindEQ("user")).CountX(ctx))
	s.Equal(2, client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(1, client.KnowledgeEntityLinkingAttribute.Query().CountX(ctx))
	s.Equal(2, client.KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestEntityLinkingAttributeConflictRollsBackEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	svc := s.makeService(tdb)
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
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:first", now), makeEvidence(&firstRef)))
	s.Require().NoError(svc.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:second", now), makeEvidence(&secondRef)))

	conflictRef := makeRef("third", map[string]string{
		"user.email":       "first@example.com",
		"user.employee_id": "second",
	})
	conflictEvent := s.createEvent(ctx, tdb, "event:third", now.Add(time.Minute))
	conflictErr := svc.IngestEvidence(ctx, conflictEvent, makeEvidence(&conflictRef))
	s.Require().ErrorIs(conflictErr, rez.ErrConflict)
	s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))
	s.Equal(2, tdb.Client(ctx).KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(2, tdb.Client(ctx).KnowledgeEntityLinkingAttribute.Query().CountX(ctx))
	s.Equal(2, tdb.Client(ctx).KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestRelationshipAliasesConvergeThroughLinkedEndpointEntities() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.makeService(tdb)
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
	s.Require().NoError(service.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:relationship:one", now),
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
	s.Require().NoError(service.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:relationship:two", now.Add(time.Minute)),
		entityEvidence(&secondSourceRef, "Alice"),
		relationshipEvidence(&secondRelationshipRef),
	))

	s.Equal(2, tdb.Client(ctx).KnowledgeEntity.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).KnowledgeRelationship.Query().CountX(ctx))
	s.Equal(5, tdb.Client(ctx).KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(5, tdb.Client(ctx).KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestRelationshipIngestionRollsBackWhenEndpointHasNoEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	service := s.makeService(tdb)
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
	event := s.createEvent(ctx, tdb, relationshipRef.ProviderResourceRef.ResourceRef, now)

	ingestErr := service.IngestEvidence(ctx, event, relationshipEvidence, sourceEvidence)
	s.Require().Error(ingestErr)
	s.ErrorContains(ingestErr, errNoExistingEndpointEntityAlias.Error())
	s.Zero(client.KnowledgeEntity.Query().CountX(ctx))
	s.Zero(client.KnowledgeRelationship.Query().CountX(ctx))
	s.Zero(client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Zero(client.KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestRelationshipIngestionEvidenceBacksEndpointsRegardlessOfInputOrder() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	service := s.makeService(tdb)
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
	event := s.createEvent(ctx, tdb, relationshipRef.ProviderResourceRef.ResourceRef, now)

	ingestErr := service.IngestEvidence(ctx, event, relationshipEvidence, targetEvidence, sourceEvidence)
	s.Require().NoError(ingestErr)
	s.Equal(2, client.KnowledgeEntity.Query().CountX(ctx))
	s.Equal(1, client.KnowledgeRelationship.Query().CountX(ctx))
	s.Equal(3, client.KnowledgeSubjectAlias.Query().CountX(ctx))
	s.Equal(3, client.KnowledgeEvidence.Query().CountX(ctx))
}

func (s *KnowledgeGraphIngestionServiceSuite) ingestEntityState(ctx context.Context, tdb rez.Database, ref *rez.KnowledgeEntityRef, effectiveAt time.Time, state schematypes.KnowledgeGraphSubjectState) *ent.KnowledgeEntity {
	evidence := rez.KnowledgeEvidenceRef{
		Kind:         ke.KindObserved,
		Assertion:    "service_observed",
		EffectiveAt:  effectiveAt,
		SubjectState: state,
		Subject:      rez.KnowledgeSubjectRef{Entity: ref},
	}
	event := s.createEvent(ctx, tdb, "event:"+uuid.NewString(), effectiveAt)
	s.Require().NoError(s.makeService(tdb).IngestEvidence(ctx, event, evidence))

	entityQuery := tdb.Client(ctx).KnowledgeEntity.Query().
		Where(kne.HasAliasesWith(ksa.ProviderResourceRef(ref.ProviderResourceRef.ResourceRef)))
	entity, queryErr := entityQuery.Only(ctx)
	s.Require().NoError(queryErr)
	return entity
}

func (s *KnowledgeGraphIngestionServiceSuite) serviceEntityRef() *rez.KnowledgeEntityRef {
	return &rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef("service:checkout"),
	}
}

func (s *KnowledgeGraphIngestionServiceSuite) TestEntityStateStoresNewestObservedEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	ref := s.serviceEntityRef()
	older := time.Now().UTC().Truncate(time.Microsecond)
	newer := older.Add(time.Minute)

	s.ingestEntityState(ctx, tdb, ref, older, schematypes.KnowledgeGraphSubjectState{
		DisplayName: "checkout",
		Description: "Takes payments",
		Properties:  map[string]any{"tier": "2"},
	})
	entity := s.ingestEntityState(ctx, tdb, ref, newer, schematypes.KnowledgeGraphSubjectState{
		DisplayName: "Checkout",
		Properties:  map[string]any{"tier": "1"},
	})

	s.Equal("Checkout", entity.State.DisplayName)
	s.Equal("Takes payments", entity.State.Description)
	s.Equal(map[string]any{"tier": "1"}, entity.State.Properties)
	s.Require().NotNil(entity.StateEffectiveAt)
	s.True(newer.Equal(*entity.StateEffectiveAt))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestEntityStateIgnoresOlderEvidenceArrivingLater() {
	ctx, tdb := s.SetupTestDatabase()
	ref := s.serviceEntityRef()
	newer := time.Now().UTC().Truncate(time.Microsecond)
	older := newer.Add(-time.Minute)

	s.ingestEntityState(ctx, tdb, ref, newer, schematypes.KnowledgeGraphSubjectState{
		DisplayName: "Checkout",
		Properties:  map[string]any{"tier": "1"},
	})
	entity := s.ingestEntityState(ctx, tdb, ref, older, schematypes.KnowledgeGraphSubjectState{
		DisplayName: "checkout-legacy",
		Description: "Old description",
		Properties:  map[string]any{"tier": "3"},
	})

	s.Equal("Checkout", entity.State.DisplayName)
	s.Empty(entity.State.Description)
	s.Equal(map[string]any{"tier": "1"}, entity.State.Properties)
	s.Require().NotNil(entity.StateEffectiveAt)
	s.True(newer.Equal(*entity.StateEffectiveAt))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestEntityStateEmptyEvidenceDoesNotBlockOlderEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	ref := s.serviceEntityRef()
	older := time.Now().UTC().Truncate(time.Microsecond)
	newer := older.Add(time.Minute)

	s.ingestEntityState(ctx, tdb, ref, newer, schematypes.KnowledgeGraphSubjectState{})
	entity := s.ingestEntityState(ctx, tdb, ref, older, schematypes.KnowledgeGraphSubjectState{
		DisplayName: "Checkout",
	})

	s.Equal("Checkout", entity.State.DisplayName)
	s.Require().NotNil(entity.StateEffectiveAt)
	s.True(older.Equal(*entity.StateEffectiveAt))
}
