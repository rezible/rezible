package eventprojection

import (
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
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

	queryEntities := s.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ReferenceIn(codeChangeRef, repoRef))
	entityCount, entityErr := queryEntities.Count(ctx)
	s.Require().NoError(entityErr)
	s.Equal(2, entityCount)

	queryRelations := s.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Kind(knowledgeRelationshipKindTouchedRepository), knr.HasSourceEntityWith(kne.Reference(codeChangeRef)))
	relationshipCount, relationshipErr := queryRelations.Count(ctx)
	s.Require().NoError(relationshipErr)
	s.Equal(1, relationshipCount)

	queryEvidence := s.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.EventID(event.ID)).
		WithAlias()
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
	s.NotEqual(uuid.Nil, relationshipEvidence.Edges.Alias.RelationshipID)
}

func (s *ProjectionServiceSuite) TestRepositoryObservationEnrichesPlaceholderWithoutLaterDegradation() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()

	baseTime := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)

	changeAttrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "repo-3",
		DisplayName:           "main@def456",
	}
	change := s.createNormalizedEvent(projections.SubjectKindCodeChange, "change-2", ne.KindObserved, baseTime, changeAttrs)
	_, projectErr := runProjection(ctx, service, change)
	s.Require().NoError(projectErr)

	repositoryAttrs := projections.CodeForgeSubjectAttributes{
		DisplayName: "Repository Three",
		URL:         "https://example.test/repo-3",
	}
	repository := s.createNormalizedEvent(projections.SubjectKindCodeForge, "repo-3", ne.KindObserved, baseTime.Add(time.Hour), repositoryAttrs)
	_, projectErr = runProjection(ctx, service, repository)
	s.Require().NoError(projectErr)

	laterChangeAttrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "repo-3",
		DisplayName:           "main@ghi789",
	}
	laterChange := s.createNormalizedEvent(projections.SubjectKindCodeChange, "change-3", ne.KindObserved, baseTime.Add(2*time.Hour), laterChangeAttrs)
	_, projectErr = runProjection(ctx, service, laterChange)
	s.Require().NoError(projectErr)

	queryAlias := s.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderSubjectRef("repo-3")).
		WithEntity()
	alias, queryErr := queryAlias.Only(ctx)
	s.Require().NoError(queryErr)
	s.Equal("Repository Three", alias.Edges.Entity.DisplayName)
	s.Equal("https://example.test/repo-3", alias.Edges.Entity.LiveProperties["url"])
}
