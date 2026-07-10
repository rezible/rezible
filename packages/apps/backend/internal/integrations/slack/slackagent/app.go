package slackagent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/predicate"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg      rez.Config
	jobs     rez.JobService
	messages rez.MessageService
	intgs    rez.IntegrationService
	users    rez.UserService
	agents   rez.AiAgentService
	events   rez.EventsService
}

func MakeApp(cfg rez.Config, jobSvc rez.JobService, msgs rez.MessageService, intgs rez.IntegrationService, users rez.UserService, agents rez.AiAgentService, events rez.EventsService) (*App, error) {
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
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
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

func (a *App) registerMessageHandlers() error {
	return errors.Join(
		a.messages.AddEventHandlers(rez.NewEventHandler("slackagent.OnAiAgentRunSnapshot", a.onAiAgentRunOutput)))
}

func (a *App) onAiAgentRunOutput(ctx context.Context, ev *rez.EventOnAiAgentRunOutput) error {
	if ev.AgentName != ai.ChatAgent.Name {
		return nil
	}

	var metadata aiChatReplyAgentRunMetadata
	if mdErr := mapstructure.Decode(ev.AgentRunMetadata, &metadata); mdErr != nil {
		return fmt.Errorf("decode metadata: %w", mdErr)
	}
	if !metadata.IsSlackReply {
		fmt.Printf("not a slack reply run?: %+v\n", ev.AgentRunMetadata)
		return nil
	}

	output, outputErr := a.agents.GetAgentRunOutput(ctx, ev.AgentOutputId)
	if outputErr != nil {
		return fmt.Errorf("get agent run output: %w", outputErr)
	}

	parsed, parseErr := ai.ChatAgent.ParseOutput(output.Data)
	if parseErr != nil {
		return fmt.Errorf("parse output: %w", parseErr)
	}

	args := SendMessageJobArgs{
		Message:        parsed.Message,
		IntegrationRef: metadata.IntegrationRef,
		Channel:        metadata.SlackReplyChannel,
		ReplyTs:        metadata.SlackReplyTs,
	}
	jobOpts := &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
	if _, cmdErr := a.jobs.Insert(ctx, args, jobOpts); cmdErr != nil {
		return fmt.Errorf("send command: %w", cmdErr)
	}
	return nil
}

type SendMessageJobArgs struct {
	IntegrationRef string `json:"integration_ref"`
	Message        string `json:"message"`
	Channel        string `json:"channel"`
	ReplyTs        string `json:"reply_ts"`
}

func (a SendMessageJobArgs) Kind() string {
	return "slack-send-message"
}

func (a *App) handleSendMessageJob(ctx context.Context, args SendMessageJobArgs) error {
	cw, wrapperErr := a.GetIntegrationClientWrapper(ctx, in.ExternalRef(args.IntegrationRef))
	if wrapperErr != nil {
		slog.Warn("failed to get slack integration client wrapper", "err", wrapperErr)
		return fmt.Errorf("get integration client wrapper: %w", wrapperErr)
	}

	//return w.PostMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
	_, _, msgErr := cw.Client().PostMessageContext(ctx, args.Channel,
		slack.MsgOptionMarkdownText(args.Message),
		slack.MsgOptionTS(args.ReplyTs))
	if msgErr != nil {
		return fmt.Errorf("post message: %w", msgErr)
	}

	return nil
}
