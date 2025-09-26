package slack

import (
	"context"
	"fmt"
	"iter"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezslack "github.com/rezible/rezible/internal/slack"
)

type UserDataProvider struct {
	client       *slack.Client
	workspaceIds []string
}

var _ rez.UserDataProvider = (*UserDataProvider)(nil)

type UserDataProviderConfig struct {
	WorkspaceIds []string `json:"workspace_ids"`
}

func DebugOnlyUserDataProvider(ctx context.Context) (*UserDataProvider, error) {
	client, clientErr := rezslack.LoadClient()
	if clientErr != nil {
		return nil, clientErr
	}
	var ids []string
	teams, _, teamsErr := client.ListTeamsContext(ctx, slack.ListTeamsParameters{})
	if teamsErr != nil {
		return nil, teamsErr
	}
	for _, team := range teams {
		ids = append(ids, team.ID)
	}
	return &UserDataProvider{client: client, workspaceIds: ids}, nil
}

func NewUserDataProvider(cfg UserDataProviderConfig) (*UserDataProvider, error) {
	client, clientErr := rezslack.LoadClient()
	if clientErr != nil {
		return nil, clientErr
	}
	return &UserDataProvider{client: client, workspaceIds: cfg.WorkspaceIds}, nil
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
	return func(yield func(*ent.User, error) bool) {
		for _, workspaceId := range p.workspaceIds {
			slackUsers, getErr := p.client.GetUsersContext(ctx,
				slack.GetUsersOptionPresence(false),
				slack.GetUsersOptionTeamID(workspaceId))
			if getErr != nil {
				yield(nil, fmt.Errorf("slack get users err: %w", getErr))
				return
			}
			for _, u := range slackUsers {
				log.Debug().Interface("user", u).Msg("slack user")
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
