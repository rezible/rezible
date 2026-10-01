package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgconn"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/tenant"
	"github.com/rezible/rezible/ent/user"
	dbservices "github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type DatabaseClientSuite struct {
	test.Suite
}

func TestDatabaseClientSuite(t *testing.T) {
	suite.Run(t, &DatabaseClientSuite{Suite: test.NewSuite()})
}

func (s *DatabaseClientSuite) createTenant(ctx context.Context, client *ent.Client) error {
	_, err := client.Tenant.Create().Save(ctx)
	return err
}

func (s *DatabaseClientSuite) requireCreateTenant(ctx context.Context, tdb rez.Database) {
	client := tdb.Client(ctx)
	_, err := client.Tenant.Create().Save(ctx)
	s.Require().NoError(err)
}

func (s *DatabaseClientSuite) tenantCount(ctx context.Context, tdb rez.Database) int {
	client := tdb.Client(ctx)
	count, err := client.Tenant.Query().Count(ctx)
	s.Require().NoError(err)
	return count
}

func (s *DatabaseClientSuite) TestClientOutsideTransaction() {
	ctx := execution.NewSystemContext(s.T().Context())
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	client := tdb.Client(ctx)

	s.NotNil(client)
	s.NotPanics(func() {
		_ = s.tenantCount(ctx, tdb)
	})
}

func (s *DatabaseClientSuite) TestWithTxCommits() {
	ctx := execution.NewSystemContext(s.T().Context())
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())

	before := s.tenantCount(ctx, tdb)
	s.Require().NoError(tdb.WithTx(ctx, s.createTenant))
	s.Equal(before+1, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestWithTxRollsBackOnError() {
	ctx := execution.NewSystemContext(s.T().Context())
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	before := s.tenantCount(ctx, tdb)

	expectedErr := errors.New("force rollback")
	var createdID int
	txErr := tdb.WithTx(ctx, func(txCtx context.Context, client *ent.Client) error {
		created, createErr := client.Tenant.Create().Save(txCtx)
		if createErr != nil {
			return fmt.Errorf("create tenant before rollback: %w", createErr)
		}
		createdID = created.ID
		queryTenant := client.Tenant.Query().
			Where(tenant.ID(createdID))
		exists, queryErr := queryTenant.Exist(txCtx)
		if queryErr != nil {
			return fmt.Errorf("read tenant inside transaction: %w", queryErr)
		}
		if !exists {
			return errors.New("created tenant is not visible inside transaction")
		}
		return expectedErr
	})
	s.ErrorIs(txErr, expectedErr)
	s.Equal(before, s.tenantCount(ctx, tdb))
	s.Require().NotZero(createdID)
	queryTenant := tdb.Client(ctx).Tenant.Query().
		Where(tenant.ID(createdID))
	exists, queryErr := queryTenant.Exist(ctx)
	s.Require().NoError(queryErr)
	s.False(exists, "rolled-back tenant must not persist")
}

func (s *DatabaseClientSuite) TestNestedWithTxSharesOuterTransaction() {
	ctx := execution.NewSystemContext(s.T().Context())

	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	before := s.tenantCount(ctx, tdb)

	s.Require().NoError(tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(txCtx, tdb)

		nestedErr := tdb.WithTx(txCtx, func(nestedCtx context.Context, _ *ent.Client) error {
			s.requireCreateTenant(nestedCtx, tdb)
			s.Equal(before+2, s.tenantCount(nestedCtx, tdb))
			return nil
		})
		if nestedErr != nil {
			return nestedErr
		}

		s.Equal(before+2, s.tenantCount(txCtx, tdb))
		return nil
	}))

	s.Equal(before+2, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestNestedWithTxErrorRollsBackOuterTransaction() {
	expectedErr := errors.New("nested rollback")
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	nestedTx := func(ctx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(ctx, tdb)
		return expectedErr
	}
	outerTx := func(ctx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(ctx, tdb)
		return tdb.WithTx(ctx, nestedTx)
	}

	ctx := execution.NewSystemContext(s.T().Context())
	before := s.tenantCount(ctx, tdb)
	err := tdb.WithTx(ctx, outerTx)
	s.ErrorIs(err, expectedErr)
	s.Equal(before, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestWithTxCommitHook() {
	ctx := execution.NewSystemContext(s.T().Context())
	committed := false
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	hook := func(next ent.Committer) ent.Committer {
		return ent.CommitFunc(func(ctx context.Context, tx *ent.Tx) error {
			if err := next.Commit(ctx, tx); err != nil {
				return err
			}
			committed = true
			return nil
		})
	}

	err := tdb.WithTx(ctx, s.createTenant, ent.WithCommitHook(hook))

	s.Require().NoError(err)
	s.True(committed)
}

func (s *DatabaseClientSuite) TestWithTxRollbackHook() {
	ctx := execution.NewSystemContext(s.T().Context())
	rolledBack := false
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	hook := func(next ent.Rollbacker) ent.Rollbacker {
		return ent.RollbackFunc(func(ctx context.Context, tx *ent.Tx) error {
			if err := next.Rollback(ctx, tx); err != nil {
				return err
			}
			rolledBack = true
			return nil
		})
	}

	expectedErr := errors.New("force rollback")
	txFn := func(txCtx context.Context, _ *ent.Client) error {
		return expectedErr
	}

	err := tdb.WithTx(ctx, txFn, ent.WithRollbackHook(hook))

	s.ErrorIs(err, expectedErr)
	s.True(rolledBack)
}

func (s *DatabaseClientSuite) TestWithTxRollsBackOnPanic() {
	ctx := execution.NewSystemContext(s.T().Context())
	_, tdb := s.SetupTestDatabase(test.WithFreshDatabase())

	panicErr := "expected"
	panicFn := func() {
		_ = tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
			s.requireCreateTenant(txCtx, tdb)
			panic(panicErr)
		})
	}

	before := s.tenantCount(ctx, tdb)
	s.PanicsWithValue(panicErr, panicFn)
	s.Equal(before, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestAcquireTxLocksRequiresTransaction() {
	ctx, tdb := s.SetupTestDatabase(test.WithFreshDatabase())

	err := tdb.AcquireTxLocks(ctx, "test", "a")

	s.Require().Error(err)
	s.Contains(err.Error(), "no active transaction")
}

func (s *DatabaseClientSuite) TestAcquireTxLocksWorksInsideTransaction() {
	ctx, tdb := s.SetupTestDatabase(test.WithFreshDatabase())

	err := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		return tdb.AcquireTxLocks(txCtx, "test", "b", "a", "a")
	})

	s.Require().NoError(err)
}

func (s *DatabaseClientSuite) TestConcurrentTxLockRequestsWithOppositeInputOrderSucceed() {
	tenantCtx, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	ctx, cancel := context.WithTimeout(tenantCtx, 10*time.Second)
	defer cancel()
	results := make(chan error, 2)
	lockFn := func(keys ...string) {
		transactionErr := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
			return tdb.AcquireTxLocks(txCtx, "test", keys...)
		})
		results <- transactionErr
	}

	go lockFn("b", "a")
	go lockFn("a", "b")

	for range 2 {
		select {
		case transactionErr := <-results:
			s.Require().NoError(transactionErr)
		case <-ctx.Done():
			s.FailNow("concurrent transaction lock requests did not finish", ctx.Err().Error())
		}
	}
}

func (s *DatabaseClientSuite) TestIsTransientErrorClassifiesPostgresConcurrencyErrors() {
	tdb := postgres.DatabaseClient{}
	s.True(tdb.IsTransientError(&pgconn.PgError{Code: "40P01"}))
	s.True(tdb.IsTransientError(fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "40001"})))
	s.False(tdb.IsTransientError(&pgconn.PgError{Code: "23505"}))
	s.False(tdb.IsTransientError(errors.New("plain error")))
}

func (s *DatabaseClientSuite) TestWithTxReturningNestedServiceRollback() {
	ctx, tdb := s.SetupTestDatabase(test.WithFreshDatabase())
	txCtx, cancelTx := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTx()

	users, serviceErr := dbservices.NewUserService(tdb, nil)
	s.Require().NoError(serviceErr)

	expectedErr := errors.New("force outer transaction rollback")
	var createdID uuid.UUID
	rolledBack := false

	result, txErr := ent.WithTxReturning(txCtx, tdb,
		func(ctx context.Context, tx *ent.Client) (*ent.User, error) {
			created, createErr := users.Set(ctx, uuid.Nil, func(m *ent.UserMutation) {
				m.SetEmail("tx-helper-" + uuid.NewString() + "@example.com")
				m.SetName("before nested update")
			})
			if createErr != nil {
				return nil, fmt.Errorf("nested user creation: %w", createErr)
			}
			if created == nil {
				return nil, errors.New("nested service returned no user")
			}
			createdID = created.ID

			// Deliberately entity-bound: the enclosing transaction is active.
			// This detects a service or helper unwrapping before outer commit.
			updateUser := created.Update().SetName("after nested update")
			if updateErr := updateUser.Exec(ctx); updateErr != nil {
				return nil, fmt.Errorf("update nested result in outer transaction: %w", updateErr)
			}

			current, queryErr := tdb.Client(ctx).User.Get(ctx, created.ID)
			if queryErr != nil {
				return nil, fmt.Errorf("read nested result in outer transaction: %w", queryErr)
			}
			if current.Name != "after nested update" {
				return nil, fmt.Errorf("nested update not visible: got name %q", current.Name)
			}

			// A non-nil callback result must be discarded when WithTx fails.
			return current, expectedErr
		},
		ent.WithRollbackHook(func(next ent.Rollbacker) ent.Rollbacker {
			return ent.RollbackFunc(func(ctx context.Context, tx *ent.Tx) error {
				if rollbackErr := next.Rollback(ctx, tx); rollbackErr != nil {
					return rollbackErr
				}
				rolledBack = true
				return nil
			})
		}),
	)
	cancelTx()

	s.Require().ErrorIs(txErr, expectedErr)
	s.Nil(result)
	s.True(rolledBack)
	s.Require().NotEqual(uuid.Nil, createdID)

	// Use the original nontransactional context to observe persisted state.
	queryUser := tdb.Client(ctx).User.Query().Where(user.ID(createdID))
	exists, queryErr := queryUser.Exist(ctx)
	s.Require().NoError(queryErr)
	s.False(exists, "nested writes must be rolled back with the outer transaction")
}
