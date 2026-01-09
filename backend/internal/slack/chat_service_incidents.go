package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
)

func (s *ChatService) addIncidentMessageHandlers() error {
	incidentEventHandler := &incidentChatEventHandler{chat: s}
	cmdsErr := s.messages.AddCommandHandlers(
		rez.NewCommandHandler("slack.CreateIncidentChannel", incidentEventHandler.createIncidentChannel))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	eventsErr := s.messages.AddEventHandlers(
		rez.NewEventHandler("slack.OnIncidentUpdate", incidentEventHandler.onUpdate))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}

	return nil
}

type incidentChatEventHandler struct {
	chat *ChatService
}

func (h *incidentChatEventHandler) onUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	inc, incErr := h.chat.incidents.Get(ctx, ev.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}

	if inc.ChatChannelID == "" {
		createCmdErr := h.chat.messages.SendCommand(ctx, &cmdCreateIncidentChannel{IncidentID: inc.ID})
		if createCmdErr != nil {
			return fmt.Errorf("failed to send create incident channel command: %w", createCmdErr)
		}
		return nil
	}

	// TODO: these should all be inserted as jobs

	if detailsErr := h.updateIncidentChannelDetailsMessage(ctx, inc); detailsErr != nil {
		log.Warn().Err(detailsErr).Msg("failed to update incident details message")
	}

	if bookmarkErr := h.updateIncidentChannelInfo(ctx, inc); bookmarkErr != nil {
		log.Warn().Err(bookmarkErr).Msg("failed to update incident channel info")
	}

	if usersErr := h.ensureIncidentChannelUsersAdded(ctx, inc); usersErr != nil {
		log.Warn().Err(usersErr).Msg("failed to add users to incident channel")
	}

	return nil
}

func getIncidentChannelName(inc *ent.Incident) string {
	return fmt.Sprintf("incident-%s", inc.Slug)
}

type cmdCreateIncidentChannel struct {
	IncidentID uuid.UUID
}

func (h *incidentChatEventHandler) createIncidentChannel(ctx context.Context, data *cmdCreateIncidentChannel) error {
	inc, incErr := h.chat.incidents.Get(ctx, data.IncidentID)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}
	if inc.ChatChannelID != "" {
		return nil
	}

	createParams := slack.CreateConversationParams{
		ChannelName: getIncidentChannelName(inc),
		TeamID:      "",
		IsPrivate:   false,
	}

	client, clientErr := getClient(ctx, h.chat.integrations)
	if clientErr != nil {
		return fmt.Errorf("failed to get client: %w", clientErr)
	}

	// TODO: check if channel exists first

	channel, createErr := client.CreateConversationContext(ctx, createParams)
	if createErr != nil {
		return fmt.Errorf("create channel: %w", createErr)
	}

	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		m.SetChatChannelID(channel.ID)
		return nil
	}
	if _, updateErr := h.chat.incidents.Set(ctx, inc.ID, setFn); updateErr != nil {
		return fmt.Errorf("set incident chatChannelID: %w", updateErr)
	}
	inc.ChatChannelID = channel.ID

	if msgErr := h.sendUserCreationMessage(ctx, inc); msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to send user incident creation message")
	}
	if annoErr := h.postIncidentAnnouncement(ctx, inc); annoErr != nil {
		log.Warn().Err(annoErr).Msg("failed to post incident announcement")
	}

	return nil
}

func (h *incidentChatEventHandler) sendUserCreationMessage(ctx context.Context, inc *ent.Incident) error {
	msQuery := inc.QueryMilestones().
		Where(im.KindEQ(im.KindResponse)).
		Where(im.Source(integrationName))
	ms, msErr := msQuery.First(ctx)
	if msErr != nil && !ent.IsNotFound(msErr) {
		return fmt.Errorf("query milestones: %w", msErr)
	}
	if ms == nil {
		return nil
	}
	// from handleIncidentModalSubmission
	parts := strings.Split(ms.ExternalID, "_")
	if len(parts) != 4 {
		log.Warn().Str("id", ms.ExternalID).Msg("invalid incident declaration milestone id")
		return nil
	}
	channelID := parts[1]
	userID := parts[2]
	// send message to user that created incident
	msgText := fmt.Sprintf("Incident created: <#%s>", inc.ChatChannelID)
	return h.chat.withClient(ctx, func(client *slack.Client) error {
		_, sendErr := client.PostEphemeralContext(ctx, channelID, userID, slack.MsgOptionText(msgText, false))
		if sendErr != nil {
			return fmt.Errorf("failed to send confirmation message: %w", sendErr)
		}
		return nil
	})
}

func (h *incidentChatEventHandler) postIncidentAnnouncement(ctx context.Context, inc *ent.Incident) error {
	sev, sevErr := h.chat.incidents.GetIncidentSeverity(ctx, inc.SeverityID)
	if sevErr != nil {
		return fmt.Errorf("failed to get incident severity: %w", sevErr)
	}

	announcementChannelId := "#incident"
	// TODO: fetch from config
	/*
		announcementChannelId, chanErr := s.getIncidentAnnouncementChannelId(ctx)
		if chanErr != nil {
			return "", fmt.Errorf("failed to get announcement channel: %w", chanErr)
		}
	*/

	headerText := fmt.Sprintf(":rotating_light: Incident Declared: <#%s> [*%s*] :rotating_light:",
		inc.ChatChannelID, sev.Name)

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

	postErr := h.chat.sendMessage(ctx, announcementChannelId, slack.MsgOptionBlocks(blocks...))
	if postErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", postErr)
	}

	return nil
}

func (h *incidentChatEventHandler) makeIncidentDetailsMessageBlocks(ctx context.Context, inc *ent.Incident) ([]slack.Block, error) {
	sev, sevErr := h.chat.incidents.GetIncidentSeverity(ctx, inc.SeverityID)
	if sevErr != nil {
		return nil, fmt.Errorf("failed to get incident severity: %w", sevErr)
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug)
	detailsText := fmt.Sprintf("*Incident Details*\n*Title:* %s\n*Severity:* %s\n*Status:* %s\n*Web:* %s",
		inc.Title, sev.Name, "OPEN", webLink)

	detailsTextBlock := &slack.TextBlockObject{
		Type: slack.MarkdownType,
		Text: detailsText,
	}
	return []slack.Block{
		slack.NewSectionBlock(detailsTextBlock, nil, nil),
	}, nil
}

func (h *incidentChatEventHandler) updateIncidentChannelDetailsMessage(ctx context.Context, inc *ent.Incident) error {
	msgBlocks, blocksErr := h.makeIncidentDetailsMessageBlocks(ctx, inc)
	if blocksErr != nil {
		return fmt.Errorf("making message blocks: %w", blocksErr)
	}

	return h.chat.withClient(ctx, func(client *slack.Client) error {
		pins, _, pinsErr := client.ListPinsContext(ctx, inc.ChatChannelID)
		if pinsErr != nil {
			return fmt.Errorf("failed to list pins: %w", pinsErr)
		}
		var existingMsgTs string
		for _, pin := range pins {
			if pin.Message != nil {
				if pin.Message.Text != "" && strings.HasPrefix(pin.Message.Text, "*Incident Details*") {
					existingMsgTs = pin.Message.Timestamp
					break
				}
			}
		}

		msgOpts := slack.MsgOptionBlocks(msgBlocks...)
		if existingMsgTs != "" {
			_, _, _, updateErr := client.UpdateMessageContext(ctx, inc.ChatChannelID, existingMsgTs, msgOpts)
			if updateErr != nil {
				return fmt.Errorf("update message: %w", updateErr)
			}
			return nil
		}

		_, msgTs, postErr := client.PostMessageContext(ctx, inc.ChatChannelID, msgOpts)
		if postErr != nil {
			return fmt.Errorf("post message: %w", postErr)
		}
		pinErr := client.AddPinContext(ctx, inc.ChatChannelID, slack.ItemRef{
			Channel:   inc.ChatChannelID,
			Timestamp: msgTs,
		})
		if pinErr != nil {
			return fmt.Errorf("pin message: %w", pinErr)
		}

		return nil
	})
}

func (h *incidentChatEventHandler) ensureIncidentChannelUsersAdded(ctx context.Context, inc *ent.Incident) error {

	return nil
}

func (h *incidentChatEventHandler) updateIncidentChannelInfo(ctx context.Context, inc *ent.Incident) error {
	sev, sevErr := h.chat.incidents.GetIncidentSeverity(ctx, inc.SeverityID)
	if sevErr != nil {
		return fmt.Errorf("failed to get incident severity: %w", sevErr)
	}

	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug)
	topic := fmt.Sprintf("[%s] %s", sev.Name, inc.Title)

	return h.chat.withClient(ctx, func(client *slack.Client) error {
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
