package db

import (
	"context"

	rez "github.com/rezible/rezible"
	_ "github.com/rezible/rezible/ent/runtime"
)

type DatabaseListener struct {
	dbc rez.Database
}

func NewListener(dbc rez.Database) DatabaseListener {
	return DatabaseListener{dbc: dbc}
}

func (l DatabaseListener) Start(ctx context.Context) error {
	return nil
}

func (l DatabaseListener) Stop(ctx context.Context) error {
	// log.Debug().Msg("Closing database connection")
	return l.dbc.Close()
}
