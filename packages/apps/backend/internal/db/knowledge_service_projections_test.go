package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/internal/projections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectRepositoryObservedMapsToRepositoryFactEvidence(t *testing.T) {
	ev := &ent.NormalizedEvent{
		ID:             uuid.New(),
		Provider:       "github",
		ProviderSource: "repositories",
		SubjectRef:     "myorg/api",
		Kind:           ne.KindRepositoryObserved,
		Attributes: projections.RepositoryObservedAttributes{
			DisplayName: "myorg/api",
			URL:         "https://github.com/myorg/api",
		}.Encode(),
	}
	proj := newKnowledgeEntityEventProjector(ev, nil)

	result := proj.projectRepositoryObserved(projections.RepositoryObserved{
		Event: ev,
		Attributes: projections.RepositoryObservedAttributes{
			DisplayName: "myorg/api",
			URL:         "https://github.com/myorg/api",
		},
	})

	require.Len(t, result.Entities, 1)
	assert.Empty(t, result.Relationships)

	entity := result.Entities[0]
	assert.Equal(t, knowledgeKindCodeRepository, entity.Kind)
	assert.Equal(t, assertionCodeRepositoryExists, entity.AssertionKind)
	assert.Equal(t, "myorg/api", entity.DisplayName)
	require.Len(t, entity.Aliases, 1)
	assert.Equal(t, EntityAliasRef{
		Provider:           "github",
		ProviderSource:     "repositories",
		ProviderSubjectRef: "myorg/api",
	}, entity.Aliases[0])
}

func TestProjectChangeEventObservedMapsChangeRepositoryRelationshipEvidence(t *testing.T) {
	ev := &ent.NormalizedEvent{
		ID:             uuid.New(),
		Provider:       "github",
		ProviderSource: "push",
		SubjectRef:     "github:myorg/api:abc123",
		Kind:           ne.KindChangeEventObserved,
		Attributes: projections.ChangeEventObservedAttributes{
			RepositoryExternalRef: "myorg/api",
			DisplayName:           "refs/heads/main",
		}.Encode(),
	}
	proj := newKnowledgeEntityEventProjector(ev, nil)

	result := proj.projectCodeChangeEventObserved(projections.ChangeEventObserved{
		Event: ev,
		Attributes: projections.ChangeEventObservedAttributes{
			RepositoryExternalRef: "myorg/api",
			DisplayName:           "refs/heads/main",
		},
	})

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
	proj := newKnowledgeEntityEventProjector(&ent.NormalizedEvent{
		OccurredAt: occurredAt,
		ReceivedAt: receivedAt,
	}, nil)

	assert.Equal(t, occurredAt, proj.observedAt())
}
