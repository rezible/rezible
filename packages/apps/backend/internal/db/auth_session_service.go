package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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

func (s *AuthSessionService) SyncFromAuthProvider(ctx context.Context, pu ent.User, po ent.Organization) (*ent.UserAuthSession, error) {
	var sess *ent.UserAuthSession
	return sess, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		org, orgErr := s.orgs.SyncFromAuthProvider(ctx, po)
		if orgErr != nil {
			return fmt.Errorf("sync organization: %w", orgErr)
		}
		ctx = execution.NewTenantContext(ctx, org.TenantID)

		usr, usrErr := s.users.SyncFromAuthProvider(ctx, pu)
		if usrErr != nil {
			return fmt.Errorf("sync user: %w", usrErr)
		}

		existing, queryErr := tx.UserAuthSession.Query().
			Where(uas.OrganizationID(org.ID), uas.UserID(usr.ID)).
			Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return fmt.Errorf("lookup existing: %w", queryErr)
		}

		expiresAt := time.Now().Add(time.Hour)
		var updated *ent.UserAuthSession
		if existing != nil {
			var updateErr error
			updated, updateErr = existing.Update().SetExpiresAt(expiresAt).Save(ctx)
			if updateErr != nil {
				return fmt.Errorf("update session: %w", updateErr)
			}
		} else {
			create := tx.UserAuthSession.Create().
				SetUserID(usr.ID).
				SetOrganizationID(org.ID).
				SetExpiresAt(expiresAt)
			var createErr error
			updated, createErr = create.Save(ctx)
			if createErr != nil {
				return fmt.Errorf("create session: %w", createErr)
			}
		}
		sess = updated.Unwrap()

		return nil
	})
}

func (s *AuthSessionService) LookupSession(ctx context.Context, id uuid.UUID) (*ent.UserAuthSession, error) {
	ctx = execution.NewSystemContext(ctx)
	return s.db.Client(ctx).UserAuthSession.Get(ctx, id)
}

func (s *AuthSessionService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).UserAuthSession.DeleteOneID(id).Exec(ctx)
}

func (s *AuthSessionService) CreateExecutionContext(ctx context.Context, sess *ent.UserAuthSession) (context.Context, error) {
	if sess.ExpiresAt.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	usr, lookupErr := s.users.Get(execution.NewSystemContext(ctx), user.ID(sess.UserID))
	if lookupErr != nil {
		slog.Debug("get user", "error", lookupErr, "userId", sess.UserID)
		return nil, rez.ErrAuthSessionInvalid
	}

	return execution.NewUserAuthContext(ctx, *usr, sess.ExpiresAt), nil
}
