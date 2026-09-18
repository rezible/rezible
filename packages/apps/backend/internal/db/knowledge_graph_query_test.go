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
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
)

type KnowledgeGraphQueryServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphQueryServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphQueryServiceSuite{Suite: test.NewSuite()})
}

type knowledgeGraphQueryFixture struct {
	database rez.Database
	ctx      context.Context
	service  *KnowledgeGraphQueryService

	a *ent.KnowledgeEntity
	b *ent.KnowledgeEntity
	x *ent.KnowledgeEntity
	y *ent.KnowledgeEntity
	u *ent.KnowledgeEntity

	aContainsX *ent.KnowledgeRelationship
	bContainsY *ent.KnowledgeRelationship
	aCallsB    *ent.KnowledgeRelationship
	xCallsY    *ent.KnowledgeRelationship
}

func (s *KnowledgeGraphQueryServiceSuite) newFixture() knowledgeGraphQueryFixture {
	ctx := s.SeedTenantContext()
	database := s.CreateTestDatabase()
	client := database.Client(ctx)
	createEntity := func(category kne.Category, kind string) *ent.KnowledgeEntity {
		entityBuilder := client.KnowledgeEntity.Create().
			SetCategory(category).
			SetKind(kind)
		entity, saveErr := entityBuilder.Save(ctx)
		s.Require().NoError(saveErr)
		return entity
	}
	createRelationship := func(source, target *ent.KnowledgeEntity, predicate knr.Predicate) *ent.KnowledgeRelationship {
		relationshipBuilder := client.KnowledgeRelationship.Create().
			SetSourceEntityID(source.ID).
			SetTargetEntityID(target.ID).
			SetPredicate(predicate)
		relationship, saveErr := relationshipBuilder.Save(ctx)
		s.Require().NoError(saveErr)
		return relationship
	}
	createAlias := func(subjectKind ksa.SubjectKind, entityID, relationshipID uuid.UUID, resourceRef string) {
		aliasBuilder := client.KnowledgeSubjectAlias.Create().
			SetSubjectKind(subjectKind).
			SetProvider("test").
			SetProviderNamespace("knowledge-graph-query-tests").
			SetProviderResourceRef(resourceRef)
		if subjectKind == ksa.SubjectKindEntity {
			aliasBuilder.SetEntityID(entityID)
		} else {
			aliasBuilder.SetRelationshipID(relationshipID)
		}
		aliasErr := aliasBuilder.Exec(ctx)
		s.Require().NoError(aliasErr)
	}

	fixture := knowledgeGraphQueryFixture{
		database: database,
		ctx:      ctx,
		a:        createEntity(kne.CategorySystem, "a"),
		b:        createEntity(kne.CategorySystem, "b"),
		x:        createEntity(kne.CategoryContainer, "x"),
		y:        createEntity(kne.CategoryContainer, "y"),
		u:        createEntity(kne.CategorySystem, "unrelated"),
	}
	fixture.aContainsX = createRelationship(fixture.a, fixture.x, knr.PredicateContains)
	fixture.bContainsY = createRelationship(fixture.b, fixture.y, knr.PredicateContains)
	fixture.aCallsB = createRelationship(fixture.a, fixture.b, knr.PredicateCalls)
	fixture.xCallsY = createRelationship(fixture.x, fixture.y, knr.PredicateCalls)
	for _, entity := range []*ent.KnowledgeEntity{fixture.a, fixture.b, fixture.x, fixture.y, fixture.u} {
		createAlias(ksa.SubjectKindEntity, entity.ID, uuid.Nil, "entity:"+entity.Kind)
	}
	for _, relationship := range []*ent.KnowledgeRelationship{
		fixture.aContainsX, fixture.bContainsY, fixture.aCallsB, fixture.xCallsY,
	} {
		createAlias(ksa.SubjectKindRelationship, uuid.Nil, relationship.ID, "relationship:"+relationship.ID.String())
	}
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)
	fixture.service = service
	return fixture
}

func graphQueryFor(level rez.KnowledgeGraphDetailLevel, rootIDs ...uuid.UUID) rez.KnowledgeGraphLevelQuery {
	return rez.KnowledgeGraphLevelQuery{
		Root: rez.KnowledgeGraphQueryRoot{
			Lens: rez.KnowledgeGraphQueryLens{EntityIDs: rootIDs},
		},
		Level: level,
	}
}

func entityIDSet(entities ent.KnowledgeEntities) map[uuid.UUID]struct{} {
	ids := make(map[uuid.UUID]struct{}, len(entities))
	for _, entity := range entities {
		ids[entity.ID] = struct{}{}
	}
	return ids
}

func connectionCountSet(connections []rez.KnowledgeGraphConnectionAggregate) map[rez.KnowledgeGraphConnection]int {
	counts := make(map[rez.KnowledgeGraphConnection]int, len(connections))
	for _, connection := range connections {
		counts[connection.Key] += connection.RelationshipCount
	}
	return counts
}

func (s *KnowledgeGraphQueryServiceSuite) TestDetailSelectionAndDirectConnections() {
	fixture := s.newFixture()
	systems, systemsErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID, fixture.a.ID, fixture.b.ID), rez.KnowledgeGraphPageParams{})
	s.Require().NoError(systemsErr)
	s.Equal(map[uuid.UUID]struct{}{fixture.a.ID: {}, fixture.b.ID: {}}, entityIDSet(systems.Entities))
	s.Equal(map[rez.KnowledgeGraphConnection]int{{
		SourceRepresentativeID: fixture.a.ID,
		TargetRepresentativeID: fixture.b.ID,
		Predicate:              knr.PredicateCalls,
	}: 1}, connectionCountSet(systems.Connections))

	runtime, runtimeErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelRuntime, fixture.a.ID, fixture.b.ID), rez.KnowledgeGraphPageParams{})
	s.Require().NoError(runtimeErr)
	s.Equal(map[uuid.UUID]struct{}{
		fixture.a.ID: {}, fixture.b.ID: {}, fixture.x.ID: {}, fixture.y.ID: {},
	}, entityIDSet(runtime.Entities))
	runtimeConnections := connectionCountSet(runtime.Connections)
	s.Equal(1, runtimeConnections[rez.KnowledgeGraphConnection{
		SourceRepresentativeID: fixture.a.ID,
		TargetRepresentativeID: fixture.b.ID,
		Predicate:              knr.PredicateCalls,
	}])
	s.Equal(1, runtimeConnections[rez.KnowledgeGraphConnection{
		SourceRepresentativeID: fixture.x.ID,
		TargetRepresentativeID: fixture.y.ID,
		Predicate:              knr.PredicateCalls,
	}])
	s.Equal(1, runtimeConnections[rez.KnowledgeGraphConnection{
		SourceRepresentativeID: fixture.a.ID,
		TargetRepresentativeID: fixture.x.ID,
		Predicate:              knr.PredicateContains,
	}])
	s.Equal(1, runtimeConnections[rez.KnowledgeGraphConnection{
		SourceRepresentativeID: fixture.b.ID,
		TargetRepresentativeID: fixture.y.ID,
		Predicate:              knr.PredicateContains,
	}])
	s.NotContains(entityIDSet(runtime.Entities), fixture.u.ID)
}

func (s *KnowledgeGraphQueryServiceSuite) TestMembershipTraversalTerminatesForSharedCycle() {
	ctx := s.SeedTenantContext()
	database := s.CreateTestDatabase()
	client := database.Client(ctx)
	createEntity := func(kind string) *ent.KnowledgeEntity {
		entityBuilder := client.KnowledgeEntity.Create().SetCategory(kne.CategorySystem).SetKind(kind)
		entity, saveErr := entityBuilder.Save(ctx)
		s.Require().NoError(saveErr)
		return entity
	}
	createContains := func(source, target *ent.KnowledgeEntity) {
		relationshipBuilder := client.KnowledgeRelationship.Create().
			SetSourceEntityID(source.ID).
			SetTargetEntityID(target.ID).
			SetPredicate(knr.PredicateContains)
		s.Require().NoError(relationshipBuilder.Exec(ctx))
	}
	p := createEntity("p")
	q := createEntity("q")
	r := createEntity("r")
	createContains(p, q)
	createContains(r, q)
	createContains(q, p)
	service, serviceErr := NewKnowledgeGraphQueryService(database)
	s.Require().NoError(serviceErr)

	result, queryErr := service.QueryGraphLevel(ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, p.ID, r.ID), rez.KnowledgeGraphPageParams{})
	s.Require().NoError(queryErr)
	s.Equal(map[uuid.UUID]struct{}{p.ID: {}, q.ID: {}, r.ID: {}}, entityIDSet(result.Entities))
	s.Len(result.Entities, 3)
	s.Len(result.Connections, 3)
}

func (s *KnowledgeGraphQueryServiceSuite) TestPaginationIncludesConnectionEndpoints() {
	fixture := s.newFixture()
	query := graphQueryFor(rez.KnowledgeGraphDetailLevelRuntime, fixture.a.ID, fixture.b.ID)
	largePage, largePageErr := fixture.service.QueryGraphLevel(fixture.ctx, query, rez.KnowledgeGraphPageParams{Limit: 100})
	s.Require().NoError(largePageErr)

	mergedEntities := make(map[uuid.UUID]struct{})
	mergedConnections := make(map[rez.KnowledgeGraphConnection]int)
	pageParams := rez.KnowledgeGraphPageParams{Limit: 1}
	for {
		page, pageErr := fixture.service.QueryGraphLevel(fixture.ctx, query, pageParams)
		s.Require().NoError(pageErr)
		pageEntityIDs := entityIDSet(page.Entities)
		for entityID := range pageEntityIDs {
			mergedEntities[entityID] = struct{}{}
		}
		for _, connection := range page.Connections {
			mergedConnections[connection.Key] += connection.RelationshipCount
			_, sourceIncluded := pageEntityIDs[connection.Key.SourceRepresentativeID]
			_, targetIncluded := pageEntityIDs[connection.Key.TargetRepresentativeID]
			s.True(sourceIncluded)
			s.True(targetIncluded)
		}
		if page.NextCursor == "" {
			break
		}
		pageParams.Cursor = page.NextCursor
	}
	s.Equal(entityIDSet(largePage.Entities), mergedEntities)
	s.Equal(connectionCountSet(largePage.Connections), mergedConnections)
	for _, connection := range largePage.Connections {
		s.Equal(1, connection.RelationshipCount)
	}
}

func (s *KnowledgeGraphQueryServiceSuite) TestInspectionUsesSelectedCanonicalRelationshipAndEdges() {
	fixture := s.newFixture()
	params := rez.ListKnowledgeGraphConnectionAggregateRelationshipsParams{
		Query: graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID, fixture.b.ID),
		Connection: rez.KnowledgeGraphConnection{
			SourceRepresentativeID: fixture.a.ID,
			TargetRepresentativeID: fixture.b.ID,
			Predicate:              knr.PredicateCalls,
		},
	}
	list, listErr := fixture.service.ListKnowledgeGraphConnectionAggregateRelationships(fixture.ctx, params)
	s.Require().NoError(listErr)
	s.Require().Len(list.Data, 1)
	s.Equal(fixture.aCallsB.ID, list.Data[0].ID)
	s.Equal(fixture.a.ID, list.Data[0].SourceEntityID)
	s.Equal(fixture.b.ID, list.Data[0].TargetEntityID)
	sourceEntity, sourceErr := list.Data[0].Edges.SourceEntityOrErr()
	s.Require().NoError(sourceErr)
	targetEntity, targetErr := list.Data[0].Edges.TargetEntityOrErr()
	s.Require().NoError(targetErr)
	s.Equal(fixture.a.ID, sourceEntity.ID)
	s.Equal(fixture.b.ID, targetEntity.ID)
	sourceAliases, sourceAliasesErr := sourceEntity.Edges.AliasesOrErr()
	s.Require().NoError(sourceAliasesErr)
	targetAliases, targetAliasesErr := targetEntity.Edges.AliasesOrErr()
	s.Require().NoError(targetAliasesErr)
	s.NotEmpty(sourceAliases)
	s.NotEmpty(targetAliases)

	hiddenParams := params
	hiddenParams.Connection = rez.KnowledgeGraphConnection{
		SourceRepresentativeID: fixture.x.ID,
		TargetRepresentativeID: fixture.y.ID,
		Predicate:              knr.PredicateCalls,
	}
	hidden, hiddenErr := fixture.service.ListKnowledgeGraphConnectionAggregateRelationships(fixture.ctx, hiddenParams)
	s.Require().NoError(hiddenErr)
	s.Empty(hidden.Data)

	visibleParams := hiddenParams
	visibleParams.Query.Level = rez.KnowledgeGraphDetailLevelRuntime
	visible, visibleErr := fixture.service.ListKnowledgeGraphConnectionAggregateRelationships(fixture.ctx, visibleParams)
	s.Require().NoError(visibleErr)
	s.Require().Len(visible.Data, 1)
	s.Equal(fixture.xCallsY.ID, visible.Data[0].ID)

	pageTwoParams := visibleParams
	pageTwoParams.Page = 2
	pageTwoParams.PageSize = 1
	pageTwo, pageTwoErr := fixture.service.ListKnowledgeGraphConnectionAggregateRelationships(fixture.ctx, pageTwoParams)
	s.Require().NoError(pageTwoErr)
	s.Empty(pageTwo.Data)
	s.Equal(1, pageTwo.Total)
}

func (s *KnowledgeGraphQueryServiceSuite) TestValidationAndEmptyInput() {
	fixture := s.newFixture()
	empty, emptyErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems), rez.KnowledgeGraphPageParams{})
	s.Require().NoError(emptyErr)
	s.NotNil(empty.Entities)
	s.NotNil(empty.Connections)
	s.Empty(empty.Entities)
	s.Empty(empty.Connections)

	_, invalidLevelErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevel(4), fixture.a.ID), rez.KnowledgeGraphPageParams{})
	s.ErrorIs(invalidLevelErr, rez.ErrInvalidInput)
	_, nilRootErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, uuid.Nil), rez.KnowledgeGraphPageParams{})
	s.ErrorIs(nilRootErr, rez.ErrInvalidInput)
	_, missingRootErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, uuid.New()), rez.KnowledgeGraphPageParams{})
	s.ErrorIs(missingRootErr, rez.ErrNotFound)
	_, negativeLimitErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID), rez.KnowledgeGraphPageParams{Limit: -1})
	s.ErrorIs(negativeLimitErr, rez.ErrInvalidInput)
	_, malformedCursorErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID), rez.KnowledgeGraphPageParams{Cursor: "not-a-cursor"})
	s.ErrorIs(malformedCursorErr, rez.ErrInvalidInput)

	firstPage, firstPageErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID, fixture.b.ID), rez.KnowledgeGraphPageParams{Limit: 1})
	s.Require().NoError(firstPageErr)
	s.NotEmpty(firstPage.NextCursor)
	_, reusedCursorErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelRuntime, fixture.a.ID, fixture.b.ID), rez.KnowledgeGraphPageParams{
		Cursor: firstPage.NextCursor,
		Limit:  1,
	})
	s.ErrorIs(reusedCursorErr, rez.ErrInvalidInput)

	_, missingTenantErr := fixture.service.QueryGraphLevel(context.Background(), graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID), rez.KnowledgeGraphPageParams{})
	s.ErrorIs(missingTenantErr, rez.ErrTenantContextMissing)
}

func (s *KnowledgeGraphQueryServiceSuite) TestTenantIsolation() {
	fixture := s.newFixture()
	systemContext := s.SystemContext()
	client := fixture.database.Client(systemContext)
	s.Require().NoError(client.Tenant.Create().Exec(systemContext))
	foreignContext := execution.NewTenantContext(s.T().Context(), 2)
	foreignEntityBuilder := client.KnowledgeEntity.Create().SetCategory(kne.CategorySystem).SetKind("foreign")
	foreignEntity, foreignEntityErr := foreignEntityBuilder.Save(foreignContext)
	s.Require().NoError(foreignEntityErr)
	foreignTargetBuilder := client.KnowledgeEntity.Create().SetCategory(kne.CategorySystem).SetKind("foreign-target")
	foreignTarget, foreignTargetErr := foreignTargetBuilder.Save(foreignContext)
	s.Require().NoError(foreignTargetErr)
	foreignRelationshipBuilder := client.KnowledgeRelationship.Create().
		SetSourceEntityID(foreignEntity.ID).
		SetTargetEntityID(foreignTarget.ID).
		SetPredicate(knr.PredicateCalls)
	s.Require().NoError(foreignRelationshipBuilder.Exec(foreignContext))

	_, foreignRootErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, foreignEntity.ID), rez.KnowledgeGraphPageParams{})
	s.ErrorIs(foreignRootErr, rez.ErrNotFound)
	result, resultErr := fixture.service.QueryGraphLevel(fixture.ctx, graphQueryFor(rez.KnowledgeGraphDetailLevelSystems, fixture.a.ID, fixture.b.ID), rez.KnowledgeGraphPageParams{})
	s.Require().NoError(resultErr)
	s.NotContains(entityIDSet(result.Entities), foreignEntity.ID)
	s.NotContains(entityIDSet(result.Entities), foreignTarget.ID)
}

func (s *KnowledgeGraphQueryServiceSuite) TestTimestampIsIgnored() {
	fixture := s.newFixture()
	pastQuery := graphQueryFor(rez.KnowledgeGraphDetailLevelRuntime, fixture.a.ID, fixture.b.ID)
	pastQuery.Root.Timestamp = time.Unix(1, 0)
	futureQuery := pastQuery
	futureQuery.Root.Timestamp = time.Now().UTC().Add(100 * 365 * 24 * time.Hour)
	past, pastErr := fixture.service.QueryGraphLevel(fixture.ctx, pastQuery, rez.KnowledgeGraphPageParams{Limit: 100})
	s.Require().NoError(pastErr)
	future, futureErr := fixture.service.QueryGraphLevel(fixture.ctx, futureQuery, rez.KnowledgeGraphPageParams{Limit: 100})
	s.Require().NoError(futureErr)
	s.Equal(entityIDSet(past.Entities), entityIDSet(future.Entities))
	s.Equal(connectionCountSet(past.Connections), connectionCountSet(future.Connections))
}
