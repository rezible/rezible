package slack

import (
	"context"
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func (s *ChatService) HandleIncidentChatUpdate(ctx context.Context, args jobs.IncidentChatUpdate) error {
	inc, incErr := s.incidents.Get(ctx, args.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to fetch incident: %w", incErr)
	}

	channelId, chanErr := s.CreateIncidentChannel(ctx, inc)
	if chanErr != nil {
		return fmt.Errorf("failed to create incident channel: %w", chanErr)
	}
	inc.ChatChannelID = channelId

	if args.Created {
		if handleErr := s.onIncidentCreated(ctx, inc); handleErr != nil {
			log.Error().Err(handleErr).Msg("failed to handle incident creation callback")
		}

		if args.OriginChannelId != "" {
			channelLink := fmt.Sprintf("<#%s>", inc.ChatChannelID)
			msgText := fmt.Sprintf("Incident created: *%s* %s", inc.Title, channelLink)
			sendErr := s.sendMessage(ctx, args.OriginChannelId, slack.MsgOptionText(msgText, false), slack.MsgOptionBroadcast())
			if sendErr != nil {
				log.Warn().Err(sendErr).Msg("failed to send confirmation message")
			}
		}
	}

	// Post incident details as pinned message
	severity := "UNKNOWN"
	if inc.Edges.Severity != nil {
		severity = inc.Edges.Severity.Name
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.FrontendUrl(), inc.Slug)
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

	_, detailsTs, postErr := s.client.PostMessageContext(ctx, inc.ChatChannelID, slack.MsgOptionBlocks(detailsBlocks...))
	if postErr != nil {
		return fmt.Errorf("failed to post incident details: %w", postErr)
	}

	pinErr := s.client.AddPinContext(ctx, inc.ChatChannelID, slack.ItemRef{
		Channel:   inc.ChatChannelID,
		Timestamp: detailsTs,
	})
	if pinErr != nil {
		log.Warn().Err(pinErr).Msg("failed to pin details message")
	}

	// TODO: check for active alerts of linked components, playbooks, etc

	return nil
}

func (s *ChatService) onIncidentCreated(ctx context.Context, inc *ent.Incident) error {
	if bookmarkErr := s.AddIncidentDetailsBookmark(ctx, inc); bookmarkErr != nil {
		log.Warn().Err(bookmarkErr).Msg("failed to add incident details bookmark")
	}

	if annoErr := s.PostIncidentAnnouncement(ctx, inc); annoErr != nil {
		log.Warn().Err(annoErr).Msg("failed to post incident announcement")
	}

	return nil
}

func (s *ChatService) getTeamId(ctx context.Context) (string, error) {
	// TODO: get this from org context!!
	teams, _, listErr := s.client.ListTeamsContext(ctx, slack.ListTeamsParameters{})
	if listErr != nil {
		return "", listErr
	}
	for _, team := range teams {
		return team.ID, nil
	}
	return "", errors.New("no teams found")
}

func (s *ChatService) CreateIncidentChannel(ctx context.Context, inc *ent.Incident) (string, error) {
	channelName := fmt.Sprintf("inc-%s", inc.Slug)

	teamId, teamErr := s.getTeamId(ctx)
	if teamErr != nil {
		return "", teamErr
	}

	channel, createErr := s.client.CreateConversationContext(ctx, slack.CreateConversationParams{
		ChannelName: channelName,
		IsPrivate:   inc.Private,
		TeamID:      teamId,
	})
	if createErr != nil {
		return "", fmt.Errorf("create channel: %w", createErr)
	}

	updated, updateErr := s.incidents.Set(ctx, inc.ID, func(m *ent.IncidentMutation) {
		m.SetChatChannelID(channel.ID)
	})
	if updateErr != nil {
		return "", fmt.Errorf("set incident chatChannelID: %w", updateErr)
	}

	// TODO: do this async in a job?
	topicErr := s.UpdateIncidentChannelTopic(ctx, updated)
	if topicErr != nil {
		log.Warn().Err(topicErr).Str("channel_id", channel.ID).Msg("failed to set initial channel topic")
	}

	return channel.ID, nil
}

func (s *ChatService) UpdateIncidentChannelTopic(ctx context.Context, inc *ent.Incident) error {
	if inc.ChatChannelID == "" {
		return nil
	}

	severity := "UNKNOWN"
	if inc.Edges.Severity != nil {
		severity = inc.Edges.Severity.Name
	}

	status := "OPEN"
	if !inc.ClosedAt.IsZero() {
		status = "CLOSED"
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.FrontendUrl(), inc.Slug)
	topic := fmt.Sprintf("[%s] %s | Status: %s | %s", severity, inc.Title, status, webLink)

	_, setErr := s.client.SetTopicOfConversationContext(ctx, inc.ChatChannelID, topic)
	if setErr != nil {
		return fmt.Errorf("failed to set channel topic: %w", setErr)
	}
	return nil
}

func (s *ChatService) AddIncidentDetailsBookmark(ctx context.Context, inc *ent.Incident) error {
	if inc.ChatChannelID == "" {
		return nil
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.FrontendUrl(), inc.Slug)
	title := "View Incident Details"

	_, addErr := s.client.AddBookmark(inc.ChatChannelID, slack.AddBookmarkParameters{
		Title: title,
		Link:  webLink,
		Type:  "link",
	})
	if addErr != nil {
		return fmt.Errorf("failed to add bookmark: %w", addErr)
	}
	return nil
}

func (s *ChatService) getIncidentAnnouncementChannelId(ctx context.Context) (string, error) {
	return "#incident", nil
}

func (s *ChatService) PostIncidentAnnouncement(ctx context.Context, inc *ent.Incident) error {
	channelId, chanErr := s.getIncidentAnnouncementChannelId(ctx)
	if chanErr != nil {
		return fmt.Errorf("failed to get announcement channel: %w", chanErr)
	}

	severity := "UNKNOWN"
	if inc.Edges.Severity != nil {
		severity = inc.Edges.Severity.Name
	}

	// Create announcement blocks
	channelLink := fmt.Sprintf("<#%s|inc-%s>", inc.ChatChannelID, inc.Slug)
	headerText := fmt.Sprintf(":rotating_light: New Incident: *%s*", inc.Title)

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
				Text: fmt.Sprintf("*Severity:* %s\n*Channel:* %s", severity, channelLink),
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
