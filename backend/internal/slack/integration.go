package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
)

const integrationName = "slack"

var supportedDataKinds = []string{"chat", "users"}

type integration struct {
	services        *rez.Services
	oauth2Config    *oauth2.Config
	eventListeners  map[string]rez.EventListener
	webhookHandlers map[string]http.Handler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{
		services:        svcs,
		oauth2Config:    loadOAuthConfig(),
		webhookHandlers: make(map[string]http.Handler),
		eventListeners:  make(map[string]rez.EventListener),
	}

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

	if i.useSocketMode() {
		sml, smlErr := newSocketModeEventListener(eh)
		if smlErr != nil {
			return fmt.Errorf("socketmode listener: %w", smlErr)
		}
		i.eventListeners["slack_socketmode"] = sml
	} else {
		wh, whErr := newWebhookListener(eh)
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

func (i *integration) useSocketMode() bool {
	return rez.Config.GetBool("slack.socketmode.enabled")
}

func (i *integration) IsAvailable() (bool, error) {
	var errs []error

	if rez.Config.GetString("slack.oauth.client_id") == "" {
		errs = append(errs, errors.New("slack.oauth.client_id not set"))
	}
	if rez.Config.GetString("slack.oauth.client_secret") == "" {
		errs = append(errs, errors.New("slack.oauth.client_secret not set"))
	}

	if i.useSocketMode() {
		if !rez.Config.SingleTenantMode() {
			errs = append(errs, errors.New("socket mode requires single tenant mode"))
		}
		if rez.Config.GetString("slack.app_token") == "" {
			errs = append(errs, errors.New("slack.app_token not set"))
		}
		if rez.Config.GetString("slack.bot_token") == "" {
			errs = append(errs, errors.New("slack.bot_token not set"))
		}
	} else {
		if rez.Config.GetString("slack.webhook_signing_secret") == "" {
			errs = append(errs, errors.New("slack.webhook_signing_secret not set"))
		}
	}
	if len(errs) > 0 {
		return false, errors.Join(errs...)
	}
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

func loadOAuthConfig() *oauth2.Config {
	clientId := rez.Config.GetString("slack.oauth.client_id")
	clientSecret := rez.Config.GetString("slack.oauth.client_secret")

	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{strings.Join(oAuthScopes, ",")},
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
	tokenScopes := mapset.NewSet(strings.Split(scope, ",")...)
	for _, s := range oAuthScopes {
		if !tokenScopes.Contains(s) {
			return nil, fmt.Errorf("missing token scope: %s", s)
		}
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
		BotUserID:   botUserId,
		Team:        *team,
		Enterprise:  enterprise,
	}

	if wh, whOk := t.Extra("incoming_webhook").(map[string]any); whOk {
		if channelId, ok := wh["channel_id"]; ok {
			if cidStr, strOk := channelId.(string); strOk {
				cfg.WebhookChannelId = cidStr
			}
		}
	}

	return json.Marshal(cfg)
}

func lookupTenantIntegration(ctx context.Context, is rez.IntegrationsService, teamId string, enterpriseId string) (*ent.Integration, error) {
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
	return newConfiguredIntegration(i.services, intg)
}

func (i *integration) getSingleTenantConfiguredIntegration(ctx context.Context) (*ConfiguredIntegration, error) {
	return nil, fmt.Errorf("not implemented")
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

type IncidentDefaults struct {
	AnnouncementChannelID     string
	ChannelNamePattern        string
	DefaultSeverityID         string
	DefaultTypeID             string
	AutoCreateVideoConference bool
	InviteMode                string
}

func ValidateIncidentDefaultsPreferences(pref map[string]any) error {
	defaults := getNestedPreferenceMap(pref, "incidentDefaults")
	if defaults == nil {
		return nil
	}

	announcementChannelID := getStringPreference(defaults, "announcementChannelId", "")
	if announcementChannelID == "" {
		return errors.New("incidentDefaults.announcementChannelId is required")
	}

	channelNamePattern := getStringPreference(defaults, "channelNamePattern", "")
	if channelNamePattern == "" {
		return errors.New("incidentDefaults.channelNamePattern is required")
	}
	if !strings.Contains(channelNamePattern, "{slug}") && !strings.Contains(channelNamePattern, "{id}") {
		return errors.New("incidentDefaults.channelNamePattern must include {slug} or {id}")
	}

	for _, value := range []struct {
		field string
		raw   string
	}{
		{field: "incidentDefaults.defaultSeverityId", raw: getStringPreference(defaults, "defaultSeverityId", "")},
		{field: "incidentDefaults.defaultTypeId", raw: getStringPreference(defaults, "defaultTypeId", "")},
	} {
		if value.raw == "" {
			continue
		}
		if _, err := uuid.Parse(value.raw); err != nil {
			return fmt.Errorf("%s must be a valid UUID", value.field)
		}
	}

	inviteMode := getStringPreference(defaults, "inviteMode", "assigned_users")
	switch inviteMode {
	case "", "assigned_users":
		return nil
	default:
		return fmt.Errorf("incidentDefaults.inviteMode %q is not supported for launch", inviteMode)
	}
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

func (ci *ConfiguredIntegration) getUserPreference(key string, defaultVal any) any {
	if pref, ok := ci.intg.UserPreferences[key]; ok {
		return pref
	}
	return defaultVal
}

func getNestedPreferenceMap(pref map[string]any, key string) map[string]any {
	raw, ok := pref[key]
	if !ok {
		return nil
	}
	if nested, ok := raw.(map[string]any); ok {
		return nested
	}
	return nil
}

func getStringPreference(pref map[string]any, key string, defaultVal string) string {
	if pref == nil {
		return defaultVal
	}
	raw, ok := pref[key]
	if !ok {
		return defaultVal
	}
	if value, ok := raw.(string); ok && value != "" {
		return value
	}
	return defaultVal
}

func getBoolPreference(pref map[string]any, key string, defaultVal bool) bool {
	if pref == nil {
		return defaultVal
	}
	raw, ok := pref[key]
	if !ok {
		return defaultVal
	}
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		return value != "" && value != "false"
	default:
		return defaultVal
	}
}

func (ci *ConfiguredIntegration) IncidentDefaults() IncidentDefaults {
	defaults := getNestedPreferenceMap(ci.intg.UserPreferences, "incidentDefaults")
	return IncidentDefaults{
		AnnouncementChannelID:     getStringPreference(defaults, "announcementChannelId", ""),
		ChannelNamePattern:        getStringPreference(defaults, "channelNamePattern", "incident-{slug}"),
		DefaultSeverityID:         getStringPreference(defaults, "defaultSeverityId", ""),
		DefaultTypeID:             getStringPreference(defaults, "defaultTypeId", ""),
		AutoCreateVideoConference: getBoolPreference(defaults, "autoCreateVideoConference", false),
		InviteMode:                getStringPreference(defaults, "inviteMode", "assigned_users"),
	}
}

func (ci *ConfiguredIntegration) EnabledDataKinds() []string {
	return supportedDataKinds
}

func (ci *ConfiguredIntegration) ChatService(ctx context.Context) (rez.ChatService, error) {
	return newChatService(ci)
}

type IntegrationConfig struct {
	AccessToken      string
	TokenType        string
	BotUserID        string
	WebhookChannelId string
	Team             teamInfo
	Enterprise       *teamInfo
}

type teamInfo struct {
	ID   string
	Name string
}

func decodeConfig(raw json.RawMessage) (*IntegrationConfig, error) {
	var cfg IntegrationConfig
	if cfgErr := json.Unmarshal(raw, &cfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to decode integration config: %w", cfgErr)
	}
	return &cfg, nil
}
