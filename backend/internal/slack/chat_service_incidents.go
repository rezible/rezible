package slack

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
)

type incidentChatEventHandler struct {
	chat *ChatService
}

func newIncidentChatEventHandler(chat *ChatService) *incidentChatEventHandler {
	return &incidentChatEventHandler{chat: chat}
}

func (h *incidentChatEventHandler) registerHandlers() error {
	cmdsErr := h.chat.messages.AddCommandHandlers(
		rez.NewCommandHandler("SlackCreateIncidentChannel", h.createIncidentChannel),
		rez.NewCommandHandler("SlackSendIncidentMilestoneMessage", h.sendIncidentMilestoneMessage))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	eventsErr := h.chat.messages.AddEventHandlers(
		rez.NewEventHandler("SlackOnIncidentUpdate", h.onIncidentUpdate),
		rez.NewEventHandler("SlackOnIncidentMilestone", h.onIncidentMilestone))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}

	return nil
}

func (h *incidentChatEventHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	log.Debug().Msg("on incident update")

	inc, incErr := h.chat.incidents.Get(ctx, ev.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}

	// incident created
	if inc.ChatChannelID == "" {
		createCmdErr := h.chat.messages.SendCommand(ctx, &cmdCreateIncidentChannel{IncidentID: inc.ID})
		if createCmdErr != nil {
			return fmt.Errorf("failed to send create incident channel command: %w", createCmdErr)
		}
		return nil
	}

	// TODO: these should all be inserted as jobs
	return h.chat.withClient(ctx, func(client *slack.Client) error {
		if channelErr := h.updateIncidentChannelProperties(ctx, client, inc); channelErr != nil {
			log.Error().Err(channelErr).Msg("failed to update incident channel")
		}

		return nil
	})
}

func (h *incidentChatEventHandler) onIncidentMilestone(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	if !ev.Created {
		return nil
	}
	cmd := &cmdSendIncidentMilestoneMessage{
		IncidentId:  ev.IncidentId,
		MilestoneId: ev.MilestoneId,
	}
	if cmdErr := h.chat.messages.SendCommand(ctx, cmd); cmdErr != nil {
		return fmt.Errorf("failed to send incident milestone message command: %w", cmdErr)
	}
	return nil
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID `json:"iid"`
	MilestoneId uuid.UUID `json:"mid"`
}

func (h *incidentChatEventHandler) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	inc, incErr := h.chat.incidents.Get(ctx, ev.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}

	if inc.ChatChannelID == "" {
		return fmt.Errorf("no chat channel for incident: %s", ev.IncidentId)
	}

	ms, msErr := inc.QueryMilestones().Where(im.ID(ev.MilestoneId)).Only(ctx)
	if msErr != nil {
		return fmt.Errorf("failed to get milestone: %w", msErr)
	}

	return h.chat.withClient(ctx, func(client *slack.Client) error {
		blocks := []slack.Block{
			slack.NewSectionBlock(plainTextBlock(ms.Description), nil, nil),
		}

		metadata := slack.SlackMetadata{
			EventType: "incident_milestone",
			EventPayload: map[string]interface{}{
				"milestone_id": ev.MilestoneId,
			},
		}

		_, _, msgErr := client.PostMessageContext(ctx, inc.ChatChannelID,
			slack.MsgOptionBlocks(blocks...), slack.MsgOptionMetadata(metadata))
		if msgErr != nil {
			return fmt.Errorf("failed to post announcement message: %w", msgErr)
		}

		return nil
	})
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

	if msgErr := h.sendUserCreatedChannelMessage(ctx, inc); msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to send user incident creation message")
	}

	if annoErr := h.postIncidentAnnouncement(ctx, inc); annoErr != nil {
		log.Warn().Err(annoErr).Msg("failed to post incident announcement")
	}

	return nil
}

func (h *incidentChatEventHandler) getSlackIncidentCreateMilestone(ctx context.Context, inc *ent.Incident) (*ent.IncidentMilestone, error) {
	msQuery := inc.QueryMilestones().Where(im.KindEQ(im.KindOpened))
	ms, msErr := msQuery.First(ctx)
	if msErr != nil && !ent.IsNotFound(msErr) {
		return nil, fmt.Errorf("query milestones: %w", msErr)
	}
	return ms, nil
}

func (h *incidentChatEventHandler) sendUserCreatedChannelMessage(ctx context.Context, inc *ent.Incident) error {
	ms, msErr := h.getSlackIncidentCreateMilestone(ctx, inc)
	if msErr != nil {
		return msErr
	}
	if ms == nil || ms.Metadata == nil {
		return nil
	}
	userId, userOk := ms.Metadata["user_id"]
	channelId, channelOk := ms.Metadata["channel_id"]
	if !userOk || !channelOk {
		log.Warn().
			Interface("metadata", ms.Metadata).
			Msg("invalid slack incident declaration milestone metadata")
		return nil
	}
	// send message to user that created incident
	msgText := fmt.Sprintf("Incident created: <#%s>", inc.ChatChannelID)
	return h.chat.withClient(ctx, func(client *slack.Client) error {
		_, sendErr := client.PostEphemeralContext(ctx, channelId, userId, slack.MsgOptionText(msgText, false))
		if sendErr != nil {
			return fmt.Errorf("failed to send confirmation message: %w", sendErr)
		}
		return nil
	})
}

func (h *incidentChatEventHandler) postIncidentAnnouncement(ctx context.Context, inc *ent.Incident) error {
	announcementChannelId, chanErr := h.chat.getIncidentAnnouncementChannelId(ctx)
	if chanErr != nil {
		return fmt.Errorf("failed to get announcement channel: %w", chanErr)
	}

	builder := newIncidentAnnouncementMessageBuilder(inc)

	postErr := h.chat.sendMessage(ctx, announcementChannelId, slack.MsgOptionBlocks(builder.build()...))
	if postErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", postErr)
	}

	return nil
}

func (h *incidentChatEventHandler) updateIncidentChannelProperties(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
	if detailsErr := h.updateIncidentChannelPinnedDetailsMessage(ctx, client, inc); detailsErr != nil {
		log.Warn().Err(detailsErr).Msg("failed to update incident details message")
	}

	if topicErr := h.updateIncidentChannelTopic(ctx, client, inc); topicErr != nil {
		log.Warn().Err(topicErr).Msg("failed to update incident channel topic")
	}

	if bookmarksErr := h.ensureIncidentChannelBookmarks(ctx, client, inc); bookmarksErr != nil {
		log.Warn().Err(bookmarksErr).Msg("failed to update incident channel bookmarks")
	}

	if usersErr := h.ensureIncidentChannelUsersAdded(ctx, client, inc); usersErr != nil {
		log.Warn().Err(usersErr).Msg("failed to add users to incident channel")
	}

	return nil
}

func (h *incidentChatEventHandler) updateIncidentChannelPinnedDetailsMessage(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
	builder := newIncidentDetailsMessageBuilder(inc)

	pins, _, pinsErr := client.ListPinsContext(ctx, inc.ChatChannelID)
	if pinsErr != nil {
		return fmt.Errorf("failed to list pins: %w", pinsErr)
	}
	var existingMsgTs string
	for _, pin := range pins {
		if pin.Message != nil && builder.isDetailsMessage(pin.Message) {
			existingMsgTs = pin.Message.Timestamp
			break
		}
	}

	msgOpts := slack.MsgOptionBlocks(builder.build()...)

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
	pinItemRef := slack.ItemRef{
		Channel:   inc.ChatChannelID,
		Timestamp: msgTs,
	}
	if pinErr := client.AddPinContext(ctx, inc.ChatChannelID, pinItemRef); pinErr != nil {
		return fmt.Errorf("pin message: %w", pinErr)
	}

	return nil
}

func (h *incidentChatEventHandler) updateIncidentChannelTopic(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
	info, infoErr := client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID:     inc.ChatChannelID,
		IncludeLocale: true,
	})
	if infoErr != nil {
		return fmt.Errorf("failed to get current channel info: %w", infoErr)
	}

	topic := fmt.Sprintf("[%s] %s", inc.Edges.Severity.Name, inc.Title)
	if info.Topic.Value != topic {
		_, setErr := client.SetTopicOfConversationContext(ctx, inc.ChatChannelID, topic)
		if setErr != nil {
			return fmt.Errorf("failed to set channel topic: %w", setErr)
		}
	}

	return nil
}

func (h *incidentChatEventHandler) ensureIncidentChannelBookmarks(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
	bookmarks, listErr := client.ListBookmarksContext(ctx, inc.ChatChannelID)
	if listErr != nil {
		return fmt.Errorf("failed to list bookmarks: %w", listErr)
	}

	title := "View Incident Details"
	for _, bookmark := range bookmarks {
		// TODO: check more thoroughly?
		if bookmark.Title == title {
			return nil
		}
	}

	_, addErr := client.AddBookmark(inc.ChatChannelID, slack.AddBookmarkParameters{
		Title: title,
		Link:  fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug),
		Type:  "link",
	})
	if addErr != nil {
		return fmt.Errorf("failed to add bookmark: %w", addErr)
	}

	return nil
}

func (h *incidentChatEventHandler) ensureIncidentChannelUsersAdded(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
	currIds, idsErr := getAllUsersInConversation(ctx, client, inc.ChatChannelID)
	if idsErr != nil {
		return fmt.Errorf("failed to get current users in conversation: %w", idsErr)
	}
	excludeIds := mapset.NewSet(currIds...)
	addIds := mapset.NewSet[string]()

	ms, msErr := h.getSlackIncidentCreateMilestone(ctx, inc)
	if msErr == nil && ms != nil && ms.Metadata != nil {
		if userId, userOk := ms.Metadata["user_id"]; userOk {
			addIds.Add(userId)
		}
	}

	missingIds := addIds.Difference(excludeIds)
	if missingIds.IsEmpty() {
		log.Debug().Msg("no users to add to incident channel")
		return nil
	}

	_, invErr := client.InviteUsersToConversationContext(ctx, inc.ChatChannelID, missingIds.ToSlice()...)
	if invErr != nil {
		log.Error().Err(invErr).Msg("failed to add users to incident channel")
		return invErr
	}

	return nil
}
