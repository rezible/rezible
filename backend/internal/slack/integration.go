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

var supportedDataKinds = []string{"chat", "users"}

type integration struct {
	services        *rez.Services
	eventListeners  map[string]rez.EventListener
	webhookHandlers map[string]http.Handler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{
		services:        svcs,
		eventListeners:  make(map[string]rez.EventListener),
		webhookHandlers: make(map[string]http.Handler),
	}

	l := newLoader(svcs)

	incMsgHandler := newIncidentChatEventHandler(l, svcs.Messages, svcs.Incidents)
	if msgsErr := incMsgHandler.registerHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("adding message handlers: %w", msgsErr)
	}

	if UseSocketMode() {
		el, lErr := newSocketModeEventListener(svcs)
		if lErr != nil {
			return nil, fmt.Errorf("failed to create event listener: %w", lErr)
		}
		intg.eventListeners["slack_socketmode"] = el
	} else {
		wh, whErr := newWebhookEventHandler(l, svcs)
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
	return supportedDataKinds
}

func (d *integration) OAuthConfigRequired() bool {
	return true
}

func (d *integration) OAuth2Config() *oauth2.Config {
	return LoadOAuthConfig()
}

func (d *integration) GetIntegrationConfigFromToken(t *oauth2.Token) (any, error) {
	getTeamInfoFromTokenExtra := func(extraKey string) (*teamInfo, error) {
		e, eOk := t.Extra(extraKey).(map[string]interface{})
		if !eOk {
			return nil, fmt.Errorf("missing or invalid extra field: %s", extraKey)
		}

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

	scope, scopeOk := t.Extra("scope").(string)
	if !scopeOk {
		return nil, fmt.Errorf("missing or invalid scope")
	}

	botUserId, botUserIdOk := t.Extra("bot_user_id").(string)
	if !botUserIdOk {
		return nil, fmt.Errorf("missing or invalid bot_user_id")
	}

	team, teamErr := getTeamInfoFromTokenExtra("team")
	if teamErr != nil {
		return nil, fmt.Errorf("invalid team info")
	}

	// isEnterprise, isEntOk := t.Extra("is_enterprise_install").(string)

	enterprise, entErr := getTeamInfoFromTokenExtra("enterprise")
	if entErr != nil {
		log.Warn().Err(entErr).Msgf("get enterprise info from token")
	}

	cfg := IntegrationConfig{
		AccessToken: t.AccessToken,
		TokenType:   t.Type(),
		Scope:       scope,
		BotUserID:   botUserId,
		Team:        *team,
		Enterprise:  enterprise,
	}

	return &cfg, nil
}

func (d *integration) GetConfiguredIntegration(i *ent.Integration) rez.ConfiguredIntegration {
	return &ConfiguredIntegration{intg: i}
}

type ConfiguredIntegration struct {
	intg *ent.Integration
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) RawConfig() json.RawMessage {
	return ci.intg.Config
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() (json.RawMessage, error) {
	return json.Marshal(ci.RawConfig())
}

func (ci *ConfiguredIntegration) UserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) EnabledDataKinds() []string {
	return supportedDataKinds
}

func (ci *ConfiguredIntegration) ChatService(ctx context.Context) (rez.ChatService, error) {
	return nil, nil
}

type IntegrationConfig struct {
	AccessToken string
	TokenType   string
	Scope       string
	BotUserID   string
	Team        teamInfo
	Enterprise  *teamInfo
}

type teamInfo struct {
	ID   string
	Name string
}

func (c *IntegrationConfig) makeClient() *slack.Client {
	return slack.New(c.AccessToken)
}

func decodeConfig(raw json.RawMessage) (*IntegrationConfig, error) {
	var cfg IntegrationConfig
	if cfgErr := json.Unmarshal(raw, &cfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to decode integration config: %w", cfgErr)
	}
	return &cfg, nil
}
