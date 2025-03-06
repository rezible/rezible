package slack

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
	"iter"
)

type DataProvider struct {
	client *slack.Client
}

// implements data provider interfaces
var (
	_ rez.UserDataProvider = (*DataProvider)(nil)
	_ rez.TeamDataProvider = (*DataProvider)(nil)
)

type DataProviderConfig struct {
	BotApiKey string `json:"bot_api_key"`
}

func NewDataProvider(cfg DataProviderConfig) (*DataProvider, error) {
	client := slack.New(cfg.BotApiKey)
	return &DataProvider{client: client}, nil
}

var (
	userDataMapping = ent.User{
		Name:     "y",
		Email:    "y",
		ChatID:   "y",
		Timezone: "y",
	}

	teamDataMapping = ent.Team{
		Name: "y",
	}
)

func (p *DataProvider) UserDataMapping() *ent.User {
	return &userDataMapping
}

func (p *DataProvider) TeamDataMapping() *ent.Team {
	return &teamDataMapping
}

func (p *DataProvider) pullTeams(ctx context.Context) iter.Seq2[*slack.Team, error] {
	var cursor string
	return func(yield func(*slack.Team, error) bool) {
		for cursor != "" {
			params := slack.ListTeamsParameters{Limit: 20, Cursor: cursor}
			teams, newCursor, listErr := p.client.ListTeamsContext(ctx, params)
			if listErr != nil {
				yield(nil, listErr)
				break
			}
			for _, team := range teams {
				yield(&team, nil)
			}
			cursor = newCursor
		}
	}
}

func (p *DataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	return func(yield func(*ent.Team, error) bool) {
		for team, teamsErr := range p.pullTeams(ctx) {
			if teamsErr != nil {
				yield(nil, fmt.Errorf("get teams: %w", teamsErr))
				return
			}

			mapped := &ent.Team{
				Name: team.Name,
			}
			yield(mapped, nil)
		}
	}
}

func (p *DataProvider) PullUsers(ctx context.Context) iter.Seq2[*ent.User, error] {
	return func(yield func(*ent.User, error) bool) {
		for team, teamsErr := range p.pullTeams(ctx) {
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
