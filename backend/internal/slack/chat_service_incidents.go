package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/jobs"
)

func (s *ChatService) setupIncidentUpdateHandler() {
	onIncidentUpdateMessage := func(ctx context.Context, m *message.Message) error {
		id, idErr := uuid.ParseBytes(m.Payload)
		if idErr != nil {
			return fmt.Errorf("failed to parse incident id: %w", idErr)
		}
		log.Debug().Msg("slack incident message handler, inserting job")
		args := handleIncidentChatUpdateArgs{IncidentId: id}
		opts := &jobs.InsertOpts{UniqueOpts: jobs.UniqueOpts{ByArgs: true}}
		if jobErr := s.jobs.Insert(ctx, args, opts); jobErr != nil {
			log.Error().Err(jobErr).Msg("failed to insert job")
		}
		return nil
	}
	s.messages.AddConsumerHandler("SlackIncidentUpdate", rez.MessageIncidentUpdatedTopic, onIncidentUpdateMessage)
	jobs.RegisterWorker(newIncidentChatUpdateWorker(s))
}

type handleIncidentChatUpdateArgs struct {
	IncidentId uuid.UUID
}

func (handleIncidentChatUpdateArgs) Kind() string { return "handle-incident-chat-update" }

type incidentChatUpdateWorker struct {
	chat *ChatService
	jobs.WorkerDefaults[handleIncidentChatUpdateArgs]
}

func newIncidentChatUpdateWorker(chat *ChatService) *incidentChatUpdateWorker {
	return &incidentChatUpdateWorker{chat: chat}
}

func (w *incidentChatUpdateWorker) Work(ctx context.Context, job *jobs.Job[handleIncidentChatUpdateArgs]) error {
	inc, incErr := w.chat.incidents.Get(ctx, job.Args.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}

	if inc.ChatChannelID == "" {
		chanId, chanErr := w.createIncidentChannel(ctx, inc)
		if chanErr != nil {
			return fmt.Errorf("failed to create incident channel: %w", chanErr)
		}
		if updateErr := inc.Update().SetChatChannelID(chanId).Exec(ctx); updateErr != nil {
			return fmt.Errorf("set incident chatChannelID: %w", updateErr)
		}
		inc.ChatChannelID = chanId
		if annoErr := w.postIncidentAnnouncement(ctx, inc); annoErr != nil {
			log.Warn().Err(annoErr).Msg("failed to post incident announcement")
		}

		if detailsErr := w.sendIncidentChannelDetailsMessage(ctx, inc); detailsErr != nil {
			log.Warn().Err(detailsErr).Msg("failed to send incident channel details message")
		}
	}

	// TODO: these should all be inserted as jobs

	if usersErr := w.ensureIncidentChannelUsersAdded(ctx, inc); usersErr != nil {
		log.Warn().Err(usersErr).Msg("failed to add users to incident channel")
	}

	if bookmarkErr := w.updateIncidentChannelInfo(ctx, inc); bookmarkErr != nil {
		log.Warn().Err(bookmarkErr).Msg("failed to update incident channel info")
	}

	return nil
}

func getIncidentChannelName(inc *ent.Incident) string {
	return fmt.Sprintf("incident-%s", inc.Slug)
}

func (w *incidentChatUpdateWorker) createIncidentChannel(ctx context.Context, inc *ent.Incident) (string, error) {
	client, clientErr := getClient(ctx, w.chat.integrations)
	if clientErr != nil {
		return "", fmt.Errorf("failed to get client: %w", clientErr)
	}

	createParams := slack.CreateConversationParams{
		ChannelName: getIncidentChannelName(inc),
		IsPrivate:   false,
	}

	decl, declErr := fetchIncidentDeclaration(ctx, inc)
	if declErr != nil {
		return "", fmt.Errorf("fetching incident declaration: %w", declErr)
	}
	if decl != nil {
		createParams.TeamID = decl.TeamID
	}

	channel, createErr := client.CreateConversationContext(ctx, createParams)
	if createErr != nil {
		return "", fmt.Errorf("create channel: %w", createErr)
	}

	// send message to user that created incident
	if decl != nil && decl.ChannelID != "" {
		msgText := fmt.Sprintf("Incident created: <#%s>", channel.ID)
		_, sendErr := client.PostEphemeralContext(ctx, decl.ChannelID, decl.UserID, slack.MsgOptionText(msgText, false))
		if sendErr != nil {
			log.Warn().Err(sendErr).Msg("failed to send confirmation message")
		}
	}

	return channel.ID, nil
}

func fetchIncidentDeclaration(ctx context.Context, inc *ent.Incident) (*slackIncidentDeclaration, error) {
	msQuery := inc.QueryMilestones().
		Where(incidentmilestone.KindEQ(incidentmilestone.KindResponse)).
		Where(incidentmilestone.Source(integrationName))
	ms, msErr := msQuery.First(ctx)
	if msErr != nil && !ent.IsNotFound(msErr) {
		return nil, fmt.Errorf("query milestones: %w", msErr)
	}
	if ms == nil {
		return nil, nil
	}
	parts := strings.Split(ms.ExternalID, "_")
	if len(parts) != 4 {
		log.Warn().Str("id", ms.ExternalID).Msg("invalid incident declaration milestone id")
		return nil, nil
	}
	return &slackIncidentDeclaration{
		TeamID:    parts[0],
		UserID:    parts[1],
		ChannelID: parts[2],
	}, nil
}

func getIncidentSeverityName(inc *ent.Incident) string {
	if inc.Edges.Severity != nil {
		return inc.Edges.Severity.Name
	}
	return "UNKNOWN"
}

func (w *incidentChatUpdateWorker) postIncidentAnnouncement(ctx context.Context, inc *ent.Incident) error {
	severity := getIncidentSeverityName(inc)

	announcementChannelId := "#incident"
	// TODO: fetch from config
	/*
		announcementChannelId, chanErr := s.getIncidentAnnouncementChannelId(ctx)
		if chanErr != nil {
			return "", fmt.Errorf("failed to get announcement channel: %w", chanErr)
		}
	*/

	headerText := fmt.Sprintf(":rotating_light: Incident Declared: <#%s> [*%s*] :rotating_light:", inc.ChatChannelID, severity)

	blocks := []slack.Block{
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: headerText,
			},
			nil, nil,
		),
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf("_%s_", inc.Title),
			},
			nil, nil,
		),
	}

	postErr := w.chat.sendMessage(ctx, announcementChannelId, slack.MsgOptionBlocks(blocks...))
	if postErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", postErr)
	}

	return nil
}

func (w *incidentChatUpdateWorker) sendIncidentChannelDetailsMessage(ctx context.Context, inc *ent.Incident) error {
	severity := getIncidentSeverityName(inc)

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug)
	detailsText := fmt.Sprintf("*Incident Details*\n*Title:* %s\n*Severity:* %s\n*Status:* %s\n*Web:* %s",
		inc.Title, severity, "OPEN", webLink)

	detailsBlocks := []slack.Block{
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: detailsText,
			},
			nil, nil,
		),
	}

	return w.chat.withClient(ctx, func(client *slack.Client) error {
		_, detailsTs, postErr := client.PostMessageContext(ctx, inc.ChatChannelID, slack.MsgOptionBlocks(detailsBlocks...))
		if postErr != nil {
			return fmt.Errorf("failed to post incident details: %w", postErr)
		}

		pinErr := client.AddPinContext(ctx, inc.ChatChannelID, slack.ItemRef{
			Channel:   inc.ChatChannelID,
			Timestamp: detailsTs,
		})
		if pinErr != nil {
			log.Warn().Err(pinErr).Msg("failed to pin details message")
		}

		// TODO: check for active alerts of linked components, playbooks, etc

		return nil
	})
}

func (w *incidentChatUpdateWorker) ensureIncidentChannelUsersAdded(ctx context.Context, inc *ent.Incident) error {

	return nil
}

func (w *incidentChatUpdateWorker) updateIncidentChannelInfo(ctx context.Context, inc *ent.Incident) error {
	severity := getIncidentSeverityName(inc)

	status := "OPEN"
	if !inc.ClosedAt.IsZero() {
		status = "CLOSED"
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug)
	topic := fmt.Sprintf("[%s] %s | %s", severity, inc.Title, status)

	return w.chat.withClient(ctx, func(client *slack.Client) error {
		_, setErr := client.SetTopicOfConversationContext(ctx, inc.ChatChannelID, topic)
		if setErr != nil {
			return fmt.Errorf("failed to set channel topic: %w", setErr)
		}

		_, addErr := client.AddBookmark(inc.ChatChannelID, slack.AddBookmarkParameters{
			Title: "View Incident Details",
			Link:  webLink,
			Type:  "link",
		})
		if addErr != nil {
			return fmt.Errorf("failed to add bookmark: %w", addErr)
		}

		return nil
	})
}
