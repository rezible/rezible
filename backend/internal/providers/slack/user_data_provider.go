package slack

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
	"iter"
)

type UserDataProvider struct {
	client *slack.Client
}

var _ rez.UserDataProvider = (*UserDataProvider)(nil)

type UserDataProviderConfig struct {
	BotApiKey string `json:"bot_api_key"`
}

func NewUserDataProvider(cfg UserDataProviderConfig) (*UserDataProvider, error) {
	client := slack.New(cfg.BotApiKey)
	return &UserDataProvider{client: client}, nil
}

var (
	userDataMapping = &ent.User{
		Name:     "y",
		Email:    "y",
		ChatID:   "y",
		Timezone: "y",
	}
)

func (p *UserDataProvider) UserDataMapping() *ent.User {
	return userDataMapping
}

func (p *UserDataProvider) PullUsers(ctx context.Context) iter.Seq2[*ent.User, error] {
	return func(yield func(*ent.User, error) bool) {
		for team, teamsErr := range pullSlackTeams(ctx, p.client) {
			if teamsErr != nil {
				yield(nil, fmt.Errorf("get teams: %w", teamsErr))
				return
			}

			slackUsers, getErr := p.client.GetUsersContext(ctx,
				slack.GetUsersOptionPresence(false),
				slack.GetUsersOptionTeamID(team.ID))
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
