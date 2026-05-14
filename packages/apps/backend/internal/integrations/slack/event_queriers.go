package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/slack-go/slack"
)

type userEventQuerier struct {
	ci     *ConfiguredIntegration
	client *slack.Client
}

func newUserEventQuerier(ci *ConfiguredIntegration) *userEventQuerier {
	return &userEventQuerier{ci: ci, client: slack.New(ci.accessToken())}
}

func (q *userEventQuerier) Provider() string {
	return integrationName
}

func (q *userEventQuerier) ProviderSource() string {
	return "users"
}

func (q *userEventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return q.pullUserObservedEvents(ctx, req.CursorAfter)
}

type userObservedPayload struct {
	SlackID   string         `json:"slack_id"`
	Email     string         `json:"email"`
	Timezone  string         `json:"timezone,omitempty"`
	UpdatedAt slack.JSONTime `json:"updated_at"`
}

func (q *userEventQuerier) makeUserObservedPayload(u slack.User) ([]byte, error) {
	payload := userObservedPayload{
		Email:     u.Profile.Email,
		SlackID:   u.ID,
		Timezone:  u.TZ,
		UpdatedAt: u.Updated,
	}
	return json.Marshal(payload)
}

func (q *userEventQuerier) pullUserObservedEvents(ctx context.Context, cursor string) iter.Seq2[*rez.ProviderEventQueryResult, error] {
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
					Provider:         integrationName,
					ProviderSource:   sourceUsers,
					ProviderEventRef: fmt.Sprintf("%s:%s:%s", q.ci.teamId(), u.ID, u.Updated),
					SubjectRef:       fmt.Sprintf("slack:%s", u.ID),
					ReceivedAt:       time.Now(),
					Payload:          payload,
					ContentType:      "application/json",
				},
			}

			if !yield(res, nil) {
				return
			}
		}
	}
}
