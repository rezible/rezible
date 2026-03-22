package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
)

const integrationName = "slack"

var supportedDataKinds = []string{"chat", "users"}

type Config struct {
	Enabled    bool   `koanf:"enabled"`
	AppToken   string `koanf:"app_token"`
	BotToken   string `koanf:"bot_token"`
	SocketMode struct {
		Enabled bool `koanf:"enabled"`
	} `koanf:"socketmode"`
	Webhooks struct {
		SigningSecret string `koanf:"signing_secret"`
	} `koanf:"webhooks"`
	OAuth struct {
		ClientId     string `koanf:"client_id"`
		ClientSecret string `koanf:"client_secret"`
	} `koanf:"oauth"`
}

func (c Config) validate() error {
	var errs []error
	if c.OAuth.ClientId == "" {
		errs = append(errs, errors.New("slack.oauth.client_id not set"))
	}
	if c.OAuth.ClientSecret == "" {
		errs = append(errs, errors.New("slack.oauth.client_secret not set"))
	}

	if c.SocketMode.Enabled {
		if !rez.Config.SingleTenantMode() {
			errs = append(errs, errors.New("socket mode requires single tenant mode"))
		}
		if c.AppToken == "" {
			errs = append(errs, errors.New("slack.app_token not set"))
		}
		if c.BotToken == "" {
			errs = append(errs, errors.New("slack.bot_token not set"))
		}
	} else {
		if c.Webhooks.SigningSecret == "" {
			errs = append(errs, errors.New("slack.webhooks.signing_secret not set"))
		}
	}
	return errors.Join(errs...)
}

type integration struct {
	cfg             Config
	services        *rez.Services
	oauth2Config    *oauth2.Config
	eventListeners  map[string]rez.EventListener
	webhookHandlers map[string]http.Handler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{
		services:        svcs,
		webhookHandlers: make(map[string]http.Handler),
		eventListeners:  make(map[string]rez.EventListener),
		cfg:             Config{},
	}

	if cfgErr := rez.Config.Unmarshal("slack", &intg.cfg); cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	oauthCfg, oauthErr := intg.loadOAuthConfig()
	if oauthErr != nil {
		return nil, fmt.Errorf("oauth config error: %w", oauthErr)
	}
	intg.oauth2Config = oauthCfg

	if handlersErr := intg.makeEventHandlers(ctx); handlersErr != nil {
		return nil, fmt.Errorf("event handlers: %w", handlersErr)
	}

	return intg, nil
}

func (i *integration) makeEventHandlers(ctx context.Context) error {
	eh, ehErr := i.makeEventHandler()
	if ehErr != nil {
		return fmt.Errorf("event handler: %w", ehErr)
	}

	if i.cfg.SocketMode.Enabled {
		sml, smlErr := i.newSocketModeEventListener(eh)
		if smlErr != nil {
			return fmt.Errorf("socketmode listener: %w", smlErr)
		}
		i.eventListeners["slack_socketmode"] = sml
	} else {
		wh, whErr := i.newWebhookListener(eh)
		if whErr != nil {
			return fmt.Errorf("webhook event handler: %w", whErr)
		}
		i.webhookHandlers["/"] = wh.Handler()
	}
	return nil
}

func (i *integration) Name() string {
	return integrationName
}

func (i *integration) IsAvailable() (bool, error) {
	if !i.cfg.Enabled {
		return false, nil
	}
	return true, i.cfg.validate()
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

var oAuthScopes = []string{
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

func (i *integration) loadOAuthConfig() (*oauth2.Config, error) {
	if i.cfg.OAuth.ClientId == "" || i.cfg.OAuth.ClientSecret == "" {
		return nil, fmt.Errorf("failed to load OAuth client id or client secret")
	}

	return &oauth2.Config{
		ClientID:     i.cfg.OAuth.ClientId,
		ClientSecret: i.cfg.OAuth.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/v2/authorize",
			TokenURL: "https://slack.com/api/oauth.v2.access",
		},
		Scopes: oAuthScopes,
	}, nil
}

func validateOauthTokenScopes(t *oauth2.Token) error {
	tokenScope, scopeOk := t.Extra("scope").(string)
	if !scopeOk {
		return fmt.Errorf("missing or invalid scope")
	}
	tokenScopes := mapset.NewSet(strings.Split(tokenScope, ",")...)
	for _, scope := range oAuthScopes {
		if !tokenScopes.Contains(scope) {
			return fmt.Errorf("missing token scope: %s", scope)
		}
	}
	return nil
}

func getTeamInfoFromTokenData(tokenData any) (map[string]any, error) {
	info := objx.New(tokenData)
	id := info.Get("id")
	name := info.Get("name")
	if !id.IsStr() || !name.IsStr() {
		return nil, fmt.Errorf("missing or invalid team info")
	}
	data := map[string]any{
		"id":   id.String(),
		"name": name.String(),
	}
	return data, nil
}

func (i *integration) ExtractIntegrationConfigFromToken(t *oauth2.Token) (map[string]any, error) {
	if scopesErr := validateOauthTokenScopes(t); scopesErr != nil {
		return nil, scopesErr
	}

	cfg := objx.Map{}
	cfg.Set(configAccessToken, t.AccessToken)

	botUserId, botUserIdOk := t.Extra("bot_user_id").(string)
	if !botUserIdOk {
		return nil, fmt.Errorf("missing or invalid bot_user_id")
	}
	cfg.Set(configBotUserID, botUserId)

	team, teamErr := getTeamInfoFromTokenData(t.Extra("team"))
	if teamErr != nil {
		return nil, fmt.Errorf("invalid team info")
	}
	cfg.Set(configTeam, team)

	enterprise, enterpriseErr := getTeamInfoFromTokenData(t.Extra("enterprise"))
	if enterpriseErr != nil {
		log.Warn().Err(enterpriseErr).Msgf("get enterprise info from token")
	} else {
		cfg.Set(configEnterprise, enterprise)
	}

	wh := objx.New(t.Extra("incoming_webhook"))
	if channelId := wh.Get("channel_id"); channelId.IsStr() {
		cfg.Set(configWebhookChannelId, channelId.String())
	}

	return cfg, nil
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return newConfiguredIntegration(i.services, intg)
}

func (i *integration) ValidateConfig(cfg map[string]any) error {
	return nil
}

func (i *integration) ValidateUserPreferences(prefs map[string]any) error {
	return nil
}

type ConfiguredIntegration struct {
	svcs *rez.Services
	intg *ent.Integration
}

func newConfiguredIntegration(svcs *rez.Services, intg *ent.Integration) *ConfiguredIntegration {
	return &ConfiguredIntegration{svcs: svcs, intg: intg}
}

func (ci *ConfiguredIntegration) tenantContext(ctx context.Context) context.Context {
	return access.TenantContext(ctx, ci.intg.TenantID)
}

func (ci *ConfiguredIntegration) config() objx.Map {
	return objx.New(ci.intg.Config)
}

const (
	configAccessToken      = "access_token"
	configBotUserID        = "bot_user_id"
	configWebhookChannelId = "webhook_channel_id"
	configTeam             = "team"
	configEnterprise       = "enterprise"
)

func (ci *ConfiguredIntegration) accessToken() string {
	return ci.config().Get(configAccessToken).String()
}

func (ci *ConfiguredIntegration) teamId() string {
	return ci.config().Get(configTeam + ".id").String()
}

const (
	userPreferencesIncidentAnnouncementChannelId = "incidents.announcement_channel_id"
	userPreferencesIncidentChannelNamePattern    = "incidents.channel_name_pattern"
	userPreferencesIncidentCreateVideoConference = "incidents.create_video_conference"
	userPreferencesIncidentInviteMode            = "incidents.invite_mode"
)

func (ci *ConfiguredIntegration) userPreferences() objx.Map {
	return objx.New(ci.intg.UserPreferences)
}

type incidentPreferences struct {
	AnnouncementChannelID     string
	ChannelNamePattern        string
	AutoCreateVideoConference bool
	InviteMode                string
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() map[string]any {
	return ci.config().Exclude([]string{})
}

func (ci *ConfiguredIntegration) GetUserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) getIncidentPreferences() incidentPreferences {
	prefs := ci.userPreferences()
	defaultAnnouncementChannelId := ci.config().Get(configWebhookChannelId).String()
	return incidentPreferences{
		AnnouncementChannelID:     prefs.Get(userPreferencesIncidentAnnouncementChannelId).Str(defaultAnnouncementChannelId),
		ChannelNamePattern:        prefs.Get(userPreferencesIncidentChannelNamePattern).Str("incident-{slug}"),
		AutoCreateVideoConference: prefs.Get(userPreferencesIncidentCreateVideoConference).Bool(),
		InviteMode:                prefs.Get(userPreferencesIncidentInviteMode).Str("assigned_users"),
	}
}

func (ci *ConfiguredIntegration) GetDataKinds() map[string]bool {
	return map[string]bool{
		"chat":  true,
		"users": true,
	}
}

func (ci *ConfiguredIntegration) MakeChatService(ctx context.Context) (rez.ChatService, error) {
	return newChatService(ci), nil
}

type installIds struct {
	TeamId       string
	EnterpriseId string
}

func (i installIds) configValues() map[string]any {
	m := map[string]any{}
	if i.TeamId != "" {
		m["team.id"] = i.TeamId
	}
	if i.EnterpriseId != "" {
		m["enterprise.id"] = i.EnterpriseId
	}
	return m
}

func lookupTenantIntegration(ctx context.Context, integrations rez.IntegrationsService, ids installIds) (*ConfiguredIntegration, error) {
	params := rez.ListIntegrationsParams{
		Names:        []string{integrationName},
		ConfigValues: ids.configValues(),
	}
	intgs, listErr := integrations.ListConfigured(access.SystemContext(ctx), params)
	if listErr != nil {
		if ent.IsNotFound(listErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing configured integrations: %w", listErr)
	}
	for _, intg := range intgs {
		if ci, ok := intg.(*ConfiguredIntegration); ok {
			return ci, nil
		}
	}
	return nil, fmt.Errorf("integration not found")
}
