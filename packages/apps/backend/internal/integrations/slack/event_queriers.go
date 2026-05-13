package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
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

type userObservedPayload struct {
	SlackID   string         `json:"slack_id"`
	Email     string         `json:"email"`
	Timezone  string         `json:"timezone,omitempty"`
	UpdatedAt slack.JSONTime `json:"updated_at"`
}

func (q *userEventQuerier) PullEvents(ctx context.Context, req rez.ProviderEventQueryRequest) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	teamIds := []string{q.ci.teamId()}
	slog.Debug("pulling slack user events", "teamIds", teamIds)
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		for _, teamId := range teamIds {
			slackUsers, getErr := q.client.GetUsersContext(ctx,
				slack.GetUsersOptionPresence(false),
				slack.GetUsersOptionTeamID(teamId))
			if getErr != nil {
				yield(nil, fmt.Errorf("slack get users err: %w", getErr))
				return
			}

			for _, u := range slackUsers {
				if u.IsBot || u.ID == "USLACKBOT" {
					continue
				}

				payload := userObservedPayload{
					Email:     u.Profile.Email,
					SlackID:   u.ID,
					Timezone:  u.TZ,
					UpdatedAt: u.Updated,
				}
				body, marshalErr := json.Marshal(payload)
				if marshalErr != nil {
					if !yield(nil, fmt.Errorf("marshal payload: %w", marshalErr)) {
						return
					}
					continue
				}

				res := &rez.ProviderEventQueryResult{
					Event: rez.ProviderEvent{
						Provider:         integrationName,
						ProviderSource:   sourceUsers,
						ProviderEventRef: fmt.Sprintf("%s:%s:%s", teamId, u.ID, u.Updated),
						SubjectRef:       fmt.Sprintf("slack:%s", u.ID),
						ReceivedAt:       time.Now(),
						Payload:          body,
						ContentType:      "application/json",
					},
				}

				if !yield(res, nil) {
					return
				}
			}
		}
	}
}
