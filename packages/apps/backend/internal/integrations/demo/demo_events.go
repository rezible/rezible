package demoprovider

import (
	"time"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/pkg/projections"
)

type demoEventPayload interface {
	resourceRef() string
	//toEvent() *ent.NormalizedEvent
}

var demoObservedAt = time.Date(2026, 5, 12, 8, 0, 0, 0, time.UTC)

type userObservedPayload struct {
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	ChatID     string    `json:"chat_id"`
	Timezone   string    `json:"timezone"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (p userObservedPayload) resourceRef() string {
	return "demo:user:" + p.ExternalID
}

var demoUserEvents = []userObservedPayload{
	{
		ExternalID: "ava-patel",
		Name:       "Ava Patel",
		Email:      "ava.patel@rezible.example",
		ChatID:     "UAVA123",
		Timezone:   "Australia/Sydney",
		UpdatedAt:  demoObservedAt,
	},
}

type codeRepositoryObservedPayload struct {
	ExternalID string    `json:"external_id"`
	FullName   string    `json:"full_name"`
	URL        string    `json:"url"`
	ObservedAt time.Time `json:"observed_at"`
}

func (p codeRepositoryObservedPayload) resourceRef() string {
	return "demo:code_repositories:" + p.ExternalID
}

var demoCodeRepositoryEvents = []codeRepositoryObservedPayload{
	{
		ExternalID: "search-api",
		FullName:   "rezible-commerce/search-api",
		URL:        "https://github.example/rezible-commerce/search-api",
		ObservedAt: demoObservedAt,
	},
}

type codeChangeObservedPayload struct {
	ExternalID       string                          `json:"external_id"`
	RepositoryRef    string                          `json:"repository_ref"`
	Title            string                          `json:"title"`
	MergedAt         time.Time                       `json:"merged_at"`
	ImpactedEntities []projections.EntityObservation `json:"impacted_entities,omitempty"`
}

func (p codeChangeObservedPayload) resourceRef() string {
	return "demo:code_change:" + p.ExternalID
}

func relatedComponent(id string, category kne.Category, kind string, displayName string) projections.EntityObservation {
	resourceRef := rez.ProviderResourceRef{
		Provider:          providerName,
		ProviderNamespace: integrationName,
		ResourceRef:       componentRef(id),
	}
	return projections.EntityObservation{
		Ref:         resourceRef,
		Category:    category,
		Kind:        kind,
		DisplayName: displayName,
	}
}

var demoCodeChangeEvents = []codeChangeObservedPayload{
	{
		ExternalID:    "pr-1842",
		RepositoryRef: "rezible-commerce/search-api",
		Title:         "PR #1842 Tune search enrichment retry policy",
		MergedAt:      time.Date(2026, 5, 12, 8, 42, 0, 0, time.UTC),
		ImpactedEntities: []projections.EntityObservation{
			relatedComponent("search_api", kne.CategoryContainer, "service", "Search API"),
			relatedComponent("elasticsearch_catalog", kne.CategoryContainer, "search_cluster", "Elasticsearch Catalog"),
		},
	},
}

type incidentObservedPayload struct {
	ResourceID    string    `json:"resource_id"`
	Title         string    `json:"title"`
	Summary       string    `json:"summary,omitempty"`
	SeverityRef   string    `json:"severity_ref"`
	TypeRef       string    `json:"type_ref"`
	OccurredAt    time.Time `json:"occurred_at"`
	ObservationID string    `json:"observation_id"`
}

func (p incidentObservedPayload) resourceRef() string {
	return "demo:incident:" + p.ResourceID
}

var demoIncidentEvents = []incidentObservedPayload{
	{
		ResourceID:    "checkout-search-timeouts",
		Title:         "Checkout search lookups timing out",
		Summary:       "Checkout requests that need product search enrichment are timing out for a subset of customers.",
		SeverityRef:   "SEV-1",
		TypeRef:       "Customer Impact",
		OccurredAt:    time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC),
		ObservationID: "checkout-search-timeouts-observed",
	},
	{
		ResourceID:    "catalog-search-stale-results",
		Title:         "Catalog search returning stale results",
		Summary:       "The catalog search index failed to refresh after the nightly product import.",
		SeverityRef:   "SEV-2",
		TypeRef:       "Data Freshness",
		OccurredAt:    time.Date(2026, 4, 18, 2, 30, 0, 0, time.UTC),
		ObservationID: "catalog-search-stale-results-observed",
	},
	{
		ResourceID:    "search-admin-dashboard-degraded",
		Title:         "Search admin dashboard degraded",
		Summary:       "Internal teams are seeing slow loads and intermittent errors in search administration views.",
		SeverityRef:   "SEV-3",
		TypeRef:       "Internal Tooling",
		OccurredAt:    time.Date(2026, 5, 14, 5, 0, 0, 0, time.UTC),
		ObservationID: "search-admin-dashboard-degraded-observed",
	},
}

type alertObservedPayload struct {
	DefinitionRef    string                          `json:"definition_ref"`
	Title            string                          `json:"title"`
	Description      string                          `json:"description,omitempty"`
	Definition       string                          `json:"definition,omitempty"`
	OccurredAt       time.Time                       `json:"occurred_at"`
	InstanceRef      string                          `json:"instance_ref"`
	ObservedEntities []projections.EntityObservation `json:"observed_entities,omitempty"`
}

func (p alertObservedPayload) resourceRef() string {
	return "demo:alert_instance:" + p.InstanceRef
}

var demoAlertEvents = []alertObservedPayload{
	{
		DefinitionRef: "search-api-latency",
		Title:         "Search API response time high",
		Description:   "p95 latency for the search API is above 2 seconds.",
		Definition:    "avg(last_5m):p95:search.api.response_time > 2000",
		OccurredAt:    time.Date(2026, 5, 12, 9, 15, 0, 0, time.UTC),
		InstanceRef:   "search-api-latency-20260512T091500Z",
		ObservedEntities: []projections.EntityObservation{
			relatedComponent("search_api", kne.CategoryContainer, "service", "Search API"),
			relatedComponent("checkout_service", kne.CategoryContainer, "service", "Checkout Listener"),
		},
	},
	{
		DefinitionRef: "elasticsearch-cpu-critical",
		Title:         "Elasticsearch cluster CPU critical",
		Description:   "Primary search cluster CPU is above 95 percent.",
		Definition:    "avg(last_5m):avg:elasticsearch.cpu.utilization > 95",
		OccurredAt:    time.Date(2026, 5, 12, 9, 28, 0, 0, time.UTC),
		InstanceRef:   "elasticsearch-cpu-critical-20260512T092800Z",
		ObservedEntities: []projections.EntityObservation{
			relatedComponent("elasticsearch_catalog", kne.CategoryContainer, "search_cluster", "Elasticsearch Catalog"),
			relatedComponent("search_api", kne.CategoryContainer, "service", "Search API"),
		},
	},
	{
		DefinitionRef: "search-index-build-failed",
		Title:         "Search index build failed",
		Description:   "Nightly catalog search index rebuild exited with a failure.",
		Definition:    "sum(last_1h):search.indexer.failures > 0",
		OccurredAt:    time.Date(2026, 5, 13, 2, 10, 0, 0, time.UTC),
		InstanceRef:   "search-index-build-failed-20260513T021000Z",
	},
	{
		DefinitionRef: "redis-search-cache-down",
		Title:         "Redis search cache down",
		Description:   "Search cache node is unreachable from application hosts.",
		Definition:    "min(last_5m):redis.search_cache.up < 1",
		OccurredAt:    time.Date(2026, 5, 13, 14, 5, 0, 0, time.UTC),
		InstanceRef:   "redis-search-cache-down-20260513T140500Z",
	},
	{
		DefinitionRef: "search-query-backlog",
		Title:         "Search query queue backing up",
		Description:   "Search query processing queue depth is above 5000 messages.",
		Definition:    "avg(last_10m):search.query_queue.depth > 5000",
		OccurredAt:    time.Date(2026, 5, 14, 4, 45, 0, 0, time.UTC),
		InstanceRef:   "search-query-backlog-20260514T044500Z",
	},
}
