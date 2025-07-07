package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB

	createAnnotationActionCallbackID = "create_annotation"
)

type webhookHandler struct {
	signingSecret string
	prov          *ChatProvider
}

func newWebhookHandler(signingSecret string, prov *ChatProvider) *webhookHandler {
	return &webhookHandler{
		signingSecret: signingSecret,
		prov:          prov,
	}
}

func (h *webhookHandler) verifyWebhook(w http.ResponseWriter, r *http.Request) error {
	bodyReader := http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes)
	body, bodyErr := io.ReadAll(bodyReader)
	if bodyErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return bodyErr
	}

	sv, svErr := slack.NewSecretsVerifier(r.Header, h.signingSecret)
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

func (h *webhookHandler) handleOptions(w http.ResponseWriter, r *http.Request) {
	if verifyErr := h.verifyWebhook(w, r); verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	log.Debug().Msg("get options")

	w.WriteHeader(http.StatusOK)
}

func (h *webhookHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if verifyErr := h.verifyWebhook(w, r); verifyErr != nil {
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
			log.Error().Err(writeErr).Msg("failed to write challenge response")
		}
	} else if ev.Type == slackevents.AppRateLimited {
		w.WriteHeader(http.StatusOK)
		log.Warn().Msg("slack app rate limited")
	} else if ev.Type == slackevents.CallbackEvent {
		w.WriteHeader(http.StatusOK)
		// TODO: task queue?
		go handleCallbackEvent(h.prov, ev)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		log.Warn().Str("type", ev.Type).Msg("failed to handle event")
	}
}

func handleCallbackEvent(p *ChatProvider, ev slackevents.EventsAPIEvent) {
	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		p.onUserHomeOpenedEvent(data)
	case *slackevents.AppMentionEvent:
		p.onMentionEvent(data)
	case *slackevents.AssistantThreadStartedEvent:
		p.onAssistantThreadStartedEvent(data)
	case *slackevents.MessageEvent:
		p.onMessageEvent(data)
	default:
		log.Debug().
			Str("innerEventType", ev.InnerEvent.Type).
			Msg("unhandled slack callback event")
	}
}

func (h *webhookHandler) handleInteractions(w http.ResponseWriter, r *http.Request) {
	if verifyErr := h.verifyWebhook(w, r); verifyErr != nil {
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

	var handlerErr error
	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		handlerErr = h.handleMessageActionInteraction(ctx, &ic)
	case slack.InteractionTypeBlockActions:
		handlerErr = h.handleBlockActionInteraction(ctx, &ic)
	case slack.InteractionTypeViewSubmission:
		handlerErr = h.handleViewSubmissionInteraction(ctx, &ic)
	default:
		handlerErr = fmt.Errorf("unknown interaction type: %s", string(ic.Type))
	}
	if handlerErr == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		log.Error().Err(handlerErr).Str("type", string(ic.Type)).Msg("failed to handle interaction")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *webhookHandler) handleMessageActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationActionCallbackID:
		return h.prov.handleAnnotationModalInteraction(ctx, ic)
	default:
		return fmt.Errorf("unknown message action callback ID: %s", ic.CallbackID)
	}
}

func (h *webhookHandler) handleBlockActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		return h.prov.handleAnnotationModalInteraction(ctx, ic)
	}
	return nil
}

func (h *webhookHandler) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case createAnnotationModalViewCallbackID:
		return h.prov.handleAnnotationModalSubmission(ctx, ic)
	default:
		return fmt.Errorf("unknown view submission callback ID: %s", ic.View.CallbackID)
	}
}
