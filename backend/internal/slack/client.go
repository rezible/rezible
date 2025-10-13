package slack

import (
	"fmt"
	"os"

	"github.com/slack-go/slack"
)

const (
	botTokenEnvVar = "SLACK_BOT_TOKEN"
	appTokenEnvVar = "SLACK_APP_TOKEN"

	enableSocketModeEnvVar = "SLACK_USE_SOCKETMODE"
)

func UseSocketMode() bool {
	return os.Getenv(enableSocketModeEnvVar) == "true"
}

func LoadClient() (*slack.Client, error) {
	botToken := os.Getenv(botTokenEnvVar)
	if botToken == "" {
		return nil, fmt.Errorf("%s environment variable not set", botTokenEnvVar)
	}

	appToken := os.Getenv(appTokenEnvVar)
	if appToken != "" && !UseSocketMode() {
		// TODO: check if socketmode enabled
	}

	client := slack.New(botToken,
		slack.OptionAppLevelToken(appToken))

	return client, nil
}
