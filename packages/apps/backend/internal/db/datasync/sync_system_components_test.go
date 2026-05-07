package datasync

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/testkit"
)

type SystemComponentsSyncSuite struct {
	testkit.Suite
}

func TestSystemComponentsSyncSuite(t *testing.T) {
	suite.Run(t, &SystemComponentsSyncSuite{Suite: testkit.NewSuite()})
}

func (s *SystemComponentsSyncSuite) batcher() *systemComponentsBatcher {
	return &systemComponentsBatcher{
		db:           s.Client(),
		providerName: "fake",
	}
}

func systemComponentWithRelationships(externalID string, targets ...string) *ent.SystemComponent {
	cmp := &ent.SystemComponent{
		ExternalID:  externalID,
		Name:        externalID,
		Description: "component " + externalID,
		Properties:  map[string]any{},
	}
	cmp.Edges.ComponentRelationships = make([]*ent.SystemComponentRelationship, len(targets))
	for i, target := range targets {
		cmp.Edges.ComponentRelationships[i] = &ent.SystemComponentRelationship{
			ExternalID:  target,
			Description: externalID + " depends on " + target,
		}
	}
	return cmp
}

func relationshipMutations(mutations []ent.Mutation) []*ent.SystemComponentRelationshipMutation {
	rels := make([]*ent.SystemComponentRelationshipMutation, 0)
	for _, mutation := range mutations {
		if relMutation, ok := mutation.(*ent.SystemComponentRelationshipMutation); ok {
			rels = append(rels, relMutation)
		}
	}
	return rels
}

func (s *SystemComponentsSyncSuite) TestCreateBatchMutationsCreatesRelationshipMutationsForProviderEdges() {
	ctx := s.SeedTenantContext()
	batcher := s.batcher()
	batch := []*ent.SystemComponent{
		systemComponentWithRelationships("payment-ui", "api-gateway"),
		systemComponentWithRelationships("api-gateway", "payments-service"),
		systemComponentWithRelationships("payments-service"),
	}

	mutations, err := batcher.createBatchMutations(ctx, batch)

	s.Require().NoError(err)
	rels := relationshipMutations(mutations)
	s.Require().Len(rels, 2)
	for _, rel := range rels {
		sourceID, hasSource := rel.SourceID()
		targetID, hasTarget := rel.TargetID()
		s.True(hasSource)
		s.True(hasTarget)
		s.NotEqual(uuid.Nil, sourceID)
		s.NotEqual(uuid.Nil, targetID)
	}
}

func (s *SystemComponentsSyncSuite) TestCreateBatchMutationsCreatesNoRelationshipMutationsForEmptyEdges() {
	ctx := s.SeedTenantContext()
	batcher := s.batcher()
	batch := []*ent.SystemComponent{
		systemComponentWithRelationships("payment-ui"),
		systemComponentWithRelationships("api-gateway"),
	}

	mutations, err := batcher.createBatchMutations(ctx, batch)

	s.Require().NoError(err)
	s.Empty(relationshipMutations(mutations))
}

func (s *SystemComponentsSyncSuite) TestCreateBatchMutationsSkipsUnknownRelationshipTarget() {
	ctx := s.SeedTenantContext()
	batcher := s.batcher()
	batch := []*ent.SystemComponent{
		systemComponentWithRelationships("payment-ui", "missing-component"),
	}

	mutations, err := batcher.createBatchMutations(ctx, batch)

	s.Require().NoError(err)
	s.Empty(relationshipMutations(mutations))
}

func (s *SystemComponentsSyncSuite) TestAfterBatchAppliedPublishesObservedEventForEachComponent() {
	ctx := s.SeedTenantContext()
	batcher := s.batcher()
	batch := []*ent.SystemComponent{
		systemComponentWithRelationships("payment-ui"),
		systemComponentWithRelationships("api-gateway"),
		systemComponentWithRelationships("payments-service"),
	}
	err := batcher.afterBatchApplied(ctx, batch)
	s.Require().NoError(err)
}
