package slack

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/slack-go/slack"
)

const integrationName = "slack"

type IntegrationConfigData struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Scope       string    `json:"scope"`
	BotUserID   string    `json:"bot_user_id"`
	Team        *teamInfo `json:"team"`
	Enterprise  *teamInfo `json:"enterprise"`
}

type teamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func loadIntegrationConfig(ctx context.Context, s rez.IntegrationsService) (*IntegrationConfigData, error) {
	params := rez.ListIntegrationsParams{
		Name: integrationName,
	}
	results, listErr := s.ListIntegrations(ctx, params)
	if listErr != nil {
		return nil, listErr
	}
	// TODO: handle multiple??
	if len(results) != 1 {
		return nil, fmt.Errorf("expected 1 integration, got %d", len(results))
	}
	var cfg IntegrationConfigData
	if jsonErr := json.Unmarshal(results[0].Config, &cfg); jsonErr != nil {
		return nil, jsonErr
	}
	return &cfg, nil
}

func getClient(ctx context.Context, s rez.IntegrationsService) (*slack.Client, error) {
	if rez.Config.SingleTenantMode() {
		return LoadSingleTenantClient()
	}
	cfg, loadErr := loadIntegrationConfig(ctx, s)
	if loadErr != nil {
		return nil, fmt.Errorf("loading integration config: %w", loadErr)
	}
	return slack.New(cfg.AccessToken), nil
}

func withClient(ctx context.Context, s rez.IntegrationsService, fn func(*slack.Client) error) error {
	client, clientErr := getClient(ctx, s)
	if clientErr != nil {
		return fmt.Errorf("get slack client: %w", clientErr)
	}
	return fn(client)
}
