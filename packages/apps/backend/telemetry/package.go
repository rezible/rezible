package telemetry

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

func PackageContext(ctx context.Context, i do.Injector) {
	do.Package(
		do.Lazy(func(i do.Injector) (rez.TelemetryService, error) {
			return NewOpenTelemetryService(ctx, do.MustInvoke[rez.Config](i))
		}),
	)(i)
}
