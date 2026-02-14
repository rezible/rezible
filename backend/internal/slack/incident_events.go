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

type incidentEventHandler struct {
	msgs      rez.MessageService
	incidents rez.IncidentService
	svcLoader *serviceLoader
}

func newIncidentEventHandler(sl *serviceLoader, msgs rez.MessageService, incidents rez.IncidentService) (*incidentEventHandler, error) {
	h := &incidentEventHandler{svcLoader: sl, msgs: msgs, incidents: incidents}
	if hErr := h.registerHandlers(); hErr != nil {
		return nil, fmt.Errorf("registering message handlers: %w", hErr)
	}
	return h, nil
}

func (h *incidentEventHandler) registerHandlers() error {
	cmdsErr := h.msgs.AddCommandHandlers(
		rez.NewCommandHandler("SlackCreateIncidentChannel", h.createIncidentChannel),
		rez.NewCommandHandler("SlackSendIncidentMilestoneMessage", h.sendIncidentMilestoneMessage))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	eventsErr := h.msgs.AddEventHandlers(
		rez.NewEventHandler("SlackOnIncidentUpdate", h.onIncidentUpdate),
		rez.NewEventHandler("SlackOnIncidentMilestone", h.onIncidentMilestone))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}

	return nil
}

func (h *incidentEventHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	chat, loadChatErr := h.svcLoader.fromContext(ctx)
	if chat == nil {
		return loadChatErr
	}

	inc, incErr := h.incidents.Get(ctx, ev.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}

	// incident created
	if inc.ChatChannelID == "" {
		createCmdErr := h.msgs.SendCommand(ctx, &cmdCreateIncidentChannel{IncidentID: inc.ID})
		if createCmdErr != nil {
			return fmt.Errorf("failed to send create incident channel command: %w", createCmdErr)
		}
		return nil
	}

	channelErr := h.updateIncidentChannel(ctx, chat, inc)
	if channelErr != nil {
		log.Error().Err(channelErr).Msg("failed to update incident channel")
	}
	return nil
}

func (h *incidentEventHandler) onIncidentMilestone(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	if !ev.Created {
		return nil
	}
	chat, loadChatErr := h.svcLoader.fromContext(ctx)
	if chat == nil {
		return loadChatErr
	}
	return h.msgs.SendCommand(ctx, &cmdSendIncidentMilestoneMessage{
		IncidentId:  ev.IncidentId,
		MilestoneId: ev.MilestoneId,
	})
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID `json:"iid"`
	MilestoneId uuid.UUID `json:"mid"`
}

func (h *incidentEventHandler) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	chat, loadChatErr := h.svcLoader.fromContext(ctx)
	if chat == nil {
		return loadChatErr
	}

	inc, incErr := h.incidents.Get(ctx, ev.IncidentId)
	if incErr != nil {
		return fmt.Errorf("failed to get incident: %w", incErr)
	}

	if inc.ChatChannelID == "" {
		// just created, don't send milestone update
		return nil
	}

	ms, msErr := h.incidents.GetIncidentMilestone(ctx, ev.MilestoneId)
	if msErr != nil {
		return fmt.Errorf("failed to get milestone: %w", msErr)
	}

	userTag := "Someone"
	if msUser := ms.Edges.User; msUser != nil {
		if msUser.ChatID != "" {
			userTag = fmt.Sprintf("<@%s>", msUser.ChatID)
		} else {
			userTag = msUser.Name
		}
	}

	// TODO: format nicely
	milestoneText := ms.Kind.String()

	descText := ""
	if ms.Description != "" {
		descText = fmt.Sprintf(" with a note: \"%s\"", ms.Description)
	}

	text := fmt.Sprintf("%s marked incident as *%s*%s", userTag, milestoneText, descText)
	textBlock := slack.NewTextBlockObject("mrkdwn", text, false, false)
	blocks := []slack.Block{
		slack.NewSectionBlock(textBlock, nil, nil),
	}

	msgTs, msgErr := chat.postMessage(ctx, inc.ChatChannelID, slack.MsgOptionBlocks(blocks...))
	if msgErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", msgErr)
	}
	ms.Metadata["msg_ts"] = msgTs
	if updateErr := ms.Update().SetMetadata(ms.Metadata).Exec(ctx); updateErr != nil {
		return fmt.Errorf("failed to update metadata: %w", updateErr)
	}

	return nil
}

func getIncidentChannelName(inc *ent.Incident) string {
	return fmt.Sprintf("incident-%s", inc.Slug)
}

type cmdCreateIncidentChannel struct {
	IncidentID uuid.UUID
}

func (h *incidentEventHandler) createIncidentChannel(ctx context.Context, data *cmdCreateIncidentChannel) error {
	inc, incErr := h.incidents.Get(ctx, data.IncidentID)
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

	chat, chatErr := h.svcLoader.fromContext(ctx)
	if chatErr != nil {
		return fmt.Errorf("load chat service: %w", chatErr)
	}

	// TODO: check if channel exists first

	channel, createErr := chat.client.CreateConversationContext(ctx, createParams)
	if createErr != nil {
		return fmt.Errorf("create channel: %w", createErr)
	}

	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		m.SetChatChannelID(channel.ID)
		return nil
	}
	if _, updateErr := h.incidents.Set(ctx, inc.ID, setFn); updateErr != nil {
		return fmt.Errorf("set incident chatChannelID: %w", updateErr)
	}
	inc.ChatChannelID = channel.ID

	if msgErr := h.sendUserCreatedChannelMessage(ctx, chat, inc); msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to send user incident creation message")
	}

	if annoErr := h.postIncidentAnnouncement(ctx, chat, inc); annoErr != nil {
		log.Warn().Err(annoErr).Msg("failed to post incident announcement")
	}

	return nil
}

func (h *incidentEventHandler) getSlackIncidentCreateMilestone(ctx context.Context, inc *ent.Incident) (*ent.IncidentMilestone, error) {
	msQuery := inc.QueryMilestones().Where(im.KindEQ(im.KindOpened))
	ms, msErr := msQuery.First(ctx)
	if msErr != nil && !ent.IsNotFound(msErr) {
		return nil, fmt.Errorf("query milestones: %w", msErr)
	}
	return ms, nil
}

func (h *incidentEventHandler) sendUserCreatedChannelMessage(ctx context.Context, chat *ChatService, inc *ent.Incident) error {
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

	_, sendErr := chat.postEphemeralMessage(ctx, channelId, userId, slack.MsgOptionText(msgText, false))
	if sendErr != nil {
		return fmt.Errorf("failed to send confirmation message: %w", sendErr)
	}
	return nil
}

func (h *incidentEventHandler) getIncidentAnnouncementChannelId(ctx context.Context) (string, error) {
	// TODO: fetch from config
	announcementChannelId := "#incident"
	return announcementChannelId, nil
}

func (h *incidentEventHandler) postIncidentAnnouncement(ctx context.Context, chat *ChatService, inc *ent.Incident) error {
	announcementChannelId, chanErr := h.getIncidentAnnouncementChannelId(ctx)
	if chanErr != nil {
		return fmt.Errorf("failed to get announcement channel: %w", chanErr)
	}

	builder := newIncidentAnnouncementMessageBuilder(inc)

	_, postErr := chat.postMessage(ctx, announcementChannelId, slack.MsgOptionBlocks(builder.build()...))
	if postErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", postErr)
	}

	return nil
}

func (h *incidentEventHandler) updateIncidentChannel(ctx context.Context, chat *ChatService, inc *ent.Incident) error {
	client := chat.client
	if detailsErr := h.updateIncidentChannelPinnedDetailsMessage(ctx, client, inc); detailsErr != nil {
		log.Warn().Err(detailsErr).Msg("failed to update incident details message")
	}

	if topicErr := h.updateIncidentChannelTopic(ctx, client, inc); topicErr != nil {
		log.Warn().Err(topicErr).Msg("failed to update incident channel topic")
	}

	conferenceAdded, bookmarksErr := h.ensureIncidentChannelBookmarks(ctx, client, inc)
	if bookmarksErr != nil {
		log.Warn().Err(bookmarksErr).Msg("failed to update incident channel bookmarks")
	}
	if conferenceAdded {
		if msgErr := h.postIncidentConferenceMessage(ctx, chat, inc); msgErr != nil {
			log.Warn().Err(msgErr).Msg("failed to post incident conference message")
		}
	}

	if usersErr := h.ensureIncidentChannelUsersAdded(ctx, client, inc); usersErr != nil {
		log.Warn().Err(usersErr).Msg("failed to add users to incident channel")
	}

	return nil
}

func (h *incidentEventHandler) updateIncidentChannelPinnedDetailsMessage(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
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

func (h *incidentEventHandler) updateIncidentChannelTopic(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
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

func (h *incidentEventHandler) ensureIncidentChannelBookmarks(ctx context.Context, client *slack.Client, inc *ent.Incident) (bool, error) {
	bookmarks, listErr := client.ListBookmarksContext(ctx, inc.ChatChannelID)
	if listErr != nil {
		return false, fmt.Errorf("failed to list bookmarks: %w", listErr)
	}

	detailsTitle := "View Incident Details"
	conferenceTitle := "Join Video Conference"
	hasDetails := false
	confBookmarkIndex := -1
	for i, bookmark := range bookmarks {
		switch bookmark.Title {
		case detailsTitle:
			hasDetails = true
		case conferenceTitle:
			confBookmarkIndex = i
		}
	}

	if !hasDetails {
		_, addErr := client.AddBookmark(inc.ChatChannelID, slack.AddBookmarkParameters{
			Title: detailsTitle,
			Link:  fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), inc.Slug),
			Type:  "link",
		})
		if addErr != nil {
			return false, fmt.Errorf("failed to add bookmark: %w", addErr)
		}
	}

	conferenceUpdated := false
	primaryConf := inc.Edges.GetPrimaryVideoConference()
	if primaryConf != nil {
		if confBookmarkIndex == -1 {
			_, addErr := client.AddBookmark(inc.ChatChannelID, slack.AddBookmarkParameters{
				Title: conferenceTitle,
				Link:  primaryConf.JoinURL,
				Emoji: ":video_camera:",
				Type:  "link",
			})
			if addErr != nil {
				return false, fmt.Errorf("failed to add conference bookmark: %w", addErr)
			}
			conferenceUpdated = true
		} else if bm := bookmarks[confBookmarkIndex]; bm.Link != primaryConf.JoinURL {
			_, editErr := client.EditBookmark(inc.ChatChannelID, bm.ID, slack.EditBookmarkParameters{
				Link: primaryConf.JoinURL,
			})
			if editErr != nil {
				return false, fmt.Errorf("failed to edit conference bookmark: %w", editErr)
			}
		}
	}
	return conferenceUpdated, nil
}

func (h *incidentEventHandler) postIncidentConferenceMessage(ctx context.Context, chat *ChatService, inc *ent.Incident) error {
	primaryConf := inc.Edges.GetPrimaryVideoConference()
	if primaryConf == nil {
		return nil
	}
	textBlock := slack.NewTextBlockObject(
		slack.MarkdownType,
		fmt.Sprintf(":video_camera: Incident video conference: %s", primaryConf.JoinURL),
		false,
		false,
	)
	_, msgErr := chat.postMessage(ctx, inc.ChatChannelID, slack.MsgOptionBlocks(slack.NewSectionBlock(textBlock, nil, nil)))
	return msgErr
}

func (h *incidentEventHandler) ensureIncidentChannelUsersAdded(ctx context.Context, client *slack.Client, inc *ent.Incident) error {
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
