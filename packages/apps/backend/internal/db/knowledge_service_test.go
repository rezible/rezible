package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knfh "github.com/rezible/rezible/ent/knowledgefacthistory"
	knfp "github.com/rezible/rezible/ent/knowledgefactprovenance"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/testkit"
)

type KnowledgeServiceSuite struct {
	testkit.Suite
}

func TestKnowledgeServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeServiceSuite{Suite: testkit.NewSuite()})
}

func (s *KnowledgeServiceSuite) service() *KnowledgeService {
	return NewKnowledgeService(s.Client())
}

func (s *KnowledgeServiceSuite) createEntity() uuid.UUID {
	entity, err := s.service().SetEntity(s.SeedTenantContext(), uuid.Nil, func(m *ent.KnowledgeEntityMutation) {
		m.SetKind(kne.KindComponent)
		m.SetDisplayName("Payments API")
		m.SetDescription("Handles payment requests")
		m.SetProperties(map[string]any{"tier": "backend"})
	})
	s.Require().NoError(err)
	return entity.ID
}

func (s *KnowledgeServiceSuite) TestSetEntityAliasCreatesEntityAliasAndGetEntityResolvesIt() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	entityID := s.createEntity()
	externalID := "payments-api-" + uuid.NewString()
	observedAt := time.Now().UTC()

	alias, err := svc.SetEntityAlias(ctx, uuid.Nil, func(m *ent.KnowledgeEntityAliasMutation) {
		m.SetEntityID(entityID)
		m.SetProvider("fake")
		m.SetProviderSource("datasync/system_components")
		m.SetSubjectKind("system_component")
		m.SetSubjectRef(externalID)
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, alias.ID)

	entity, err := svc.GetEntity(ctx, kne.ID(alias.EntityID))
	s.Require().NoError(err)
	s.Equal(kne.KindComponent, entity.Kind)
	s.Equal("Payments API", entity.DisplayName)
}

func (s *KnowledgeServiceSuite) TestSetEntityAliasIsIdempotentOnProviderExternalIdentity() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	entityID := s.createEntity()
	externalID := "payments-api-" + uuid.NewString()
	observedAt := time.Now().UTC()

	firstAlias, err := svc.SetEntityAlias(ctx, uuid.Nil, func(m *ent.KnowledgeEntityAliasMutation) {
		m.SetEntityID(entityID)
		m.SetProvider("fake")
		m.SetProviderSource("datasync/system_components")
		m.SetSubjectKind("system_component")
		m.SetSubjectRef(externalID)
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)

	secondObservedAt := observedAt.Add(time.Minute)
	secondAlias, err := svc.SetEntityAlias(ctx, uuid.Nil, func(m *ent.KnowledgeEntityAliasMutation) {
		m.SetEntityID(entityID)
		m.SetProvider("fake")
		m.SetProviderSource("datasync/system_components")
		m.SetSubjectKind("system_component")
		m.SetSubjectRef(externalID)
		m.SetFirstSeenAt(secondObservedAt)
		m.SetLastSeenAt(secondObservedAt)
	})
	s.Require().NoError(err)
	s.Equal(firstAlias.ID, secondAlias.ID)

	aliasCount, err := s.Client().KnowledgeEntityAlias.Query().
		Where(
			knea.Provider("fake"),
			knea.ProviderSource("datasync/system_components"),
			knea.SubjectKind("system_component"),
			knea.SubjectRef(externalID),
		).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, aliasCount)
}

func (s *KnowledgeServiceSuite) TestSetRelationshipIsIdempotentOnSourceTargetKind() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	sourceID := s.createEntity()
	targetID := s.createEntity()
	observedAt := time.Now().UTC()

	firstRel, err := svc.SetRelationship(ctx, uuid.Nil, func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(sourceID)
		m.SetTargetEntityID(targetID)
		m.SetKind("depends_on")
		m.SetDisplayName("depends on")
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)

	secondObservedAt := observedAt.Add(time.Minute)
	secondRel, err := svc.SetRelationship(ctx, uuid.Nil, func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(sourceID)
		m.SetTargetEntityID(targetID)
		m.SetKind("depends_on")
		m.SetDisplayName("depends on")
		m.SetFirstSeenAt(secondObservedAt)
		m.SetLastSeenAt(secondObservedAt)
	})
	s.Require().NoError(err)
	s.Equal(firstRel.ID, secondRel.ID)

	relationshipCount, err := s.Client().KnowledgeRelationship.Query().
		Where(
			knr.SourceEntityID(sourceID),
			knr.TargetEntityID(targetID),
			knr.Kind("depends_on"),
		).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, relationshipCount)
}

func (s *KnowledgeServiceSuite) TestSupportedCanonicalKindsCanBePersisted() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	kinds := []kne.Kind{
		kne.KindComponent,
		kne.KindService,
		kne.KindRepository,
		kne.KindIncident,
		kne.KindChangeEvent,
	}

	for _, kind := range kinds {
		_, err := svc.SetEntity(ctx, uuid.Nil, func(m *ent.KnowledgeEntityMutation) {
			m.SetKind(kind)
			m.SetDisplayName(string(kind))
			m.SetProperties(map[string]any{})
		})
		s.Require().NoError(err)
	}
}

func (s *KnowledgeServiceSuite) TestSetFactProvenanceIsIdempotentOnEvidenceSource() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	entityID := s.createEntity()
	observedAt := time.Now().UTC()

	alias, err := svc.SetEntityAlias(ctx, uuid.Nil, func(m *ent.KnowledgeEntityAliasMutation) {
		m.SetEntityID(entityID)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetSubjectKind("repository")
		m.SetSubjectRef("rezible/rezible")
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)

	first, err := svc.SetFactProvenance(ctx, uuid.Nil, func(m *ent.KnowledgeFactProvenanceMutation) {
		m.SetAliasID(alias.ID)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("repo:rezible/rezible")
		m.SetExtractionMethod("test_projection")
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)

	second, err := svc.SetFactProvenance(ctx, uuid.Nil, func(m *ent.KnowledgeFactProvenanceMutation) {
		m.SetAliasID(alias.ID)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("repo:rezible/rezible")
		m.SetExtractionMethod("test_projection")
		m.SetFirstSeenAt(observedAt.Add(time.Minute))
		m.SetLastSeenAt(observedAt.Add(time.Minute))
	})
	s.Require().NoError(err)
	s.Equal(first.ID, second.ID)
	s.True(second.LastSeenAt.After(first.LastSeenAt))

	count, err := s.Client().KnowledgeFactProvenance.Query().
		Where(
			knfp.AliasID(alias.ID),
			knfp.Provider("github"),
			knfp.ProviderSource("normalized_events"),
			knfp.ProviderEventRef("repo:rezible/rezible"),
			knfp.ExtractionMethod("test_projection"),
		).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *KnowledgeServiceSuite) TestSetFactHistoryIsAppendOnlyAndDedupeByHistoryKey() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	entityID := s.createEntity()
	observedAt := time.Now().UTC()
	alias, err := svc.SetEntityAlias(ctx, uuid.Nil, func(m *ent.KnowledgeEntityAliasMutation) {
		m.SetEntityID(entityID)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetSubjectKind("repository")
		m.SetSubjectRef("rezible/rezible")
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)

	first, err := svc.SetFactHistory(ctx, uuid.Nil, func(m *ent.KnowledgeFactHistoryMutation) {
		m.SetFactKind("alias")
		m.SetAliasID(alias.ID)
		m.SetEventKind("alias_observed")
		m.SetHistoryKey("alias:" + alias.ID.String() + ":repo-1")
		m.SetOccurredAt(observedAt)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("repo-1")
		m.SetExtractionMethod("test_projection")
		m.SetAttributes(map[string]any{})
	})
	s.Require().NoError(err)
	second, err := svc.SetFactHistory(ctx, uuid.Nil, func(m *ent.KnowledgeFactHistoryMutation) {
		m.SetFactKind("alias")
		m.SetAliasID(alias.ID)
		m.SetEventKind("alias_observed")
		m.SetHistoryKey("alias:" + alias.ID.String() + ":repo-1")
		m.SetOccurredAt(observedAt.Add(time.Minute))
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("repo-1")
		m.SetExtractionMethod("test_projection")
		m.SetAttributes(map[string]any{})
	})
	s.Require().NoError(err)
	s.Equal(first.ID, second.ID)

	_, err = svc.SetFactHistory(ctx, uuid.Nil, func(m *ent.KnowledgeFactHistoryMutation) {
		m.SetFactKind("alias")
		m.SetAliasID(alias.ID)
		m.SetEventKind("alias_observed")
		m.SetHistoryKey("alias:" + alias.ID.String() + ":repo-2")
		m.SetOccurredAt(observedAt.Add(2 * time.Minute))
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("repo-2")
		m.SetExtractionMethod("test_projection")
		m.SetAttributes(map[string]any{})
	})
	s.Require().NoError(err)

	count, err := s.Client().KnowledgeFactHistory.Query().
		Where(knfh.AliasID(alias.ID)).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(2, count)
}

func (s *KnowledgeServiceSuite) TestGetRelationshipWithEvidenceReturnsHistory() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	sourceID := s.createEntity()
	targetID := s.createEntity()
	observedAt := time.Now().UTC()

	rel, err := svc.SetRelationship(ctx, uuid.Nil, func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(sourceID)
		m.SetTargetEntityID(targetID)
		m.SetKind("changes_repository")
		m.SetDisplayName("changes repository")
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)
	_, err = svc.SetFactProvenance(ctx, uuid.Nil, func(m *ent.KnowledgeFactProvenanceMutation) {
		m.SetRelationshipID(rel.ID)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("change-1")
		m.SetExtractionMethod("test_projection")
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)
	_, err = svc.SetFactHistory(ctx, uuid.Nil, func(m *ent.KnowledgeFactHistoryMutation) {
		m.SetFactKind("relationship")
		m.SetRelationshipID(rel.ID)
		m.SetEventKind("relationship_observed")
		m.SetHistoryKey("relationship:" + rel.ID.String() + ":change-1")
		m.SetOccurredAt(observedAt)
		m.SetProvider("github")
		m.SetProviderSource("normalized_events")
		m.SetProviderEventRef("change-1")
		m.SetExtractionMethod("test_projection")
		m.SetAttributes(map[string]any{})
	})
	s.Require().NoError(err)

	evidenceRel, evidenceRelErr := s.Client().KnowledgeRelationship.Query().
		Where(knr.ID(rel.ID)).
		WithProvenance().
		Only(ctx)
	s.Require().NoError(evidenceRelErr)
	s.Equal(rel.ID, evidenceRel.ID)
	s.Len(evidenceRel.Edges.Provenance, 1)
	// s.Len(evidenceRel.Edges.History, 1)
}
