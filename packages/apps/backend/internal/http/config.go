package http

import (
	"fmt"
	"net/url"

	rez "github.com/rezible/rezible"
)

type Config struct {
	Host           string               `koanf:"host"`
	Port           string               `koanf:"port"`
	BasePath       string               `koanf:"base_path"`
	DocumentsProxy DocumentsProxyConfig `koanf:"documents_proxy"`
}

type DocumentsProxyConfig struct {
	Enabled   bool   `koanf:"enabled"`
	ProxyHost string `koanf:"proxy_host"`
	serverUrl *url.URL
}

func loadConfig() (Config, error) {
	cfg := Config{
		Host:     rez.Config.GetString("HOST", "0.0.0.0"),
		Port:     rez.Config.GetString("PORT", "7002"),
		BasePath: "",
		DocumentsProxy: DocumentsProxyConfig{
			Enabled:   false,
			ProxyHost: "localhost:7003",
		},
	}
	if cfgErr := rez.Config.Unmarshal("server.http", &cfg); cfgErr != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", cfgErr)
	}
	if cfg.DocumentsProxy.Enabled {
		proxyUrl, parseErr := url.Parse("ws://" + cfg.DocumentsProxy.ProxyHost)
		if parseErr != nil {
			return cfg, fmt.Errorf("failed to parse documents_proxy.proxy_host: %w", parseErr)
		}
		cfg.DocumentsProxy.serverUrl = proxyUrl
	}

	return cfg, nil
}
