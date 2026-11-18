package eventprojection

import (
	"time"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) TestCodeChangeProjection() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)
	mergedAt := time.Date(2026, 10, 7, 13, 50, 0, 0, time.UTC)
	impacted := projections.EntityObservation{
		Ref:      rez.ProviderResourceRef{Provider: "test", ProviderNamespace: "projection-tests", ResourceRef: "service:checkout"},
		Category: kne.CategoryContainer, Kind: "service", DisplayName: "Checkout",
		Description: "Original description", Properties: map[string]any{"criticality": "high"},
	}
	topology := projections.SystemComponentEventAttributes{Category: impacted.Category, Kind: impacted.Kind, DisplayName: impacted.DisplayName, Description: impacted.Description, Properties: impacted.Properties}
	topologyEvent := s.createNormalizedEvent(ctx, tdb, projections.KindSystemComponent, impacted.Ref.ResourceRef, mergedAt.Add(-time.Hour), topology)
	_, topologyErr := runProjection(ctx, service, topologyEvent)
	s.Require().NoError(topologyErr)
	impacted.Description = "Untrusted change description"
	impacted.Properties = map[string]any{"criticality": "low"}
	repository := githubRepositoryObservation("42", "Acme/Checkout", "acme/checkout")
	attributes := projections.CodeChangeEventAttributes{
		Repository: repository, DisplayName: "Raise checkout timeout",
		// The repository takes priority when the same alias appears in both roles.
		ImpactedEntities: []projections.EntityObservation{impacted, repository},
		Merge: &projections.CodeChangeMerge{
			MergedAt: mergedAt, MergeCommitSha: "abc123", BaseRef: "main", IntoDefaultBranch: true, Number: 7,
		},
	}
	event := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, "change:42:pr:7", mergedAt.Add(time.Minute), attributes)
	for range 2 {
		_, projectErr := runProjection(ctx, service, event)
		s.Require().NoError(projectErr)
	}
	point, pointErr := tdb.Client(ctx).ChangePoint.Query().Only(ctx)
	s.Require().NoError(pointErr)
	s.Equal(event.ID, point.LatestEventID)
	s.True(mergedAt.Equal(point.OccurredAt))
	services := s.entitiesOfKind(ctx, tdb, kne.CategoryContainer, "service")
	s.Require().Len(services, 1)
	s.Equal("Original description", services[0].State.Description)
	s.Equal(map[string]any{"criticality": "high"}, services[0].State.Properties)
	repositories := s.entitiesOfKind(ctx, tdb, kne.CategoryCode, "repository")
	s.Require().Len(repositories, 1)
	s.assertInvolvement(ctx, tdb, projections.AssertionChangeRepository, repositories[0].ID, 1)
	s.assertInvolvement(ctx, tdb, projections.AssertionChangeImpacted, services[0].ID, 1)

	attributes.Merge = nil
	attributes.ImpactedEntities = nil
	push := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, "change:42:push:abc", mergedAt.Add(2*time.Minute), attributes)
	_, pushErr := runProjection(ctx, service, push)
	s.Require().NoError(pushErr)
	points, countErr := tdb.Client(ctx).ChangePoint.Query().Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, points, "a push creates no change point")
	entities, entityErr := tdb.Client(ctx).KnowledgeEntity.Query().Count(ctx)
	s.Require().NoError(entityErr)
	s.Equal(2, entities, "changes create only involved entities")
}
