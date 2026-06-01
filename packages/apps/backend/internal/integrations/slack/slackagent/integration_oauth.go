package slackagent

import (
	"fmt"
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"
)

func (i *Integration) OAuthConfigRequired() bool {
	return true
}

func (i *Integration) OAuth2Config() *oauth2.Config {
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

func makeOAuth2Config(cfg rez.IntegrationsConfigSlack) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.OAuthClientId,
		ClientSecret: cfg.OAuthClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/v2/authorize",
			TokenURL: "https://slack.com/api/oauth.v2.access",
		},
		Scopes: oAuthScopes,
	}
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

func (i *Integration) ExtractIntegrationOptionsFromToken(t *oauth2.Token) ([]rez.ExternalIntegrationOption, error) {
	if scopesErr := validateOauthTokenScopes(t); scopesErr != nil {
		return nil, scopesErr
	}
	if isEnterprise, ok := t.Extra("is_enterprise_install").(bool); ok && isEnterprise {
		return nil, fmt.Errorf("slack enterprise org-wide installs are not supported yet")
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
		slog.Warn("slack get enterprise info from token", "error", enterpriseErr)
	} else {
		cfg.Set(configEnterprise, enterprise)
	}

	wh := objx.New(t.Extra("incoming_webhook"))
	if channelId := wh.Get("channel_id"); channelId.IsStr() {
		cfg.Set(configWebhookChannelId, channelId.String())
	}

	teamID, _ := team["id"].(string)
	teamName, _ := team["name"].(string)
	return []rez.ExternalIntegrationOption{{
		ExternalRef: teamID,
		DisplayName: teamName,
		Config:      cfg,
	}}, nil
}
