package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
	v1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/pkg/situations"
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

func (s *KnowledgeGraphQueryServiceSuite) TestEntitySummariesCarryDisplayName() {
	ctx, database, service := s.newService()
	named, createErr := database.Client(ctx).KnowledgeEntity.Create().
		SetID(queryTestUUID(1)).
		SetCategory(kne.CategoryContainer).
		SetKind("service").
		SetState(schematypes.KnowledgeGraphSubjectState{DisplayName: "checkout-api"}).
		Save(ctx)
	s.Require().NoError(createErr)
	unnamed := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "service")

	page := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	s.Require().Equal([]uuid.UUID{named.ID, unnamed.ID}, entityPageIDs(page))
	s.Equal("checkout-api", page.Entities[0].State.DisplayName)
	s.Empty(page.Entities[1].State.DisplayName)
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
	first := s.createRelationship(ctx, database, queryTestUUID(101), selectedA.ID, outsideA.ID, knr.PredicateDependsOn)
	second := s.createRelationship(ctx, database, queryTestUUID(102), outsideB.ID, selectedB.ID, knr.PredicateReadsFrom)
	bothSelected := s.createRelationship(ctx, database, queryTestUUID(103), selectedA.ID, selectedB.ID, knr.PredicateDependsOn)
	selfLoop := s.createRelationship(ctx, database, queryTestUUID(104), selectedA.ID, selectedA.ID, knr.PredicateObserves)
	s.createRelationship(ctx, database, queryTestUUID(105), outsideA.ID, outsideB.ID, knr.PredicateDependsOn)

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
	first := s.createRelationship(ctx, database, queryTestUUID(10), selected.ID, outside.ID, knr.PredicateDependsOn)
	second := s.createRelationship(ctx, database, queryTestUUID(20), outside.ID, selected.ID, knr.PredicateContains)
	s.createRelationship(ctx, database, queryTestUUID(25), selected.ID, outside.ID, knr.PredicateObserves)
	selection := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{"service"}},
	})
	predicates := []knr.Predicate{knr.PredicateDependsOn, knr.PredicateContains, knr.PredicateDependsOn}
	limit := 1
	firstPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         predicates,
		Limit:              &limit,
	})
	s.Equal([]uuid.UUID{first.ID}, relationshipPageIDs(firstPage))
	s.Require().NotNil(firstPage.NextCursor)
	s.Equal([]knr.Predicate{knr.PredicateDependsOn, knr.PredicateContains, knr.PredicateDependsOn}, predicates)

	liveRelationship := s.createRelationship(ctx, database, queryTestUUID(15), outside.ID, selected.ID, knr.PredicateDependsOn)
	continuationPredicates := []knr.Predicate{knr.PredicateContains, knr.PredicateDependsOn}
	secondPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         continuationPredicates,
		Cursor:             firstPage.NextCursor,
		Limit:              &limit,
	})
	s.Equal([]uuid.UUID{liveRelationship.ID}, relationshipPageIDs(secondPage))
	s.Require().NotNil(secondPage.NextCursor)
	s.Equal([]knr.Predicate{knr.PredicateContains, knr.PredicateDependsOn}, continuationPredicates)

	thirdPage := s.expandRelationships(ctx, service, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: selection.EntitySelectionRef,
		Predicates:         []knr.Predicate{knr.PredicateDependsOn, knr.PredicateContains},
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
	newRelationship := s.createRelationship(ctx, database, queryTestUUID(101), outside.ID, newEntity.ID, knr.PredicateDependsOn)
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
	s.createRelationship(ctx, database, queryTestUUID(10), selected.ID, outside.ID, knr.PredicateDependsOn)
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
			SetPredicate(knr.PredicateDependsOn)
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
		s.ErrorIs(queryErr, errs.ErrInvalidInput)
	}
	invalidRelationshipLimits := []int{0, maxRelationshipLimit + 1}
	for _, invalidLimit := range invalidRelationshipLimits {
		_, queryErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
			EntitySelectionRef: emptySelection.EntitySelectionRef,
			Limit:              &invalidLimit,
		})
		s.ErrorIs(queryErr, errs.ErrInvalidInput)
	}

	_, categoryErr := service.SelectGraphEntities(ctx, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Categories: []kne.Category{"unknown_category"}},
	})
	s.ErrorIs(categoryErr, errs.ErrInvalidInput)
	_, kindErr := service.SelectGraphEntities(ctx, rez.SelectKnowledgeGraphEntitiesParams{
		Filter: rez.KnowledgeEntityFilter{Kinds: []string{""}},
	})
	s.ErrorIs(kindErr, errs.ErrInvalidInput)
	_, predicateErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: emptySelection.EntitySelectionRef,
		Predicates:         []knr.Predicate{"unknown_predicate"},
	})
	s.ErrorIs(predicateErr, errs.ErrInvalidInput)

	emptyCursor := rez.RelationshipExpansionCursor("")
	_, cursorErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{
		EntitySelectionRef: emptySelection.EntitySelectionRef,
		Cursor:             &emptyCursor,
	})
	s.ErrorIs(cursorErr, errs.ErrInvalidInput)
	_, referenceErr := service.ExpandGraphRelationships(ctx, rez.ExpandKnowledgeGraphRelationshipsParams{})
	s.ErrorIs(referenceErr, errs.ErrInvalidInput)

	selected := s.createEntity(ctx, database, queryTestUUID(1), kne.CategoryContainer, "service")
	outside := s.createEntity(ctx, database, queryTestUUID(2), kne.CategoryContainer, "database")
	s.createEntity(ctx, database, queryTestUUID(3), kne.CategoryContainer, "service")
	s.createRelationship(ctx, database, queryTestUUID(10), selected.ID, outside.ID, knr.PredicateDependsOn)
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
	s.ErrorIs(entityCursorErr, errs.ErrInvalidInput)

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
		Predicates:         []knr.Predicate{knr.PredicateDependsOn},
		Cursor:             relationshipPage.NextCursor,
		Limit:              &relationshipLimit,
	})
	s.ErrorIs(relationshipCursorErr, errs.ErrInvalidInput)
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
		Authority:   schematypes.KnowledgeAuthorityDeclared,
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
		Authority:   schematypes.KnowledgeAuthorityDeclared,
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
		SetLastOccurredAt(now).
		SetReceivedAt(now).
		SetAttributes([]byte(`{}`))
	event, createErr := createEvent.Save(ctx)
	s.Require().NoError(createErr)

	ingestion, ingestionErr := NewKnowledgeGraphIngestionService(database, discardedObservationNotices{})
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

func (s *KnowledgeGraphQueryServiceSuite) TestListAncestors() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	commerce := s.structureEntityRef("commerce", kne.CategorySystemFunction)
	core := s.structureEntityRef("core", kne.CategorySystem)
	checkout := s.structureEntityRef("checkout", kne.CategoryContainer)
	orchestrator := s.structureEntityRef("orchestrator", kne.CategoryComponent)
	s.ingestStructure(ctx, database,
		s.observeEntity(commerce), s.observeEntity(core), s.observeEntity(checkout), s.observeEntity(orchestrator),
		s.observeContains(commerce, core), s.observeContains(core, checkout), s.observeContains(checkout, orchestrator),
	)
	commerceID, coreID := s.ingestedEntityID(ctx, database, commerce), s.ingestedEntityID(ctx, database, core)
	checkoutID, orchestratorID := s.ingestedEntityID(ctx, database, checkout), s.ingestedEntityID(ctx, database, orchestrator)

	ancestors, listErr := service.ListAncestors(ctx, []uuid.UUID{orchestratorID, commerceID})
	s.Require().NoError(listErr)
	s.Require().Len(ancestors[orchestratorID], 3)
	s.Equal([]uuid.UUID{checkoutID, coreID, commerceID}, []uuid.UUID{
		ancestors[orchestratorID][0].ID, ancestors[orchestratorID][1].ID, ancestors[orchestratorID][2].ID,
	}, "nearest first")
	s.NotContains(ancestors, commerceID, "a root has no ancestors")

	// The map reads the tree from the selection page.
	page := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	parents := make(map[uuid.UUID]*uuid.UUID)
	for _, entity := range page.Entities {
		parents[entity.ID] = entity.ParentEntityID()
	}
	s.Equal(&checkoutID, parents[orchestratorID])
	s.Nil(parents[commerceID])

	// A flag left pointing at a withdrawn ancestor ends the chain below it.
	database.Client(ctx).KnowledgeEntity.UpdateOneID(coreID).SetWithdrawnAt(time.Now()).ExecX(ctx)
	ancestors, listErr = service.ListAncestors(ctx, []uuid.UUID{orchestratorID})
	s.Require().NoError(listErr)
	s.Require().Len(ancestors[orchestratorID], 1)
	s.Equal(checkoutID, ancestors[orchestratorID][0].ID)

	// The structure lock keeps cycles out, so a cycle written around it is corrupt data, and ends the chain.
	database.Client(ctx).KnowledgeEntity.UpdateOneID(coreID).ClearWithdrawnAt().ExecX(ctx)
	createCycle := database.Client(ctx).KnowledgeRelationship.Create().
		SetPredicate(knr.PredicateContains).
		SetSourceEntityID(orchestratorID).
		SetTargetEntityID(commerceID).
		SetIsParent(true)
	createCycle.ExecX(ctx)
	ancestors, listErr = service.ListAncestors(ctx, []uuid.UUID{orchestratorID})
	s.Require().NoError(listErr)
	s.Require().Len(ancestors[orchestratorID], 3)
	s.Equal([]uuid.UUID{checkoutID, coreID, commerceID}, []uuid.UUID{
		ancestors[orchestratorID][0].ID, ancestors[orchestratorID][1].ID, ancestors[orchestratorID][2].ID,
	})
}

// TestSelectionListsOtherParentIDs reads, for each selected entity, the sources of the contains claims that were not
// chosen as its parent, oldest claim first.
func (s *KnowledgeGraphQueryServiceSuite) TestSelectionListsOtherParentIDs() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), database)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	core := graph.entity(kne.CategorySystem, "system", "provider-a", "system:core")
	experience := graph.entity(kne.CategorySystem, "system", "provider-a", "system:experience")
	commerce := graph.entity(kne.CategorySystem, "system", "provider-a", "system:commerce")
	checkoutAPI := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:checkout-api")
	checkoutAPI.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout-api"}
	checkout := graph.entity(kne.CategoryContainer, "service", "provider-b", "service:checkout")
	checkout.LinkingAttributes = projections.LinkingAttributes{projections.LinkingAttributeServiceName: "checkout"}
	graph.record(ctx, start,
		graph.observeEntity(core, "core"), graph.observeEntity(experience, "experience"),
		graph.observeEntity(commerce, "commerce"), graph.observeEntity(checkoutAPI, "checkout-api"),
		graph.observeRelationship(graph.relationship(knr.PredicateContains, core, checkoutAPI)),
	)
	graph.record(ctx, start.Add(time.Minute), graph.observeRelationship(graph.relationship(knr.PredicateContains, experience, checkoutAPI)))
	graph.record(ctx, start.Add(2*time.Minute), graph.observeRelationship(graph.relationship(knr.PredicateContains, commerce, checkoutAPI)))
	graph.record(ctx, start.Add(3*time.Minute),
		graph.observeEntity(checkout, "checkout"),
		graph.observeRelationship(graph.relationship(knr.PredicateContains, checkoutAPI, checkout)),
	)
	// Merging the two services turns the claim between them into the survivor's claim to contain itself.
	proposal := database.Client(ctx).KnowledgeProposal.Query().OnlyX(ctx)
	mergeParams := rez.MergeEntitiesParams{
		ProposalID: proposal.ID,
		EntityIDs:  [2]uuid.UUID{graph.entityID(ctx, checkoutAPI), graph.entityID(ctx, checkout)},
	}
	merged, mergeErr := graph.ingestion.MergeEntities(ctx, mergeParams)
	s.Require().NoError(mergeErr)
	s.Require().True(merged.Merged)
	checkoutID := merged.SurvivorID
	s.Require().True(database.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.SourceEntityID(checkoutID), knr.TargetEntityID(checkoutID)).
		ExistX(ctx))
	coreID, experienceID, commerceID := graph.entityID(ctx, core), graph.entityID(ctx, experience), graph.entityID(ctx, commerce)

	page := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})

	s.Equal([]uuid.UUID{experienceID, commerceID}, page.OtherParentIDs[checkoutID],
		"the chosen parent and a claim to contain itself are left out")
	s.NotContains(page.OtherParentIDs, coreID, "an entity without claims has no other parents")

	// A withdrawn source's claim is not current.
	database.Client(ctx).KnowledgeEntity.UpdateOneID(experienceID).SetWithdrawnAt(time.Now()).ExecX(ctx)
	page = s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	s.Equal([]uuid.UUID{commerceID}, page.OtherParentIDs[checkoutID])
}

func (s *KnowledgeGraphQueryServiceSuite) TestListAncestorsOfADeepChain() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	const depth = 20
	refs := make([]*rez.KnowledgeEntityRef, depth+1)
	var evidence []rez.KnowledgeEvidenceRef
	for level := range refs {
		refs[level] = s.structureEntityRef(fmt.Sprintf("component-%d", level), kne.CategoryComponent)
		evidence = append(evidence, s.observeEntity(refs[level]))
		if level > 0 {
			evidence = append(evidence, s.observeContains(refs[level-1], refs[level]))
		}
	}
	s.ingestStructure(ctx, database, evidence...)

	leafID := s.ingestedEntityID(ctx, database, refs[depth])
	ancestors, listErr := service.ListAncestors(ctx, []uuid.UUID{leafID})
	s.Require().NoError(listErr)
	s.Require().Len(ancestors[leafID], depth, "no depth cap")
	for i, ancestor := range ancestors[leafID] {
		s.Equal(s.ingestedEntityID(ctx, database, refs[depth-1-i]), ancestor.ID, "nearest first")
	}
}

func (s *KnowledgeGraphQueryServiceSuite) TestResolveRepresentatives() {
	ctx, database := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	checkout := s.structureEntityRef("checkout", kne.CategoryContainer)
	orders := s.structureEntityRef("orders", kne.CategoryComponent)
	pricing := s.structureEntityRef("pricing", kne.CategoryComponent)
	cluster := s.structureEntityRef("cluster", kne.CategoryInfrastructure)
	repository := s.structureEntityRef("repository", kne.CategoryCode)
	s.ingestStructure(ctx, database,
		s.observeEntity(checkout), s.observeEntity(orders), s.observeEntity(pricing),
		s.observeEntity(cluster), s.observeEntity(repository),
		s.observeContains(checkout, orders), s.observeContains(orders, pricing),
	)
	checkoutID, pricingID := s.ingestedEntityID(ctx, database, checkout), s.ingestedEntityID(ctx, database, pricing)
	clusterID, repositoryID := s.ingestedEntityID(ctx, database, cluster), s.ingestedEntityID(ctx, database, repository)

	resolved, resolveErr := service.ResolveRepresentatives(ctx, []uuid.UUID{pricingID, clusterID, repositoryID}, situations.RepresentativeCategories, 2)
	s.Require().NoError(resolveErr)
	s.Equal(map[uuid.UUID]uuid.UUID{pricingID: checkoutID, clusterID: clusterID}, resolved)

	resolved, resolveErr = service.ResolveRepresentatives(ctx, []uuid.UUID{pricingID}, situations.RepresentativeCategories, 1)
	s.Require().NoError(resolveErr)
	s.Empty(resolved, "the container is two steps away")

	withdrawCluster := s.observeEntity(cluster)
	withdrawCluster.Kind = ke.KindDeleted
	s.ingestStructure(ctx, database, withdrawCluster)
	resolved, resolveErr = service.ResolveRepresentatives(ctx, []uuid.UUID{clusterID}, situations.RepresentativeCategories, 2)
	s.Require().NoError(resolveErr)
	s.Empty(resolved, "a withdrawn entity represents nothing")
}

// testGraph records evidence through ingestion, as providers do.
type testGraph struct {
	require   *require.Assertions
	db        rez.Database
	ingestion *KnowledgeGraphIngestionService
}

func newTestGraph(require *require.Assertions, tdb rez.Database) *testGraph {
	ingestion, ingestionErr := NewKnowledgeGraphIngestionService(tdb, discardedObservationNotices{})
	require.NoError(ingestionErr)
	return &testGraph{require: require, db: tdb, ingestion: ingestion}
}

func (g *testGraph) entity(category kne.Category, kind, provider, resource string) rez.KnowledgeEntityRef {
	return rez.KnowledgeEntityRef{
		Category: category,
		Kind:     kind,
		ProviderResourceRef: rez.ProviderResourceRef{
			Provider:          provider,
			ProviderNamespace: "tests",
			ResourceRef:       resource,
		},
	}
}

func (g *testGraph) relationship(predicate knr.Predicate, source, target rez.KnowledgeEntityRef) rez.KnowledgeRelationshipRef {
	providerRef := source.ProviderResourceRef
	providerRef.ResourceRef = fmt.Sprintf("%s:%s:%s", source.ProviderResourceRef.ResourceRef, predicate, target.ProviderResourceRef.ResourceRef)
	return rez.KnowledgeRelationshipRef{
		Predicate:           predicate,
		ProviderResourceRef: providerRef,
		Source:              source,
		Target:              target,
	}
}

func (g *testGraph) observeEntity(ref rez.KnowledgeEntityRef, name string) rez.KnowledgeEvidenceRef {
	return rez.KnowledgeEvidenceRef{
		Authority:    schematypes.KnowledgeAuthorityDeclared,
		Kind:         ke.KindObserved,
		Assertion:    ref.Kind + "_observed",
		Subject:      rez.KnowledgeSubjectRef{Entity: &ref},
		SubjectState: schematypes.KnowledgeGraphSubjectState{DisplayName: name},
	}
}

func (g *testGraph) observeRelationship(ref rez.KnowledgeRelationshipRef) rez.KnowledgeEvidenceRef {
	return rez.KnowledgeEvidenceRef{
		Authority: schematypes.KnowledgeAuthorityDeclared,
		Kind:      ke.KindObserved,
		Assertion: string(ref.Predicate) + "_observed",
		Subject:   rez.KnowledgeSubjectRef{Relationship: &ref},
	}
}

// record ingests the evidence from one event effective at the given time.
func (g *testGraph) record(ctx context.Context, effectiveAt time.Time, evidence ...rez.KnowledgeEvidenceRef) {
	createEvent := g.db.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("tests").
		SetProviderResourceRef(uuid.NewString()).
		SetProviderEventSource("tests").
		SetProviderEventRef(uuid.NewString()).
		SetKind("test").
		SetOccurredAt(effectiveAt).
		SetLastOccurredAt(effectiveAt).
		SetReceivedAt(effectiveAt).
		SetAttributes([]byte(`{}`))
	event, eventErr := createEvent.Save(ctx)
	g.require.NoError(eventErr)
	for index := range evidence {
		evidence[index].EffectiveAt = effectiveAt
	}
	g.require.NoError(g.ingestion.IngestEvidence(ctx, event, evidence...))
}

// changePointEvent records an event of a change point with the given involvement, effective at the event's
// time, as a projection would before recording the change point.
func (g *testGraph) changePointEvent(ctx context.Context, resourceRef, kind string, at time.Time, attributes string, involvement ...rez.KnowledgeEvidenceRef) *ent.NormalizedEvent {
	createEvent := g.db.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("tests").
		SetProviderResourceRef(resourceRef).
		SetProviderEventSource("changes").
		SetProviderEventRef(uuid.NewString()).
		SetKind(kind).
		SetOccurredAt(at).
		SetLastOccurredAt(at).
		SetReceivedAt(at).
		SetAttributes([]byte(attributes))
	event, eventErr := createEvent.Save(ctx)
	g.require.NoError(eventErr)
	for index := range involvement {
		involvement[index].EffectiveAt = at
	}
	g.require.NoError(g.ingestion.IngestEvidence(ctx, event, involvement...))
	return event
}

func (g *testGraph) alias(ctx context.Context, ref rez.ProviderResourceRef) *ent.KnowledgeSubjectAlias {
	queryAlias := g.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.Provider(ref.Provider), ksa.ProviderNamespace(ref.ProviderNamespace), ksa.ProviderResourceRef(ref.ResourceRef))
	alias, aliasErr := queryAlias.Only(ctx)
	g.require.NoError(aliasErr)
	return alias
}

func (g *testGraph) entityID(ctx context.Context, ref rez.KnowledgeEntityRef) uuid.UUID {
	return *g.alias(ctx, ref.ProviderResourceRef).EntityID
}

func (g *testGraph) relationshipID(ctx context.Context, ref rez.KnowledgeRelationshipRef) uuid.UUID {
	return *g.alias(ctx, ref.ProviderResourceRef).RelationshipID
}

// latestEvidenceID returns the newest evidence recorded against the subject's alias.
func (g *testGraph) latestEvidenceID(ctx context.Context, ref rez.ProviderResourceRef) uuid.UUID {
	queryEvidence := g.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.SubjectAliasID(g.alias(ctx, ref).ID)).
		Order(ke.ByEffectiveAt(sql.OrderDesc()))
	evidence, evidenceErr := queryEvidence.First(ctx)
	g.require.NoError(evidenceErr)
	return evidence.ID
}

func (s *KnowledgeGraphQueryServiceSuite) TestGetEntityExpands() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	checkout.LinkingAttributes = testEntityLinkingAttributes{"service.name": "checkout-api"}
	graph.record(ctx, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC), graph.observeEntity(checkout, "checkout-api"))
	checkoutID := graph.entityID(ctx, checkout)

	plain, plainErr := service.GetEntity(ctx, checkoutID, rez.KnowledgeEntityExpand{})
	s.Require().NoError(plainErr)
	s.Require().Len(plain.Edges.Aliases, 1)
	s.Nil(plain.Edges.Aliases[0].Edges.Evidence)
	s.Nil(plain.Edges.LinkingAttributes)

	expanded, expandedErr := service.GetEntity(ctx, checkoutID, rez.KnowledgeEntityExpand{AliasEvidence: true, LinkingAttributes: true})
	s.Require().NoError(expandedErr)
	s.Require().Len(expanded.Edges.Aliases, 1)
	s.Len(expanded.Edges.Aliases[0].Edges.Evidence, 1)
	s.Require().Len(expanded.Edges.LinkingAttributes, 1)
	s.Equal("service.name", expanded.Edges.LinkingAttributes[0].Attribute)
}

func (s *KnowledgeGraphQueryServiceSuite) TestGetEvidenceExpandsSubject() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	database := graph.entity(kne.CategoryContainer, "database", "kubernetes", "orders-db")
	dependsOn := graph.relationship(knr.PredicateDependsOn, checkout, database)
	graph.record(ctx, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		graph.observeEntity(checkout, "checkout-api"),
		graph.observeEntity(database, "orders-db"),
		graph.observeRelationship(dependsOn))

	entityEvidenceID := graph.latestEvidenceID(ctx, checkout.ProviderResourceRef)
	plain, plainErr := service.GetEvidence(ctx, entityEvidenceID, rez.KnowledgeEvidenceExpand{})
	s.Require().NoError(plainErr)
	s.Nil(plain.Edges.SubjectAlias.Edges.Entity)

	onEntity, entityErr := service.GetEvidence(ctx, entityEvidenceID, rez.KnowledgeEvidenceExpand{Subject: true})
	s.Require().NoError(entityErr)
	s.Equal(graph.entityID(ctx, checkout), onEntity.Edges.SubjectAlias.Edges.Entity.ID)

	relationshipEvidenceID := graph.latestEvidenceID(ctx, dependsOn.ProviderResourceRef)
	onRelationship, relationshipErr := service.GetEvidence(ctx, relationshipEvidenceID, rez.KnowledgeEvidenceExpand{Subject: true})
	s.Require().NoError(relationshipErr)
	relationship := onRelationship.Edges.SubjectAlias.Edges.Relationship
	s.Require().NotNil(relationship)
	s.Equal(graph.entityID(ctx, checkout), relationship.Edges.SourceEntity.ID)
	s.Equal(graph.entityID(ctx, database), relationship.Edges.TargetEntity.ID)
}

func (s *KnowledgeGraphQueryServiceSuite) TestListEntityRelationshipsByNewestEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	t0 := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	database := graph.entity(kne.CategoryContainer, "database", "kubernetes", "orders-db")
	team := graph.entity(kne.CategoryActor, "team", "github", "payments")
	queue := graph.entity(kne.CategoryContainer, "queue", "kubernetes", "orders-queue")
	dependsOnDatabase := graph.relationship(knr.PredicateDependsOn, checkout, database)
	ownsCheckout := graph.relationship(knr.PredicateOwns, team, checkout)
	publishesToQueue := graph.relationship(knr.PredicatePublishesTo, checkout, queue)

	graph.record(ctx, t0,
		graph.observeEntity(checkout, "checkout-api"),
		graph.observeEntity(queue, "orders-queue"),
		graph.observeRelationship(publishesToQueue))
	graph.record(ctx, t0.Add(time.Minute),
		graph.observeEntity(database, "orders-db"),
		graph.observeRelationship(dependsOnDatabase))
	graph.record(ctx, t0.Add(2*time.Minute),
		graph.observeRelationship(dependsOnDatabase))
	graph.record(ctx, t0.Add(3*time.Minute),
		graph.observeEntity(team, "Payments"),
		graph.observeRelationship(ownsCheckout))
	checkoutID := graph.entityID(ctx, checkout)
	params := rez.ListKnowledgeRelationshipsParams{
		ListParams:          ent.ListParams{Page: 1, PageSize: 2},
		EntityID:            checkoutID,
		NewestEvidenceFirst: true,
		Expand:              rez.KnowledgeRelationshipExpand{Endpoints: true},
	}
	firstPage, firstErr := service.ListRelationships(ctx, params)
	s.Require().NoError(firstErr)
	s.Equal(3, firstPage.Total, "all current relationships are counted")
	s.Require().Len(firstPage.Data, 2)
	s.Equal(graph.relationshipID(ctx, ownsCheckout), firstPage.Data[0].ID)
	s.Equal(graph.entityID(ctx, team), firstPage.Data[0].Edges.SourceEntity.ID, "an incoming relationship")
	s.Equal(graph.relationshipID(ctx, dependsOnDatabase), firstPage.Data[1].ID)

	params.Page = 2
	secondPage, secondErr := service.ListRelationships(ctx, params)
	s.Require().NoError(secondErr)
	s.Require().Len(secondPage.Data, 1)
	s.Equal(graph.relationshipID(ctx, publishesToQueue), secondPage.Data[0].ID)

}

func (s *KnowledgeGraphQueryServiceSuite) TestSummarizeEntityEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	t0 := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	alert := graph.entity(kne.CategorySignal, "alert", "alertmanager", "CheckoutErrors")
	database := graph.entity(kne.CategoryContainer, "database", "kubernetes", "orders-db")
	alertObserves := graph.relationship(knr.PredicateObserves, alert, checkout)
	databaseDependency := graph.relationship(knr.PredicateDependsOn, checkout, database)

	graph.record(ctx, t0, graph.observeEntity(checkout, "checkout-api"))
	for firing := 1; firing <= 3; firing++ {
		graph.record(ctx, t0.Add(time.Duration(firing)*time.Hour),
			graph.observeEntity(alert, "CheckoutErrors"),
			graph.observeRelationship(alertObserves))
	}
	graph.record(ctx, t0.Add(150*time.Minute),
		graph.observeEntity(database, "orders-db"),
		graph.observeRelationship(databaseDependency))

	params := rez.SummarizeKnowledgeEvidenceParams{
		ListParams:  ent.ListParams{Page: 1, PageSize: 2},
		EntityID:    graph.entityID(ctx, checkout),
		ByAssertion: true,
		Expand:      rez.KnowledgeEvidenceSummaryExpand{Alias: true},
	}
	summaries, summarizeErr := service.SummarizeEvidence(ctx, params)
	s.Require().NoError(summarizeErr)
	s.Equal(3, summaries.Total, "the alert's repeats, the dependency and the service itself")
	s.Require().Len(summaries.Data, 2)

	alertSummary := summaries.Data[0]
	s.Equal(graph.relationshipID(ctx, alertObserves), alertSummary.Alias.Edges.Relationship.ID)
	s.Equal("observes_observed", alertSummary.Assertion)
	s.Equal(3, alertSummary.Count)
	s.Equal(t0.Add(time.Hour), alertSummary.EarliestAt)
	s.Equal(t0.Add(3*time.Hour), alertSummary.LatestAt)
	s.Equal(graph.latestEvidenceID(ctx, alertObserves.ProviderResourceRef), alertSummary.LatestID)
	s.Equal("alertmanager", alertSummary.Alias.Provider)

	deploymentSummary := summaries.Data[1]
	s.Equal(graph.relationshipID(ctx, databaseDependency), deploymentSummary.Alias.Edges.Relationship.ID)
	s.Equal(1, deploymentSummary.Count)
	s.Equal(graph.entityID(ctx, database), deploymentSummary.Alias.Edges.Relationship.Edges.TargetEntity.ID)

	params.ListParams = ent.ListParams{}
	params.Start = t0.Add(2 * time.Hour)
	params.End = t0.Add(150 * time.Minute)
	bounded, boundedErr := service.SummarizeEvidence(ctx, params)
	s.Require().NoError(boundedErr)
	s.Require().Len(bounded.Data, 2)
	s.Equal(deploymentSummary.SubjectAliasID, bounded.Data[0].SubjectAliasID)
	s.Equal(alertSummary.SubjectAliasID, bounded.Data[1].SubjectAliasID)
	s.Equal(1, bounded.Data[1].Count, "only the alert's second firing is in the window")
}

// A merge leaves the merged entity and the relationship it folded as redirects. Neither is current, get by ID
// returns each with its redirect, and a read keyed by the merged entity's ID reads the survivor's.
func (s *KnowledgeGraphQueryServiceSuite) TestMergedSubjectsRedirect() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)

	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	alert := graph.entity(kne.CategorySignal, "alert", "alertmanager", "CheckoutErrors")
	observes := graph.relationship(knr.PredicateObserves, alert, checkout)
	graph.record(ctx, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		graph.observeEntity(checkout, "checkout-api"),
		graph.observeEntity(alert, "Checkout errors"),
		graph.observeRelationship(observes))
	survivorID := graph.entityID(ctx, checkout)
	alertID := graph.entityID(ctx, alert)
	keptID := graph.relationshipID(ctx, observes)

	createMerged := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	merged := createMerged.SaveX(ctx)
	createFolded := client.KnowledgeRelationship.Create().
		SetPredicate(knr.PredicateObserves).
		SetSourceEntityID(alertID).
		SetTargetEntityID(merged.ID)
	folded := createFolded.SaveX(ctx)
	client.KnowledgeRelationship.UpdateOneID(folded.ID).SetMergedIntoID(keptID).ExecX(ctx)
	client.KnowledgeEntity.UpdateOneID(merged.ID).SetMergedIntoID(survivorID).ExecX(ctx)

	for _, includeWithdrawn := range []bool{false, true} {
		entities, entitiesErr := service.ListEntities(ctx, rez.ListKnowledgeEntitiesParams{IncludeWithdrawn: includeWithdrawn})
		s.Require().NoError(entitiesErr)
		entityIDs := make([]uuid.UUID, len(entities.Data))
		for i, entity := range entities.Data {
			entityIDs[i] = entity.ID
		}
		s.ElementsMatch([]uuid.UUID{survivorID, alertID}, entityIDs, "include withdrawn %t", includeWithdrawn)

		relationships, relationshipsErr := service.ListRelationships(ctx, rez.ListKnowledgeRelationshipsParams{IncludeWithdrawn: includeWithdrawn})
		s.Require().NoError(relationshipsErr)
		s.Require().Len(relationships.Data, 1, "include withdrawn %t", includeWithdrawn)
		s.Equal(keptID, relationships.Data[0].ID)
	}
	selected := s.selectEntities(ctx, service, rez.SelectKnowledgeGraphEntitiesParams{})
	s.Len(selected.Entities, 2, "selection leaves the merged entity out")

	got, getErr := service.GetEntity(ctx, merged.ID, rez.KnowledgeEntityExpand{})
	s.Require().NoError(getErr)
	s.Equal(&survivorID, got.MergedIntoID)
	gotRelationship, getRelationshipErr := service.GetRelationship(ctx, folded.ID)
	s.Require().NoError(getRelationshipErr)
	s.Equal(&keptID, gotRelationship.MergedIntoID)

	byMerged, byMergedErr := service.ListRelationships(ctx, rez.ListKnowledgeRelationshipsParams{EntityID: merged.ID})
	s.Require().NoError(byMergedErr)
	s.Require().Len(byMerged.Data, 1)
	s.Equal(keptID, byMerged.Data[0].ID, "the merged entity's relationships are the survivor's")

	summarize := func(entityID uuid.UUID) *ent.ListResult[rez.KnowledgeEvidenceSummary] {
		summaries, summarizeErr := service.SummarizeEvidence(ctx, rez.SummarizeKnowledgeEvidenceParams{EntityID: entityID})
		s.Require().NoError(summarizeErr)
		return summaries
	}
	survivorSummaries := summarize(survivorID)
	s.Equal(2, survivorSummaries.Total)
	s.Equal(survivorSummaries.Data, summarize(merged.ID).Data)
}

func (s *KnowledgeGraphQueryServiceSuite) TestSummarizeAliasEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	t0 := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	graph.record(ctx, t0, graph.observeEntity(checkout, "checkout"))
	checkoutRestarted := graph.observeEntity(checkout, "checkout-api")
	checkoutRestarted.Assertion = "service_restarted"
	graph.record(ctx, t0.Add(time.Minute), checkoutRestarted)
	aliasID := graph.alias(ctx, checkout.ProviderResourceRef).ID

	summaries, summarizeErr := service.SummarizeEvidence(ctx, rez.SummarizeKnowledgeEvidenceParams{AliasIDs: []uuid.UUID{aliasID}})
	s.Require().NoError(summarizeErr)
	s.Equal(1, summaries.Total, "assertions are summarized together")
	s.Require().Len(summaries.Data, 1)
	summary := summaries.Data[0]
	s.Equal(aliasID, summary.SubjectAliasID)
	s.Nil(summary.Alias)
	s.Empty(summary.Assertion)
	s.Equal(2, summary.Count)
	s.Equal(t0, summary.EarliestAt)
	s.Equal(t0.Add(time.Minute), summary.LatestAt)
	s.Equal("checkout-api", summary.LatestState.DisplayName)

	nothing, nothingErr := service.SummarizeEvidence(ctx, rez.SummarizeKnowledgeEvidenceParams{})
	s.Require().NoError(nothingErr)
	s.Zero(nothing.Total)
	s.Empty(nothing.Data)
}

func (s *KnowledgeGraphQueryServiceSuite) TestSummarizeEvidenceIsTenantScoped() {
	ownerCtx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	graph.record(ownerCtx, time.Now().UTC(), graph.observeEntity(checkout, "checkout-api"))
	checkoutID := graph.entityID(ownerCtx, checkout)
	aliasID := graph.alias(ownerCtx, checkout.ProviderResourceRef).ID
	otherTenantCtx, _ := s.SetupTestDatabase()

	owned, ownerErr := service.SummarizeEvidence(ownerCtx, rez.SummarizeKnowledgeEvidenceParams{EntityID: checkoutID})
	s.Require().NoError(ownerErr)
	s.Equal(1, owned.Total)

	_, entityErr := service.SummarizeEvidence(otherTenantCtx, rez.SummarizeKnowledgeEvidenceParams{EntityID: checkoutID})
	s.ErrorIs(entityErr, errs.ErrNotFound)
	byAlias, aliasErr := service.SummarizeEvidence(otherTenantCtx, rez.SummarizeKnowledgeEvidenceParams{AliasIDs: []uuid.UUID{aliasID}})
	s.Require().NoError(aliasErr)
	s.Zero(byAlias.Total)
}

func (s *KnowledgeGraphQueryServiceSuite) TestListEntitiesByLinkingAttribute() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	payments := graph.entity(kne.CategoryContainer, "service", "kubernetes", "payments-api")
	payments.LinkingAttributes = testEntityLinkingAttributes{"service.name": "payments-api"}
	checkout := graph.entity(kne.CategoryContainer, "service", "kubernetes", "checkout-api")
	checkout.LinkingAttributes = testEntityLinkingAttributes{"service.name": "checkout-api"}
	paymentsDatabase := graph.entity(kne.CategoryContainer, "database", "kubernetes", "payments-api-db")
	alert := graph.entity(kne.CategorySignal, "alert", "provider-a", "payments-alert")
	alert.LinkingAttributes = testEntityLinkingAttributes{"service.name": "payments-alert"}
	graph.record(ctx, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		graph.observeEntity(payments, "payments-api"),
		graph.observeEntity(paymentsDatabase, "payments-api"),
		graph.observeEntity(alert, "payments-api"),
		graph.observeEntity(checkout, "checkout-api"))

	find := func(contains string) []uuid.UUID {
		params := rez.ListKnowledgeEntitiesParams{
			Categories:           []kne.Category{kne.CategoryContainer},
			LinkingAttribute:     "service.name",
			LinkingValueContains: contains,
		}
		listed, listErr := service.ListEntities(ctx, params)
		s.Require().NoError(listErr)
		ids := make([]uuid.UUID, len(listed.Data))
		for index, entity := range listed.Data {
			ids[index] = entity.ID
		}
		return ids
	}

	paymentsID, checkoutID := graph.entityID(ctx, payments), graph.entityID(ctx, checkout)
	s.Equal([]uuid.UUID{paymentsID}, find("payments-api"),
		"only the entity holding the attribute, not one named alike without it")
	s.Equal([]uuid.UUID{paymentsID}, find("payments"), "a value containing the text")
	s.ElementsMatch([]uuid.UUID{paymentsID, checkoutID}, find(""), "every entity holding the attribute")
	s.Empty(find("payments-alert"), "an entity of another category is not listed")
	s.Empty(find("orders-api"))
}

func (s *KnowledgeGraphQueryServiceSuite) TestListEntitiesSearch() {
	ctx, tdb := s.SetupTestDatabase()
	service, serviceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	checkout := graph.entity(kne.CategoryContainer, "service", "provider-a", "checkout-api")
	checkout.LinkingAttributes = testEntityLinkingAttributes{"service.name": "checkout-api"}
	payments := graph.entity(kne.CategoryContainer, "service", "provider-a", "payments-svc")
	payments.LinkingAttributes = testEntityLinkingAttributes{"service.name": "payments-svc"}
	graph.record(ctx, time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		graph.observeEntity(checkout, "Checkout API"),
		graph.observeEntity(payments, "Payments"),
	)
	search := func(text string) []*ent.KnowledgeEntity {
		params := rez.ListKnowledgeEntitiesParams{
			ListParams:            ent.ListParams{Search: text},
			WithLinkingAttributes: true,
		}
		result, listErr := service.ListEntities(ctx, params)
		s.Require().NoError(listErr)
		return result.Data
	}

	byDisplayName := search("CHECKOUT a")
	byLinkingValue := search("SVC")
	byNeither := search("inventory")

	s.Require().Len(byDisplayName, 1)
	s.Equal(graph.entityID(ctx, checkout), byDisplayName[0].ID)
	s.Require().Len(byDisplayName[0].Edges.LinkingAttributes, 1)
	s.Equal("checkout-api", byDisplayName[0].Edges.LinkingAttributes[0].Value)
	s.Require().Len(byLinkingValue, 1)
	s.Equal(graph.entityID(ctx, payments), byLinkingValue[0].ID)
	s.Empty(byNeither)
}

func (s *KnowledgeGraphQueryServiceSuite) TestListEntityAliases() {
	ctx, tdb := s.SetupTestDatabase()
	registry, registryErr := integrations.NewRegistry(&testIntegration{name: "provider-a"})
	s.Require().NoError(registryErr)
	service, serviceErr := NewKnowledgeEntityAliasService(tdb, registry)
	s.Require().NoError(serviceErr)
	graph := newTestGraph(s.Require(), tdb)
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	fromAlerts := graph.entity(kne.CategoryContainer, "service", "provider-a", "service:checkout-api")
	fromAlerts.LinkingAttributes = testEntityLinkingAttributes{"service.name": "checkout-api"}
	fromLogs := graph.entity(kne.CategoryContainer, "service", "provider-b", "logs:service_name:checkout-api")
	fromLogs.LinkingAttributes = testEntityLinkingAttributes{"service.name": "checkout-api"}
	graph.record(ctx, start, graph.observeEntity(fromAlerts, "checkout-api"))
	graph.record(ctx, start.Add(time.Hour), graph.observeEntity(fromLogs, ""))
	graph.record(ctx, start.Add(2*time.Hour), graph.observeEntity(fromAlerts, "checkout-api"))
	entityID := graph.entityID(ctx, fromAlerts)

	aliases, aliasesErr := service.ListEntityAliases(ctx, []uuid.UUID{entityID})

	s.Require().NoError(aliasesErr)
	s.Require().Len(aliases[entityID], 2)
	newest, older := aliases[entityID][0], aliases[entityID][1]
	s.Equal("provider-a", newest.Alias.Provider)
	s.Equal("Test Package", newest.IntegrationDisplayName)
	s.Equal("checkout-api", newest.Name)
	s.True(start.Add(2 * time.Hour).Equal(newest.LastRecordedAt))
	s.Nil(newest.Alias.Edges.MergedByProposal)
	s.Equal("provider-b", older.Alias.Provider)
	s.Empty(older.IntegrationDisplayName, "a provider with no registered integration has no display name")
	s.Equal("logs:service_name:checkout-api", older.Name, "an alias stating no name is named by its resource ref")
	s.True(start.Add(time.Hour).Equal(older.LastRecordedAt))
}

// TestGetEntityWithAliasesReadsParentClaims reads an entity with a parent claim and another claim from a source
// that has a parent of its own, through the API response.
func (s *KnowledgeGraphQueryServiceSuite) TestGetEntityWithAliasesReadsParentClaims() {
	ctx, tdb := s.SetupTestDatabase()
	registry, registryErr := integrations.NewRegistry()
	s.Require().NoError(registryErr)
	service, serviceErr := NewKnowledgeEntityAliasService(tdb, registry)
	s.Require().NoError(serviceErr)
	commerce := s.structureEntityRef("commerce", kne.CategorySystemFunction)
	core := s.structureEntityRef("core", kne.CategorySystem)
	experience := s.structureEntityRef("experience", kne.CategorySystem)
	checkout := s.structureEntityRef("checkout", kne.CategoryContainer)
	s.ingestStructure(ctx, tdb,
		s.observeEntity(commerce), s.observeEntity(core), s.observeEntity(experience), s.observeEntity(checkout),
		s.observeContains(commerce, experience),
	)
	s.ingestStructure(ctx, tdb, s.observeContains(core, checkout))
	s.ingestStructure(ctx, tdb, s.observeContains(experience, checkout))
	commerceID, coreID := s.ingestedEntityID(ctx, tdb, commerce), s.ingestedEntityID(ctx, tdb, core)
	experienceID, checkoutID := s.ingestedEntityID(ctx, tdb, experience), s.ingestedEntityID(ctx, tdb, checkout)

	read, readErr := service.GetEntityWithAliases(ctx, checkoutID, rez.KnowledgeEntityExpand{})

	s.Require().NoError(readErr)
	response := v1.KnowledgeGraphEntityWithAliasesFromRez(read, time.Now()).Attributes
	s.Equal(&coreID, response.ParentId)
	s.NotNil(response.ParentRelationshipId)
	s.Equal([]string{"test"}, response.ParentProviders)
	s.Require().Len(response.OtherParentClaims, 1)
	other := response.OtherParentClaims[0]
	s.Equal(experienceID, other.Parent.ID)
	s.Equal(&commerceID, other.Parent.ParentId, "the other claim's source keeps its own parent")
	s.Equal([]string{"test"}, other.Providers)
}
