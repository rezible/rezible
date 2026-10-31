package db

import (
	"context"
	"strings"
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
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/test"
)

type KnowledgeGraphIngestionServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphIngestionServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphIngestionServiceSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeGraphIngestionServiceSuite) makeService(tdb rez.Database) *KnowledgeGraphIngestionService {
	service, serviceErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(serviceErr)
	return service
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
	s.Require().ErrorIs(conflictErr, errs.ErrConflict)
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
		Predicate:           knr.PredicateOwns,
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
		Predicate:           knr.PredicateDependsOn,
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
		Predicate:           knr.PredicateDependsOn,
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

func (s *KnowledgeGraphIngestionServiceSuite) TestEntityStateKeepsNewerFieldsAgainstOlderEvidence() {
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
	s.Equal("Old description", entity.State.Description, "only the older evidence states a description")
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

func (s *KnowledgeGraphIngestionServiceSuite) TestIngestRejectsInvalidRelationship() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.makeService(tdb)
	now := time.Now().UTC()
	entity := func(name string, category kne.Category, kind string) rez.KnowledgeEntityRef {
		return rez.KnowledgeEntityRef{Category: category, Kind: kind, ProviderResourceRef: s.testProviderRef("invalid:" + name)}
	}
	system := entity("system", kne.CategorySystem, "subsystem")
	cluster := entity("cluster", kne.CategoryInfrastructure, "kubernetes_cluster")
	team := entity("team", kne.CategoryActor, "team")
	service1 := entity("service", kne.CategoryContainer, "service")

	tests := []struct {
		name      string
		predicate knr.Predicate
		source    rez.KnowledgeEntityRef
		target    rez.KnowledgeEntityRef
	}{
		{"cross-hierarchy contains", knr.PredicateContains, system, cluster},
		{"runs_on from a team", knr.PredicateRunsOn, team, cluster},
		{"observes from a container", knr.PredicateObserves, service1, system},
		{"self contains", knr.PredicateContains, system, system},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			relationshipRef := rez.KnowledgeRelationshipRef{
				Predicate:           tc.predicate,
				ProviderResourceRef: s.testProviderRef("invalid:relationship:" + tc.name),
				Source:              tc.source,
				Target:              tc.target,
			}
			observe := func(subject rez.KnowledgeSubjectRef) rez.KnowledgeEvidenceRef {
				return rez.KnowledgeEvidenceRef{Kind: ke.KindObserved, Assertion: "observed", EffectiveAt: now, Subject: subject}
			}
			evidence := []rez.KnowledgeEvidenceRef{
				observe(rez.KnowledgeSubjectRef{Entity: &tc.source}),
				observe(rez.KnowledgeSubjectRef{Relationship: &relationshipRef}),
			}
			if tc.target != tc.source {
				evidence = append(evidence, observe(rez.KnowledgeSubjectRef{Entity: &tc.target}))
			}
			ingestErr := service.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:"+tc.name, now), evidence...)
			s.ErrorIs(ingestErr, errs.ErrInvalidInput)
			s.Zero(tdb.Client(ctx).KnowledgeSubjectAlias.Query().Where(ksa.ProviderResourceRefHasPrefix("invalid:")).CountX(ctx))
		})
	}
}

// observeService ingests one observation of the checkout service through a source's alias, linked by name.
func (s *KnowledgeGraphIngestionServiceSuite) observeService(ctx context.Context, tdb rez.Database, source string, evidence rez.KnowledgeEvidenceRef) *ent.KnowledgeEntity {
	ref := &rez.KnowledgeEntityRef{
		Category:            kne.CategoryContainer,
		Kind:                "service",
		ProviderResourceRef: s.testProviderRef(source + ":checkout"),
		LinkingAttributes:   testEntityLinkingAttributes{"service.name": "checkout"},
	}
	evidence.Assertion = "service_observed"
	evidence.Subject = rez.KnowledgeSubjectRef{Entity: ref}
	if evidence.Kind == "" {
		evidence.Kind = ke.KindObserved
	}
	event := s.createEvent(ctx, tdb, "event:"+uuid.NewString(), evidence.EffectiveAt)
	s.Require().NoError(s.makeService(tdb).IngestEvidence(ctx, event, evidence))
	entityQuery := tdb.Client(ctx).KnowledgeEntity.Query().
		Where(kne.HasAliasesWith(ksa.ProviderResourceRef(ref.ProviderResourceRef.ResourceRef)))
	entity, queryErr := entityQuery.Only(ctx)
	s.Require().NoError(queryErr)
	return entity
}

func (s *KnowledgeGraphIngestionServiceSuite) TestSubjectStateMerge() {
	ctx, tdb := s.SetupTestDatabase()
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

	s.observeService(ctx, tdb, "grafana", rez.KnowledgeEvidenceRef{
		EffectiveAt:  day(6),
		ValidFor:     3 * time.Hour,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	})
	s.observeService(ctx, tdb, "catalogue", rez.KnowledgeEvidenceRef{
		EffectiveAt:  day(1),
		StateMode:    ke.StateModeReplace,
		SubjectState: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core", "lifecycle": "production", "owner": "commerce"}},
	})
	s.observeService(ctx, tdb, "alerts", rez.KnowledgeEvidenceRef{
		EffectiveAt:  day(3),
		SubjectState: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "critical", "runbook": "https://runbooks.test/checkout"}},
	})
	entity := s.observeService(ctx, tdb, "catalogue", rez.KnowledgeEvidenceRef{
		EffectiveAt:  day(6),
		StateMode:    ke.StateModeReplace,
		SubjectState: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core", "owner": "commerce"}},
	})

	s.Equal("checkout", entity.State.DisplayName)
	s.Equal(map[string]any{"tier": "core", "owner": "commerce", "runbook": "https://runbooks.test/checkout"}, entity.State.Properties,
		"the snapshot unsets only its own lifecycle, and the alert's runbook survives")
	s.Nil(entity.ExpiresAt, "the catalogue alias never expires")
	s.True(day(6).Equal(*entity.LastObservedAt))
}

func (s *KnowledgeGraphIngestionServiceSuite) TestWithdrawal() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.makeService(tdb)
	query, queryServiceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queryServiceErr)
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	team := rez.KnowledgeEntityRef{Category: kne.CategoryActor, Kind: "team", ProviderResourceRef: s.testProviderRef("team:search")}
	user := rez.KnowledgeEntityRef{Category: kne.CategoryActor, Kind: "user", ProviderResourceRef: s.testProviderRef("user:ava")}
	membership := rez.KnowledgeRelationshipRef{Predicate: knr.PredicateMemberOf, ProviderResourceRef: s.testProviderRef("membership:ava"), Source: user, Target: team}
	evidence := func(kind ke.Kind, at time.Time, subject rez.KnowledgeSubjectRef) rez.KnowledgeEvidenceRef {
		return rez.KnowledgeEvidenceRef{Kind: kind, Assertion: "observed", EffectiveAt: at, Subject: subject,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Search"}}
	}
	ingest := func(at time.Time, refs ...rez.KnowledgeEvidenceRef) {
		s.Require().NoError(service.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:"+uuid.NewString(), at), refs...))
	}
	ingest(at,
		evidence(ke.KindObserved, at, rez.KnowledgeSubjectRef{Entity: &team}),
		evidence(ke.KindObserved, at, rez.KnowledgeSubjectRef{Entity: &user}),
		evidence(ke.KindObserved, at, rez.KnowledgeSubjectRef{Relationship: &membership}),
	)
	listed := func(includeWithdrawn bool) (int, int) {
		entities, entitiesErr := query.ListEntities(ctx, rez.ListKnowledgeEntitiesParams{IncludeWithdrawn: includeWithdrawn})
		s.Require().NoError(entitiesErr)
		relationships, relationshipsErr := query.ListRelationships(ctx, rez.ListKnowledgeRelationshipsParams{IncludeWithdrawn: includeWithdrawn})
		s.Require().NoError(relationshipsErr)
		return len(entities.Data), len(relationships.Data)
	}

	ingest(at.Add(time.Hour), evidence(ke.KindDeleted, at.Add(time.Hour), rez.KnowledgeSubjectRef{Entity: &team}))
	entities, relationships := listed(false)
	s.Equal(1, entities, "the withdrawn team leaves default reads")
	s.Zero(relationships, "and so does the membership naming it")
	entities, relationships = listed(true)
	s.Equal(2, entities)
	s.Equal(1, relationships)

	ingest(at.Add(2*time.Hour), evidence(ke.KindObserved, at.Add(2*time.Hour), rez.KnowledgeSubjectRef{Entity: &team}))
	entities, relationships = listed(false)
	s.Equal(2, entities, "a later observation restores the team")
	s.Equal(1, relationships)
}

func (s *KnowledgeGraphIngestionServiceSuite) TestExpiry() {
	ctx, tdb := s.SetupTestDatabase()
	synced := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)

	entity := s.observeService(ctx, tdb, "grafana", rez.KnowledgeEvidenceRef{
		EffectiveAt:  synced,
		ValidFor:     3 * time.Hour,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	})
	s.False(knowledgegraph.IsStale(entity.ExpiresAt, synced.Add(3*time.Hour)))
	s.True(knowledgegraph.IsStale(entity.ExpiresAt, synced.Add(3*time.Hour+time.Minute)))

	entity = s.observeService(ctx, tdb, "alerts", rez.KnowledgeEvidenceRef{EffectiveAt: synced})
	s.False(knowledgegraph.IsStale(entity.ExpiresAt, synced.Add(24*time.Hour)), "the alert-observed alias never expires")
}

// parentFixture ingests logical entities and contains claims between them, each in its own event, in order.
type parentFixture struct {
	s       *KnowledgeGraphIngestionServiceSuite
	ctx     context.Context
	tdb     rez.Database
	at      time.Time
	ids     map[string]uuid.UUID
	queries *KnowledgeGraphQueryService
}

func (s *KnowledgeGraphIngestionServiceSuite) newParentFixture() *parentFixture {
	ctx, tdb := s.SetupTestDatabase()
	queries, queriesErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queriesErr)
	return &parentFixture{s: s, ctx: ctx, tdb: tdb, at: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), ids: make(map[string]uuid.UUID), queries: queries}
}

func (f *parentFixture) entityRef(name string) rez.KnowledgeEntityRef {
	category := kne.CategorySystem
	if strings.HasPrefix(name, "service") {
		category = kne.CategoryContainer
	}
	return rez.KnowledgeEntityRef{Category: category, Kind: "test", ProviderResourceRef: f.s.testProviderRef("parent:" + name)}
}

func (f *parentFixture) ingest(refs ...rez.KnowledgeEvidenceRef) {
	f.at = f.at.Add(time.Hour)
	for i := range refs {
		refs[i].Assertion, refs[i].EffectiveAt = "observed", f.at
	}
	event := f.s.createEvent(f.ctx, f.tdb, "event:"+uuid.NewString(), f.at)
	f.s.Require().NoError(f.s.makeService(f.tdb).IngestEvidence(f.ctx, event, refs...))
}

func (f *parentFixture) entity(kind ke.Kind, name string) {
	ref := f.entityRef(name)
	f.ingest(rez.KnowledgeEvidenceRef{Kind: kind, Subject: rez.KnowledgeSubjectRef{Entity: &ref}})
	entityQuery := f.tdb.Client(f.ctx).KnowledgeEntity.Query().
		Where(kne.HasAliasesWith(ksa.ProviderResourceRef(ref.ProviderResourceRef.ResourceRef)))
	f.ids[name] = entityQuery.OnlyIDX(f.ctx)
}

func (f *parentFixture) contains(kind ke.Kind, parent, child string) {
	ref := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateContains,
		ProviderResourceRef: f.s.testProviderRef("parent:contains:" + parent + ":" + child),
		Source:              f.entityRef(parent),
		Target:              f.entityRef(child),
	}
	f.ingest(rez.KnowledgeEvidenceRef{Kind: kind, Subject: rez.KnowledgeSubjectRef{Relationship: &ref}})
}

func (f *parentFixture) parentOf(name string) string {
	parentID := f.tdb.Client(f.ctx).KnowledgeEntity.GetX(f.ctx, f.ids[name]).ParentID
	for candidate, id := range f.ids {
		if parentID != nil && id == *parentID {
			return candidate
		}
	}
	return ""
}

func (s *KnowledgeGraphIngestionServiceSuite) TestParentChoice() {
	s.Run("the first claim wins and the others are other claims", func() {
		f := s.newParentFixture()
		f.entity(ke.KindObserved, "commerce")
		f.entity(ke.KindObserved, "experience")
		f.entity(ke.KindObserved, "service-search")
		f.contains(ke.KindObserved, "commerce", "service-search")
		f.contains(ke.KindObserved, "experience", "service-search")
		s.Equal("commerce", f.parentOf("service-search"))

		claims, claimsErr := f.queries.ListParentClaims(f.ctx, []uuid.UUID{f.ids["service-search"]})
		s.Require().NoError(claimsErr)
		s.Require().Len(claims[f.ids["service-search"]], 1)
		s.Equal(f.ids["experience"], claims[f.ids["service-search"]][0].SourceEntityID)

		f.contains(ke.KindDeleted, "commerce", "service-search")
		s.Equal("experience", f.parentOf("service-search"), "a withdrawn claim stops being a candidate")
	})

	s.Run("a withdrawn parent falls back to the next claim, and back when restored", func() {
		f := s.newParentFixture()
		f.entity(ke.KindObserved, "commerce")
		f.entity(ke.KindObserved, "experience")
		f.entity(ke.KindObserved, "service-search")
		f.contains(ke.KindObserved, "commerce", "service-search")
		f.contains(ke.KindObserved, "experience", "service-search")

		f.entity(ke.KindDeleted, "commerce")
		s.Equal("experience", f.parentOf("service-search"))
		f.entity(ke.KindObserved, "commerce")
		s.Equal("commerce", f.parentOf("service-search"))
	})

	s.Run("a claim that would make a cycle is skipped", func() {
		f := s.newParentFixture()
		f.entity(ke.KindObserved, "service-a")
		f.entity(ke.KindObserved, "service-b")
		f.contains(ke.KindObserved, "service-a", "service-b")
		f.contains(ke.KindObserved, "service-b", "service-a")
		s.Equal("service-a", f.parentOf("service-b"))
		s.Empty(f.parentOf("service-a"))
	})
}
