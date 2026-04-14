package oidc

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type tokenProvider struct {
	idVerifier   *oidc.IDTokenVerifier
	audienceOpt  oauth2.AuthCodeOption
	appClientCfg oauth2.Config
}

func makeTokenProvider(ctx context.Context, cfg oidcConfig) (*tokenProvider, error) {
	apiAudience := rez.Config.ApiUrl()
	if apiAudience == "" {
		return nil, fmt.Errorf("no api url configured, can't verify token audience")
	}
	oidcProv, oidcProvErr := oidc.NewProvider(ctx, cfg.Issuer)
	if oidcProvErr != nil {
		return nil, fmt.Errorf("create oidc provider: %w", oidcProvErr)
	}
	appClientId := cfg.AppClient.Id
	return &tokenProvider{
		appClientCfg: oauth2.Config{
			ClientID:    appClientId,
			Scopes:      cfg.AppClient.Scopes,
			RedirectURL: cfg.AppClient.RedirectUri,
			Endpoint:    oidcProv.Endpoint(),
		},
		audienceOpt: oauth2.SetAuthURLParam("resource", apiAudience),
		// audience for id token is the client application, not resource
		idVerifier: oidcProv.VerifierContext(ctx, &oidc.Config{ClientID: appClientId}),
	}, nil
}

func (p *tokenProvider) exchangeAppClientAuthCode(ctx context.Context, code string, verifier string) (*oauth2.Token, error) {
	return p.appClientCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier), p.audienceOpt)
}

func (p *tokenProvider) refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	expiredTokenSource := p.appClientCfg.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second), // set expired to force a refresh request
	})
	return expiredTokenSource.Token()
}

func (p *tokenProvider) verifyIdToken(ctx context.Context, idTokenStr string) (*verifiedSessionToken, error) {
	if idTokenStr == "" {
		return nil, rez.ErrAuthSessionMissing
	}
	idToken, verifyErr := p.idVerifier.Verify(ctx, idTokenStr)
	if verifyErr != nil {
		log.Debug().Err(verifyErr).Msg("id token verification failed")
		return nil, rez.ErrAuthSessionInvalid
	}
	var claims IdTokenClaims
	if claimsErr := idToken.Claims(&claims); claimsErr != nil {
		log.Debug().Err(verifyErr).Msg("id token claims failed")
		return nil, rez.ErrAuthSessionInvalid
	}
	return &verifiedSessionToken{idToken, claims}, nil
}

type IdTokenClaims struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type verifiedSessionToken struct {
	id     *oidc.IDToken
	claims IdTokenClaims
}

func (t *verifiedSessionToken) getUser() ent.User {
	return ent.User{
		AuthProviderID: t.id.Subject,
		Email:          t.claims.Email,
		Name:           t.claims.Name,
		Edges:          ent.UserEdges{},
	}
}

func (t *verifiedSessionToken) getOrganization() ent.Organization {
	return ent.Organization{
		Name: "test org",
	}
}
