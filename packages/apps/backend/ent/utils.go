package ent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"entgo.io/ent"
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

type TxOption func(*TxOptions)

type TxOptions struct {
	OnCommit   []CommitHook
	OnRollback []RollbackHook
}

func WithCommitHook(h CommitHook) TxOption {
	return func(opts *TxOptions) {
		opts.OnCommit = append(opts.OnCommit, h)
	}
}

func WithRollbackHook(h RollbackHook) TxOption {
	return func(opts *TxOptions) {
		opts.OnRollback = append(opts.OnRollback, h)
	}
}

type EntityMutator[T any, M ent.Mutation] interface {
	Save(context.Context) (T, error)
	Mutation() M
}

func ExtractPgxTx(txClient *Tx) (pgx.Tx, error) {
	// extract pgx transaction from driver (hacky but eh)
	txDrv, drvOk := txClient.config.driver.(*txDriver)
	if !drvOk {
		return nil, errors.New("ent: pgx.Tx does not support driver")
	}
	pgxDrvTx, pgOk := txDrv.tx.(*entpgx.PgxPoolTx)
	if !pgOk {
		return nil, errors.New("ent: pgx.Tx does not support driver")
	}
	return pgxDrvTx.PGXTransaction(), nil
}

func ExecTx(ctx context.Context, query string, args ...any) error {
	txClient := TxFromContext(ctx)
	if txClient == nil {
		return errors.New("ent: no transaction in context")
	}
	txDrv, drvOk := txClient.config.driver.(*txDriver)
	if !drvOk {
		return errors.New("ent: tx does not support driver")
	}
	return txDrv.Exec(ctx, query, args, nil)
}

type ListResult[T any] struct {
	Data  []*T
	Count int
}

type listQuery[T any, Q any] interface {
	All(ctx context.Context) ([]*T, error)
	Count(ctx context.Context) (int, error)
	Limit(limit int) Q
	Offset(offset int) Q
}

func DoListQuery[T any, Q any](ctx context.Context, query listQuery[T, Q], p ListParams) (*ListResult[T], error) {
	res := &ListResult[T]{
		Data:  make([]*T, 0),
		Count: 0,
	}
	ctx = p.GetQueryContext(ctx)
	if p.Count {
		count, queryErr := query.Count(ctx)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return nil, fmt.Errorf("count: %w", queryErr)
		}
		res.Count = count
	}
	if !p.Count || res.Count > 0 {
		query.Offset(p.Offset)
		query.Limit(p.GetLimit())
		results, queryErr := query.All(ctx)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return nil, fmt.Errorf("list: %w", queryErr)
		}
		res.Data = results
	}
	return res, nil
}
