package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/knowledgeentityalias"
	"github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/testkit"
	"github.com/stretchr/testify/suite"
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
		m.SetKind(knowledgeentity.KindComponent)
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
		m.SetSource("datasync/system_components")
		m.SetExternalKind("system_component")
		m.SetExternalID(externalID)
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, alias.ID)

	entity, err := svc.GetEntity(ctx, knowledgeentity.ID(alias.EntityID))
	s.Require().NoError(err)
	s.Equal(knowledgeentity.KindComponent, entity.Kind)
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
		m.SetSource("datasync/system_components")
		m.SetExternalKind("system_component")
		m.SetExternalID(externalID)
		m.SetFirstSeenAt(observedAt)
		m.SetLastSeenAt(observedAt)
	})
	s.Require().NoError(err)

	secondObservedAt := observedAt.Add(time.Minute)
	secondAlias, err := svc.SetEntityAlias(ctx, uuid.Nil, func(m *ent.KnowledgeEntityAliasMutation) {
		m.SetEntityID(entityID)
		m.SetProvider("fake")
		m.SetSource("datasync/system_components")
		m.SetExternalKind("system_component")
		m.SetExternalID(externalID)
		m.SetFirstSeenAt(secondObservedAt)
		m.SetLastSeenAt(secondObservedAt)
	})
	s.Require().NoError(err)
	s.Equal(firstAlias.ID, secondAlias.ID)

	aliasCount, err := s.Client().KnowledgeEntityAlias.Query().
		Where(
			knowledgeentityalias.Provider("fake"),
			knowledgeentityalias.Source("datasync/system_components"),
			knowledgeentityalias.ExternalKind("system_component"),
			knowledgeentityalias.ExternalID(externalID),
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
			knowledgerelationship.SourceEntityID(sourceID),
			knowledgerelationship.TargetEntityID(targetID),
			knowledgerelationship.Kind("depends_on"),
		).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, relationshipCount)
}

func (s *KnowledgeServiceSuite) TestSupportedCanonicalKindsCanBePersisted() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	kinds := []knowledgeentity.Kind{
		knowledgeentity.KindComponent,
		knowledgeentity.KindService,
		knowledgeentity.KindRepository,
		knowledgeentity.KindIncident,
		knowledgeentity.KindChangeEvent,
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
