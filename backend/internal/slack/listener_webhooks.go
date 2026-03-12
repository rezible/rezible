package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
)

type WebhookListener struct {
	handler       *eventHandler
	signingSecret string
}

func newWebhookListener(handler *eventHandler) (*WebhookListener, error) {
	signingSecret := rez.Config.GetString("slack.webhook_signing_secret")
	if signingSecret == "" {
		return nil, fmt.Errorf("slack.webhook_signing_secret not set")
	}

	return &WebhookListener{
		handler:       handler,
		signingSecret: signingSecret,
	}, nil
}

func (l *WebhookListener) Handler() *chi.Mux {
	r := chi.NewMux()
	r.Use(middleware.Timeout(3 * time.Second))
	r.Use(l.requestVerifierMiddleware)
	r.HandleFunc("/commands", l.onCommands)
	r.HandleFunc("/interaction", l.onInteraction)
	r.HandleFunc("/options", l.onOptions)
	r.HandleFunc("/events", l.onEventsApi)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("path", r.URL.Path).
			Msg("slack webhook handler not found")
		w.WriteHeader(http.StatusOK)
	})
	return r
}

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB
)

func (l *WebhookListener) requestVerifierMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sv, svErr := slack.NewSecretsVerifier(r.Header, l.signingSecret)
		if svErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reader := http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes)
		defer func(r io.ReadCloser) {
			if closeErr := r.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("failed to close webhook body reader")
			}
		}(reader)

		body, bodyErr := io.ReadAll(reader)
		if bodyErr != nil {
			mbErr := &http.MaxBytesError{}
			if maxBytes := errors.As(bodyErr, &mbErr); maxBytes {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error().Err(bodyErr).Msg("failed to read webhook body")
			}
			return
		}

		if _, writeErr := sv.Write(body); writeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error().Err(writeErr).Msg("failed to write payload to verifier")
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

func (l *WebhookListener) onCommands(w http.ResponseWriter, r *http.Request) {
	cmd, parseErr := slack.SlashCommandParse(r)
	if parseErr != nil {
		log.Error().Err(parseErr).Msg("failed to parse slash command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if handlerErr := l.handler.SlashCommand(r.Context(), cmd); handlerErr != nil {
		log.Error().Err(handlerErr).Msg("failed to handle command event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookListener) onInteraction(w http.ResponseWriter, r *http.Request) {
	payload := r.PostFormValue("payload")
	if payload == "" {
		log.Debug().Msg("empty interaction payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ic, icErr := slack.InteractionCallbackParse(r)
	if icErr != nil {
		log.Debug().Err(icErr).Msg("failed to parse interaction callback")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if handlerErr := l.handler.InteractionCallback(r.Context(), &ic); handlerErr != nil {
		log.Error().Err(handlerErr).Msg("failed to handle interaction event message")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookListener) onOptions(w http.ResponseWriter, r *http.Request) {
	// TODO, not currently used
	body := []byte("")
	if handlerErr := l.handler.Options(r.Context(), body); handlerErr != nil {
		log.Error().Err(handlerErr).Msg("failed to handle options event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (l *WebhookListener) onEventsApi(w http.ResponseWriter, r *http.Request) {
	body, bodyErr := io.ReadAll(r.Body)
	if bodyErr != nil {
		log.Error().Err(bodyErr).Msg("failed to read webhook body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// skip using a verification token as middleware verified via header
	ev, evErr := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if evErr != nil {
		log.Error().Err(evErr).Msg("failed to parse event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ev.Type == slackevents.URLVerification {
		l.handleUrlVerificationEvent(w, body)
	} else if ev.Type == slackevents.AppRateLimited {
		log.Warn().Msg("slack app rate limited")
		w.WriteHeader(http.StatusOK)
	} else if ev.Type == slackevents.CallbackEvent {
		if handleErr := l.handler.CallbackEvent(r.Context(), &ev); handleErr != nil {
			log.Error().Err(handleErr).Msg("failed to handle callback event")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		log.Warn().Str("type", ev.Type).Msg("didnt handle webhook event")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (l *WebhookListener) handleUrlVerificationEvent(w http.ResponseWriter, body []byte) {
	var res *slackevents.ChallengeResponse
	if jsonErr := json.Unmarshal(body, &res); jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text")
	_, writeErr := w.Write([]byte(res.Challenge))
	if writeErr != nil {
		log.Error().Err(writeErr).Msg("failed to write url verification challenge response")
	}
}
