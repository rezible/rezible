package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeEntityAliasRef(ev *ent.NormalizedEvent) EntityAliasRef {
	return EntityAliasRef{
		Provider:           ev.Provider,
		ProviderSource:     ev.ProviderSource,
		ProviderSubjectRef: ev.ProviderSubjectRef,
	}
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
	ref := makeEntityAliasRef(ev)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectCodeForgeEvent(&projections.CodeForgeEvent{Event: ev, Attributes: attrs})

	require.Len(t, result.Entities, 1)
	assert.Empty(t, result.Relationships)

	entity := result.Entities[0]
	assert.Equal(t, knowledgeKindCodeRepository, entity.Kind)
	assert.Equal(t, assertionCodeRepositoryExists, entity.AssertionKind)
	assert.Equal(t, "myorg/api", entity.DisplayName)
	require.Len(t, entity.Aliases, 1)
	assert.Equal(t, ref, entity.Aliases[0])
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
	assert.Equal(t, assertionCodeChangeObserved, result.Entities[0].AssertionKind)
	assert.Equal(t, knowledgeKindCodeRepository, result.Entities[1].Kind)
	assert.Equal(t, assertionCodeRepositoryExists, result.Entities[1].AssertionKind)
	assert.True(t, result.Entities[1].IsPlaceholder)

	require.Len(t, result.Relationships, 1)
	rel := result.Relationships[0]
	assert.Equal(t, relationshipKindTouched, rel.Kind)
	assert.Equal(t, assertionCodeChangeTouchedRepository, rel.AssertionKind)
	assert.Equal(t, result.Entities[0].Aliases[0], rel.FromAlias)
	assert.Equal(t, result.Entities[1].Aliases[0], rel.ToAlias)
}

func TestProjectorObservedAtPrefersOccurredAt(t *testing.T) {
	occurredAt := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)
	receivedAt := occurredAt.Add(time.Hour)
	ev := &ent.NormalizedEvent{OccurredAt: occurredAt, ReceivedAt: receivedAt}

	assert.Equal(t, occurredAt, newKnowledgeEntityEventProjector(ev, nil).resolveEventObservedAt(ev))
}

func TestProjectSystemComponentObservedMapsToEntityEvidence(t *testing.T) {
	attrs := projections.SystemComponentSubjectAttributes{
		ExternalRef: "fake:component:search_api",
		Kind:        "service",
		DisplayName: "Search API",
		Description: "Product search query API.",
		Properties: map[string]any{
			"criticality": "high",
		},
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		Provider:           "fake",
		ProviderSource:     "system_topology",
		ProviderSubjectRef: "fake:component:search_api",
		SubjectKind:        projections.SubjectKindSystemComponent.String(),
	}
	var encodeAttrsErr error
	ev.Attributes, encodeAttrsErr = projections.EncodeAttributes(attrs)
	require.NoError(t, encodeAttrsErr)
	ref := makeEntityAliasRef(ev)

	proj := newKnowledgeEntityEventProjector(ev, nil)
	result := proj.projectSystemComponentEvent(&projections.SystemComponentEvent{Event: ev, Attributes: attrs})

	require.Len(t, result.Entities, 1)
	assert.Empty(t, result.Relationships)
	entity := result.Entities[0]
	assert.Equal(t, attrs.Kind, entity.Kind)
	assert.Equal(t, assertionSystemComponentExists, entity.AssertionKind)
	assert.Equal(t, attrs.DisplayName, entity.DisplayName)
	assert.Equal(t, attrs.Description, entity.Description)
	assert.Equal(t, attrs.Properties["criticality"], entity.Properties["criticality"])
	require.Len(t, entity.Aliases, 1)
	assert.Equal(t, ref, entity.Aliases[0])
}

func TestProjectSystemRelationshipObservedMapsEndpointsAndRelationshipEvidence(t *testing.T) {
	attrs := projections.SystemRelationshipSubjectAttributes{
		ExternalRef:       "fake:relationship:checkout_service:calls:search_api",
		Kind:              "calls",
		DisplayName:       "Checkout Service calls Search API",
		SourceExternalRef: "fake:component:checkout_service",
		SourceKind:        "service",
		SourceDisplayName: "Checkout Service",
		TargetExternalRef: "fake:component:search_api",
		TargetKind:        "service",
		TargetDisplayName: "Search API",
		Properties: map[string]any{
			"critical_path": true,
		},
	}
	ev := &ent.NormalizedEvent{
		ID:                 uuid.New(),
		Provider:           "fake",
		ProviderSource:     "system_topology",
		ProviderSubjectRef: "fake:relationship:checkout_service:calls:search_api",
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
	assert.Equal(t, assertionSystemRelationshipExists, relationship.AssertionKind)
	assert.Equal(t, "Checkout Service calls Search API", relationship.DisplayName)
	assert.Equal(t, result.Entities[0].Aliases[0], relationship.FromAlias)
	assert.Equal(t, result.Entities[1].Aliases[0], relationship.ToAlias)
	assert.Equal(t, true, relationship.Properties["critical_path"])
}
