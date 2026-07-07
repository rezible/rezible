package github

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
	"golang.org/x/oauth2"
)

func (i *Integration) OAuthInstallRequired() bool {
	return true
}

func (i *Integration) OAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     i.cfg.App.ClientID,
		ClientSecret: i.cfg.App.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://github.com/login/oauth/authorize",
			TokenURL:  "https://github.com/login/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func (i *Integration) RetrieveInstallationTargetOptions(ctx context.Context, t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	if t == nil || t.AccessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}
	installations, err := i.listUserInstallations(ctx, t.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("list github installations: %w", err)
	}
	return i.makeInstallationTargetOptions(installations)
}

func (i *Integration) listUserInstallations(ctx context.Context, accessToken string) ([]*gh.Installation, error) {
	client := gh.NewClient(nil).WithAuthToken(accessToken)
	var all []*gh.Installation
	opts := &gh.ListOptions{PerPage: 100}
	for page := 1; ; page++ {
		opts.Page = page
		installations, resp, err := client.Apps.ListUserInstallations(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("list user installations page %d: %w", page, err)
		}
		all = append(all, installations...)
		if resp.NextPage == 0 {
			break
		}
	}
	return all, nil
}

func (i *Integration) makeInstallationTargetOptions(installations []*gh.Installation) ([]rez.IntegrationInstallationTarget, error) {
	if len(installations) == 0 {
		return nil, fmt.Errorf("no valid github app installations")
	}
	options := make([]rez.IntegrationInstallationTarget, 0, len(installations))
	for _, inst := range installations {
		if inst.GetAppID() != i.cfg.App.AppID {
			continue
		}
		cfg, cfgErr := i.MakeInstallationConfigFromGithub(inst)
		if cfgErr != nil {
			return nil, fmt.Errorf("make installation config: %w", cfgErr)
		}
		options = append(options, rez.IntegrationInstallationTarget{
			IntegrationName: integrationName,
			DisplayName:     cfg.Org,
			Config:          cfg,
		})
	}
	return options, nil
}
