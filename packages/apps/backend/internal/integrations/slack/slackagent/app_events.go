package slackagent

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	as "github.com/rezible/rezible/ent/agentsession"
	"github.com/rezible/rezible/ent/predicate"
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
			return a.onMessageEvent(ctx, data)
		default:
			slog.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
			return nil
		}
	}
}

type aiChatAgentSessionMetadata struct {
	IsSlack           bool   `mapstructure:"slack"`
	IntegrationRef    string `mapstructure:"integration_ref"`
	SlackReplyChannel string `mapstructure:"slack_reply_channel"`
	SlackReplyTs      string `mapstructure:"slack_reply_ts"`
}

func (m aiChatAgentSessionMetadata) Encode() (map[string]any, error) {
	var md map[string]any
	return md, mapstructure.Decode(m, &md)
}

func (a *App) startOrContinueAgentThreadReply(ctx context.Context, userId uuid.UUID, msg string, md map[string]any) error {
	listSessionsParams := rez.ListAgentSessionsParams{
		ListParams: ent.ListParams{Limit: 1},
		Predicates: []predicate.AgentSession{as.OwnerUserID(userId)},
		Metadata:   md,
	}
	sessions, sessionsErr := a.agents.ListAgentSessions(ctx, listSessionsParams)
	if sessionsErr != nil && !ent.IsNotFound(sessionsErr) {
		return fmt.Errorf("failed to lookup agent sessions: %w", sessionsErr)
	}
	if len(sessions.Data) == 1 {
		slog.Debug("continuing existing agent session in thread")
		session := sessions.Data[0]
		params := &rez.RequestAgentTurnParams{
			Input:        &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(msg)},
			ParentTurnID: nil,
		}
		if _, requestErr := a.agents.RequestAgentTurn(ctx, session.ID, params); requestErr != nil {
			slog.Error("failed to request chat agent turn", "error", requestErr)
		}
		return nil
	}
	createSessionParams := rez.CreateAgentSessionParams{
		AgentName:   rezai.ChatAgent.Name,
		OwnerUserID: &userId,
		Input:       rezai.ChatAgentInput{UserId: userId, Message: msg},
		Metadata:    md,
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

	md := aiChatAgentSessionMetadata{
		IsSlack:           true,
		IntegrationRef:    cw.Integration().ExternalRef,
		SlackReplyChannel: data.Channel,
		SlackReplyTs:      replyTs,
	}
	sessionMetadata, mdErr := md.Encode()
	if mdErr != nil {
		return fmt.Errorf("failed to create agent session metadata: %w", mdErr)
	}

	usr, usrErr := a.users.Get(ctx, user.ChatID(data.User))
	if usrErr != nil {
		return fmt.Errorf("failed to lookup chat user: %w", usrErr)
	}

	cleanedText := strings.TrimSpace(mentionRe.ReplaceAllString(data.Text, ""))
	fmt.Printf("clean mention text: '%s'\n", cleanedText)

	return a.startOrContinueAgentThreadReply(ctx, usr.ID, cleanedText, sessionMetadata)
}

func (a *App) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
	//slog.Debug("message event", "message", data)
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
