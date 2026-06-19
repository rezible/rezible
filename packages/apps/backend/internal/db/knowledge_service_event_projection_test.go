package db

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/projections"
	"github.com/rezible/rezible/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KnowledgeServiceProjectionSuite struct {
	testkit.Suite
}

func TestKnowledgeServiceProjectionSuite(t *testing.T) {
	suite.Run(t, &KnowledgeServiceProjectionSuite{Suite: testkit.NewSuite()})
}

func (s *KnowledgeServiceProjectionSuite) newKnowledgeService() *KnowledgeService {
	return NewKnowledgeService(s.Database())
}

func (s *KnowledgeServiceProjectionSuite) createNormalizedEvent(subjectKind projections.SubjectKind, providerSubjectRef string, occurredAt time.Time, attrs map[string]any) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	ev, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("projection").
		SetProviderEventRef("event-" + uuid.NewString()).
		SetProviderSubjectRef(providerSubjectRef).
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(subjectKind.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt.Add(time.Minute)).
		SetAttributes(attrs).
		Save(ctx)
	s.Require().NoError(err)
	return ev
}

func (s *KnowledgeServiceProjectionSuite) mustEncodeAttrs(attrs any) map[string]any {
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	return encoded
}

func (s *KnowledgeServiceProjectionSuite) TestCodeChangeProjectionPersistsEvidenceAndIsIdempotent() {
	ctx := s.SeedTenantContext()
	svc := s.newKnowledgeService()
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	attrs := s.mustEncodeAttrs(projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "repo-1",
		DisplayName:           "main@abc123",
	})
	ev := s.createNormalizedEvent(projections.SubjectKindCodeChange, "change-1", occurredAt, attrs)

	s.Require().NoError(svc.HandleEventProjection(ctx, ev))
	s.Require().NoError(svc.HandleEventProjection(ctx, ev))

	entities, err := s.Client(ctx).KnowledgeEntity.Query().
		Where(kne.KindIn(knowledgeKindCodeChange, knowledgeKindCodeRepository)).
		All(ctx)
	s.Require().NoError(err)
	s.Len(entities, 2)

	relationships, err := s.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Kind(relationshipKindTouched)).
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(relationships, 1)
	s.True(relationships[0].FirstObservedAt.Equal(occurredAt))
	s.True(relationships[0].LastObservedAt.Equal(occurredAt))

	evidence, err := s.Client(ctx).KnowledgeEvidence.Query().All(ctx)
	s.Require().NoError(err)
	s.Len(evidence, 3)
	for _, item := range evidence {
		s.Equal(ev.ID, item.EventID)
		s.True(item.ObservedAt.Equal(occurredAt))
		s.NotEmpty(item.SubjectType)
	}

	relationshipEvidence, err := s.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.SubjectTypeEQ(knev.SubjectTypeRelationship)).
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(relationshipEvidence.RelationshipID)
	s.Equal(relationships[0].ID, *relationshipEvidence.RelationshipID)
	s.Equal(assertionCodeChangeTouchedRepository, relationshipEvidence.Assertion)
}

func (s *KnowledgeServiceProjectionSuite) TestEntityProjectionRecordsChangedEvidenceWhenPropertiesChange() {
	ctx := s.SeedTenantContext()
	svc := s.newKnowledgeService()

	subjectRef := "repo-2"

	firstAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	firstAttrs := s.mustEncodeAttrs(projections.CodeForgeSubjectAttributes{
		DisplayName: "repo-2",
		URL:         "https://example.test/repo-2",
	})
	first := s.createNormalizedEvent(projections.SubjectKindCodeForge, subjectRef, firstAt, firstAttrs)
	s.Require().NoError(svc.HandleEventProjection(ctx, first))

	secondAt := firstAt.Add(time.Hour)
	secondAttrs := s.mustEncodeAttrs(projections.CodeForgeSubjectAttributes{
		DisplayName: "repo-2 renamed",
		URL:         "https://example.test/repo-2-renamed",
	})
	second := s.createNormalizedEvent(projections.SubjectKindCodeForge, subjectRef, secondAt, secondAttrs)
	s.Require().NoError(svc.HandleEventProjection(ctx, second))

	dbc := s.Client(ctx)
	queryAlias := dbc.KnowledgeEntityAlias.Query().
		Where(knea.ProviderSubjectRef(subjectRef)).
		WithEntity()
	alias, aliasErr := queryAlias.Only(ctx)
	s.Require().NoError(aliasErr)

	entity := alias.Edges.Entity
	s.Require().NotNil(entity)
	s.Equal("repo-2 renamed", entity.DisplayName)
	s.True(entity.FirstObservedAt.Equal(firstAt))
	s.True(entity.LastObservedAt.Equal(secondAt))

	queryEvidence := s.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.EntityID(entity.ID)).
		Order(ent.Asc(knev.FieldObservedAt))
	evidence, evidenceErr := queryEvidence.All(ctx)
	s.Require().NoError(evidenceErr)
	s.Require().Len(evidence, 2)
	s.Equal(knev.EvidenceKindObserved, evidence[0].EvidenceKind)
	s.Equal(knev.EvidenceKindChanged, evidence[1].EvidenceKind)
}

func (s *KnowledgeServiceProjectionSuite) TestPlaceholderRepositoryIsEnrichedByRepositoryObservation() {
	ctx := s.SeedTenantContext()
	svc := s.newKnowledgeService()
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	change := s.createNormalizedEvent(projections.SubjectKindCodeChange, "change-2", occurredAt, s.mustEncodeAttrs(projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "repo-3",
		DisplayName:           "main@def456",
	}))
	repo := s.createNormalizedEvent(projections.SubjectKindCodeForge, "repo-3", occurredAt.Add(time.Hour), s.mustEncodeAttrs(projections.CodeForgeSubjectAttributes{
		DisplayName: "Repository Three",
		URL:         "https://example.test/repo-3",
	}))

	s.Require().NoError(svc.HandleEventProjection(ctx, change))
	s.Require().NoError(svc.HandleEventProjection(ctx, repo))

	alias, err := s.Client(ctx).KnowledgeEntityAlias.Query().
		Where(knea.ProviderSubjectRef("repo-3")).
		WithEntity().
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(alias.Edges.Entity)
	s.Equal("Repository Three", alias.Edges.Entity.DisplayName)
	s.Equal("https://example.test/repo-3", alias.Edges.Entity.Properties["url"])
}

func (s *KnowledgeServiceProjectionSuite) TestAliasConflictFailsLoudly() {
	ctx := s.SeedTenantContext()
	svc := s.newKnowledgeService()
	first, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName("Service A").
		Save(ctx)
	s.Require().NoError(err)
	second, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName("Service B").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).KnowledgeEntityAlias.Create().
		SetEntityID(first.ID).
		SetProvider("test").
		SetProviderSubjectRef("service-a").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).KnowledgeEntityAlias.Create().
		SetEntityID(second.ID).
		SetProvider("test").
		SetProviderSubjectRef("service-b").
		Save(ctx)
	s.Require().NoError(err)

	attrs := s.mustEncodeAttrs(projections.SystemComponentSubjectAttributes{
		ExternalRef: "service-a",
		Kind:        "service",
		DisplayName: "Service A",
	})
	ev := s.createNormalizedEvent(projections.SubjectKindSystemComponent, "service-a", time.Now().UTC(), attrs)

	proj := rez.ProjectedKnowledgeEntity{
		Kind:              "service",
		DisplayName:       "Service A",
		EvidenceAssertion: assertionSystemComponentExists,
		AliasRefs: []ent.KnowledgeEntityAliasRef{
			{Provider: "test", ProviderSubjectRef: "service-a"},
			{Provider: "test", ProviderSubjectRef: "service-b"},
		},
	}
	_, err = svc.ResolveProjectedEntity(ctx, ev, proj)
	s.Require().Error(err)
	s.True(strings.Contains(err.Error(), "different entities"))
}

func TestProjectCodeForgeObservedMapsToRepositoryFactEvidence(t *testing.T) {
	attrs := projections.CodeForgeSubjectAttributes{
		DisplayName: "myorg/api",
		URL:         "https://github.com/myorg/api",
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		SubjectKind:        projections.SubjectKindCodeForge.String(),
		Provider:           "github",
		ProviderSource:     "repositories",
		ProviderSubjectRef: "myorg/api",
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)
	proj := newKnowledgeEntityEventProjector(ev, nil)

	result := proj.projectCodeForgeEvent(&projections.CodeForgeEvent{Event: ev, Attributes: attrs})

	require.Len(t, result.Entities, 1)
	assert.Empty(t, result.Relationships)

	entity := result.Entities[0]
	assert.Equal(t, knowledgeKindCodeRepository, entity.Kind)
	assert.Equal(t, assertionCodeRepositoryExists, entity.EvidenceAssertion)
	assert.Equal(t, "myorg/api", entity.DisplayName)
	require.Len(t, entity.AliasRefs, 1)
	assert.Equal(t, ev.MakeEntityAliasRef(), entity.AliasRefs[0])
}

func TestProjectCodeChangeEventObservedMapsChangeRepositoryRelationshipEvidence(t *testing.T) {
	attrs := projections.CodeChangeSubjectAttributes{
		RepositoryExternalRef: "myorg/api",
		DisplayName:           "refs/heads/main",
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		Provider:           "github",
		ProviderSource:     "push",
		ProviderSubjectRef: "github:myorg/api:abc123",
		SubjectKind:        projections.SubjectKindCodeChange.String(),
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectCodeChangeEvent(&projections.CodeChangeEvent{Event: ev, Attributes: attrs})

	require.Len(t, result.Entities, 2)
	assert.Equal(t, knowledgeKindCodeChange, result.Entities[0].Kind)
	assert.Equal(t, assertionCodeChangeObserved, result.Entities[0].EvidenceAssertion)
	assert.Equal(t, knowledgeKindCodeRepository, result.Entities[1].Kind)
	assert.Equal(t, assertionCodeRepositoryExists, result.Entities[1].EvidenceAssertion)
	assert.True(t, result.Entities[1].IsPlaceholder)

	require.Len(t, result.Relationships, 1)
	rel := result.Relationships[0]
	assert.Equal(t, relationshipKindTouched, rel.Kind)
	assert.Equal(t, assertionCodeChangeTouchedRepository, rel.EvidenceAssertion)
	assert.Equal(t, result.Entities[0].AliasRefs[0], rel.FromAliasRef)
	assert.Equal(t, result.Entities[1].AliasRefs[0], rel.ToAliasRef)
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

	require.Len(t, result.Entities, 3)
	require.Len(t, result.Relationships, 2)
	related := result.Relationships[1]
	assert.Equal(t, relationshipKindRelatedTo, related.Kind)
	assert.Equal(t, assertionCodeChangeRelatedEntity, related.EvidenceAssertion)
	assert.Equal(t, "demo:component:search_api", related.ToAliasRef.ProviderSubjectRef)
}

func TestProjectChatMessageEventMapsRelatedEntityEvidence(t *testing.T) {
	attrs := projections.ChatMessageAttributes{
		ConversationExternalRef: "#inc-checkout",
		Body:                    "Search API latency moved after the retry policy change.",
		SenderExternalRef:       "demo:user:ava",
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
		ProviderSource:     "chat_messages",
		ProviderSubjectRef: "demo:chat_message:1",
		SubjectKind:        projections.SubjectKindChatMessage.String(),
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectChatMessageEvent(&projections.ChatMessage{Event: ev, Attributes: attrs})

	require.Len(t, result.Entities, 2)
	require.Len(t, result.Relationships, 1)
	assert.Equal(t, knowledgeKindChatMessage, result.Entities[0].Kind)
	assert.Equal(t, assertionChatMessageRelatedEntity, result.Relationships[0].EvidenceAssertion)
	assert.Equal(t, "demo:component:search_api", result.Relationships[0].ToAliasRef.ProviderSubjectRef)
}

func TestProjectorObservedAtPrefersOccurredAt(t *testing.T) {
	occurredAt := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)
	receivedAt := occurredAt.Add(time.Hour)
	ev := &ent.NormalizedEvent{OccurredAt: occurredAt, ReceivedAt: receivedAt}

	assert.Equal(t, occurredAt, ev.DeriveObservedAt())
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

	require.Len(t, result.Entities, 1)
	assert.Empty(t, result.Relationships)
	entity := result.Entities[0]
	assert.Equal(t, attrs.Kind, entity.Kind)
	assert.Equal(t, assertionSystemComponentExists, entity.EvidenceAssertion)
	assert.Equal(t, attrs.DisplayName, entity.DisplayName)
	assert.Equal(t, attrs.Description, entity.Description)
	assert.Equal(t, attrs.Properties["criticality"], entity.Properties["criticality"])
	require.Len(t, entity.AliasRefs, 1)
	assert.Equal(t, ev.MakeEntityAliasRef(), entity.AliasRefs[0])
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

	require.Len(t, result.Entities, 2)
	assert.Equal(t, "service", result.Entities[0].Kind)
	assert.Equal(t, "Checkout Service", result.Entities[0].DisplayName)
	assert.True(t, result.Entities[0].IsPlaceholder)
	assert.Equal(t, "service", result.Entities[1].Kind)
	assert.Equal(t, "Search API", result.Entities[1].DisplayName)
	assert.True(t, result.Entities[1].IsPlaceholder)

	require.Len(t, result.Relationships, 1)
	relationship := result.Relationships[0]
	assert.Equal(t, "calls", relationship.Kind)
	assert.Equal(t, assertionSystemRelationshipExists, relationship.EvidenceAssertion)
	assert.Equal(t, "Checkout Service calls Search API", relationship.DisplayName)
	assert.Equal(t, result.Entities[0].AliasRefs[0], relationship.FromAliasRef)
	assert.Equal(t, result.Entities[1].AliasRefs[0], relationship.ToAliasRef)
	assert.Equal(t, true, relationship.Properties["critical_path"])
}
