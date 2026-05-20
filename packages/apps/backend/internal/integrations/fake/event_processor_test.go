package fakeprovider

import (
	"encoding/json"
	"testing"
	"time"

	rez "github.com/rezible/rezible"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/integrations/projections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessAlertObservedEvent(t *testing.T) {
	occurredAt := time.Date(2026, 5, 12, 9, 15, 0, 0, time.UTC)
	payload := alertObservedPayload{
		Title:       "Search API response time high",
		Description: "p95 latency is above threshold.",
		Definition:  "avg(last_5m):p95:search.api.response_time > 2000",
		OccurredAt:  occurredAt,
	}
	payloadBytes, jsonErr := json.Marshal(payload)
	require.NoError(t, jsonErr)
	prov := rez.ProviderEvent{
		Provider:         integrationName,
		ProviderSource:   sourceAlerts,
		ProviderEventRef: "fake:alerts:search-api-latency",
		SubjectRef:       "fake:alert:search-api-latency",
		ReceivedAt:       occurredAt,
		Payload:          payloadBytes,
	}
	events, procErr := (&eventProcessor{}).Process(t.Context(), prov)

	require.NoError(t, procErr)
	require.Len(t, events, 1)
	ev := events[0]
	assert.Equal(t, ne.KindAlertObserved, ev.Kind)
	assert.Equal(t, "alert", ev.SubjectKind)
	assert.Equal(t, "fake:alert:search-api-latency", ev.SubjectRef)
	assert.Equal(t, occurredAt, ev.OccurredAt)

	decoded, decodeErr := projections.DecodeEvent[projections.AlertObservedAttributes](ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, "Search API response time high", decoded.Attributes.Title)
	assert.Equal(t, "p95 latency is above threshold.", decoded.Attributes.Description)
}

func TestProcessIncidentObservedEvent(t *testing.T) {
	occurredAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)
	payload := incidentObservedPayload{
		Title:        "Checkout search lookups timing out",
		Summary:      "Checkout requests are timing out.",
		SeverityName: "SEV-1",
		SeverityRank: 1,
		TypeName:     "Customer Impact",
		OccurredAt:   occurredAt,
	}
	payloadBytes, jsonErr := json.Marshal(payload)
	require.NoError(t, jsonErr)

	events, err := (&eventProcessor{}).Process(t.Context(), rez.ProviderEvent{
		Provider:         integrationName,
		ProviderSource:   sourceIncidents,
		ProviderEventRef: "fake:incidents:checkout-search-timeouts",
		SubjectRef:       "fake:incident:checkout-search-timeouts",
		ReceivedAt:       occurredAt,
		Payload:          payloadBytes,
	})

	require.NoError(t, err)
	require.Len(t, events, 1)
	ev := events[0]
	assert.Equal(t, ne.KindIncidentObserved, ev.Kind)
	assert.Equal(t, "incident", ev.SubjectKind)
	assert.Equal(t, "fake:incident:checkout-search-timeouts", ev.SubjectRef)
	assert.Equal(t, occurredAt, ev.OccurredAt)

	decoded, decodeErr := projections.DecodeEvent[projections.IncidentObservedAttributes](ev)
	require.NoError(t, decodeErr)
	assert.Equal(t, payload.Title, decoded.Attributes.Title)
	assert.Equal(t, payload.SeverityName, decoded.Attributes.SeverityName)
	assert.Equal(t, payload.SeverityRank, decoded.Attributes.SeverityRank)
}
