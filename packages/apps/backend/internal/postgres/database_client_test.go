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
	"github.com/rezible/rezible/ent/schema"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	"github.com/rezible/rezible/ent/tenant"
	"github.com/rezible/rezible/ent/user"
	dbservices "github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/pkg/errs"
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

func (s *DatabaseClientSuite) TestIsForeignKeyErrorMatchesNamedConstraints() {
	tdb := postgres.DatabaseClient{}
	violation := &pgconn.PgError{Code: "23503", ConstraintName: "named_fk"}
	s.True(tdb.IsForeignKeyError(fmt.Errorf("wrapped: %w", violation), "other_fk", "named_fk"))
	s.False(tdb.IsForeignKeyError(violation, "other_fk"))
	s.False(tdb.IsForeignKeyError(&pgconn.PgError{Code: "23505", ConstraintName: "named_fk"}, "named_fk"))
	s.False(tdb.IsForeignKeyError(errors.New("plain error"), "named_fk"))
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

func (s *DatabaseClientSuite) createTeam(ctx context.Context, tdb rez.Database) *ent.Team {
	createTeam := tdb.Client(ctx).Team.Create().
		SetSlug(uuid.NewString()).
		SetName("team")
	created, createErr := createTeam.Save(ctx)
	s.Require().NoError(createErr)
	return created
}

func (s *DatabaseClientSuite) TestCreateUsesTheContextTenant() {
	ctx, tdb := s.SetupTestDatabase()
	otherCtx, _ := s.SetupTestDatabase()
	other := s.createTeam(otherCtx, tdb)

	createInOther := tdb.Client(ctx).Team.Create().
		SetTenantID(other.TenantID).
		SetSlug(uuid.NewString()).
		SetName("other")
	s.ErrorIs(createInOther.Exec(ctx), errs.ErrTenantContextMissing)

	systemCtx := execution.NewSystemContext(s.T().Context())
	createWithoutTenant := tdb.Client(systemCtx).Team.Create().
		SetTenantID(other.TenantID).
		SetSlug(uuid.NewString()).
		SetName("system")
	s.ErrorIs(createWithoutTenant.Exec(systemCtx), errs.ErrTenantContextMissing)

	count, countErr := tdb.Client(systemCtx).Team.Query().
		Where(team.NameIn("other", "system")).
		Count(systemCtx)
	s.Require().NoError(countErr)
	s.Zero(count)
}

func (s *DatabaseClientSuite) TestWriteReferencingAMissingRowIsNotFound() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	owner, ownerErr := client.User.Query().Only(ctx)
	s.Require().NoError(ownerErr)
	existing := s.createTeam(ctx, tdb)
	createIncident := client.Incident.Create().
		SetSlug(uuid.NewString()).
		SetTitle("incident")
	incident, incidentErr := createIncident.Save(ctx)
	s.Require().NoError(incidentErr)

	createMembership := client.TeamMembership.Create().
		SetTeamID(uuid.New()).
		SetUserID(owner.ID)
	s.ErrorIs(createMembership.Exec(ctx), errs.ErrNotFound, "required field")
	setSeverity := client.Incident.UpdateOneID(incident.ID).SetSeverityID(uuid.New())
	s.ErrorIs(setSeverity.Exec(ctx), errs.ErrNotFound, "optional field")
	addRoster := client.Team.UpdateOneID(existing.ID).AddOncallRosterIDs(uuid.New())
	s.ErrorIs(addRoster.Exec(ctx), errs.ErrNotFound, "many-to-many edge")
	bulk := client.TeamMembership.CreateBulk(
		client.TeamMembership.Create().SetTeamID(existing.ID).SetUserID(owner.ID),
		client.TeamMembership.Create().SetTeamID(uuid.New()).SetUserID(owner.ID),
	)
	s.ErrorIs(bulk.Exec(ctx), errs.ErrNotFound, "bulk create")
	upsert := client.TeamMembership.Create().
		SetTeamID(uuid.New()).
		SetUserID(owner.ID).
		OnConflictColumns(teammembership.FieldTeamID, teammembership.FieldUserID).
		UpdateNewValues()
	s.ErrorIs(upsert.Exec(ctx), errs.ErrNotFound, "upsert")

	memberships, membershipsErr := client.TeamMembership.Query().Count(ctx)
	s.Require().NoError(membershipsErr)
	s.Zero(memberships)
	reloaded, reloadErr := client.Incident.Get(ctx, incident.ID)
	s.Require().NoError(reloadErr)
	s.Nil(reloaded.SeverityID)
}

// The test database's context is a tenant context with no user, as the provider webhook and chat handlers create.
func (s *DatabaseClientSuite) TestWriteReferencingAnotherTenantsRowIsNotFound() {
	ctx, tdb := s.SetupTestDatabase()
	otherCtx, _ := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	otherClient := tdb.Client(otherCtx)
	owner, ownerErr := client.User.Query().Only(ctx)
	s.Require().NoError(ownerErr)
	team := s.createTeam(ctx, tdb)
	otherTeam := s.createTeam(otherCtx, tdb)
	severity := client.IncidentSeverity.Create().SetName("sev").SetRank(1).SaveX(ctx)
	otherSeverity := otherClient.IncidentSeverity.Create().SetName("sev").SetRank(1).SaveX(otherCtx)
	incident := client.Incident.Create().SetSlug(uuid.NewString()).SetTitle("incident").SaveX(ctx)
	otherIncident := otherClient.Incident.Create().SetSlug(uuid.NewString()).SetTitle("other").SaveX(otherCtx)
	otherRoster := otherClient.OncallRoster.Create().SetName("roster").SetSlug(uuid.NewString()).SaveX(otherCtx)

	setSeverity := client.Incident.UpdateOneID(incident.ID).SetSeverityID(otherSeverity.ID)
	s.ErrorIs(setSeverity.Exec(ctx), errs.ErrNotFound, "unique edge")
	addIncident := client.IncidentSeverity.UpdateOneID(severity.ID).AddIncidentIDs(otherIncident.ID)
	s.ErrorIs(addIncident.Exec(ctx), errs.ErrNotFound, "non-unique edge")
	addRoster := client.Team.UpdateOneID(team.ID).AddOncallRosterIDs(otherRoster.ID)
	s.ErrorIs(addRoster.Exec(ctx), errs.ErrNotFound, "many-to-many edge")
	bulk := client.TeamMembership.CreateBulk(
		client.TeamMembership.Create().SetTeamID(team.ID).SetUserID(owner.ID),
		client.TeamMembership.Create().SetTeamID(otherTeam.ID).SetUserID(owner.ID),
	)
	s.ErrorIs(bulk.Exec(ctx), errs.ErrNotFound, "bulk create")
	upsert := client.TeamMembership.Create().
		SetTeamID(otherTeam.ID).
		SetUserID(owner.ID).
		OnConflictColumns(teammembership.FieldTeamID, teammembership.FieldUserID).
		UpdateNewValues()
	s.ErrorIs(upsert.Exec(ctx), errs.ErrNotFound, "upsert")

	s.Zero(client.TeamMembership.Query().CountX(ctx))
	s.Nil(client.Incident.GetX(ctx, incident.ID).SeverityID)
	s.Nil(otherClient.Incident.GetX(otherCtx, otherIncident.ID).SeverityID)
	systemCtx := execution.NewSystemContext(s.T().Context())
	s.Zero(tdb.Client(systemCtx).OncallRoster.QueryTeams(otherRoster).CountX(systemCtx))

	sameTenant := client.Incident.UpdateOneID(incident.ID).SetSeverityID(severity.ID)
	s.Require().NoError(sameTenant.Exec(ctx))
}

func (s *DatabaseClientSuite) TestReportMissingReferencesLeavesAMissingRowToItsForeignKey() {
	ctx, tdb := s.SetupTestDatabase()
	otherCtx, _ := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	incident := client.Incident.Create().SetSlug(uuid.NewString()).SetTitle("incident").SaveX(ctx)
	otherSeverity := tdb.Client(otherCtx).IncidentSeverity.Create().SetName("sev").SetRank(1).SaveX(otherCtx)
	reportCtx := schema.ReportMissingReferences(ctx)

	checkErr := client.Incident.UpdateOneID(incident.ID).SetSeverityID(uuid.New()).Exec(ctx)
	s.ErrorIs(checkErr, errs.ErrNotFound)
	_, isPgErr := errors.AsType[*pgconn.PgError](checkErr)
	s.False(isPgErr, "rejected before the write")

	fkErr := client.Incident.UpdateOneID(incident.ID).SetSeverityID(uuid.New()).Exec(reportCtx)
	s.ErrorIs(fkErr, errs.ErrNotFound)
	pgErr, isPgErr := errors.AsType[*pgconn.PgError](fkErr)
	s.Require().True(isPgErr, "rejected by the foreign key")
	s.Equal("23503", pgErr.Code)

	crossErr := client.Incident.UpdateOneID(incident.ID).SetSeverityID(otherSeverity.ID).Exec(reportCtx)
	s.ErrorIs(crossErr, errs.ErrNotFound, "another tenant's row is still not found")
}

func (s *DatabaseClientSuite) TestDeletingAReferencedRowKeepsItsError() {
	ctx, tdb := s.SetupTestDatabase()
	client := tdb.Client(ctx)
	owner, ownerErr := client.User.Query().Only(ctx)
	s.Require().NoError(ownerErr)
	createMembership := client.TeamMembership.Create().
		SetTeamID(s.createTeam(ctx, tdb).ID).
		SetUserID(owner.ID)
	s.Require().NoError(createMembership.Exec(ctx))

	deleteErr := client.User.DeleteOneID(owner.ID).Exec(ctx)

	s.True(ent.IsConstraintError(deleteErr), "%v", deleteErr)
	s.NotErrorIs(deleteErr, errs.ErrNotFound)
}

func (s *DatabaseClientSuite) TestUpsertMatchingAnotherTenantsKeyInserts() {
	ctx, tdb := s.SetupTestDatabase()
	otherCtx, _ := s.SetupTestDatabase()
	other := s.createTeam(otherCtx, tdb)

	upsert := tdb.Client(ctx).Team.Create().
		SetSlug(other.Slug).
		SetName("upserted").
		OnConflictColumns(team.FieldTenantID, team.FieldSlug).
		UpdateName()
	id, upsertErr := upsert.ID(ctx)
	s.Require().NoError(upsertErr)

	s.NotEqual(other.ID, id)
	reloaded, reloadErr := tdb.Client(otherCtx).Team.Get(otherCtx, other.ID)
	s.Require().NoError(reloadErr)
	s.Equal(other.Name, reloaded.Name)
}
