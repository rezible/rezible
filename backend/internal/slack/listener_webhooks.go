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
	"github.com/rs/zerolog/log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
)

type WebhookEventHandler struct {
	loader       *loader
	msgs         rez.MessageService
	bodyVerifier webhookVerifierFunc
}

func newWebhookEventHandler(l *loader, svcs *rez.Services) (*WebhookEventHandler, error) {
	whVerifier, verifierErr := makeWebhookVerifier()
	if verifierErr != nil {
		return nil, verifierErr
	}

	wh := &WebhookEventHandler{
		loader:       l,
		msgs:         svcs.Messages,
		bodyVerifier: whVerifier,
	}

	if msgsErr := wh.addMessageHandlers(); msgsErr != nil {
		return nil, msgsErr
	}

	return wh, nil
}

func (wh *WebhookEventHandler) addMessageHandlers() error {
	cmdsErr := wh.msgs.AddCommandHandlers(
		rez.NewCommandHandler("SlackHandleCommandEvent", wh.handleCommandEvent),
		rez.NewCommandHandler("SlackHandleInteractionEvent", wh.handleInteractionEvent))
	if cmdsErr != nil {
		return fmt.Errorf("command handlers: %w", cmdsErr)
	}
	evsErr := wh.msgs.AddEventHandlers(
		rez.NewEventHandler("SlackHandleCallbackEvent", wh.handleCallbackEvent))
	if evsErr != nil {
		return fmt.Errorf("event handlers: %w", evsErr)
	}
	return nil
}

func (wh *WebhookEventHandler) Handler() *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Timeout(3 * time.Second))
	mux.HandleFunc("/commands", wh.onCommandsWebhook)
	mux.HandleFunc("/interaction", wh.onInteractionsWebhook)
	mux.HandleFunc("/options", wh.onOptionsWebhook)
	mux.HandleFunc("/events", wh.onEventsWebhook)
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("path", r.URL.Path).
			Msg("slack webhook handler not found")
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB
)

type webhookVerifierFunc func(w http.ResponseWriter, r *http.Request) ([]byte, error)

func makeWebhookVerifier() (webhookVerifierFunc, error) {
	signingSecret := rez.Config.GetString("slack.webhook_signing_secret")
	if signingSecret == "" {
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
	if sendCmdErr := wh.msgs.SendCommand(r.Context(), cmd); sendCmdErr != nil {
		log.Error().Err(sendCmdErr).Msg("failed to publish slash command")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) handleCommandEvent(ctx context.Context, cmd *slack.SlashCommand) error {
	chat, _, chatErr := wh.loader.loadByTenantLookup(ctx, cmd.TeamID, cmd.EnterpriseID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}
	handled, payload, handleErr := chat.handleSlashCommand(ctx, cmd)
	if handleErr != nil {
		log.Error().Err(handleErr).Msg("failed to handle slash command")
		return handleErr
	}
	if !handled {
		log.Warn().Str("command", cmd.Command).Msg("unknown slack command, ignoring")
		return nil
	}
	if payload != nil {
		msg := slack.MsgOptionBlocks(payload.Blocks.BlockSet...)
		_, postErr := chat.postEphemeralMessage(ctx, cmd.ChannelID, cmd.UserID, msg)
		return postErr
	}
	return nil
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

	cmdErr := wh.msgs.SendCommand(r.Context(), webhookInteractionEvent{Payload: payload})
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

	chat, tenantCtx, chatErr := wh.loader.loadByTenantLookup(ctx, ic.Team.ID, ic.Enterprise.ID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}

	handled, _, handlerErr := chat.handleInteractionEvent(tenantCtx, &ic)
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
		pubErr := wh.msgs.PublishEvent(r.Context(), webhookCallbackEvent{Body: body})
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
	chat, tenantCtx, chatErr := wh.loader.loadByTenantLookup(ctx, cbe.TeamID, cbe.EnterpriseID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}
	handled, handlerErr := chat.handleCallbackEvent(tenantCtx, &cbe)
	if handlerErr != nil {
		return fmt.Errorf("failed to handle callback event: %w", handlerErr)
	}
	if !handled {
		log.Warn().Msg("didnt handle callback event")
	}
	return nil
}
