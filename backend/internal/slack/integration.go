package slack

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
)

const integrationName = "slack"

type IntegrationConfigData struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Scope       string    `json:"scope"`
	BotUserID   string    `json:"bot_user_id"`
	Team        teamInfo  `json:"team"`
	Enterprise  *teamInfo `json:"enterprise"`
}

type teamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getTeamInfoFromTokenExtra(e map[string]interface{}) (*teamInfo, error) {
	id, idOk := e["id"].(string)
	if !idOk {
		return nil, fmt.Errorf("missing or invalid id")
	}
	name, nameOk := e["name"].(string)
	if !nameOk {
		return nil, fmt.Errorf("missing or invalid name")
	}
	return &teamInfo{ID: id, Name: name}, nil
}

func getIntegrationConfigFromOAuthToken(t *oauth2.Token) (*IntegrationConfigData, error) {
	scope, scopeOk := t.Extra("scope").(string)
	if !scopeOk {
		return nil, fmt.Errorf("missing or invalid scope")
	}

	botUserId, botUserIdOk := t.Extra("bot_user_id").(string)
	if !botUserIdOk {
		return nil, fmt.Errorf("missing or invalid bot_user_id")
	}

	teamRaw, teamOk := t.Extra("team").(map[string]interface{})
	if !teamOk {
		return nil, fmt.Errorf("missing or invalid team")
	}

	team, teamErr := getTeamInfoFromTokenExtra(teamRaw)
	if teamErr != nil {
		return nil, fmt.Errorf("invalid team info")
	}

	cfg := IntegrationConfigData{
		AccessToken: t.AccessToken,
		TokenType:   t.Type(),
		Scope:       scope,
		BotUserID:   botUserId,
		Team:        *team,
	}

	// isEnterprise, isEntOk := t.Extra("is_enterprise_install").(string)

	if enterprise, eOk := t.Extra("enterprise").(map[string]interface{}); eOk {
		e, entErr := getTeamInfoFromTokenExtra(enterprise)
		if entErr != nil {
			log.Error().Err(entErr).Msgf("get enterprise info from token")
		}
		cfg.Enterprise = e
	}

	return &cfg, nil
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
