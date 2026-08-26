package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/slack-go/slack"
)

func (i *Integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	ii, iiErr := i.makeInstalledIntegration(intg)
	if iiErr != nil {
		return nil, fmt.Errorf("make installed integration: %w", iiErr)
	}
	return &eventQuerier{ii: ii, client: slack.New(ii.config.AccessToken)}, nil
}

type eventQuerier struct {
	ii     *InstalledIntegration
	client *slack.Client
}

func (q *eventQuerier) Integration() *ent.Integration {
	return q.ii.intg
}

func (q *eventQuerier) QueryProviderEvents(ctx context.Context, cursors rez.ProviderEventQuerySourceCursors) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if usersCursor, ok := integrations.GetSourceQueryCursor(cursors, sourceUsers); ok {
			for ev, err := range q.pullUserObservedEvents(ctx, usersCursor) {
				if !yield(ev, err) {
					return
				}
			}
		}
		if teamsCursor, ok := integrations.GetSourceQueryCursor(cursors, sourceTeams); ok {
			for ev, err := range q.pullTeamObservedEvents(ctx, teamsCursor) {
				if !yield(ev, err) {
					return
				}
			}
		}
	}
}

func makeUserObservedPayload(u slack.User) userObservedPayload {
	return userObservedPayload{
		SlackID:   u.ID,
		Name:      u.Name,
		Email:     u.Profile.Email,
		Timezone:  u.TZ,
		UpdatedAt: u.Updated,
	}
}

type userObservedPayload struct {
	Name      string         `json:"name"`
	SlackID   string         `json:"slack_id"`
	Email     string         `json:"email"`
	Timezone  string         `json:"timezone,omitempty"`
	UpdatedAt slack.JSONTime `json:"updated_at"`
}

func (p userObservedPayload) makeSubjectAttributes() projections.UserSubjectAttributes {
	return projections.UserSubjectAttributes{
		ExternalRef: p.SlackID,
		Name:        p.Name,
		Email:       p.Email,
		ChatId:      p.SlackID,
		Timezone:    p.Timezone,
	}
}

func makeTeamObservedPayload(g slack.UserGroup) teamObservedPayload {
	chatChannelID := ""
	if len(g.Prefs.Channels) > 0 {
		chatChannelID = g.Prefs.Channels[0]
	}
	return teamObservedPayload{
		ExternalRef:   g.ID,
		Name:          g.Name,
		Slug:          g.Handle,
		ChatChannelID: chatChannelID,
		UpdatedAt:     g.DateUpdate,
		Deleted:       !g.DateDelete.Time().IsZero(),
	}
}

type teamObservedPayload struct {
	ExternalRef   string         `json:"external_ref"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	ChatChannelID string         `json:"chat_channel_id,omitempty"`
	UpdatedAt     slack.JSONTime `json:"updated_at"`
	Deleted       bool           `json:"deleted"`
}

func (p teamObservedPayload) makeSubjectAttributes() projections.TeamSubjectAttributes {
	return projections.TeamSubjectAttributes{
		ExternalRef:   p.ExternalRef,
		Name:          p.Name,
		Slug:          p.Slug,
		ChatChannelId: p.ChatChannelID,
	}
}

type teamMembershipObservedPayload struct {
	Team teamObservedPayload `json:"team"`
	User userObservedPayload `json:"user"`
}

func (q *eventQuerier) pullUserObservedEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	var teamId string
	if q.ii.config.Team != nil {
		teamId = q.ii.config.Team.Id
	}
	pullUserEvents := func() ([]rez.ProviderEvent, error) {
		slackUsers, getErr := q.client.GetUsersContext(ctx,
			slack.GetUsersOptionPresence(false),
			slack.GetUsersOptionTeamID(teamId))
		if getErr != nil {
			return nil, fmt.Errorf("slack get users err: %w", getErr)
		}

		var events []rez.ProviderEvent
		for _, u := range slackUsers {
			if u.IsBot || u.ID == "USLACKBOT" || u.Profile.Email == "" {
				continue
			}

			payload, payloadErr := json.Marshal(makeUserObservedPayload(u))
			if payloadErr != nil {
				return nil, fmt.Errorf("make payload: %w", payloadErr)
			}

			events = append(events, rez.ProviderEvent{
				Provider:           integrationName,
				ProviderSource:     sourceUsers,
				ProviderEventRef:   fmt.Sprintf("%s:%s:%s", teamId, u.ID, u.Updated),
				ProviderSubjectRef: fmt.Sprintf("slack:%s", u.ID),
				ReceivedAt:         time.Now(),
				Payload:            payload,
				ContentType:        "application/json",
			})
		}

		return events, nil
	}
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		events, eventsErr := pullUserEvents()
		if eventsErr != nil {
			yield(nil, eventsErr)
			return
		}
		nextCursor := fmt.Sprintf("%d", time.Now().UTC().Unix())
		for _, event := range events {
			if !yield(&rez.ProviderEventQueryResult{Event: event, SourceCursorAfter: &nextCursor}, nil) {
				return
			}
		}
	}
}

func (q *eventQuerier) pullTeamObservedEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		var teamID string
		if q.ii.config.Team != nil {
			teamID = q.ii.config.Team.Id
		}
		users, usersErr := q.client.GetUsersContext(ctx,
			slack.GetUsersOptionPresence(false),
			slack.GetUsersOptionTeamID(teamID))
		if usersErr != nil {
			yield(nil, fmt.Errorf("slack get users: %w", usersErr))
			return
		}
		usersByID := make(map[string]slack.User, len(users))
		for _, user := range users {
			usersByID[user.ID] = user
		}

		groups, groupsErr := q.client.GetUserGroupsContext(ctx,
			slack.GetUserGroupsOptionTeamID(teamID),
			slack.GetUserGroupsOptionIncludeDisabled(true),
			slack.GetUserGroupsOptionIncludeUsers(true))
		if groupsErr != nil {
			yield(nil, fmt.Errorf("slack get user groups: %w", groupsErr))
			return
		}

		nextCursor := fmt.Sprintf("%d", time.Now().UTC().Unix())
		for _, group := range groups {
			teamPayload := makeTeamObservedPayload(group)

			teamPayloadBytes, marshalTeamErr := json.Marshal(teamPayload)
			if marshalTeamErr != nil {
				yield(nil, fmt.Errorf("make team payload: %w", marshalTeamErr))
				return
			}

			teamEvent := rez.ProviderEvent{
				Provider:           integrationName,
				ProviderSource:     sourceTeams,
				ProviderEventRef:   fmt.Sprintf("%s:%s:%d", teamID, group.ID, group.DateUpdate),
				ProviderSubjectRef: group.ID,
				ReceivedAt:         time.Now().UTC(),
				Payload:            teamPayloadBytes,
				ContentType:        "application/json",
			}
			if !yield(&rez.ProviderEventQueryResult{Event: teamEvent, SourceCursorAfter: &nextCursor}, nil) {
				return
			}
			if teamPayload.Deleted {
				continue
			}
			for _, userID := range group.Users {
				usr, exists := usersByID[userID]
				if !exists || usr.IsBot || usr.Profile.Email == "" {
					continue
				}
				membershipPayload := teamMembershipObservedPayload{
					Team: teamPayload,
					User: makeUserObservedPayload(usr),
				}
				membershipPayloadBytes, marshalMembershipErr := json.Marshal(membershipPayload)
				if marshalMembershipErr != nil {
					yield(nil, fmt.Errorf("marshal membership payload: %w", marshalMembershipErr))
					return
				}
				membershipEvent := rez.ProviderEvent{
					Provider:           integrationName,
					ProviderSource:     sourceTeamMemberships,
					ProviderEventRef:   fmt.Sprintf("%s:%s:%s:%d", teamID, group.ID, userID, group.DateUpdate),
					ProviderSubjectRef: fmt.Sprintf("slack:%s:%s", group.ID, userID),
					ReceivedAt:         time.Now().UTC(),
					Payload:            membershipPayloadBytes,
					ContentType:        "application/json",
				}
				if !yield(&rez.ProviderEventQueryResult{Event: membershipEvent, SourceCursorAfter: &nextCursor}, nil) {
					return
				}
			}
		}
	}
}
