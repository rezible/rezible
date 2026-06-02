package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
)

type githubClient struct {
	ci     *InstalledIntegration
	cfg    *installationConfig
	client *github.Client
}

func newAppClient(cfg rez.IntegrationsConfigGithub, ii *InstalledIntegration) (*githubClient, error) {
	icfg, icfgErr := ii.config()
	if icfgErr != nil {
		return nil, icfgErr
	}
	transport, transportErr := ghinstallation.New(
		http.DefaultTransport,
		cfg.App.AppID,
		icfg.InstallationID,
		[]byte(cfg.App.PrivateKeyPEM),
	)
	if transportErr != nil {
		return nil, fmt.Errorf("create github app transport: %w", transportErr)
	}
	client := github.NewClient(&http.Client{Transport: transport})

	return &githubClient{ci: ii, cfg: icfg, client: client}, nil
}

func (c *githubClient) ListRepositories(ctx context.Context) ([]*github.Repository, error) {
	var all []*github.Repository
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for page := 1; ; page++ {
		opts.Page = page
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, c.cfg.Org, opts)
		if err != nil {
			return nil, fmt.Errorf("list org repos page %d: %w", page, err)
		}
		all = append(all, repos...)
		if resp.NextPage == 0 {
			break
		}
	}
	return all, nil
}
