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
)

type OrganizationsService struct {
	db *ent.Client
}

func NewOrganizationsService(db *ent.Client) (*OrganizationsService, error) {
	return &OrganizationsService{db: db}, nil
}

func (s *OrganizationsService) GetById(ctx context.Context, id uuid.UUID) (*ent.Organization, error) {
	return s.db.Organization.Get(ctx, id)
}

func (s *OrganizationsService) GetCurrent(ctx context.Context) (*ent.Organization, error) {
	// scoped by tenant id in context
	return s.db.Organization.Query().First(ctx)
}

func (s *OrganizationsService) FindOrCreateFromProvider(ctx context.Context, o ent.Organization) (*ent.Organization, error) {
	orgQuery := s.db.Organization.Query().
		Where(organization.ExternalID(o.ExternalID))

	org, orgErr := orgQuery.Only(access.SystemContext(ctx))
	if orgErr != nil && !ent.IsNotFound(orgErr) {
		return nil, fmt.Errorf("failed to query organization: %w", orgErr)
	}
	if org != nil {
		return org, nil
	}
	if !rez.Config.AllowTenantCreation() {
		return nil, rez.ErrInvalidTenant
	}

	var createdOrg *ent.Organization
	createTenantOrgFn := func(tx *ent.Tx) error {
		tnt, tenantErr := tx.Tenant.Create().Save(access.SystemContext(ctx))
		if tenantErr != nil {
			return fmt.Errorf("create tenant: %w", tenantErr)
		}

		org, orgErr = tx.Organization.Create().
			SetTenant(tnt).
			SetExternalID(o.ExternalID).
			SetName(o.Name).
			Save(access.TenantSystemContext(ctx, tnt.ID))
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

func (s *OrganizationsService) CompleteSetup(ctx context.Context, id uuid.UUID) error {
	o, orgErr := s.GetById(ctx, id)
	if orgErr != nil {
		return orgErr
	}
	return o.Update().SetInitialSetupAt(time.Now()).Exec(ctx)
}
