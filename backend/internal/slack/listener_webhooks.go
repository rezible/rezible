package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB
)

type webhookVerifierFunc func(w http.ResponseWriter, r *http.Request) ([]byte, error)

func makeWebhookVerifier() (webhookVerifierFunc, error) {
	signingSecret := rez.Config.GetString("slack.webhook_signing_secret")
	if signingSecret == "" && !UseSocketMode() {
		return nil, errors.New("slack.webhook_signing_secret not set")
	}
	verifyFn := func(w http.ResponseWriter, r *http.Request) ([]byte, error) {
		sv, svErr := slack.NewSecretsVerifier(r.Header, signingSecret)
		if svErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, svErr
		}

		bodyReader := http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes)
		body, bodyErr := io.ReadAll(bodyReader)
		if bodyErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, bodyErr
		}

		if _, writeErr := sv.Write(body); writeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return nil, writeErr
		}

		if verificationErr := sv.Ensure(); verificationErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, verificationErr
		}

		return body, nil
	}
	return verifyFn, nil
}

type WebhookEventHandler struct {
	chat         *ChatService
	bodyVerifier webhookVerifierFunc
}

func NewWebhookEventListener(chat *ChatService) (*WebhookEventHandler, error) {
	whVerifier, verifierErr := makeWebhookVerifier()
	if verifierErr != nil {
		return nil, verifierErr
	}

	wh := &WebhookEventHandler{
		chat:         chat,
		bodyVerifier: whVerifier,
	}

	cmdsErr := chat.messages.AddCommandHandlers(
		rez.NewCommandHandler("SlackHandleCommandEvent", wh.handleCommandEvent),
		rez.NewCommandHandler("SlackHandleInteractionEvent", wh.handleInteractionEvent))
	if cmdsErr != nil {
		return nil, fmt.Errorf("command handlers: %w", cmdsErr)
	}
	evsErr := chat.messages.AddEventHandlers(
		rez.NewEventHandler("SlackHandleCallbackEvent", wh.handleCallbackEvent))
	if evsErr != nil {
		return nil, fmt.Errorf("event handlers: %w", evsErr)
	}

	return wh, nil
}

func (wh *WebhookEventHandler) Handler() http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Timeout(3 * time.Second))
	mux.HandleFunc("/commands", wh.onCommandsWebhook)
	mux.HandleFunc("/interaction", wh.onInteractionsWebhook)
	mux.HandleFunc("/options", wh.onOptionsWebhook)
	mux.HandleFunc("/events", wh.onEventsWebhook)
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Msg("slack webhook handler not found")
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (wh *WebhookEventHandler) onCommandsWebhook(w http.ResponseWriter, r *http.Request) {
	body, verifyErr := wh.bodyVerifier(w, r)
	if verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	cmd, parseErr := slack.SlashCommandParse(r)
	if parseErr != nil {
		log.Error().Err(parseErr).Msg("failed to parse slash command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmdErr := wh.chat.messages.SendCommand(r.Context(), cmd)
	if cmdErr != nil {
		log.Error().Err(cmdErr).Msg("failed to publish slash command")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) handleCommandEvent(ctx context.Context, cmd *slack.SlashCommand) error {
	_, _, handleErr := wh.chat.handleSlashCommand(ctx, cmd)
	return handleErr
}

func (wh *WebhookEventHandler) onInteractionsWebhook(w http.ResponseWriter, r *http.Request) {
	body, verifyErr := wh.bodyVerifier(w, r)
	if verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	payload := r.PostFormValue("payload")
	if payload == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmdErr := wh.chat.messages.SendCommand(r.Context(), webhookInteractionEvent{Payload: payload})
	if cmdErr != nil {
		log.Error().Err(cmdErr).Msg("failed to publish interaction event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type webhookInteractionEvent struct {
	Payload string
}

func (wh *WebhookEventHandler) handleInteractionEvent(ctx context.Context, ie *webhookInteractionEvent) error {
	var ic slack.InteractionCallback
	if jsonErr := ic.UnmarshalJSON([]byte(ie.Payload)); jsonErr != nil {
		log.Debug().Err(jsonErr).Msg("failed to unmarshal interaction payload")
		return nil
	}

	handled, _, handlerErr := wh.chat.handleInteractionEvent(ctx, &ic)
	if handlerErr != nil {
		log.Error().Err(handlerErr).Str("type", string(ic.Type)).Msg("failed to handle interaction")
		return handlerErr
	}
	if !handled {
		log.Warn().Str("type", string(ic.Type)).Msg("didnt handle webhook interaction event")
	}
	return nil
}

func (wh *WebhookEventHandler) onOptionsWebhook(w http.ResponseWriter, r *http.Request) {
	_, verifyErr := wh.bodyVerifier(w, r)
	if verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	log.Debug().Msg("get options")

	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) onEventsWebhook(w http.ResponseWriter, r *http.Request) {
	body, verifyErr := wh.bodyVerifier(w, r)
	if verifyErr != nil {
		log.Error().Err(verifyErr).Msg("failed to verify webhook body")
		return
	}

	ev, evErr := slackevents.ParseEvent(body,
		// skip using the verification token as we verified via header
		slackevents.OptionNoVerifyToken())
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
		pubErr := wh.chat.messages.PublishEvent(r.Context(), webhookCallbackEvent{Body: body})
		if pubErr != nil {
			log.Error().Err(pubErr).Msg("failed to publish callback command")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Warn().Str("type", ev.Type).Msg("didnt handle webhook event")
	w.WriteHeader(http.StatusBadRequest)
}

type webhookCallbackEvent struct {
	Body []byte
}

func (wh *WebhookEventHandler) handleCallbackEvent(ctx context.Context, ev *webhookCallbackEvent) error {
	cbe, parseErr := slackevents.ParseEvent(ev.Body,
		// skip using the verification token as we verified via header
		slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("failed to parse event: %w", parseErr)
	}
	tenantId, tenantIdErr := wh.chat.lookupTeamTenantId(ctx, cbe.TeamID, cbe.EnterpriseID)
	if tenantIdErr != nil {
		return fmt.Errorf("failed to get tenant id: %w", tenantIdErr)
	}
	handled, handlerErr := wh.chat.handleCallbackEvent(access.TenantContext(ctx, tenantId), &cbe)
	if handlerErr != nil {
		return fmt.Errorf("failed to handle callback event: %w", handlerErr)
	}
	if !handled {
		log.Warn().Msg("didnt handle callback event")
	}
	return nil
}
