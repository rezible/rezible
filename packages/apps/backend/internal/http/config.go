package http

import (
	"fmt"
	"net/url"

	rez "github.com/rezible/rezible"
)

type Config struct {
	Host           string               `cfg:"host"`
	Port           string               `cfg:"port"`
	BasePath       string               `cfg:"base_path"`
	DocumentsProxy DocumentsProxyConfig `cfg:"documents_proxy"`
}

type DocumentsProxyConfig struct {
	Enabled   bool   `cfg:"enabled"`
	ProxyHost string `cfg:"proxy_host"`
	serverUrl *url.URL
}

func loadConfig(cl rez.ConfigLoader) (Config, error) {
	cfg := Config{
		Host:     cl.GetString("HOST", "0.0.0.0"),
		Port:     cl.GetString("PORT", "7002"),
		BasePath: "",
		DocumentsProxy: DocumentsProxyConfig{
			Enabled:   false,
			ProxyHost: "localhost:7003",
		},
	}
	if cfgErr := cl.Unmarshal("server.http", &cfg); cfgErr != nil {
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
