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
	Page            int
	PageSize        int
	IncludeArchived bool
	OrderAsc        bool
}

func (p ListParams) GetPage() int {
	if p.Page < 1 {
		return 1
	}
	return p.Page
}

func (p ListParams) GetPageSize() int {
	if p.PageSize < 1 {
		return 25
	}
	return p.PageSize
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

type transactionRunner interface {
	WithTx(context.Context, func(context.Context, *Client) error, ...TxOption) error
}

// WithTxReturning runs fn using db.WithTx and returns its value only when
// WithTx succeeds. On error it returns the zero value of T.
// It preserves entity transaction bindings; it does not unwrap entities.
// When joining an existing transaction, success does not commit that transaction.
func WithTxReturning[T any](
	ctx context.Context,
	db transactionRunner,
	fn func(context.Context, *Client) (T, error),
	opts ...TxOption,
) (T, error) {
	var result T
	txFn := func(txCtx context.Context, tx *Client) error {
		var callbackErr error
		result, callbackErr = fn(txCtx, tx)
		return callbackErr
	}
	if txErr := db.WithTx(ctx, txFn, opts...); txErr != nil {
		var zero T
		return zero, txErr
	}
	return result, nil
}

type TxOptions struct {
	OnCommit   []CommitHook
	OnRollback []RollbackHook
}

func WithCommitHook(h CommitHook) TxOption {
	return func(opts *TxOptions) {
		opts.OnCommit = append(opts.OnCommit, h)
	}
}

func WithAfterCommitFunc(fn func()) TxOption {
	return WithCommitHook(func(next Committer) Committer {
		return CommitFunc(func(ctx context.Context, tx *Tx) error {
			if err := next.Commit(ctx, tx); err != nil {
				return err
			}
			fn()
			return nil
		})
	})
}

func WithRollbackHook(h RollbackHook) TxOption {
	return func(opts *TxOptions) {
		opts.OnRollback = append(opts.OnRollback, h)
	}
}

type EntityMutator[T any, M ent.Mutation] interface {
	Save(context.Context) (T, error)
	Exec(context.Context) error
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
	Data     []*T
	Page     int
	PageSize int
	Total    int
}

type listQuery[T any, Q any] interface {
	All(ctx context.Context) ([]*T, error)
	Count(ctx context.Context) (int, error)
	Limit(limit int) Q
	Offset(offset int) Q
}

func DoListQuery[T any, Q any](ctx context.Context, query listQuery[T, Q], p ListParams) (*ListResult[T], error) {
	page := p.GetPage()
	pageSize := p.GetPageSize()
	res := &ListResult[T]{
		Data:     make([]*T, 0),
		Page:     page,
		PageSize: pageSize,
	}
	ctx = p.GetQueryContext(ctx)
	count, queryErr := query.Count(ctx)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("count: %w", queryErr)
	}
	res.Total = count
	if res.Total == 0 {
		return res, nil
	}
	query.Offset((page - 1) * pageSize)
	query.Limit(pageSize)
	results, queryErr := query.All(ctx)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("list: %w", queryErr)
	}
	res.Data = results
	return res, nil
}
