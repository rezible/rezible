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
	loader       *serviceLoader
	msgs         rez.MessageService
	bodyVerifier webhookVerifierFunc
}

func newWebhookEventHandler(l *serviceLoader, svcs *rez.Services) (*WebhookEventHandler, error) {
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
		rez.NewCommandHandler("slack.webhooks.command", wh.processCommandEvent),
		rez.NewCommandHandler("slack.webhooks.interaction", wh.processInteractionEvent))
	if cmdsErr != nil {
		return fmt.Errorf("command handlers: %w", cmdsErr)
	}
	evsErr := wh.msgs.AddEventHandlers(
		rez.NewEventHandler("slack.webhooks.callback_event", wh.processCallbackEvent))
	if evsErr != nil {
		return fmt.Errorf("event handlers: %w", evsErr)
	}
	return nil
}

var (
	maxWebhookPayloadBytes = int64(4<<20 + 1) // 4 MB
)

type webhookVerifierFunc func(w http.ResponseWriter, r *http.Request) ([]byte, bool, error)

func makeWebhookVerifier() (webhookVerifierFunc, error) {
	signingSecret := rez.Config.GetString("slack.webhook_signing_secret")
	if signingSecret == "" {
		return nil, fmt.Errorf("slack.webhook_signing_secret not set")
	}
	verifyFn := func(w http.ResponseWriter, r *http.Request) ([]byte, bool, error) {
		sv, svErr := slack.NewSecretsVerifier(r.Header, signingSecret)
		if svErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, false, fmt.Errorf("making secrets verifier: %w", svErr)
		}

		body, bodyErr := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookPayloadBytes))
		if bodyErr != nil {
			mbErr := &http.MaxBytesError{}
			if maxBytes := errors.As(bodyErr, &mbErr); maxBytes {
				w.WriteHeader(http.StatusBadRequest)
				return nil, false, nil
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return nil, false, fmt.Errorf("reading body: %w", bodyErr)
			}
		}

		if _, writeErr := sv.Write(body); writeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return nil, false, fmt.Errorf("writing body to verifier: %w", writeErr)
		}

		if verificationErr := sv.Ensure(); verificationErr != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, false, nil
		}

		return body, true, nil
	}
	return verifyFn, nil
}

func (wh *WebhookEventHandler) Handler() *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.Timeout(3 * time.Second))
	mux.HandleFunc("/commands", wh.handleCommands)
	mux.HandleFunc("/interaction", wh.handleInteraction)
	mux.HandleFunc("/options", wh.handleOptions)
	mux.HandleFunc("/events", wh.handleEvents)
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("path", r.URL.Path).
			Msg("slack webhook handler not found")
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func (wh *WebhookEventHandler) handleCommands(w http.ResponseWriter, r *http.Request) {
	body, verified, verifyErr := wh.bodyVerifier(w, r)
	if !verified {
		if verifyErr != nil {
			log.Error().Err(verifyErr).Msg("failed to verify commands webhook body")
		}
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	cmd, parseErr := slack.SlashCommandParse(r)
	if parseErr != nil {
		log.Error().Err(parseErr).Msg("failed to parse slash command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if msgErr := wh.msgs.SendCommand(r.Context(), webhookCommandEvent{Command: cmd}); msgErr != nil {
		log.Error().Err(msgErr).Msg("failed to publish slash command message")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) handleInteraction(w http.ResponseWriter, r *http.Request) {
	body, verified, verifyErr := wh.bodyVerifier(w, r)
	if !verified {
		if verifyErr != nil {
			log.Error().Err(verifyErr).Msg("failed to verify commands webhook body")
		}
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	payload := r.PostFormValue("payload")
	if payload == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if msgErr := wh.msgs.SendCommand(r.Context(), webhookInteractionEvent{Payload: payload}); msgErr != nil {
		log.Error().Err(msgErr).Msg("failed to publish interaction event message")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) handleOptions(w http.ResponseWriter, r *http.Request) {
	_, verified, verifyErr := wh.bodyVerifier(w, r)
	if !verified {
		if verifyErr != nil {
			log.Error().Err(verifyErr).Msg("failed to verify options webhook body")
		}
		return
	}

	log.Debug().Msg("get options")

	w.WriteHeader(http.StatusOK)
}

func (wh *WebhookEventHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	body, verified, verifyErr := wh.bodyVerifier(w, r)
	if !verified {
		if verifyErr != nil {
			log.Error().Err(verifyErr).Msg("failed to verify events webhook body")
		}
		return
	}

	// skip using the verification token as we verified via header
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
		if msgErr := wh.msgs.PublishEvent(r.Context(), webhookCallbackEvent{Body: body}); msgErr != nil {
			log.Error().Err(msgErr).Msg("failed to publish callback event message")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Warn().Str("type", ev.Type).Msg("didnt handle webhook event")
	w.WriteHeader(http.StatusBadRequest)
}

type webhookCommandEvent struct {
	Command slack.SlashCommand
}

func (wh *WebhookEventHandler) processCommandEvent(baseCtx context.Context, ev *webhookCommandEvent) error {
	cmd := ev.Command
	chat, ctx, chatErr := wh.loader.fromTenantLookup(baseCtx, cmd.TeamID, cmd.EnterpriseID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}
	handled, payload, handleErr := chat.handleSlashCommand(ctx, &cmd)
	if handleErr != nil {
		log.Error().Err(handleErr).Msg("failed to handle slash command")
		return handleErr
	}
	if !handled {
		log.Warn().Str("command", cmd.Command).Msg("unknown slack command, ignoring")
		return nil
	}
	if payload != nil {
		msgOpts := slack.MsgOptionBlocks(payload.Blocks.BlockSet...)
		_, postErr := chat.postEphemeralMessage(ctx, cmd.ChannelID, cmd.UserID, msgOpts)
		return postErr
	}
	return nil
}

type webhookInteractionEvent struct {
	Payload string
}

func (wh *WebhookEventHandler) processInteractionEvent(ctx context.Context, ie *webhookInteractionEvent) error {
	var ic slack.InteractionCallback
	if jsonErr := ic.UnmarshalJSON([]byte(ie.Payload)); jsonErr != nil {
		log.Debug().Err(jsonErr).Msg("failed to unmarshal interaction payload")
		return nil
	}

	chat, tenantCtx, chatErr := wh.loader.fromTenantLookup(ctx, ic.Team.ID, ic.Enterprise.ID)
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

type webhookCallbackEvent struct {
	Body []byte
}

func (wh *WebhookEventHandler) processCallbackEvent(ctx context.Context, ev *webhookCallbackEvent) error {
	cbe, parseErr := slackevents.ParseEvent(ev.Body,
		// skip using the verification token as we verified via header
		slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("failed to parse event: %w", parseErr)
	}
	chat, tenantCtx, chatErr := wh.loader.fromTenantLookup(ctx, cbe.TeamID, cbe.EnterpriseID)
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
