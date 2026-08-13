package slackagent

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	asb "github.com/rezible/rezible/ent/agentsessionbinding"
	"github.com/rezible/rezible/ent/user"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func (a *App) RespondEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{
		slackevents.AppHomeOpened,
		slackevents.AppMention,
		slackevents.AssistantThreadStarted,
		slackevents.Message,
	}
}

func (a *App) EventsApiHandler() slackintegration.EventsApiHandler {
	return func(ctx context.Context, ii *ent.Integration, ev *slackevents.EventsAPIEvent) error {
		cw, cwErr := slackintegration.NewClientWrapper(ii)
		if cwErr != nil {
			return fmt.Errorf("failed to create client wrapper: %w", cwErr)
		}

		switch data := ev.InnerEvent.Data.(type) {
		case *slackevents.AppHomeOpenedEvent:
			return a.onUserHomeOpenedEvent(ctx, cw, data)
		case *slackevents.AppMentionEvent:
			return a.onMentionEvent(ctx, cw, data)
		case *slackevents.AssistantThreadStartedEvent:
			return a.onAssistantThreadStartedEvent(ctx, data)
		case *slackevents.MessageEvent:
			return a.onMessageEvent(ctx, cw, data)
		default:
			slog.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
			return nil
		}
	}
}

const (
	slackAgentBindingSource             = "slack"
	slackAgentBindingResourceKindThread = "thread"
)

func slackThreadResourceRef(channelID string, threadTs string) string {
	return channelID + ":" + threadTs
}

func (a *App) lookupSlackThreadBinding(ctx context.Context, integrationID uuid.UUID, channelID string, threadTs string) (*ent.AgentSessionBinding, error) {
	return a.agents.LookupAgentSessionBinding(ctx,
		asb.IntegrationID(integrationID),
		asb.Source(slackAgentBindingSource),
		asb.ResourceKind(slackAgentBindingResourceKindThread),
		asb.ResourceRef(slackThreadResourceRef(channelID, threadTs)))
}

func (a *App) startAgentThreadReply(ctx context.Context, cw *slackintegration.ClientWrapper, userId uuid.UUID, msg string, channelID string, threadTs string) error {
	bindingParams := rez.AgentSessionBindingParams{
		IntegrationID: &cw.Integration().ID,
		Source:        slackAgentBindingSource,
		ResourceKind:  slackAgentBindingResourceKindThread,
		ResourceRef:   slackThreadResourceRef(channelID, threadTs),
		Metadata: map[string]any{
			"channel_id": channelID,
			"thread_ts":  threadTs,
		},
	}
	createSessionParams := rez.CreateAgentSessionParams{
		AgentName:   rezai.ChatAgent.Name,
		OwnerUserID: &userId,
		Input:       rezai.ChatAgentInput{UserId: userId, Message: msg},
		Bindings:    []rez.AgentSessionBindingParams{bindingParams},
	}
	if _, sessionErr := a.agents.CreateAgentSession(ctx, createSessionParams); sessionErr != nil {
		slog.Error("failed to create chat agent session", "error", sessionErr)
	}
	return nil
}

var mentionRe = regexp.MustCompile(`<@([^>]+)>`)

func (a *App) onMentionEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	usr, usrErr := a.users.Get(ctx, user.ChatID(data.User))
	if usrErr != nil {
		return fmt.Errorf("failed to lookup chat user: %w", usrErr)
	}

	cleanedText := strings.TrimSpace(mentionRe.ReplaceAllString(data.Text, ""))

	binding, bindingErr := a.lookupSlackThreadBinding(ctx, cw.Integration().ID, data.Channel, replyTs)
	if bindingErr != nil && !ent.IsNotFound(bindingErr) {
		return fmt.Errorf("failed to lookup agent session binding: %w", bindingErr)
	}
	if binding == nil {
		return a.startAgentThreadReply(ctx, cw, usr.ID, cleanedText, data.Channel, replyTs)
	}

	slog.Debug("continuing existing agent session in thread")
	params := &rez.RequestAgentTurnParams{
		Input: &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(cleanedText)},
	}
	_, requestErr := a.agents.RequestAgentTurn(ctx, binding.AgentSessionID, params)
	if requestErr != nil {
		slog.Error("failed to request chat agent turn", "error", requestErr)
	}
	return nil
}

func (a *App) onMessageEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.MessageEvent) error {
	if data.User == "" || data.BotID != "" || data.SubType != "" || data.ThreadTimeStamp == "" || data.ThreadTimeStamp == data.TimeStamp {
		return nil
	}
	cleanedText := strings.TrimSpace(mentionRe.ReplaceAllString(data.Text, ""))
	if cleanedText == "" {
		return nil
	}

	binding, bindingErr := a.lookupSlackThreadBinding(ctx, cw.Integration().ID, data.Channel, data.ThreadTimeStamp)
	if bindingErr != nil {
		if ent.IsNotFound(bindingErr) {
			return nil
		}
		return fmt.Errorf("lookup slack thread binding: %w", bindingErr)
	}

	slog.Debug("TODO: check if message is directed at agent", "binding", binding)

	return nil
}

func (a *App) onAssistantThreadStartedEvent(ctx context.Context, data *slackevents.AssistantThreadStartedEvent) error {
	slog.Debug("assistant thread started")
	return nil
}

func (a *App) onUserHomeOpenedEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.AppHomeOpenedEvent) error {
	homeView, viewErr := makeUserHomeView(ctx)
	if viewErr != nil || homeView == nil {
		return fmt.Errorf("failed to create user home view: %w", viewErr)
	}

	req := slack.PublishViewContextRequest{
		UserID: data.User,
		View:   *homeView,
		Hash:   nil,
	}
	resp, publishErr := cw.Client().PublishViewContext(ctx, req)
	if publishErr != nil {
		slackintegration.LogSlackViewErrorResponse(slog.Default(), publishErr, resp)
		return fmt.Errorf("failed to publish user home view: %w", publishErr)
	}

	return nil
}
