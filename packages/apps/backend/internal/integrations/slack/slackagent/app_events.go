package slackagent

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
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

func (a *App) onMentionEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.AppMentionEvent) error {
	usr, usrErr := a.users.Get(ctx, user.ChatID(data.User))
	if usrErr != nil {
		return fmt.Errorf("failed to lookup chat user: %w", usrErr)
	}
	return a.onAgentMentionedByUser(ctx, usr, cw.Integration(), data)
}

var mentionRe = regexp.MustCompile(`<@([^>]+)>`)

func (a *App) onMessageEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.MessageEvent) error {
	if data.User == "" || data.BotID != "" || data.SubType != "" || data.ThreadTimeStamp == "" || data.ThreadTimeStamp == data.TimeStamp {
		return nil
	}

	if cleanedText := strings.TrimSpace(mentionRe.ReplaceAllString(data.Text, "")); cleanedText == "" {
		return nil
	}

	res := &agentThreadBindingResource{
		ChannelId: data.Channel,
		ThreadTs:  data.ThreadTimeStamp,
	}

	binding, bindingErr := a.lookupSlackThreadBinding(ctx, cw.Integration().ID, res)
	if bindingErr != nil {
		if ent.IsNotFound(bindingErr) {
			return nil
		}
		return fmt.Errorf("lookup slack thread binding: %w", bindingErr)
	}

	//usr, usrErr := a.users.Get(ctx, user.ChatID(data.User))
	//if usrErr != nil {
	//	return fmt.Errorf("failed to lookup chat user: %w", usrErr)
	//}

	return a.onBoundAgentThreadUserMessage(ctx, binding, data.Message)
}

func (a *App) onAssistantThreadStartedEvent(ctx context.Context, data *slackevents.AssistantThreadStartedEvent) error {
	slog.Debug("assistant thread started")
	return nil
}

func (a *App) onUserHomeOpenedEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.AppHomeOpenedEvent) error {
	homeView, viewErr := makeUserHomeViewRequest(ctx)
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
