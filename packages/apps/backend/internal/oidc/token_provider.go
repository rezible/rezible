package oidc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rezible/rezible/ent"
	"golang.org/x/oauth2"
)

type tokenProvider struct {
	oauthConfig oauth2.Config
	verifier    *oidc.IDTokenVerifier
}

func (cfg *configOidc) makeTokenProvider(ctx context.Context) (*tokenProvider, error) {
	oidcProv, oidcProvErr := oidc.NewProvider(ctx, cfg.IssuerUrl)
	if oidcProvErr != nil {
		return nil, fmt.Errorf("create oidc provider: %w", oidcProvErr)
	}

	return &tokenProvider{
		verifier: oidcProv.VerifierContext(ctx, &oidc.Config{ClientID: cfg.ClientId}),
		oauthConfig: oauth2.Config{
			ClientID:    cfg.ClientId,
			RedirectURL: cfg.ClientRedirectUri,
			Scopes:      cfg.ClientScopes,
			Endpoint:    oidcProv.Endpoint(),
		},
	}, nil
}

func (p *tokenProvider) exchangeToken(ctx context.Context, code string, verifier string) (*oauth2.Token, error) {
	return p.oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}

func (p *tokenProvider) refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	expiredTokenSource := p.oauthConfig.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second), // set expired to force a refresh request
	})
	return expiredTokenSource.Token()
}

func (p *tokenProvider) verifyIdToken(ctx context.Context, idTokenStr string) (*verifiedIdToken, error) {
	idToken, verifyErr := p.verifier.Verify(ctx, idTokenStr)
	if verifyErr != nil {
		return nil, fmt.Errorf("verify: %w", verifyErr)
	}
	var claims IdTokenClaims
	if claimsErr := idToken.Claims(&claims); claimsErr != nil {
		return nil, fmt.Errorf("get claims: %w", claimsErr)
	}
	return &verifiedIdToken{idToken, claims}, nil
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
