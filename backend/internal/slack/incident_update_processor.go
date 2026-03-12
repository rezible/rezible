package slack

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
)

type incidentUpdateProcessor struct {
	chat      *ChatService
	incidents rez.IncidentService
	messages  rez.MessageService
	inc       *ent.Incident
}

func newIncidentUpdateProcessor(chat *ChatService, services *rez.Services, inc *ent.Incident) *incidentUpdateProcessor {
	return &incidentUpdateProcessor{
		chat:      chat,
		incidents: services.Incidents,
		messages:  services.Messages,
		inc:       inc,
	}
}

func (p *incidentUpdateProcessor) processUpdate(ctx context.Context) error {
	if p.inc.ChatChannelID == "" {
		createCmdErr := p.messages.SendCommand(ctx, &cmdCreateIncidentChannel{IncidentId: p.inc.ID})
		if createCmdErr != nil {
			return fmt.Errorf("failed to send create incident channel command: %w", createCmdErr)
		}
		return nil
	}
	return p.updateIncidentChannel(ctx)
}

func (p *incidentUpdateProcessor) sendIncidentMilestoneMessage(ctx context.Context, milestoneId uuid.UUID) error {
	if p.inc.ChatChannelID == "" {
		// just created, don't send milestone update
		return nil
	}
	ms, msErr := p.incidents.GetIncidentMilestone(ctx, milestoneId)
	if msErr != nil {
		return fmt.Errorf("failed to get milestone: %w", msErr)
	}

	userTag := "Somebody"
	if msUser := ms.Edges.User; msUser != nil {
		if msUser.ChatID != "" {
			userTag = fmt.Sprintf("<@%s>", msUser.ChatID)
		} else {
			userTag = msUser.Name
		}
	}

	verb := "updated"
	label := strings.ToUpper(ms.Kind.String())
	switch ms.Kind {
	case im.KindImpact:
		verb = "reported"
	case im.KindMitigation:
		verb = "marked"
	case im.KindResolution:
		verb = "resolved"
	}

	bodyText := fmt.Sprintf("%s %s the incident as *%s*", userTag, verb, label)
	blocks := []slack.Block{
		slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", bodyText, false, false), nil, nil),
	}
	if ms.Description != "" {
		blocks = append(blocks, slack.NewContextBlock(
			"",
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("Note: %s", ms.Description), false, false),
		))
	}

	msgTs, msgErr := p.chat.postMessage(ctx, p.inc.ChatChannelID, slack.MsgOptionBlocks(blocks...))
	if msgErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", msgErr)
	}
	ms.Metadata["msg_ts"] = msgTs
	if updateErr := ms.Update().SetMetadata(ms.Metadata).Exec(ctx); updateErr != nil {
		return fmt.Errorf("failed to update metadata: %w", updateErr)
	}

	return nil
}

func formatIncidentChannelName(pattern string, inc *ent.Incident) string {
	if pattern == "" {
		pattern = "incident-{slug}"
	}
	name := strings.ReplaceAll(pattern, "{slug}", inc.Slug)
	name = strings.ReplaceAll(name, "{id}", inc.ID.String())
	name = strings.ReplaceAll(name, "{title}", slug.Make(inc.Title))
	return name
}

func isSlackChannelNameTakenError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "name_taken")
}

func (p *incidentUpdateProcessor) getIncidentChannelName() (string, error) {
	return formatIncidentChannelName(p.chat.ci.IncidentDefaults().ChannelNamePattern, p.inc), nil
}

func (p *incidentUpdateProcessor) findConversationByName(ctx context.Context, client *slack.Client, channelName string) (*slack.Channel, error) {
	params := &slack.GetConversationsParameters{
		ExcludeArchived: true,
		Limit:           200,
		Types:           []string{"public_channel"},
	}
	for {
		channels, cursor, listErr := client.GetConversationsContext(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("list conversations: %w", listErr)
		}
		for _, channel := range channels {
			if channel.Name == channelName || channel.NameNormalized == channelName {
				return &channel, nil
			}
		}
		if cursor == "" {
			return nil, nil
		}
		params.Cursor = cursor
	}
}

func (p *incidentUpdateProcessor) linkIncidentChannel(ctx context.Context, channelID string) error {
	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		m.SetChatChannelID(channelID)
		return nil
	}
	if _, updateErr := p.incidents.Set(ctx, p.inc.ID, setFn); updateErr != nil {
		return fmt.Errorf("set incident chatChannelID: %w", updateErr)
	}
	p.inc.ChatChannelID = channelID
	return nil
}

func (p *incidentUpdateProcessor) createIncidentChannel(ctx context.Context) error {
	channelName, nameErr := p.getIncidentChannelName()
	if nameErr != nil {
		return fmt.Errorf("resolve incident channel name: %w", nameErr)
	}

	createParams := slack.CreateConversationParams{
		ChannelName: channelName,
		TeamID:      "",
		IsPrivate:   false,
	}

	channel, createErr := p.chat.client.CreateConversationContext(ctx, createParams)
	if createErr != nil {
		if isSlackChannelNameTakenError(createErr) {
			log.Warn().
				Err(createErr).
				Str("incident_id", p.inc.ID.String()).
				Str("channel_name", channelName).
				Msg("slack incident channel already exists, attempting relink")

			existingChannel, lookupErr := p.findConversationByName(ctx, p.chat.client, channelName)
			if lookupErr != nil {
				return fmt.Errorf("lookup existing channel after name_taken: %w", lookupErr)
			}
			if existingChannel == nil {
				return fmt.Errorf("slack reported name_taken but no conversation named %q was found", channelName)
			}
			channel = existingChannel
		} else {
			return fmt.Errorf("create channel: %w", createErr)
		}
	}

	if linkErr := p.linkIncidentChannel(ctx, channel.ID); linkErr != nil {
		log.Error().
			Err(linkErr).
			Str("incident_id", p.inc.ID.String()).
			Str("channel_id", channel.ID).
			Str("channel_name", channelName).
			Msg("failed to persist slack incident channel link")
		return linkErr
	}

	if msgErr := p.sendUserCreatedChannelMessage(ctx); msgErr != nil {
		log.Warn().
			Err(msgErr).
			Str("incident_id", p.inc.ID.String()).
			Str("channel_id", p.inc.ChatChannelID).
			Msg("failed to send user incident creation message")
	}

	if annoErr := p.postIncidentAnnouncement(ctx); annoErr != nil {
		log.Warn().
			Err(annoErr).
			Str("incident_id", p.inc.ID.String()).
			Str("channel_id", p.inc.ChatChannelID).
			Msg("failed to post incident announcement")
	}

	return nil
}

func (p *incidentUpdateProcessor) getSlackIncidentCreateMilestone(ctx context.Context) (*ent.IncidentMilestone, error) {
	msQuery := p.inc.QueryMilestones().Where(im.KindEQ(im.KindOpened))
	ms, msErr := msQuery.First(ctx)
	if msErr != nil && !ent.IsNotFound(msErr) {
		return nil, fmt.Errorf("query milestones: %w", msErr)
	}
	return ms, nil
}

func (p *incidentUpdateProcessor) sendUserCreatedChannelMessage(ctx context.Context) error {
	ms, msErr := p.getSlackIncidentCreateMilestone(ctx)
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
	msgText := fmt.Sprintf("Incident created: <#%s>", p.inc.ChatChannelID)

	_, sendErr := p.chat.postEphemeralMessage(ctx, channelId, userId, slack.MsgOptionText(msgText, false))
	if sendErr != nil {
		return fmt.Errorf("failed to send confirmation message: %w", sendErr)
	}
	return nil
}

func (p *incidentUpdateProcessor) getIncidentAnnouncementChannelId() (string, error) {
	announcementChannelID := p.chat.ci.IncidentDefaults().AnnouncementChannelID
	if announcementChannelID != "" {
		return announcementChannelID, nil
	}
	if cfg, cfgErr := decodeConfig(p.chat.ci.RawConfig()); cfgErr == nil && cfg.WebhookChannelId != "" {
		return cfg.WebhookChannelId, nil
	}
	return "", errors.New("no announcementChannelId configured")
}

func (p *incidentUpdateProcessor) postIncidentAnnouncement(ctx context.Context) error {
	announcementChannelId, chanErr := p.getIncidentAnnouncementChannelId()
	if chanErr != nil {
		return fmt.Errorf("failed to get announcement channel: %w", chanErr)
	}

	builder := newIncidentAnnouncementMessageBuilder(p.inc)

	_, postErr := p.chat.postMessage(ctx, announcementChannelId, slack.MsgOptionBlocks(builder.build()...))
	if postErr != nil {
		return fmt.Errorf("failed to post announcement message: %w", postErr)
	}

	return nil
}

func (p *incidentUpdateProcessor) updateIncidentChannel(ctx context.Context) error {
	if detailsErr := p.updateIncidentChannelPinnedDetailsMessage(ctx); detailsErr != nil {
		log.Warn().Err(detailsErr).Msg("failed to update incident details message")
	}

	if topicErr := p.updateIncidentChannelTopic(ctx); topicErr != nil {
		log.Warn().Err(topicErr).Msg("failed to update incident channel topic")
	}

	conferenceAdded, bookmarksErr := p.ensureIncidentChannelBookmarks(ctx)
	if bookmarksErr != nil {
		log.Warn().Err(bookmarksErr).Msg("failed to update incident channel bookmarks")
	}
	if conferenceAdded {
		if msgErr := p.postIncidentConferenceMessage(ctx); msgErr != nil {
			log.Warn().Err(msgErr).Msg("failed to post incident conference message")
		}
	}

	if usersErr := p.ensureIncidentChannelUsersAdded(ctx); usersErr != nil {
		log.Warn().Err(usersErr).Msg("failed to add users to incident channel")
	}

	return nil
}

func (p *incidentUpdateProcessor) updateIncidentChannelPinnedDetailsMessage(ctx context.Context) error {
	builder := newIncidentDetailsMessageBuilder(p.inc)

	pins, _, pinsErr := p.chat.client.ListPinsContext(ctx, p.inc.ChatChannelID)
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
		_, _, _, updateErr := p.chat.client.UpdateMessageContext(ctx, p.inc.ChatChannelID, existingMsgTs, msgOpts)
		if updateErr != nil {
			return fmt.Errorf("update message: %w", updateErr)
		}
		return nil
	}

	_, msgTs, postErr := p.chat.client.PostMessageContext(ctx, p.inc.ChatChannelID, msgOpts)
	if postErr != nil {
		return fmt.Errorf("post message: %w", postErr)
	}
	pinItemRef := slack.ItemRef{
		Channel:   p.inc.ChatChannelID,
		Timestamp: msgTs,
	}
	if pinErr := p.chat.client.AddPinContext(ctx, p.inc.ChatChannelID, pinItemRef); pinErr != nil {
		return fmt.Errorf("pin message: %w", pinErr)
	}

	return nil
}

func (p *incidentUpdateProcessor) updateIncidentChannelTopic(ctx context.Context) error {
	info, infoErr := p.chat.client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{
		ChannelID:     p.inc.ChatChannelID,
		IncludeLocale: true,
	})
	if infoErr != nil {
		return fmt.Errorf("failed to get current channel info: %w", infoErr)
	}

	topic := fmt.Sprintf("[%s] %s", p.inc.Edges.Severity.Name, p.inc.Title)
	if info.Topic.Value != topic {
		_, setErr := p.chat.client.SetTopicOfConversationContext(ctx, p.inc.ChatChannelID, topic)
		if setErr != nil {
			return fmt.Errorf("failed to set channel topic: %w", setErr)
		}
	}

	return nil
}

func (p *incidentUpdateProcessor) ensureIncidentChannelBookmarks(ctx context.Context) (bool, error) {
	bookmarks, listErr := p.chat.client.ListBookmarksContext(ctx, p.inc.ChatChannelID)
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
		_, addErr := p.chat.client.AddBookmark(p.inc.ChatChannelID, slack.AddBookmarkParameters{
			Title: detailsTitle,
			Link:  fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), p.inc.Slug),
			Type:  "link",
		})
		if addErr != nil {
			return false, fmt.Errorf("failed to add bookmark: %w", addErr)
		}
	}

	conferenceUpdated := false
	primaryConf := p.inc.Edges.GetPrimaryVideoConference()
	if primaryConf != nil {
		if confBookmarkIndex == -1 {
			_, addErr := p.chat.client.AddBookmark(p.inc.ChatChannelID, slack.AddBookmarkParameters{
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
			_, editErr := p.chat.client.EditBookmark(p.inc.ChatChannelID, bm.ID, slack.EditBookmarkParameters{
				Link: primaryConf.JoinURL,
			})
			if editErr != nil {
				return false, fmt.Errorf("failed to edit conference bookmark: %w", editErr)
			}
		}
	}
	return conferenceUpdated, nil
}

func (p *incidentUpdateProcessor) postIncidentConferenceMessage(ctx context.Context) error {
	primaryConf := p.inc.Edges.GetPrimaryVideoConference()
	if primaryConf == nil {
		return nil
	}
	textBlock := slack.NewTextBlockObject(
		slack.MarkdownType,
		fmt.Sprintf(":video_camera: Incident video conference: %s", primaryConf.JoinURL),
		false,
		false,
	)
	_, msgErr := p.chat.postMessage(ctx, p.inc.ChatChannelID,
		slack.MsgOptionBlocks(slack.NewSectionBlock(textBlock, nil, nil)))
	return msgErr
}

func (p *incidentUpdateProcessor) ensureIncidentChannelUsersAdded(ctx context.Context) error {
	currIds, idsErr := getAllUsersInConversation(ctx, p.chat.client, p.inc.ChatChannelID)
	if idsErr != nil {
		return fmt.Errorf("failed to get current users in conversation: %w", idsErr)
	}
	excludeIds := mapset.NewSet(currIds...)
	addIds := mapset.NewSet[string]()

	ms, msErr := p.getSlackIncidentCreateMilestone(ctx)
	if msErr == nil && ms != nil && ms.Metadata != nil {
		if userId, userOk := ms.Metadata["user_id"]; userOk {
			addIds.Add(userId)
		}
	}
	for _, assignment := range p.inc.Edges.RoleAssignments {
		if assignment == nil || assignment.Edges.User == nil {
			continue
		}
		if assignment.Edges.User.ChatID == "" {
			log.Warn().
				Str("incident_id", p.inc.ID.String()).
				Str("user_id", assignment.Edges.User.ID.String()).
				Msg("skipping incident channel invite for user without slack mapping")
			continue
		}
		addIds.Add(assignment.Edges.User.ChatID)
	}

	missingIds := addIds.Difference(excludeIds)
	if missingIds.IsEmpty() {
		log.Debug().Msg("no users to add to incident channel")
		return nil
	}

	_, invErr := p.chat.client.InviteUsersToConversationContext(ctx, p.inc.ChatChannelID, missingIds.ToSlice()...)
	if invErr != nil {
		log.Error().Err(invErr).Msg("failed to add users to incident channel")
		return invErr
	}

	return nil
}
