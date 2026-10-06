package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/pkg/execution"
)

type webhookHandler struct {
	secret        string
	provEvents    rez.ProviderEventPipelineService
	installations rez.IntegrationInstallationLookup
}

func newWebhookHandler(secret string, provEvents rez.ProviderEventPipelineService, installations rez.IntegrationInstallationLookup) http.Handler {
	return &webhookHandler{secret: secret, provEvents: provEvents, installations: installations}
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

	if !h.validateHMAC(body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "missing X-GitHub-Event header", http.StatusBadRequest)
		return
	}

	deliveryRef := r.Header.Get("X-GitHub-Delivery")

	providerNamespace := ""
	if eventType == sourcePushEvent {
		var payload pushEventPayload
		if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
			slog.Error("failed to unmarshal payload", "error", jsonErr.Error())
		}
		providerNamespace = fmt.Sprintf("%d", payload.Repository.Owner.ID)
	} else if eventType == sourcePullEvent {
		var payload pullRequestPayload
		if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
			slog.Error("failed to unmarshal payload", "error", jsonErr.Error())
		}
		providerNamespace = fmt.Sprintf("%d", payload.Repository.Owner.ID)
	} else {
		w.WriteHeader(http.StatusOK)
		return
	}
	if deliveryRef == "" || providerNamespace == "0" {
		http.Error(w, "missing github event identity", http.StatusBadRequest)
		return
	}

	pe := rez.ProviderEvent{
		Provider:            ProviderName,
		ProviderNamespace:   providerNamespace,
		ProviderEventSource: eventType,
		ProviderEventRef:    deliveryRef,
		Attributes:          body,
		ReceivedAt:          time.Now().UTC(),
	}

	code, ingestErr := h.ingestEvent(r.Context(), pe)
	if ingestErr != nil {
		http.Error(w, ingestErr.Error(), code)
		return
	}

	w.WriteHeader(code)
}

func (h *webhookHandler) ingestEvent(ctx context.Context, pe rez.ProviderEvent) (int, error) {
	lookupInstallation := in.And(in.Name(integrationName), in.ProviderInstallationRef(pe.ProviderNamespace))
	intg, lookupErr := h.installations.LookupInstallation(execution.NewSystemContext(ctx), lookupInstallation)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return http.StatusNotFound, fmt.Errorf("no installation for github account")
		}
		slog.ErrorContext(ctx, "failed to lookup github installation", "error", lookupErr)
		return http.StatusInternalServerError, lookupErr
	}

	ctx = execution.NewTenantContext(ctx, intg.TenantID)

	if ingestErr := h.provEvents.Ingest(ctx, pe); ingestErr != nil {
		slog.ErrorContext(ctx, "failed to ingest github webhook event", "error", ingestErr)
		return http.StatusInternalServerError, ingestErr
	}

	return http.StatusOK, nil
}

type pushEventPayload struct {
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
		Owner    struct {
			ID int64 `json:"id"`
		} `json:"owner"`
	} `json:"repository"`
}

type pullRequestPayload struct {
	Number     int `json:"number"`
	Repository struct {
		FullName string `json:"full_name"`
		Owner    struct {
			ID int64 `json:"id"`
		} `json:"owner"`
	} `json:"repository"`
}

func (h *webhookHandler) validateHMAC(body []byte, signature string) bool {
	sig := strings.TrimPrefix(signature, "sha256=")
	if len(sig) == 0 || len(h.secret) == 0 {
		return false
	}
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(body)
	expected := mac.Sum(nil)
	if !hmac.Equal(sigBytes, expected) {
		slog.Debug("github webhook signature mismatch",
			"expected", fmt.Sprintf("%x", expected),
			"got", fmt.Sprintf("%x", sigBytes),
		)
		return false
	}
	return true
}
