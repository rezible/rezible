package river

import (
	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(func(i do.Injector) (rez.JobsService, error) {
		return NewJobService(do.MustInvoke[rez.TelemetryService](i), do.MustInvoke[*pgxpool.Pool](i))
	}),
)
