package slackagent

import (
	"context"
	"fmt"
	"log/slog"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/ai"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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
			return a.onMessageEvent(ctx, data)
		default:
			slog.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
			return nil
		}
	}
}

func (a *App) onMentionEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	usr, usrErr := a.users.Get(ctx, user.ChatID(data.User))
	if usrErr != nil {
		return fmt.Errorf("failed to lookup chat user: %w", usrErr)
	}

	fmt.Printf("slack mention event: \n%+v\n\n", data)
	createRunParams := rez.CreateAgentRunParams{
		OwnerUserID: usr.ID,
		Input: ai.ChatAgentInput{
			UserId:  usr.ID,
			Message: data.Text,
		},
		Metadata: map[string]any{
			"integration_ref":        cw.Integration().ExternalRef,
			"slack_reply_channel_id": data.Channel,
			"slack_reply_thread_id":  replyTs,
		},
	}
	_, runErr := a.agents.CreateAgentRun(ctx, ai.ChatAgent.Name, createRunParams)
	if runErr != nil {
		slog.Error("failed to create chat agent run", "error", runErr)
	}
	return nil
}

func (a *App) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
	slog.Debug("message event", "message", data)
	/*
		threadTs := data.ThreadTimeStamp
		// TODO check if thread is 'monitored'

		slog.Debug("message event",
			"type", data.ChannelType,
			"text", data.Text,
			"thread", threadTs,
			"user", data.User,
		)
	*/

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
