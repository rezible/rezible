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
	client       *slack.Client
	workspaceIds []string
}

var _ rez.UserDataProvider = (*UserDataProvider)(nil)

type UserDataProviderConfig struct {
	WorkspaceIds []string `json:"workspace_ids"`
}

func NewUserDataProvider() (*UserDataProvider, error) {
	client, clientErr := LoadSingleTenantClient()
	if clientErr != nil {
		return nil, clientErr
	}
	return &UserDataProvider{client: client, workspaceIds: nil}, nil
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
