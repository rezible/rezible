package slackincidents

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/schema"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const (
	// maxTopicRunes is Slack's limit on a channel topic.
	maxTopicRunes = 250
	// maxChannelNameRunes is Slack's limit on a channel name.
	maxChannelNameRunes = 80
)

const bookmarkTitleIncident = "Open in Rezible"

// incidentChannel is an incident with the installation that runs its response channel.
type incidentChannel struct {
	inc         *ent.Incident
	incidentUrl *url.URL
	client      *slackintegration.ClientWrapper
	config      *slackintegration.InstallationConfig
	settings    *UserSettingsIncidents
}

// loadIncidentChannel reads the incident and the installation as they are now. It returns nil when there is no
// channel work to do: the incident was archived or has no declarer (it was observed, not run in Rezible),
// incident management is off, or the app is not installed.
func (a *App) loadIncidentChannel(ctx context.Context, incidentID uuid.UUID) (*incidentChannel, error) {
	inc, getErr := a.incidents.GetIncident(ctx, incident.ID(incidentID))
	if getErr != nil {
		if !ent.IsNotFound(getErr) {
			return nil, fmt.Errorf("get incident: %w", getErr)
		}
		archived, archivedErr := a.incidents.GetIncident(schema.IncludeArchived(ctx), incident.ID(incidentID))
		if archivedErr != nil {
			return nil, fmt.Errorf("get archived incident: %w", archivedErr)
		}
		if archived.ArchiveTime == nil {
			return nil, fmt.Errorf("incident %s not found", incidentID)
		}
		return nil, nil
	}
	if inc.DeclaredByID == nil {
		return nil, nil
	}
	enabled, enabledErr := a.orgs.IsIncidentManagementEnabled(ctx)
	if enabledErr != nil {
		return nil, fmt.Errorf("check incident management: %w", enabledErr)
	}
	if !enabled {
		return nil, nil
	}
	client, clientErr := a.getEnabledIntegrationClient(ctx)
	if clientErr != nil {
		return nil, fmt.Errorf("get slack incidents client: %w", clientErr)
	}
	if client == nil {
		return nil, nil
	}
	config, configErr := slackintegration.GetValidatedConfig(client.Integration().InstallationConfig)
	if configErr != nil {
		return nil, fmt.Errorf("slack incidents config: %w", configErr)
	}
	settings, settingsErr := readIncidentSettings(client.Integration())
	if settingsErr != nil {
		return nil, settingsErr
	}
	incUrl, incUrlErr := a.cfg.App.GetFrontendUrl("incidents", inc.Slug)
	if incUrlErr != nil {
		return nil, fmt.Errorf("incident url: %w", incUrlErr)
	}
	return &incidentChannel{inc: inc, incidentUrl: incUrl, client: client, config: config, settings: settings}, nil
}

// syncIncidentChannel requests the incident's channel when it has none, and otherwise brings the channel up
// to date. Each step runs even when another fails; the job fails at the end so River retries.
func (a *App) syncIncidentChannel(ctx context.Context, incidentID uuid.UUID) error {
	c, loadErr := a.loadIncidentChannel(ctx, incidentID)
	if loadErr != nil || c == nil {
		return loadErr
	}
	if c.inc.ChatChannelID == "" || c.inc.ChatDetailsMessageID == "" {
		args := createIncidentChannelJobArgs{IncidentId: c.inc.ID}
		if _, jobErr := a.jobs.Insert(ctx, args, nil); jobErr != nil {
			return fmt.Errorf("insert create incident channel job: %w", jobErr)
		}
	}
	if c.inc.ChatChannelID == "" {
		return nil
	}

	steps := []struct {
		name string
		run  func(context.Context) error
	}{
		{"update details message", c.updateDetailsMessage},
		{"update channel topic", c.updateTopic},
		{"add channel bookmark", c.ensureBookmark},
		{"invite people", c.inviteMembers},
	}
	var stepErrs []error
	for _, step := range steps {
		if stepErr := step.run(ctx); stepErr != nil {
			slog.WarnContext(ctx, "incident channel sync step failed",
				"step", step.name,
				"error", stepErr,
				"incident_id", c.inc.ID.String(),
				"channel_id", c.inc.ChatChannelID,
			)
			stepErrs = append(stepErrs, fmt.Errorf("%s: %w", step.name, stepErr))
		}
	}
	return errors.Join(stepErrs...)
}

// createIncidentChannel creates the response channel when there is none, and posts and pins its details
// message when there is none. Only this job, unique by incident, posts the details message, so concurrent
// syncs cannot post it twice. The channel is stored last: that publishes an incident update, and its sync
// edits the stored details message with anything that changed meanwhile. Then it announces the incident.
func (a *App) createIncidentChannel(ctx context.Context, incidentID uuid.UUID) error {
	c, loadErr := a.loadIncidentChannel(ctx, incidentID)
	if loadErr != nil || c == nil {
		return loadErr
	}
	created := c.inc.ChatChannelID == ""
	if created {
		channelID, createErr := c.createChannel(ctx)
		if createErr != nil {
			return createErr
		}
		c.inc.ChatChannelID = channelID
	}
	if c.inc.ChatDetailsMessageID == "" {
		messageID, postErr := c.postDetailsMessage(ctx)
		if postErr != nil {
			return postErr
		}
		if setErr := a.incidents.SetChatDetailsMessage(ctx, c.inc.ID, messageID); setErr != nil {
			return fmt.Errorf("set incident details message: %w", setErr)
		}
	}
	if !created {
		return nil
	}

	if setErr := a.incidents.SetChatChannel(ctx, c.inc.ID, c.inc.ChatChannelID); setErr != nil {
		return fmt.Errorf("set incident chat channel: %w", setErr)
	}
	if announceErr := c.postAnnouncement(ctx); announceErr != nil {
		slog.WarnContext(ctx, "failed to post incident announcement",
			"error", announceErr,
			"incident_id", c.inc.ID.String(),
		)
	}
	return nil
}

// createChannel creates the response channel, or finds the one a failed attempt created.
func (c *incidentChannel) createChannel(ctx context.Context) (string, error) {
	channelName := c.channelName()
	createParams := slack.CreateConversationParams{
		ChannelName: channelName,
		IsPrivate:   false,
	}
	channel, createErr := c.client.Client().CreateConversationContext(ctx, createParams)
	if createErr != nil {
		if !strings.Contains(createErr.Error(), "name_taken") {
			return "", fmt.Errorf("create channel: %w", createErr)
		}
		slog.WarnContext(ctx, "slack incident channel already exists, attempting relink",
			"error", createErr,
			"incident_id", c.inc.ID.String(),
			"channel_name", channelName,
		)
		existingChannel, lookupErr := c.findConversationByName(ctx, channelName)
		if lookupErr != nil {
			return "", fmt.Errorf("lookup existing channel after name_taken: %w", lookupErr)
		}
		if existingChannel == nil {
			return "", fmt.Errorf("slack reported name_taken but no conversation named %q was found", channelName)
		}
		channel = existingChannel
	}
	return channel.ID, nil
}

// postIncidentMilestone posts the milestone to the response channel, once. Before the channel exists there is
// nothing to post to: the details message created with the channel shows the latest update.
func (a *App) postIncidentMilestone(ctx context.Context, incidentID uuid.UUID, milestoneID uuid.UUID) error {
	c, loadErr := a.loadIncidentChannel(ctx, incidentID)
	if loadErr != nil || c == nil || c.inc.ChatChannelID == "" {
		return loadErr
	}
	ms, msErr := a.incidents.GetMilestone(ctx, milestoneID)
	if msErr != nil {
		return fmt.Errorf("get milestone: %w", msErr)
	}
	if ms.Metadata["msg_ts"] != "" {
		return nil
	}

	message := incidentMessages{inc: c.inc, incidentUrl: c.incidentUrl}
	msgOpts := message.milestoneMessageOptions(ms)
	msgTs, postErr := c.client.PostMessage(ctx, c.inc.ChatChannelID, msgOpts...)
	if postErr != nil {
		return fmt.Errorf("post milestone message: %w", postErr)
	}
	metadata := map[string]string{"msg_ts": msgTs}
	for key, value := range ms.Metadata {
		if key != "msg_ts" {
			metadata[key] = value
		}
	}
	if updateErr := a.incidents.SetMilestoneMetadata(ctx, ms.ID, metadata); updateErr != nil {
		return fmt.Errorf("store milestone message: %w", updateErr)
	}
	return nil
}

// channelName fills the pattern, in lowercase and within Slack's limit on channel names.
func (c *incidentChannel) channelName() string {
	name := strings.ReplaceAll(c.settings.ChannelNamePattern, "{id}", c.inc.ID.String())
	name = strings.ReplaceAll(name, "{slug}", c.inc.Slug)
	name = strings.ReplaceAll(name, "{title}", slug.Make(c.inc.Title))
	runes := []rune(strings.ToLower(name))
	return string(runes[:min(len(runes), maxChannelNameRunes)])
}

func (c *incidentChannel) findConversationByName(ctx context.Context, channelName string) (*slack.Channel, error) {
	params := &slack.GetConversationsParameters{
		ExcludeArchived: true,
		Limit:           200,
		Types:           []string{"public_channel"},
	}
	for {
		channels, cursor, listErr := c.client.Client().GetConversationsContext(ctx, params)
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

func (c *incidentChannel) postAnnouncement(ctx context.Context) error {
	channelID := c.settings.announcementChannel(c.config)
	if channelID == "" {
		return fmt.Errorf("no announcement channel")
	}
	message := incidentMessages{inc: c.inc, incidentUrl: c.incidentUrl}
	if _, postErr := c.client.PostMessage(ctx, channelID, message.announcementMessageOptions()...); postErr != nil {
		return fmt.Errorf("post announcement: %w", postErr)
	}
	return nil
}

// postDetailsMessage posts and pins the details message, and returns its timestamp. A failed pin is logged
// rather than returned, so a retry does not post the message again.
func (c *incidentChannel) postDetailsMessage(ctx context.Context) (string, error) {
	message := incidentMessages{inc: c.inc, incidentUrl: c.incidentUrl}
	msgTs, postErr := c.client.PostMessage(ctx, c.inc.ChatChannelID, message.detailsMessageOptions()...)
	if postErr != nil {
		return "", fmt.Errorf("post details message: %w", postErr)
	}
	pinRef := slack.ItemRef{Channel: c.inc.ChatChannelID, Timestamp: msgTs}
	if pinErr := c.client.Client().AddPinContext(ctx, c.inc.ChatChannelID, pinRef); pinErr != nil {
		slog.WarnContext(ctx, "failed to pin incident details message",
			"error", pinErr,
			"incident_id", c.inc.ID.String(),
		)
	}
	return msgTs, nil
}

// updateDetailsMessage edits the stored details message. Until one is stored, the channel job posts it.
func (c *incidentChannel) updateDetailsMessage(ctx context.Context) error {
	if c.inc.ChatDetailsMessageID == "" {
		return nil
	}
	message := incidentMessages{inc: c.inc, incidentUrl: c.incidentUrl}
	return c.client.UpdateMessage(ctx, c.inc.ChatChannelID, c.inc.ChatDetailsMessageID, message.detailsMessageOptions()...)
}

func (c *incidentChannel) updateTopic(ctx context.Context) error {
	client := c.client.Client()
	info, infoErr := client.GetConversationInfoContext(ctx, &slack.GetConversationInfoInput{ChannelID: c.inc.ChatChannelID})
	if infoErr != nil {
		return fmt.Errorf("get channel info: %w", infoErr)
	}
	topic := c.inc.Title
	if sev := c.inc.Edges.Severity; sev != nil {
		topic = fmt.Sprintf("[%s] %s", sev.Name, topic)
	}
	topic = slackintegration.TruncateRunes(topic, maxTopicRunes)
	escapedTopic := slackintegration.TruncateRunes(slackintegration.EscapeText(topic), maxTopicRunes)
	// The topic is sent escaped. Slack may return it as sent or unescaped, so either means it is current.
	if info.Topic.Value == escapedTopic || info.Topic.Value == topic {
		return nil
	}
	topic = escapedTopic
	if _, setErr := client.SetTopicOfConversationContext(ctx, c.inc.ChatChannelID, topic); setErr != nil {
		return fmt.Errorf("set channel topic: %w", setErr)
	}
	return nil
}

func (c *incidentChannel) ensureBookmark(ctx context.Context) error {
	client := c.client.Client()
	bookmarks, listErr := client.ListBookmarksContext(ctx, c.inc.ChatChannelID)
	if listErr != nil {
		return fmt.Errorf("list bookmarks: %w", listErr)
	}
	for _, bookmark := range bookmarks {
		if bookmark.Title == bookmarkTitleIncident {
			return nil
		}
	}
	params := slack.AddBookmarkParameters{
		Title: bookmarkTitleIncident,
		Link:  c.incidentUrl.String(),
		Type:  "link",
	}
	if _, addErr := client.AddBookmarkContext(ctx, c.inc.ChatChannelID, params); addErr != nil {
		return fmt.Errorf("add bookmark: %w", addErr)
	}
	return nil
}

// inviteMembers invites the declarer and everyone holding a role who isn't in the channel yet. Members are
// found in Slack by the chat ID the Slack agent's member sync sets.
func (c *incidentChannel) inviteMembers(ctx context.Context) error {
	people := make([]*ent.User, 0, len(c.inc.Edges.RoleAssignments)+1)
	if declarer := c.inc.Edges.DeclaredBy; declarer != nil {
		people = append(people, declarer)
	}
	for _, assignment := range c.inc.Edges.RoleAssignments {
		if assignment.Edges.User != nil {
			people = append(people, assignment.Edges.User)
		}
	}
	invite := mapset.NewSet[string]()
	for _, person := range people {
		if person.ChatID == "" {
			slog.InfoContext(ctx, "not inviting a member without a slack account to the incident channel",
				"incident_id", c.inc.ID.String(),
				"user_id", person.ID.String(),
			)
			continue
		}
		invite.Add(person.ChatID)
	}
	if invite.IsEmpty() {
		return nil
	}

	currentIds, idsErr := slackintegration.GetAllUsersInConversation(ctx, c.client, c.inc.ChatChannelID)
	if idsErr != nil {
		return fmt.Errorf("list channel members: %w", idsErr)
	}
	missing := invite.Difference(mapset.NewSet(currentIds...))
	if missing.IsEmpty() {
		return nil
	}
	if _, inviteErr := c.client.Client().InviteUsersToConversationContext(ctx, c.inc.ChatChannelID, missing.ToSlice()...); inviteErr != nil {
		return fmt.Errorf("invite to channel: %w", inviteErr)
	}
	return nil
}
