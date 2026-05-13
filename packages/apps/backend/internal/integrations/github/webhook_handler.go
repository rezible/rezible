package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	rez "github.com/rezible/rezible"
)

type webhookHandler struct {
	services *rez.Services
	secret   string
}

func newWebhookHandler(secret string, svcs *rez.Services) http.Handler {
	return &webhookHandler{secret: secret, services: svcs}
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

	subjectRef := fmt.Sprintf("github:%s", eventType)
	if eventType == sourcePushEvent {
		var payload pushEventPayload
		if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
			slog.Error("failed to unmarshal payload", "error", jsonErr.Error())
		}
		if payload.Repository.FullName != "" && payload.After != "" {
			subjectRef = fmt.Sprintf("github:%s:%s", payload.Repository.FullName, payload.After)
		}
	} else if eventType == sourcePullEvent {
		var payload pullRequestPayload
		if jsonErr := json.Unmarshal(body, &payload); jsonErr != nil {
			slog.Error("failed to unmarshal payload", "error", jsonErr.Error())
		}
		if payload.Repository.FullName != "" && payload.Number != 0 {
			subjectRef = fmt.Sprintf("github:%s:pr:%d", payload.Repository.FullName, payload.Number)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		return
	}
	//} else if deliveryRef != "" {
	//	subjectRef = fmt.Sprintf("github:%s:%s", eventType, deliveryRef)
	//}

	pe := rez.ProviderEvent{
		Provider:         integrationName,
		ProviderSource:   eventType,
		ProviderEventRef: deliveryRef,
		SubjectRef:       subjectRef,
		Payload:          body,
	}

	if _, ingestErr := h.services.ProviderEvents.Ingest(r.Context(), pe); ingestErr != nil {
		slog.ErrorContext(r.Context(), "failed to ingest github webhook event", "error", ingestErr)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type pushEventPayload struct {
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type pullRequestPayload struct {
	Number     int `json:"number"`
	Repository struct {
		FullName string `json:"full_name"`
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
