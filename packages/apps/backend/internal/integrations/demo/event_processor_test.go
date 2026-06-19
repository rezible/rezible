package demoprovider

import (
	"testing"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/projections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type convertablePayload interface {
	toEvent() (*rez.ProviderEvent, error)
}

func processPayload(t *testing.T, ev convertablePayload) ent.NormalizedEvents {
	provEvent, eventErr := ev.toEvent()
	require.NoError(t, eventErr)
	require.NotNil(t, provEvent)
	events, procErr := (&eventProcessor{event: provEvent}).process()
	require.NoError(t, procErr)
	return events
}

func TestProcessAlertObservedEvent(t *testing.T) {
	payload := alertObservedPayload{
		Title:           "Search API response time high",
		Description:     "p95 latency is above threshold.",
		Definition:      "avg(last_5m):p95:search.api.response_time > 2000",
		OccurredAt:      time.Date(2026, 5, 12, 9, 15, 0, 0, time.UTC),
		RelatedEntities: []projections.RelatedEntityRef{relatedComponent("search_api", "service", "Search API")},
	}
	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.Equal(t, ne.ActivityKindObserved, ev.ActivityKind)
	assert.True(t, projections.SubjectKindAlert.Matches(ev))
	assert.Equal(t, payload.getSubjectRef(), ev.ProviderSubjectRef)
	assert.Equal(t, payload.OccurredAt, ev.OccurredAt)

	decoded, decodeErr := projections.DecodeAlertEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, "Search API response time high", decoded.Attributes.Title)
	assert.Equal(t, "p95 latency is above threshold.", decoded.Attributes.Description)
	require.Len(t, decoded.Attributes.RelatedEntities, 1)
	assert.Equal(t, componentRef("search_api"), decoded.Attributes.RelatedEntities[0].ExternalRef)
}

func TestProcessIncidentObservedEvent(t *testing.T) {
	payload := incidentObservedPayload{
		ExternalID:  "foobar",
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		OccurredAt:  time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC),
	}
	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindIncident.Matches(ev))
	assert.Equal(t, payload.getSubjectRef(), ev.ProviderSubjectRef)
	assert.Equal(t, payload.OccurredAt, ev.OccurredAt)

	decoded, decodeErr := projections.DecodeIncidentEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.Title, decoded.Attributes.Title)
	assert.Equal(t, payload.SeverityRef, decoded.Attributes.SeverityRef)
}

func TestProcessCodeRepositoryObservedEvent(t *testing.T) {
	payload := demoCodeRepositoryEvents[0]

	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindCodeForge.Matches(ev))
	assert.Equal(t, payload.getSubjectRef(), ev.ProviderSubjectRef)

	decoded, decodeErr := projections.DecodeCodeForgeEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.FullName, decoded.Attributes.DisplayName)
	assert.Equal(t, payload.URL, decoded.Attributes.URL)
}

func TestProcessCodeChangeObservedEvent(t *testing.T) {
	payload := demoCodeChangeEvents[0]

	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindCodeChange.Matches(ev))
	assert.Equal(t, payload.getSubjectRef(), ev.ProviderSubjectRef)

	decoded, decodeErr := projections.DecodeCodeChangeEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.RepositoryExternalRef, decoded.Attributes.RepositoryExternalRef)
	assert.Equal(t, payload.Title, decoded.Attributes.DisplayName)
	require.Len(t, decoded.Attributes.RelatedEntities, 2)
}

func TestProcessChatMessageObservedEvent(t *testing.T) {
	payload := demoChatMessageEvents[0]

	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindChatMessage.Matches(ev))
	assert.Equal(t, ne.ActivityKindReceived, ev.ActivityKind)

	decoded, decodeErr := projections.DecodeChatMessageEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.Body, decoded.Attributes.Body)
	assert.Equal(t, payload.ConversationExternalRef, decoded.Attributes.ConversationExternalRef)
	require.Len(t, decoded.Attributes.RelatedEntities, 3)
}

func TestProcessPlaybookObservedEvent(t *testing.T) {
	payload := demoPlaybookEvents[0]

	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindPlaybook.Matches(ev))

	decoded, decodeErr := projections.DecodePlaybookEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.Title, decoded.Attributes.Title)
	assert.Contains(t, decoded.Attributes.Content, "Search API p95 latency")
	require.Len(t, decoded.Attributes.RelatedAlerts, 2)
}

func TestProcessIncidentImpactObservedEvent(t *testing.T) {
	payload := demoIncidentImpactEvents[0]

	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindIncidentImpact.Matches(ev))

	decoded, decodeErr := projections.DecodeIncidentImpactEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.IncidentExternalRef, decoded.Attributes.IncidentExternalRef)
	assert.Equal(t, payload.EntityExternalRef, decoded.Attributes.EntityExternalRef)
}

func TestProcessTopologyComponentObservedEvent(t *testing.T) {
	ownerTeam := "commerce_team"
	payload := &topologyComponentObservedPayload{
		ExternalRef: "demo:component:search_api",
		Kind:        "service",
		DisplayName: "Search API",
		Description: "Product search query API.",
		Properties: map[string]any{
			"criticality": "high",
			"owner_team":  ownerTeam,
		},
	}

	events := processPayload(t, payload)
	require.Len(t, events, 1)

	ev := events[0]
	assert.True(t, projections.SubjectKindSystemComponent.Matches(ev))
	assert.Equal(t, payload.getSubjectRef(), ev.ProviderSubjectRef)

	decoded, decodeErr := projections.DecodeSystemComponentEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.DisplayName, decoded.Attributes.DisplayName)
	assert.Equal(t, payload.Kind, decoded.Attributes.Kind)
	assert.Equal(t, ownerTeam, decoded.Attributes.Properties["owner_team"])
}

func TestProcessTopologyRelationshipObservedEvent(t *testing.T) {
	payload := &topologyRelationshipObservedPayload{
		ExternalRef:       "demo:relationship:checkout_service:calls:search_api",
		Kind:              "calls",
		DisplayName:       "Checkout Listener calls Search API",
		SourceExternalRef: "demo:component:checkout_service",
		SourceKind:        "service",
		SourceDisplayName: "Checkout Listener",
		TargetExternalRef: "demo:component:search_api",
		TargetKind:        "service",
		TargetDisplayName: "Search API",
		Properties: map[string]any{
			"critical_path": true,
		},
	}

	events := processPayload(t, payload)
	require.Len(t, events, 1)
	ev := events[0]
	assert.True(t, projections.SubjectKindSystemRelationship.Matches(ev))
	assert.Equal(t, payload.getSubjectRef(), ev.ProviderSubjectRef)

	decoded, decodeErr := projections.DecodeSystemRelationshipEvent(ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.Kind, decoded.Attributes.Kind)
	assert.Equal(t, payload.SourceExternalRef, decoded.Attributes.SourceExternalRef)
	assert.Equal(t, payload.TargetExternalRef, decoded.Attributes.TargetExternalRef)
	assert.Equal(t, true, decoded.Attributes.Properties["critical_path"])
}
