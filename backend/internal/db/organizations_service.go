package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/jobs"
	"github.com/rs/zerolog/log"
)

type OrganizationsService struct {
	db   *ent.Client
	jobs rez.JobsService
}

func NewOrganizationsService(db *ent.Client, jobs rez.JobsService) (*OrganizationsService, error) {
	return &OrganizationsService{db: db, jobs: jobs}, nil
}

func (s *OrganizationsService) GetById(ctx context.Context, id uuid.UUID) (*ent.Organization, error) {
	return s.db.Organization.Get(ctx, id)
}

func (s *OrganizationsService) GetCurrent(ctx context.Context) (*ent.Organization, error) {
	// scoped by tenant id in context
	return s.db.Organization.Query().First(ctx)
}

func (s *OrganizationsService) FindOrCreateFromProvider(ctx context.Context, o ent.Organization) (*ent.Organization, error) {
	// Required to query without a tenant id set
	ctx = access.SystemContext(ctx)

	orgQuery := s.db.Organization.Query().
		Where(organization.ExternalID(o.ExternalID))

	org, orgErr := orgQuery.Only(ctx)
	if orgErr != nil && !ent.IsNotFound(orgErr) {
		return nil, fmt.Errorf("failed to query organization: %w", orgErr)
	}
	if org != nil {
		return org, nil
	}
	if !rez.Config.AllowTenantCreation() {
		return nil, rez.ErrInvalidTenant
	}

	var createdTenant *ent.Tenant
	var createdOrg *ent.Organization
	createTenantOrgFn := func(tx *ent.Tx) error {
		var createErr error

		createTenant := tx.Tenant.Create()
		createdTenant, createErr = createTenant.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create tenant: %w", createErr)
		}
		ctx = access.TenantContext(ctx, createdTenant.ID)

		createOrg := tx.Organization.Create().
			SetTenant(createdTenant).
			SetExternalID(o.ExternalID).
			SetName(o.Name)
		org, orgErr = createOrg.Save(ctx)
		if orgErr != nil {
			return fmt.Errorf("create organization: %w", orgErr)
		}
		createdOrg = org

		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, createTenantOrgFn); txErr != nil {
		return nil, txErr
	}

	return createdOrg, nil
}

func (s *OrganizationsService) CompleteSetup(ctx context.Context, org *ent.Organization) error {
	args := jobs.SyncIntegrationsData{
		OrganizationId: org.ID,
		IgnoreHistory:  true,
		CreateDefaults: true,
	}
	if jobErr := s.jobs.Insert(ctx, args, nil); jobErr != nil {
		log.Error().Err(jobErr).Msg("failed to insert sync job")
	}

	return s.db.Organization.UpdateOne(org).SetInitialSetupAt(time.Now()).Exec(ctx)
}
