package oidc

import rez "github.com/rezible/rezible"

type Config struct {
	SessionSecret []byte     `koanf:"session_secret"`
	Oidc          oidcConfig `koanf:"oidc"`
}

type oidcConfig struct {
	Issuer    string           `koanf:"issuer"`
	AppClient oidcClientConfig `koanf:"app_client"`
}

type oidcClientConfig struct {
	Id          string   `koanf:"id"`
	Scopes      []string `koanf:"scopes"`
	RedirectUri string   `koanf:"redirect_uri"`
}

var defaultClientScopes = []string{"openid", "profile", "email", "offline_access"}

func loadConfig() (*Config, error) {
	cfg := Config{
		Oidc: oidcConfig{
			AppClient: oidcClientConfig{
				Scopes: defaultClientScopes,
			},
		},
	}
	if cfgErr := rez.Config.Unmarshal("auth", &cfg); cfgErr != nil {
		return nil, cfgErr
	}
	return &cfg, nil
}
