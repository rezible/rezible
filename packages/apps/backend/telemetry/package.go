package telemetry

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

func PackageContext(ctx context.Context, i do.Injector) {
	do.Package(
		do.Lazy(func(i do.Injector) (rez.TelemetryService, error) {
			cfg, cfgErr := loadConfig(do.MustInvoke[rez.ConfigLoader](i))
			if cfgErr != nil {
				return nil, fmt.Errorf("failed to load config: %w", cfgErr)
			}
			return newOpenTelemetryService(ctx, cfg)
		}),
	)(i)
}
