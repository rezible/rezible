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
	secret   string
	services *rez.Services
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

	if h.secret != "" {
		sig := r.Header.Get("X-Hub-Signature-256")
		if sig != "" {
			if !validateHMAC(body, sig, h.secret) {
				http.Error(w, "invalid signature", http.StatusUnauthorized)
				return
			}
		}
	}

	pe := rez.ProviderEvent{
		Provider:            integrationName,
		ProviderSource:      r.Header.Get("X-GitHub-Event"),
		SubjectRef:          githubWebhookSubjectRef(r.Header.Get("X-GitHub-Event"), r.Header.Get("X-GitHub-Delivery"), body),
		Payload:             body,
		ProviderDeliveryRef: r.Header.Get("X-GitHub-Delivery"),
	}

	if _, ingestErr := h.services.ProviderEvents.Ingest(r.Context(), pe); ingestErr != nil {
		slog.ErrorContext(r.Context(), "failed to ingest github webhook event", "error", ingestErr)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func githubWebhookSubjectRef(event, delivery string, body []byte) string {
	switch event {
	case "push":
		var payload struct {
			After      string `json:"after"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
		}
		if err := json.Unmarshal(body, &payload); err == nil && payload.Repository.FullName != "" && payload.After != "" {
			return fmt.Sprintf("github:%s:%s", payload.Repository.FullName, payload.After)
		}
	case "pull_request":
		var payload struct {
			Number     int `json:"number"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
		}
		if err := json.Unmarshal(body, &payload); err == nil && payload.Repository.FullName != "" && payload.Number != 0 {
			return fmt.Sprintf("github:%s:pr:%d", payload.Repository.FullName, payload.Number)
		}
	}
	if delivery != "" {
		return fmt.Sprintf("github:%s:%s", event, delivery)
	}
	return fmt.Sprintf("github:%s", event)
}

func validateHMAC(body []byte, signature, secret string) bool {
	sig := strings.TrimPrefix(signature, "sha256=")
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
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
