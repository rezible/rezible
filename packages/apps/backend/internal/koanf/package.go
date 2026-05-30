package koanf

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

func PackageContext(ctx context.Context, i do.Injector) {
	do.Package(
		do.Lazy(func(i do.Injector) (rez.ConfigLoader, error) {
			opts := ConfigLoaderOptions{LoadEnvironment: true}
			return NewConfigLoader(ctx, opts)
		}),
	)(i)
}
