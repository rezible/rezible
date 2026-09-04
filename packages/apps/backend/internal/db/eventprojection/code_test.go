package eventprojection

import (
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) TestCodeChangeProjectionPersistsEvidenceAndIsIdempotent() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	service := s.projectionService(tdb)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	repoRef := "repo-1"
	codeChangeRef := "code-change-1"
	repositoryResourceRef := rez.ProviderResourceRef{
		Provider:          "test",
		ProviderNamespace: "projection-tests",
		ResourceRef:       repoRef,
	}
	attrs := projections.CodeChangeEventAttributes{
		Repository: projections.EntityObservation{
			Ref:         repositoryResourceRef,
			Category:    kne.CategoryCode,
			Kind:        knowledgeEntityKindRepository,
			DisplayName: "Repository One",
		},
		DisplayName: "main@abc123",
		ImpactedEntities: []projections.EntityObservation{
			{
				Ref:         rez.ProviderResourceRef{Provider: "test", ProviderNamespace: "account-a", ResourceRef: "service-1"},
				Category:    kne.CategoryContainer,
				Kind:        "service",
				DisplayName: "Service A",
			},
			{
				Ref:         rez.ProviderResourceRef{Provider: "test", ProviderNamespace: "account-b", ResourceRef: "service-1"},
				Category:    kne.CategoryContainer,
				Kind:        "service",
				DisplayName: "Service B",
			},
		},
	}
	event := s.createNormalizedEvent(tdb, projections.KindCodeChange, codeChangeRef, occurredAt, attrs)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	_, projectErr = runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	client := tdb.Client(ctx)
	queryEntities := client.KnowledgeEntity.Query()
	entityCount, entityErr := queryEntities.Count(ctx)
	s.Require().NoError(entityErr)
	s.Equal(4, entityCount)

	queryRelations := client.KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateTouches))
	relationshipCount, relationshipErr := queryRelations.Count(ctx)
	s.Require().NoError(relationshipErr)
	s.Equal(1, relationshipCount)
	impactRelationships, impactErr := client.KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateImpacts)).
		WithAliases().
		All(ctx)
	s.Require().NoError(impactErr)
	s.Require().Len(impactRelationships, 2)
	s.NotEqual(impactRelationships[0].Edges.Aliases[0].ProviderResourceRef, impactRelationships[1].Edges.Aliases[0].ProviderResourceRef)
	for _, relationship := range impactRelationships {
		s.Equal("rezible", relationship.Edges.Aliases[0].Provider)
		s.Empty(relationship.Edges.Aliases[0].ProviderNamespace)
	}

	queryEvidence := client.KnowledgeEvidence.Query().
		Where(ke.EventID(event.ID)).
		WithSubjectAlias()
	evidence, evidenceErr := queryEvidence.All(ctx)
	s.Require().NoError(evidenceErr)
	s.Len(evidence, 7)

	var relationshipEvidence *ent.KnowledgeEvidence
	for _, item := range evidence {
		if item.Assertion == knowledgeAssertionCodeChangeRepository {
			relationshipEvidence = item
			break
		}
	}
	s.Require().NotNil(relationshipEvidence)
	s.NotNil(relationshipEvidence.Edges.SubjectAlias.RelationshipID)
}
