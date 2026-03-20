package oidc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rezible/rezible/ent"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
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
	clientSecret := rez.Config.GetString("auth.oidc_client_secret")
	if clientId == "" || clientSecret == "" {
		return fmt.Errorf("auth.oidc_client_id, auth.oidc_client_secret must be set")
	}

	// Fetch Dex's public keys from: {dexIssuer}/.well-known/openid-configuration
	// Cache and auto-rotate the JWKS
	issuerUrl := rez.Config.GetString("auth.issuer_url")
	if len(issuerUrl) == 0 {
		return fmt.Errorf("auth issuer url must be set")
	}

	prov, provErr := oidc.NewProvider(ctx, issuerUrl)
	if provErr != nil {
		return fmt.Errorf("failed to create oidc provider: %w", provErr)
	}
	s.provider = prov

	s.oauthConfig = oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Endpoint:     prov.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
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

func (s *AuthService) CreateClientAuthSession(ctx context.Context, token string, verifier string) ([]http.Cookie, error) {
	tokenResp, exchangeErr := s.oauthConfig.Exchange(ctx, token, oauth2.VerifierOption(verifier))
	if exchangeErr != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", exchangeErr)
	}
	return oapiv1.MakeAuthSessionCookies(*tokenResp), nil
}

func (s *AuthService) RefreshClientAuthSession(ctx context.Context, refreshCookie http.Cookie) ([]http.Cookie, error) {
	// TODO
	return nil, nil
}

func (s *AuthService) ClearClientAuthSession(ctx context.Context) ([]http.Cookie, error) {
	// TODO
	return nil, nil
}

func (s *AuthService) CreateVerifiedAuthSessionContext(ctx context.Context, rawIdToken string) (context.Context, error) {
	idToken, verifyErr := s.tokenVerifier.Verify(ctx, rawIdToken)
	if verifyErr != nil {
		return nil, rez.ErrAuthSessionUnauthorized
	}
	return s.createAuthSessionContext(ctx, idToken)
}

type IdTokenClaims struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}

type authUserSessionContextKey struct{}

func (s *AuthService) createAuthSessionContext(ctx context.Context, idToken *oidc.IDToken) (context.Context, error) {
	var claims IdTokenClaims
	if claimsErr := idToken.Claims(&claims); claimsErr != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", claimsErr)
	}

	if len(claims.Email) == 0 {
		return nil, rez.ErrInvalidUser
	}
	provUser := ent.User{
		Email: claims.Email,
	}

	usr, usrErr := s.users.FindOrCreateAuthProviderUser(ctx, provUser)
	if usrErr != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", usrErr)
	}

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
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}
