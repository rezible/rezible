package koanf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"

	rez "github.com/rezible/rezible"
)

type ConfigLoader struct {
	loader    *koanf.Koanf
	validator *validator.Validate
	opts      ConfigLoaderOptions
}

type ConfigLoaderOptions struct {
	LoadEnvironment     bool
	Overrides           map[string]any
	SkipValidation      bool
	LogValidationErrors bool
}

const (
	delim     = "."
	structTag = "cfg"
)

func NewConfigLoader(opts ConfigLoaderOptions) *ConfigLoader {
	return &ConfigLoader{
		loader:    koanf.New(delim),
		validator: validator.New(validator.WithRequiredStructEnabled()),
		opts:      opts,
	}
}

func (c *ConfigLoader) LoadConfig(ctx context.Context) (rez.Config, error) {
	cfg := rez.DefaultConfig()

	if c.opts.LoadEnvironment {
		if envErr := c.loadEnvironment(); envErr != nil {
			return cfg, fmt.Errorf("failed to load env provider: %w", envErr)
		}
	}

	if c.opts.Overrides != nil {
		overrideLoader := koanf.New(delim)
		for k, v := range c.opts.Overrides {
			if ovrErr := overrideLoader.Set(k, v); ovrErr != nil {
				return cfg, fmt.Errorf("failed to set override (%s=%s): %w", k, v, ovrErr)
			}
		}
		if mergeErr := c.loader.Merge(overrideLoader); mergeErr != nil {
			return cfg, fmt.Errorf("failed to merge overrides: %w", mergeErr)
		}
	}

	cfgErr := c.loader.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: structTag})
	if cfgErr != nil {
		return cfg, fmt.Errorf("unmarshal: %w", cfgErr)
	}

	if c.opts.SkipValidation {
		return cfg, nil
	}

	if validationErr := c.validator.StructCtx(ctx, cfg); validationErr != nil {
		if errs, ok := errors.AsType[validator.ValidationErrors](validationErr); ok && len(errs) > 0 {
			msgs := make([]string, len(errs))
			for i, verr := range errs {
				msgs[i] = fmt.Sprintf("[%d: %s]", i, verr)
			}
			if c.opts.LogValidationErrors {
				slog.Error("validation errors:" + strings.Join(msgs, "; "))
			} else {
				return cfg, fmt.Errorf("validation: %s", strings.Join(msgs, " "))
			}
		}
	}
	return cfg, nil
}

func (c *ConfigLoader) loadEnvironment() error {
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
	return c.loader.Load(envProv, nil)
}
