package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezible/rezible/ent"
	"github.com/samber/do/v2"
)

func PackageContext(ctx context.Context, i do.Injector) {
	do.Package(
		do.Lazy(func(i do.Injector) (*DatabaseClient, error) {
			return NewDatabaseClient(ctx)
		}),
		do.Lazy(func(i do.Injector) (*ent.Client, error) {
			return do.MustInvoke[*DatabaseClient](i).Client(), nil
		}),
		do.Lazy(func(i do.Injector) (*pgxpool.Pool, error) {
			return do.MustInvoke[*DatabaseClient](i).Pool(), nil
		}),
	)(i)
}
