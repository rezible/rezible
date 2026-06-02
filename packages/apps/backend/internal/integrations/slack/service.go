package slackintegration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/execution"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"golang.org/x/oauth2"
)

type (
	SlashCommandHandler        func(context.Context, *ent.Integration, *slack.SlashCommand) (*slack.Blocks, error)
	EventsApiHandler           func(context.Context, *ent.Integration, *slackevents.EventsAPIEvent) error
	InteractionCallbackHandler func(context.Context, *ent.Integration, *slack.InteractionCallback) error
)

type Service struct {
	integrationName string
	cfg             rez.IntegrationsConfigSlackApp

	msgs  rez.MessageService
	intgs rez.IntegrationService
	users rez.UserService

	eventHandler *EventHandler
	oauthHandler *oauthHandler

	webhookHandler     http.Handler
	socketModeListener *socketModeListener

	slashCommandHandlers        map[string]SlashCommandHandler
	eventsApiHandler            EventsApiHandler
	interactionCallbackHandlers map[slack.InteractionType]InteractionCallbackHandler
}

type NewServiceParams struct {
	AppConfig            rez.IntegrationsConfigSlackApp
	IntegrationName      string
	MessageService       rez.MessageService
	ProviderEventService rez.ProviderEventService
	OAuthScopes          []string

	EventsApiHandler            EventsApiHandler
	SlashCommandHandlers        map[string]SlashCommandHandler
	InteractionCallbackHandlers map[slack.InteractionType]InteractionCallbackHandler
}

func NewService(p NewServiceParams) (*Service, error) {
	cfg := p.AppConfig
	s := &Service{
		integrationName:             p.IntegrationName,
		cfg:                         cfg,
		webhookHandler:              http.NotFoundHandler(),
		oauthHandler:                NewOAuthHandler(cfg.OAuthClientId, cfg.OAuthClientSecret, p.OAuthScopes),
		eventsApiHandler:            p.EventsApiHandler,
		slashCommandHandlers:        p.SlashCommandHandlers,
		interactionCallbackHandlers: p.InteractionCallbackHandlers,
	}

	if !cfg.Enabled {
		return s, nil
	}

	s.eventHandler = &EventHandler{
		integrationName:           p.IntegrationName,
		messages:                  p.MessageService,
		respondEventTypes:         mapset.NewSet[slackevents.EventsAPIType](),
		provEvents:                p.ProviderEventService,
		publishProviderEventTypes: mapset.NewSet[slackevents.EventsAPIType](),
	}

	if p.AppConfig.EnableSocketMode {
		s.socketModeListener = makeSocketModeListener(s.createSingleTenantClient(), s.eventHandler)
	} else {
		whSecret := p.AppConfig.WebhookSigningSecret
		if whSecret == "" {
			return nil, fmt.Errorf("webhook signing secret not set")
		}
		s.webhookHandler = makeWebhookHandler(whSecret, s.eventHandler)
	}

	return s, nil
}

func (s *Service) createSingleTenantClient() *slack.Client {
	return slack.New(s.cfg.BotToken, slack.OptionAppLevelToken(s.cfg.AppToken))
}

func (s *Service) WebhookHandler() http.Handler {
	return s.webhookHandler
}

func (s *Service) Start(ctx context.Context) error {
	if s.socketModeListener != nil {
		return s.socketModeListener.Start(ctx)
	}
	return nil
}

func (s *Service) Shutdown(ctx context.Context) error {
	if s.socketModeListener != nil {
		return s.socketModeListener.Shutdown(ctx)
	}
	return nil
}

func (s *Service) registerMessageHandlers() error {
	return errors.Join(
		s.msgs.AddEventHandlers(
			rez.NewEventHandler(s.integrationName+".slash_command", s.handleSlashCommand),
			rez.NewEventHandler(s.integrationName+".interaction_callback", s.handleInteractionCallback),
			rez.NewEventHandler(s.integrationName+".events_api_callback", s.handleEventsApiCallbackEvent),
		),
	)
}

func (s *Service) OAuth2Config() *oauth2.Config {
	return s.oauthHandler.OAuth2Config()
}

func (s *Service) RetrieveInstallationTargetOptions(ctx context.Context, t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	return s.oauthHandler.ExtractInstallationTargetFromToken(t)
}

func (s *Service) createInstallationContext(ctx context.Context, ids IntegrationInstallIds) (*ent.Integration, context.Context, error) {
	intg, lookupErr := s.intgs.LookupByRef(execution.NewSystemContext(ctx), s.integrationName, ids.asRef())
	if lookupErr != nil {
		return nil, ctx, fmt.Errorf("listing configured integrations: %w", lookupErr)
	}
	ctx = execution.NewTenantContext(ctx, intg.TenantID)
	return intg, ctx, nil
}

func (s *Service) handleEventsApiCallbackEvent(baseCtx context.Context, ev *handleEventsApiCallbackEvent) error {
	if s.integrationName != ev.integrationName {
		return nil
	}
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	installIds := IntegrationInstallIds{TeamId: cb.TeamID, EnterpriseId: cb.EnterpriseID}
	intg, ctx, intgsErr := s.createInstallationContext(baseCtx, installIds)
	if intgsErr != nil {
		return fmt.Errorf("lookup integration: %w", intgsErr)
	}
	if err := s.eventsApiHandler(ctx, intg, &cb); err != nil {
		return fmt.Errorf("handle events api event: %w", err)
	}
	return nil
}

func (s *Service) createUserContext(ctx context.Context, userId string) (context.Context, error) {
	usr, usrErr := s.users.Get(ctx, user.ChatID(userId))
	if usrErr != nil {
		slog.ErrorContext(ctx, "failed to lookup chat user",
			"error", usrErr,
			"chat_id", userId,
		)
		return nil, fmt.Errorf("lookup user: %w", usrErr)
	}
	return execution.NewUserAuthContext(ctx, *usr, time.Time{}), nil
}

func (s *Service) handleInteractionCallback(baseCtx context.Context, ev *interactionCallbackEvent) error {
	if s.integrationName != ev.integrationName {
		return nil
	}

	var ic slack.InteractionCallback
	if err := ic.UnmarshalJSON(ev.Data); err != nil {
		return fmt.Errorf("invalid interaction payload: %w", err)
	}

	handler, ok := s.interactionCallbackHandlers[ic.Type]
	if !ok {
		// log unhandled
		slog.Warn("unhandled interaction type")
		return nil
	}

	installIds := IntegrationInstallIds{TeamId: ic.Team.ID, EnterpriseId: ic.Enterprise.ID}
	intg, ctx, intgsErr := s.createInstallationContext(baseCtx, installIds)
	if intgsErr != nil {
		return fmt.Errorf("lookup integration: %w", intgsErr)
	}

	var userErr error
	ctx, userErr = s.createUserContext(ctx, ic.User.ID)
	if userErr != nil {
		return fmt.Errorf("lookup user: %w", userErr)
	}

	return handler(ctx, intg, &ic)
}

func (s *Service) handleSlashCommand(baseCtx context.Context, ev *slashCommandEvent) error {
	if s.integrationName != ev.integrationName {
		return nil
	}

	cmd := ev.Command
	handler, hasHandler := s.slashCommandHandlers[cmd.Command]
	if !hasHandler {
		slog.Debug("unknown slack command, ignoring", "command", cmd.Command)
		return nil
	}

	installIds := IntegrationInstallIds{TeamId: cmd.TeamID, EnterpriseId: cmd.EnterpriseID}
	intg, ctx, intgsErr := s.createInstallationContext(baseCtx, installIds)
	if intgsErr != nil {
		return fmt.Errorf("lookup integration: %w", intgsErr)
	}

	var userErr error
	ctx, userErr = s.createUserContext(ctx, cmd.UserID)
	if userErr != nil {
		return fmt.Errorf("lookup user: %w", userErr)
	}

	response, handlerErr := handler(ctx, intg, &cmd)
	if handlerErr != nil {
		return fmt.Errorf("handling command: %w", handlerErr)
	}
	if response != nil {
		//_, msgErr := s.postEphemeralMessage(ctx, cmd.ChannelID, cmd.UserID, slack.MsgOptionBlocks(response.BlockSet...))
		//if msgErr != nil {
		//	return fmt.Errorf("failed to post ephemeral message: %w", msgErr)
		//}
	}
	return nil
}
