package demoprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/projections"
)

type webhookHandler struct {
	router     http.Handler
	provEvents rez.ProviderEventPipelineService
}

func newWebhookHandler(provEvents rez.ProviderEventPipelineService) http.Handler {
	h := &webhookHandler{provEvents: provEvents}
	r := chi.NewRouter()
	r.Post("/", h.handlePost)
	return r
}

type webhookPayload struct {
	EventKind string `json:"kind"`
	TenantId  *int   `json:"tenant_id,omitempty"`
}

func (h *webhookHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var payload webhookPayload
	if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	tenantId := 1
	if payload.TenantId != nil {
		tenantId = *payload.TenantId
	}
	ctx := execution.NewTenantContext(r.Context(), tenantId)

	var handlerErr error
	if payload.EventKind == webhookEventKindAlert {
		handlerErr = h.handleDemoAlertEvent(ctx, body)
	}
	if handlerErr != nil {
		http.Error(w, handlerErr.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

const (
	webhookEventKindAlert = "alert"
)

func (h *webhookHandler) handleDemoAlertEvent(ctx context.Context, body []byte) error {
	now := time.Now().UTC()
	payload := alertObservedPayload{
		ExternalID:  "search-api-latency",
		Title:       "Search API response time high",
		Description: "p95 latency for the search API is above 2 seconds.",
		Definition:  "avg(last_5m):p95:search.api.response_time > 2000",
		OccurredAt:  now,
		InstanceRef: fmt.Sprintf("search-api-latency-%s", now.String()),
		RelatedEntities: []projections.RelatedEntityRef{
			relatedComponent("search_api", "service", "Search API"),
		},
	}
	ev, jsonErr := payload.toEvent()
	if jsonErr != nil || ev == nil {
		return fmt.Errorf("json marshal alert: %w", jsonErr)
	}
	if ingestErr := h.provEvents.Ingest(ctx, *ev); ingestErr != nil {
		return fmt.Errorf("ingest error: %w", ingestErr)
	}
	return nil
}
