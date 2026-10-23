package slackintegration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
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

	MakeMessageHandlers() []rez.MessageEventHandler

	PublishEventTypes() []slackevents.EventsAPIType
	RespondEventTypes() []slackevents.EventsAPIType
}

type createInstallationCtxFunc func(context.Context, InstallationIds) (context.Context, error)

type AppService[A App] struct {
	app             A
	integrationName string

	msgs          rez.MessageQueue
	installations rez.IntegrationInstallationLookup
	users         rez.UserService

	oauthHandler                *oauthHandler
	webhookHandler              http.Handler
	socketModeListener          *socketModeListener
	slashCommandHandlers        map[string]SlashCommandHandler
	eventsApiHandler            EventsApiHandler
	interactionCallbackHandlers map[slack.InteractionType]InteractionCallbackHandler
}

func (s *AppService[A]) App() A { return s.app }

type AppServiceDependencies struct {
	MessageQueue                 rez.MessageQueue
	Installations                rez.IntegrationInstallationLookup
	UserService                  rez.UserService
	ProviderEventPipelineService rez.ProviderEventPipelineService
}

func NewAppService[A App](app A, d *AppServiceDependencies) (*AppService[A], error) {
	cfg := app.Config()
	s := &AppService[A]{
		app:                         app,
		integrationName:             app.IntegrationName(),
		msgs:                        d.MessageQueue,
		installations:               d.Installations,
		users:                       d.UserService,
		oauthHandler:                newOAuthHandler(cfg.OAuthClientId, cfg.OAuthClientSecret, app.OAuthScopes()),
		webhookHandler:              http.NotFoundHandler(),
		slashCommandHandlers:        app.SlashCommandHandlers(),
		eventsApiHandler:            app.EventsApiHandler(),
		interactionCallbackHandlers: app.InteractionCallbackHandlers(),
	}

	if cfg.Enabled {
		instCtxFn := func(ctx context.Context, ids InstallationIds) (context.Context, error) {
			_, instCtx, lookupErr := s.createInstallationContext(ctx, ids)
			return instCtx, lookupErr
		}
		eventHandler := makeAppEventHandler(app, d.MessageQueue, d.ProviderEventPipelineService, instCtxFn)

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

func (s *AppService[A]) HasLifecycle() bool {
	return s.socketModeListener != nil
}

func (s *AppService[A]) Run(ctx context.Context, ready chan<- struct{}) error {
	if s.socketModeListener == nil {
		return fmt.Errorf("socketModeListener is nil")
	}

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return s.socketModeListener.runClient(groupCtx)
	})

	var readyOnce sync.Once
	onSocketListenerConnected := func() {
		readyOnce.Do(func() {
			close(ready)
		})
	}

	group.Go(func() error {
		return s.socketModeListener.runEventConsumerLoop(groupCtx, onSocketListenerConnected)
	})

	if runErr := group.Wait(); runErr != nil && !(errors.Is(runErr, context.Canceled) && ctx.Err() != nil) {
		return runErr
	}
	return nil
}

func (s *AppService[A]) Shutdown(context.Context) error {
	return nil
}

func (s *AppService[A]) Config() rez.IntegrationsConfigSlackApp {
	return s.app.Config()
}

func (s *AppService[A]) MakeMessageHandlers() []rez.MessageEventHandler {
	return append(s.app.MakeMessageHandlers(),
		messages.NewEventHandler(s.integrationName+".slash_command", s.handleSlashCommand),
		messages.NewEventHandler(s.integrationName+".interaction_callback", s.handleInteractionCallback),
		messages.NewEventHandler(s.integrationName+".events_api_callback", s.handleEventsApiCallbackEvent),
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
	intg, lookupErr := s.installations.LookupInstallation(execution.NewSystemContext(ctx), lookupIntegrationsPred)
	if lookupErr != nil {
		return nil, ctx, fmt.Errorf("listing configured integrations: %w", lookupErr)
	}
	return intg, execution.NewTenantContext(ctx, intg.TenantID), nil
}

func (s *AppService[A]) handleEventsApiCallbackEvent(baseCtx context.Context, ev *eventsApiCallbackEvent) error {
	if s.integrationName != ev.IntegrationName {
		return nil
	}
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	installIds := InstallationIds{TeamId: cb.TeamID, EnterpriseId: cb.EnterpriseID}
	intg, ctx, intgsErr := s.createInstallationContext(baseCtx, installIds)
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

	// Slack users the member sync has not seen yet continue as the installation's tenant, with no user;
	// handlers that need one check for it.
	userCtx, userErr := LookupUserContext(ctx, s.users, ic.User.ID)
	if userErr != nil && !errors.Is(userErr, ErrChatUserNotFound) {
		return fmt.Errorf("lookup user: %w", userErr)
	}
	if userErr == nil {
		ctx = userCtx
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
	intg, installCtx, intgsErr := s.createInstallationContext(baseCtx, installIds)
	if intgsErr != nil {
		return fmt.Errorf("lookup integration: %w", intgsErr)
	}
	ctx := installCtx

	userCtx, userErr := LookupUserContext(ctx, s.users, cmd.UserID)
	if errors.Is(userErr, ErrChatUserNotFound) {
		return s.replyUnknownUser(ctx, &cmd)
	}
	if userErr != nil {
		return fmt.Errorf("lookup user: %w", userErr)
	}
	ctx = userCtx

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

// replyUnknownUser tells the Slack user who ran the command, and only them, that Rezible does not know them
// yet.
func (s *AppService[A]) replyUnknownUser(ctx context.Context, cmd *slack.SlashCommand) error {
	reply := &slack.WebhookMessage{
		Text:         UnknownUserText,
		ResponseType: slack.ResponseTypeEphemeral,
	}
	if replyErr := slack.PostWebhookContext(ctx, cmd.ResponseURL, reply); replyErr != nil {
		return fmt.Errorf("reply to unknown slack user: %w", replyErr)
	}
	return nil
}
