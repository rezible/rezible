package slack

import (
	"errors"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
)

func UseSocketMode() bool {
	return rez.Config.GetBool("slack.socketmode.enabled")
}

func LoadClient() (*slack.Client, error) {
	botToken := rez.Config.GetString("slack.bot_token")
	if botToken == "" {
		return nil, errors.New("slack.bot_token not set")
	}

	appToken := rez.Config.GetString("slack.app_token")
	if appToken != "" && !UseSocketMode() {
		return nil, errors.New("slack.app_token not set")
	}

	client := slack.New(botToken,
		slack.OptionAppLevelToken(appToken))

	return client, nil
}

func LoadOAuthConfig() *oauth2.Config {
	//appId := rez.Config.GetString("slack.app_id")
	clientId := rez.Config.GetString("slack.oauth_client_id")
	clientSecret := rez.Config.GetString("slack.oauth_client_secret")
	scopes := []string{"app_mentions:read",
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
