package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
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

func (s *SystemAnalysisServiceSuite) service(tdb rez.Database, knowledge rez.KnowledgeGraphQueryService) *SystemAnalysisService {
	svc, svcErr := NewSystemAnalysisService(tdb, knowledge)
	s.Require().NoError(svcErr)
	return svc
}

func (s *SystemAnalysisServiceSuite) createAnalysis(ctx context.Context, tdb rez.Database) *ent.SystemAnalysis {
	create := tdb.Client(ctx).SystemAnalysis.Create()
	analysis, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return analysis
}

func (s *SystemAnalysisServiceSuite) createNormalizedEvent(ctx context.Context, tdb rez.Database, subjectRef string) *ent.NormalizedEvent {
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

func (s *SystemAnalysisServiceSuite) createGraphFixture(ctx context.Context, tdb rez.Database) systemAnalysisGraphFixture {
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
		SetEventID(s.createNormalizedEvent(ctx, tdb, alias.ProviderResourceRef).ID).
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
	ctx, tdb := s.SetupTestDatabase()
	fixture := s.createGraphFixture(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})
	analysis := s.createAnalysis(ctx, tdb)

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
	ctx, tdb := s.SetupTestDatabase()
	fixture := s.createGraphFixture(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})
	analysis := s.createAnalysis(ctx, tdb)

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
	ctx, tdb := s.SetupTestDatabase()
	fixture := s.createGraphFixture(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})
	analysis := s.createAnalysis(ctx, tdb)
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId: analysis.ID,
		EntityIds:  []uuid.UUID{fixture.Source.ID},
	}))

	entryOccAt := time.Now()
	firstParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.KindContext,
		Title:      "First",
		Body:       "Test context",
		OccurredAt: &entryOccAt,
	}
	first, firstErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, firstParams)
	s.Require().NoError(firstErr)

	secondOccAt := entryOccAt.Add(time.Second)
	secondParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.KindObservation,
		Title:      "Second",
		OccurredAt: &secondOccAt,
	}
	second, secondErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, secondParams)
	s.Require().NoError(secondErr)
	s.Empty(second.Body)

	_, subjectErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(second.ID)
		m.SetRole("affected")
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
	s.Equal("affected", subject.Role)

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
	ctx, tdb := s.SetupTestDatabase()
	fixture := s.createGraphFixture(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})
	analysis := s.createAnalysis(ctx, tdb)
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId: analysis.ID,
		EntityIds:  []uuid.UUID{fixture.Source.ID},
	}))

	invalidParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.Kind("timeline_event"),
		Title:      "Invalid",
	}
	_, invalidErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, invalidParams)
	s.Require().Error(invalidErr)

	createParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.KindObservation,
		Title:      "Original",
	}
	entry, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, createParams)
	s.Require().NoError(createErr)

	_, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("affected")
		m.SetKnowledgeEntityID(fixture.Source.ID)
	})
	s.Require().NoError(addErr)

	updateParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.KindObservation,
		Title:      "Updated",
	}
	_, updateErr := svc.SetSystemAnalysisEntry(ctx, entry.ID, updateParams)
	s.Require().NoError(updateErr)

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
	ctx, tdb := s.SetupTestDatabase()
	fixture := s.createGraphFixture(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})
	analysis := s.createAnalysis(ctx, tdb)
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      analysis.ID,
		RelationshipIds: []uuid.UUID{fixture.Relationship.ID},
	}))

	createParams := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.KindObservation,
		Title:      "Observation",
	}
	entry, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, createParams)
	s.Require().NoError(createErr)

	_, missingErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("missing")
	})
	s.ErrorIs(missingErr, rez.ErrUnprocessableInput)

	_, multipleErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("multiple")
		m.SetKnowledgeEntityID(fixture.Source.ID)
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
	})
	s.ErrorIs(multipleErr, rez.ErrUnprocessableInput)

	entitySubject, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("context")
		m.SetKnowledgeEntityID(fixture.Source.ID)
	})
	s.Require().NoError(addErr)
	s.NotEqual(uuid.Nil, entitySubject.ID)
	s.Require().NotNil(entitySubject.KnowledgeEntityID)

	relationshipSubject, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("contributing")
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
	})
	s.Require().NoError(addErr)
	s.Require().NotNil(relationshipSubject.KnowledgeRelationshipID)

	evidenceSubject, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(entry.ID)
		m.SetRole("supports")
		m.SetKnowledgeEvidenceID(fixture.Evidence.ID)
	})
	s.Require().NoError(addErr)
	s.Require().NotNil(evidenceSubject.KnowledgeEvidenceID)

	_, retargetErr := svc.SetSystemAnalysisEntrySubject(ctx, entitySubject.ID, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
	})
	s.ErrorIs(retargetErr, rez.ErrInvalidInput)

}

func (s *SystemAnalysisServiceSuite) TestIncludeSystemAnalysisSubjectsAddsRelationshipEndpoints() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(ctx, tdb)
	analysis := s.createAnalysis(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})

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
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(ctx, tdb)
	analysis := s.createAnalysis(ctx, tdb)
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

func (s *SystemAnalysisServiceSuite) TestSetSystemAnalysisEntryCreatesSubjectsAtomicallyAndAppends() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	fixture := s.createGraphFixture(ctx, tdb)
	analysis := s.createAnalysis(ctx, tdb)
	svc := s.service(tdb, &KnowledgeGraphQueryService{db: tdb})
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      analysis.ID,
		RelationshipIds: []uuid.UUID{fixture.Relationship.ID},
		EntityIds:       []uuid.UUID{fixture.Source.ID},
	}))
	now := time.Now()
	params := rez.SetSystemAnalysisEntryParams{
		AnalysisID: analysis.ID,
		Kind:       sae.KindFinding,
		OccurredAt: &now,
		Title:      "Database dependency",
		Body:       "API relies on DB.",
		SetSubjects: []rez.SetSystemAnalysisEntrySubjectParams{
			{Role: "affected", KnowledgeEntityID: &fixture.Source.ID},
			{Role: "contributing", KnowledgeRelationshipID: &fixture.Relationship.ID},
			{Role: "supports", KnowledgeEvidenceID: &fixture.Evidence.ID},
		},
	}
	entry, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, params)
	s.Require().NoError(createErr)
	s.Equal(sae.KindFinding, entry.Kind)
	s.Equal(1, entry.Sequence)
	s.Equal("Database dependency", entry.Title)
	s.Equal("API relies on DB.", entry.Body)
	s.Equal(3, client.SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID)).CountX(ctx))
	s.Equal(2, client.SystemAnalysisEntity.Query().Where(saentity.AnalysisID(analysis.ID)).CountX(ctx),
		"entry attachments do not change prepared graph membership")

	params.SetSubjects = nil
	second, secondErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, params)
	s.Require().NoError(secondErr)
	s.Equal(2, second.Sequence)
}
