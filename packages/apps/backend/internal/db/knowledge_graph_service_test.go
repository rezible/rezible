package db

import (
	"fmt"
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

func (s *KnowledgeGraphServiceSuite) knowledgeService(tdb rez.Database) *KnowledgeGraphService {
	return &KnowledgeGraphService{db: tdb}
}

func (s *KnowledgeGraphServiceSuite) createEvent(tdb rez.Database, kind projections.SubjectKind, eventKind ne.Kind, subjectRef string, occurredAt time.Time, attributes any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttributes, encodeErr := projections.EncodeAttributes(attributes)
	s.Require().NoError(encodeErr)
	event, createErr := tdb.Client(ctx).NormalizedEvent.Create().
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
	tdb := s.CreateTestDatabase()
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()

	apiSubjectRef := "service:api"
	apiAliasRef := ent.KnowledgeSubjectAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: apiSubjectRef,
	}
	apiEntityRef := ent.KnowledgeEntityRef{
		Category:        kne.CategoryContainer,
		Kind:            "service",
		SubjectAliasRef: apiAliasRef,
	}

	apiEvent := s.createEvent(tdb, projections.SubjectKindSystemComponent, ne.KindObserved, apiSubjectRef, now.Add(-2*time.Hour), struct{}{})
	apiEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now.Add(-2 * time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Old API name",
		},
		SubjectEntity: &apiEntityRef,
	}
	apiAlias, ingestApiErr := service.IngestSubjectEvidence(ctx, apiEvent, apiEvidenceRef)
	s.Require().NoError(ingestApiErr)
	s.Require().NotNil(apiAlias.EntityID)
	apiEntityId := *apiAlias.EntityID

	latestEvent := s.createEvent(tdb, projections.SubjectKindSystemComponent, ne.KindObserved, apiSubjectRef, now.Add(-time.Hour), struct{}{})
	apiUpdateEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now.Add(-time.Hour),
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "API",
		},
		SubjectEntity: &apiEntityRef,
	}
	s.Require().NoError(service.IngestEvidence(ctx, latestEvent, apiUpdateEvidenceRef))

	databaseSubjectRef := "service:database"
	databaseEntityAlias := ent.KnowledgeSubjectAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: databaseSubjectRef,
	}
	databaseEntity := ent.KnowledgeEntityRef{
		Category:        kne.CategoryContainer,
		Kind:            "database",
		SubjectAliasRef: databaseEntityAlias,
	}

	databaseEvent := s.createEvent(tdb, projections.SubjectKindSystemComponent, ne.KindObserved, databaseSubjectRef, now.Add(-time.Hour), struct{}{})
	databaseEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:          ke.KindObserved,
		Assertion:     "component_exists",
		EffectiveAt:   now.Add(-time.Hour),
		SubjectState:  schematypes.KnowledgeGraphSubjectState{DisplayName: "Database"},
		SubjectEntity: &databaseEntity,
	}
	s.Require().NoError(service.IngestEvidence(ctx, databaseEvent, databaseEvidenceRef))

	apiDbSubjectRef := "api-uses-database"
	apiDbRelationshipAlias := ent.KnowledgeSubjectAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: apiDbSubjectRef,
	}
	apiDbRelationshipRef := ent.KnowledgeRelationshipRef{
		Predicate:       knr.PredicateUses,
		SubjectAliasRef: apiDbRelationshipAlias,
		Source:          apiEntityRef,
		Target:          databaseEntity,
	}
	relEvidenceRef := ent.KnowledgeEvidenceRef{
		Kind:                ke.KindObserved,
		Assertion:           "relationship_exists",
		EffectiveAt:         now,
		SubjectState:        schematypes.KnowledgeGraphSubjectState{DisplayName: "uses"},
		SubjectRelationship: &apiDbRelationshipRef,
	}
	relationshipEvent := s.createEvent(tdb, projections.SubjectKindSystemRelationship, ne.KindObserved, apiDbSubjectRef, now, struct{}{})
	s.Require().NoError(service.IngestEvidence(ctx, relationshipEvent, relEvidenceRef))

	apiEntity, getApiEntityErr := service.GetEntity(ctx, apiEntityId)
	s.Require().NoError(getApiEntityErr)

	currEv := apiEntity.LatestEvidence()
	s.Require().NotNil(currEv)
	s.Require().NotNil(currEv.SubjectState)
	s.Equal(apiUpdateEvidenceRef.SubjectState.DisplayName, currEv.SubjectState.DisplayName)

	viewParams := rez.GetKnowledgeGraphViewParams{Depth: 1, EntityID: apiEntity.ID}
	view, viewErr := service.GetView(ctx, viewParams)
	s.Require().NoError(viewErr)
	s.Len(view.Entities, 2)
	s.Len(view.Relationships, 1)
	s.Equal(apiEntity.ID, view.RootID)
}

func (s *KnowledgeGraphServiceSuite) TestBoundedViewIncludesIsolatedRootEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.knowledgeService(tdb)
	now := time.Now().UTC()
	apiAliasRef := ent.KnowledgeSubjectAliasRef{
		Provider:           "test",
		ProviderSource:     "knowledge-graph-tests",
		ProviderSubjectRef: "service:api",
	}
	apiEntityRef := ent.KnowledgeEntityRef{
		Category:        kne.CategoryContainer,
		Kind:            "service",
		SubjectAliasRef: apiAliasRef,
	}

	event := s.createEvent(tdb, projections.SubjectKindSystemComponent, ne.KindObserved, apiAliasRef.ProviderSubjectRef, now, struct{}{})
	evidenceRef := ent.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "component_exists",
		EffectiveAt: now,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: "API",
		},
		SubjectEntity: &apiEntityRef,
	}

	r := s.Require()
	alias, ingestErr := service.IngestSubjectEvidence(ctx, event, evidenceRef)
	r.NoError(ingestErr)
	r.NotNil(alias.EntityID)

	rootEntityId := *alias.EntityID

	fmt.Println("entities:")
	for _, e := range tdb.Client(ctx).KnowledgeEntity.Query().AllX(ctx) {
		fmt.Printf("  %+v\n", e)
	}

	fmt.Println("relationships:")
	for _, rel := range tdb.Client(ctx).KnowledgeRelationship.Query().AllX(ctx) {
		fmt.Printf("  %+v\n", rel)
	}

	viewParams := rez.GetKnowledgeGraphViewParams{
		Depth:    1,
		EntityID: rootEntityId,
	}
	view, viewErr := service.GetView(ctx, viewParams)
	r.NoError(viewErr)
	r.Equal(rootEntityId, view.RootID)
	r.Len(view.Entities, 1)
	r.Len(view.Relationships, 0)
	r.Equal(rootEntityId, view.Entities[0].ID)
}

func (s *KnowledgeGraphServiceSuite) TestKnowledgeSubjectAliasIntegrity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	service := s.knowledgeService(tdb)

	source := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryActor).
		SetKind("team").
		SaveX(ctx)
	target := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service").
		SaveX(ctx)
	relationship := client.KnowledgeRelationship.Create().
		SetPredicate(knr.PredicateOwns).
		SetSourceEntityID(source.ID).
		SetTargetEntityID(target.ID).
		SaveX(ctx)

	_, invalidSubjectErr := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderSource("aliases").
		SetProviderSubjectRef("invalid-subject").
		SetEntityID(source.ID).
		SetRelationshipID(relationship.ID).
		Save(ctx)
	s.Require().Error(invalidSubjectErr)

	client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderSource("aliases").
		SetProviderSubjectRef("shared-resource").
		SetEntityID(source.ID).
		SaveX(ctx)
	_, duplicateResourceErr := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindRelationship).
		SetProvider("test").
		SetProviderSource("aliases").
		SetProviderSubjectRef("shared-resource").
		SetRelationshipID(relationship.ID).
		Save(ctx)
	s.Require().Error(duplicateResourceErr)

	ref := ent.KnowledgeRelationshipRef{
		Predicate: knr.PredicateOwns,
		SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
			Provider:           "test",
			ProviderSource:     "aliases",
			ProviderSubjectRef: "shared-resource",
		},
	}
	_, wrongSubjectKindErr := service.lookupExistingRelationshipByRef(ctx, ref, source.ID, target.ID)
	s.Require().ErrorIs(wrongSubjectKindErr, rez.ErrConflict)

	client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindRelationship).
		SetProvider("test").
		SetProviderSource("aliases").
		SetProviderSubjectRef("relationship-resource").
		SetRelationshipID(relationship.ID).
		SaveX(ctx)
	ref.SubjectAliasRef.ProviderSubjectRef = "relationship-resource"
	_, wrongTopologyErr := service.lookupExistingRelationshipByRef(ctx, ref, target.ID, source.ID)
	s.Require().ErrorIs(wrongTopologyErr, rez.ErrConflict)

	client.KnowledgeRelationship.DeleteOneID(relationship.ID).ExecX(ctx)
	remainingRelationshipAliases := client.KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderSubjectRef("relationship-resource")).
		CountX(ctx)
	s.Zero(remainingRelationshipAliases)
}
