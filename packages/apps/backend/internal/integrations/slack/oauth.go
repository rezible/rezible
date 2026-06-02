package slackintegration

import (
	"fmt"
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"
)

type oauthHandler struct {
	cfg *oauth2.Config
}

func NewOAuthHandler(clientId string, clientSecret string, scopes []string) *oauthHandler {
	return &oauthHandler{
		cfg: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://slack.com/oauth/v2/authorize",
				TokenURL: "https://slack.com/api/oauth.v2.access",
			},
			Scopes: scopes,
		},
	}
}

func (h *oauthHandler) OAuth2Config() *oauth2.Config {
	return h.cfg
}

func (h *oauthHandler) validateOAuthTokenScopes(t *oauth2.Token) error {
	tokenScope, scopeOk := t.Extra("scope").(string)
	if !scopeOk {
		return fmt.Errorf("missing or invalid scope")
	}
	tokenScopes := mapset.NewSet(strings.Split(tokenScope, ",")...)
	for _, scope := range h.cfg.Scopes {
		if !tokenScopes.Contains(scope) {
			return fmt.Errorf("missing token scope: %s", scope)
		}
	}
	return nil
}

func (h *oauthHandler) getTeamInfoFromTokenData(tokenData any) (*TeamInfo, error) {
	fmt.Printf("token data team info: %+v\n", tokenData)
	data, ok := tokenData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid token data")
	}
	var teamInfo TeamInfo
	if decodeErr := mapstructure.Decode(data, &teamInfo); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode: %v", decodeErr)
	}
	if teamInfo.Id == "" {
		return nil, nil
	}
	return &teamInfo, nil
}

func (h *oauthHandler) ExtractInstallationTargetFromToken(t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	if scopesErr := h.validateOAuthTokenScopes(t); scopesErr != nil {
		return nil, scopesErr
	}
	if t.TokenType != "bot" {
		return nil, fmt.Errorf("invalid token type: %s", t.TokenType)
	}

	cfg := &InstallationConfig{
		AccessToken: t.AccessToken,
	}

	isEnterpriseInstall, castOk := t.Extra("is_enterprise_install").(bool)
	if castOk {
		cfg.IsEnterpriseInstall = isEnterpriseInstall
	}

	botUserId, botUserIdOk := t.Extra("bot_user_id").(string)
	if !botUserIdOk {
		return nil, fmt.Errorf("missing or invalid bot_user_id")
	}
	cfg.BotUserID = botUserId

	team, teamErr := h.getTeamInfoFromTokenData(t.Extra("team"))
	if teamErr != nil {
		if !cfg.IsEnterpriseInstall {
			return nil, fmt.Errorf("invalid team info: %v", teamErr)
		}
	}
	cfg.Team = team

	enterprise, enterpriseErr := h.getTeamInfoFromTokenData(t.Extra("enterprise"))
	if enterpriseErr != nil {
		slog.Warn("slack get enterprise info from token", "error", enterpriseErr)
	}
	cfg.Enterprise = enterprise

	wh := objx.New(t.Extra("incoming_webhook"))
	if channelId := wh.Get("channel_id"); channelId.IsStr() {
		cfg.WebhookChannelId = channelId.String()
	}

	target, targetErr := MakeConfigInstallationTarget(cfg)
	if targetErr != nil {
		return nil, fmt.Errorf("failed to create target: %v", targetErr)
	}

	return []rez.IntegrationInstallationTarget{*target}, nil
}
