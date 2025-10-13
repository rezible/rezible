package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	signingSecretEnvVar = "SLACK_WEBHOOK_SIGNING_SECRET"
)

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB
)

type WebhookEventHandler struct {
	chatSvc       *ChatService
	signingSecret string
}

func NewWebhookEventHandler(chatSvc *ChatService) (*WebhookEventHandler, error) {
	signingSecret := os.Getenv(signingSecretEnvVar)
	if signingSecret == "" && !UseSocketMode() {
		return nil, fmt.Errorf("%s environment variable not set", signingSecretEnvVar)
	}
	return &WebhookEventHandler{chatSvc: chatSvc, signingSecret: signingSecret}, nil
}

func (wh *WebhookEventHandler) Handler() http.Handler {
	mux := chi.NewMux()
	mux.HandleFunc("/options", wh.handleOptionsWebhook)
	mux.HandleFunc("/events", wh.handleEventsWebhook)
	mux.HandleFunc("/interaction", wh.handleInteractionsWebhook)
	return mux
}

func (wh *WebhookEventHandler) verifyWebhook(w http.ResponseWriter, r *http.Request) error {
	bodyReader := http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes)
	body, bodyErr := io.ReadAll(bodyReader)
	if bodyErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return bodyErr
	}

	sv, svErr := slack.NewSecretsVerifier(r.Header, wh.signingSecret)
	if svErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return svErr
	}

	if _, writeErr := sv.Write(body); writeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return writeErr
	}

	if verificationErr := sv.Ensure(); verificationErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return verificationErr
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return nil
}

func (wh *WebhookEventHandler) handleOptionsWebhook(w http.ResponseWriter, r *http.Request) {
	if verifyErr := wh.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	log.Debug().Msg("get options")

	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) handleEventsWebhook(w http.ResponseWriter, r *http.Request) {
	if verifyErr := wh.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	body, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ev, evErr := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if evErr != nil {
		log.Error().Err(evErr).Msg("failed to parse event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if ev.Type == slackevents.URLVerification {
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
		return
	}

	if ev.Type == slackevents.AppRateLimited {
		log.Warn().Msg("slack app rate limited")
		w.WriteHeader(http.StatusOK)
		return
	}

	if ev.Type == slackevents.CallbackEvent {
		handled, cbErr := wh.chatSvc.onCallbackEventReceived(r.Context(), ev)
		if cbErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else if !handled {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}

	log.Warn().Str("type", ev.Type).Msg("didnt handle webhook event")
	w.WriteHeader(http.StatusBadRequest)
}

func (wh *WebhookEventHandler) handleInteractionsWebhook(w http.ResponseWriter, r *http.Request) {
	if verifyErr := wh.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	payload := r.PostFormValue("payload")
	if payload == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var ic slack.InteractionCallback
	if jsonErr := json.Unmarshal([]byte(payload), &ic); jsonErr != nil {
		log.Debug().Err(jsonErr).Msg("failed to unmarshal interaction payload")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()
	handled, handlerErr := wh.chatSvc.onInteractionEventReceived(ctx, &ic)
	if handlerErr != nil {
		log.Error().Err(handlerErr).Str("type", string(ic.Type)).Msg("failed to handle interaction")
		w.WriteHeader(http.StatusInternalServerError)
	} else if !handled {
		log.Warn().Str("type", string(ic.Type)).Msg("didnt handle webhook interaction event")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
