package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type WebhookHandler struct {
	eventHandler  *EventHandler
	signingSecret string
}

func MakeWebhookHandler(signingSecret string, evth *EventHandler) (http.Handler, error) {
	if signingSecret == "" {
		return nil, fmt.Errorf("slack.webhooks.signing_secret not set")
	}

	h := &WebhookHandler{
		signingSecret: signingSecret,
		eventHandler:  evth,
	}

	return h.Handler(), nil
}

func (l *WebhookHandler) Handler() *chi.Mux {
	r := chi.NewMux()
	r.Use(middleware.Timeout(3 * time.Second))
	r.Use(l.requestVerifierMiddleware)
	r.HandleFunc("/commands", l.onCommands)
	r.HandleFunc("/interaction", l.onInteraction)
	r.HandleFunc("/options", l.onOptions)
	r.HandleFunc("/events", l.onEventsApi)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("slack webhook handler not found", "path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	return r
}

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB
)

func (l *WebhookHandler) requestVerifierMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sv, svErr := slack.NewSecretsVerifier(r.Header, l.signingSecret)
		if svErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reader := http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes)
		defer func(r io.ReadCloser) {
			if closeErr := r.Close(); closeErr != nil {
				slog.Error("failed to close webhook body reader", "error", closeErr)
			}
		}(reader)

		body, bodyErr := io.ReadAll(reader)
		if bodyErr != nil {
			mbErr := &http.MaxBytesError{}
			if maxBytes := errors.As(bodyErr, &mbErr); maxBytes {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				slog.Error("failed to read webhook body", "error", bodyErr)
			}
			return
		}

		if _, writeErr := sv.Write(body); writeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			slog.Error("failed to write payload to verifier", "error", writeErr)
			return
		}

		if verificationErr := sv.Ensure(); verificationErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)
	})
}

func (l *WebhookHandler) onCommands(w http.ResponseWriter, r *http.Request) {
	cmd, parseErr := slack.SlashCommandParse(r)
	if parseErr != nil {
		slog.Error("failed to parse slash command", "error", parseErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if handlerErr := l.eventHandler.OnSlashCommand(r.Context(), cmd); handlerErr != nil {
		slog.Error("failed to handle command event", "error", handlerErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookHandler) onInteraction(w http.ResponseWriter, r *http.Request) {
	payload := r.PostFormValue("payload")
	if payload == "" {
		slog.Warn("empty interaction payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if handlerErr := l.eventHandler.OnInteractionCallback(r.Context(), []byte(payload)); handlerErr != nil {
		slog.Error("failed to handle interaction event message", "error", handlerErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookHandler) onOptions(w http.ResponseWriter, r *http.Request) {
	// TODO, not currently used
	body := []byte("")
	if handlerErr := l.eventHandler.OnOptions(r.Context(), body); handlerErr != nil {
		slog.Error("failed to handle options event", "error", handlerErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookHandler) onEventsApi(w http.ResponseWriter, r *http.Request) {
	body, bodyErr := io.ReadAll(r.Body)
	if bodyErr != nil {
		slog.Error("failed to read webhook body", "error", bodyErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// skip using a verification token as middleware verified via header
	evt, evErr := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if evErr != nil {
		slog.Error("failed to parse event", "error", evErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if evt.Type == slackevents.URLVerification {
		if verifyErr := l.handleUrlVerificationEvent(w, body); verifyErr != nil {
			slog.Error("failed to handle url verification event", "error", verifyErr)
		}
		return
	}

	var handleErr error
	if evt.Type == slackevents.CallbackEvent {
		cb, cbOk := evt.Data.(*slackevents.EventsAPICallbackEvent)
		if !cbOk {
			slog.Error("failed to cast callback event")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		handleErr = l.eventHandler.OnCallbackEvent(r.Context(), cb, body)
	} else if evt.Type == slackevents.AppRateLimited {
		handleErr = l.eventHandler.OnAppRateLimitedEvent(r.Context())
	} else {
		handleErr = fmt.Errorf("unhandled event type: %s", evt.Type)
	}
	if handleErr != nil {
		slog.Error("failed to handle eventsAPI event", "error", handleErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookHandler) handleUrlVerificationEvent(w http.ResponseWriter, body []byte) error {
	var res *slackevents.ChallengeResponse
	if jsonErr := json.Unmarshal(body, &res); jsonErr != nil {
		return fmt.Errorf("unmarshal body: %w", jsonErr)
	}
	w.Header().Set("Content-Type", "text")
	if _, writeErr := w.Write([]byte(res.Challenge)); writeErr != nil {
		return fmt.Errorf("write challenge response body: %w", writeErr)
	}
	return nil
}
