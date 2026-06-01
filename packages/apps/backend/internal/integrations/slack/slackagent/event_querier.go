package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

func (i *Integration) MakeProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	return newEventQuerier(i.makeConfiguredIntegration(intg)), nil
}

type eventQuerier struct {
	ci     *ConfiguredIntegration
	client *slack.Client
}

func newEventQuerier(ci *ConfiguredIntegration) *eventQuerier {
	return &eventQuerier{ci: ci, client: slack.New(ci.accessToken())}
}

func (q *eventQuerier) Provider() string {
	return integrationName
}

func (q *eventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if usersCursor, ok := req.SourceCursors[sourceUsers]; ok || len(req.SourceCursors) == 0 {
			for ev, evErr := range q.pullUserObservedEvents(ctx, usersCursor) {
				if !yield(ev, evErr) {
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
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		slackUsers, getErr := q.client.GetUsersContext(ctx,
			slack.GetUsersOptionPresence(false),
			slack.GetUsersOptionTeamID(q.ci.teamId()))
		if getErr != nil {
			yield(nil, fmt.Errorf("slack get users err: %w", getErr))
			return
		}

		for _, u := range slackUsers {
			if u.IsBot || u.ID == "USLACKBOT" {
				continue
			}

			payload, payloadErr := q.makeUserObservedPayload(u)
			if payloadErr != nil {
				if !yield(nil, fmt.Errorf("make payload: %w", payloadErr)) {
					return
				}
				continue
			}

			res := &rez.ProviderEventQueryResult{
				Event: rez.ProviderEvent{
					Provider:           integrationName,
					ProviderSource:     sourceUsers,
					ProviderEventRef:   fmt.Sprintf("%s:%s:%s", q.ci.teamId(), u.ID, u.Updated),
					ProviderSubjectRef: fmt.Sprintf("slack:%s", u.ID),
					ReceivedAt:         time.Now(),
					Payload:            payload,
					ContentType:        "application/json",
				},
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}
