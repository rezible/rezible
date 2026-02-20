package postgres

import (
	"database/sql"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
)

func NewTestDB(t *testing.T) *sql.DB {
	t.Parallel()
	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		User:       "postgres",
		Password:   "password",
		Host:       "localhost",
		Port:       "5433",
		Options:    "sslmode=disable",
	}
	gm := golangmigrator.New("migrations")
	return pgtestdb.New(t, pgxConf, gm)
}
