package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
	v1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/pkg/projections"
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
		SetLastOccurredAt(occurredAt).
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
		SetAuthority(schematypes.KnowledgeAuthorityDeclared).
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderNamespace("knowledge-graph-tests").
		SetProviderResourceRef("shared-resource").
		SetEntityID(first.ID)
	createFirstAlias.SaveX(ctx)
	createSecondAlias := client.KnowledgeSubjectAlias.Create().
		SetAuthority(schematypes.KnowledgeAuthorityDeclared).
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
		Authority:    schematypes.KnowledgeAuthorityDeclared,
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
		Authority:    schematypes.KnowledgeAuthorityDeclared,
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

// A projection whose ref resolves to an entity a merge has redirected writes its evidence and state on the
// survivor. The merged entity still holds a name here, as it does for a projection that read the name before the
// merge moved it.
func (s *KnowledgeGraphIngestionServiceSuite) TestProjectionResolvingAMergedEntityWritesTheSurvivor() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	graph := newTestGraph(s.Require(), tdb)
	at := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	checkout := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:checkout")
	graph.record(ctx, at, graph.observeEntity(checkout, "checkout"))
	survivorID := graph.entityID(ctx, checkout)

	createMerged := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	merged := createMerged.SaveX(ctx)
	createName := client.KnowledgeEntityLinkingAttribute.Create().
		SetEntityID(merged.ID).
		SetAttribute(projections.LinkingAttributeServiceName).
		SetValue("checkout-api")
	createName.ExecX(ctx)
	client.KnowledgeEntity.UpdateOneID(merged.ID).SetMergedIntoID(survivorID).ExecX(ctx)

	renamed := graph.entity(kne.CategoryContainer, "service", "provider-b", "service:checkout-api")
	renamed.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout-api"}
	graph.record(ctx, at.Add(time.Minute), graph.observeEntity(renamed, "checkout-api"))

	alias := graph.alias(ctx, renamed.ProviderResourceRef)
	s.Equal(&survivorID, alias.EntityID)
	s.Equal(1, client.KnowledgeEvidence.Query().Where(ke.SubjectAliasID(alias.ID)).CountX(ctx))
	survivor := client.KnowledgeEntity.GetX(ctx, survivorID)
	s.Require().NotNil(survivor.LastObservedAt)
	s.True(survivor.LastObservedAt.Equal(at.Add(time.Minute)), "the survivor's state takes the new evidence")
	s.Nil(client.KnowledgeEntity.GetX(ent.IncludeMerged(ctx), merged.ID).LastObservedAt, "the merged entity is left as it was")
}

// A merge that commits while a projection waits for its entity lock moves the linking attributes. The projection
// reads them after the lock, so it does not propose its own name as another entity's. An interceptor commits the
// merge at the lock query.
func (s *KnowledgeGraphIngestionServiceSuite) TestMergeBeforeTheEntityLockProposesNoIdentityConflict() {
	cases := []struct {
		name string
		ref  *rez.KnowledgeEntityRef
	}{
		{"the alias names the merged entity", s.proposalServiceRef("provider-a", "service-1", "checkout")},
		{"the alias names the survivor", s.proposalServiceRef("provider-b", "service:payments", "checkout")},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			ctx, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
			client := tdb.Client(ctx)
			s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-a", "service-1", "checkout"))
			s.mustObserve(ctx, tdb, s.proposalServiceRef("provider-b", "service:payments", "payments"))
			mergedID := s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "checkout")
			survivorID := s.entityNamed(ctx, tdb, projections.LinkingAttributeServiceName, "payments")

			type nestedQuery struct{}
			merged := false
			client.Intercept(ent.InterceptFunc(func(next ent.Querier) ent.Querier {
				return ent.QuerierFunc(func(queryCtx context.Context, q ent.Query) (ent.Value, error) {
					if _, isEntity := q.(*ent.KnowledgeEntityQuery); isEntity && !merged && queryCtx.Value(nestedQuery{}) == nil {
						merged = true
						client.KnowledgeSubjectAlias.Update().Where(ksa.EntityID(mergedID)).SetEntityID(survivorID).ExecX(ctx)
						client.KnowledgeEntityLinkingAttribute.Update().Where(kela.EntityID(mergedID)).SetEntityID(survivorID).ExecX(ctx)
						client.KnowledgeEntity.UpdateOneID(mergedID).SetMergedIntoID(survivorID).ExecX(ctx)
					}
					return next.Query(context.WithValue(queryCtx, nestedQuery{}, true), q)
				})
			}))

			s.mustObserve(ctx, tdb, tc.ref)
			s.True(merged)
			s.Empty(s.proposals(ctx, tdb), "the projection's own name is not another entity's")
			alias := client.KnowledgeSubjectAlias.Query().Where(ksa.ProviderResourceRef(tc.ref.ProviderResourceRef.ResourceRef)).OnlyX(ctx)
			s.Equal(&survivorID, alias.EntityID)
		})
	}
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
			Authority:    schematypes.KnowledgeAuthorityDeclared,
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
			Authority:    schematypes.KnowledgeAuthorityDeclared,
			Kind:         ke.KindObserved,
			Assertion:    "entity_observed",
			EffectiveAt:  now,
			SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: name},
			Subject:      rez.KnowledgeSubjectRef{Entity: ref},
		}
	}
	relationshipEvidence := func(ref *rez.KnowledgeRelationshipRef) rez.KnowledgeEvidenceRef {
		return rez.KnowledgeEvidenceRef{
			Authority:   schematypes.KnowledgeAuthorityDeclared,
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
		Authority:    schematypes.KnowledgeAuthorityDeclared,
		Kind:         ke.KindObserved,
		Assertion:    "source_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &sourceRef},
	}
	relationshipEvidence := rez.KnowledgeEvidenceRef{
		Authority:   schematypes.KnowledgeAuthorityDeclared,
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
		Authority:    schematypes.KnowledgeAuthorityDeclared,
		Kind:         ke.KindObserved,
		Assertion:    "source_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "API"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &sourceRef},
	}
	targetEvidence := rez.KnowledgeEvidenceRef{
		Authority:    schematypes.KnowledgeAuthorityDeclared,
		Kind:         ke.KindObserved,
		Assertion:    "target_observed",
		EffectiveAt:  now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "Primary database"},
		Subject:      rez.KnowledgeSubjectRef{Entity: &targetRef},
	}
	relationshipEvidence := rez.KnowledgeEvidenceRef{
		Authority:   schematypes.KnowledgeAuthorityDeclared,
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
		Authority:    schematypes.KnowledgeAuthorityDeclared,
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

func (s *KnowledgeGraphIngestionServiceSuite) TestUnchangedObservationLeavesTheEntityRow() {
	ctx, tdb := s.SetupTestDatabase()
	ref := s.serviceEntityRef()
	observedAt := time.Now().UTC().Truncate(time.Microsecond)
	state := schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout", Properties: map[string]any{"tier": "1"}}

	first := s.ingestEntityState(ctx, tdb, ref, observedAt, state)
	again := s.ingestEntityState(ctx, tdb, ref, observedAt, state)

	s.True(first.UpdatedAt.Equal(again.UpdatedAt), "a second event stating the same state at the same time changes nothing")
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
				return rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityDeclared, Kind: ke.KindObserved, Assertion: "observed", EffectiveAt: now, Subject: subject}
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
			var invalid *invalidRelationshipRefError
			s.True(errors.As(ingestErr, &invalid))
			s.Zero(tdb.Client(ctx).KnowledgeSubjectAlias.Query().Where(ksa.ProviderResourceRefHasPrefix("invalid:")).CountX(ctx))
		})
	}
}

func (s *KnowledgeGraphIngestionServiceSuite) TestEvidenceAuthority() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.makeService(tdb)
	at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ref := s.serviceEntityRef()
	ingest := func(authority schematypes.KnowledgeAuthority) error {
		at = at.Add(time.Hour)
		evidence := rez.KnowledgeEvidenceRef{Authority: authority, Kind: ke.KindObserved, Assertion: "observed", EffectiveAt: at,
			Subject: rez.KnowledgeSubjectRef{Entity: ref}}
		return service.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:"+uuid.NewString(), at), evidence)
	}
	authority := func() schematypes.KnowledgeAuthority {
		return tdb.Client(ctx).KnowledgeSubjectAlias.Query().OnlyX(ctx).Authority
	}

	s.ErrorIs(ingest(""), errs.ErrInvalidInput, "a ref must state its authority")
	s.Zero(tdb.Client(ctx).KnowledgeSubjectAlias.Query().CountX(ctx))

	s.Require().NoError(ingest(schematypes.KnowledgeAuthorityObserved))
	s.Equal(schematypes.KnowledgeAuthorityObserved, authority())
	s.Require().NoError(ingest(schematypes.KnowledgeAuthorityDeclared))
	s.Equal(schematypes.KnowledgeAuthorityDeclared, authority(), "higher evidence raises the alias")
	s.Require().NoError(ingest(schematypes.KnowledgeAuthorityObserved))
	s.Equal(schematypes.KnowledgeAuthorityDeclared, authority(), "lower evidence leaves it")

	// The column stores each authority's rank, so SQL orders and compares authorities in authority order.
	observedOnly := &rez.KnowledgeEntityRef{Category: kne.CategoryContainer, Kind: "service",
		ProviderResourceRef: s.testProviderRef("service:observed-only")}
	at = at.Add(time.Hour)
	observedEvidence := rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityObserved, Kind: ke.KindObserved,
		Assertion: "observed", EffectiveAt: at, Subject: rez.KnowledgeSubjectRef{Entity: observedOnly}}
	s.Require().NoError(service.IngestEvidence(ctx, s.createEvent(ctx, tdb, "event:"+uuid.NewString(), at), observedEvidence))
	byAuthority := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Order(ksa.ByAuthority(sql.OrderDesc())).
		AllX(ctx)
	s.Require().Len(byAuthority, 2)
	s.Equal(schematypes.KnowledgeAuthorityDeclared, byAuthority[0].Authority)
	s.Equal(schematypes.KnowledgeAuthorityObserved, byAuthority[1].Authority)
	aboveObserved := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.AuthorityGT(schematypes.KnowledgeAuthorityObserved))
	s.Equal(byAuthority[0].ID, aboveObserved.OnlyX(ctx).ID)

	// One event names a subject twice: the alias takes the higher authority in either order, though only the
	// first ref is applied as evidence.
	for _, order := range [][2]schematypes.KnowledgeAuthority{
		{schematypes.KnowledgeAuthorityObserved, schematypes.KnowledgeAuthorityDeclared},
		{schematypes.KnowledgeAuthorityDeclared, schematypes.KnowledgeAuthorityObserved},
	} {
		s.Run(string(order[0])+" first", func() {
			ctx, tdb := s.SetupTestDatabase()
			at := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
			ref := &rez.KnowledgeEntityRef{Category: kne.CategoryContainer, Kind: "service",
				ProviderResourceRef: s.testProviderRef("service:twice-" + string(order[0]))}
			refs := make([]rez.KnowledgeEvidenceRef, 0, len(order))
			for _, authority := range order {
				refs = append(refs, rez.KnowledgeEvidenceRef{Authority: authority, Kind: ke.KindObserved, Assertion: "observed",
					EffectiveAt: at, Subject: rez.KnowledgeSubjectRef{Entity: ref}})
			}
			event := s.createEvent(ctx, tdb, "event:"+uuid.NewString(), at)
			s.Require().NoError(s.makeService(tdb).IngestEvidence(ctx, event, refs...))
			queryAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
				Where(ksa.ProviderResourceRef(ref.ProviderResourceRef.ResourceRef))
			s.Equal(schematypes.KnowledgeAuthorityDeclared, queryAlias.OnlyX(ctx).Authority)
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
		Authority:    schematypes.KnowledgeAuthorityObserved,
		EffectiveAt:  day(6),
		ValidFor:     3 * time.Hour,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	})
	s.observeService(ctx, tdb, "catalogue", rez.KnowledgeEvidenceRef{
		Authority:    schematypes.KnowledgeAuthorityDeclared,
		EffectiveAt:  day(1),
		StateMode:    ke.StateModeReplace,
		SubjectState: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "core", "lifecycle": "production", "owner": "commerce"}},
	})
	s.observeService(ctx, tdb, "alerts", rez.KnowledgeEvidenceRef{
		Authority:    schematypes.KnowledgeAuthorityObserved,
		EffectiveAt:  day(3),
		SubjectState: schematypes.KnowledgeGraphSubjectState{Properties: map[string]any{"tier": "critical", "runbook": "https://runbooks.test/checkout"}},
	})
	entity := s.observeService(ctx, tdb, "catalogue", rez.KnowledgeEvidenceRef{
		Authority:    schematypes.KnowledgeAuthorityDeclared,
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
		return rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityDeclared, Kind: kind, Assertion: "observed", EffectiveAt: at, Subject: subject,
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
		Authority:    schematypes.KnowledgeAuthorityObserved,
		EffectiveAt:  synced,
		ValidFor:     3 * time.Hour,
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout"},
	})
	s.False(knowledgegraph.IsStale(entity.ExpiresAt, synced.Add(3*time.Hour)))
	s.True(knowledgegraph.IsStale(entity.ExpiresAt, synced.Add(3*time.Hour+time.Minute)))

	entity = s.observeService(ctx, tdb, "alerts", rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityObserved, EffectiveAt: synced})
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

// prepare creates the next event and stamps the refs with its time, for ingestion. A ref that already has an
// effective time keeps it.
func (f *parentFixture) prepare(refs ...rez.KnowledgeEvidenceRef) *ent.NormalizedEvent {
	f.at = f.at.Add(time.Hour)
	for i := range refs {
		refs[i].Assertion = "observed"
		if refs[i].EffectiveAt.IsZero() {
			refs[i].EffectiveAt = f.at
		}
	}
	return f.s.createEvent(f.ctx, f.tdb, "event:"+uuid.NewString(), f.at)
}

func (f *parentFixture) ingest(refs ...rez.KnowledgeEvidenceRef) {
	event := f.prepare(refs...)
	f.s.Require().NoError(f.s.makeService(f.tdb).IngestEvidence(f.ctx, event, refs...))
	assertPlacement(f.ctx, f.s.T(), f.tdb)
}

func (f *parentFixture) entity(kind ke.Kind, name string) {
	ref := f.entityRef(name)
	f.ingest(rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityDeclared, Kind: kind, Subject: rez.KnowledgeSubjectRef{Entity: &ref}})
	entityQuery := f.tdb.Client(f.ctx).KnowledgeEntity.Query().
		Where(kne.HasAliasesWith(ksa.ProviderResourceRef(ref.ProviderResourceRef.ResourceRef)))
	f.ids[name] = entityQuery.OnlyIDX(f.ctx)
}

func (f *parentFixture) containsRef(kind ke.Kind, parent, child string) rez.KnowledgeEvidenceRef {
	ref := rez.KnowledgeRelationshipRef{
		Predicate:           knr.PredicateContains,
		ProviderResourceRef: f.s.testProviderRef("parent:contains:" + parent + ":" + child),
		Source:              f.entityRef(parent),
		Target:              f.entityRef(child),
	}
	return rez.KnowledgeEvidenceRef{Authority: schematypes.KnowledgeAuthorityDeclared, Kind: kind, Subject: rez.KnowledgeSubjectRef{Relationship: &ref}}
}

func (f *parentFixture) contains(kind ke.Kind, parent, child string) {
	f.ingest(f.containsRef(kind, parent, child))
}

func (f *parentFixture) isFallback(name string) bool {
	return f.tdb.Client(f.ctx).KnowledgeEntity.GetX(f.ctx, f.ids[name]).ParentIsFallback
}

func (f *parentFixture) parentOf(name string) string {
	claim := f.parentClaim(name)
	for candidate, id := range f.ids {
		if claim != nil && id == claim.SourceEntityID {
			return candidate
		}
	}
	return ""
}

// parentClaim returns the claim flagged as the entity's parent, nil when there is none.
func (f *parentFixture) parentClaim(name string) *ent.KnowledgeRelationship {
	queryFlagged := f.tdb.Client(f.ctx).KnowledgeRelationship.Query().
		Where(knr.TargetEntityID(f.ids[name]), knr.IsParent(true))
	flagged := queryFlagged.AllX(f.ctx)
	f.s.Require().LessOrEqual(len(flagged), 1)
	if len(flagged) == 0 {
		return nil
	}
	return flagged[0]
}

// assertPlacement recomputes every entity's parent choice from the committed claims, with
// knowledgegraph.RankParentClaims and the committed parent graph for the cycle check, and asserts that the flagged
// claim is that choice and is current, and that parent_is_fallback is set exactly when a better-ranked claim was
// skipped for closing a cycle. Rows a merge left as redirects are frozen, and queries leave them out.
func assertPlacement(ctx context.Context, t *testing.T, database rez.Database) {
	t.Helper()
	client := database.Client(ctx)
	entities := client.KnowledgeEntity.Query().AllX(ctx)
	queryContains := client.KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateContains)).
		WithAliases(liveParentClaimAliases)
	contains := queryContains.AllX(ctx)

	withdrawn := make(map[uuid.UUID]bool, len(entities))
	for _, entity := range entities {
		withdrawn[entity.ID] = entity.WithdrawnAt != nil
	}
	parentOf := make(map[uuid.UUID]uuid.UUID)
	flagged := make(map[uuid.UUID]uuid.UUID)
	candidates := make(map[uuid.UUID][]*ent.KnowledgeRelationship)
	for _, relationship := range contains {
		isCurrent := relationship.WithdrawnAt == nil && !withdrawn[relationship.SourceEntityID] && !withdrawn[relationship.TargetEntityID]
		if relationship.IsParent {
			require.True(t, isCurrent, "the parent claim %s is current", relationship.ID)
			parentOf[relationship.TargetEntityID] = relationship.SourceEntityID
			flagged[relationship.TargetEntityID] = relationship.ID
		}
		if isCurrent && relationship.SourceEntityID != relationship.TargetEntityID {
			candidates[relationship.TargetEntityID] = append(candidates[relationship.TargetEntityID], relationship)
		}
	}
	hasAncestor := func(entityID, ancestorID uuid.UUID) bool {
		visited := mapset.NewSet[uuid.UUID]()
		for current, found := entityID, true; found && visited.Add(current); current, found = parentOf[current] {
			if current == ancestorID {
				return true
			}
		}
		return false
	}
	for _, entity := range entities {
		claims := candidates[entity.ID]
		slices.SortFunc(claims, func(a, b *ent.KnowledgeRelationship) int {
			return knowledgegraph.RankParentClaims(parentClaimOf(a), parentClaimOf(b))
		})
		chosen, isFallback := uuid.Nil, false
		for _, claim := range claims {
			if hasAncestor(claim.SourceEntityID, entity.ID) {
				isFallback = true
				continue
			}
			chosen = claim.ID
			break
		}
		require.Equal(t, chosen, flagged[entity.ID], "the parent claim of %s is its best claim closing no cycle", entity.ID)
		require.Equal(t, isFallback, entity.ParentIsFallback, "the fallback mark of %s", entity.ID)
	}
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
		chosenID, remainingID := f.parentClaim("service-search").ID, claims[f.ids["service-search"]][0].ID

		f.contains(ke.KindDeleted, "commerce", "service-search")
		s.Equal(remainingID, f.parentClaim("service-search").ID, "a withdrawn claim stops being a candidate")
		s.False(f.tdb.Client(f.ctx).KnowledgeRelationship.GetX(f.ctx, chosenID).IsParent)
	})

	s.Run("a child has at most one parent claim", func() {
		f := s.newParentFixture()
		f.entity(ke.KindObserved, "commerce")
		f.entity(ke.KindObserved, "experience")
		f.entity(ke.KindObserved, "service-search")
		f.contains(ke.KindObserved, "commerce", "service-search")
		f.contains(ke.KindObserved, "experience", "service-search")
		claims, claimsErr := f.queries.ListParentClaims(f.ctx, []uuid.UUID{f.ids["service-search"]})
		s.Require().NoError(claimsErr)
		s.Require().Len(claims[f.ids["service-search"]], 1)

		flagErr := f.tdb.Client(f.ctx).KnowledgeRelationship.UpdateOneID(claims[f.ids["service-search"]][0].ID).
			SetIsParent(true).
			Exec(f.ctx)
		s.True(ent.IsConstraintError(flagErr), "the unique index refuses a second parent claim: %v", flagErr)
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

	s.Run("a claim closing a deep cycle is skipped", func() {
		f := s.newParentFixture()
		const depth = 20
		name := func(level int) string { return fmt.Sprintf("service-%d", level) }
		for level := range depth + 1 {
			f.entity(ke.KindObserved, name(level))
		}
		for level := range depth {
			f.contains(ke.KindObserved, name(level), name(level+1))
		}
		f.contains(ke.KindObserved, name(depth), name(0))
		s.Equal(name(depth-1), f.parentOf(name(depth)))
		s.Empty(f.parentOf(name(0)))
	})

	s.Run("a fallback parent gives way to the best claim once its cycle is gone", func() {
		f := s.newParentFixture()
		for _, name := range []string{"service-a", "service-b", "service-c", "service-d"} {
			f.entity(ke.KindObserved, name)
		}
		f.contains(ke.KindObserved, "service-a", "service-b")
		f.contains(ke.KindObserved, "service-b", "service-a")
		f.contains(ke.KindObserved, "service-c", "service-a")
		s.Equal("service-c", f.parentOf("service-a"), "the best claim would close a cycle")
		s.True(f.isFallback("service-a"))

		f.contains(ke.KindObserved, "service-d", "service-b")
		f.contains(ke.KindDeleted, "service-a", "service-b")
		s.Equal("service-d", f.parentOf("service-b"))
		s.Equal("service-b", f.parentOf("service-a"))
		s.False(f.isFallback("service-a"))
	})

	s.Run("a fallback that moves frees an earlier fallback's claim", func() {
		f := s.newParentFixture()
		// a is created before b, so the fallback pass visits a first.
		for _, name := range []string{"service-a", "service-b", "service-c", "service-d", "service-e"} {
			f.entity(ke.KindObserved, name)
		}
		idA, idB := f.ids["service-a"], f.ids["service-b"]
		s.Require().Negative(bytes.Compare(idA[:], idB[:]))
		// Observed first and withdrawn, the claims c → a and d → b rank first on a and b once restored.
		f.contains(ke.KindObserved, "service-c", "service-a")
		f.contains(ke.KindDeleted, "service-c", "service-a")
		f.contains(ke.KindObserved, "service-d", "service-b")
		f.contains(ke.KindDeleted, "service-d", "service-b")
		f.contains(ke.KindObserved, "service-e", "service-a")
		f.contains(ke.KindObserved, "service-a", "service-b")
		f.contains(ke.KindObserved, "service-b", "service-c")
		f.contains(ke.KindObserved, "service-b", "service-d")
		f.contains(ke.KindObserved, "service-c", "service-a")
		f.contains(ke.KindObserved, "service-d", "service-b")
		f.contains(ke.KindObserved, "service-e", "service-d")
		s.Equal("service-e", f.parentOf("service-a"))
		s.Equal("service-a", f.parentOf("service-b"))
		s.True(f.isFallback("service-a"))
		s.True(f.isFallback("service-b"))

		f.contains(ke.KindDeleted, "service-b", "service-d")
		s.Equal("service-e", f.parentOf("service-d"))
		s.Equal("service-d", f.parentOf("service-b"), "d no longer descends from b")
		s.Equal("service-c", f.parentOf("service-a"), "c no longer descends from a once b moved")
		s.False(f.isFallback("service-a"))
		s.False(f.isFallback("service-b"))
	})
}

func (s *KnowledgeGraphIngestionServiceSuite) TestParentRanking() {
	declared, observed := schematypes.KnowledgeAuthorityDeclared, schematypes.KnowledgeAuthorityObserved
	day := func(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }
	claim := func(f *parentFixture, authority schematypes.KnowledgeAuthority, at time.Time, parent, child string) rez.KnowledgeEvidenceRef {
		ref := f.containsRef(ke.KindObserved, parent, child)
		ref.Authority, ref.EffectiveAt = authority, at
		return ref
	}
	entities := func(f *parentFixture, names ...string) {
		for _, name := range names {
			f.entity(ke.KindObserved, name)
		}
	}

	s.Run("authority first", func() {
		f := s.newParentFixture()
		entities(f, "commerce", "experience", "service-search")
		f.ingest(claim(f, observed, day(1), "commerce", "service-search"))
		f.ingest(claim(f, declared, day(3), "experience", "service-search"))
		s.Equal("experience", f.parentOf("service-search"))
	})

	s.Run("first observation, not insertion", func() {
		f := s.newParentFixture()
		entities(f, "commerce", "experience", "service-search")
		f.ingest(claim(f, declared, day(3), "commerce", "service-search"))
		f.ingest(claim(f, declared, day(1), "experience", "service-search"))
		s.Equal("experience", f.parentOf("service-search"))
	})

	s.Run("rebuilding in another order gives the same parents", func() {
		build := func(reverse bool) map[string]string {
			f := s.newParentFixture()
			entities(f, "platform", "commerce", "experience", "service-search", "service-cart")
			claims := []rez.KnowledgeEvidenceRef{
				claim(f, observed, day(1), "commerce", "service-search"),
				claim(f, declared, day(4), "experience", "service-search"),
				claim(f, declared, day(2), "platform", "service-search"),
				claim(f, observed, day(2), "commerce", "service-cart"),
				claim(f, observed, day(1), "experience", "service-cart"),
				claim(f, declared, day(3), "platform", "commerce"),
			}
			if reverse {
				slices.Reverse(claims)
			}
			for _, ref := range claims {
				f.ingest(ref)
			}
			parents := make(map[string]string)
			for name := range f.ids {
				parents[name] = f.parentOf(name)
			}
			return parents
		}
		forward := build(false)
		s.Equal(map[string]string{"platform": "", "commerce": "platform", "experience": "",
			"service-search": "platform", "service-cart": "experience"}, forward)
		s.Equal(forward, build(true))
	})

	s.Run("an older observation of a withdrawn claim leaves it withdrawn", func() {
		f := s.newParentFixture()
		entities(f, "commerce", "service-search")
		f.ingest(claim(f, observed, day(2), "commerce", "service-search"))
		withdrawal := claim(f, observed, day(3), "commerce", "service-search")
		withdrawal.Kind = ke.KindDeleted
		f.ingest(withdrawal)
		s.Empty(f.parentOf("service-search"))

		f.ingest(claim(f, declared, day(1), "commerce", "service-search"))
		s.Empty(f.parentOf("service-search"), "the alias stays withdrawn, so its earlier observation and higher authority rank nothing")
	})

	s.Run("a declared alias on another claim moves the parent in the same write", func() {
		f := s.newParentFixture()
		entities(f, "commerce", "experience", "service-search")
		f.ingest(claim(f, observed, day(1), "commerce", "service-search"))
		f.ingest(claim(f, observed, day(2), "experience", "service-search"))
		s.Equal("commerce", f.parentOf("service-search"))

		catalogue := claim(f, declared, day(5), "experience", "service-search")
		catalogue.Subject.Relationship.ProviderResourceRef = f.s.testProviderRef("catalogue:contains:experience:service-search")
		f.ingest(catalogue)
		s.Equal("experience", f.parentOf("service-search"))
		s.Equal(1, f.tdb.Client(f.ctx).KnowledgeRelationship.Query().
			Where(knr.SourceEntityID(f.ids["experience"])).
			CountX(f.ctx), "the catalogue's alias joins the existing claim")
	})
}

// TestListingReobservationOfAClaimKeepsItsParent re-observes a time-bounded contains claim through a repeated
// listing: the claim's freshness advances, and its target keeps its parent.
func (s *KnowledgeGraphIngestionServiceSuite) TestListingReobservationOfAClaimKeepsItsParent() {
	f := s.newParentFixture()
	f.entity(ke.KindObserved, "commerce")
	f.entity(ke.KindObserved, "service-search")
	claims := []rez.KnowledgeEvidenceRef{f.containsRef(ke.KindObserved, "commerce", "service-search")}
	claims[0].ValidFor = 3 * time.Hour
	event := f.prepare(claims...)
	claim := claims[0]
	ingestion := s.makeService(f.tdb)
	ingester := NewProviderEventIngester(f.tdb, nil, &evidenceProjector{ingestion, claims}, ingestion)
	s.Require().NoError(ingester.ProjectNormalizedEvent(f.ctx, event.ID, false))
	s.Equal("commerce", f.parentOf("service-search"))
	relationshipID := f.tdb.Client(f.ctx).KnowledgeRelationship.Query().OnlyIDX(f.ctx)

	repeatedAt := claim.EffectiveAt.Add(2 * time.Hour)
	s.Require().NoError(ingester.ReobserveNormalizedEvent(f.ctx, event.ID, repeatedAt))

	relationship := f.tdb.Client(f.ctx).KnowledgeRelationship.GetX(f.ctx, relationshipID)
	s.Require().NotNil(relationship.ExpiresAt)
	s.True(relationship.ExpiresAt.Equal(repeatedAt.Add(claim.ValidFor)), "the repeat extends the claim's validity")
	s.Equal("commerce", f.parentOf("service-search"))
}

// evidenceProjector exercises the projection runner with a batch produced by a provider projector.
type evidenceProjector struct {
	ingestion *KnowledgeGraphIngestionService
	refs      []rez.KnowledgeEvidenceRef
}

func (p *evidenceProjector) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(ctx context.Context, event *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		return nil, p.ingestion.IngestEvidence(ctx, event, p.refs...)
	}, true
}

func (s *KnowledgeGraphIngestionServiceSuite) TestProjectionSkipsInvalidRelationships() {
	ctx, tdb := s.SetupTestDatabase()
	graph := newTestGraph(s.Require(), tdb)
	source := graph.entity(kne.CategoryContainer, "service", "test", "source")
	target := graph.entity(kne.CategoryContainer, "service", "test", "target")
	missing := graph.entity(kne.CategoryContainer, "service", "test", "missing")
	existing := graph.relationship(knr.PredicateDependsOn, source, target)
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	graph.record(ctx, at, graph.observeEntity(source, "source"), graph.observeEntity(target, "target"), graph.observeRelationship(existing))
	mismatch := existing
	mismatch.Source, mismatch.Target = target, source
	invalid := graph.relationship(knr.PredicateRunsOn, source, target)
	self := graph.relationship(knr.PredicateContains, source, source)
	absent := graph.relationship(knr.PredicateDependsOn, source, missing)
	refs := []rez.KnowledgeEvidenceRef{graph.observeEntity(source, "updated"), graph.observeRelationship(existing),
		graph.observeRelationship(mismatch), graph.observeRelationship(invalid), graph.observeRelationship(self), graph.observeRelationship(absent)}
	for i := range refs {
		refs[i].EffectiveAt = at.Add(time.Hour)
	}
	ingestion := s.makeService(tdb)
	ingester := NewProviderEventIngester(tdb, nil, &evidenceProjector{ingestion, refs}, ingestion)
	event := s.createEvent(ctx, tdb, "partial", at.Add(time.Hour))
	s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, event.ID, false))
	receipt := tdb.Client(ctx).NormalizedEventProjection.Query().Where(nep.EventID(event.ID)).OnlyX(ctx)
	reasons := []schematypes.SkippedEvidenceReason{}
	for _, entry := range receipt.SkippedRefs {
		reasons = append(reasons, entry.Reason)
		s.NotEmpty(entry.ResourceRef)
		s.NotEmpty(entry.Error)
	}
	s.ElementsMatch([]schematypes.SkippedEvidenceReason{schematypes.SkippedEvidenceReasonTopologyMismatch, schematypes.SkippedEvidenceReasonPredicateCategories, schematypes.SkippedEvidenceReasonSelfContains, schematypes.SkippedEvidenceReasonMissingEndpoint}, reasons)
	payload, marshalErr := json.Marshal(v1.EventProjectionFromEnt(receipt))
	s.Require().NoError(marshalErr)
	var response map[string]any
	s.Require().NoError(json.Unmarshal(payload, &response))
	attributes := response["attributes"].(map[string]any)
	s.NotContains(attributes, "skipped_refs")
	apiSkipped := attributes["skippedRefs"].([]any)
	s.Require().Len(apiSkipped, len(receipt.SkippedRefs))
	for i, skipped := range receipt.SkippedRefs {
		s.Equal(map[string]any{"reason": string(skipped.Reason), "resourceRef": skipped.ResourceRef, "error": skipped.Error}, apiSkipped[i])
	}
	s.Equal(2, tdb.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).CountX(ctx))
	s.Equal(1, tdb.Client(ctx).KnowledgeRelationship.Query().CountX(ctx))
	s.Equal("updated", tdb.Client(ctx).KnowledgeEntity.GetX(ctx, graph.entityID(ctx, source)).State.DisplayName)
}

func (s *KnowledgeGraphIngestionServiceSuite) TestProjectionRetriesMissingEndpoints() {
	for _, endpointArrives := range []bool{true, false} {
		name := "endpoint never arrives"
		if endpointArrives {
			name = "endpoint event projects before retry"
		}
		s.Run(name, func() {
			ctx, tdb := s.SetupTestDatabase()
			graph := newTestGraph(s.Require(), tdb)
			source := graph.entity(kne.CategoryContainer, "service", "test", "source")
			target := graph.entity(kne.CategoryContainer, "service", "test", "target")
			relationship := graph.relationship(knr.PredicateDependsOn, source, target)
			at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
			refs := []rez.KnowledgeEvidenceRef{graph.observeEntity(source, "source"), graph.observeRelationship(relationship)}
			for i := range refs {
				refs[i].EffectiveAt = at
			}
			ingestion := s.makeService(tdb)
			projector := &evidenceProjector{ingestion, refs}
			ingester := NewProviderEventIngester(tdb, nil, projector, ingestion)
			pipeline := &ProviderEventPipelineService{db: tdb, ingester: ingester}
			worker := river.WorkFunc(pipeline.HandleEventProjectionJob)
			event := s.createEvent(ctx, tdb, "relationship", at)
			job := &river.Job[jobs.ProjectNormalizedEvent]{
				JobRow: &rivertype.JobRow{Attempt: 1}, Args: jobs.ProjectNormalizedEvent{EventId: event.ID},
			}
			lastFailedAttempt := 4
			if endpointArrives {
				lastFailedAttempt = 1
			}
			for ; job.Attempt <= lastFailedAttempt; job.Attempt++ {
				projectionErr := worker.Work(ctx, job)
				s.Require().Error(projectionErr)
				s.True(jobs.IsRetryableError(projectionErr))
				s.Zero(tdb.Client(ctx).NormalizedEventProjection.Query().Where(nep.EventID(event.ID)).CountX(ctx))
				s.Zero(tdb.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).CountX(ctx))
				s.Zero(tdb.Client(ctx).KnowledgeSubjectAlias.Query().CountX(ctx), "alias creation rolls back too")
			}
			if endpointArrives {
				endpointEvent := s.createEvent(ctx, tdb, "endpoint", at)
				endpointRef := graph.observeEntity(target, "target")
				endpointRef.EffectiveAt = at
				projector.refs = []rez.KnowledgeEvidenceRef{endpointRef}
				s.Require().NoError(ingester.ProjectNormalizedEvent(ctx, endpointEvent.ID, false))
				projector.refs = refs
			}
			s.Require().NoError(worker.Work(ctx, job))
			receipt := tdb.Client(ctx).NormalizedEventProjection.Query().Where(nep.EventID(event.ID)).OnlyX(ctx)
			if endpointArrives {
				s.Empty(receipt.SkippedRefs)
				s.Equal(2, tdb.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).CountX(ctx))
				s.Equal(1, tdb.Client(ctx).KnowledgeRelationship.Query().CountX(ctx))
			} else {
				s.Require().Len(receipt.SkippedRefs, 1)
				s.Equal(schematypes.SkippedEvidenceReasonMissingEndpoint, receipt.SkippedRefs[0].Reason)
				s.Equal(relationship.ProviderResourceRef.ResourceRef, receipt.SkippedRefs[0].ResourceRef)
				s.Equal(1, tdb.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).CountX(ctx))
				s.Zero(tdb.Client(ctx).KnowledgeRelationship.Query().CountX(ctx))
			}
			s.Equal("source", tdb.Client(ctx).KnowledgeEntity.GetX(ctx, graph.entityID(ctx, source)).State.DisplayName)
		})
	}
}
