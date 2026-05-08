package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/testkit"
)

type SystemTopologyServiceSuite struct {
	testkit.Suite
}

func TestSystemTopologyServiceSuite(t *testing.T) {
	suite.Run(t, &SystemTopologyServiceSuite{Suite: testkit.NewSuite()})
}

func (s *SystemTopologyServiceSuite) service() *SystemTopologyService {
	svc, err := NewSystemTopologyService(s.Client())
	s.Require().NoError(err)
	return svc
}

func (s *SystemTopologyServiceSuite) createSystemTopologyEntity(name string) *ent.KnowledgeEntity {
	entity, err := s.Client().KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName(name).
		SetProperties(map[string]any{}).
		Save(s.SeedTenantContext())
	s.Require().NoError(err)
	return entity
}

func (s *SystemTopologyServiceSuite) createRelationship(source, target *ent.KnowledgeEntity) *ent.KnowledgeRelationship {
	rel, err := s.Client().KnowledgeRelationship.Create().
		SetSourceEntityID(source.ID).
		SetTargetEntityID(target.ID).
		SetKind("depends_on").
		SetDisplayName("depends on").
		SetProperties(map[string]any{"criticality": "high"}).
		Save(s.SeedTenantContext())
	s.Require().NoError(err)
	return rel
}

func (s *SystemTopologyServiceSuite) TestCreateSnapshotDeduplicatesRelationshipsAcrossRootTraversals() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	source := s.createSystemTopologyEntity("Payments API " + uuid.NewString())
	target := s.createSystemTopologyEntity("Ledger DB " + uuid.NewString())
	rel := s.createRelationship(source, target)

	snapshot, err := svc.CreateSnapshot(ctx, rez.CreateSystemTopologySnapshotParams{
		RootEntityIDs: []uuid.UUID{source.ID, target.ID},
		Depth:         1,
	})
	s.Require().NoError(err)
	s.Require().Len(snapshot.Edges.Relationships, 1)
	s.Equal(rel.ID, *snapshot.Edges.Relationships[0].KnowledgeRelationshipID)
}

func (s *SystemTopologyServiceSuite) TestListEntitiesIncludesRelationshipEdgesForCounts() {
	ctx := s.SeedTenantContext()
	svc := s.service()
	source := s.createSystemTopologyEntity("Checkout API " + uuid.NewString())
	target := s.createSystemTopologyEntity("Inventory API " + uuid.NewString())
	s.createRelationship(source, target)

	result, err := svc.ListEntities(ctx, rez.ListSystemTopologyEntitiesParams{
		ListParams: ent.ListParams{
			Limit: 10,
		},
		Kinds: []string{"service"},
	})
	s.Require().NoError(err)

	relationshipsByEntityID := map[uuid.UUID]int{}
	for _, entity := range result.Data {
		relationshipsByEntityID[entity.ID] += len(entity.Edges.SourceRelationships)
		relationshipsByEntityID[entity.ID] += len(entity.Edges.TargetRelationships)
	}
	s.Equal(1, relationshipsByEntityID[source.ID])
	s.Equal(1, relationshipsByEntityID[target.ID])

	relCount, err := s.Client().KnowledgeRelationship.Query().
		Where(knr.SourceEntityID(source.ID), knr.TargetEntityID(target.ID)).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, relCount)
}
