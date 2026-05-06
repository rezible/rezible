package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
		Provider:  integrationName,
		Source:    r.Header.Get("X-GitHub-Event"),
		Payload:   body,
		DedupeKey: r.Header.Get("X-GitHub-Delivery"),
	}

	if ingestErr := h.services.ProviderEvents.Ingest(r.Context(), pe); ingestErr != nil {
		slog.ErrorContext(r.Context(), "failed to ingest github webhook event", "error", ingestErr)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
