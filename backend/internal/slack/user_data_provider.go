package slack

import (
	"context"
	"fmt"
	"iter"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type UserDataProvider struct {
	client  *slack.Client
	teamIds []string
}

var _ rez.UserDataProvider = (*UserDataProvider)(nil)

func NewUserDataProvider(cfg IntegrationConfigData) (*UserDataProvider, error) {
	var teamIds []string
	if cfg.Team != nil {
		teamIds = []string{cfg.Team.ID}
	}
	return &UserDataProvider{client: slack.New(cfg.AccessToken), teamIds: teamIds}, nil
}

var userDataMapping = &ent.User{
	Name:     "y",
	Email:    "y",
	ChatID:   "y",
	Timezone: "y",
}

func (p *UserDataProvider) UserDataMapping() *ent.User {
	return userDataMapping
}

func (p *UserDataProvider) PullUsers(ctx context.Context) iter.Seq2[*ent.User, error] {
	if p.teamIds == nil {
		teams, _, listErr := p.client.ListTeamsContext(ctx, slack.ListTeamsParameters{})
		if listErr == nil {
			p.teamIds = make([]string, len(teams))
			for i, team := range teams {
				p.teamIds[i] = team.ID
			}
		}
	}
	return func(yield func(*ent.User, error) bool) {
		for _, teamId := range p.teamIds {
			slackUsers, getErr := p.client.GetUsersContext(ctx,
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
				yield(&ent.User{
					Name:     u.Profile.RealNameNormalized,
					Email:    u.Profile.Email,
					ChatID:   u.ID,
					Timezone: u.TZ,
				}, nil)
			}
		}
	}
}
