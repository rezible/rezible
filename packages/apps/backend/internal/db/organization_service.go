package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
)

type OrganizationService struct {
	db   rez.Database
	jobs rez.JobService
}

func NewOrganizationService(db rez.Database, jobs rez.JobService) (*OrganizationService, error) {
	return &OrganizationService{db: db, jobs: jobs}, nil
}

func (s *OrganizationService) Get(ctx context.Context, p predicate.Organization) (*ent.Organization, error) {
	return s.db.Client(ctx).Organization.Query().Where(p).First(ctx)
}

func (s *OrganizationService) SyncFromAuthProvider(ctx context.Context, po ent.Organization) (*ent.Organization, error) {
	if po.AuthProviderID == "" {
		return nil, fmt.Errorf("no auth provider id set")
	}
	ctx = execution.NewSystemContext(ctx)
	existing, getErr := s.Get(ctx, organization.AuthProviderID(po.AuthProviderID))
	if getErr != nil && !ent.IsNotFound(getErr) {
		return nil, fmt.Errorf("fetch existing: %w", getErr)
	}
	if existing != nil {
		// TODO: check if should sync every time
		// if !AlwaysSyncAuthDetails { return existing, nil }
		ctx = execution.NewTenantContext(ctx, existing.TenantID)
	}

	var updated *ent.Organization
	updateTx := func(txCtx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.Organization, *ent.OrganizationMutation]
		if existing != nil {
			mutator = tx.Organization.UpdateOne(existing)
		} else {
			tnt, tntErr := tx.Tenant.Create().Save(ctx)
			if tntErr != nil {
				return fmt.Errorf("create tenant: %w", tntErr)
			}
			mutator = tx.Organization.Create().
				SetTenantID(tnt.ID).
				SetAuthProviderID(po.AuthProviderID)
			ctx = execution.NewTenantContext(ctx, tnt.ID)
		}

		m := mutator.Mutation()
		m.SetName(po.Name)

		var saveErr error
		updated, saveErr = mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save org: %w", saveErr)
		}

		return nil
	}
	if txErr := s.db.WithTx(ctx, updateTx); txErr != nil {
		return nil, fmt.Errorf("update: %w", txErr)
	}

	return updated, nil
}

func (s *OrganizationService) CompleteSetup(ctx context.Context, org *ent.Organization) error {
	update := s.db.Client(ctx).Organization.UpdateOne(org).SetInitialSetupAt(time.Now())
	if updateErr := update.Exec(ctx); updateErr != nil {
		return fmt.Errorf("update: %w", updateErr)
	}
	args := jobs.SyncIntegrationEventsArgs{
		SyncReason: "inital_setup",
	}
	if _, jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
		slog.Error("failed to insert sync job", "error", jobErr)
	}
	return nil
}
