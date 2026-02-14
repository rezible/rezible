package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
)

const integrationName = "slack"

var supportedDataKinds = []string{"chat", "users"}

type integration struct {
	services             *rez.Services
	oauth2Config         *oauth2.Config
	eventListeners       map[string]rez.EventListener
	webhookHandlers      map[string]http.Handler
	incidentEventHandler *incidentEventHandler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	sl := newServiceLoader(svcs)

	incHandler, incHandlerErr := newIncidentEventHandler(sl, svcs.Messages, svcs.Incidents)
	if incHandlerErr != nil {
		return nil, fmt.Errorf("incident event handler: %w", incHandlerErr)
	}

	intg := &integration{
		services:             svcs,
		oauth2Config:         loadOAuthConfig(),
		eventListeners:       make(map[string]rez.EventListener),
		webhookHandlers:      make(map[string]http.Handler),
		incidentEventHandler: incHandler,
	}

	if UseSocketMode() {
		el, lErr := newSocketModeEventListener(svcs)
		if lErr != nil {
			return nil, fmt.Errorf("failed to create event listener: %w", lErr)
		}
		intg.eventListeners["slack_socketmode"] = el
	} else {
		wh, whErr := newWebhookEventHandler(sl, svcs)
		if whErr != nil {
			return nil, fmt.Errorf("webhook event handler: %w", whErr)
		}
		intg.webhookHandlers["/"] = wh.Handler()
	}

	return intg, nil
}

func (i *integration) Name() string {
	return integrationName
}

func (i *integration) IsAvailable() (bool, error) {
	// TODO: check config
	return true, nil
}

func (i *integration) EventListeners() map[string]rez.EventListener {
	return i.eventListeners
}

func (i *integration) WebhookHandlers() map[string]http.Handler {
	return i.webhookHandlers
}

func (i *integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *integration) OAuthConfigRequired() bool {
	return true
}

func (i *integration) OAuth2Config() *oauth2.Config {
	return i.oauth2Config
}

func loadOAuthConfig() *oauth2.Config {
	clientId := rez.Config.GetString("slack.oauth_client_id")
	clientSecret := rez.Config.GetString("slack.oauth_client_secret")
	scopes := []string{
		"app_mentions:read",
		"assistant:write",
		"channels:history",
		"channels:join",
		"channels:read",
		"chat:write",
		"chat:write.customize",
		"chat:write.public",
		"commands",
		"groups:history",
		"groups:read",
		"im:history",
		"im:read",
		"im:write",
		"im:write.topic",
		"incoming-webhook",
		"metadata.message:read",
		"mpim:history",
		"pins:read",
		"reactions:read",
		"usergroups:read",
		"users.profile:read",
		"users:read",
		"users:read.email",
		"channels:write.topic",
		"channels:manage",
		"channels:write.invites",
	}

	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{strings.Join(scopes, ",")},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/v2/authorize",
			TokenURL: "https://slack.com/api/oauth.v2.access",
		},
	}
}

func (i *integration) ExtractIntegrationConfigFromToken(t *oauth2.Token) (json.RawMessage, error) {
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

	return json.Marshal(cfg)
}

func lookupIntegration(ctx context.Context, is rez.IntegrationsService, teamId string, enterpriseId string) (*ent.Integration, error) {
	vals := make(map[string]any)
	if teamId != "" {
		vals["Team.ID"] = teamId
	}
	if enterpriseId != "" {
		vals["Enterprise.ID"] = enterpriseId
	}
	return is.LookupByConfigValues(access.SystemContext(ctx), integrationName, vals)
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return &ConfiguredIntegration{svcs: i.services, intg: intg}
}

type ConfiguredIntegration struct {
	svcs *rez.Services
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
	cfg, cfgErr := decodeConfig(ci.intg.Config)
	if cfgErr != nil {
		return nil, fmt.Errorf("unable to decode config: %w", cfgErr)
	}
	return newChatService(cfg.makeClient(), ci.svcs), nil
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
