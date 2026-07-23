package eventprojection

import (
	"time"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) TestOlderObservationDoesNotReplaceLiveState() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()

	baseTime := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	newerAttrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "component:api",
		Kind:        "service",
		DisplayName: "API v2",
		Properties:  map[string]any{"version": "v2"},
	}
	newer := s.createNormalizedEvent(projections.SubjectKindSystemComponent, "component:api", ne.KindObserved, baseTime, newerAttrs)

	olderAttrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "component:api",
		Kind:        "service",
		DisplayName: "API v1",
		Properties:  map[string]any{"version": "v1"},
	}
	older := s.createNormalizedEvent(projections.SubjectKindSystemComponent, "component:api", ne.KindObserved, baseTime.Add(-time.Hour), olderAttrs)

	_, projectErr := runProjection(ctx, service, newer)
	s.Require().NoError(projectErr)
	_, projectErr = runProjection(ctx, service, older)
	s.Require().NoError(projectErr)

	queryAlias := s.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.SubjectKindEQ(ksa.SubjectKindEntity), ksa.ProviderSubjectRef("component:api")).
		WithEntity()
	alias, queryErr := queryAlias.Only(ctx)
	s.Require().NoError(queryErr)
	s.Equal("API v2", alias.Edges.Entity.DisplayName)
	s.Equal("v2", alias.Edges.Entity.LiveProperties["version"])
	s.True(alias.FirstObservedAt.Equal(older.OccurredAt))
	s.True(alias.LastObservedAt.Equal(newer.OccurredAt))
}

func (s *ProjectionServiceSuite) TestDeletionAndLaterObservationUpdateAliasState() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()

	subjRef := "deleted-component"
	baseTime := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "component:deleted",
		Kind:        "service",
		DisplayName: "Test Service",
	}

	observed := s.createNormalizedEvent(projections.SubjectKindSystemComponent, subjRef, ne.KindObserved, baseTime, attrs)
	_, projectErr := runProjection(ctx, service, observed)
	s.Require().NoError(projectErr)

	deleted := s.createNormalizedEvent(projections.SubjectKindSystemComponent, subjRef, ne.KindDeleted, baseTime.Add(time.Hour), attrs)
	_, projectErr = runProjection(ctx, service, deleted)
	s.Require().NoError(projectErr)

	queryAlias := s.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.SubjectKindEQ(ksa.SubjectKindEntity), ksa.ProviderSubjectRef(subjRef))
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr)
	s.Require().NotNil(alias.DeletedAt)
	s.True(alias.DeletedAt.Equal(deleted.OccurredAt))

	queryEvidence := s.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EventID(deleted.ID))
	deletedEvidence, evidenceErr := queryEvidence.Only(ctx)
	s.Require().NoError(evidenceErr)
	s.Equal(ke.EvidenceKindDeleted, deletedEvidence.EvidenceKind)

	restored := s.createNormalizedEvent(projections.SubjectKindSystemComponent, subjRef, ne.KindObserved, baseTime.Add(2*time.Hour), attrs)
	_, projectErr = runProjection(ctx, service, restored)
	s.Require().NoError(projectErr)

	restoredAlias, restoredAliasErr := s.Client(ctx).KnowledgeSubjectAlias.Get(ctx, alias.ID)
	s.Require().NoError(restoredAliasErr)
	s.Nil(restoredAlias.DeletedAt)
}

func (s *ProjectionServiceSuite) TestSystemRelationshipProjectsEndpointsAndEvidence() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()
	attrs := projections.SystemRelationshipSubjectAttributes{
		ExternalRef:       "checkout:calls:search",
		Kind:              "calls",
		SourceExternalRef: "component:checkout",
		SourceKind:        "service",
		SourceDisplayName: "Checkout",
		TargetExternalRef: "component:search",
		TargetKind:        "service",
		TargetDisplayName: "Search",
		Properties:        map[string]any{"critical_path": true},
	}
	event := s.createNormalizedEvent(projections.SubjectKindSystemRelationship, "checkout:calls:search", ne.KindObserved, time.Now(), attrs)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	entityCount, entityErr := s.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ReferenceIn("component:checkout", "component:search")).
		Count(ctx)
	s.Require().NoError(entityErr)
	s.Equal(2, entityCount)
	relationshipCount, relationshipErr := s.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Kind("calls"), knr.HasSourceEntityWith(kne.Reference("component:checkout"))).
		Count(ctx)
	s.Require().NoError(relationshipErr)
	s.Equal(1, relationshipCount)
	evidenceCount, evidenceErr := s.Client(ctx).KnowledgeEvidence.Query().Where(ke.EventID(event.ID)).Count(ctx)
	s.Require().NoError(evidenceErr)
	s.Equal(3, evidenceCount)
}
