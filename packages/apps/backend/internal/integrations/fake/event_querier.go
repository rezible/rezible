package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
)

func (i *Integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.IntegrationEventQuerier, error) {
	return newEventQuerier(&InstalledIntegration{intg: intg}), nil
}

type eventQuerier struct {
	ii *InstalledIntegration
}

func newEventQuerier(ci *InstalledIntegration) *eventQuerier {
	return &eventQuerier{ii: ci}
}

func (q *eventQuerier) Integration() *ent.Integration {
	return q.ii.intg
}

func (q *eventQuerier) PullEvents(ctx context.Context, cursors map[string]string) iter.Seq2[*rez.IntegrationEventQueryResult, error] {
	return func(yield func(*rez.IntegrationEventQueryResult, error) bool) {
		if cursor, shouldQuery := integrations.GetSourceQueryCursor(cursors, sourceIncidents); shouldQuery {
			for ev, evErr := range q.pullIncidentEvents(cursor) {
				if !yield(ev, evErr) {
					return
				}
			}
		}
		if cursor, shouldQuery := integrations.GetSourceQueryCursor(cursors, sourceAlerts); shouldQuery {
			for ev, evErr := range q.pullAlertEvents(cursor) {
				if !yield(ev, evErr) {
					return
				}
			}
		}
		if cursor, shouldQuery := integrations.GetSourceQueryCursor(cursors, sourceTopology); shouldQuery {
			for ev, evErr := range q.pullTopologyEvents(cursor) {
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

func (p incidentObservedPayload) getEventRef() string {
	return "fake:incidents:" + p.ObservationID
}

func (p incidentObservedPayload) getSubjectRef() string {
	return "fake:incidents:" + p.ExternalID
}

func (p incidentObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	enc, jsonErr := json.Marshal(p)
	if jsonErr != nil {
		return nil, jsonErr
	}
	prov := &rez.ProviderEvent{
		Provider:           integrationName,
		ProviderSource:     sourceIncidents,
		ProviderEventRef:   p.getEventRef(),
		ProviderSubjectRef: p.getSubjectRef(),
		ReceivedAt:         p.OccurredAt,
		Payload:            enc,
	}
	return prov, nil
}

func (q *eventQuerier) pullIncidentEvents(cursor string) iter.Seq2[*rez.IntegrationEventQueryResult, error] {
	return func(yield func(*rez.IntegrationEventQueryResult, error) bool) {
		for _, payload := range fakeIncidentEvents {
			if cursor != "" && payload.ObservationID <= cursor {
				continue
			}
			payloadBytes, jsonErr := json.Marshal(payload)
			if jsonErr != nil {
				yield(nil, fmt.Errorf("json marshal incident: %w", jsonErr))
				return
			}
			res := &rez.IntegrationEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:           integrationName,
					ProviderSource:     sourceIncidents,
					ProviderEventRef:   fmt.Sprintf("fake:%s:%s", sourceIncidents, payload.ObservationID),
					ProviderSubjectRef: fmt.Sprintf("fake:incident:%s", payload.ExternalID),
					ReceivedAt:         payload.OccurredAt,
					Payload:            payloadBytes,
					ContentType:        "application/json",
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

func (p alertObservedPayload) getEventRef() string {
	return "fake:alerts:" + p.InstanceRef
}

func (p alertObservedPayload) getSubjectRef() string {
	return "fake:alert:" + p.ExternalID
}

func (p alertObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	enc, jsonErr := json.Marshal(p)
	if jsonErr != nil {
		return nil, jsonErr
	}
	prov := &rez.ProviderEvent{
		Provider:           integrationName,
		ProviderSource:     sourceAlerts,
		ProviderEventRef:   p.getEventRef(),
		ProviderSubjectRef: p.getSubjectRef(),
		ReceivedAt:         p.OccurredAt,
		Payload:            enc,
	}
	return prov, nil
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

func (q *eventQuerier) pullAlertEvents(cursor string) iter.Seq2[*rez.IntegrationEventQueryResult, error] {
	return func(yield func(*rez.IntegrationEventQueryResult, error) bool) {
		for _, payload := range fakeAlertEvents {
			if cursor != "" && payload.InstanceRef <= cursor {
				continue
			}
			ev, jsonErr := payload.toEvent()
			if jsonErr != nil || ev == nil {
				yield(nil, fmt.Errorf("json marshal alert: %w", jsonErr))
				return
			}
			res := &rez.IntegrationEventQueryResult{
				Event:             *ev,
				SourceCursorAfter: new(payload.InstanceRef),
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}
