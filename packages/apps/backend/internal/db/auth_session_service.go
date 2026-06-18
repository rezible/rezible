package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/user"
	uas "github.com/rezible/rezible/ent/userauthsession"
	"github.com/rezible/rezible/execution"
)

type AuthSessionService struct {
	db    rez.Database
	orgs  rez.OrganizationService
	users rez.UserService
}

func NewAuthSessionService(db rez.Database, orgs rez.OrganizationService, users rez.UserService) *AuthSessionService {
	return &AuthSessionService{db: db, orgs: orgs, users: users}
}

func (s *AuthSessionService) CreateFromUserAuth(ctx context.Context, ps *rez.UserAuthProviderSession) (*ent.UserAuthSession, error) {
	var sess *ent.UserAuthSession
	return sess, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		ctx = execution.NewSystemContext(ctx)

		org, orgErr := s.syncAuthProviderOrg(ctx, tx, &ps.Org)
		if orgErr != nil {
			return fmt.Errorf("sync org: %w", orgErr)
		}
		ctx = execution.NewTenantContext(ctx, org.TenantID)

		usr, userErr := s.syncAuthProviderUser(ctx, &ps.User)
		if userErr != nil {
			return fmt.Errorf("sync user: %w", userErr)
		}

		deleteExisting := tx.UserAuthSession.Delete().
			Where(uas.And(uas.OrganizationID(org.ID), uas.UserID(usr.ID)))
		numDeleted, delErr := deleteExisting.Exec(ctx)
		if delErr != nil {
			return fmt.Errorf("delete existing session: %w", delErr)
		}
		if numDeleted > 0 {
			slog.Debug("deleted existing sessions", "count", numDeleted)
		}

		create := tx.UserAuthSession.Create().
			SetUserID(usr.ID).
			SetOrganizationID(org.ID).
			SetExpiresAt(time.Now().Add(time.Hour))
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create session: %w", createErr)
		}
		sess = created.Unwrap()

		return nil
	})
}

func (s *AuthSessionService) syncAuthProviderOrg(ctx context.Context, c *ent.Client, po *ent.Organization) (*ent.Organization, error) {
	existing, lookupErr := s.orgs.Get(ctx, organization.AuthProviderID(po.AuthProviderID))
	if lookupErr != nil {
		return nil, fmt.Errorf("lookup organization: %w", lookupErr)
	}

	var orgId uuid.UUID
	if existing != nil {
		ctx = execution.NewTenantContext(ctx, existing.TenantID)
		orgId = existing.ID

		isEqual := po.Name == existing.Name
		if isEqual {
			return existing, nil
		}
	} else {
		// TODO: new tenant for each org?
		tnt, saveTntErr := c.Tenant.Create().Save(ctx)
		if saveTntErr != nil {
			return nil, fmt.Errorf("create tenant: %w", saveTntErr)
		}
		ctx = execution.NewTenantContext(ctx, tnt.ID)
	}

	return s.orgs.Set(ctx, orgId, func(m *ent.OrganizationMutation) {
		m.SetAuthProviderID(po.AuthProviderID)
		m.SetName(po.Name)
	})
}

func (s *AuthSessionService) syncAuthProviderUser(ctx context.Context, pu *ent.User) (*ent.User, error) {
	existing, lookupErr := s.users.Get(ctx, user.AuthProviderID(pu.AuthProviderID))
	if lookupErr != nil {
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

	return s.users.Set(ctx, userId, func(m *ent.UserMutation) {
		m.SetAuthProviderID(pu.AuthProviderID)
		m.SetName(pu.Name)
		m.SetEmail(pu.Email)
	})
}

func (s *AuthSessionService) CreateFromApiToken(ctx context.Context, token string) (*ent.UserAuthSession, error) {
	return nil, fmt.Errorf("not supported")
}

func (s *AuthSessionService) Get(ctx context.Context, id uuid.UUID) (*ent.UserAuthSession, error) {
	ctx = execution.NewSystemContext(ctx)
	return s.db.Client(ctx).UserAuthSession.Get(ctx, id)
}

func (s *AuthSessionService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).UserAuthSession.DeleteOneID(id).Exec(ctx)
}
