package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
)

type Config struct {
	SessionSecret []byte `koanf:"session_secret"`
	Oidc          struct {
		ClientId  string   `koanf:"client_id"`
		IssuerUrl string   `koanf:"issuer_url"`
		Scopes    []string `koanf:"scopes"`
	} `koanf:"oidc"`
}

type AuthService struct {
	orgs  rez.OrganizationService
	users rez.UserService
	teams rez.TeamService

	cfg           Config
	oauthConfig   oauth2.Config
	provider      *oidc.Provider
	tokenVerifier *oidc.IDTokenVerifier
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService, teams rez.TeamService) (*AuthService, error) {
	s := &AuthService{
		orgs:  orgs,
		users: users,
		teams: teams,
	}

	var cfg Config
	if cfgErr := rez.Config.Unmarshal("auth", &cfg); cfgErr != nil {
		return nil, fmt.Errorf("auth config error: %w", cfgErr)
	}
	s.cfg = cfg

	if providerErr := s.initOidcProvider(ctx); providerErr != nil {
		return nil, fmt.Errorf("failed to init oidc provider: %w", providerErr)
	}

	if oauthCfgErr := s.initOAuthConfig(); oauthCfgErr != nil {
		return nil, fmt.Errorf("failed to init oauth config: %w", oauthCfgErr)
	}

	return s, nil
}

func (s *AuthService) initOidcProvider(ctx context.Context) error {
	prov, provErr := oidc.NewProvider(ctx, s.cfg.Oidc.IssuerUrl)
	if provErr != nil {
		return fmt.Errorf("failed to create oidc provider: %w", provErr)
	}
	s.provider = prov

	s.tokenVerifier = prov.VerifierContext(ctx, &oidc.Config{
		ClientID:                   s.cfg.Oidc.ClientId,
		SkipClientIDCheck:          false,
		SkipExpiryCheck:            false,
		SkipIssuerCheck:            false,
		InsecureSkipSignatureCheck: false,
	})

	return nil
}

func (s *AuthService) initOAuthConfig() error {
	redirectUrl, redirectErr := url.JoinPath(rez.Config.AppUrl(), "auth", "callback")
	if redirectErr != nil {
		return fmt.Errorf("failed to create oidc redirect url: %w", redirectErr)
	}

	const defaultScopes = "openid offline_access profile email groups federated:id federated_claims"
	scopes := s.cfg.Oidc.Scopes
	if len(scopes) == 0 {
		scopes = strings.Split(defaultScopes, " ")
	}

	s.oauthConfig = oauth2.Config{
		ClientID:    s.cfg.Oidc.ClientId,
		Endpoint:    s.provider.Endpoint(),
		RedirectURL: redirectUrl,
		Scopes:      scopes,
	}

	return nil
}

func (s *AuthService) CreateAuthSessionContext(ctx context.Context, idTokenStr string) (context.Context, error) {
	return s.createAuthSessionContext(ctx, idTokenStr, false)
}

func (s *AuthService) GetAuthSession(ctx context.Context) rez.AuthSession {
	return s.getAuthSessionContext(ctx)
}

func (s *AuthService) CompleteClientAuthSessionFlow(ctx context.Context, code string, verifier string) ([]http.Cookie, error) {
	tokenResp, exchangeErr := s.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if exchangeErr != nil {
		return nil, fmt.Errorf("exchange token: %w", exchangeErr)
	}
	authCtx, ctxErr := s.createAuthSessionContext(ctx, tokenResp.AccessToken, true)
	if ctxErr != nil {
		return nil, fmt.Errorf("create auth session context: %w", ctxErr)
	}
	return oapiv1.MakeAuthSessionCookies(authCtx, *tokenResp), nil
}

func (s *AuthService) RefreshClientAuthSession(ctx context.Context, refreshToken string) ([]http.Cookie, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh cookie provided")
	}
	freshToken, refreshErr := s.fetchRefreshedToken(ctx, refreshToken)
	if refreshErr != nil {
		log.Error().Err(refreshErr).Msg("exchange token")
		return nil, fmt.Errorf("fetch refreshed token: %w", refreshErr)
	}
	return oapiv1.MakeAuthSessionCookies(ctx, *freshToken), nil
}

func (s *AuthService) fetchRefreshedToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	expiredToken := &oauth2.Token{
		RefreshToken: refreshToken,
		// set expired to force refetch
		Expiry: time.Now().Add(-time.Second),
	}
	tokenSource := s.oauthConfig.TokenSource(ctx, expiredToken)
	newToken, exchangeErr := tokenSource.Token()
	if exchangeErr != nil {
		return nil, fmt.Errorf("force refresh token exchange: %w", exchangeErr)
	}
	return newToken, nil
}

func (s *AuthService) ClearClientAuthSession() ([]http.Cookie, error) {
	return oapiv1.MakeLogoutAuthSessionCookies(), nil
}

type IdTokenClaims struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type verifiedIdToken struct {
	idToken *oidc.IDToken
	claims  IdTokenClaims
}

func (t *verifiedIdToken) getDomain() string {
	if parts := strings.Split(t.claims.Email, "@"); len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func (t *verifiedIdToken) getUser() ent.User {
	return ent.User{
		AuthProviderID: t.idToken.Subject,
		Email:          t.claims.Email,
		Name:           t.claims.Name,
	}
}

func (s *AuthService) getVerifiedIdToken(ctx context.Context, idTokenStr string) (*verifiedIdToken, error) {
	if idTokenStr == "" {
		return nil, rez.ErrAuthSessionMissing
	}
	idToken, verifyErr := s.tokenVerifier.Verify(ctx, idTokenStr)
	if verifyErr != nil {
		return nil, fmt.Errorf("verify: %w", verifyErr)
	}
	var claims IdTokenClaims
	if claimsErr := idToken.Claims(&claims); claimsErr != nil {
		return nil, fmt.Errorf("get claims: %w", claimsErr)
	}
	return &verifiedIdToken{idToken, claims}, nil
}

type authUserSessionContextKey struct{}

func (s *AuthService) createAuthSessionContext(ctx context.Context, idTokenStr string, create bool) (context.Context, error) {
	token, tokenErr := s.getVerifiedIdToken(ctx, idTokenStr)
	if tokenErr != nil {
		return nil, fmt.Errorf("get verified id token: %w", tokenErr)
	}

	var usrErr error
	var usr *ent.User
	if create {
		org, orgErr := s.orgs.FindOrCreateFromDomain(ctx, token.getDomain())
		if orgErr != nil {
			return nil, fmt.Errorf("find org: %w", orgErr)
		}
		ctx = access.WithOrganization(ctx, org)

		usr, usrErr = s.users.FindOrCreateFromAuth(ctx, token.getUser())
		if usrErr != nil {
			return nil, fmt.Errorf("find or create auth user: %w", usrErr)
		}
	} else {
		usr, usrErr = s.users.LookupUserByAuthProviderId(access.SystemContext(ctx), token.getUser().AuthProviderID)
		if usrErr != nil {
			return nil, fmt.Errorf("lookup auth provider user: %w", usrErr)
		}
	}
	ctx = access.WithUser(ctx, usr)

	return context.WithValue(ctx, authUserSessionContextKey{}, newAuthSession(usr, token)), nil
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
