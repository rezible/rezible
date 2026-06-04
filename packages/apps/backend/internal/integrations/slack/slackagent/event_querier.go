package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
)

func (i *Integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	return newEventQuerier(i.makeInstalledIntegration(intg))
}

type eventQuerier struct {
	ii     *InstalledIntegration
	cfg    *slackintegration.InstallationConfig
	client *slack.Client
}

func newEventQuerier(ii *InstalledIntegration) (*eventQuerier, error) {
	cfg, cfgErr := slackintegration.DecodeInstallationConfig(ii.intg)
	if cfgErr != nil {
		return nil, cfgErr
	}
	return &eventQuerier{ii: ii, cfg: cfg, client: slack.New(cfg.AccessToken)}, nil
}

func (q *eventQuerier) Integration() *ent.Integration {
	return q.ii.intg
}

func (q *eventQuerier) QueryProviderEvents(ctx context.Context, cursors map[string]string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if usersCursor, ok := integrations.GetSourceQueryCursor(cursors, sourceUsers); ok {
			for ev, err := range q.pullUserObservedEvents(ctx, usersCursor) {
				if !yield(ev, err) {
					return
				}
			}
		}
	}
}

type userObservedPayload struct {
	Name      string         `json:"name"`
	SlackID   string         `json:"slack_id"`
	Email     string         `json:"email"`
	Timezone  string         `json:"timezone,omitempty"`
	UpdatedAt slack.JSONTime `json:"updated_at"`
}

func (q *eventQuerier) makeUserObservedPayload(u slack.User) ([]byte, error) {
	payload := userObservedPayload{
		Name:      u.Name,
		Email:     u.Profile.Email,
		SlackID:   u.ID,
		Timezone:  u.TZ,
		UpdatedAt: u.Updated,
	}
	return json.Marshal(payload)
}

func (q *eventQuerier) pullUserObservedEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	var teamId string
	if q.cfg.Team != nil {
		teamId = q.cfg.Team.Id
	}
	pullEvents := func() ([]rez.ProviderEvent, error) {
		slackUsers, getErr := q.client.GetUsersContext(ctx,
			slack.GetUsersOptionPresence(false),
			slack.GetUsersOptionTeamID(teamId))
		if getErr != nil {
			return nil, fmt.Errorf("slack get users err: %w", getErr)
		}

		var events []rez.ProviderEvent
		for _, u := range slackUsers {
			if u.IsBot || u.ID == "USLACKBOT" {
				continue
			}

			payload, payloadErr := q.makeUserObservedPayload(u)
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
		events, eventsErr := pullEvents()
		if eventsErr != nil {
			yield(nil, eventsErr)
			return
		}
		for _, event := range events {
			if !yield(&rez.ProviderEventQueryResult{Event: event, SourceCursorAfter: nil}, nil) {
				return
			}
		}
	}
}
