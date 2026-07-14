package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KnowledgeServiceProjectionSuite struct {
	test.Suite
}

func TestKnowledgeServiceProjectionSuite(t *testing.T) {
	suite.Run(t, &KnowledgeServiceProjectionSuite{Suite: test.NewSuite()})
}

func (s *KnowledgeServiceProjectionSuite) newKnowledgeService() *KnowledgeIngestionService {
	return NewKnowledgeIngestionService(s.Database())
}

func (s *KnowledgeServiceProjectionSuite) createNormalizedEvent(subjectKind projections.SubjectKind, providerSubjectRef string, occurredAt time.Time, attrs any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encodedAttrs, attrsErr := projections.EncodeAttributes(attrs)
	s.Require().NoError(attrsErr)
	createEvent := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("projection").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetProviderSubjectRef(providerSubjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(subjectKind.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt.Add(time.Minute)).
		SetAttributes(encodedAttrs)
	ev, err := createEvent.Save(ctx)
	s.Require().NoError(err)
	return ev
}

func (s *KnowledgeServiceProjectionSuite) TestCodeChangeProjectionPersistsEvidenceAndIsIdempotent() {
	ctx := s.SeedTenantContext()
	svc := s.newKnowledgeService()
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "repo-1",
		DisplayName:           "main@abc123",
	}
	ev := s.createNormalizedEvent(projections.SubjectKindCodeChange, "change-1", occurredAt, attrs)

	_, projErr := svc.HandleEventProjection(ctx, ev)
	s.Require().NoError(projErr)

	_, projErr = svc.HandleEventProjection(ctx, ev)
	s.Require().NoError(projErr)

	entities, keErr := s.Client(ctx).KnowledgeEntity.Query().All(ctx)
	s.Require().NoError(keErr)
	s.Len(entities, 1)

	evidence, evErr := s.Client(ctx).KnowledgeEvidence.Query().All(ctx)
	s.Require().NoError(evErr)
	s.Len(evidence, 1)
	for _, item := range evidence {
		s.Equal(ev.ID, item.EventID)
		s.NotEmpty(item.EvidenceKind)
	}

	//relationshipEvidence, err := s.Client(ctx).KnowledgeEvidence.Query().
	//	Where(knev.SubjectTypeEQ(knev.SubjectTypeRelationship)).
	//	Only(ctx)
	//s.Require().NoError(err)
	//s.Require().NotNil(relationshipEvidence.RelationshipID)
	//s.Equal(relationships[0].ID, *relationshipEvidence.RelationshipID)
	//s.Equal(assertionCodeChangeTouchedRepository, relationshipEvidence.Assertion)
}

func (s *KnowledgeServiceProjectionSuite) TestPlaceholderRepositoryIsEnrichedByRepositoryObservation() {
	ctx := s.SeedTenantContext()
	svc := s.newKnowledgeService()
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	attrsChange := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "repo-3",
		DisplayName:           "main@def456",
	}
	change := s.createNormalizedEvent(projections.SubjectKindCodeChange, "change-2", occurredAt, attrsChange)

	attrsRepo := projections.CodeForgeSubjectAttributes{
		DisplayName: "Repository Three",
		URL:         "https://example.test/repo-3",
	}
	repo := s.createNormalizedEvent(projections.SubjectKindCodeForge, "repo-3", occurredAt.Add(time.Hour), attrsRepo)

	_, projChangeErr := svc.HandleEventProjection(ctx, change)
	s.Require().NoError(projChangeErr)

	_, projRepoErr := svc.HandleEventProjection(ctx, repo)
	s.Require().NoError(projRepoErr)

	alias, err := s.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ProviderSubjectRef("repo-3")).
		WithEntity().
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(alias.Edges.Entity)
	s.Equal("Repository Three", alias.Edges.Entity.DisplayName)
	//s.Equal("https://example.test/repo-3", alias.Edges.Entity.Properties["url"])
}

func TestProjectCodeChangeEventMapsRelatedEntities(t *testing.T) {
	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "myorg/api",
		DisplayName:           "PR #1 tune retries",
		RelatedEntities: []projections.RelatedEntityRef{
			{
				ExternalRef: "demo:component:search_api",
				Kind:        "service",
				DisplayName: "Search API",
			},
		},
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		Provider:           "demo",
		ProviderSource:     "code_changes",
		ProviderSubjectRef: "demo:code_change:pr-1",
		SubjectKind:        projections.SubjectKindCodeChange.String(),
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectCodeChangeEvent(&projections.CodeChangeEvent{Event: ev, Attributes: attrs})

	require.Len(t, result, 2)
	//require.Len(t, result.Relationships, 2)
	//related := result.Relationships[1]
	//assert.Equal(t, relationshipKindRelatedTo, related.Kind)
	//assert.Equal(t, assertionCodeChangeRelatedEntity, related.EvidenceAssertion)
	//assert.Equal(t, "demo:component:search_api", related.ToAliasRef.ProviderSubjectRef)
}

func TestProjectSystemComponentObservedMapsToEntityEvidence(t *testing.T) {
	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "demo:component:search_api",
		Kind:        "service",
		DisplayName: "Search API",
		Description: "Product search query API.",
		Properties: map[string]any{
			"criticality": "high",
		},
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		Provider:           "demo",
		ProviderSource:     "system_topology",
		ProviderSubjectRef: "demo:component:search_api",
		SubjectKind:        projections.SubjectKindSystemComponent.String(),
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectSystemComponentEvent(&projections.SystemComponentEvent{Event: ev, Attributes: attrs})

	require.Len(t, result, 1)
}

func TestProjectSystemRelationshipObservedMapsEndpointsAndRelationshipEvidence(t *testing.T) {
	attrs := projections.SystemRelationshipSubjectAttributes{
		ExternalRef:       "demo:relationship:checkout_service:calls:search_api",
		Kind:              "calls",
		DisplayName:       "Checkout Service calls Search API",
		SourceExternalRef: "demo:component:checkout_service",
		SourceKind:        "service",
		SourceDisplayName: "Checkout Service",
		TargetExternalRef: "demo:component:search_api",
		TargetKind:        "service",
		TargetDisplayName: "Search API",
		Properties: map[string]any{
			"critical_path": true,
		},
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		Provider:           "demo",
		ProviderSource:     "system_topology",
		ProviderSubjectRef: "demo:relationship:checkout_service:calls:search_api",
		SubjectKind:        projections.SubjectKindSystemRelationship.String(),
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectSystemRelationshipEvent(&projections.SystemRelationshipEvent{Event: ev, Attributes: attrs})

	require.Len(t, result, 3)
}
