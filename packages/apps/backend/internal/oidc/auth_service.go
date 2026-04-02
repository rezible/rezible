package oidc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
)

type Config struct {
	SessionSecret []byte            `koanf:"session_secret"`
	Oidc          configOidc        `koanf:"oidc"`
	AllowAccess   configAllowAccess `koanf:"allow_access"`
}

type configOidc struct {
	IssuerUrl         string   `koanf:"issuer_url"`
	ClientId          string   `koanf:"client_id"`
	ClientRedirectUri string   `koanf:"client_redirect_uri"`
	ClientScopes      []string `koanf:"client_scopes"`
}
type configAllowAccess struct {
	Public         bool     `koanf:"public_signup"`
	Domains        []string `koanf:"domains"`
	allowedDomains mapset.Set[string]
}

func (a *configAllowAccess) DomainAllowed(domain string) bool {
	return a.Public || a.allowedDomains.Contains(domain)
}

func loadConfig() (*Config, error) {
	var defaultClientScopes = []string{"openid", "offline_access", "profile", "email", "groups", "federated:id", "federated_claims"}
	cfg := Config{
		Oidc: configOidc{ClientScopes: defaultClientScopes},
	}
	if cfgErr := rez.Config.Unmarshal("auth", &cfg); cfgErr != nil {
		return nil, cfgErr
	}
	if !cfg.AllowAccess.Public {
		cfg.AllowAccess.allowedDomains = mapset.NewSet(cfg.AllowAccess.Domains...)
	}
	return &cfg, nil
}

type AuthService struct {
	orgs  rez.OrganizationService
	users rez.UserService
	teams rez.TeamService

	cfg      Config
	provider *tokenProvider
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService, teams rez.TeamService) (*AuthService, error) {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	s := &AuthService{
		orgs:  orgs,
		users: users,
		teams: teams,
		cfg:   *cfg,
	}

	if _, provErr := s.getTokenProvider(ctx); provErr != nil {
		log.Warn().Err(provErr).Msg("failed to load token provider")
	}

	return s, nil
}

func (s *AuthService) getTokenProvider(ctx context.Context) (*tokenProvider, error) {
	if s.provider == nil {
		prov, provErr := s.cfg.Oidc.makeTokenProvider(ctx)
		if provErr != nil {
			return nil, fmt.Errorf("make token provider: %w", provErr)
		}
		s.provider = prov
	}
	return s.provider, nil
}

func (s *AuthService) CreateAuthSessionContext(ctx context.Context, idTokenStr string) (context.Context, error) {
	return s.createAuthSessionContext(ctx, idTokenStr, false)
}

func (s *AuthService) GetAuthSession(ctx context.Context) rez.AuthSession {
	return s.getAuthSessionContext(ctx)
}

func (s *AuthService) CompleteClientAuthSessionFlow(ctx context.Context, code string, verifier string) ([]http.Cookie, error) {
	prov, provErr := s.getTokenProvider(ctx)
	if provErr != nil {
		return nil, provErr
	}
	token, exchangeErr := prov.exchangeToken(ctx, code, verifier)
	if exchangeErr != nil {
		return nil, fmt.Errorf("exchange token: %w", exchangeErr)
	}
	authCtx, ctxErr := s.createAuthSessionContext(ctx, token.AccessToken, true)
	if ctxErr != nil {
		return nil, fmt.Errorf("create auth session context: %w", ctxErr)
	}
	return oapiv1.MakeAuthSessionCookies(authCtx, *token), nil
}

func (s *AuthService) RefreshClientAuthSession(ctx context.Context, refreshToken string) ([]http.Cookie, error) {
	prov, provErr := s.getTokenProvider(ctx)
	if provErr != nil {
		return nil, provErr
	}
	freshToken, refreshErr := prov.refreshToken(ctx, refreshToken)
	if refreshErr != nil {
		return nil, fmt.Errorf("fetch refreshed token: %w", refreshErr)
	}
	return oapiv1.MakeAuthSessionCookies(ctx, *freshToken), nil
}

func (s *AuthService) ClearClientAuthSession() ([]http.Cookie, error) {
	return oapiv1.MakeLogoutAuthSessionCookies(), nil
}

type authUserSessionContextKey struct{}

func (s *AuthService) createAuthSessionContext(ctx context.Context, idTokenStr string, create bool) (context.Context, error) {
	if idTokenStr == "" {
		return nil, rez.ErrAuthSessionMissing
	}
	prov, provErr := s.getTokenProvider(ctx)
	if provErr != nil {
		return nil, provErr
	}
	token, tokenErr := prov.verifyIdToken(ctx, idTokenStr)
	if tokenErr != nil {
		log.Debug().Err(tokenErr).Msg("failed to verify id token")
		return nil, rez.ErrAuthSessionInvalid
	}

	if !s.cfg.AllowAccess.DomainAllowed(token.getDomain()) {
		return nil, rez.ErrDomainNotAllowed
	}

	usr, usrErr := s.matchAuthUser(ctx, token, create)
	if usrErr != nil {
		return nil, fmt.Errorf("match auth user: %w", usrErr)
	}
	return s.makeAuthSessionContext(ctx, usr, token), nil
}

func (s *AuthService) matchAuthUser(ctx context.Context, token *verifiedIdToken, create bool) (*ent.User, error) {
	if !create {
		usr, usrErr := s.users.Get(access.SystemContext(ctx), user.AuthProviderID(token.getUser().AuthProviderID))
		if usrErr != nil {
			return nil, fmt.Errorf("lookup auth provider user: %w", usrErr)
		}
		return usr, nil
	}

	org, orgErr := s.orgs.FindOrCreateFromDomain(ctx, token.getDomain())
	if orgErr != nil {
		return nil, fmt.Errorf("find org: %w", orgErr)
	}
	ctx = access.TenantContext(ctx, org.TenantID)

	usr, usrErr := s.users.FindOrCreateAuthProviderUser(ctx, token.getUser())
	if usrErr != nil {
		return nil, fmt.Errorf("find or create auth user: %w", usrErr)
	}
	return usr, nil
}

func (s *AuthService) makeAuthSessionContext(ctx context.Context, usr *ent.User, token *verifiedIdToken) context.Context {
	return context.WithValue(access.WithUser(ctx, usr), authUserSessionContextKey{}, newAuthSession(usr, token))
}

func (s *AuthService) getAuthSessionContext(ctx context.Context) AuthSession {
	if sess, ok := ctx.Value(authUserSessionContextKey{}).(AuthSession); ok {
		return sess
	}
	return AuthSession{}
}

type AuthSession struct {
	userId    uuid.UUID
	scopes    []string
	expiresAt time.Time
}

func newAuthSession(u *ent.User, t *verifiedIdToken) AuthSession {
	return AuthSession{
		userId:    u.ID,
		scopes:    t.claims.Scopes,
		expiresAt: t.idToken.Expiry,
	}
}

func (a AuthSession) UserId() uuid.UUID {
	return a.userId
}

func (a AuthSession) Scopes() []string {
	return a.scopes
}

func (a AuthSession) ExpiresAt() time.Time {
	return a.expiresAt
}
