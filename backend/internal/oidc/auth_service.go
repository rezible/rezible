package oidc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
)

type AuthService struct {
	orgs          rez.OrganizationService
	users         rez.UserService
	oauthConfig   oauth2.Config
	provider      *oidc.Provider
	tokenVerifier *oidc.IDTokenVerifier
	sessionSecret []byte
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (*AuthService, error) {
	s := &AuthService{orgs: orgs, users: users}

	secretKey := []byte(rez.Config.GetString("auth.session_secret_key"))
	if len(secretKey) == 0 {
		return nil, fmt.Errorf("auth session secret key must be set")
	}
	s.sessionSecret = secretKey

	if providerErr := s.initProvider(ctx); providerErr != nil {
		return nil, fmt.Errorf("failed to init oidc provider: %w", providerErr)
	}

	return s, nil
}

func (s *AuthService) initProvider(ctx context.Context) error {
	clientId := rez.Config.GetStringOr("auth.oidc_client_id", "rezible-app")
	if clientId == "" {
		return fmt.Errorf("auth.oidc_client_id, auth.oidc_client_secret must be set")
	}

	issuerUrl := rez.Config.GetString("auth.oidc_issuer_url")
	if len(issuerUrl) == 0 {
		return fmt.Errorf("auth issuer url must be set")
	}

	prov, provErr := oidc.NewProvider(ctx, issuerUrl)
	if provErr != nil {
		return fmt.Errorf("failed to create oidc provider: %w", provErr)
	}
	s.provider = prov

	s.oauthConfig = oauth2.Config{
		ClientID:    clientId,
		Endpoint:    prov.Endpoint(),
		RedirectURL: rez.Config.AppUrl() + "/auth/callback",
		Scopes: []string{
			oidc.ScopeOpenID,
			oidc.ScopeOfflineAccess,
			"email",
			"profile",
			"groups",
			"federated:id",
			"federated_claims",
		},
	}
	s.tokenVerifier = prov.VerifierContext(ctx, &oidc.Config{
		ClientID:                   clientId,
		SkipClientIDCheck:          false,
		SkipExpiryCheck:            false,
		SkipIssuerCheck:            false,
		InsecureSkipSignatureCheck: false,
	})

	return nil
}

func (s *AuthService) CompleteClientAuthSessionFlow(ctx context.Context, code string, verifier string) ([]http.Cookie, error) {
	tokenResp, exchangeErr := s.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if exchangeErr != nil {
		log.Error().Err(exchangeErr).Msg("failed to exchange token")
		return nil, fmt.Errorf("failed to exchange token: %w", exchangeErr)
	}
	authCtx, ctxErr := s.CreateVerifiedAuthSessionContext(ctx, tokenResp.AccessToken)
	if ctxErr != nil {
		return nil, fmt.Errorf("failed to create auth session context: %w", ctxErr)
	}
	return oapiv1.MakeAuthSessionCookies(authCtx, *tokenResp), nil
}

func (s *AuthService) RefreshClientAuthSession(ctx context.Context, refreshCookie http.Cookie) ([]http.Cookie, error) {
	if refreshCookie.Value == "" {
		return nil, fmt.Errorf("no refresh cookie provided")
	}
	expiredToken := &oauth2.Token{
		RefreshToken: refreshCookie.Value,
		Expiry:       time.Now().Add(-time.Second),
	}
	tokenSource := s.oauthConfig.TokenSource(ctx, expiredToken)
	newToken, exchangeErr := tokenSource.Token()
	if exchangeErr != nil {
		log.Error().Err(exchangeErr).Msg("failed to exchange token")
		return nil, fmt.Errorf("failed to exchange token: %w", exchangeErr)
	}
	return oapiv1.MakeAuthSessionCookies(ctx, *newToken), nil
}

func (s *AuthService) ClearClientAuthSession(ctx context.Context) ([]http.Cookie, error) {
	return oapiv1.MakeLogoutAuthSessionCookies(), nil
}

type IdTokenClaims struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}

func (c IdTokenClaims) getDomain() string {
	emailParts := strings.Split(c.Email, "@")
	if len(emailParts) != 2 {
		return ""
	}
	return emailParts[1]
}

func (s *AuthService) getUserFromTokenClaims(claims IdTokenClaims) ent.User {
	return ent.User{AuthProviderID: claims.Sub, Email: claims.Email}
}

type authUserSessionContextKey struct{}

func (s *AuthService) CreateVerifiedAuthSessionContext(ctx context.Context, idTokenStr string) (context.Context, error) {
	if idTokenStr == "" {
		return nil, rez.ErrAuthSessionMissing
	}
	idToken, verifyErr := s.tokenVerifier.Verify(ctx, idTokenStr)
	if verifyErr != nil {
		return nil, fmt.Errorf("failed to verify token: %w", verifyErr)
	}

	var claims IdTokenClaims
	if claimsErr := idToken.Claims(&claims); claimsErr != nil {
		return nil, fmt.Errorf("id token claims: %w", claimsErr)
	}

	ctx = access.SystemContext(ctx)
	org, orgErr := s.orgs.FindOrCreateFromProviderDomain(ctx, claims.getDomain())
	if orgErr != nil {
		return nil, fmt.Errorf("find org: %w", orgErr)
	}
	ctx = access.TenantContext(ctx, org.TenantID)

	usr, usrErr := s.users.FindOrCreateAuthProviderUser(ctx, s.getUserFromTokenClaims(claims))
	if usrErr != nil {
		return nil, fmt.Errorf("find user: %w", usrErr)
	}
	ctx = s.users.CreateUserAccessContext(ctx, usr)

	sess := &rez.AuthSession{
		UserId:    usr.ID,
		Scopes:    claims.Scopes,
		ExpiresAt: idToken.Expiry,
	}
	return context.WithValue(ctx, authUserSessionContextKey{}, sess), nil
}

func (s *AuthService) GetAuthSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authUserSessionContextKey{}).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrAuthSessionMissing
	}
	return sess, nil
}
