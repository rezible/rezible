package slack

import (
	"context"
	"fmt"
	"iter"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type dataProvider struct {
	client  *slack.Client
	teamIds []string
}

var (
	_ rez.TeamDataProvider = (*dataProvider)(nil)
	_ rez.UserDataProvider = (*dataProvider)(nil)
)

func newDataProvider(intg *ent.Integration) (*dataProvider, error) {
	// TODO: handle nil integration (single tenant mode)
	cfg, cfgErr := decodeConfig(intg.Config)
	if cfgErr != nil {
		return nil, cfgErr
	}
	return &dataProvider{
		client:  slack.New(cfg.AccessToken),
		teamIds: []string{cfg.Team.ID},
	}, nil
}

func (i *integration) MakeUserDataProvider(ctx context.Context, intg *ent.Integration) (rez.UserDataProvider, error) {
	return newDataProvider(intg)
}

func (i *integration) MakeTeamDataProvider(ctx context.Context, intg *ent.Integration) (rez.TeamDataProvider, error) {
	return newDataProvider(intg)
}

var userDataMapping = &ent.User{
	Name:     "y",
	Email:    "y",
	ChatID:   "y",
	Timezone: "y",
}

func (p *dataProvider) UserDataMapping() *ent.User {
	return userDataMapping
}

func (p *dataProvider) PullUsers(ctx context.Context) iter.Seq2[*ent.User, error] {
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

var teamDataMapping = ent.Team{
	Name: "y",
	Edges: ent.TeamEdges{
		Users: []*ent.User{{
			ChatID: "y",
		}},
	},
}

func (p *dataProvider) TeamDataMapping() *ent.Team {
	return &teamDataMapping
}

func (p *dataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	return func(yield func(*ent.Team, error) bool) {
		for _, teamId := range p.teamIds {
			userGroups, userGroupsErr := p.client.GetUserGroupsContext(ctx,
				slack.GetUserGroupsOptionTeamID(teamId),
				slack.GetUserGroupsOptionIncludeUsers(true))
			if userGroupsErr != nil {
				yield(nil, fmt.Errorf("error getting user groups: %w", userGroupsErr))
				return
			}

			for _, userGroup := range userGroups {
				ugUsers := make([]*ent.User, len(userGroup.Users))
				for i, userId := range userGroup.Users {
					ugUsers[i] = &ent.User{ChatID: userId}
				}
				mapped := &ent.Team{
					ExternalID: userGroup.ID,
					Name:       userGroup.Name,
					Edges: ent.TeamEdges{
						Users: ugUsers,
					},
				}
				yield(mapped, nil)
			}
		}
	}
}
