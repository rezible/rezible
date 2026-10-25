package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/ent/userauthsession"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/suite"
)

type AuthSessionServiceSuite struct {
	test.Suite
}

func TestAuthSessionServiceSuite(t *testing.T) {
	suite.Run(t, &AuthSessionServiceSuite{Suite: test.NewSuite()})
}

func (s *AuthSessionServiceSuite) TestCreatingSessionPreservesExistingSessions() {
	sessions, _ := s.newServices()
	providerSession := &rez.UserAuthProviderSession{
		User:      ent.User{Email: "session-test@example.com", Name: "Session Test", AuthProviderID: "session-test-user"},
		Org:       ent.Organization{Name: "Session Test Org", AuthProviderID: "session-test-org"},
		ExpiresAt: time.Now().Add(time.Hour),
	}

	first, firstErr := sessions.CreateFromUserAuthResponse(s.T().Context(), providerSession)
	s.Require().NoError(firstErr)
	second, secondErr := sessions.CreateFromUserAuthResponse(s.T().Context(), providerSession)
	s.Require().NoError(secondErr)
	s.NotEqual(first.ID, second.ID)

	loaded, lookupErr := sessions.LookupSession(s.T().Context(), first.ID)
	s.Require().NoError(lookupErr)
	s.Equal(first.ID, loaded.ID)
}

func (s *AuthSessionServiceSuite) TestLookupSessionMissingSessionIsNotFound() {
	sessions, _ := s.newServices()

	_, lookupErr := sessions.LookupSession(s.T().Context(), uuid.New())
	s.Require().Error(lookupErr)
	s.True(ent.IsNotFound(lookupErr))
}

func (s *AuthSessionServiceSuite) newServices() (*AuthSessionService, *UserService) {
	_, database := s.SetupTestDatabase()
	jobs := mocks.NewMockJobService(s.T())
	organizations, _ := NewOrganizationService(database, jobs)
	users, _ := NewUserService(database, organizations)
	sessions, _ := NewAuthSessionService(database, organizations, users)
	return sessions, users
}

func (s *AuthSessionServiceSuite) signIn(sessions *AuthSessionService, orgID, userID string, role user.Role) *ent.UserAuthSession {
	sess, err := sessions.CreateFromUserAuthResponse(s.T().Context(), &rez.UserAuthProviderSession{
		User:         ent.User{Email: userID + "@example.com", Name: userID, AuthProviderID: userID},
		Org:          ent.Organization{Name: orgID, AuthProviderID: orgID},
		Role:         role,
		RefreshToken: "refresh-" + userID,
		ExpiresAt:    time.Now().Add(time.Hour),
	})
	s.Require().NoError(err)
	return sess
}

func (s *AuthSessionServiceSuite) userRole(users *UserService, sess *ent.UserAuthSession) user.Role {
	ctx := execution.NewTenantContext(s.T().Context(), sess.TenantID)
	usr, err := users.Get(ctx, user.ID(sess.UserID))
	s.Require().NoError(err)
	return usr.Role
}

func (s *AuthSessionServiceSuite) TestSigningInTakesTheProviderRole() {
	sessions, users := s.newServices()

	// A member creating the organization stays a member.
	created := s.signIn(sessions, "org-a", "creator", user.RoleMember)
	s.Equal(userauthsession.RoleMember, created.Role)
	s.Equal(user.RoleMember, s.userRole(users, created))

	asAdmin := s.signIn(sessions, "org-a", "person", user.RoleAdmin)
	asMember := s.signIn(sessions, "org-a", "person", user.RoleMember)
	s.Equal(asAdmin.UserID, asMember.UserID)
	s.Equal(user.RoleMember, s.userRole(users, asMember))

	loaded, err := sessions.LookupSession(s.T().Context(), asAdmin.ID)
	s.Require().NoError(err)
	s.Equal(userauthsession.RoleAdmin, loaded.Role)
	s.True(execution.GetContext(execution.NewUserContext(s.T().Context(), loaded)).HasScope(execution.ScopeOrganizationManage))
	s.False(execution.GetContext(execution.NewUserContext(s.T().Context(), asMember)).HasScope(execution.ScopeOrganizationManage))

	// The same provider user in another organization has its own role.
	otherOrg := s.signIn(sessions, "org-b", "person", user.RoleAdmin)
	s.NotEqual(asMember.TenantID, otherOrg.TenantID)
	s.Equal(user.RoleAdmin, s.userRole(users, otherOrg))
	s.Equal(user.RoleMember, s.userRole(users, asMember))
}

func (s *AuthSessionServiceSuite) TestSigningInWithoutProviderRole() {
	sessions, users := s.newServices()

	first := s.signIn(sessions, "org", "first", "")
	second := s.signIn(sessions, "org", "second", "")
	s.Equal(userauthsession.RoleAdmin, first.Role)
	s.Equal(userauthsession.RoleMember, second.Role)

	s.Equal(user.RoleAdmin, s.userRole(users, s.signIn(sessions, "org", "first", "")))
	s.Equal(user.RoleMember, s.userRole(users, s.signIn(sessions, "org", "second", "")))
}

var errRefreshRefused = errors.New("invalid_grant")

func (s *AuthSessionServiceSuite) TestRefresh() {
	refreshed := func(orgID, userID string) *rez.UserAuthProviderSession {
		return &rez.UserAuthProviderSession{
			User:         ent.User{AuthProviderID: userID},
			Org:          ent.Organization{AuthProviderID: orgID},
			Role:         user.RoleMember,
			RefreshToken: "rotated",
		}
	}
	for _, tc := range []struct {
		name     string
		timeLeft time.Duration
		exchange func(*rez.UserAuthProviderSession) (*rez.UserAuthProviderSession, error)
		errIs    error
		kept     bool
		extended bool
	}{
		{"extends a session in its last minutes", 4 * time.Minute, func(ps *rez.UserAuthProviderSession) (*rez.UserAuthProviderSession, error) {
			return ps, nil
		}, nil, true, true},
		{"leaves a session with more time", 10 * time.Minute, nil, nil, true, false},
		{"deletes a refused session", 4 * time.Minute, func(*rez.UserAuthProviderSession) (*rez.UserAuthProviderSession, error) {
			return nil, fmt.Errorf("%w: %w", errs.ErrAuthSessionInvalid, errRefreshRefused)
		}, errRefreshRefused, false, false},
		{"deletes a session refreshed as another user", 4 * time.Minute, func(ps *rez.UserAuthProviderSession) (*rez.UserAuthProviderSession, error) {
			return refreshed(ps.Org.AuthProviderID, "someone-else"), nil
		}, errs.ErrAuthSessionInvalid, false, false},
		{"deletes a session refreshed in another organization", 4 * time.Minute, func(ps *rez.UserAuthProviderSession) (*rez.UserAuthProviderSession, error) {
			return refreshed("another-org", ps.User.AuthProviderID), nil
		}, errs.ErrAuthSessionInvalid, false, false},
		{"keeps a session when the provider fails", 4 * time.Minute, func(*rez.UserAuthProviderSession) (*rez.UserAuthProviderSession, error) {
			return nil, errors.New("provider unreachable")
		}, nil, true, false},
	} {
		s.Run(tc.name, func() {
			sessions, users := s.newServices()
			orgID := "refresh-org-" + uuid.NewString()
			sess := s.signIn(sessions, orgID, "person", user.RoleAdmin)
			tenantCtx := execution.NewTenantContext(s.T().Context(), sess.TenantID)
			expiresAt := time.Now().Add(tc.timeLeft)
			s.Require().NoError(sessions.db.Client(tenantCtx).UserAuthSession.UpdateOneID(sess.ID).SetExpiresAt(expiresAt).Exec(tenantCtx))
			usr, userErr := users.Get(tenantCtx, user.ID(sess.UserID))
			s.Require().NoError(userErr)

			exchange := func(_ context.Context, refreshToken string) (*rez.UserAuthProviderSession, error) {
				s.Require().NotNil(tc.exchange, "unexpected exchange")
				s.Equal("refresh-"+usr.AuthProviderID, refreshToken)
				return tc.exchange(refreshed(orgID, usr.AuthProviderID))
			}
			result, refreshErr := sessions.Refresh(s.T().Context(), sess.ID, exchange)
			switch {
			case tc.errIs != nil:
				s.ErrorIs(refreshErr, tc.errIs)
			case tc.kept && !tc.extended && tc.exchange != nil:
				s.Require().Error(refreshErr)
				s.NotErrorIs(refreshErr, errs.ErrAuthSessionInvalid)
			default:
				s.Require().NoError(refreshErr)
				s.Equal(sess.ID, result.ID)
			}

			loaded, lookupErr := sessions.LookupSession(s.T().Context(), sess.ID)
			if !tc.kept {
				s.True(ent.IsNotFound(lookupErr))
				return
			}
			s.Require().NoError(lookupErr)
			if !tc.extended {
				s.WithinDuration(expiresAt, loaded.ExpiresAt, time.Second)
				s.Equal(userauthsession.RoleAdmin, loaded.Role)
				return
			}
			s.WithinDuration(time.Now().Add(time.Hour), loaded.ExpiresAt, time.Minute)
			s.Equal(userauthsession.RoleMember, loaded.Role)
			s.Equal([]byte("rotated"), loaded.ProviderRefreshToken)
			s.Equal(user.RoleMember, s.userRole(users, loaded))
			s.False(execution.GetContext(execution.NewUserContext(s.T().Context(), loaded)).HasScope(execution.ScopeOrganizationManage))
		})
	}
}

func (s *AuthSessionServiceSuite) TestRefreshWithoutRefreshTokenLeavesSession() {
	sessions, _ := s.newServices()
	sess := s.signIn(sessions, "refresh-org-"+uuid.NewString(), "person", user.RoleAdmin)
	tenantCtx := execution.NewTenantContext(s.T().Context(), sess.TenantID)
	expiresAt := time.Now().Add(4 * time.Minute)
	update := sessions.db.Client(tenantCtx).UserAuthSession.UpdateOneID(sess.ID).
		SetExpiresAt(expiresAt).
		ClearProviderRefreshToken()
	s.Require().NoError(update.Exec(tenantCtx))

	exchange := func(context.Context, string) (*rez.UserAuthProviderSession, error) {
		s.Fail("unexpected exchange")
		return nil, nil
	}
	result, refreshErr := sessions.Refresh(s.T().Context(), sess.ID, exchange)
	s.Require().NoError(refreshErr)
	s.WithinDuration(expiresAt, result.ExpiresAt, time.Second)
}
