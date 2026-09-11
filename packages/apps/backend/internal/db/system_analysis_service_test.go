package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	saentity "github.com/rezible/rezible/ent/systemanalysisentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sarel "github.com/rezible/rezible/ent/systemanalysisrelationship"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type SystemAnalysisServiceSuite struct {
	test.Suite
}

type systemAnalysisGraphFixture struct {
	Source       *ent.KnowledgeEntity
	Target       *ent.KnowledgeEntity
	Relationship *ent.KnowledgeRelationship
	Evidence     *ent.KnowledgeEvidence
}

func TestSystemAnalysisServiceSuite(t *testing.T) {
	suite.Run(t, &SystemAnalysisServiceSuite{Suite: test.NewSuite()})
}

func (s *SystemAnalysisServiceSuite) service(tdb rez.Database, knowledge rez.KnowledgeGraphService) *SystemAnalysisService {
	svc, svcErr := NewSystemAnalysisService(tdb, knowledge)
	s.Require().NoError(svcErr)
	return svc
}

func (s *SystemAnalysisServiceSuite) createAnalysis(tdb rez.Database) *ent.SystemAnalysis {
	ctx := s.SeedTenantContext()
	create := tdb.Client(ctx).SystemAnalysis.Create()
	analysis, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return analysis
}

func (s *SystemAnalysisServiceSuite) createNormalizedEvent(tdb rez.Database, subjectRef string) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	now := time.Now().UTC()
	encodedAttributes, encodeErr := projections.EncodeAttributes(struct{}{})
	s.Require().NoError(encodeErr)
	create := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("system-analysis-tests").
		SetProviderResourceRef(subjectRef).
		SetProviderEventSource("system-analysis-tests").
		SetProviderEventRef(uuid.NewString()).
		SetKind(projections.KindSystemComponent).
		SetOccurredAt(now).
		SetReceivedAt(now).
		SetAttributes(encodedAttributes)
	event, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return event
}

func (s *SystemAnalysisServiceSuite) createGraphFixture(tdb rez.Database) systemAnalysisGraphFixture {
	ctx := s.SeedTenantContext()
	client := tdb.Client(ctx)
	now := time.Now().UTC()

	createSource := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	source, sourceErr := createSource.Save(ctx)
	s.Require().NoError(sourceErr)
	createTarget := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("database")
	target, targetErr := createTarget.Save(ctx)
	s.Require().NoError(targetErr)
	createRelationship := client.KnowledgeRelationship.Create().
		SetPredicate(knr.PredicateUses).
		SetSourceEntityID(source.ID).
		SetTargetEntityID(target.ID)
	relationship, relationshipErr := createRelationship.Save(ctx)
	s.Require().NoError(relationshipErr)

	createAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderNamespace("system-analysis-tests").
		SetProviderResourceRef("service:" + uuid.NewString()).
		SetEntityID(source.ID)
	alias, aliasErr := createAlias.Save(ctx)
	s.Require().NoError(aliasErr)
	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(s.createNormalizedEvent(tdb, alias.ProviderResourceRef).ID).
		SetSubjectAliasID(alias.ID).
		SetKind(kev.KindObserved).
		SetAssertion("component_exists").
		SetEffectiveAt(now).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{DisplayName: "Checkout API"})
	evidence, evidenceErr := createEvidence.Save(ctx)
	s.Require().NoError(evidenceErr)

	return systemAnalysisGraphFixture{
		Source:       source,
		Target:       target,
		Relationship: relationship,
		Evidence:     evidence,
	}
}

func (s *SystemAnalysisServiceSuite) TestAnalysisEntityMutationsAndDelete() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	fixture := s.createGraphFixture(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})
	analysis := s.createAnalysis(tdb)

	node, createErr := svc.SetSystemAnalysisEntity(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntityMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKnowledgeEntityID(fixture.Source.ID)
		m.SetPosX(10)
		m.SetPosY(20)
		m.SetDescriptionOverride("checkout service")
	})
	s.Require().NoError(createErr)
	s.Equal(fixture.Source.ID, node.KnowledgeEntityID)
	s.Require().NotNil(node.PosX)
	s.Require().NotNil(node.PosY)
	s.Equal(10.0, *node.PosX)
	s.Equal(20.0, *node.PosY)

	_, retargetErr := svc.SetSystemAnalysisEntity(ctx, node.ID, func(m *ent.SystemAnalysisEntityMutation) {
		m.SetKnowledgeEntityID(fixture.Target.ID)
	})
	s.ErrorIs(retargetErr, rez.ErrInvalidInput)

	updated, updateErr := svc.SetSystemAnalysisEntity(ctx, node.ID, func(m *ent.SystemAnalysisEntityMutation) {
		m.SetPosX(30)
		m.SetPosY(40)
	})
	s.Require().NoError(updateErr)
	s.Equal(30.0, *updated.PosX)
	s.Equal(40.0, *updated.PosY)

	params := rez.ListSystemAnalysisEntitiesParams{
		ListParams: ent.ListParams{},
		Predicates: []predicate.SystemAnalysisEntity{saentity.AnalysisID(analysis.ID)},
	}
	nodes, listErr := svc.ListSystemAnalysisEntities(ctx, params)
	s.Require().NoError(listErr)
	s.Equal(1, nodes.Total)
	s.Require().Len(nodes.Data, 1)
	s.Equal(node.ID, nodes.Data[0].ID)

	s.Require().NoError(svc.DeleteSystemAnalysisEntity(ctx, node.ID))
	nodeQuery := tdb.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.ID(node.ID))
	s.Equal(0, nodeQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestAnalysisRelationshipDerivesEndpointEntities() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	fixture := s.createGraphFixture(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})
	analysis := s.createAnalysis(tdb)

	relationship, createErr := svc.SetSystemAnalysisRelationship(ctx, uuid.Nil, func(m *ent.SystemAnalysisRelationshipMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
		m.SetDescriptionOverride("service uses database")
	})
	s.Require().NoError(createErr)
	s.Equal(fixture.Relationship.ID, relationship.KnowledgeRelationshipID)

	entityParams := rez.ListSystemAnalysisEntitiesParams{
		ListParams: ent.ListParams{Page: 2, PageSize: 1},
		Predicates: []predicate.SystemAnalysisEntity{saentity.AnalysisID(analysis.ID)},
	}
	nodes, listNodesErr := svc.ListSystemAnalysisEntities(ctx, entityParams)
	s.Require().NoError(listNodesErr)
	s.Equal(2, nodes.Total)
	s.Require().Len(nodes.Data, 1)

	relsParams := rez.ListSystemAnalysisRelationshipsParams{
		ListParams: ent.ListParams{},
		Predicates: []predicate.SystemAnalysisRelationship{sarel.AnalysisID(analysis.ID)},
	}
	relationships, listRelationshipsErr := svc.ListSystemAnalysisRelationships(ctx, relsParams)
	s.Require().NoError(listRelationshipsErr)
	s.Equal(1, relationships.Total)
	s.Require().Len(relationships.Data, 1)
	s.Require().NotNil(relationships.Data[0].Edges.KnowledgeRelationship)

	_, retargetErr := svc.SetSystemAnalysisRelationship(ctx, relationship.ID, func(m *ent.SystemAnalysisRelationshipMutation) {
		m.SetKnowledgeRelationshipID(uuid.New())
	})
	s.ErrorIs(retargetErr, rez.ErrInvalidInput)

	layout := map[string]any{"curve": "smooth"}
	updated, updateErr := svc.SetSystemAnalysisRelationship(ctx, relationship.ID, func(m *ent.SystemAnalysisRelationshipMutation) {
		m.SetLayout(layout)
	})
	s.Require().NoError(updateErr)
	s.Equal(layout, updated.Layout)

	client := tdb.Client(ctx)
	sourceNode := client.SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID), saentity.KnowledgeEntityID(fixture.Source.ID)).
		OnlyX(ctx)
	s.Require().NoError(svc.DeleteSystemAnalysisEntity(ctx, sourceNode.ID))
	relationshipQuery := client.SystemAnalysisRelationship.Query().
		Where(sarel.ID(relationship.ID))
	s.Equal(0, relationshipQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestListEntriesOrdersAndLoadsSubjects() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	fixture := s.createGraphFixture(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})
	analysis := s.createAnalysis(tdb)

	entryOccAt := time.Now()
	first, firstErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindContext)
		m.SetTitle("First")
		m.SetProperties(map[string]any{"source": "test"})
		m.SetOccurredAt(entryOccAt)
	})
	s.Require().NoError(firstErr)

	second, secondErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindObservation)
		m.SetTitle("Second")
		m.SetOccurredAt(entryOccAt.Add(time.Second))
	})
	s.Require().NoError(secondErr)
	s.Empty(second.Properties)

	_, subjectErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(second.ID)
		m.SetRole("primary")
		m.SetKnowledgeEntityID(fixture.Source.ID)
	})
	s.Require().NoError(subjectErr)

	params := rez.ListSystemAnalysisEntriesParams{
		Predicates: []predicate.SystemAnalysisEntry{sae.AnalysisID(analysis.ID)},
	}
	entries, listErr := svc.ListSystemAnalysisEntries(ctx, params)
	s.Require().NoError(listErr)
	s.Require().Len(entries.Data, 2)
	s.Equal(first.ID, entries.Data[0].ID)

	entrySecond := entries.Data[1]
	s.Equal(second.ID, entrySecond.ID)

	secondSubjects, secondSubjectsErr := entrySecond.Edges.SubjectsOrErr()
	s.Require().NoError(secondSubjectsErr)
	s.Require().Len(secondSubjects, 1)
	subject := secondSubjects[0]
	s.Equal("primary", subject.Role)

	params2 := rez.ListSystemAnalysisEntriesParams{
		Predicates: []predicate.SystemAnalysisEntry{sae.AnalysisID(analysis.ID)},
		ListParams: ent.ListParams{Page: 2, PageSize: 1},
	}
	page, pageErr := svc.ListSystemAnalysisEntries(ctx, params2)
	s.Require().NoError(pageErr)
	s.Equal(2, page.Total)
	s.Require().Len(page.Data, 1)
	s.Equal(second.ID, page.Data[0].ID)
}

func (s *SystemAnalysisServiceSuite) TestEntryMutationsValidateUpdateAndDelete() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	fixture := s.createGraphFixture(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})
	analysis := s.createAnalysis(tdb)

	_, invalidErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.Kind("timeline_event"))
		m.SetTitle("Invalid")
	})
	s.Require().Error(invalidErr)

	entry, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindObservation)
		m.SetTitle("Original")
	})
	s.Require().NoError(createErr)

	_, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("primary")
		m.SetKnowledgeEntityID(fixture.Source.ID)
	})
	s.Require().NoError(addErr)

	title := "Updated"
	properties := map[string]any{"confidence": "high"}
	updated, updateErr := svc.SetSystemAnalysisEntry(ctx, entry.ID, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetTitle(title)
		m.SetProperties(properties)
	})
	s.Require().NoError(updateErr)
	s.Equal(title, updated.Title)
	s.Equal(properties, updated.Properties)

	s.Require().NoError(svc.DeleteSystemAnalysisEntry(ctx, entry.ID))
	client := tdb.Client(ctx)
	entryQuery := client.SystemAnalysisEntry.Query().
		Where(sae.ID(entry.ID))
	subjectsQuery := client.SystemAnalysisEntrySubject.Query().
		Where(saes.EntryID(entry.ID))
	s.Equal(0, entryQuery.CountX(ctx))
	s.Equal(0, subjectsQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestEntrySubjectRequiresExactlyOneGraphReference() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	fixture := s.createGraphFixture(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})
	analysis := s.createAnalysis(tdb)

	entry, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindObservation)
		m.SetTitle("Observation")
	})
	s.Require().NoError(createErr)

	_, missingErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("missing")
	})
	s.ErrorIs(missingErr, rez.ErrInvalidInput)

	_, multipleErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("multiple")
		m.SetKnowledgeEntityID(fixture.Source.ID)
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
	})
	s.ErrorIs(multipleErr, rez.ErrInvalidInput)

	entitySubject, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("entity")
		m.SetKnowledgeEntityID(fixture.Source.ID)
	})
	s.Require().NoError(addErr)
	s.NotEqual(uuid.Nil, entitySubject.ID)
	s.Require().NotNil(entitySubject.KnowledgeEntityID)

	relationshipSubject, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("relationship")
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
	})
	s.Require().NoError(addErr)
	s.Require().NotNil(relationshipSubject.KnowledgeRelationshipID)

	evidenceSubject, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("evidence")
		m.SetKnowledgeEvidenceID(fixture.Evidence.ID)
	})
	s.Require().NoError(addErr)
	s.Require().NotNil(evidenceSubject.KnowledgeEvidenceID)

	createNormalizedEvent := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").SetProviderNamespace("analysis").SetProviderResourceRef(uuid.NewString()).
		SetKind("log").SetProviderEventSource("analysis").SetProviderEventRef(uuid.NewString()).
		SetAttributes([]byte(`{"message":"connection timeout"}`)).SetOccurredAt(time.Now()).SetReceivedAt(time.Now())
	normalizedEvent, normalizedEventErr := createNormalizedEvent.Save(ctx)
	s.Require().NoError(normalizedEventErr)
	observationSubject, observationSubjectErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("observation")
		m.SetNormalizedEventID(normalizedEvent.ID)
	})
	s.Require().NoError(observationSubjectErr)
	s.Equal(normalizedEvent.ID, *observationSubject.NormalizedEventID)
	_, duplicateObservationErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("observation")
		m.SetNormalizedEventID(normalizedEvent.ID)
	})
	s.Error(duplicateObservationErr)

	_, multipleObservationErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("invalid")
		m.SetKnowledgeEntityID(fixture.Source.ID)
		m.SetNormalizedEventID(normalizedEvent.ID)
	})
	s.ErrorIs(multipleObservationErr, rez.ErrInvalidInput)

	_, retargetErr := svc.SetSystemAnalysisEntrySubject(ctx, entitySubject.ID, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
	})
	s.ErrorIs(retargetErr, rez.ErrInvalidInput)

	updated, updateErr := svc.SetSystemAnalysisEntrySubject(ctx, entitySubject.ID, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetRole("renamed")
	})
	s.Require().NoError(updateErr)
	s.Equal("renamed", updated.Role)

	s.Require().NoError(svc.DeleteSystemAnalysisEntrySubject(ctx, relationshipSubject.ID))
	subjectQuery := tdb.Client(ctx).SystemAnalysisEntrySubject.Query().
		Where(saes.ID(relationshipSubject.ID))
	s.Equal(0, subjectQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestGetGraphUsesSubjectBeforeScope() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	createScope := client.KnowledgeEntity.Create().
		SetCategory(kne.CategorySystem).
		SetKind("system")
	scope, scopeErr := createScope.Save(ctx)
	s.Require().NoError(scopeErr)
	createSubject := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	subject, subjectErr := createSubject.Save(ctx)
	s.Require().NoError(subjectErr)
	createAnalysis := client.SystemAnalysis.Create().
		SetScopeEntityID(scope.ID).
		SetSubjectEntityID(subject.ID)
	analysis, createErr := createAnalysis.Save(ctx)
	s.Require().NoError(createErr)

	knowledge := mocks.NewMockKnowledgeGraphService(s.T())
	expectedParams := rez.GetKnowledgeGraphViewParams{
		EntityID:               subject.ID,
		Depth:                  2,
		RelationshipPredicates: []string{"uses"},
	}
	knowledge.EXPECT().
		GetView(mock.Anything, expectedParams).
		Return(&rez.KnowledgeGraphView{RootID: subject.ID}, nil).
		Once()
	svc := s.service(tdb, knowledge)

	view, viewErr := svc.GetSystemAnalysisGraph(ctx, analysis.ID, rez.GetKnowledgeGraphViewParams{
		Depth:                  2,
		RelationshipPredicates: []string{"uses"},
	})
	s.Require().NoError(viewErr)
	s.Equal(subject.ID, view.RootID)
}

func (s *SystemAnalysisServiceSuite) TestGetGraphDoesNotUseScope() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	createScope := client.KnowledgeEntity.Create().
		SetCategory(kne.CategorySystem).
		SetKind("system")
	scope, scopeErr := createScope.Save(ctx)
	s.Require().NoError(scopeErr)
	createAnalysis := client.SystemAnalysis.Create().
		SetScopeEntityID(scope.ID)
	analysis, createErr := createAnalysis.Save(ctx)
	s.Require().NoError(createErr)

	knowledge := mocks.NewMockKnowledgeGraphService(s.T())
	knowledge.EXPECT().
		GetView(mock.Anything, rez.GetKnowledgeGraphViewParams{}).
		Return(&rez.KnowledgeGraphView{}, nil).
		Once()
	svc := s.service(tdb, knowledge)

	view, viewErr := svc.GetSystemAnalysisGraph(ctx, analysis.ID, rez.GetKnowledgeGraphViewParams{})
	s.Require().NoError(viewErr)
	s.Equal(uuid.Nil, view.RootID)
}

func (s *SystemAnalysisServiceSuite) TestGetGraphFallsBackToKnowledgeGraphDefault() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	analysis := s.createAnalysis(tdb)

	knowledge := mocks.NewMockKnowledgeGraphService(s.T())
	knowledge.EXPECT().
		GetView(mock.Anything, rez.GetKnowledgeGraphViewParams{}).
		Return(&rez.KnowledgeGraphView{}, nil).
		Once()
	svc := s.service(tdb, knowledge)

	_, viewErr := svc.GetSystemAnalysisGraph(ctx, analysis.ID, rez.GetKnowledgeGraphViewParams{})
	s.Require().NoError(viewErr)
}

func (s *SystemAnalysisServiceSuite) TestHasSystemAnalysisEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(tdb)
	analysis := client.SystemAnalysis.Create().SetSubjectEntityID(fixture.Source.ID).SaveX(ctx)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})

	included, includedErr := svc.HasSystemAnalysisEntity(ctx, analysis.ID, fixture.Source.ID)
	s.Require().NoError(includedErr)
	s.False(included)

	client.SystemAnalysisEntity.Create().SetAnalysisID(analysis.ID).SetKnowledgeEntityID(fixture.Source.ID).SaveX(ctx)
	included, includedErr = svc.HasSystemAnalysisEntity(ctx, analysis.ID, fixture.Source.ID)
	s.Require().NoError(includedErr)
	s.True(included)
}

func (s *SystemAnalysisServiceSuite) TestIncludeSystemAnalysisSubjectsAddsRelationshipEndpoints() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(tdb)
	analysis := s.createAnalysis(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})

	includeParams := rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      analysis.ID,
		EntityIds:       []uuid.UUID{fixture.Source.ID},
		RelationshipIds: []uuid.UUID{fixture.Relationship.ID},
	}
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, includeParams))
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, includeParams))

	entities := client.SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID)).
		AllX(ctx)
	s.Require().Len(entities, 2)
	for _, entity := range entities {
		s.Equal(analysis.ID, entity.AnalysisID)
	}
	relationships := client.SystemAnalysisRelationship.Query().
		Where(sarel.AnalysisID(analysis.ID)).
		AllX(ctx)
	s.Require().Len(relationships, 1)
	s.Equal(analysis.ID, relationships[0].AnalysisID)
	s.Equal(fixture.Relationship.ID, relationships[0].KnowledgeRelationshipID)
}

func (s *SystemAnalysisServiceSuite) TestEntrySubjectDatabaseRequiresExactlyOneGraphReference() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(tdb)
	analysis := s.createAnalysis(tdb)
	entry := client.SystemAnalysisEntry.Create().
		SetAnalysisID(analysis.ID).
		SetKind(sae.KindObservation).
		SetTitle("Observation").
		SaveX(ctx)

	_, missingErr := client.SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("missing").
		Save(ctx)
	s.Require().Error(missingErr)
	_, multipleErr := client.SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("multiple").
		SetKnowledgeEntityID(fixture.Source.ID).
		SetKnowledgeRelationshipID(fixture.Relationship.ID).
		Save(ctx)
	s.Require().Error(multipleErr)
}

func (s *SystemAnalysisServiceSuite) TestEntrySubjectDatabaseRequiresExactlyOneIncludingNormalizedEvent() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(tdb)
	analysis := s.createAnalysis(tdb)
	createEntry := client.SystemAnalysisEntry.Create().SetAnalysisID(analysis.ID).SetKind(sae.KindObservation).SetTitle("Observation")
	entry := createEntry.SaveX(ctx)
	createEvidence := client.NormalizedEvent.Create().SetProvider("test").SetProviderNamespace("analysis").SetProviderResourceRef(uuid.NewString()).SetKind("log").SetProviderEventSource("analysis").SetProviderEventRef(uuid.NewString()).SetAttributes([]byte(`{"message":"timeout"}`)).SetOccurredAt(time.Now()).SetReceivedAt(time.Now())
	evidence, evidenceErr := createEvidence.Save(ctx)
	s.Require().NoError(evidenceErr)

	missingCreate := client.SystemAnalysisEntrySubject.Create().SetEntryID(entry.ID).SetRole("missing")
	_, missingErr := missingCreate.Save(ctx)
	s.Require().Error(missingErr)
	multipleCreate := client.SystemAnalysisEntrySubject.Create().SetEntryID(entry.ID).SetRole("multiple").SetKnowledgeEvidenceID(fixture.Evidence.ID).SetNormalizedEventID(evidence.ID)
	_, multipleErr := multipleCreate.Save(ctx)
	s.Require().Error(multipleErr)
	validCreate := client.SystemAnalysisEntrySubject.Create().SetEntryID(entry.ID).SetRole("observation").SetNormalizedEventID(evidence.ID)
	valid, validErr := validCreate.Save(ctx)
	s.Require().NoError(validErr)
	s.Equal(evidence.ID, *valid.NormalizedEventID)
}

func (s *SystemAnalysisServiceSuite) TestSetSystemAnalysisEntryCreatesSubjectsAtomicallyAndAppends() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(tdb)
	analysis := s.createAnalysis(tdb)
	svc := s.service(tdb, &KnowledgeGraphService{db: tdb})
	now := time.Now()
	setEntry := func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindFinding)
		m.SetOccurredAt(now)
		m.SetTitle("Database dependency")
		m.SetBody("API relies on DB.")
	}
	setSubjects := []func(*ent.SystemAnalysisEntrySubjectMutation){
		func(m *ent.SystemAnalysisEntrySubjectMutation) {
			m.SetRole("primary")
			m.SetKnowledgeEntityID(fixture.Source.ID)
		},
		func(m *ent.SystemAnalysisEntrySubjectMutation) {
			m.SetRole("contributing")
			m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
		},
		func(m *ent.SystemAnalysisEntrySubjectMutation) {
			m.SetRole("evidence_for")
			m.SetKnowledgeEvidenceID(fixture.Evidence.ID)
		},
	}
	entry, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, setEntry, setSubjects...)
	s.Require().NoError(createErr)
	s.Equal(sae.KindFinding, entry.Kind)
	s.Equal(1, entry.Sequence)
	s.Equal("Database dependency", entry.Title)
	s.Equal("API relies on DB.", entry.Body)
	s.Equal(3, client.SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID)).CountX(ctx))
	s.Equal(0, client.SystemAnalysisEntity.Query().Where(saentity.AnalysisID(analysis.ID)).CountX(ctx), "findings do not auto-include graph subjects")

	second, secondErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, setEntry)
	s.Require().NoError(secondErr)
	s.Equal(2, second.Sequence)

	missingID := uuid.New()
	_, invalidErr := svc.SetSystemAnalysisEntry(
		ctx,
		uuid.Nil,
		setEntry,
		func(m *ent.SystemAnalysisEntrySubjectMutation) {
			m.SetRole("primary")
			m.SetKnowledgeEntityID(fixture.Source.ID)
		},
		func(m *ent.SystemAnalysisEntrySubjectMutation) {
			m.SetRole("missing")
			m.SetKnowledgeEvidenceID(missingID)
		},
	)
	s.Require().Error(invalidErr)
	s.Equal(2, client.SystemAnalysisEntry.Query().Where(sae.AnalysisID(analysis.ID)).CountX(ctx))
	s.Equal(3, client.SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID)).CountX(ctx))
}
