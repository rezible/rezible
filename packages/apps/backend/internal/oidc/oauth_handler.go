package oidc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	orgScope = "organization"
)

type oauthHandler struct {
	cfg oidcConfig

	singleTenantOrg *ent.Organization

	resourceOption      oauth2.AuthCodeOption
	provider            *oidc.Provider
	oauthCfg            *oauth2.Config
	accessTokenVerifier *oidc.IDTokenVerifier
	idTokenVerifier     *oidc.IDTokenVerifier
}

func makeOAuthHandler(ctx context.Context, cfg Config) (*oauthHandler, error) {
	apiAudience := rez.Config.ApiUrl()
	if apiAudience == "" {
		return nil, fmt.Errorf("no api url configured, can't verify token audience")
	}

	h := &oauthHandler{
		cfg:            cfg.Oidc,
		resourceOption: oauth2.SetAuthURLParam("resource", apiAudience),
	}

	if rez.Config.SingleTenantMode() {
		h.singleTenantOrg = &ent.Organization{
			AuthProviderID: "default",
			Name:           cfg.SingleTenantOrgName,
		}
	}

	return h, nil
}

func (h *oauthHandler) scopes() []string {
	scopes := []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile", "email"}
	if h.singleTenantOrg == nil {
		scopes = append(scopes, orgScope)
	}
	return scopes
}

func (h *oauthHandler) ensureProvider(ctx context.Context) error {
	if h.provider == nil {
		var provErr error
		h.provider, provErr = oidc.NewProvider(ctx, h.cfg.Issuer)
		if provErr != nil {
			return fmt.Errorf("create oidc provider: %w", provErr)
		}
		accessTokenAudience := rez.Config.ApiUrl()
		if h.singleTenantOrg != nil {
			accessTokenAudience = h.cfg.ClientID
		}
		h.oauthCfg = &oauth2.Config{
			ClientID:     h.cfg.ClientID,
			ClientSecret: h.cfg.ClientSecret,
			Scopes:       h.scopes(),
			RedirectURL:  h.cfg.RedirectUrl,
			Endpoint:     h.provider.Endpoint(),
		}
		h.accessTokenVerifier = h.provider.VerifierContext(ctx, &oidc.Config{ClientID: accessTokenAudience})
		h.idTokenVerifier = h.provider.VerifierContext(ctx, &oidc.Config{ClientID: h.oauthCfg.ClientID})
	}
	return nil
}

type AuthFlowState struct {
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"code_verifier"`
	ReturnTo     string `json:"return_to"`
}

func (h *oauthHandler) createAuthRedirect(r *http.Request) (string, *AuthFlowState, error) {
	if cfgErr := h.ensureProvider(r.Context()); cfgErr != nil {
		return "", nil, cfgErr
	}
	state, nonce, randErr := createRandomValues()
	if randErr != nil {
		return "", nil, randErr
	}
	returnTo := r.URL.Query().Get("return_to")
	if returnTo == "" {
		returnTo = "/"
	}
	if !strings.HasPrefix(returnTo, "/") || strings.HasPrefix(returnTo, "//") {
		return "", nil, fmt.Errorf("invalid return_to")
	}
	codeVerifier := oauth2.GenerateVerifier()
	// TODO: encode this as the oauth state itself? (instead of a cookie)
	vs := &AuthFlowState{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ReturnTo:     returnTo,
	}
	authURL := h.oauthCfg.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(codeVerifier),
		h.resourceOption,
	)
	return authURL, vs, nil
}

func (h *oauthHandler) doCallbackExchange(r *http.Request, s AuthFlowState) (*userAuthInfo, error) {
	q := r.URL.Query()
	code := q.Get("code")
	if code == "" {
		return nil, fmt.Errorf("missing code")
	}
	if q.Get("state") != s.State {
		return nil, fmt.Errorf("invalid state")
	}
	ctx := r.Context()
	if cfgErr := h.ensureProvider(ctx); cfgErr != nil {
		return nil, cfgErr
	}
	token, exchangeErr := h.oauthCfg.Exchange(ctx, code,
		oauth2.VerifierOption(s.CodeVerifier), h.resourceOption)
	if exchangeErr != nil {
		return nil, fmt.Errorf("token exchange failed: %w", exchangeErr)
	}
	if !token.Valid() {
		return nil, fmt.Errorf("invalid token")
	}
	return h.extractTokenInfo(ctx, token, s.Nonce)
}

func (h *oauthHandler) refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if cfgErr := h.ensureProvider(ctx); cfgErr != nil {
		return nil, cfgErr
	}
	expiredTokenSource := h.oauthCfg.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second), // set expired to force a refresh request
	})
	return expiredTokenSource.Token()
}

type userAuthInfo struct {
	expiresAt time.Time
	user      ent.User
	org       ent.Organization
}

type accessTokenClaims struct {
	Scopes           []string `json:"scopes"`
	OrganizationId   string   `json:"rez-org-id"`
	OrganizationName string   `json:"rez-org-name"`
}

type idTokenClaims struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	OrgId   string `json:"org_id"`
	OrgName string `json:"org_name"`
}

func (h *oauthHandler) extractTokenInfo(ctx context.Context, t *oauth2.Token, nonce string) (*userAuthInfo, error) {
	at, atErr := h.accessTokenVerifier.Verify(ctx, t.AccessToken)
	if atErr != nil {
		return nil, fmt.Errorf("access token verification failed: %w", atErr)
	}
	var atClaims accessTokenClaims
	if atClaimsErr := at.Claims(&atClaims); atClaimsErr != nil {
		return nil, fmt.Errorf("failed to parse access token claims: %w", atClaimsErr)
	}

	idTokenStr, idOk := t.Extra("id_token").(string)
	if !idOk {
		return nil, fmt.Errorf("no id_token")
	}
	id, idTokenErr := h.idTokenVerifier.Verify(ctx, idTokenStr)
	if idTokenErr != nil {
		return nil, fmt.Errorf("verify id token: %w", idTokenErr)
	}
	if id.Nonce != nonce {
		return nil, fmt.Errorf("invalid id token nonce")
	}
	if hashErr := id.VerifyAccessToken(t.AccessToken); hashErr != nil {
		return nil, fmt.Errorf("verify access token: %w", hashErr)
	}

	var idClaims idTokenClaims
	if claimsErr := id.Claims(&idClaims); claimsErr != nil {
		return nil, rez.ErrAuthSessionInvalid
	}

	info := &userAuthInfo{
		expiresAt: at.Expiry,
		user: ent.User{
			AuthProviderID: idClaims.Sub,
			Email:          idClaims.Email,
			Name:           idClaims.Name,
		},
		org: ent.Organization{
			AuthProviderID: atClaims.OrganizationId,
			Name:           atClaims.OrganizationName,
		},
	}

	if h.singleTenantOrg != nil {
		log.Debug().Msg("using single tenant organization")
		info.org = *h.singleTenantOrg
	}

	return info, nil
}
