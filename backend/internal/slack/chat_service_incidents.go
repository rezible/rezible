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
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/jobs"
)

func (s *ChatService) incidentUpdateMessageHandler(m *message.Message) error {
	id, idErr := uuid.ParseBytes(m.Payload)
	if idErr != nil {
		return fmt.Errorf("failed to parse incident id: %w", idErr)
	}
	log.Debug().Msg("message handler, inserting job")
	params := jobs.InsertJobParams{
		Args:       jobs.HandleIncidentChatUpdate{IncidentId: id},
		Uniqueness: &jobs.JobUniquenessOpts{Args: true},
	}
	if jobErr := s.jobs.Insert(access.SystemContext(context.Background()), params); jobErr != nil {
		log.Error().Err(jobErr).Msg("failed to insert job")
	}

	return nil
}

func (s *ChatService) HandleIncidentChatUpdate(ctx context.Context, args jobs.HandleIncidentChatUpdate) error {
	inc, incErr := s.incidents.Get(access.SystemContext(ctx), args.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}
	ctx = access.TenantContext(ctx, inc.TenantID)

	if inc.ChatChannelID == "" {
		chanId, chanErr := s.createIncidentChannel(ctx, inc)
		if chanErr != nil {
			return fmt.Errorf("failed to create incident channel: %w", chanErr)
		}
		inc.ChatChannelID = chanId
	}

	// TODO: these should all be inserted as jobs

	if usersErr := s.ensureIncidentChannelUsersAdded(ctx, inc); usersErr != nil {
		log.Warn().Err(usersErr).Msg("failed to add users to incident channel")
	}

	if bookmarkErr := s.updateIncidentChannelInfo(ctx, inc); bookmarkErr != nil {
		log.Warn().Err(bookmarkErr).Msg("failed to update incident channel info")
	}

	return nil
}

func getIncidentChannelName(inc *ent.Incident) string {
	return fmt.Sprintf("incident-%s", inc.Slug)
}

func (s *ChatService) createIncidentChannel(ctx context.Context, inc *ent.Incident) (string, error) {
	client, clientErr := getClient(ctx, s.integrations)
	if clientErr != nil {
		return "", fmt.Errorf("failed to get client: %w", clientErr)
	}

	decl, declErr := fetchIncidentDeclaration(ctx, inc)
	if declErr != nil {
		return "", fmt.Errorf("fetching incident declaration: %w", declErr)
	}

	params := slack.CreateConversationParams{
		ChannelName: getIncidentChannelName(inc),
		IsPrivate:   false,
	}
	if decl != nil {
		params.TeamID = decl.TeamID
	}

	channel, createErr := client.CreateConversationContext(ctx, params)
	if createErr != nil {
		return "", fmt.Errorf("create channel: %w", createErr)
	}

	setFn := func(m *ent.IncidentMutation) {
		m.SetChatChannelID(channel.ID)
	}
	_, updateErr := s.incidents.Set(ctx, inc.ID, setFn, nil)
	if updateErr != nil {
		return "", fmt.Errorf("set incident chatChannelID: %w", updateErr)
	}
	inc.ChatChannelID = channel.ID

	announcementChannelId := "#incident"
	// TODO: fetch from config
	/*
		announcementChannelId, chanErr := s.getIncidentAnnouncementChannelId(ctx)
		if chanErr != nil {
			return "", fmt.Errorf("failed to get announcement channel: %w", chanErr)
		}
	*/

	// send message to user that created incident
	if decl != nil && decl.ChannelID != "" && decl.ChannelID != announcementChannelId {
		msgText := fmt.Sprintf("Incident created: <#%s>", channel.ID)
		_, sendErr := client.PostEphemeralContext(ctx, decl.ChannelID, decl.UserID, slack.MsgOptionText(msgText, false))
		if sendErr != nil {
			log.Warn().Err(sendErr).Msg("failed to send confirmation message")
		}
	}

	if annoErr := s.postIncidentAnnouncement(ctx, inc, announcementChannelId); annoErr != nil {
		log.Warn().Err(annoErr).Msg("failed to post incident announcement")
	}

	if detailsErr := s.sendIncidentChannelDetailsMessage(ctx, inc); detailsErr != nil {
		log.Warn().Err(detailsErr).Msg("failed to send incident channel details message")
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

func (s *ChatService) postIncidentAnnouncement(ctx context.Context, inc *ent.Incident, channelId string) error {
	severity := getIncidentSeverityName(inc)

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

	postErr := s.sendMessage(ctx, channelId, slack.MsgOptionBlocks(blocks...))
	if postErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", postErr)
	}

	return nil
}

func (s *ChatService) sendIncidentChannelDetailsMessage(ctx context.Context, inc *ent.Incident) error {
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

	return s.withClient(ctx, func(client *slack.Client) error {
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

func (s *ChatService) ensureIncidentChannelUsersAdded(ctx context.Context, inc *ent.Incident) error {

	return nil
}

func (s *ChatService) updateIncidentChannelInfo(ctx context.Context, inc *ent.Incident) error {
	severity := getIncidentSeverityName(inc)

	status := "OPEN"
	if !inc.ClosedAt.IsZero() {
		status = "CLOSED"
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug)
	topic := fmt.Sprintf("[%s] %s | %s | %s", severity, inc.Title, status)

	return s.withClient(ctx, func(client *slack.Client) error {
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
