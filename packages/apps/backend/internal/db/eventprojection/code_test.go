package eventprojection

import (
	"time"

	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) TestCodeChangeProjectionPersistsEvidenceAndIsIdempotent() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	repoRef := "repo-1"
	codeChangeRef := "code-change-1"
	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: repoRef,
		DisplayName:           "main@abc123",
	}
	event := s.createNormalizedEvent(projections.SubjectKindCodeChange, codeChangeRef, ne.KindObserved, occurredAt, attrs)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	_, projectErr = runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	queryEntities := s.Client(ctx).KnowledgeEntity.Query()
	entityCount, entityErr := queryEntities.Count(ctx)
	s.Require().NoError(entityErr)
	s.Equal(2, entityCount)

	queryRelations := s.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Kind(knowledgeRelationshipKindTouchedRepository))
	relationshipCount, relationshipErr := queryRelations.Count(ctx)
	s.Require().NoError(relationshipErr)
	s.Equal(1, relationshipCount)

	queryEvidence := s.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EventID(event.ID)).
		WithSubjectAlias()
	evidence, evidenceErr := queryEvidence.All(ctx)
	s.Require().NoError(evidenceErr)
	s.Len(evidence, 3)

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
