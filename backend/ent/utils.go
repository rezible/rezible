package ent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/rezible/rezible/ent/entpgx"
)

type ListParams struct {
	Search          string
	Offset          int
	Limit           int
	Count           bool
	IncludeArchived bool
	OrderAsc        bool
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

type CountableQuery[T any] interface {
	All(ctx context.Context) ([]T, error)
	Count(ctx context.Context) (int, error)
}

func RunCountableQuery[T any](ctx context.Context, q CountableQuery[T], doCount bool) ([]T, int, error) {
	var count int
	var queryErr error
	results := make([]T, 0)
	if doCount {
		count, queryErr = q.Count(ctx)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("count: %w", queryErr)
		}
	}
	if !doCount || count > 0 {
		results, queryErr = q.All(ctx)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("list: %w", queryErr)
		}
	}
	return results, count, nil
}
