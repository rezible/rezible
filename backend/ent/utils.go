package ent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
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

func (p ListParams) GetOrder() entsql.OrderTermOption {
	if p.OrderAsc {
		return entsql.OrderAsc()
	}
	return entsql.OrderDesc()
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

type ListResult[T any] struct {
	Data  []T
	Count int
}

type listQuery[T any, Q any] interface {
	All(ctx context.Context) ([]T, error)
	Count(ctx context.Context) (int, error)
	Limit(limit int) Q
	Offset(offset int) Q
}

func DoListQuery[T any, Q any](ctx context.Context, query listQuery[T, Q], p ListParams) (ListResult[T], error) {
	res := ListResult[T]{
		Data:  make([]T, 0),
		Count: 0,
	}
	ctx = p.GetQueryContext(ctx)
	if p.Count {
		count, queryErr := query.Count(ctx)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return res, fmt.Errorf("count: %w", queryErr)
		}
		res.Count = count
	}
	if !p.Count || res.Count > 0 {
		query.Offset(p.Offset)
		query.Limit(p.GetLimit())
		results, queryErr := query.All(ctx)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return res, fmt.Errorf("list: %w", queryErr)
		}
		res.Data = results
	}
	return res, nil
}
