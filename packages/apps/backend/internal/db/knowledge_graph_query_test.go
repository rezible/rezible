package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
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
