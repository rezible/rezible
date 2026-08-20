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
	ne "github.com/rezible/rezible/ent/normalizedevent"
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

func (s *SystemAnalysisServiceSuite) service(knowledge rez.KnowledgeGraphService) *SystemAnalysisService {
	svc, svcErr := NewSystemAnalysisService(s.Database(), knowledge)
	s.Require().NoError(svcErr)
	return svc
}

func (s *SystemAnalysisServiceSuite) createAnalysis() *ent.SystemAnalysis {
	ctx := s.SeedTenantContext()
	create := s.Client(ctx).SystemAnalysis.Create()
	analysis, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return analysis
}

func (s *SystemAnalysisServiceSuite) createNormalizedEvent(subjectRef string) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	now := time.Now().UTC()
	encodedAttributes, encodeErr := projections.EncodeAttributes(struct{}{})
	s.Require().NoError(encodeErr)
	create := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("system-analysis-tests").
		SetProviderEventRef(uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(projections.SubjectKindSystemComponent.String()).
		SetOccurredAt(now).
		SetReceivedAt(now).
		SetAttributes(encodedAttributes)
	event, createErr := create.Save(ctx)
	s.Require().NoError(createErr)
	return event
}

func (s *SystemAnalysisServiceSuite) createGraphFixture() systemAnalysisGraphFixture {
	ctx := s.SeedTenantContext()
	client := s.Client(ctx)
	now := time.Now().UTC()

	createSource := client.KnowledgeEntity.Create().
		SetKind(kne.KindContainer).
		SetSubkind("service")
	source, sourceErr := createSource.Save(ctx)
	s.Require().NoError(sourceErr)
	createTarget := client.KnowledgeEntity.Create().
		SetKind(kne.KindContainer).
		SetSubkind("database")
	target, targetErr := createTarget.Save(ctx)
	s.Require().NoError(targetErr)
	createRelationship := client.KnowledgeRelationship.Create().
		SetKind(knr.KindInteractsWith).
		SetSubkind("uses").
		SetSourceEntityID(source.ID).
		SetTargetEntityID(target.ID)
	relationship, relationshipErr := createRelationship.Save(ctx)
	s.Require().NoError(relationshipErr)

	createAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("test").
		SetProviderSource("system-analysis-tests").
		SetProviderSubjectRef("service:" + uuid.NewString()).
		SetEntityID(source.ID)
	alias, aliasErr := createAlias.Save(ctx)
	s.Require().NoError(aliasErr)
	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(s.createNormalizedEvent(alias.ProviderSubjectRef).ID).
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
	fixture := s.createGraphFixture()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})
	analysis := s.createAnalysis()

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
		ListParams: ent.ListParams{Count: true},
		Predicates: []predicate.SystemAnalysisEntity{saentity.AnalysisID(analysis.ID)},
	}
	nodes, listErr := svc.ListSystemAnalysisEntities(ctx, params)
	s.Require().NoError(listErr)
	s.Equal(1, nodes.Count)
	s.Require().Len(nodes.Data, 1)
	s.Equal(node.ID, nodes.Data[0].ID)

	s.Require().NoError(svc.DeleteSystemAnalysisEntity(ctx, node.ID))
	nodeQuery := s.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.ID(node.ID))
	s.Equal(0, nodeQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestAnalysisRelationshipDerivesEndpointEntities() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})
	analysis := s.createAnalysis()

	relationship, createErr := svc.SetSystemAnalysisRelationship(ctx, uuid.Nil, func(m *ent.SystemAnalysisRelationshipMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKnowledgeRelationshipID(fixture.Relationship.ID)
		m.SetDescriptionOverride("service uses database")
	})
	s.Require().NoError(createErr)
	s.Equal(fixture.Relationship.ID, relationship.KnowledgeRelationshipID)

	entityParams := rez.ListSystemAnalysisEntitiesParams{
		ListParams: ent.ListParams{Count: true, Limit: 1, Offset: 1},
		Predicates: []predicate.SystemAnalysisEntity{saentity.AnalysisID(analysis.ID)},
	}
	nodes, listNodesErr := svc.ListSystemAnalysisEntities(ctx, entityParams)
	s.Require().NoError(listNodesErr)
	s.Equal(2, nodes.Count)
	s.Require().Len(nodes.Data, 1)

	relsParams := rez.ListSystemAnalysisRelationshipsParams{
		ListParams: ent.ListParams{Count: true},
		Predicates: []predicate.SystemAnalysisRelationship{sarel.AnalysisID(analysis.ID)},
	}
	relationships, listRelationshipsErr := svc.ListSystemAnalysisRelationships(ctx, relsParams)
	s.Require().NoError(listRelationshipsErr)
	s.Equal(1, relationships.Count)
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

	sourceNode := s.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID), saentity.KnowledgeEntityID(fixture.Source.ID)).
		OnlyX(ctx)
	s.Require().NoError(svc.DeleteSystemAnalysisEntity(ctx, sourceNode.ID))
	relationshipQuery := s.Client(ctx).SystemAnalysisRelationship.Query().
		Where(sarel.ID(relationship.ID))
	s.Equal(0, relationshipQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestListEntriesOrdersAndLoadsSubjects() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})
	analysis := s.createAnalysis()

	first, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindContext)
		m.SetTitle("First")
		m.SetProperties(map[string]any{"source": "test"})
	})
	s.Require().NoError(createErr)
	second, createErr := svc.SetSystemAnalysisEntry(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysis.ID)
		m.SetKind(sae.KindObservation)
		m.SetTitle("Second")
	})
	s.Require().NoError(createErr)
	s.Empty(second.Properties)

	_, addErr := svc.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(second.ID)
		m.SetRole("primary")
		m.SetKnowledgeEntityID(fixture.Source.ID)
	})
	s.Require().NoError(addErr)

	params := rez.ListSystemAnalysisEntriesParams{
		ListParams: ent.ListParams{Count: true},
		Predicates: []predicate.SystemAnalysisEntry{sae.AnalysisID(analysis.ID)},
	}
	entries, listErr := svc.ListSystemAnalysisEntries(ctx, params)
	s.Require().NoError(listErr)
	s.Equal(2, entries.Count)
	s.Require().Len(entries.Data, 2)
	s.Equal(first.ID, entries.Data[0].ID)
	s.Equal(second.ID, entries.Data[1].ID)
	s.Require().Len(entries.Data[1].Edges.Subjects, 1)
	subject := entries.Data[1].Edges.Subjects[0]
	s.Equal("primary", subject.Role)
	s.Require().NotNil(subject.Edges.KnowledgeEntity)
	s.Equal(fixture.Source.ID, subject.Edges.KnowledgeEntity.ID)
	s.NotEmpty(subject.Edges.KnowledgeEntity.Edges.Aliases)

	params2 := rez.ListSystemAnalysisEntriesParams{
		ListParams: ent.ListParams{Count: true, Limit: 1, Offset: 1},
		Predicates: []predicate.SystemAnalysisEntry{sae.AnalysisID(analysis.ID)},
	}
	page, pageErr := svc.ListSystemAnalysisEntries(ctx, params2)
	s.Require().NoError(pageErr)
	s.Equal(2, page.Count)
	s.Require().Len(page.Data, 1)
	s.Equal(second.ID, page.Data[0].ID)
}

func (s *SystemAnalysisServiceSuite) TestEntryMutationsValidateUpdateAndDelete() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})
	analysis := s.createAnalysis()

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
	entryQuery := s.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.ID(entry.ID))
	subjectsQuery := s.Client(ctx).SystemAnalysisEntrySubject.Query().
		Where(saes.EntryID(entry.ID))
	s.Equal(0, entryQuery.CountX(ctx))
	s.Equal(0, subjectsQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestEntrySubjectRequiresExactlyOneGraphReference() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})
	analysis := s.createAnalysis()
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
	subjectQuery := s.Client(ctx).SystemAnalysisEntrySubject.Query().
		Where(saes.ID(relationshipSubject.ID))
	s.Equal(0, subjectQuery.CountX(ctx))
}

func (s *SystemAnalysisServiceSuite) TestGetGraphUsesSubjectBeforeScope() {
	ctx := s.SeedTenantContext()
	client := s.Client(ctx)
	createScope := client.KnowledgeEntity.Create().
		SetKind(kne.KindSystem).
		SetSubkind("system")
	scope, scopeErr := createScope.Save(ctx)
	s.Require().NoError(scopeErr)
	createSubject := client.KnowledgeEntity.Create().
		SetKind(kne.KindContainer).
		SetSubkind("service")
	subject, subjectErr := createSubject.Save(ctx)
	s.Require().NoError(subjectErr)
	createAnalysis := client.SystemAnalysis.Create().
		SetScopeEntityID(scope.ID).
		SetSubjectEntityID(subject.ID)
	analysis, createErr := createAnalysis.Save(ctx)
	s.Require().NoError(createErr)

	knowledge := mocks.NewMockKnowledgeGraphService(s.T())
	expectedParams := rez.GetKnowledgeGraphViewParams{
		EntityID:          subject.ID,
		Depth:             2,
		RelationshipKinds: []string{"interacts_with"},
	}
	knowledge.EXPECT().
		GetView(mock.Anything, expectedParams).
		Return(&rez.KnowledgeGraphView{RootID: subject.ID}, nil).
		Once()
	svc := s.service(knowledge)

	view, viewErr := svc.GetSystemAnalysisGraph(ctx, analysis.ID, rez.GetKnowledgeGraphViewParams{
		Depth:             2,
		RelationshipKinds: []string{"interacts_with"},
	})
	s.Require().NoError(viewErr)
	s.Equal(subject.ID, view.RootID)
}

func (s *SystemAnalysisServiceSuite) TestGetGraphDoesNotUseScope() {
	ctx := s.SeedTenantContext()
	createScope := s.Client(ctx).KnowledgeEntity.Create().
		SetKind(kne.KindSystem).
		SetSubkind("system")
	scope, scopeErr := createScope.Save(ctx)
	s.Require().NoError(scopeErr)
	createAnalysis := s.Client(ctx).SystemAnalysis.Create().
		SetScopeEntityID(scope.ID)
	analysis, createErr := createAnalysis.Save(ctx)
	s.Require().NoError(createErr)

	knowledge := mocks.NewMockKnowledgeGraphService(s.T())
	knowledge.EXPECT().
		GetView(mock.Anything, rez.GetKnowledgeGraphViewParams{}).
		Return(&rez.KnowledgeGraphView{}, nil).
		Once()
	svc := s.service(knowledge)

	view, viewErr := svc.GetSystemAnalysisGraph(ctx, analysis.ID, rez.GetKnowledgeGraphViewParams{})
	s.Require().NoError(viewErr)
	s.Equal(uuid.Nil, view.RootID)
}

func (s *SystemAnalysisServiceSuite) TestGetGraphFallsBackToKnowledgeGraphDefault() {
	ctx := s.SeedTenantContext()
	analysis := s.createAnalysis()

	knowledge := mocks.NewMockKnowledgeGraphService(s.T())
	knowledge.EXPECT().
		GetView(mock.Anything, rez.GetKnowledgeGraphViewParams{}).
		Return(&rez.KnowledgeGraphView{}, nil).
		Once()
	svc := s.service(knowledge)

	_, viewErr := svc.GetSystemAnalysisGraph(ctx, analysis.ID, rez.GetKnowledgeGraphViewParams{})
	s.Require().NoError(viewErr)
}

func (s *SystemAnalysisServiceSuite) TestHasSystemAnalysisEntity() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	analysis := s.Client(ctx).SystemAnalysis.Create().SetSubjectEntityID(fixture.Source.ID).SaveX(ctx)
	svc := s.service(&KnowledgeGraphService{db: s.Database()})

	included, includedErr := svc.HasSystemAnalysisEntity(ctx, analysis.ID, fixture.Source.ID)
	s.Require().NoError(includedErr)
	s.False(included)

	s.Client(ctx).SystemAnalysisEntity.Create().SetAnalysisID(analysis.ID).SetKnowledgeEntityID(fixture.Source.ID).SaveX(ctx)
	included, includedErr = svc.HasSystemAnalysisEntity(ctx, analysis.ID, fixture.Source.ID)
	s.Require().NoError(includedErr)
	s.True(included)
}

func (s *SystemAnalysisServiceSuite) TestIncludeSystemAnalysisSubjectsAddsRelationshipEndpoints() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	analysis := s.createAnalysis()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})

	includeParams := rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      analysis.ID,
		EntityIds:       []uuid.UUID{fixture.Source.ID},
		RelationshipIds: []uuid.UUID{fixture.Relationship.ID},
	}
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, includeParams))
	s.Require().NoError(svc.IncludeSystemAnalysisSubjects(ctx, includeParams))

	entities := s.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysis.ID)).
		AllX(ctx)
	s.Require().Len(entities, 2)
	for _, entity := range entities {
		s.Equal(analysis.ID, entity.AnalysisID)
	}
	relationships := s.Client(ctx).SystemAnalysisRelationship.Query().
		Where(sarel.AnalysisID(analysis.ID)).
		AllX(ctx)
	s.Require().Len(relationships, 1)
	s.Equal(analysis.ID, relationships[0].AnalysisID)
	s.Equal(fixture.Relationship.ID, relationships[0].KnowledgeRelationshipID)
}

func (s *SystemAnalysisServiceSuite) TestEntrySubjectDatabaseRequiresExactlyOneGraphReference() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	analysis := s.createAnalysis()
	entry := s.Client(ctx).SystemAnalysisEntry.Create().
		SetAnalysisID(analysis.ID).
		SetKind(sae.KindObservation).
		SetTitle("Observation").
		SaveX(ctx)

	_, missingErr := s.Client(ctx).SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("missing").
		Save(ctx)
	s.Require().Error(missingErr)
	_, multipleErr := s.Client(ctx).SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("multiple").
		SetKnowledgeEntityID(fixture.Source.ID).
		SetKnowledgeRelationshipID(fixture.Relationship.ID).
		Save(ctx)
	s.Require().Error(multipleErr)
}

func (s *SystemAnalysisServiceSuite) TestSetSystemAnalysisEntryCreatesSubjectsAtomicallyAndAppends() {
	ctx := s.SeedTenantContext()
	fixture := s.createGraphFixture()
	analysis := s.createAnalysis()
	svc := s.service(&KnowledgeGraphService{db: s.Database()})
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
	s.Equal(3, s.Client(ctx).SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID)).CountX(ctx))
	s.Equal(0, s.Client(ctx).SystemAnalysisEntity.Query().Where(saentity.AnalysisID(analysis.ID)).CountX(ctx), "findings do not auto-include graph subjects")

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
	s.Equal(2, s.Client(ctx).SystemAnalysisEntry.Query().Where(sae.AnalysisID(analysis.ID)).CountX(ctx))
	s.Equal(3, s.Client(ctx).SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID)).CountX(ctx))
}
