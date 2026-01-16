package slack

import (
	"context"
	"fmt"
	"iter"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type TeamDataProvider struct {
	client  *slack.Client
	teamIds []string
}

var _ rez.TeamDataProvider = (*TeamDataProvider)(nil)

func NewTeamDataProvider(intg *ent.Integration) (*TeamDataProvider, error) {
	cfg, cfgErr := decodeConfig(intg)
	if cfgErr != nil {
		return nil, cfgErr
	}
	teamIds := []string{cfg.Team.ID}
	return &TeamDataProvider{client: slack.New(cfg.AccessToken), teamIds: teamIds}, nil
}

var teamDataMapping = ent.Team{
	Name: "y",
	Edges: ent.TeamEdges{
		Users: []*ent.User{{
			ChatID: "y",
		}},
	},
}

func (p *TeamDataProvider) TeamDataMapping() *ent.Team {
	return &teamDataMapping
}

func (p *TeamDataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	return func(yield func(*ent.Team, error) bool) {
		for _, teamId := range p.teamIds {
			userGroups, userGroupsErr := p.client.GetUserGroupsContext(ctx,
				slack.GetUserGroupsOptionWithTeamID(teamId),
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
