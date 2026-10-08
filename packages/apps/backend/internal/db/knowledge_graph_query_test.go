package db

import (
	"context"
	"fmt"
	"slices"
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
	"github.com/rezible/rezible/pkg/knowledgegraph"
	"github.com/rezible/rezible/test"
)

type KnowledgeGraphQueryServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphQueryServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphQueryServiceSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeGraphQueryServiceSuite) newService() (context.Context, rez.Database, *KnowledgeGraphQueryService) {
	// These fixtures reuse fixed entity and relationship primary keys, which are unique across tenants.
	ctx, database := s.SetupTestDatabase(test.WithFreshDatabase())
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	return ctx, database, service
}

func (s *KnowledgeGraphQueryServiceSuite) createEntity(ctx context.Context, database rez.Database, id uuid.UUID, category kne.Category, kind string) *ent.KnowledgeEntity {
	entity, createErr := database.Client(ctx).KnowledgeEntity.Create().
		SetID(id).
		SetCategory(category).
		SetKind(kind).
		Save(ctx)
	s.Require().NoError(createErr)
	return entity
}

func (s *KnowledgeGraphQueryServiceSuite) createRelationship(ctx context.Context, database rez.Database, id uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, relationshipPredicate knr.Predicate) *ent.KnowledgeRelationship {
	relationship, createErr := database.Client(ctx).KnowledgeRelationship.Create().
		SetID(id).
		SetSourceEntityID(sourceID).
		SetTargetEntityID(targetID).
		SetPredicate(relationshipPredicate).
		Save(ctx)
	s.Require().NoError(createErr)
	return relationship
}

func queryTestUUID(value uint64) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012x", value))
}

func (s *KnowledgeGraphQueryServiceSuite) selectEntities(ctx context.Context, service *KnowledgeGraphQueryService, params rez.SelectKnowledgeGraphEntitiesParams) *rez.KnowledgeGraphEntitiesPage {
	page, queryErr := service.SelectGraphEntities(ctx, params)
	s.Require().NoError(queryErr)
	return page
}

func (s *KnowledgeGraphQueryServiceSuite) expandRelationships(ctx context.Context, service *KnowledgeGraphQueryService, params rez.ExpandKnowledgeGraphRelationshipsParams) *rez.KnowledgeGraphRelationshipsPage {
	page, queryErr := service.ExpandGraphRelationships(ctx, params)
	s.Require().NoError(queryErr)
	return page
}

func (s *KnowledgeGraphQueryServiceSuite) TestEntityFiltersAndNormalization() {
	ctx, database, service := s.newService()
	actorUser := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryActor, "user")
	actorTeam := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryActor, "team")
	serviceEntity := s.createEntity(ctx, database, queryTestUUID(3), kne.CategoryContainer, "service")
	apiEntity := s.createEntity(ctx, database, queryTestUUID(4), kne.CategoryContainer, "api")
	caseEntity := s.createEntity(ctx, database, queryTestUUID(5), kne.CategoryContainer, "Service")

	categoryPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Categories: []kne.Category{kne.CategoryActor}},
	})
	s.Equal([]uuid.UUID{actorUser.ID, actorTeam.ID}, entityPageIDs(categoryPage))

	kindPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
	})
	s.Equal([]uuid.UUID{serviceEntity.ID}, entityPageIDs(kindPage))
	s.NotContains(entityPageIDs(kindPage), caseEntity.ID)

	combinedPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{
			Categories: []kne.Category{kne.CategoryActor},
			Kinds:      []string{"team", "user"},
		},
	})
	s.Equal([]uuid.UUID{actorUser.ID, actorTeam.ID}, entityPageIDs(combinedPage))

	unfilteredPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	s.Equal([]uuid.UUID{actorUser.ID, actorTeam.ID, serviceEntity.ID, apiEntity.ID, caseEntity.ID}, entityPageIDs(unfilteredPage))

	categoryValues := []kne.Category{kne.CategoryContainer, kne.CategoryActor, kne.CategoryActor}
	kindValues := []string{"service", "user", "service"}
	normalizedPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Categories: categoryValues, Kinds: kindValues},
	})
	s.Equal([]kne.Category{kne.CategoryContainer, kne.CategoryActor, kne.CategoryActor}, categoryValues)
	s.Equal([]string{"service", "user", "service"}, kindValues)
	reorderedPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{
			Categories: []kne.Category{kne.CategoryActor, kne.CategoryContainer},
			Kinds:      []string{"user", "service"},
		},
	})
	s.Equal(normalizedPage.EntitySelectionRef, reorderedPage.EntitySelectionRef)

	nilFilterPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	emptyFilterPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Categories: []kne.Category{}, Kinds: []string{}},
	})
	s.Equal(nilFilterPage.EntitySelectionRef, emptyFilterPage.EntitySelectionRef)
}

func (s *KnowledgeGraphQueryServiceSuite) TestEntityKeysetPaginationAndPageReferences() {
	ctx, database, service := s.newService()
	first := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryActor, "user")
	second := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryActor, "user")
	third := s.createEntity(ctx, database, queryTestUUID(3), kne.CategoryActor, "user")
	limit := 2
	params := rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"user"}},
		Limit:  &limit,
	}
	firstPage := s.selectEntities(ctx, service, params)
	s.Equal([]uuid.UUID{first.ID, second.ID}, entityPageIDs(firstPage))
	s.Require().NotNil(firstPage.NextCursor)

	firstSelection, decodeFirstErr := service.decodeRef[knowledgeEntitySelection](string(firstPage.EntitySelectionRef))
	s.Require().NoError(decodeFirstErr)
	s.Empty(firstSelection.AfterID)
	s.Equal(limit, firstSelection.Limit)

	params.Cursor = firstPage.NextCursor
	params.Filter = rez.KnowledgeEntityFilter{Categories: []kne.Category{}, Kinds: []string{"user", "user"}}
	secondPage := s.selectEntities(ctx, service, params)
	s.Equal([]uuid.UUID{third.ID}, entityPageIDs(secondPage))
	s.Nil(secondPage.NextCursor)

	secondSelection, decodeSecondErr := service.decodeRef[knowledgeEntitySelection](string(secondPage.EntitySelectionRef))
	s.Require().NoError(decodeSecondErr)
	s.Require().NotNil(secondSelection.AfterID)
	s.Equal(second.ID, *secondSelection.AfterID)
	s.Equal([]string{"user"}, secondSelection.Kinds)

	exactLimit := 3
	exactPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{Limit: &exactLimit})
	s.Len(exactPage.Entities, exactLimit)
	s.Nil(exactPage.NextCursor)

	missingPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"missing"}},
	})
	s.Empty(missingPage.Entities)
	s.NotEmpty(missingPage.EntitySelectionRef)
	s.Nil(missingPage.NextCursor)
}

func (s *KnowledgeGraphQueryServiceSuite) TestExpansionMatchesEitherEndpointAndReturnsEachRelationshipOnce() {
	ctx, database, service := s.newService()
	selectedA := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryContainer, "service")
	selectedB := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "service")
	outsideA := s.createEntity(ctx, database, queryTestUUID(3), kne.CategoryContainer, "database")
	outsideB := s.createEntity(ctx, database, queryTestUUID(4), kne.CategoryContainer, "database")
	first := s.createRelationship(ctx, database, queryTestUUID(101), selectedA.ID, outsideA.ID, knr.PredicateUses)
	second := s.createRelationship(ctx, database, queryTestUUID(102), outsideB.ID, selectedB.ID, knr.PredicateReadsFrom)
	bothSelected := s.createRelationship(ctx, database, queryTestUUID(103), selectedA.ID, selectedB.ID, knr.PredicateDependsOn)
	selfLoop := s.createRelationship(ctx, database, queryTestUUID(104), selectedA.ID, selectedA.ID, knr.PredicateObserves)
	s.createRelationship(ctx, database, queryTestUUID(105), outsideA.ID, outsideB.ID, knr.PredicateUses)

	entityLimit := 1
	firstSelection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
		Limit:  &entityLimit,
	})
	secondSelection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
		Cursor: firstSelection.NextCursor,
		Limit:  &entityLimit,
	})
	firstExpansion := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: firstSelection.EntitySelectionRef,
	})
	secondExpansion := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: secondSelection.EntitySelectionRef,
	})

	s.Equal([]uuid.UUID{first.ID, bothSelected.ID, selfLoop.ID}, relationshipPageIDs(firstExpansion))
	s.Equal([]uuid.UUID{second.ID, bothSelected.ID}, relationshipPageIDs(secondExpansion))
	for _, relationship := range firstExpansion.Relationships {
		s.Nil(relationship.Edges.SourceEntity)
		s.Nil(relationship.Edges.TargetEntity)
		s.Nil(relationship.Edges.Aliases)
	}
}
func (s *KnowledgeGraphQueryServiceSuite) TestRelationshipFiltersPaginationAndLiveReplay() {
	ctx, database, service := s.newService()
	selected := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryContainer, "service")
	outside := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "database")
	first := s.createRelationship(ctx, database, queryTestUUID(10), selected.ID, outside.ID, knr.PredicateUses)
	second := s.createRelationship(ctx, database, queryTestUUID(20), outside.ID, selected.ID, knr.PredicateContains)
	s.createRelationship(ctx, database, queryTestUUID(25), selected.ID, outside.ID, knr.PredicateObserves)
	selection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
	})
	predicates := []knr.Predicate{knr.PredicateUses, knr.PredicateContains, knr.PredicateUses}
	limit := 1
	firstPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         predicates,
		Limit:              &limit,
	})
	s.Equal([]uuid.UUID{first.ID}, relationshipPageIDs(firstPage))
	s.Require().NotNil(firstPage.NextCursor)
	s.Equal([]knr.Predicate{knr.PredicateUses, knr.PredicateContains, knr.PredicateUses}, predicates)

	liveRelationship := s.createRelationship(ctx, database, queryTestUUID(15), outside.ID, selected.ID, knr.PredicateUses)
	continuationPredicates := []knr.Predicate{knr.PredicateContains, knr.PredicateUses}
	secondPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         continuationPredicates,
		Cursor:             firstPage.NextCursor,
		Limit:              &limit,
	})
	s.Equal([]uuid.UUID{liveRelationship.ID}, relationshipPageIDs(secondPage))
	s.Require().NotNil(secondPage.NextCursor)
	s.Equal([]knr.Predicate{knr.PredicateContains, knr.PredicateUses}, continuationPredicates)

	thirdPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         []knr.Predicate{knr.PredicateUses, knr.PredicateContains},
		Cursor:             secondPage.NextCursor,
		Limit:              &limit,
	})
	s.Equal([]uuid.UUID{second.ID}, relationshipPageIDs(thirdPage))
	s.Nil(thirdPage.NextCursor)
}

func (s *KnowledgeGraphQueryServiceSuite) TestSelectionReplayTracksGraphChangesAndEmptySelections() {
	ctx, database, service := s.newService()
	selection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"late_entity"}},
	})
	descriptor, decodeErr := service.decodeRef[knowledgeEntitySelection](string(selection.EntitySelectionRef))
	s.Require().NoError(decodeErr)
	s.Equal([]string{"late_entity"}, descriptor.Kinds)
	s.Equal(defaultEntityQueryLimit, descriptor.Limit)
	s.Empty(descriptor.AfterID)
	s.Empty(selection.Entities)

	emptyExpansion := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
	})
	s.Empty(emptyExpansion.Relationships)
	s.Nil(emptyExpansion.NextCursor)

	newEntity := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryContainer, "late_entity")
	outside := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "database")
	newRelationship := s.createRelationship(ctx, database, queryTestUUID(101), outside.ID, newEntity.ID, knr.PredicateUses)
	replayedExpansion := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
	})
	s.Equal([]uuid.UUID{newRelationship.ID}, relationshipPageIDs(replayedExpansion))
}

func (s *KnowledgeGraphQueryServiceSuite) TestQueryLimits() {
	ctx, database, service := s.newService()

	defaultEntityPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	defaultEntityDescriptor, decodeErr := service.decodeRef[knowledgeEntitySelection](string(defaultEntityPage.EntitySelectionRef))
	s.Require().NoError(decodeErr)
	s.Equal(defaultEntityQueryLimit, defaultEntityDescriptor.Limit)

	maximumEntityLimit := maxEntityQueryLimit
	maximumEntityPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{Limit: &maximumEntityLimit})
	maximumEntityDescriptor, maximumDecodeErr := service.decodeRef[knowledgeEntitySelection](string(maximumEntityPage.EntitySelectionRef))
	s.Require().NoError(maximumDecodeErr)
	s.Equal(maxEntityQueryLimit, maximumEntityDescriptor.Limit)

	selected := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryContainer, "service")
	outside := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "database")
	s.createRelationship(ctx, database, queryTestUUID(10), selected.ID, outside.ID, knr.PredicateUses)
	s.createRelationship(ctx, database, queryTestUUID(20), selected.ID, outside.ID, knr.PredicateContains)
	selection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
	})

	maximumRelationshipLimit := maxRelationshipLimit
	maximumRelationshipPage, maxRelationshipErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Limit:              &maximumRelationshipLimit,
	})
	s.Require().NoError(maxRelationshipErr)
	s.Equal([]uuid.UUID{queryTestUUID(10), queryTestUUID(20)}, relationshipPageIDs(maximumRelationshipPage))

	defaultRelationshipPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
	})
	s.Equal([]uuid.UUID{queryTestUUID(10), queryTestUUID(20)}, relationshipPageIDs(defaultRelationshipPage))

	bulkEntities := make([]*ent.KnowledgeEntityCreate, defaultRelationshipLimit+1)
	bulkRelationships := make([]*ent.KnowledgeRelationshipCreate, len(bulkEntities))
	for index := range bulkRelationships {
		targetID := queryTestUUID(uint64(index + 1000))
		bulkEntities[index] = database.Client(ctx).KnowledgeEntity.Create().
			SetID(targetID).
			SetCategory(kne.CategoryContainer).
			SetKind("database")
		bulkRelationships[index] = database.Client(ctx).KnowledgeRelationship.Create().
			SetID(queryTestUUID(uint64(index + 1000))).
			SetSourceEntityID(selected.ID).
			SetTargetEntityID(targetID).
			SetPredicate(knr.PredicateUses)
	}
	createEntities := database.Client(ctx).KnowledgeEntity.CreateBulk(bulkEntities...)
	s.Require().NoError(createEntities.Exec(ctx))
	createRelationships := database.Client(ctx).KnowledgeRelationship.CreateBulk(bulkRelationships...)
	s.Require().NoError(createRelationships.Exec(ctx))

	defaultRelationshipPage = s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
	})
	s.Len(defaultRelationshipPage.Relationships, defaultRelationshipLimit)
	s.Require().NotNil(defaultRelationshipPage.NextCursor)
	continuation := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Cursor:             defaultRelationshipPage.NextCursor,
	})
	s.Len(continuation.Relationships, 3)
	s.Nil(continuation.NextCursor)
}

func (s *KnowledgeGraphQueryServiceSuite) TestQueryValidation() {
	ctx, database, service := s.newService()
	emptySelection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"no_match"}},
	})

	invalidEntityLimits := []int{0, maxEntityQueryLimit + 1}
	for _, invalidLimit := range invalidEntityLimits {
		_, queryErr := service.SelectGraphEntities(ctx, rez.SelectKnowledgeGraphEntitiesParams{Limit: &invalidLimit})
		s.ErrorIs(queryErr, rez.ErrInvalidInput)
	}
	invalidRelationshipLimits := []int{0, maxRelationshipLimit + 1}
	for _, invalidLimit := range invalidRelationshipLimits {
		_, queryErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
			EntitySelectionRef: emptySelection.EntitySelectionRef,
			Limit:              &invalidLimit,
		})
		s.ErrorIs(queryErr, rez.ErrInvalidInput)
	}

	_, categoryErr := service.SelectGraphEntities(ctx, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Categories: []kne.Category{"unknown_category"}},
	})
	s.ErrorIs(categoryErr, rez.ErrInvalidInput)
	_, kindErr := service.SelectGraphEntities(ctx, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{""}},
	})
	s.ErrorIs(kindErr, rez.ErrInvalidInput)
	_, predicateErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: emptySelection.EntitySelectionRef,
		Predicates:         []knr.Predicate{"unknown_predicate"},
	})
	s.ErrorIs(predicateErr, rez.ErrInvalidInput)

	emptyCursor := rez.RelationshipExpansionCursor("")
	_, cursorErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: emptySelection.EntitySelectionRef,
		Cursor:             &emptyCursor,
	})
	s.ErrorIs(cursorErr, rez.ErrInvalidInput)
	_, referenceErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{})
	s.ErrorIs(referenceErr, rez.ErrInvalidInput)

	selected := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryContainer, "service")
	outside := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "database")
	s.createEntity(ctx, database, queryTestUUID(3), kne.CategoryContainer, "service")
	s.createRelationship(ctx, database, queryTestUUID(10), selected.ID, outside.ID, knr.PredicateUses)
	s.createRelationship(ctx, database, queryTestUUID(20), selected.ID, outside.ID, knr.PredicateContains)
	entityLimit := 1
	entityPage := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
		Limit:  &entityLimit,
	})
	_, entityCursorErr := service.SelectGraphEntities(ctx, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{}},
		Cursor: entityPage.NextCursor,
		Limit:  &entityLimit,
	})
	s.ErrorIs(entityCursorErr, rez.ErrInvalidInput)

	selection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
	})
	relationshipLimit := 1
	relationshipPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Limit:              &relationshipLimit,
	})
	s.Require().NotNil(relationshipPage.NextCursor)
	_, relationshipCursorErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         []knr.Predicate{knr.PredicateUses},
		Cursor:             relationshipPage.NextCursor,
		Limit:              &relationshipLimit,
	})
	s.ErrorIs(relationshipCursorErr, rez.ErrInvalidInput)
}

func entityPageIDs(page *rez.KnowledgeGraphEntitiesPage) []uuid.UUID {
	ids := make([]uuid.UUID, len(page.Entities))
	for index, entity := range page.Entities {
		ids[index] = entity.ID
	}
	return ids
}

func relationshipPageIDs(page *rez.KnowledgeGraphRelationshipsPage) []uuid.UUID {
	ids := make([]uuid.UUID, len(page.Relationships))
	for index, relationship := range page.Relationships {
		ids[index] = relationship.ID
	}
	return ids
}

func (s *KnowledgeGraphQueryServiceSuite) structureEntityRef(name string, category kne.Category) *rez.KnowledgeEntityRef {
	return &rez.KnowledgeEntityRef{
		Category: category,
		Kind:     string(category),
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "test",
			ProviderNamespace: "knowledge-graph-tests",
			ResourceRef:       "entity:" + name,
		},
	}
}

func (s *KnowledgeGraphQueryServiceSuite) observeEntity(ref *rez.KnowledgeEntityRef) rez.KnowledgeEvidenceRef {
	return rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "entity_observed",
		EffectiveAt: time.Now().UTC(),
		Subject:     rez.KnowledgeSubjectRef{Entity: ref},
	}
}

func (s *KnowledgeGraphQueryServiceSuite) observeContains(parent *rez.KnowledgeEntityRef, child *rez.KnowledgeEntityRef) rez.KnowledgeEvidenceRef {
	ref := rez.KnowledgeRelationshipRef{
		Predicate: knr.PredicateContains,
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "test",
			ProviderNamespace: "knowledge-graph-tests",
			ResourceRef:       "contains:" + parent.ProviderResourceRef.ResourceRef + ":" + child.ProviderResourceRef.ResourceRef,
		},
		Source: *parent,
		Target: *child,
	}
	return rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "relationship_observed",
		EffectiveAt: time.Now().UTC(),
		Subject:     rez.KnowledgeSubjectRef{Relationship: &ref},
	}
}

func (s *KnowledgeGraphQueryServiceSuite) ingestStructure(ctx context.Context, database rez.Database, evidence ...rez.KnowledgeEvidenceRef) {
	now := time.Now().UTC()
	createEvent := database.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("knowledge-graph-tests").
		SetProviderResourceRef("structure").
		SetProviderEventSource("knowledge-graph-tests").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetKind("system_component").
		SetOccurredAt(now).
		SetReceivedAt(now).
		SetAttributes([]byte(`{}`))
	event, createErr := createEvent.Save(ctx)
	s.Require().NoError(createErr)

	ingestion, ingestionErr := NewKnowledgeGraphIngestionService(database)
	s.Require().NoError(ingestionErr)
	s.Require().NoError(ingestion.IngestEvidence(ctx, event, evidence...))
}

func (s *KnowledgeGraphQueryServiceSuite) ingestedEntityID(ctx context.Context, database rez.Database, ref *rez.KnowledgeEntityRef) uuid.UUID {
	entityQuery := database.Client(ctx).KnowledgeEntity.Query().
		Where(kne.HasAliasesWith(ksa.ProviderResourceRef(ref.ProviderResourceRef.ResourceRef)))
	id, queryErr := entityQuery.OnlyID(ctx)
	s.Require().NoError(queryErr)
	return id
}

func (s *KnowledgeGraphQueryServiceSuite) TestResolveStructureClimbsStoredRelationships() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	checkout := s.structureEntityRef("checkout", kne.CategoryContainer)
	orders := s.structureEntityRef("orders", kne.CategoryComponent)
	pricing := s.structureEntityRef("pricing", kne.CategoryComponent)
	s.ingestStructure(ctx, database,
		s.observeEntity(checkout),
		s.observeEntity(orders),
		s.observeEntity(pricing),
		s.observeContains(checkout, orders),
		s.observeContains(orders, pricing),
	)
	checkoutID := s.ingestedEntityID(ctx, database, checkout)
	pricingID := s.ingestedEntityID(ctx, database, pricing)

	// The container is exactly MaxDepth steps away, so loading one step too few would miss it.
	params := rez.ResolveStructureParams{
		EntityIDs:        []uuid.UUID{pricingID, checkoutID},
		TargetCategories: knowledgegraph.StructureLevelRuntime.Categories(),
		MaxDepth:         2,
	}
	resolved, resolveErr := service.ResolveStructure(ctx, params)
	s.Require().NoError(resolveErr)

	expected := map[uuid.UUID][]uuid.UUID{
		pricingID:  {checkoutID},
		checkoutID: {checkoutID},
	}
	s.Equal(expected, resolved)
}

// relatedEventFixture is an entity related to a target, ingested with its state effective at a time.
type relatedEventFixture struct {
	name       string
	category   kne.Category
	kind       string
	at         time.Time
	properties map[string]any
	predicate  knr.Predicate
	target     *rez.KnowledgeEntityRef
	// reversed makes the target the relationship's source.
	reversed bool
}

func (s *KnowledgeGraphQueryServiceSuite) ingestRelatedEvent(ctx context.Context, database rez.Database, fixture relatedEventFixture) uuid.UUID {
	ref := &rez.KnowledgeEntityRef{
		Category: fixture.category,
		Kind:     fixture.kind,
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "test",
			ProviderNamespace: "knowledge-graph-tests",
			ResourceRef:       "event:" + fixture.name,
		},
	}
	eventEvidence := rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "event_observed",
		EffectiveAt: fixture.at,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: fixture.name,
			Properties:  fixture.properties,
		},
		Subject: rez.KnowledgeSubjectRef{Entity: ref},
	}
	relationship := rez.KnowledgeRelationshipRef{
		Predicate: fixture.predicate,
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          "test",
			ProviderNamespace: "knowledge-graph-tests",
			ResourceRef:       "relationship:" + fixture.name,
		},
		Source: *ref,
		Target: *fixture.target,
	}
	if fixture.reversed {
		relationship.Source, relationship.Target = relationship.Target, relationship.Source
	}
	relationshipEvidence := rez.KnowledgeEvidenceRef{
		Kind:        ke.KindObserved,
		Assertion:   "relationship_observed",
		EffectiveAt: fixture.at,
		Subject:     rez.KnowledgeSubjectRef{Relationship: &relationship},
	}
	s.ingestStructure(ctx, database,
		s.observeEntity(fixture.target),
		eventEvidence,
		relationshipEvidence,
	)
	return s.ingestedEntityID(ctx, database, ref)
}

type relatedEventsFixture struct {
	service         uuid.UUID
	at              time.Time
	firstProduction uuid.UUID
	staging         uuid.UUID
	// failedProduction and laterProduction share a time.
	failedProduction uuid.UUID
	laterProduction  uuid.UUID
}

// seedRelatedEvents ingests checkout's deployments a minute apart, and entities that do not match a lookup
// of checkout's deployments by impacts.
func (s *KnowledgeGraphQueryServiceSuite) seedRelatedEvents(ctx context.Context, database rez.Database) *relatedEventsFixture {
	at := time.Date(2026, 10, 7, 14, 0, 0, 0, time.UTC)
	checkout := s.structureEntityRef("checkout", kne.CategoryContainer)
	search := s.structureEntityRef("search", kne.CategoryContainer)
	deployment := func(name string, offset time.Duration, environment string, status string) relatedEventFixture {
		return relatedEventFixture{
			name:     name,
			category: kne.CategoryEvent,
			kind:     "deployment",
			at:       at.Add(offset),
			properties: map[string]any{
				"environment": environment,
				"status":      status,
			},
			predicate: knr.PredicateImpacts,
			target:    checkout,
		}
	}

	f := &relatedEventsFixture{at: at}
	f.firstProduction = s.ingestRelatedEvent(ctx, database, deployment("first-production", 0, "production", "succeeded"))
	f.staging = s.ingestRelatedEvent(ctx, database, deployment("staging", time.Minute, "staging", "succeeded"))
	f.failedProduction = s.ingestRelatedEvent(ctx, database, deployment("failed-production", 2*time.Minute, "production", "failed"))
	laterProduction := deployment("later-production", 2*time.Minute, "production", "succeeded")
	laterProduction.properties["attempt"] = 2
	laterProduction.properties["rollback"] = true
	laterProduction.properties["version"] = `v1 "beta" it's`
	f.laterProduction = s.ingestRelatedEvent(ctx, database, laterProduction)

	otherService := deployment("search-production", time.Minute, "production", "succeeded")
	otherService.target = search
	s.ingestRelatedEvent(ctx, database, otherService)

	otherPredicate := deployment("touches-checkout", time.Minute, "production", "succeeded")
	otherPredicate.predicate = knr.PredicateTouches
	s.ingestRelatedEvent(ctx, database, otherPredicate)

	otherKind := deployment("code-change", time.Minute, "production", "succeeded")
	otherKind.kind = "code_change"
	s.ingestRelatedEvent(ctx, database, otherKind)

	// The service impacts the event: the relationship runs the other way.
	reversed := deployment("reversed", time.Minute, "production", "succeeded")
	reversed.reversed = true
	s.ingestRelatedEvent(ctx, database, reversed)

	// A Kubernetes Deployment is a deployment that is not an event.
	notAnEvent := deployment("kubernetes-deployment", time.Minute, "production", "succeeded")
	notAnEvent.category = kne.CategoryContainer
	s.ingestRelatedEvent(ctx, database, notAnEvent)

	f.service = s.ingestedEntityID(ctx, database, checkout)
	return f
}

// tied orders events that share a time by ID.
func (f *relatedEventsFixture) tied() []uuid.UUID {
	ids := []uuid.UUID{f.failedProduction, f.laterProduction}
	slices.SortFunc(ids, func(a, b uuid.UUID) int {
		return strings.Compare(a.String(), b.String())
	})
	return ids
}

func (s *KnowledgeGraphQueryServiceSuite) listRelatedEvents(ctx context.Context, service *KnowledgeGraphQueryService, params rez.ListRelatedEventsParams) ([]uuid.UUID, bool) {
	s.T().Helper()
	related, listErr := service.ListRelatedEvents(ctx, params)
	s.Require().NoError(listErr)
	ids := make([]uuid.UUID, len(related.Events))
	for index, event := range related.Events {
		ids[index] = event.ID
	}
	return ids, related.Cut
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsMatchesKindPredicateAndProperties() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	f := s.seedRelatedEvents(ctx, database)
	params := rez.ListRelatedEventsParams{
		EntityID:  f.service,
		Predicate: knr.PredicateImpacts,
		Kind:      "deployment",
		Limit:     10,
	}

	all, _ := s.listRelatedEvents(ctx, service, params)
	expectedAll := append(f.tied(), f.staging, f.firstProduction)
	s.Equal(expectedAll, all, "only the entity's deployment events, newest first with ties by ID")

	params.PropertyEquals = map[string]string{"environment": "production"}
	production, _ := s.listRelatedEvents(ctx, service, params)
	expectedProduction := append(f.tied(), f.firstProduction)
	s.Equal(expectedProduction, production)

	params.PropertyEquals = map[string]string{"environment": "production", "status": "succeeded"}
	succeeded, _ := s.listRelatedEvents(ctx, service, params)
	s.Equal([]uuid.UUID{f.laterProduction, f.firstProduction}, succeeded, "every property filter must match")
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsMatchesAbsentProperties() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	f := s.seedRelatedEvents(ctx, database)
	params := rez.ListRelatedEventsParams{
		EntityID:       f.service,
		Predicate:      knr.PredicateImpacts,
		Kind:           "deployment",
		PropertyAbsent: []string{"version"},
		Limit:          10,
	}

	withoutVersion, _ := s.listRelatedEvents(ctx, service, params)
	expected := []uuid.UUID{f.failedProduction, f.staging, f.firstProduction}
	s.Equal(expected, withoutVersion, "only events without the property")

	params.PropertyAbsent = []string{"version", "environment"}
	withoutEither, _ := s.listRelatedEvents(ctx, service, params)
	s.Empty(withoutEither, "every absent filter must match")

	params.PropertyAbsent = nil
	params.PropertyEquals = map[string]string{"version": ""}
	emptyVersion, _ := s.listRelatedEvents(ctx, service, params)
	s.Empty(emptyVersion, "an absent property never matches an equality filter")

	checkout := s.structureEntityRef("checkout", kne.CategoryContainer)
	nullVersion := relatedEventFixture{
		name:       "null-version",
		category:   kne.CategoryEvent,
		kind:       "deployment",
		at:         f.at.Add(-time.Minute),
		properties: map[string]any{"version": nil},
		predicate:  knr.PredicateImpacts,
		target:     checkout,
	}
	nullVersionID := s.ingestRelatedEvent(ctx, database, nullVersion)
	emptyStringVersion := relatedEventFixture{
		name:       "empty-version",
		category:   kne.CategoryEvent,
		kind:       "deployment",
		at:         f.at.Add(-2 * time.Minute),
		properties: map[string]any{"version": ""},
		predicate:  knr.PredicateImpacts,
		target:     checkout,
	}
	emptyStringVersionID := s.ingestRelatedEvent(ctx, database, emptyStringVersion)

	params.PropertyEquals = nil
	params.PropertyAbsent = []string{"version"}
	withNull, _ := s.listRelatedEvents(ctx, service, params)
	expectedWithNull := []uuid.UUID{f.failedProduction, f.staging, f.firstProduction, nullVersionID}
	s.Equal(expectedWithNull, withNull, "a JSON null is absent; an empty string is not")

	params.PropertyAbsent = nil
	params.PropertyEquals = map[string]string{"version": ""}
	emptyString, _ := s.listRelatedEvents(ctx, service, params)
	s.Equal([]uuid.UUID{emptyStringVersionID}, emptyString, "an empty string equals an empty value")
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsWindowBounds() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	f := s.seedRelatedEvents(ctx, database)

	cases := map[string]struct {
		fromInclusive bool
		toInclusive   bool
		expected      []uuid.UUID
	}{
		"both inclusive": {
			fromInclusive: true,
			toInclusive:   true,
			expected:      append(f.tied(), f.staging, f.firstProduction),
		},
		"both exclusive": {
			expected: []uuid.UUID{f.staging},
		},
		"from inclusive": {
			fromInclusive: true,
			expected:      []uuid.UUID{f.staging, f.firstProduction},
		},
		"to inclusive": {
			toInclusive: true,
			expected:    append(f.tied(), f.staging),
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			params := rez.ListRelatedEventsParams{
				EntityID:      f.service,
				Predicate:     knr.PredicateImpacts,
				Kind:          "deployment",
				From:          f.at,
				To:            f.at.Add(2 * time.Minute),
				FromInclusive: tc.fromInclusive,
				ToInclusive:   tc.toInclusive,
				Limit:         10,
			}

			events, _ := s.listRelatedEvents(ctx, service, params)

			s.Equal(tc.expected, events)
		})
	}

	onlyFrom := rez.ListRelatedEventsParams{
		EntityID:  f.service,
		Predicate: knr.PredicateImpacts,
		Kind:      "deployment",
		From:      f.at.Add(time.Minute),
		Limit:     10,
	}
	afterStart, _ := s.listRelatedEvents(ctx, service, onlyFrom)
	s.Equal(f.tied(), afterStart, "a zero To leaves the window open")
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsLimit() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	f := s.seedRelatedEvents(ctx, database)
	params := rez.ListRelatedEventsParams{
		EntityID:  f.service,
		Predicate: knr.PredicateImpacts,
		Kind:      "deployment",
		Limit:     2,
	}

	cutEvents, cut := s.listRelatedEvents(ctx, service, params)
	s.Equal(f.tied(), cutEvents)
	s.True(cut, "more events matched than the limit")

	params.Limit = 4
	allEvents, allCut := s.listRelatedEvents(ctx, service, params)
	s.Len(allEvents, 4)
	s.False(allCut, "exactly the limit matched")

	for _, limit := range []int{0, -1} {
		params.Limit = limit
		_, listErr := service.ListRelatedEvents(ctx, params)
		s.ErrorIs(listErr, rez.ErrInvalidInput, "limit %d", limit)
	}
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsComparesPropertiesAsText() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	f := s.seedRelatedEvents(ctx, database)

	cases := map[string]struct {
		property string
		value    string
		expected []uuid.UUID
	}{
		"number":                {property: "attempt", value: "2", expected: []uuid.UUID{f.laterProduction}},
		"boolean":               {property: "rollback", value: "true", expected: []uuid.UUID{f.laterProduction}},
		"quoted value":          {property: "version", value: `v1 "beta" it's`, expected: []uuid.UUID{f.laterProduction}},
		"injection-like value":  {property: "status", value: `succeeded' OR '1'='1`, expected: []uuid.UUID{}},
		"value of another type": {property: "attempt", value: "two", expected: []uuid.UUID{}},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			params := rez.ListRelatedEventsParams{
				EntityID:       f.service,
				Predicate:      knr.PredicateImpacts,
				Kind:           "deployment",
				PropertyEquals: map[string]string{tc.property: tc.value},
				Limit:          10,
			}

			events, _ := s.listRelatedEvents(ctx, service, params)

			s.Equal(tc.expected, events)
		})
	}
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsRejectsInvalidParams() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	valid := rez.ListRelatedEventsParams{
		EntityID:  uuid.New(),
		Predicate: knr.PredicateImpacts,
		Kind:      "deployment",
		Limit:     10,
	}

	cases := map[string]func(*rez.ListRelatedEventsParams){
		"zero limit":          func(p *rez.ListRelatedEventsParams) { p.Limit = 0 },
		"negative limit":      func(p *rez.ListRelatedEventsParams) { p.Limit = -1 },
		"empty kind":          func(p *rez.ListRelatedEventsParams) { p.Kind = "" },
		"unknown predicate":   func(p *rez.ListRelatedEventsParams) { p.Predicate = "knows" },
		"empty predicate":     func(p *rez.ListRelatedEventsParams) { p.Predicate = "" },
		"quote in key":        func(p *rez.ListRelatedEventsParams) { p.PropertyEquals = map[string]string{`status' OR '1'='1`: "x"} },
		"path operator key":   func(p *rez.ListRelatedEventsParams) { p.PropertyEquals = map[string]string{"status'->>'x": "x"} },
		"empty property key":  func(p *rez.ListRelatedEventsParams) { p.PropertyEquals = map[string]string{"": "x"} },
		"quote in absent key": func(p *rez.ListRelatedEventsParams) { p.PropertyAbsent = []string{`status' OR '1'='1`} },
		"empty absent key":    func(p *rez.ListRelatedEventsParams) { p.PropertyAbsent = []string{""} },
	}
	for name, change := range cases {
		s.Run(name, func() {
			params := valid
			change(&params)

			_, listErr := service.ListRelatedEvents(ctx, params)

			s.ErrorIs(listErr, rez.ErrInvalidInput)
		})
	}
}

func (s *KnowledgeGraphQueryServiceSuite) TestListRelatedEventsIsTenantScoped() {
	ownerCtx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	f := s.seedRelatedEvents(ownerCtx, database)
	otherTenantCtx, _ := s.SetupTestDatabase()
	params := rez.ListRelatedEventsParams{
		EntityID:  f.service,
		Predicate: knr.PredicateImpacts,
		Kind:      "deployment",
		Limit:     10,
	}

	ownerEvents, _ := s.listRelatedEvents(ownerCtx, service, params)
	otherTenantEvents, _ := s.listRelatedEvents(otherTenantCtx, service, params)

	s.Len(ownerEvents, 4)
	s.Empty(otherTenantEvents, "another tenant's entity ID finds nothing")
}
