package slack

import (
	"errors"

	rez "github.com/rezible/rezible"
	"github.com/slack-go/slack"
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
