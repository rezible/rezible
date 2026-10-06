package demoprovider

import (
	"fmt"
	"slices"
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

type teamObservedPayload struct {
	ResourceID    string    `json:"resource_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	ChatChannelID string    `json:"chat_channel_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (p teamObservedPayload) resourceRef() string {
	return "demo:team:" + p.ResourceID
}

var demoTeamEvents = []teamObservedPayload{
	{
		ResourceID:    "search-platform",
		Name:          "Search Platform",
		Slug:          "search-platform",
		ChatChannelID: "CSEARCH123",
		UpdatedAt:     demoObservedAt,
	},
}

type teamMembershipObservedPayload struct {
	TeamResourceID string `json:"team_resource_id"`
	UserExternalID string `json:"user_external_id"`
	Role           string `json:"role"`
}

func (p teamMembershipObservedPayload) resourceRef() string {
	return "demo:team-membership:" + p.TeamResourceID + ":" + p.UserExternalID
}

var demoTeamMembershipEvents = []teamMembershipObservedPayload{
	{
		TeamResourceID: "search-platform",
		UserExternalID: "ava-patel",
		Role:           "member",
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
		Provider:          ProviderName,
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
		RepositoryRef: "demo:code_repositories:search-api",
		Title:         "PR #1842 Tune search enrichment retry policy",
		MergedAt:      time.Date(2026, 5, 12, 8, 42, 0, 0, time.UTC),
		ImpactedEntities: []projections.EntityObservation{
			relatedComponent("search_api", kne.CategoryContainer, "service", "Search API"),
			relatedComponent("elasticsearch_catalog", kne.CategoryContainer, "search_cluster", "Elasticsearch Catalog"),
		},
	},
}

type incidentObservedPayload struct {
	ResourceID      string     `json:"resource_id"`
	Title           string     `json:"title"`
	Summary         string     `json:"summary,omitempty"`
	SeverityRef     string     `json:"severity_ref"`
	TypeRef         string     `json:"type_ref"`
	ResponseState   string     `json:"response_state"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	SourceUpdatedAt time.Time  `json:"source_updated_at"`
	SourceURL       string     `json:"source_url"`
	OccurredAt      time.Time  `json:"occurred_at"`
	ObservationID   string     `json:"observation_id"`
}

func (p incidentObservedPayload) resourceRef() string {
	return "demo:incident:" + p.ResourceID
}

func (p incidentObservedPayload) eventRef() string {
	return "demo:incidents:" + p.ObservationID
}

func (p incidentObservedPayload) cursorAfter() string {
	return fmt.Sprintf("z:%s\x1f%s\x1f%s", p.SourceUpdatedAt.UTC().Format(time.RFC3339Nano), p.resourceRef(), p.ObservationID)
}

var demoIncidentEvents = []incidentObservedPayload{
	{
		ResourceID:      "catalog-search-stale-results",
		Title:           "Catalog search returning stale results",
		Summary:         "The catalog search index failed to refresh after the nightly product import.",
		SeverityRef:     "SEV-2",
		TypeRef:         "Data Freshness",
		ResponseState:   "resolved",
		SourceUpdatedAt: time.Date(2026, 4, 18, 3, 5, 0, 0, time.UTC),
		SourceURL:       "https://status.demo.example/incidents/catalog-search-stale-results",
		OccurredAt:      time.Date(2026, 4, 18, 2, 30, 0, 0, time.UTC),
		ObservationID:   "catalog-search-stale-results-resolved-time-unknown",
	},
	{
		ResourceID:      "checkout-search-timeouts",
		Title:           "Checkout search lookups timing out",
		Summary:         "Checkout requests that need product search enrichment are timing out for a subset of customers.",
		SeverityRef:     "SEV-1",
		TypeRef:         "Customer Impact",
		ResponseState:   "resolved",
		ResolvedAt:      demoTime(time.Date(2026, 5, 12, 10, 5, 0, 0, time.UTC)),
		SourceUpdatedAt: time.Date(2026, 5, 12, 10, 5, 0, 0, time.UTC),
		SourceURL:       "https://status.demo.example/incidents/checkout-search-timeouts",
		OccurredAt:      time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC),
		ObservationID:   "checkout-search-timeouts-resolved",
	},
	{
		ResourceID:      "checkout-search-timeouts",
		Title:           "Checkout search lookups timing out",
		Summary:         "Checkout requests that need product search enrichment are timing out for a subset of customers.",
		SeverityRef:     "SEV-1",
		TypeRef:         "Customer Impact",
		ResponseState:   "resolved",
		ResolvedAt:      demoTime(time.Date(2026, 5, 12, 10, 5, 0, 0, time.UTC)),
		SourceUpdatedAt: time.Date(2026, 5, 12, 10, 5, 0, 0, time.UTC),
		SourceURL:       "https://status.demo.example/incidents/checkout-search-timeouts",
		OccurredAt:      time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC),
		ObservationID:   "checkout-search-timeouts-resolved-repeat",
	},
	{
		ResourceID:      "search-admin-dashboard-degraded",
		Title:           "Search admin dashboard degraded",
		Summary:         "Internal teams are seeing slow loads and intermittent errors in search administration views.",
		SeverityRef:     "SEV-3",
		TypeRef:         "Internal Tooling",
		ResponseState:   "resolved",
		SourceUpdatedAt: time.Date(2026, 5, 14, 5, 0, 0, 0, time.UTC),
		SourceURL:       "https://status.demo.example/incidents/search-admin-dashboard-degraded",
		OccurredAt:      time.Date(2026, 5, 14, 5, 0, 0, 0, time.UTC),
		ObservationID:   "search-admin-dashboard-resolved-first-observation",
	},
}

func demoTime(value time.Time) *time.Time {
	return &value
}

type alertObservedPayload struct {
	DefinitionRef            string                          `json:"definition_ref"`
	Title                    string                          `json:"title"`
	Description              string                          `json:"description,omitempty"`
	Definition               string                          `json:"definition,omitempty"`
	State                    string                          `json:"state"`
	InstanceID               string                          `json:"instance_id,omitempty"`
	Labels                   map[string]string               `json:"labels,omitempty"`
	Summary                  string                          `json:"summary,omitempty"`
	IdentityGroupLabels      []string                        `json:"identity_group_labels,omitempty"`
	Severity                 string                          `json:"severity,omitempty"`
	StartedAt                time.Time                       `json:"started_at"`
	EndedAt                  *time.Time                      `json:"ended_at,omitempty"`
	ResolutionTimeoutSeconds *int                            `json:"resolution_timeout_seconds,omitempty"`
	OccurredAt               time.Time                       `json:"occurred_at"`
	NotificationRef          string                          `json:"notification_ref"`
	ObservedEntities         []projections.EntityObservation `json:"observed_entities,omitempty"`
}

// resourceRef identifies the definition, so all of a definition's notifications share it.
func (p alertObservedPayload) resourceRef() string {
	return "demo:alert:" + p.DefinitionRef
}

func (p alertObservedPayload) eventRef() string {
	return "demo:alerts:" + p.NotificationRef
}

func (p alertObservedPayload) cursorAfter() string {
	return p.eventRef()
}

// receivedAt is when the notification was delivered: as it occurred.
func (p alertObservedPayload) receivedAt() time.Time {
	return p.OccurredAt
}

// resolvedDemoAlert is the resolved notification ending a firing demo alert's window.
func resolvedDemoAlert(firing alertObservedPayload, resolvedAt time.Time) alertObservedPayload {
	resolved := firing
	resolved.State = projections.AlertStateResolved
	resolved.OccurredAt = resolvedAt
	resolved.EndedAt = &resolvedAt
	resolved.NotificationRef = resolvedAt.Format("20060102T150405Z") + "-" + firing.DefinitionRef + "-resolved"
	return resolved
}

var (
	demoSearchLatencyFiring = alertObservedPayload{
		DefinitionRef:   "search-api-latency",
		Title:           "Search API response time high",
		Description:     "p95 latency for the search API is above 2 seconds.",
		Definition:      "avg(last_5m):p95:search.api.response_time > 2000",
		State:           projections.AlertStateFiring,
		Labels:          map[string]string{"service": "search-api", "env": "production"},
		Summary:         "Search API p95 latency is 2.8 seconds.",
		Severity:        "warning",
		StartedAt:       time.Date(2026, 5, 12, 9, 15, 0, 0, time.UTC),
		OccurredAt:      time.Date(2026, 5, 12, 9, 15, 0, 0, time.UTC),
		NotificationRef: "20260512T091500Z-search-api-latency-firing",
		ObservedEntities: []projections.EntityObservation{
			relatedComponent("search_api", kne.CategoryContainer, "service", "Search API"),
			relatedComponent("checkout_service", kne.CategoryContainer, "service", "Checkout Listener"),
		},
	}
	demoElasticsearchCPUFiring = alertObservedPayload{
		DefinitionRef:   "elasticsearch-cpu-critical",
		Title:           "Elasticsearch cluster CPU critical",
		Description:     "Primary search cluster CPU is above 95 percent.",
		Definition:      "avg(last_5m):avg:elasticsearch.cpu.utilization > 95",
		State:           projections.AlertStateFiring,
		Labels:          map[string]string{"cluster": "elasticsearch-catalog", "env": "production"},
		Summary:         "Catalog cluster CPU is at 97 percent.",
		Severity:        "critical",
		StartedAt:       time.Date(2026, 5, 12, 9, 28, 0, 0, time.UTC),
		OccurredAt:      time.Date(2026, 5, 12, 9, 28, 0, 0, time.UTC),
		NotificationRef: "20260512T092800Z-elasticsearch-cpu-critical-firing",
		ObservedEntities: []projections.EntityObservation{
			relatedComponent("elasticsearch_catalog", kne.CategoryContainer, "search_cluster", "Elasticsearch Catalog"),
			relatedComponent("search_api", kne.CategoryContainer, "service", "Search API"),
		},
	}
	demoIndexBuildFiring = alertObservedPayload{
		DefinitionRef:   "search-index-build-failed",
		Title:           "Search index build failed",
		Description:     "Nightly catalog search index rebuild exited with a failure.",
		Definition:      "sum(last_1h):search.indexer.failures > 0",
		State:           projections.AlertStateFiring,
		Labels:          map[string]string{"job": "catalog-indexer", "env": "production"},
		Severity:        "warning",
		StartedAt:       time.Date(2026, 5, 13, 2, 10, 0, 0, time.UTC),
		OccurredAt:      time.Date(2026, 5, 13, 2, 10, 0, 0, time.UTC),
		NotificationRef: "20260513T021000Z-search-index-build-failed-firing",
	}
	demoRedisCacheFiring = alertObservedPayload{
		DefinitionRef:   "redis-search-cache-down",
		Title:           "Redis search cache down",
		Description:     "Search cache node is unreachable from application hosts.",
		Definition:      "min(last_5m):redis.search_cache.up < 1",
		State:           projections.AlertStateFiring,
		Labels:          map[string]string{"instance": "redis-search-cache-1", "env": "production"},
		Severity:        "critical",
		StartedAt:       time.Date(2026, 5, 13, 14, 5, 0, 0, time.UTC),
		OccurredAt:      time.Date(2026, 5, 13, 14, 5, 0, 0, time.UTC),
		NotificationRef: "20260513T140500Z-redis-search-cache-down-firing",
	}
	demoQueryBacklogFiring = alertObservedPayload{
		DefinitionRef:   "search-query-backlog",
		Title:           "Search query queue backing up",
		Description:     "Search query processing queue depth is above 5000 messages.",
		Definition:      "avg(last_10m):search.query_queue.depth > 5000",
		State:           projections.AlertStateFiring,
		Labels:          map[string]string{"queue": "search-queries", "env": "production"},
		Severity:        "warning",
		StartedAt:       time.Date(2026, 5, 14, 4, 45, 0, 0, time.UTC),
		OccurredAt:      time.Date(2026, 5, 14, 4, 45, 0, 0, time.UTC),
		NotificationRef: "20260514T044500Z-search-query-backlog-firing",
	}
)

// demoIncidentAlertsEnd is the fixtures' last notification of the demo incident's alerts.
var demoIncidentAlertsEnd = time.Date(2026, 5, 12, 9, 43, 0, 0, time.UTC)

// demoAlertEvents are ordered by notification ref, which is also their delivery order. The search latency
// alert of the demo incident never resolves, so its situation stays open.
var demoAlertEvents = []alertObservedPayload{
	demoSearchLatencyFiring,
	demoElasticsearchCPUFiring,
	resolvedDemoAlert(demoElasticsearchCPUFiring, demoIncidentAlertsEnd),
	demoIndexBuildFiring,
	resolvedDemoAlert(demoIndexBuildFiring, time.Date(2026, 5, 13, 2, 55, 0, 0, time.UTC)),
	demoRedisCacheFiring,
	resolvedDemoAlert(demoRedisCacheFiring, time.Date(2026, 5, 13, 14, 13, 0, 0, time.UTC)),
	demoQueryBacklogFiring,
	resolvedDemoAlert(demoQueryBacklogFiring, time.Date(2026, 5, 14, 5, 15, 0, 0, time.UTC)),
}

// demoAlertEventsAt is demoAlertEvents with the demo incident's alerts moved so their last notification lands a
// minute before the integration was installed, keeping their situation live. The times depend only on the
// install time, so every pull delivers the same notifications; refs are unchanged, so sync cursors still apply.
func demoAlertEventsAt(installedAt time.Time) []alertObservedPayload {
	shift := installedAt.Add(-time.Minute).Sub(demoIncidentAlertsEnd)
	incidentAlerts := []string{demoSearchLatencyFiring.DefinitionRef, demoElasticsearchCPUFiring.DefinitionRef}
	events := make([]alertObservedPayload, 0, len(demoAlertEvents))
	for _, event := range demoAlertEvents {
		if slices.Contains(incidentAlerts, event.DefinitionRef) {
			event.StartedAt = event.StartedAt.Add(shift)
			event.OccurredAt = event.OccurredAt.Add(shift)
			if event.EndedAt != nil {
				event.EndedAt = new(event.EndedAt.Add(shift))
			}
		}
		events = append(events, event)
	}
	return events
}
