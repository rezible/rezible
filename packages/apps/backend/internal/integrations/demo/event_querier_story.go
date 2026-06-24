package demoprovider

import (
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/projections"
)

var demoObservedAt = time.Date(2026, 5, 12, 8, 0, 0, 0, time.UTC)

type userObservedPayload struct {
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	ChatID     string    `json:"chat_id"`
	Timezone   string    `json:"timezone"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (p userObservedPayload) getEventRef() string {
	return "demo:users:" + p.ExternalID
}

func (p userObservedPayload) getSubjectRef() string {
	return "demo:user:" + p.ExternalID
}

func (p userObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	return payloadToEvent(sourceUsers, p.getEventRef(), p.getSubjectRef(), p.UpdatedAt, p)
}

var demoUserEvents = []userObservedPayload{
	{
		ExternalID: "ava-patel",
		Name:       "Ava Patel",
		Email:      "ava.patel@rezible.example",
		ChatID:     "UAVA123",
		Timezone:   "Australia/Perth",
		UpdatedAt:  demoObservedAt,
	},
}

type codeRepositoryObservedPayload struct {
	ExternalID string    `json:"external_id"`
	FullName   string    `json:"full_name"`
	URL        string    `json:"url"`
	ObservedAt time.Time `json:"observed_at"`
}

func (p codeRepositoryObservedPayload) getEventRef() string {
	return "demo:code_repositories:" + p.ExternalID
}

func (p codeRepositoryObservedPayload) getSubjectRef() string {
	return p.FullName
}

func (p codeRepositoryObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	return payloadToEvent(sourceCodeRepos, p.getEventRef(), p.getSubjectRef(), p.ObservedAt, p)
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
	ExternalID            string                         `json:"external_id"`
	RepositoryExternalRef string                         `json:"repository_external_ref"`
	Title                 string                         `json:"title"`
	MergedAt              time.Time                      `json:"merged_at"`
	RelatedEntities       []projections.RelatedEntityRef `json:"related_entities,omitempty"`
}

func (p codeChangeObservedPayload) getEventRef() string {
	return "demo:code_changes:" + p.ExternalID
}

func (p codeChangeObservedPayload) getSubjectRef() string {
	return "demo:code_change:" + p.ExternalID
}

func (p codeChangeObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	return payloadToEvent(sourceCodeChanges, p.getEventRef(), p.getSubjectRef(), p.MergedAt, p)
}

func relatedComponent(id, kind, displayName string) projections.RelatedEntityRef {
	return projections.RelatedEntityRef{
		ExternalRef: componentRef(id),
		Kind:        kind,
		DisplayName: displayName,
	}
}

var demoCodeChangeEvents = []codeChangeObservedPayload{
	{
		ExternalID:            "pr-1842",
		RepositoryExternalRef: "rezible-commerce/search-api",
		Title:                 "PR #1842 Tune search enrichment retry policy",
		MergedAt:              time.Date(2026, 5, 12, 8, 42, 0, 0, time.UTC),
		RelatedEntities: []projections.RelatedEntityRef{
			relatedComponent("search_api", "service", "Search API"),
			relatedComponent("elasticsearch_catalog", "search_cluster", "Elasticsearch Catalog"),
		},
	},
}

type chatMessageObservedPayload struct {
	ExternalID              string                         `json:"external_id"`
	ConversationExternalRef string                         `json:"conversation_external_ref"`
	Body                    string                         `json:"body"`
	SenderExternalRef       string                         `json:"sender_external_ref"`
	ThreadExternalRef       string                         `json:"thread_external_ref,omitempty"`
	OccurredAt              time.Time                      `json:"occurred_at"`
	RelatedEntities         []projections.RelatedEntityRef `json:"related_entities,omitempty"`
}

func (p chatMessageObservedPayload) getEventRef() string {
	return "demo:chat_messages:" + p.ExternalID
}

func (p chatMessageObservedPayload) getSubjectRef() string {
	return "demo:chat_message:" + p.ExternalID
}

func (p chatMessageObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	return payloadToEvent(sourceChatMessages, p.getEventRef(), p.getSubjectRef(), p.OccurredAt, p)
}

var demoChatMessageEvents = []chatMessageObservedPayload{
	{
		ExternalID:              "checkout-thread-001",
		ConversationExternalRef: "#inc-checkout-search-timeouts",
		Body:                    "p95 checkout latency moved right after PR #1842. Search API is retrying Elasticsearch more aggressively than expected.",
		SenderExternalRef:       "demo:user:ava-patel",
		ThreadExternalRef:       "checkout-search-timeouts",
		OccurredAt:              time.Date(2026, 5, 12, 9, 38, 0, 0, time.UTC),
		RelatedEntities: []projections.RelatedEntityRef{
			relatedComponent("checkout_service", "service", "Checkout Listener"),
			relatedComponent("search_api", "service", "Search API"),
			relatedComponent("elasticsearch_catalog", "search_cluster", "Elasticsearch Catalog"),
		},
	},
}

type playbookObservedPayload struct {
	ExternalID               string    `json:"external_id"`
	Title                    string    `json:"title"`
	Content                  string    `json:"content"`
	UpdatedAt                time.Time `json:"updated_at"`
	RelatedAlertExternalRefs []string  `json:"related_alert_external_refs,omitempty"`
}

func (p playbookObservedPayload) getEventRef() string {
	return "demo:playbooks:" + p.ExternalID
}

func (p playbookObservedPayload) getSubjectRef() string {
	return "demo:playbook:" + p.ExternalID
}

func (p playbookObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	return payloadToEvent(sourcePlaybooks, p.getEventRef(), p.getSubjectRef(), p.UpdatedAt, p)
}

var demoPlaybookEvents = []playbookObservedPayload{
	{
		ExternalID: "checkout-search-latency",
		Title:      "Checkout search latency triage",
		Content:    "Check Search API p95 latency, Elasticsearch CPU, retry volume, and rollback PR #1842 if retry amplification continues.",
		UpdatedAt:  demoObservedAt,
		RelatedAlertExternalRefs: []string{
			"demo:alert:search-api-latency",
			"demo:alert:elasticsearch-cpu-critical",
		},
	},
}

type incidentImpactObservedPayload struct {
	ExternalID          string    `json:"external_id"`
	IncidentExternalRef string    `json:"incident_external_ref"`
	EntityExternalRef   string    `json:"entity_external_ref"`
	EntityKind          string    `json:"entity_kind"`
	EntityDisplayName   string    `json:"entity_display_name"`
	Source              string    `json:"source,omitempty"`
	Note                string    `json:"note,omitempty"`
	ObservedAt          time.Time `json:"observed_at"`
}

func (p incidentImpactObservedPayload) getEventRef() string {
	return "demo:incident_impacts:" + p.ExternalID
}

func (p incidentImpactObservedPayload) getSubjectRef() string {
	return "demo:incident_impact:" + p.ExternalID
}

func (p incidentImpactObservedPayload) toEvent() (*rez.ProviderEvent, error) {
	return payloadToEvent(sourceIncidentImpacts, p.getEventRef(), p.getSubjectRef(), p.ObservedAt, p)
}

var demoIncidentImpactEvents = []incidentImpactObservedPayload{
	{
		ExternalID:          "checkout-current-search-api",
		IncidentExternalRef: "demo:incident:checkout-search-timeouts",
		EntityExternalRef:   componentRef("search_api"),
		EntityKind:          "service",
		EntityDisplayName:   "Search API",
		Source:              "demo",
		Note:                "Checkout enrichment depends on Search API responses during payment initiation.",
		ObservedAt:          time.Date(2026, 5, 12, 9, 36, 0, 0, time.UTC),
	},
	{
		ExternalID:          "checkout-current-checkout-service",
		IncidentExternalRef: "demo:incident:checkout-search-timeouts",
		EntityExternalRef:   componentRef("checkout_service"),
		EntityKind:          "service",
		EntityDisplayName:   "Checkout Listener",
		Source:              "demo",
		Note:                "Customer checkout path is timing out while waiting for search enrichment.",
		ObservedAt:          time.Date(2026, 5, 12, 9, 36, 0, 0, time.UTC),
	},
	{
		ExternalID:          "checkout-prior-search-api",
		IncidentExternalRef: "demo:incident:catalog-search-stale-results",
		EntityExternalRef:   componentRef("search_api"),
		EntityKind:          "service",
		EntityDisplayName:   "Search API",
		Source:              "demo",
		Note:                "Previous search degradation involved the same query and index path.",
		ObservedAt:          time.Date(2026, 4, 18, 2, 35, 0, 0, time.UTC),
	},
}

type eventPayload interface {
	getEventRef() string
	getSubjectRef() string
	toEvent() (*rez.ProviderEvent, error)
}

func payloadToEvent(source, eventRef, subjectRef string, receivedAt time.Time, payload any) (*rez.ProviderEvent, error) {
	enc, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return &rez.ProviderEvent{
		Provider:           integrationName,
		ProviderSource:     source,
		ProviderEventRef:   eventRef,
		ProviderSubjectRef: subjectRef,
		ReceivedAt:         receivedAt,
		Payload:            enc,
		ContentType:        "application/json",
	}, nil
}

func pullPayloadEvents[P eventPayload](items []P, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		for _, payload := range items {
			cursorAfter := payload.getEventRef()
			if cursor != "" && cursorAfter <= cursor {
				continue
			}
			ev, evErr := payload.toEvent()
			if evErr != nil || ev == nil {
				yield(nil, fmt.Errorf("marshal demo event: %w", evErr))
				return
			}
			res := &rez.ProviderEventQueryResult{
				Event:             *ev,
				SourceCursorAfter: new(cursorAfter),
			}
			if !yield(res, nil) {
				return
			}
		}
	}
}

func (q *eventQuerier) pullUserEvents(cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return pullPayloadEvents(demoUserEvents, cursor)
}

func (q *eventQuerier) pullCodeRepositoryEvents(cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return pullPayloadEvents(demoCodeRepositoryEvents, cursor)
}

func (q *eventQuerier) pullCodeChangeEvents(cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return pullPayloadEvents(demoCodeChangeEvents, cursor)
}

func (q *eventQuerier) pullChatMessageEvents(cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return pullPayloadEvents(demoChatMessageEvents, cursor)
}

func (q *eventQuerier) pullPlaybookEvents(cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return pullPayloadEvents(demoPlaybookEvents, cursor)
}

func (q *eventQuerier) pullIncidentImpactEvents(cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return pullPayloadEvents(demoIncidentImpactEvents, cursor)
}
