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

func (s *OrganizationsService) createForDomain(ctx context.Context, domain string) (*ent.Organization, error) {
	var createdOrg *ent.Organization
	createFn := func(tx *ent.Tx) error {
		tenantCreate := tx.Tenant.Create()
		createdTenant, tenantErr := tenantCreate.Save(access.SystemContext(ctx))
		if tenantErr != nil {
			return fmt.Errorf("save tenant: %w", tenantErr)
		}
		ctx = access.TenantContext(ctx, createdTenant.ID)

		orgCreate := tx.Organization.Create().
			SetTenant(createdTenant).
			SetDomain(domain).
			SetName(domain)

		var createErr error
		createdOrg, createErr = orgCreate.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("save organization: %w", createErr)
		}

		return nil
	}
	return createdOrg, ent.WithTx(ctx, s.db, createFn)
}

func (s *OrganizationsService) FindOrCreateFromDomain(ctx context.Context, domain string) (*ent.Organization, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain empty")
	}

	orgQuery := s.db.Organization.Query().Where(organization.Domain(domain))

	org, orgErr := orgQuery.Only(access.SystemContext(ctx))
	if org != nil {
		return org, nil
	} else if orgErr != nil && !ent.IsNotFound(orgErr) {
		return nil, fmt.Errorf("query organization: %w", orgErr)
	}

	createdOrg, createErr := s.createForDomain(ctx, domain)
	if createErr != nil {
		return nil, fmt.Errorf("create organization: %w", createErr)
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
