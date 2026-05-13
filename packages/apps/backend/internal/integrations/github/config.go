package github

import (
	"errors"
	"fmt"
)

type Config struct {
	Enabled       bool   `cfg:"enabled"`
	WebhookSecret string `cfg:"webhook_secret"`
	App           struct {
		AppID         int64  `cfg:"app_id"`
		ClientID      string `cfg:"client_id"`
		ClientSecret  string `cfg:"client_secret"`
		PrivateKeyPEM string `cfg:"private_key_pem"`
	} `cfg:"app"`
}

func (c Config) validate() error {
	var errs []error
	if c.App.AppID == 0 {
		errs = append(errs, fmt.Errorf("github.app.app_id not set"))
	}
	if c.App.ClientID == "" {
		errs = append(errs, fmt.Errorf("github.app.client_id not set"))
	}
	if c.App.ClientSecret == "" {
		errs = append(errs, fmt.Errorf("github.app.client_secret not set"))
	}
	if c.App.PrivateKeyPEM == "" {
		errs = append(errs, fmt.Errorf("github.app.private_key_pem not set"))
	}
	return errors.Join(errs...)
}
