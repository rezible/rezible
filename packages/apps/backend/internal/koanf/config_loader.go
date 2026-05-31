package koanf

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"

	rez "github.com/rezible/rezible"
)

type ConfigLoader struct {
	loader    *koanf.Koanf
	validator *validator.Validate
}

type ConfigLoaderOptions struct {
	LoadEnvironment bool
	Overrides       map[string]any
}

const (
	delim     = "."
	structTag = "cfg"
)

func NewConfigLoader(opts ConfigLoaderOptions) (*ConfigLoader, error) {
	k := koanf.New(delim)
	v := validator.New(validator.WithRequiredStructEnabled())

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

	cfg := &ConfigLoader{
		loader:    k,
		validator: v,
	}

	return cfg, nil
}

func (c *ConfigLoader) LoadConfig(ctx context.Context) (rez.Config, error) {
	cfg := rez.DefaultConfig()
	cfgErr := c.loader.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: structTag})
	if cfgErr != nil {
		return cfg, fmt.Errorf("unmarshal: %w", cfgErr)
	}

	if validationErr := c.validator.StructCtx(ctx, cfg); validationErr != nil {
		if errs, ok := errors.AsType[validator.ValidationErrors](validationErr); ok && len(errs) > 0 {
			msgs := make([]string, len(errs))
			for i, verr := range errs {
				msgs[i] = fmt.Sprintf("[%d: %s]", i, verr)
			}
			return cfg, fmt.Errorf("validation: %s", strings.Join(msgs, " "))
		}
	}
	return cfg, nil
}
