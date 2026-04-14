package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type oauthHandler struct {
	idVerifier  *oidc.IDTokenVerifier
	audienceOpt oauth2.AuthCodeOption
	oauthCfg    oauth2.Config
}

func makeOAuthHandler(ctx context.Context, cfg oidcConfig) (*oauthHandler, error) {
	apiAudience := rez.Config.ApiUrl()
	if apiAudience == "" {
		return nil, fmt.Errorf("no api url configured, can't verify token audience")
	}
	oidcProv, oidcProvErr := oidc.NewProvider(ctx, cfg.Issuer)
	if oidcProvErr != nil {
		return nil, fmt.Errorf("create oidc provider: %w", oidcProvErr)
	}
	redirectUri, redirectErr := url.JoinPath(rez.Config.AppUrl(), "/api/auth/callback")
	if redirectErr != nil {
		return nil, fmt.Errorf("redirect url: %w", redirectErr)
	}
	return &oauthHandler{
		oauthCfg: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile", "email"},
			RedirectURL:  redirectUri,
			Endpoint:     oidcProv.Endpoint(),
		},
		audienceOpt: oauth2.SetAuthURLParam("resource", apiAudience),
		// audience for id token is the client application, not resource
		idVerifier: oidcProv.VerifierContext(ctx, &oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

type AuthVerifyState struct {
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURL  string `json:"redirect_url"`
}

func (p *oauthHandler) createAuthRedirect() (string, *AuthVerifyState, error) {
	state, nonce, randErr := createRandomValues()
	if randErr != nil {
		return "", nil, randErr
	}
	codeVerifier := oauth2.GenerateVerifier()
	vs := &AuthVerifyState{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		RedirectURL:  "/",
	}
	authURL := p.oauthCfg.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(codeVerifier),
		oidc.Nonce(nonce),
		p.audienceOpt,
	)
	return authURL, vs, nil
}

func (p *oauthHandler) doCallbackTokenExchange(r *http.Request, state string, verifier string) (*oauth2.Token, error) {
	q := r.URL.Query()
	code := q.Get("code")
	if code == "" {
		return nil, fmt.Errorf("missing code")
	}
	if q.Get("state") != state {
		return nil, fmt.Errorf("invalid state")
	}

	token, exchangeErr := p.oauthCfg.Exchange(r.Context(), code, oauth2.VerifierOption(verifier))
	if exchangeErr != nil {
		return nil, fmt.Errorf("token exchange failed: %w", exchangeErr)
	}
	return token, nil
}

type IdTokenClaims struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	OrgId   string `json:"org_id"`
	OrgName string `json:"org_name"`
}

func (p *oauthHandler) verifyIdToken(ctx context.Context, idTokenStr string, nonce string) (*verifiedSessionToken, error) {
	idToken, err := p.idVerifier.Verify(ctx, idTokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid id_token: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, fmt.Errorf("invalid nonce")
	}

	var claims IdTokenClaims
	if claimsErr := idToken.Claims(&claims); claimsErr != nil {
		return nil, rez.ErrAuthSessionInvalid
	}
	return &verifiedSessionToken{id: idToken, claims: claims}, nil
}

func (p *oauthHandler) refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	expiredTokenSource := p.oauthCfg.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second), // set expired to force a refresh request
	})
	return expiredTokenSource.Token()
}

type verifiedSessionToken struct {
	id     *oidc.IDToken
	claims IdTokenClaims
}

func (t *verifiedSessionToken) getUser() ent.User {
	return ent.User{
		AuthProviderID: t.claims.Sub,
		Email:          t.claims.Email,
		Name:           t.claims.Name,
	}
}

func (t *verifiedSessionToken) getOrganization() ent.Organization {
	return ent.Organization{
		AuthProviderID: t.claims.OrgId,
		Name:           t.claims.OrgName,
	}
}
