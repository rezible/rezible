package river

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/samber/do/v2"
)

func ProvideJobService(ctx context.Context, inj do.Injector) (rez.JobsService, error) {
	return NewJobService(ctx, do.MustInvoke[*pgxpool.Pool](inj))
}
