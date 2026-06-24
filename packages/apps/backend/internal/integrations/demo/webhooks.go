package demoprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	rez "github.com/rezible/rezible"
)

type webhookHandler struct {
	provEvents rez.ProviderEventPipelineService
}

func newWebhookHandler(provEvents rez.ProviderEventPipelineService) http.Handler {
	return &webhookHandler{provEvents: provEvents}
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	eventType := r.Header.Get("X-Rez-Demo-Event")
	if eventType == "" {
		http.Error(w, "missing X-Rez-Demo-Event header", http.StatusBadRequest)
		return
	}

	var handlerErr error
	if eventType == webhookEventTypeAlert {
		handlerErr = h.handleDemoAlertEvent(r.Context(), body)
	}
	if handlerErr != nil {
		http.Error(w, handlerErr.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

const (
	webhookEventTypeAlert = "alert"
)

type alertEventPayload struct {
	EventId string `json:"eventId"`
	AlertId string `json:"id"`
	Name    string `json:"name"`
}

func (p alertEventPayload) makeProviderEvent(raw []byte) rez.ProviderEvent {
	return rez.ProviderEvent{
		Provider:           providerName,
		ProviderSource:     fmt.Sprintf("%s:webhook", integrationName),
		ProviderEventRef:   p.EventId,
		ProviderSubjectRef: p.AlertId,
		Payload:            raw,
	}
}

func (h *webhookHandler) handleDemoAlertEvent(ctx context.Context, body []byte) error {
	var payload alertEventPayload
	if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
		return fmt.Errorf("failed to unmarshal alert event payload: %w", jsonErr)
	}

	pe := payload.makeProviderEvent(body)
	if ingestErr := h.provEvents.Ingest(ctx, pe); ingestErr != nil {
		return fmt.Errorf("ingest error: %w", ingestErr)
	}
	return nil
}
