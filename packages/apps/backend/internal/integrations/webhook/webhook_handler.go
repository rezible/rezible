package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
)

// maxDeliveryBytes bounds a delivery body while it is read.
const maxDeliveryBytes = 64 << 10

// webhookHandler receives deliveries at /{token}. The token authenticates the delivery and identifies the
// installation; it is a secret and is never logged.
type webhookHandler struct {
	logger        *slog.Logger
	clock         rez.Clock
	events        rez.ProviderEventPipelineService
	installations rez.IntegrationInstallationLookup
}

func newWebhookHandler(logger *slog.Logger, clock rez.Clock, events rez.ProviderEventPipelineService, installations rez.IntegrationInstallationLookup) http.Handler {
	h := &webhookHandler{
		logger:        logger,
		clock:         clock,
		events:        events,
		installations: installations,
	}
	r := chi.NewRouter()
	r.Post("/{token}", h.handleDelivery)
	return r
}

// deliveryOutcome is logged once per delivery.
type deliveryOutcome struct {
	installationID uuid.UUID
	status         int
	queued         int
	// rejection is the field error returned to the caller with a 400.
	rejection *fieldError
	err       error
}

func (h *webhookHandler) handleDelivery(w http.ResponseWriter, r *http.Request) {
	outcome := h.receiveDelivery(w, r)
	h.logOutcome(r.Context(), outcome)
	if outcome.status == http.StatusNoContent {
		w.WriteHeader(outcome.status)
		return
	}
	if outcome.rejection != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(outcome.status)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": outcome.rejection.Error()})
		return
	}
	http.Error(w, http.StatusText(outcome.status), outcome.status)
}

func (h *webhookHandler) receiveDelivery(w http.ResponseWriter, r *http.Request) deliveryOutcome {
	tokenHash := integrations.WebhookTokenHash(chi.URLParam(r, "token"))
	lookup := in.And(in.Name(integrationName), in.WebhookTokenHash(tokenHash))
	intg, lookupErr := h.installations.LookupInstallation(execution.NewSystemContext(r.Context()), lookup)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return deliveryOutcome{status: http.StatusNotFound}
		}
		return deliveryOutcome{status: http.StatusServiceUnavailable, err: lookupErr}
	}
	outcome := deliveryOutcome{installationID: intg.ID}
	ctx := execution.NewTenantContext(r.Context(), intg.TenantID)

	body, readErr := io.ReadAll(http.MaxBytesReader(w, r.Body, maxDeliveryBytes))
	if readErr != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(readErr, &tooLarge) {
			outcome.status = http.StatusRequestEntityTooLarge
		} else {
			outcome.status = http.StatusBadRequest
		}
		outcome.err = readErr
		return outcome
	}

	config, configErr := decodeInstallationConfig(intg.InstallationConfig)
	if configErr != nil {
		outcome.status = http.StatusServiceUnavailable
		outcome.err = configErr
		return outcome
	}
	event, mapErr := presets[config.Preset].MapDelivery(intg.ID, body, h.clock.Now())
	if mapErr != nil {
		var rejection *fieldError
		if errors.As(mapErr, &rejection) {
			outcome.status = http.StatusBadRequest
			outcome.rejection = rejection
		} else {
			outcome.status = http.StatusServiceUnavailable
		}
		outcome.err = mapErr
		return outcome
	}

	if ingestErr := h.events.IngestMany(ctx, []rez.ProviderEvent{*event}); ingestErr != nil {
		outcome.status = http.StatusServiceUnavailable
		outcome.err = ingestErr
		return outcome
	}
	outcome.status = http.StatusNoContent
	outcome.queued = 1
	return outcome
}

func (h *webhookHandler) logOutcome(ctx context.Context, outcome deliveryOutcome) {
	level := slog.LevelInfo
	if outcome.status >= http.StatusInternalServerError {
		level = slog.LevelError
	} else if outcome.status >= http.StatusBadRequest {
		level = slog.LevelWarn
	}
	attrs := []slog.Attr{
		slog.Int("status", outcome.status),
		slog.Int("queued_events", outcome.queued),
	}
	if outcome.installationID != uuid.Nil {
		attrs = append(attrs, slog.String("integration_id", outcome.installationID.String()))
	}
	if outcome.err != nil {
		attrs = append(attrs, slog.String("error", outcome.err.Error()))
	}
	h.logger.LogAttrs(ctx, level, "webhook delivery", attrs...)
}
