package github

import (
	"context"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
	"golang.org/x/oauth2"
)

func (i *Integration) OAuthInstallRequired() bool {
	return true
}

func (i *Integration) OAuth2Config() *oauth2.Config {
	return i.oauth2Config
}

func (i *Integration) loadOAuthConfig() *oauth2.Config {
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

func (i *Integration) makeInstallationTargetOptions(installations []*gh.Installation) ([]rez.IntegrationInstallationTarget, error) {
	matches := make([]*gh.Installation, 0, len(installations))
	for _, installation := range installations {
		if installation == nil || installation.GetID() == 0 || installation.GetAccount().GetLogin() == "" {
			continue
		}
		if installation.GetAppID() != i.cfg.App.AppID {
			continue
		}
		matches = append(matches, installation)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid github app installations found for authenticated user")
	}

	options := make([]rez.IntegrationInstallationTarget, 0, len(matches))
	for _, installation := range matches {
		option := rez.IntegrationInstallationTarget{
			ExternalRef: strconv.FormatInt(installation.GetID(), 10),
			DisplayName: installation.GetAccount().GetLogin(),
			InstallationConfig: map[string]any{
				configOrg:            installation.GetAccount().GetLogin(),
				configInstallationID: installation.GetID(),
			},
		}
		options = append(options, option)
	}
	return options, nil
}
