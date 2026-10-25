package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/ent/userauthsession"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
)

const (
	authSessionLifetime = time.Hour
	// authSessionRefreshWindow is how close to its expiry a session must be before a refresh asks the provider.
	authSessionRefreshWindow = 5 * time.Minute
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

		usr, userErr := s.syncAuthProviderUser(ctx, &ps.User, ps.Role, isNewOrg)
		if userErr != nil {
			return nil, fmt.Errorf("sync user: %w", userErr)
		}

		create := tx.UserAuthSession.Create().
			SetUserID(usr.ID).
			SetOrganizationID(org.ID).
			SetRole(userauthsession.Role(usr.Role)).
			SetExpiresAt(time.Now().Add(authSessionLifetime))
		if ps.RefreshToken != "" {
			create.SetProviderRefreshToken([]byte(ps.RefreshToken))
		}
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

// syncAuthProviderUser creates or updates the organization's user for the provider's user. A role the provider
// states is the user's role. Without one, an existing user keeps theirs, the user who creates the organization is
// its admin, and anyone else is a member.
func (s *AuthSessionService) syncAuthProviderUser(ctx context.Context, pu *ent.User, role user.Role, isNewOrg bool) (*ent.User, error) {
	existing, lookupErr := s.users.Get(ctx, user.AuthProviderID(pu.AuthProviderID))
	if lookupErr != nil && !ent.IsNotFound(lookupErr) {
		return nil, fmt.Errorf("lookup user: %w", lookupErr)
	}

	if role == "" {
		switch {
		case existing != nil:
			role = existing.Role
		case isNewOrg:
			role = user.RoleAdmin
		default:
			role = user.RoleMember
		}
	}

	var userId uuid.UUID
	if existing != nil {
		isEqual := pu.Name == existing.Name && pu.Email == existing.Email && role == existing.Role
		if isEqual {
			return existing, nil
		}
		userId = existing.ID
	}

	setFn := func(m *ent.UserMutation) {
		m.SetAuthProviderID(pu.AuthProviderID)
		m.SetName(pu.Name)
		m.SetEmail(pu.Email)
		m.SetRole(role)
	}
	return s.users.Set(ctx, userId, setFn)
}

func (s *AuthSessionService) CreateForToken(ctx context.Context, token string) (*ent.UserAuthSession, error) {
	return nil, errs.ErrAuthSessionInvalid
}

// Refresh extends a session in its last minutes with a new token from the provider, taking the role it states.
// A session without a refresh token is returned unchanged. A session the provider refuses, or whose token names
// another user or organization, is deleted.
func (s *AuthSessionService) Refresh(ctx context.Context, id uuid.UUID, exchange rez.AuthSessionRefreshExchange) (*ent.UserAuthSession, error) {
	stored, lookupErr := s.LookupSession(ctx, id)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil, errs.ErrAuthSessionInvalid
		}
		return nil, fmt.Errorf("lookup session: %w", lookupErr)
	}
	ctx = execution.NewTenantContext(ctx, stored.TenantID)

	// A refused refresh deletes the session and returns no error from the transaction, so the deletion commits;
	// the refusal is returned once it has.
	var refusal error
	refreshed, refreshErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.UserAuthSession, error) {
		// Tabs refresh the same session at once. The lock lets the first exchange the refresh token while the
		// rest wait and then find the session extended, so concurrent tabs never reuse one refresh token. It does
		// not make the exchange exactly-once: a lost response or a failed commit can still spend a token.
		lock := tx.UserAuthSession.Query().
			Where(userauthsession.ID(id)).
			ForUpdate()
		sess, lockErr := lock.Only(ctx)
		if lockErr != nil {
			if ent.IsNotFound(lockErr) {
				return nil, errs.ErrAuthSessionInvalid
			}
			return nil, fmt.Errorf("lock session: %w", lockErr)
		}
		timeLeft := time.Until(sess.ExpiresAt)
		if timeLeft <= 0 {
			return nil, errs.ErrAuthSessionExpired
		}
		if timeLeft > authSessionRefreshWindow || len(sess.ProviderRefreshToken) == 0 {
			return sess, nil
		}

		ps, exchangeErr := exchange(ctx, string(sess.ProviderRefreshToken))
		if exchangeErr != nil {
			if errors.Is(exchangeErr, errs.ErrAuthSessionInvalid) {
				refusal = exchangeErr
				return nil, tx.UserAuthSession.DeleteOneID(id).Exec(ctx)
			}
			return nil, fmt.Errorf("refresh with provider: %w", exchangeErr)
		}
		usr, userErr := tx.User.Get(ctx, sess.UserID)
		if userErr != nil {
			return nil, fmt.Errorf("get session user: %w", userErr)
		}
		org, orgErr := tx.Organization.Get(ctx, sess.OrganizationID)
		if orgErr != nil {
			return nil, fmt.Errorf("get session organization: %w", orgErr)
		}
		if ps.User.AuthProviderID != usr.AuthProviderID || ps.Org.AuthProviderID != org.AuthProviderID {
			refusal = fmt.Errorf("%w: refresh names another user or organization", errs.ErrAuthSessionInvalid)
			return nil, tx.UserAuthSession.DeleteOneID(id).Exec(ctx)
		}

		update := tx.UserAuthSession.UpdateOneID(id).
			SetExpiresAt(time.Now().Add(authSessionLifetime))
		if ps.Role != "" {
			if roleErr := tx.User.UpdateOneID(usr.ID).SetRole(ps.Role).Exec(ctx); roleErr != nil {
				return nil, fmt.Errorf("set user role: %w", roleErr)
			}
			update.SetRole(userauthsession.Role(ps.Role))
		}
		if ps.RefreshToken != "" {
			update.SetProviderRefreshToken([]byte(ps.RefreshToken))
		}
		return update.Save(ctx)
	})
	if refreshErr != nil {
		return nil, refreshErr
	}
	if refusal != nil {
		return nil, refusal
	}
	return refreshed, nil
}

func (s *AuthSessionService) LookupSession(ctx context.Context, id uuid.UUID) (*ent.UserAuthSession, error) {
	ctx = execution.NewSystemContext(ctx)
	return s.db.Client(ctx).UserAuthSession.Get(ctx, id)
}

func (s *AuthSessionService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).UserAuthSession.DeleteOneID(id).Exec(ctx)
}
