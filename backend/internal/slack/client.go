package slack

import (
	"context"
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"golang.org/x/oauth2"
)

type loader struct {
	svcs *rez.Services
}

func newLoader(svcs *rez.Services) *loader {
	return &loader{svcs: svcs}
}

func (l *loader) loadWithIntegration(intg *ent.Integration) (*ChatService, error) {
	cfg, cfgErr := decodeConfig(intg.Config)
	if cfgErr != nil {
		return nil, fmt.Errorf("unable to decode config: %w", cfgErr)
	}
	return newChatService(cfg.makeClient(), l.svcs), nil
}

func (l *loader) loadByTenantLookup(ctx context.Context, teamId string, enterpriseId string) (*ChatService, context.Context, error) {
	vals := make(map[string]any)
	if teamId != "" {
		vals["Team.ID"] = teamId
	}
	if enterpriseId != "" {
		vals["Enterprise.ID"] = enterpriseId
	}
	intg, lookupErr := l.svcs.Integrations.LookupByConfigValues(access.SystemContext(ctx), integrationName, vals)
	if lookupErr != nil {
		return nil, nil, lookupErr
	}
	tenantCtx := access.TenantContext(ctx, intg.TenantID)
	chat, chatErr := l.loadWithIntegration(intg)
	if chatErr != nil {
		return nil, nil, fmt.Errorf("load chat service failed: %w", chatErr)
	}
	return chat, tenantCtx, nil
}

func (l *loader) loadFromContext(ctx context.Context) (*ChatService, error) {
	intg, lookupErr := l.svcs.Integrations.Get(ctx, integrationName)
	if lookupErr != nil {
		return nil, lookupErr
	}
	return l.loadWithIntegration(intg)
}

func LoadOAuthConfig() *oauth2.Config {
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
