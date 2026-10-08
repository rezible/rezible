package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationrole"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
)

type AuthSessionService struct {
	db    rez.Database
	orgs  rez.OrganizationService
	users rez.UserService
}

func NewAuthSessionService(db rez.Database, orgs rez.OrganizationService, users rez.UserService) (*AuthSessionService, error) {
	return &AuthSessionService{db: db, orgs: orgs, users: users}, nil
}

func (s *AuthSessionService) CreateFromUserAuthResponse(ctx context.Context, ps *rez.UserAuthProviderSession) (*ent.UserAuthSession, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.UserAuthSession, error) {
		ctx = execution.NewSystemContext(ctx)

		org, isNewOrg, orgErr := s.syncAuthProviderOrg(ctx, tx, &ps.Org)
		if orgErr != nil {
			return nil, fmt.Errorf("sync org: %w", orgErr)
		}
		ctx = execution.NewTenantContext(ctx, org.TenantID)

		usr, userErr := s.syncAuthProviderUser(ctx, &ps.User)
		if userErr != nil {
			return nil, fmt.Errorf("sync user: %w", userErr)
		}

		if isNewOrg {
			createAdminRole := tx.OrganizationRole.Create().
				SetUserID(usr.ID).
				SetOrganizationID(org.ID).
				SetRole(organizationrole.RoleAdmin)
			if roleErr := createAdminRole.Exec(ctx); roleErr != nil {
				return nil, fmt.Errorf("create admin role: %w", roleErr)
			}
		}

		create := tx.UserAuthSession.Create().
			SetUserID(usr.ID).
			SetOrganizationID(org.ID).
			SetExpiresAt(time.Now().Add(time.Hour))
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create session: %w", createErr)
		}
		return created, nil
	})
}

func (s *AuthSessionService) syncAuthProviderOrg(ctx context.Context, c *ent.Client, po *ent.Organization) (*ent.Organization, bool, error) {
	existing, lookupErr := s.orgs.Get(ctx, organization.AuthProviderID(po.AuthProviderID))
	if lookupErr != nil && !ent.IsNotFound(lookupErr) {
		return nil, false, fmt.Errorf("lookup organization: %w", lookupErr)
	}

	isCreate := existing == nil

	if existing != nil {
		return existing, isCreate, nil
	}
	// TODO: new tenant for each org?
	tnt, saveTntErr := c.Tenant.Create().Save(ctx)
	if saveTntErr != nil {
		return nil, false, fmt.Errorf("create tenant: %w", saveTntErr)
	}
	ctx = execution.NewTenantContext(ctx, tnt.ID)

	setFn := func(m *ent.OrganizationMutation) {
		m.SetAuthProviderID(po.AuthProviderID)
		m.SetName(po.Name)
	}
	org, setErr := s.orgs.Set(ctx, uuid.Nil, setFn)
	if setErr != nil {
		return nil, false, fmt.Errorf("set org: %w", setErr)
	}
	return org, isCreate, nil
}

func (s *AuthSessionService) syncAuthProviderUser(ctx context.Context, pu *ent.User) (*ent.User, error) {
	existing, lookupErr := s.users.Get(ctx, user.AuthProviderID(pu.AuthProviderID))
	if lookupErr != nil && !ent.IsNotFound(lookupErr) {
		return nil, fmt.Errorf("lookup user: %w", lookupErr)
	}

	var userId uuid.UUID
	if existing != nil {
		isEqual := pu.Name == existing.Name && pu.Email == existing.Email
		if isEqual {
			return existing, nil
		}
		userId = existing.ID
	}

	setFn := func(m *ent.UserMutation) {
		m.SetAuthProviderID(pu.AuthProviderID)
		m.SetName(pu.Name)
		m.SetEmail(pu.Email)
	}
	return s.users.Set(ctx, userId, setFn)
}

func (s *AuthSessionService) CreateForToken(ctx context.Context, token string) (*ent.UserAuthSession, error) {
	return nil, errs.ErrAuthSessionInvalid
}

func (s *AuthSessionService) LookupSession(ctx context.Context, id uuid.UUID) (*ent.UserAuthSession, error) {
	ctx = execution.NewSystemContext(ctx)
	return s.db.Client(ctx).UserAuthSession.Get(ctx, id)
}

func (s *AuthSessionService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).UserAuthSession.DeleteOneID(id).Exec(ctx)
}
