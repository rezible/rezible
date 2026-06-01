package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/testkit"
	"github.com/stretchr/testify/suite"
)

type DatabaseClientSuite struct {
	testkit.Suite
}

func TestDatabaseClientSuite(t *testing.T) {
	suite.Run(t, &DatabaseClientSuite{Suite: testkit.NewSuite()})
}

func (s *DatabaseClientSuite) tenantCount(ctx context.Context) int {
	count, err := s.Database().Client(ctx).Tenant.Query().Count(ctx)
	s.Require().NoError(err)
	return count
}

func (s *DatabaseClientSuite) createTenant(ctx context.Context, client *ent.Client) error {
	_, err := client.Tenant.Create().Save(ctx)
	return err
}

func (s *DatabaseClientSuite) requireCreateTenant(ctx context.Context) {
	_, err := s.Database().Client(ctx).Tenant.Create().Save(ctx)
	s.Require().NoError(err)
}

func (s *DatabaseClientSuite) TestClientOutsideTransaction() {
	ctx := s.SystemContext()

	s.NotNil(s.Database().Client(ctx))
	s.NotPanics(func() {
		_ = s.tenantCount(ctx)
	})
}

func (s *DatabaseClientSuite) TestWithTxCommits() {
	ctx := s.SystemContext()
	before := s.tenantCount(ctx)
	s.Require().NoError(s.Database().WithTx(ctx, s.createTenant))
	s.Equal(before+1, s.tenantCount(ctx))
}

func (s *DatabaseClientSuite) TestWithTxRollsBackOnError() {
	ctx := s.SystemContext()
	before := s.tenantCount(ctx)

	expectedErr := fmt.Errorf("force rollback")
	txErr := s.Database().WithTx(ctx, func(txCtx context.Context, client *ent.Client) error {
		_ = s.createTenant(txCtx, client)
		return expectedErr
	})
	s.ErrorIs(txErr, expectedErr)
	s.Equal(before, s.tenantCount(ctx))
}

func (s *DatabaseClientSuite) TestNestedWithTxSharesOuterTransaction() {
	ctx := s.SystemContext()
	before := s.tenantCount(ctx)

	db := s.Database()

	s.Require().NoError(db.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(txCtx)

		nestedErr := db.WithTx(txCtx, func(nestedCtx context.Context, _ *ent.Client) error {
			s.requireCreateTenant(nestedCtx)
			s.Equal(before+2, s.tenantCount(nestedCtx))
			return nil
		})
		if nestedErr != nil {
			return nestedErr
		}

		s.Equal(before+2, s.tenantCount(txCtx))
		return nil
	}))

	s.Equal(before+2, s.tenantCount(ctx))
}

func (s *DatabaseClientSuite) TestNestedWithTxErrorRollsBackOuterTransaction() {
	expectedErr := errors.New("nested rollback")
	nestedTx := func(ctx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(ctx)
		return expectedErr
	}
	outerTx := func(ctx context.Context, _ *ent.Client) error {
		s.requireCreateTenant(ctx)
		return s.Database().WithTx(ctx, nestedTx)
	}

	ctx := s.SystemContext()
	before := s.tenantCount(ctx)
	err := s.Database().WithTx(ctx, outerTx)
	s.ErrorIs(err, expectedErr)
	s.Equal(before, s.tenantCount(ctx))
}

func (s *DatabaseClientSuite) TestWithTxCommitHook() {
	ctx := s.SystemContext()
	committed := false
	hook := func(next ent.Committer) ent.Committer {
		return ent.CommitFunc(func(ctx context.Context, tx *ent.Tx) error {
			if err := next.Commit(ctx, tx); err != nil {
				return err
			}
			committed = true
			return nil
		})
	}

	err := s.Database().WithTx(ctx, s.createTenant, ent.WithCommitHook(hook))

	s.Require().NoError(err)
	s.True(committed)
}

func (s *DatabaseClientSuite) TestWithTxRollbackHook() {
	ctx := s.SystemContext()
	rolledBack := false
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

	err := s.Database().WithTx(ctx, txFn, ent.WithRollbackHook(hook))

	s.ErrorIs(err, expectedErr)
	s.True(rolledBack)
}

func (s *DatabaseClientSuite) TestWithTxRollsBackOnPanic() {
	ctx := s.SystemContext()

	panicErr := "expected"
	panicFn := func() {
		_ = s.Database().WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
			s.requireCreateTenant(txCtx)
			panic(panicErr)
		})
	}

	before := s.tenantCount(ctx)
	s.PanicsWithValue(panicErr, panicFn)
	s.Equal(before, s.tenantCount(ctx))
}
