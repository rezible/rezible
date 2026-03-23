package slack

import (
	"errors"

	rez "github.com/rezible/rezible"
)

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
