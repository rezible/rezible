package koanf

import (
	"context"
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

type BaseConfig struct {
	DebugMode        bool   `koanf:"debug_mode"`
	SingleTenantMode bool   `koanf:"single_tenant_mode"`
	AppUrl           string `koanf:"app_url"`
	ApiUrl           string `koanf:"api_url"`
}

type Config struct {
	k *koanf.Koanf

	baseCfg BaseConfig
}

type ConfigLoaderOptions struct {
	LoadEnvironment bool
	Overrides       map[string]any
}

const delim = "."

func NewConfigLoader(ctx context.Context, opts ConfigLoaderOptions) (*Config, error) {
	k := koanf.New(delim)

	if opts.LoadEnvironment {
		prefix := ""
		envProv := env.Provider(delim, env.Opt{
			TransformFunc: func(k, v string) (string, any) {
				k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, prefix)), "__", delim)
				if strings.Contains(v, " ") {
					return k, strings.Split(v, " ")
				}
				return k, v
			},
		})
		if envErr := k.Load(envProv, nil); envErr != nil {
			return nil, fmt.Errorf("failed to load env provider: %w", envErr)
		}
	}

	var baseCfg BaseConfig
	if cfgErr := k.Unmarshal("", &baseCfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to unmarshal app config: %w", cfgErr)
	}

	//for key, val := range opts.Overrides {
	//	v.Set(key, val)
	//}

	return &Config{k: k, baseCfg: baseCfg}, nil
}

func (c *Config) Exists(key string) bool {
	return c.k.Exists(key)
}

func (c *Config) Unmarshal(key string, v any) error {
	return c.k.Unmarshal(key, v)
}

func (c *Config) GetString(key string, fallback string) string {
	if c.Exists(key) {
		return c.k.String(key)
	}
	return fallback
}

func (c *Config) DebugMode() bool {
	return c.baseCfg.DebugMode
}

func (c *Config) ApiUrl() string {
	return c.baseCfg.ApiUrl
}

func (c *Config) AppUrl() string {
	return c.baseCfg.AppUrl
}

func (c *Config) SingleTenantMode() bool {
	return c.baseCfg.SingleTenantMode
}
