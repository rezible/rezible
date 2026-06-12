package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationpreferences"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
)

type OrganizationService struct {
	db   rez.Database
	jobs rez.JobService
}

func NewOrganizationService(db rez.Database, jobs rez.JobService) *OrganizationService {
	return &OrganizationService{db: db, jobs: jobs}
}

func (s *OrganizationService) Get(ctx context.Context, p predicate.Organization) (*ent.Organization, error) {
	query := s.db.Client(ctx).Organization.Query().
		Where(p).
		WithPreferences()
	return query.Only(ctx)
}

func (s *OrganizationService) getByAuthProviderId(ctx context.Context, authId string) (*ent.Organization, error) {
	query := s.db.Client(ctx).Organization.Query().
		Where(organization.AuthProviderID(authId)).
		WithPreferences()
	return query.Only(ctx)
}

func (s *OrganizationService) SyncFromAuthProvider(ctx context.Context, po ent.Organization) (*ent.Organization, error) {
	var org *ent.Organization
	return org, s.db.WithTx(execution.NewSystemContext(ctx), func(ctx context.Context, tx *ent.Client) error {
		existing, getErr := s.getByAuthProviderId(ctx, po.AuthProviderID)
		if getErr != nil && !ent.IsNotFound(getErr) {
			return fmt.Errorf("fetch existing: %w", getErr)
		}
		if existing != nil {
			if po.Name == existing.Name {
				org = existing
				return nil
			}
			ctx = execution.NewTenantContext(ctx, existing.TenantID)
		}

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

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save org: %w", saveErr)
		}

		org = saved

		return nil
	})
}

var orgInitialSetupIntegrationSyncJob = jobs.SyncIntegrationEventsArgs{SyncReason: "org_initial_setup"}

func (s *OrganizationService) SetPreferences(ctx context.Context, orgId uuid.UUID, setFn func(*ent.OrganizationPreferencesMutation)) (*ent.OrganizationPreferences, error) {
	var prefs *ent.OrganizationPreferences
	return prefs, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		queryExisting := tx.OrganizationPreferences.Query().
			Where(organizationpreferences.OrganizationID(orgId))
		curr, queryErr := queryExisting.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return queryErr
		}

		var mutator ent.EntityMutator[*ent.OrganizationPreferences, *ent.OrganizationPreferencesMutation]
		if curr != nil {
			mutator = curr.Update()
		} else {
			mutator = tx.OrganizationPreferences.Create().
				SetOrganizationID(orgId)
		}
		m := mutator.Mutation()
		setFn(m)

		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save: %w", saveErr)
		}

		wasInitialSetup := !updated.InitialSetupAt.IsZero() && (curr == nil || curr.InitialSetupAt.IsZero())
		if wasInitialSetup {
			if _, jobErr := s.jobs.Insert(ctx, orgInitialSetupIntegrationSyncJob, nil); jobErr != nil {
				slog.Error("failed to insert org sync job", "error", jobErr)
			}
		}
		prefs = updated
		return nil
	})
}
