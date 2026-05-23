package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func (i *integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	return newEventQuerier(&ConfiguredIntegration{intg: intg}), nil
}

type eventQuerier struct {
	ci *ConfiguredIntegration
}

func newEventQuerier(ci *ConfiguredIntegration) *eventQuerier {
	return &eventQuerier{ci: ci}
}

func (q *eventQuerier) Provider() string {
	return integrationName
}

func (q *eventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if incidentsCursor, shouldQuery := req.GetSourceCursor(sourceIncidents); shouldQuery {
			for ev, evErr := range q.pullIncidentEvents(ctx, incidentsCursor) {
				if !yield(ev, evErr) {
					return
				}
			}
		}
		if alertsCursor, shouldQuery := req.GetSourceCursor(sourceAlerts); shouldQuery {
			for ev, evErr := range q.pullAlertEvents(ctx, alertsCursor) {
				if !yield(ev, evErr) {
					return
				}
			}
		}
		if topologyCursor, shouldQuery := req.GetSourceCursor(sourceTopology); shouldQuery {
			for ev, evErr := range q.pullTopologyEvents(ctx, topologyCursor) {
				if !yield(ev, evErr) {
					return
				}
			}
		}
	}
}

var fakeIncidentEvents = []incidentObservedPayload{
	{
		ExternalID:    "checkout-search-timeouts",
		Title:         "Checkout search lookups timing out",
		Summary:       "Checkout requests that need product search enrichment are timing out for a subset of customers.",
		SeverityRef:   "SEV-1",
		TypeRef:       "Customer Impact",
		OccurredAt:    time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC),
		ObservationID: "checkout-search-timeouts-observed",
	},
	{
		ExternalID:    "catalog-search-stale-results",
		Title:         "Catalog search returning stale results",
		Summary:       "The catalog search index failed to refresh after the nightly product import.",
		SeverityRef:   "SEV-2",
		TypeRef:       "Data Freshness",
		OccurredAt:    time.Date(2026, 5, 13, 2, 30, 0, 0, time.UTC),
		ObservationID: "catalog-search-stale-results-observed",
	},
	{
		ExternalID:    "search-admin-dashboard-degraded",
		Title:         "Search admin dashboard degraded",
		Summary:       "Internal teams are seeing slow loads and intermittent errors in search administration views.",
		SeverityRef:   "SEV-3",
		TypeRef:       "Internal Tooling",
		OccurredAt:    time.Date(2026, 5, 14, 5, 0, 0, 0, time.UTC),
		ObservationID: "search-admin-dashboard-degraded-observed",
	},
}

type incidentObservedPayload struct {
	ExternalID    string    `json:"external_id"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary,omitempty"`
	SeverityRef   string    `json:"severity_ref"`
	TypeRef       string    `json:"type_ref"`
	OccurredAt    time.Time `json:"occurred_at"`
	ObservationID string    `json:"observation_id"`
}

func (q *eventQuerier) pullIncidentEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		for _, payload := range fakeIncidentEvents {
			if cursor != "" && payload.ObservationID <= cursor {
				continue
			}
			payloadBytes, jsonErr := json.Marshal(payload)
			if jsonErr != nil {
				yield(nil, fmt.Errorf("json marshal incident: %w", jsonErr))
				return
			}
			res := &rez.ProviderEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:         integrationName,
					ProviderSource:   sourceIncidents,
					ProviderEventRef: fmt.Sprintf("fake:%s:%s", sourceIncidents, payload.ObservationID),
					SubjectRef:       fmt.Sprintf("fake:incident:%s", payload.ExternalID),
					ReceivedAt:       payload.OccurredAt,
					Payload:          payloadBytes,
					ContentType:      "application/json",
				},
				SourceCursorAfter: new(payload.ObservationID),
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}

type alertObservedPayload struct {
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Definition  string    `json:"definition,omitempty"`
	OccurredAt  time.Time `json:"occurred_at"`
	InstanceRef string    `json:"instance_ref"`
}

var fakeAlertEvents = []alertObservedPayload{
	{
		ExternalID:  "search-api-latency",
		Title:       "Search API response time high",
		Description: "p95 latency for the search API is above 2 seconds.",
		Definition:  "avg(last_5m):p95:search.api.response_time > 2000",
		OccurredAt:  time.Date(2026, 5, 12, 9, 15, 0, 0, time.UTC),
		InstanceRef: "search-api-latency-20260512T091500Z",
	},
	{
		ExternalID:  "elasticsearch-cpu-critical",
		Title:       "Elasticsearch cluster CPU critical",
		Description: "Primary search cluster CPU is above 95 percent.",
		Definition:  "avg(last_5m):avg:elasticsearch.cpu.utilization > 95",
		OccurredAt:  time.Date(2026, 5, 12, 9, 28, 0, 0, time.UTC),
		InstanceRef: "elasticsearch-cpu-critical-20260512T092800Z",
	},
	{
		ExternalID:  "search-index-build-failed",
		Title:       "Search index build failed",
		Description: "Nightly catalog search index rebuild exited with a failure.",
		Definition:  "sum(last_1h):search.indexer.failures > 0",
		OccurredAt:  time.Date(2026, 5, 13, 2, 10, 0, 0, time.UTC),
		InstanceRef: "search-index-build-failed-20260513T021000Z",
	},
	{
		ExternalID:  "redis-search-cache-down",
		Title:       "Redis search cache down",
		Description: "Search cache node is unreachable from application hosts.",
		Definition:  "min(last_5m):redis.search_cache.up < 1",
		OccurredAt:  time.Date(2026, 5, 13, 14, 5, 0, 0, time.UTC),
		InstanceRef: "redis-search-cache-down-20260513T140500Z",
	},
	{
		ExternalID:  "search-query-backlog",
		Title:       "Search query queue backing up",
		Description: "Search query processing queue depth is above 5000 messages.",
		Definition:  "avg(last_10m):search.query_queue.depth > 5000",
		OccurredAt:  time.Date(2026, 5, 14, 4, 45, 0, 0, time.UTC),
		InstanceRef: "search-query-backlog-20260514T044500Z",
	},
}

func (q *eventQuerier) pullAlertEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		for _, payload := range fakeAlertEvents {
			if cursor != "" && payload.InstanceRef <= cursor {
				continue
			}
			payloadBytes, jsonErr := json.Marshal(payload)
			if jsonErr != nil {
				yield(nil, fmt.Errorf("json marshal alert: %w", jsonErr))
				return
			}
			res := &rez.ProviderEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:         integrationName,
					ProviderSource:   sourceAlerts,
					ProviderEventRef: fmt.Sprintf("fake:%s:%s", sourceAlerts, payload.InstanceRef),
					SubjectRef:       fmt.Sprintf("fake:alert:%s", payload.ExternalID),
					ReceivedAt:       payload.OccurredAt,
					Payload:          payloadBytes,
					ContentType:      "application/json",
				},
				SourceCursorAfter: new(payload.InstanceRef),
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}
