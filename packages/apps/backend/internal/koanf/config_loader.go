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
	opts      Options
}

type Options struct {
	LoadEnvironment     bool
	Overrides           map[string]any
	SkipValidation      bool
	LogValidationErrors bool
}

const (
	delim     = "."
	structTag = "cfg"
)

func LoadConfig(ctx context.Context, opts Options) (*rez.Config, error) {
	return NewConfigLoader(opts).LoadConfig(ctx)
}

func NewConfigLoader(opts Options) *ConfigLoader {
	return &ConfigLoader{
		loader:    koanf.New(delim),
		validator: validator.New(validator.WithRequiredStructEnabled()),
		opts:      opts,
	}
}

func (c *ConfigLoader) LoadConfig(ctx context.Context) (*rez.Config, error) {
	cfg := rez.DefaultConfig()

	if c.opts.LoadEnvironment {
		if envErr := c.loadEnvironment(); envErr != nil {
			return nil, fmt.Errorf("failed to load env provider: %w", envErr)
		}
	}

	if c.opts.Overrides != nil {
		overrideLoader := koanf.New(delim)
		for k, v := range c.opts.Overrides {
			if ovrErr := overrideLoader.Set(k, v); ovrErr != nil {
				return nil, fmt.Errorf("failed to set override (%s=%s): %w", k, v, ovrErr)
			}
		}
		if mergeErr := c.loader.Merge(overrideLoader); mergeErr != nil {
			return nil, fmt.Errorf("failed to merge overrides: %w", mergeErr)
		}
	}

	cfgErr := c.loader.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: structTag})
	if cfgErr != nil {
		return nil, fmt.Errorf("unmarshal: %w", cfgErr)
	}

	if !c.opts.SkipValidation {
		if validationErr := c.validateConfig(ctx, cfg); validationErr != nil {
			return nil, fmt.Errorf("failed to validate config:\n%w", validationErr)
		}
	}

	return &cfg, nil
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

func (c *ConfigLoader) validateConfig(ctx context.Context, cfg rez.Config) error {
	var validationErrs []error
	if validationErr := c.validator.StructCtx(ctx, cfg); validationErr != nil {
		if errs, ok := errors.AsType[validator.ValidationErrors](validationErr); ok && len(errs) > 0 {
			for _, verr := range errs {
				if c.opts.LogValidationErrors {
					slog.Error("field error:" + verr.Error())
				} else {
					validationErrs = append(validationErrs, verr)
				}
			}
		}
	}
	return errors.Join(validationErrs...)
}
