package slackagent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/riverqueue/river"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	asb "github.com/rezible/rezible/ent/agentsessionbinding"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/predicate"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
)

type App struct {
	cfg      rez.Config
	jobs     rez.JobService
	messages rez.MessageService
	intgs    rez.IntegrationService
	users    rez.UserService
	agents   rez.AgentSessionService
	events   rez.EventsService
}

func MakeApp(cfg rez.Config, jobSvc rez.JobService, msgs rez.MessageService, intgs rez.IntegrationService, users rez.UserService, agents rez.AgentSessionService, events rez.EventsService) (*App, error) {
	h := &App{
		cfg:      cfg,
		jobs:     jobSvc,
		messages: msgs,
		intgs:    intgs,
		users:    users,
		agents:   agents,
		events:   events,
	}

	jobs.RegisterWorkerFunc(h.handleSendMessageJob)
	jobs.RegisterWorkerFunc(h.handleCheckThreadResponseRequired)

	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (a *App) registerMessageHandlers() error {
	return a.messages.AddHandlers(
		messages.NewEventHandler("slackagent.OnAiAgentTurnFinished", a.onAiAgentTurnFinished))
}

func (a *App) IntegrationName() string {
	return integrationName
}

func (a *App) Config() rez.IntegrationsConfigSlackApp {
	return a.cfg.Integrations.Slack.Agent
}

func (a *App) PublishProviderEventPipelineEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{}
}

func (a *App) OAuthScopes() []string {
	return []string{
		"app_mentions:read",
		"assistant:write",
		"channels:history",
		"channels:join",
		"channels:read",
		"chat:write",
		"chat:write.customize",
		"chat:write.public",
		"commands",
		"groups:history",
		"groups:read",
		"im:history",
		"im:read",
		"im:write",
		"im:write.topic",
		"incoming-webhook",
		"metadata.message:read",
		"mpim:history",
		"pins:read",
		"reactions:read",
		"usergroups:read",
		"users.profile:read",
		"users:read",
		"users:read.email",
		"channels:write.topic",
		"channels:manage",
		"channels:write.invites",
	}
}

func (a *App) findMessageIdForAlertEvent(ctx context.Context, alertID uuid.UUID) (slackintegration.MessageId, error) {
	_ = ctx
	_ = alertID
	return "", nil
}

func (a *App) GetIntegrationClientWrapper(ctx context.Context, preds ...predicate.Integration) (*slackintegration.ClientWrapper, error) {
	preds = append(preds, in.IntegrationName(integrationName))
	intg, intgErr := a.intgs.LookupInstallation(ctx, in.And(preds...))
	if intgErr != nil {
		return nil, fmt.Errorf("query slack agent integration: %w", intgErr)
	}
	return slackintegration.NewClientWrapper(intg)
}

func (a *App) lookupAiAgentSessionBinding(ctx context.Context, sessionId uuid.UUID) (*ent.AgentSessionBinding, error) {
	preds := []predicate.AgentSessionBinding{
		asb.AgentSessionID(sessionId),
		asb.Source(slackAgentBindingSource),
		asb.ResourceKind(slackAgentBindingResourceKindThread),
	}
	binding, bindingErr := a.agents.LookupAgentSessionBinding(ctx, preds...)
	if bindingErr != nil {
		if ent.IsNotFound(bindingErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup: %w", bindingErr)
	}
	if binding.IntegrationID == nil {
		return nil, fmt.Errorf("slack session binding missing integration id")
	}
	return binding, nil
}

func (a *App) onAiAgentTurnFinished(ctx context.Context, ev *rezai.EventOnAgentTurnFinished) error {
	if len(ev.Response.Text()) == 0 {
		return nil
	}

	binding, bindingErr := a.lookupAiAgentSessionBinding(ctx, ev.AgentSessionId)
	if bindingErr != nil {
		return fmt.Errorf(" session binding: %w", bindingErr)
	} else if binding == nil {
		return nil
	}

	channelID, threadTs, ok := strings.Cut(binding.ResourceRef, ":")
	if !ok || channelID == "" || threadTs == "" {
		return fmt.Errorf("invalid slack thread resource ref %q", binding.ResourceRef)
	}

	args := SendMessageJobArgs{
		Message:       ev.Response.Text(),
		IntegrationID: *binding.IntegrationID,
		Channel:       channelID,
		ReplyTs:       threadTs,
	}
	if _, cmdErr := a.jobs.Insert(ctx, args, nil); cmdErr != nil {
		return fmt.Errorf("insert job: %w", cmdErr)
	}

	return nil
}

type CheckAgentThreadResponseRequiredArgs struct {
	BindingId uuid.UUID `json:"binding_id" river:"unique"`
	ReplyTs   string    `json:"reply_ts" river:"unique"`
}

func (a CheckAgentThreadResponseRequiredArgs) Kind() string {
	return "slack-check-agent-thread-response-required"
}

func (CheckAgentThreadResponseRequiredArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

func (a *App) handleCheckThreadResponseRequired(ctx context.Context, args CheckAgentThreadResponseRequiredArgs) error {
	binding, bindingErr := a.agents.LookupAgentSessionBinding(ctx, asb.ID(args.BindingId))
	if bindingErr != nil {
		return fmt.Errorf("lookup slack agent session binding: %w", bindingErr)
	}
	if binding.IntegrationID == nil {
		return nil
	}

	_, wrapperErr := a.GetIntegrationClientWrapper(ctx, in.ID(*binding.IntegrationID))
	if wrapperErr != nil {
		slog.Warn("failed to get slack integration client wrapper", "err", wrapperErr)
		return fmt.Errorf("get integration client wrapper: %w", wrapperErr)
	}

	slog.Debug("TODO: check if message is directed at agent", "binding", binding)

	return nil
}

type SendMessageJobArgs struct {
	IntegrationID uuid.UUID `json:"integration_id" river:"unique"`
	Message       string    `json:"message" river:"unique"`
	Channel       string    `json:"channel"`
	ReplyTs       string    `json:"reply_ts" river:"unique"`
}

func (a SendMessageJobArgs) Kind() string {
	return "slack-send-message"
}

func (SendMessageJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

func (a *App) handleSendMessageJob(ctx context.Context, args SendMessageJobArgs) error {
	cw, wrapperErr := a.GetIntegrationClientWrapper(ctx, in.ID(args.IntegrationID))
	if wrapperErr != nil {
		slog.Warn("failed to get slack integration client wrapper", "err", wrapperErr)
		return fmt.Errorf("get integration client wrapper: %w", wrapperErr)
	}

	_, _, msgErr := cw.Client().PostMessageContext(ctx, args.Channel,
		slack.MsgOptionMarkdownText(args.Message),
		slack.MsgOptionTS(args.ReplyTs))
	if msgErr != nil {
		return fmt.Errorf("post message: %w", msgErr)
	}

	return nil
}
