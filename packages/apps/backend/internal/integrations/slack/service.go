package slackintegration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"golang.org/x/oauth2"
)

type (
	SlashCommandHandler        func(context.Context, *ent.Integration, *slack.SlashCommand) (*slack.Blocks, error)
	EventsApiHandler           func(context.Context, *ent.Integration, *slackevents.EventsAPIEvent) error
	InteractionCallbackHandler func(context.Context, *ent.Integration, *slack.InteractionCallback) error
)

type App interface {
	IntegrationName() string
	Config() rez.IntegrationsConfigSlackApp
	OAuthScopes() []string

	EventsApiHandler() EventsApiHandler
	SlashCommandHandlers() map[string]SlashCommandHandler
	InteractionCallbackHandlers() map[slack.InteractionType]InteractionCallbackHandler

	PublishProviderEventPipelineEventTypes() []slackevents.EventsAPIType
	RespondEventTypes() []slackevents.EventsAPIType
}

type AppService[A App] struct {
	app             A
	integrationName string

	msgs  rez.MessageService
	intgs rez.IntegrationService
	users rez.UserService

	oauthHandler                *oauthHandler
	webhookHandler              http.Handler
	socketModeListener          *socketModeListener
	slashCommandHandlers        map[string]SlashCommandHandler
	eventsApiHandler            EventsApiHandler
	interactionCallbackHandlers map[slack.InteractionType]InteractionCallbackHandler
}

func NewAppService[A App](app A, msgs rez.MessageService, intgs rez.IntegrationService, users rez.UserService, eventPipeline rez.ProviderEventPipelineService) (*AppService[A], error) {
	cfg := app.Config()
	s := &AppService[A]{
		app:                         app,
		integrationName:             app.IntegrationName(),
		msgs:                        msgs,
		intgs:                       intgs,
		users:                       users,
		oauthHandler:                newOAuthHandler(cfg.OAuthClientId, cfg.OAuthClientSecret, app.OAuthScopes()),
		webhookHandler:              http.NotFoundHandler(),
		slashCommandHandlers:        app.SlashCommandHandlers(),
		eventsApiHandler:            app.EventsApiHandler(),
		interactionCallbackHandlers: app.InteractionCallbackHandlers(),
	}

	if cfg.Enabled {
		if msgErr := s.registerMessageHandlers(); msgErr != nil {
			return nil, fmt.Errorf("register message handlers: %w", msgErr)
		}

		eventHandler := makeAppEventHandler(app, msgs, eventPipeline)

		if cfg.EnableSocketMode {
			socketModeClient := slack.New(cfg.BotToken, slack.OptionAppLevelToken(cfg.AppToken))
			s.socketModeListener = makeSocketModeListener(socketModeClient, eventHandler)
		} else {
			s.webhookHandler = makeWebhookHandler(cfg.WebhookSigningSecret, eventHandler)
		}
	}

	return s, nil
}

func (s *AppService[A]) WebhookHandler() http.Handler {
	return s.webhookHandler
}

func (s *AppService[A]) Lifecycle() *rez.ServiceLifecycle {
	if s.socketModeListener != nil {
		return s.socketModeListener.Lifecycle()
	}
	return nil
}

func (s *AppService[A]) App() A {
	return s.app
}

func (s *AppService[A]) registerMessageHandlers() error {
	return errors.Join(
		s.msgs.AddHandlers(
			messages.NewEventHandler(s.integrationName+".slash_command", s.handleSlashCommand),
			messages.NewEventHandler(s.integrationName+".interaction_callback", s.handleInteractionCallback),
			messages.NewEventHandler(s.integrationName+".events_api_callback", s.handleEventsApiCallbackEvent),
		),
	)
}

func (s *AppService[A]) OAuth2Config() *oauth2.Config {
	return s.oauthHandler.OAuth2Config()
}

func (s *AppService[A]) ValidateInstallationConfig(raw []byte) (rez.IntegrationInstallationConfig, error) {
	credentials, configErr := GetValidatedConfig(raw)
	if configErr != nil {
		return nil, configErr
	}
	return MakeInstallationConfig(s.integrationName, credentials), nil
}

func (s *AppService[A]) RetrieveInstallationTargetOptions(ctx context.Context, t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	cfg, cfgErr := s.oauthHandler.ExtractInstallationConfigFromToken(t)
	if cfgErr != nil {
		return nil, fmt.Errorf("extract config: %w", cfgErr)
	}
	targets := []rez.IntegrationInstallationTarget{{
		DisplayName: cfg.DisplayName(),
		Config:      MakeInstallationConfig(s.integrationName, cfg),
	}}
	return targets, nil
}

func (s *AppService[A]) createInstallationContext(ctx context.Context, ids InstallationIds) (*ent.Integration, context.Context, error) {
	lookupIntegrationsPred := in.And(in.Name(s.integrationName), in.ProviderInstallationRef(ids.InstallationTargetResourceRef()))
	intg, lookupErr := s.intgs.LookupInstallation(execution.NewSystemContext(ctx), lookupIntegrationsPred)
	if lookupErr != nil {
		return nil, ctx, fmt.Errorf("listing configured integrations: %w", lookupErr)
	}
	return intg, execution.NewTenantContext(ctx, intg.TenantID), nil
}

func (s *AppService[A]) LookupChatUser(ctx context.Context, userId string) (*ent.User, error) {
	return s.users.Get(ctx, user.ChatID(userId))
}

func (s *AppService[A]) createUserContext(ctx context.Context, userId string) (context.Context, error) {
	usr, usrErr := s.LookupChatUser(ctx, userId)
	if usrErr != nil {
		slog.ErrorContext(ctx, "failed to lookup chat user",
			"error", usrErr,
			"chat_id", userId,
		)
		return nil, fmt.Errorf("lookup user: %w", usrErr)
	}
	sess := &ent.UserAuthSession{
		TenantID: usr.TenantID,
		UserID:   usr.ID,
	}
	return execution.NewUserContext(ctx, sess), nil
}

func (s *AppService[A]) handleEventsApiCallbackEvent(baseCtx context.Context, ev *handleEventsApiCallbackEvent) error {
	if s.integrationName != ev.IntegrationName {
		return nil
	}
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	intg, ctx, intgsErr := s.createInstallationContext(baseCtx, InstallationIds{TeamId: cb.TeamID, EnterpriseId: cb.EnterpriseID})
	if intgsErr != nil {
		return fmt.Errorf("lookup integration: %w", intgsErr)
	}
	return s.eventsApiHandler(ctx, intg, &cb)
}

func (s *AppService[A]) handleInteractionCallback(baseCtx context.Context, ev *interactionCallbackEvent) error {
	if s.integrationName != ev.IntegrationName {
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

	installIds := InstallationIds{TeamId: ic.Team.ID, EnterpriseId: ic.Enterprise.ID}
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

func (s *AppService[A]) handleSlashCommand(baseCtx context.Context, ev *slashCommandEvent) error {
	if s.integrationName != ev.IntegrationName {
		return nil
	}

	cmd := ev.Command
	handler, hasHandler := s.slashCommandHandlers[cmd.Command]
	if !hasHandler {
		slog.Debug("unknown slack command, ignoring", "command", cmd.Command)
		return nil
	}

	installIds := InstallationIds{TeamId: cmd.TeamID, EnterpriseId: cmd.EnterpriseID}
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
