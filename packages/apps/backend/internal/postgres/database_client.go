package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/entpgx"
	"github.com/rezible/rezible/ent/hook"
	_ "github.com/rezible/rezible/ent/runtime"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
)

// foreignKeyViolation is the SQLSTATE of a foreign key violation.
const foreignKeyViolation = "23503"

type DatabaseClient struct {
	client *ent.Client
	pool   *ConnectionPool
}

type PgxPool = pgxpool.Pool

// NewPgxPoolDatabaseClient opens a client whose encrypted fields use encryptionKey.
func NewPgxPoolDatabaseClient(pool *ConnectionPool, encryptionKey []byte) (*DatabaseClient, error) {
	client, clientErr := newDatabaseClient(entpgx.NewPgxPoolDriver(pool.Pool), encryptionKey)
	if clientErr == nil {
		client.pool = pool
	}
	return client, clientErr
}

func newDatabaseClient(driver dialect.Driver, encryptionKey []byte) (*DatabaseClient, error) {
	if keyErr := schematypes.SetFieldEncryptionKey(encryptionKey); keyErr != nil {
		return nil, fmt.Errorf("set encryption key: %w", keyErr)
	}
	dbc := &DatabaseClient{}
	dbc.client = ent.NewClient(ent.Driver(driver))
	// The first hook runs outermost: references are canonical before the tenant reference check counts them.
	dbc.client.Use(dbc.ensureTenantIdSetHook(), dbc.canonicalKnowledgeReferencesHook(), dbc.tenantReferencesHook(), dbc.missingReferenceHook())
	dbc.client.Intercept(dbc.setTenantContextInterceptor())
	return dbc, nil
}

func (dbc *DatabaseClient) Client(ctx context.Context) *ent.Client {
	if tx := ent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return dbc.client
}

func (dbc *DatabaseClient) AcquireTxLocks(ctx context.Context, namespace string, keys ...string) error {
	if ent.TxFromContext(ctx) == nil {
		return fmt.Errorf("no active transaction")
	}
	if namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	tenantScope := "tenant:system"
	if tenantId, tenantOk := execution.GetContext(ctx).TenantID(); tenantOk {
		tenantScope = fmt.Sprintf("tenant:%d", tenantId)
	}

	lockKeys := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		lockKey := namespace + "\x1f" + tenantScope + "\x1f" + key
		if _, exists := seen[lockKey]; !exists {
			seen[lockKey] = struct{}{}
			lockKeys = append(lockKeys, lockKey)
		}
	}
	sort.Strings(lockKeys)

	for _, key := range lockKeys {
		if lockErr := ent.ExecTx(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1, 0))", key); lockErr != nil {
			return fmt.Errorf("acquire transaction lock: %w", lockErr)
		}
	}
	return nil
}

func (dbc *DatabaseClient) IsTransientError(err error) bool {
	if err == nil {
		return false
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code == "40P01" || pgErr.Code == "40001"
	}
	return false
}

func (dbc *DatabaseClient) IsConstraintError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok || pgErr.Code != "23505" {
		return "", false
	}
	return pgErr.ConstraintName, true
}

func (dbc *DatabaseClient) WithTx(ctx context.Context, fn func(txCtx context.Context, tx *ent.Client) error, opts ...ent.TxOption) error {
	if tx := ent.TxFromContext(ctx); tx != nil {
		applyTxOptions(tx, opts...)
		return fn(ctx, tx.Client())
	}

	tx, txErr := dbc.client.Tx(ctx)
	if txErr != nil {
		return fmt.Errorf("begin transaction: %w", txErr)
	}
	applyTxOptions(tx, opts...)

	finished := false
	defer func() {
		if v := recover(); v != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				panic(fmt.Errorf("%v: rollback transaction: %w", v, rbErr))
			}
			panic(v)
		} else if !finished {
			_ = tx.Rollback()
		}
	}()

	txCtx := ent.NewTxContext(ctx, tx)
	if fnErr := fn(txCtx, tx.Client()); fnErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w: rollback transaction: %w", fnErr, rbErr)
		}
		finished = true
		return fnErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit transaction: %w", commitErr)
	}
	finished = true
	return nil
}

func applyTxOptions(tx *ent.Tx, opts ...ent.TxOption) {
	txOpts := &ent.TxOptions{}
	for _, opt := range opts {
		opt(txOpts)
	}
	for _, hook := range txOpts.OnCommit {
		tx.OnCommit(hook)
	}
	for _, hook := range txOpts.OnRollback {
		tx.OnRollback(hook)
	}
}

func (dbc *DatabaseClient) RequireUpToDateMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(dbc.pool.Pool)
	defer closeDatabaseResource("migration startup check", db)
	_, pending, statusErr := migrations.Status(ctx, db)
	if statusErr != nil {
		return fmt.Errorf("get migration status: %w", statusErr)
	}
	if len(pending) > 0 {
		return fmt.Errorf("database migrations pending: %s", strings.Join(pending, ", "))
	}
	return nil
}

func closeDatabaseResource(name string, c io.Closer) {
	if closeErr := c.Close(); closeErr != nil && !errors.Is(closeErr, sql.ErrConnDone) {
		slog.Error("failed to close "+name, "error", closeErr)
	}
}

func (dbc *DatabaseClient) Shutdown() error {
	// The pool is owned by ConnectionPool. Ent's Close method closes its driver,
	// which would also stop River while it is still draining.
	return nil
}

func (dbc *DatabaseClient) setTenantContextInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(q ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			if tenantId, tenantIdSet := execution.GetContext(ctx).TenantID(); tenantIdSet {
				ctx = entsql.WithIntVar(ctx, "app.current_tenant", tenantId)
			}
			return q.Query(ctx, query)
		})
	})
}

func (dbc *DatabaseClient) ensureTenantIdSetHook() ent.Hook {
	type tenantedMutation interface {
		SetTenantID(int)
	}
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if tm, ok := m.(tenantedMutation); ok {
				tenantId, tenantIdSet := execution.GetContext(ctx).TenantID()
				if !tenantIdSet {
					return nil, errs.ErrTenantContextMissing
				}
				if setTenantId, alreadySet := m.Field("tenant_id"); alreadySet && setTenantId != tenantId {
					return nil, errs.ErrTenantContextMissing
				}
				tm.SetTenantID(tenantId)
			}
			return next.Mutate(ctx, m)
		})
	}
}

// canonicalKnowledgeReferencesHook makes a write naming a merged knowledge subject store its survivor, so call
// sites need not know about merges. It reads each subject FOR KEY SHARE, which holds off a merge only inside a
// transaction; outside one, a racing write can still store the merged subject, which stays readable.
func (dbc *DatabaseClient) canonicalKnowledgeReferencesHook() ent.Hook {
	type clientMutation interface {
		Client() *ent.Client
	}
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			for _, field := range ent.KnowledgeReferenceFields[m.Type()] {
				value, isSet := m.Field(field.Name)
				if !isSet {
					continue
				}
				id, isID := value.(uuid.UUID)
				if !isID {
					return nil, fmt.Errorf("canonicalize %s.%s: %T is not a UUID", m.Type(), field.Name, value)
				}
				resolve := ent.ResolveKnowledgeEntityID
				if field.Relationship {
					resolve = ent.ResolveKnowledgeRelationshipID
				}
				canonical, resolveErr := resolve(ctx, m.(clientMutation).Client(), id)
				if ent.IsNotFound(resolveErr) {
					// Left to the tenant reference check and the foreign key to report.
					continue
				} else if resolveErr != nil {
					return nil, fmt.Errorf("canonicalize %s.%s: %w", m.Type(), field.Name, resolveErr)
				}
				if canonical != id {
					if setErr := m.SetField(field.Name, canonical); setErr != nil {
						return nil, setErr
					}
				}
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// tenantReferencesHook rejects a write that sets or adds an edge to a missing row or one in another tenant, before
// the write runs. Archived rows and merged knowledge subjects count as present: neither is a tenant boundary.
func (dbc *DatabaseClient) tenantReferencesHook() ent.Hook {
	type tenantReferencesMutation interface {
		CheckTenantReferences(context.Context) error
	}
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if rm, ok := m.(tenantReferencesMutation); ok {
				if checkErr := rm.CheckTenantReferences(ent.IncludeMerged(ent.IncludeArchived(ctx))); checkErr != nil {
					if ent.IsNotFound(checkErr) {
						return nil, fmt.Errorf("%w: %w", errs.ErrNotFound, checkErr)
					}
					return nil, fmt.Errorf("check tenant references: %w", checkErr)
				}
			}
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

// missingReferenceHook reports a foreign key violation on a write as not found, keeping the violation in the
// chain; a delete of a row still referenced keeps its error.
func (dbc *DatabaseClient) missingReferenceHook() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			value, mutateErr := next.Mutate(ctx, m)
			if pgErr, ok := errors.AsType[*pgconn.PgError](mutateErr); ok && pgErr.Code == foreignKeyViolation {
				return nil, fmt.Errorf("%w: %w", errs.ErrNotFound, pgErr)
			}
			return value, mutateErr
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}
