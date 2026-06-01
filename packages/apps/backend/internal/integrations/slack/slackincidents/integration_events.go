package slackincidents

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/internal/integrations/slack"
	goslack "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var (
	supportedIncidentSubcommands = []string{"new", "update", "status", "help"}
	incidentCommandsFormatted    = strings.Join(supportedIncidentSubcommands, ", ")
)

func (i *Integration) handleIncidentCommand(ctx context.Context, cmd *goslack.SlashCommand) (*goslack.Blocks, error) {
	// are we currently in an incident channel?
	var channelIncidentId uuid.UUID
	inc, incErr := i.incidents.Get(ctx, incident.ChatChannelID(cmd.ChannelID))
	if incErr != nil && !ent.IsNotFound(incErr) {
		slog.Error("unable to get incident by channel", "error", incErr)
		return slack.CommandErrorResponse(incErr.Error()), nil
	} else if inc != nil {
		channelIncidentId = inc.ID
	}

	subcmd := ""
	if args := strings.Split(cmd.Text, " "); len(args) > 0 {
		subcmd = args[0]
		slog.Debug("incident command", "text", cmd.Text, "subcmd", subcmd)
	}

	if subcmd == "" || subcmd == "update" || subcmd == "new" {
		meta := incidentDetailsModalViewMetadata{
			CommandChannelId: cmd.ChannelID,
			UserId:           cmd.UserID,
			IncidentId:       channelIncidentId,
		}
		if subcmd == "new" {
			meta.IncidentId = uuid.Nil
		}
		if subcmd == "update" && inc == nil {
			return slack.CommandErrorResponse("Not in an incident channel"), nil
		}
		// TODO: load these properly
		prefs := incidentPreferences{}
		view, viewErr := i.makeIncidentDetailsModalView(ctx, prefs, &meta)
		if viewErr != nil {
			slog.Error("failed creating incident details view", "error", viewErr)
			return slack.CommandErrorResponse("Failed to create incident details modal"), viewErr
		}
		if openModalErr := i.service.OpenModalView(ctx, cmd.TriggerID, *view); openModalErr != nil {
			return slack.CommandErrorResponse("Failed to open incident details modal"), openModalErr
		}
	} else if subcmd == "status" {
		if inc == nil {
			return slack.CommandErrorResponse("Not in an incident channel. Supported subcommands: " + incidentCommandsFormatted), nil
		}
		meta := incidentMilestoneModalViewMetadata{
			UserId:     cmd.UserID,
			IncidentId: channelIncidentId,
		}
		view, viewErr := i.makeIncidentMilestoneModalView(ctx, &meta)
		if viewErr != nil {
			slog.Error("failed creating incident milestone view", "error", viewErr)
			return slack.CommandErrorResponse("Failed to create incident milestone view"), viewErr
		}
		if openModalErr := i.service.OpenModalView(ctx, cmd.TriggerID, *view); openModalErr != nil {
			return slack.CommandErrorResponse("Failed to open incident milestone view"), openModalErr
		}
	} else if subcmd == "help" {
		// TODO
	} else {
		return slack.CommandErrorResponse(
			fmt.Sprintf("Invalid incident command '%s'. Supported subcommands: %s", subcmd, incidentCommandsFormatted),
		), nil
	}

	return nil, nil
}

var respondCallbackInnerEvents = mapset.NewSet(
	slackevents.AppHomeOpened,
	slackevents.AppMention,
	slackevents.AssistantThreadStarted,
	slackevents.Message,
)

func (i *Integration) handleCallbackEvent(ctx context.Context, ev *slackevents.EventsAPIEvent) error {
	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		return i.onUserHomeOpenedEvent(ctx, data)
	case *slackevents.AppMentionEvent:
		return i.onMentionEvent(ctx, data)
	case *slackevents.AssistantThreadStartedEvent:
		return i.onAssistantThreadStartedEvent(ctx, data)
	case *slackevents.MessageEvent:
		return i.onMessageEvent(ctx, data)
	default:
		slog.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
		return nil
	}
}

func (i *Integration) onMentionEvent(ctx context.Context, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	slog.Debug("mention event", "replyTs", replyTs)
	return nil
}

func (i *Integration) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
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

func (i *Integration) onAssistantThreadStartedEvent(ctx context.Context, data *slackevents.AssistantThreadStartedEvent) error {
	slog.Debug("assistant thread started")
	return nil
}

func (i *Integration) onUserHomeOpenedEvent(ctx context.Context, data *slackevents.AppHomeOpenedEvent) error {
	homeView, viewErr := makeUserHomeView(ctx)
	if viewErr != nil || homeView == nil {
		return fmt.Errorf("failed to create user home view: %w", viewErr)
	}

	req := goslack.PublishViewContextRequest{
		UserID: data.User,
		View:   *homeView,
		Hash:   nil,
	}
	resp, publishErr := i.service.GetClient(ctx).PublishViewContext(ctx, req)
	if publishErr != nil {
		slack.LogSlackViewErrorResponse(slog.Default(), publishErr, resp)
		return fmt.Errorf("failed to publish user home view: %w", publishErr)
	}

	return nil
}
