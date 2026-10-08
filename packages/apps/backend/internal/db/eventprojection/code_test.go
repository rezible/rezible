package eventprojection

import (
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) TestCodeChangeProjectionPersistsEvidenceAndIsIdempotent() {
	ctx, tdb := s.SetupTestDatabase()
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
				Ref: rez.ProviderResourceRef{
					Provider:          "test",
					ProviderNamespace: "account-a",
					ResourceRef:       "service-1",
				},
				Category:    kne.CategoryContainer,
				Kind:        "service",
				DisplayName: "Service A",
			},
			{
				Ref: rez.ProviderResourceRef{
					Provider:          "test",
					ProviderNamespace: "account-b",
					ResourceRef:       "service-1",
				},
				Category:    kne.CategoryContainer,
				Kind:        "service",
				DisplayName: "Service B",
			},
		},
	}
	event := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, codeChangeRef, occurredAt, attrs)

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
	queryImpactRelationships := client.KnowledgeRelationship.Query().
		Where(knr.PredicateEQ(knr.PredicateImpacts)).
		WithAliases()
	impactRelationships, impactErr := queryImpactRelationships.All(ctx)
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

func (s *ProjectionServiceSuite) TestPullRequestMergeIsRecorded() {
	ctx, tdb := s.SetupTestDatabase()
	service := s.projectionService(tdb)
	mergedAt := time.Date(2026, 10, 7, 13, 50, 0, 0, time.UTC)
	// The merge's time decides when the change happened, whatever time its event carries.
	occurredAt := mergedAt.Add(time.Minute)
	pushedAt := mergedAt.Add(2 * time.Minute)

	repository := githubRepositoryObservation("42", "Acme/Checkout", "acme/checkout")
	merged := projections.CodeChangeEventAttributes{
		Repository:  repository,
		DisplayName: "Raise checkout timeout",
		Merge: &projections.CodeChangeMerge{
			MergedAt:          mergedAt,
			MergeCommitSha:    "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392",
			BaseRef:           "main",
			IntoDefaultBranch: true,
			Number:            7,
			URL:               "https://github.com/Acme/Checkout/pull/7",
			AuthorLogin:       "octocat",
		},
	}
	push := projections.CodeChangeEventAttributes{
		Repository:  repository,
		DisplayName: "refs/heads/main",
	}

	mergeEvent := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, "change:42:pr:7", occurredAt, merged)
	_, projectErr := runProjection(ctx, service, mergeEvent)
	s.Require().NoError(projectErr)

	pushEvent := s.createNormalizedEvent(ctx, tdb, projections.KindCodeChange, "change:42:abc123", pushedAt, push)
	_, projectErr = runProjection(ctx, service, pushEvent)
	s.Require().NoError(projectErr)

	queryChangeAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderResourceRef("change:42:pr:7"))
	changeAlias, aliasErr := queryChangeAlias.Only(ctx)
	s.Require().NoError(aliasErr)
	change, changeErr := tdb.Client(ctx).KnowledgeEntity.Get(ctx, *changeAlias.EntityID)
	s.Require().NoError(changeErr)

	s.Equal(kne.CategoryEvent, change.Category)
	s.Equal(knowledgeEntityKindCodeChange, change.Kind)
	s.Equal("Raise checkout timeout", change.State.DisplayName)
	s.Require().NotNil(change.StateEffectiveAt)
	s.True(mergedAt.Equal(*change.StateEffectiveAt), "the change is an event at its merge time")
	expected := map[string]any{
		"merged_at":           "2026-10-07T13:50:00Z",
		"merge_commit_sha":    "4f1c0d9e8b7a6f5e4d3c2b1a09f8e7d6c5b4a392",
		"base_ref":            "main",
		"into_default_branch": "true",
		"number":              "7",
		"url":                 "https://github.com/Acme/Checkout/pull/7",
		"author":              "octocat",
	}
	s.Equal(expected, change.State.Properties)

	queryPushAlias := tdb.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderResourceRef("change:42:abc123"))
	pushAlias, pushAliasErr := queryPushAlias.Only(ctx)
	s.Require().NoError(pushAliasErr)
	pushed, pushedErr := tdb.Client(ctx).KnowledgeEntity.Get(ctx, *pushAlias.EntityID)
	s.Require().NoError(pushedErr)
	s.Empty(pushed.State.Properties, "a push records no merge")
}
