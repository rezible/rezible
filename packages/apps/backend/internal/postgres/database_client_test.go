package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres"
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
	ctx := s.SystemContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)

	s.NotNil(client)
	s.NotPanics(func() {
		_ = s.tenantCount(ctx, tdb)
	})
}

func (s *DatabaseClientSuite) TestWithTxCommits() {
	ctx := s.SystemContext()
	tdb := s.CreateTestDatabase()

	before := s.tenantCount(ctx, tdb)
	s.Require().NoError(tdb.WithTx(ctx, s.createTenant))
	s.Equal(before+1, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestWithTxRollsBackOnError() {
	ctx := s.SystemContext()
	tdb := s.CreateTestDatabase()
	before := s.tenantCount(ctx, tdb)

	expectedErr := fmt.Errorf("force rollback")
	txErr := tdb.WithTx(ctx, func(txCtx context.Context, client *ent.Client) error {
		_ = s.createTenant(txCtx, client)
		return expectedErr
	})
	s.ErrorIs(txErr, expectedErr)
	s.Equal(before, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestNestedWithTxSharesOuterTransaction() {
	ctx := s.SystemContext()

	tdb := s.CreateTestDatabase()
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
	tdb := s.CreateTestDatabase()
	nestedTx := func(ctx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(ctx, tdb)
		return expectedErr
	}
	outerTx := func(ctx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(ctx, tdb)
		return tdb.WithTx(ctx, nestedTx)
	}

	ctx := s.SystemContext()
	before := s.tenantCount(ctx, tdb)
	err := tdb.WithTx(ctx, outerTx)
	s.ErrorIs(err, expectedErr)
	s.Equal(before, s.tenantCount(ctx, tdb))
}

func (s *DatabaseClientSuite) TestWithTxCommitHook() {
	ctx := s.SystemContext()
	committed := false
	tdb := s.CreateTestDatabase()
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
	ctx := s.SystemContext()
	rolledBack := false
	tdb := s.CreateTestDatabase()
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
	ctx := s.SystemContext()
	tdb := s.CreateTestDatabase()

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
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	err := tdb.AcquireTxLocks(ctx, "test", "a")

	s.Require().Error(err)
	s.Contains(err.Error(), "no active transaction")
}

func (s *DatabaseClientSuite) TestAcquireTxLocksWorksInsideTransaction() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	err := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		return tdb.AcquireTxLocks(txCtx, "test", "b", "a", "a")
	})

	s.Require().NoError(err)
}

func (s *DatabaseClientSuite) TestAcquireTxLocksSortsOppositeInputOrder() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	lockFn := func(keys ...string) {
		defer wg.Done()
		err := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
			return tdb.AcquireTxLocks(txCtx, "test", keys...)
		})
		errs <- err
	}

	wg.Add(2)
	go lockFn("b", "a")
	go lockFn("a", "b")
	wg.Wait()
	close(errs)

	for err := range errs {
		s.Require().NoError(err)
	}
}

func (s *DatabaseClientSuite) TestIsTransientErrorClassifiesPostgresConcurrencyErrors() {
	tdb := postgres.DatabaseClient{}
	s.True(tdb.IsTransientError(&pgconn.PgError{Code: "40P01"}))
	s.True(tdb.IsTransientError(fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "40001"})))
	s.False(tdb.IsTransientError(&pgconn.PgError{Code: "23505"}))
	s.False(tdb.IsTransientError(errors.New("plain error")))
}
