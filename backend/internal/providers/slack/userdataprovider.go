package slack

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/twohundreds/rezible/ent"
	"iter"
)

type UserDataProvider struct {
	client *slack.Client
}

type UserDataProviderConfig struct {
	BotApiKey string `json:"bot_api_key"`
}

func NewUserDataProvider(cfg UserDataProviderConfig) (*UserDataProvider, error) {
	client := slack.New(cfg.BotApiKey)
	return &UserDataProvider{client: client}, nil
}

var (
	userDataMapping = ent.User{
		Name:     "y",
		Email:    "y",
		ChatID:   "y",
		Timezone: "y",
	}
)

func (p *UserDataProvider) UserDataMapping() *ent.User {
	return &userDataMapping
}

// TODO: integrate this properly
func listTeamIds(ctx context.Context, client *slack.Client) ([]string, error) {
	var ids []string
	var cursor string
	var listErr error
	for {
		var teams []slack.Team
		teams, cursor, listErr = client.ListTeamsContext(ctx, slack.ListTeamsParameters{Limit: 20, Cursor: cursor})
		for _, t := range teams {
			ids = append(ids, t.ID)
		}
		if cursor == "" || listErr != nil {
			break
		}
	}
	return ids, listErr
}

func (p *UserDataProvider) PullUsers(ctx context.Context) iter.Seq2[*ent.User, error] {
	return func(yield func(*ent.User, error) bool) {
		teamIds, teamsErr := listTeamIds(ctx, p.client)
		if teamsErr != nil {
			yield(nil, fmt.Errorf("get teams: %w", teamsErr))
			return
		}

		for _, teamId := range teamIds {
			slackUsers, getErr := p.client.GetUsersContext(ctx,
				slack.GetUsersOptionLimit(20),
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
				mapped := &ent.User{
					Name:     u.Profile.RealNameNormalized,
					Email:    u.Profile.Email,
					ChatID:   u.ID,
					Timezone: u.TZ,
				}
				yield(mapped, nil)
			}
		}
	}
}
