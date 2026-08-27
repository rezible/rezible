package slackagent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	asb "github.com/rezible/rezible/ent/agentsessionbinding"
	in "github.com/rezible/rezible/ent/integration"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/riverqueue/river"
	"github.com/slack-go/slack"
)

type SendMessageJobArgs struct {
	IntegrationID uuid.UUID `json:"integration_id" river:"unique"`
	Message       string    `json:"message" river:"unique"`
	Channel       string    `json:"channel"`
	ReplyTs       string    `json:"reply_ts" river:"unique"`
}

func (SendMessageJobArgs) Kind() string {
	return "slack-agent-send-message"
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

type HandleBoundAgentThreadMessagedArgs struct {
	BindingId uuid.UUID `json:"binding_id" river:"unique"`
	MessageTs string    `json:"message_ts" river:"unique"`
}

func (HandleBoundAgentThreadMessagedArgs) Kind() string {
	return "slack-agent-handle-bound-thread-response"
}

func (HandleBoundAgentThreadMessagedArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		MaxAttempts: 2,
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: jobs.UniqueStateNonCompleted,
		},
	}
}

func (a *App) RegisterJobs(registry *jobs.Registry) error {
	if registerErr := registry.AddWorkerFunc(a.handleSendMessageJob); registerErr != nil {
		return fmt.Errorf("send Slack Agent message: %w", registerErr)
	}
	if registerErr := registry.AddWorkerFunc(a.handleBoundAgentThreadMessagedJob); registerErr != nil {
		return fmt.Errorf("handle bound Slack Agent thread message: %w", registerErr)
	}
	return nil
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

func (a *App) handleBoundAgentThreadMessagedJob(ctx context.Context, args HandleBoundAgentThreadMessagedArgs) error {
	binding, bindingErr := a.agents.LookupAgentSessionBinding(ctx, asb.ID(args.BindingId))
	if bindingErr != nil {
		return fmt.Errorf("lookup slack agent session binding: %w", bindingErr)
	}

	res, resErr := a.getBoundThreadResource(binding)
	if resErr != nil {
		return fmt.Errorf("bound thread resource: %w", resErr)
	}

	intg, intgErr := binding.Edges.IntegrationOrErr()
	if intgErr != nil {
		return fmt.Errorf("no integration for binding: %w", intgErr)
	}

	cw, cwErr := slackintegration.NewClientWrapper(intg)
	if cwErr != nil {
		return fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	client := cw.Client()
	msgs, msgsErr := a.getThreadMessageContext(ctx, client, res, args.MessageTs, 3)
	if msgsErr != nil {
		return fmt.Errorf("failed to get thread message context: %w", msgsErr)
	}
	if len(msgs) == 0 {
		return fmt.Errorf("no messages")
	}

	usernamesMap, usernamesErr := a.getUserNames(ctx, client, msgs)
	if usernamesErr != nil {
		return fmt.Errorf("failed to get user profiles: %w", usernamesErr)
	}

	usrMsg := msgs[len(msgs)-1]
	workflowInput := rezai.ClassifyAgentThreadResponseInput{
		PreviousMessages: make([]string, max(0, len(msgs)-1)),
	}
	for i, msg := range msgs {
		name := msg.User
		if username, ok := usernamesMap[msg.User]; ok {
			name = username
		}
		msgText := fmt.Sprintf("%s: %s", name, msg.Text)
		if i < len(msgs)-1 {
			workflowInput.PreviousMessages[i] = msgText
		} else {
			workflowInput.UserMessage = msgText
		}
	}

	output, workflowErr := a.responseClassifier.Run(ctx, workflowInput)
	if workflowErr != nil {
		return fmt.Errorf("check response required: %w", workflowErr)
	}
	if !output.ShouldReply {
		return nil
	}
	params := &rez.RequestAgentTurnParams{
		Input: &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(usrMsg.Text)},
	}
	if _, requestErr := a.agents.RequestAgentTurn(ctx, binding.AgentSessionID, params); requestErr != nil {
		return fmt.Errorf("request agent turn: %w", requestErr)
	}
	return nil
}
