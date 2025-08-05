package slack

import (
	"context"
	"fmt"
	"iter"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezslack "github.com/rezible/rezible/internal/slack"
)

type TeamDataProvider struct {
	client *slack.Client
}

var _ rez.TeamDataProvider = (*TeamDataProvider)(nil)

type TeamDataProviderConfig struct{}

func NewTeamDataProvider(cfg TeamDataProviderConfig) (*TeamDataProvider, error) {
	client, clientErr := rezslack.LoadClient()
	if clientErr != nil {
		return nil, clientErr
	}
	return &TeamDataProvider{client: client}, nil
}

var (
	teamDataMapping = ent.Team{
		Name: "y",
		Edges: ent.TeamEdges{
			Users: []*ent.User{&ent.User{
				ChatID: "y",
			}},
		},
	}
)

func (p *TeamDataProvider) TeamDataMapping() *ent.Team {
	return &teamDataMapping
}

func pullSlackTeams(ctx context.Context, client *slack.Client) iter.Seq2[*slack.Team, error] {
	var cursor string
	return func(yield func(*slack.Team, error) bool) {
		for {
			params := slack.ListTeamsParameters{Limit: 20, Cursor: cursor}
			slackTeams, newCursor, listErr := client.ListTeamsContext(ctx, params)
			if listErr != nil {
				yield(nil, listErr)
				break
			}
			for _, slackTeam := range slackTeams {
				yield(&slackTeam, nil)
			}
			if newCursor == "" {
				break
			}
			cursor = newCursor
		}
	}
}

func (p *TeamDataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	return func(yield func(*ent.Team, error) bool) {
		for slackTeam, teamsErr := range pullSlackTeams(ctx, p.client) {
			if teamsErr != nil {
				yield(nil, fmt.Errorf("get teams: %w", teamsErr))
				return
			}

			userGroups, userGroupsErr := p.client.GetUserGroupsContext(ctx,
				slack.GetUserGroupsOptionWithTeamID(slackTeam.ID),
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
					ProviderID: userGroup.ID,
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
