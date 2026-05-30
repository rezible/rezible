package koanf

import (
	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (*Config, error) {
		return NewConfigLoader(ConfigLoaderOptions{LoadEnvironment: true})
	}),
	do.Bind[*Config, rez.ConfigLoader](),
)
