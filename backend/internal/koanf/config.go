package koanf

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	k *koanf.Koanf
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

	//for key, val := range opts.Overrides {
	//	v.Set(key, val)
	//}

	return &Config{k: k}, nil
}

func (c *Config) Unmarshal(key string, v any) error {
	return c.k.Unmarshal(key, v)
}

func (c *Config) GetString(key string) string {
	return c.k.String(key)
}

func (c *Config) GetStringOr(key string, orDefault string) string {
	if c.k.Exists(key) {
		return c.k.String(key)
	}
	return orDefault
}

func (c *Config) GetStrings(key string) []string {
	return c.k.Strings(key)
}

func (c *Config) GetBool(key string) bool {
	return c.k.Bool(key)
}

func (c *Config) GetBoolOr(key string, orDefault bool) bool {
	if c.k.Exists(key) {
		return c.k.Bool(key)
	}
	return orDefault
}

func (c *Config) GetDuration(key string) time.Duration {
	return c.k.Duration(key)
}

func (c *Config) GetDurationOr(key string, orDefault time.Duration) time.Duration {
	if c.k.Exists(key) {
		return c.k.Duration(key)
	}
	return orDefault
}

func (c *Config) DebugMode() bool {
	return c.GetBool("debug_mode")
}

func (c *Config) DataSyncMode() bool {
	return c.GetBool("datasync_mode")
}

func (c *Config) ServeFrontend() bool {
	return c.GetBool("serve_frontend")
}

func (c *Config) ListenHost() string {
	return c.GetStringOr("host", "0.0.0.0")
}

func (c *Config) ListenPort() string {
	return c.GetStringOr("port", "7002")
}

func ensureSlashPrefix(s string) string {
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	return s
}

func (c *Config) ApiPath() string {
	return ensureSlashPrefix(c.GetStringOr("api_path", "/api"))
}

func (c *Config) WebhooksPath() string {
	return ensureSlashPrefix(c.GetStringOr("webhooks_path", "/webhooks"))
}

func (c *Config) AppUrl() string {
	return c.GetString("app_url")
}

// TODO: tighten these up

func (c *Config) SingleTenantMode() bool {
	return c.GetBool("single_tenant_mode")
}

func (c *Config) AllowUserCreation() bool {
	return !c.SingleTenantMode()
}

func (c *Config) AllowTenantCreation() bool {
	return !c.SingleTenantMode() && !c.GetBool("disable_tenant_creation")
}
