package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
)

const integrationName = "slack"

type integration struct {
	chatService     *ChatService
	eventListeners  map[string]rez.EventListener
	webhookHandlers map[string]http.Handler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	cs, csErr := NewChatService(ctx, svcs)
	if csErr != nil {
		return nil, fmt.Errorf("failed to create chat service: %w", csErr)
	}

	intg := &integration{
		chatService:     cs,
		eventListeners:  make(map[string]rez.EventListener),
		webhookHandlers: make(map[string]http.Handler),
	}

	if UseSocketMode() {
		el, lErr := NewSocketModeEventListener(cs)
		if lErr != nil {
			return nil, fmt.Errorf("failed to create event listener: %w", lErr)
		}
		intg.eventListeners["slack_socketmode"] = el
	} else {
		wh, whErr := NewWebhookEventHandler(cs)
		if whErr != nil {
			return nil, fmt.Errorf("webhook event handler: %w", whErr)
		}
		intg.webhookHandlers["/"] = wh.Handler()
	}

	return intg, nil
}

func (d *integration) Name() string {
	return integrationName
}

func (d *integration) Enabled() bool {
	// TODO: check config
	return true
}

func (d *integration) EventListeners() map[string]rez.EventListener {
	return d.eventListeners
}

func (d *integration) WebhookHandlers() map[string]http.Handler {
	return d.webhookHandlers
}

func (d *integration) SupportedDataKinds() []string {
	return []string{"chat", "users"}
}

func (d *integration) GetChatService() rez.ChatService {
	return d.chatService
}

func (d *integration) OAuthConfigRequired() bool {
	return true
}

func (d *integration) ValidateConfig(cfg json.RawMessage) (bool, error) {

	return true, nil
}

func (d *integration) GetUserConfig(rawCfg json.RawMessage) (json.RawMessage, error) {
	var cfg IntegrationConfig
	if jsonErr := json.Unmarshal(rawCfg, &cfg); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal config: %w", jsonErr)
	}
	return json.Marshal(cfg.UserConfig)
}

func (d *integration) MergeUserConfig(rawCfg json.RawMessage, rawUserCfg json.RawMessage) (json.RawMessage, error) {
	var cfg IntegrationConfig
	if jsonErr := json.Unmarshal(rawCfg, &cfg); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal config: %w", jsonErr)
	}
	if jsonErr := json.Unmarshal(rawUserCfg, &cfg.UserConfig); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal user config: %w", jsonErr)
	}
	return json.Marshal(cfg)
}

func (d *integration) OAuth2Config() *oauth2.Config {
	return LoadOAuthConfig()
}

func (d *integration) GetIntegrationConfigFromToken(token *oauth2.Token) (any, error) {
	return getIntegrationConfigFromOAuthToken(token)
}

type IntegrationConfig struct {
	AccessToken string
	TokenType   string
	Scope       string
	BotUserID   string
	Team        teamInfo
	Enterprise  *teamInfo
	UserConfig  IntegrationUserConfig
}

type teamInfo struct {
	ID   string
	Name string
}

type IntegrationUserConfig struct {
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

func getIntegrationConfigFromOAuthToken(t *oauth2.Token) (*IntegrationConfig, error) {
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

	cfg := IntegrationConfig{
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

func decodeConfig(intg *ent.Integration) (*IntegrationConfig, error) {
	if intg.Name != integrationName {
		return nil, fmt.Errorf("invalid integration name")
	}
	var cfg IntegrationConfig
	if cfgErr := json.Unmarshal(intg.Config, &cfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to decode integration config: %w", cfgErr)
	}
	return &cfg, nil
}

func loadIntegrationConfig(ctx context.Context, s rez.IntegrationsService) (*IntegrationConfig, error) {
	params := rez.ListIntegrationsParams{
		Name: integrationName,
	}
	results, listErr := s.ListIntegrations(ctx, params)
	if listErr != nil {
		return nil, listErr
	}
	if len(results) != 1 {
		return nil, fmt.Errorf("expected 1 integration, got %d", len(results))
	}
	return decodeConfig(results[0])
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
