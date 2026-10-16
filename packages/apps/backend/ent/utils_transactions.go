package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
)

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

func WithLockStrength(strength sql.LockStrength) func(*sql.LockOptions) {
	return func(opts *sql.LockOptions) {
		opts.Strength = strength
	}
}

// KeyShareLock makes a ForShare lock FOR KEY SHARE. Use it for a reader that must block a FOR UPDATE of the row,
// such as one that deletes it, without blocking other readers that also update the row's non-key columns.
var KeyShareLock = WithLockStrength(sql.LockKeyShare)

// KeyUpdateLock makes a ForUpdate lock FOR NO KEY UPDATE. Use it to serialize updates of a row's non-key columns
// without blocking KeyShareLock holders.
var KeyUpdateLock = WithLockStrength(sql.LockNoKeyUpdate)
