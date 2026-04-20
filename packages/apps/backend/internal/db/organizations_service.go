package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/jobs"
)

type OrganizationsService struct {
	db   *ent.Client
	jobs rez.JobsService
}

func NewOrganizationsService(db *ent.Client, jobs rez.JobsService) (*OrganizationsService, error) {
	return &OrganizationsService{db: db, jobs: jobs}, nil
}

func (s *OrganizationsService) Get(ctx context.Context, p predicate.Organization) (*ent.Organization, error) {
	return s.db.Organization.Query().Where(p).First(ctx)
}

func (s *OrganizationsService) SyncFromAuthProvider(ctx context.Context, po ent.Organization) (*ent.Organization, error) {
	if po.AuthProviderID == "" {
		return nil, fmt.Errorf("no auth provider id set")
	}
	ctx = access.SystemContext(ctx)
	existing, getErr := s.Get(ctx, organization.AuthProviderID(po.AuthProviderID))
	if getErr != nil && !ent.IsNotFound(getErr) {
		return nil, fmt.Errorf("fetch existing: %w", getErr)
	}
	if existing != nil {
		// TODO: check if should sync every time
		// if !AlwaysSyncAuthDetails { return existing, nil }
		ctx = access.TenantContext(ctx, existing.TenantID)
	}

	var updated *ent.Organization
	updateTx := func(tx *ent.Tx) error {
		var mutator ent.EntityMutator[*ent.Organization, *ent.OrganizationMutation]
		if existing == nil {
			tnt, tntErr := tx.Tenant.Create().Save(ctx)
			if tntErr != nil {
				return fmt.Errorf("create tenant: %w", tntErr)
			}
			mutator = tx.Organization.Create().
				SetTenantID(tnt.ID).
				SetAuthProviderID(po.AuthProviderID)
			ctx = access.TenantContext(ctx, tnt.ID)
		} else {
			mutator = tx.Organization.UpdateOne(existing)
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
	if txErr := ent.WithTx(ctx, s.db, updateTx); txErr != nil {
		return nil, fmt.Errorf("update: %w", txErr)
	}

	return updated, nil
}

func (s *OrganizationsService) CompleteSetup(ctx context.Context, org *ent.Organization) error {
	args := jobs.SyncIntegrationsData{
		OrganizationId: org.ID,
		IgnoreHistory:  true,
		CreateDefaults: true,
	}
	if jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
		slog.Error("failed to insert sync job", "error", jobErr)
	}

	return s.db.Organization.UpdateOne(org).SetInitialSetupAt(time.Now()).Exec(ctx)
}
