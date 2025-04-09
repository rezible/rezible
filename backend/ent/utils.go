package ent

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/rezible/rezible/ent/entpgx"
)

type ListParams struct {
	Search          string
	Offset          int
	Limit           int
	IncludeArchived bool
}

func (p ListParams) GetLimit() int {
	if p.Limit == 0 {
		return 10
	}
	return p.Limit
}

func (p ListParams) GetQueryContext(parent context.Context) context.Context {
	if p.IncludeArchived {
		return context.WithValue(parent, "include_archived", true)
	}
	return parent
}

func WithTx(ctx context.Context, client *Client, fn func(tx *Tx) error) error {
	tx, txErr := client.Tx(ctx)
	if txErr != nil {
		return txErr
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if fnErr := fn(tx); fnErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			fnErr = fmt.Errorf("%w: rolling back transaction: %v", fnErr, rbErr)
		}
		return fnErr
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("committing transaction: %w", commitErr)
	}
	return nil
}

func ExtractPgxTx(txClient *Tx) (pgx.Tx, error) {
	// extract pgx transaction from driver (hacky but eh)
	txDrv, drvOk := txClient.config.driver.(*txDriver)
	if !drvOk {
		return nil, errors.New("ent: pgx.Tx does not support driver")
	}
	pgxDrvTx, pgOk := txDrv.tx.(*entpgx.EntPgxPoolTx)
	if !pgOk {
		return nil, errors.New("ent: pgx.Tx does not support driver")
	}
	return pgxDrvTx.PGXTransaction(), nil
}
