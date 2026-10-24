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
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
)

type (
	// SlashCommandHandler returns the blocks to reply with privately, or nil for no reply.
	SlashCommandHandler func(context.Context, *ent.Integration, *slack.SlashCommand) (*slack.Blocks, error)
	EventsApiHandler    func(context.Context, *ent.Integration, *slackevents.EventsAPIEvent) error
	// InteractionCallbackHandler may return errors to show on a submitted modal, which then stays open.
	InteractionCallbackHandler func(context.Context, *ent.Integration, *slack.InteractionCallback) (ViewErrors, error)
)

// ViewErrors are messages shown under a submitted modal's blocks, by block ID.
type ViewErrors map[string]string

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
		eventHandler := makeAppEventHandler(app, d.MessageQueue, d.ProviderEventPipelineService, instCtxFn,
			s.handleSlashCommand, s.handleInteractionCallback)

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

// handleInteractionCallback runs the interaction's handler while Slack waits. A submitted modal shows errors
// on its blocks; any other interaction's error is replied to privately.
func (s *AppService[A]) handleInteractionCallback(ctx context.Context, data []byte) (*slack.ViewSubmissionResponse, error) {
	var ic slack.InteractionCallback
	if jsonErr := ic.UnmarshalJSON(data); jsonErr != nil {
		return nil, fmt.Errorf("invalid interaction payload: %w", jsonErr)
	}
	handler, ok := s.interactionCallbackHandlers[ic.Type]
	if !ok {
		slog.WarnContext(ctx, "unhandled slack interaction type", "type", string(ic.Type))
		return nil, nil
	}

	var response *slack.ViewSubmissionResponse
	_ = execution.Do(ctx, "slack.interaction", func(ctx context.Context) error {
		viewErrors, handleErr := s.runInteraction(ctx, handler, &ic)
		if handleErr != nil {
			text := errorReplyText(ctx, "slack interaction failed", handleErr)
			if ic.Type == slack.InteractionTypeViewSubmission {
				viewErrors = ViewErrors{firstInputBlockID(ic.View): text}
			} else {
				replyPrivately(ctx, ic.ResponseURL, text)
			}
		}
		if len(viewErrors) > 0 {
			response = slack.NewErrorsViewSubmissionResponse(viewErrors)
		}
		// Only internal errors mark the span; refusals with public text are expected.
		if errs.CodeOf(handleErr) != errs.CodeInternal {
			return nil
		}
		return handleErr
	}, attribute.String("slack_interaction", string(ic.Type)))
	return response, nil
}

func (s *AppService[A]) runInteraction(ctx context.Context, handler InteractionCallbackHandler, ic *slack.InteractionCallback) (ViewErrors, error) {
	installIds := InstallationIds{TeamId: ic.Team.ID, EnterpriseId: ic.Enterprise.ID}
	intg, installCtx, intgsErr := s.createInstallationContext(ctx, installIds)
	if intgsErr != nil {
		return nil, fmt.Errorf("lookup integration: %w", intgsErr)
	}
	ctx = installCtx

	// Slack users the member sync has not seen yet continue as the installation's tenant, with no user;
	// handlers that need one check for it.
	userCtx, userErr := LookupUserContext(ctx, s.users, ic.User.ID)
	if userErr != nil && !errors.Is(userErr, ErrChatUserNotFound) {
		return nil, fmt.Errorf("lookup user: %w", userErr)
	}
	if userErr == nil {
		ctx = userCtx
	}

	return handler(ctx, intg, ic)
}

// handleSlashCommand runs the command's handler while Slack waits, and returns its reply, shown only to the
// person who ran it.
func (s *AppService[A]) handleSlashCommand(ctx context.Context, cmd slack.SlashCommand) *slack.WebhookMessage {
	handler, hasHandler := s.slashCommandHandlers[cmd.Command]
	if !hasHandler {
		slog.DebugContext(ctx, "unknown slack command, ignoring", "command", cmd.Command)
		return nil
	}

	var reply *slack.Blocks
	_ = execution.Do(ctx, "slack.command", func(ctx context.Context) error {
		blocks, commandErr := s.runSlashCommand(ctx, handler, &cmd)
		if commandErr != nil {
			blocks = CommandReply(errorReplyText(ctx, "slack command failed", commandErr))
		}
		reply = blocks
		// Only internal errors mark the span; refusals with public text are expected.
		if errs.CodeOf(commandErr) != errs.CodeInternal {
			return nil
		}
		return commandErr
	}, attribute.String("slack_command", cmd.Command))
	if reply == nil {
		return nil
	}
	return &slack.WebhookMessage{ResponseType: slack.ResponseTypeEphemeral, Blocks: reply}
}

func (s *AppService[A]) runSlashCommand(ctx context.Context, handler SlashCommandHandler, cmd *slack.SlashCommand) (*slack.Blocks, error) {
	installIds := InstallationIds{TeamId: cmd.TeamID, EnterpriseId: cmd.EnterpriseID}
	intg, installCtx, intgsErr := s.createInstallationContext(ctx, installIds)
	if intgsErr != nil {
		return nil, fmt.Errorf("lookup integration: %w", intgsErr)
	}

	userCtx, userErr := LookupUserContext(installCtx, s.users, cmd.UserID)
	if errors.Is(userErr, ErrChatUserNotFound) {
		return CommandReply(UnknownUserText), nil
	}
	if userErr != nil {
		return nil, fmt.Errorf("lookup user: %w", userErr)
	}
	return handler(userCtx, intg, cmd)
}

// errorReplyText logs a handler's error once and returns what to tell the person: its public message, or a
// reference to the trace when it has none.
func errorReplyText(ctx context.Context, msg string, handlerErr error) string {
	if errs.CodeOf(handlerErr) != errs.CodeInternal {
		slog.DebugContext(ctx, msg, "error", handlerErr)
		return errs.PublicMessage(handlerErr)
	}
	slog.ErrorContext(ctx, msg, "error", handlerErr)
	if spanContext := trace.SpanContextFromContext(ctx); spanContext.HasTraceID() {
		return "Something went wrong. Reference: " + spanContext.TraceID().String()
	}
	return "Something went wrong."
}

// firstInputBlockID is the block a modal's error is shown on when it concerns no block in particular.
func firstInputBlockID(view slack.View) string {
	for _, block := range view.Blocks.BlockSet {
		if input, isInput := block.(*slack.InputBlock); isInput {
			return input.BlockID
		}
	}
	return ""
}

// replyPrivately tells the person who acted, and only them, through the interaction's response URL. Actions
// inside a modal have none, so their errors are only logged.
func replyPrivately(ctx context.Context, responseURL string, text string) {
	if responseURL == "" {
		return
	}
	reply := &slack.WebhookMessage{
		Text:         text,
		ResponseType: slack.ResponseTypeEphemeral,
	}
	if replyErr := slack.PostWebhookContext(ctx, responseURL, reply); replyErr != nil {
		slog.ErrorContext(ctx, "reply to slack interaction", "error", replyErr)
	}
}
