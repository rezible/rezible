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
	client       *slack.Client
	workspaceIds []string
}

var _ rez.TeamDataProvider = (*TeamDataProvider)(nil)

type TeamDataProviderConfig struct {
	WorkspaceIds []string `json:"workspace_ids"`
}

func NewTeamDataProvider(cfg TeamDataProviderConfig) (*TeamDataProvider, error) {
	client, clientErr := rezslack.LoadSingleTenantClient()
	if clientErr != nil {
		return nil, clientErr
	}
	return &TeamDataProvider{client: client}, nil
}

var (
	teamDataMapping = ent.Team{
		Name: "y",
		Edges: ent.TeamEdges{
			Users: []*ent.User{{
				ChatID: "y",
			}},
		},
	}
)

func (p *TeamDataProvider) TeamDataMapping() *ent.Team {
	return &teamDataMapping
}

func (p *TeamDataProvider) PullTeams(ctx context.Context) iter.Seq2[*ent.Team, error] {
	return func(yield func(*ent.Team, error) bool) {
		for _, workspaceId := range p.workspaceIds {
			userGroups, userGroupsErr := p.client.GetUserGroupsContext(ctx,
				slack.GetUserGroupsOptionWithTeamID(workspaceId),
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
