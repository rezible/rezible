package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/errs"
)

const (
	orgScope = "org"
)

type oidcHandler struct {
	oidcIssuer          string
	authStateCookiePath string
	multiTenant         bool

	codec *cookieCodec

	// resourceOptions name the API as the access token's resource; empty in single-tenant mode.
	resourceOptions []oauth2.AuthCodeOption
	provider        *oidc.Provider
	oauthCfg        *oauth2.Config

	accessTokenCfg      *oidc.Config
	accessTokenVerifier *oidc.IDTokenVerifier

	idTokenCfg      *oidc.Config
	idTokenVerifier *oidc.IDTokenVerifier
}

func newOidcHandler(cfg rez.Config) (*oidcHandler, error) {
	oauthRedirectUrl := cfg.HttpServer.Auth.Oidc.RedirectUrl
	if oauthRedirectUrl == "" {
		feRedirectUrl, urlErr := cfg.App.GetFrontendUrl(cfg.App.FrontendApiPath, "/auth/callback")
		if urlErr != nil {
			return nil, fmt.Errorf("oauth redirect url: %w", urlErr)
		}
		oauthRedirectUrl = feRedirectUrl.String()
	}

	codec, codecErr := newCookieCodec(cfg.HttpServer.Auth.SessionSecret)
	if codecErr != nil {
		return nil, fmt.Errorf("cookie codec: %w", codecErr)
	}

	multiTenant := !cfg.App.SingleTenant.Enabled
	accessTokenAudience := cfg.HttpServer.Auth.Oidc.ClientID
	scopes := []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "profile", "email"}
	var resourceOptions []oauth2.AuthCodeOption
	if multiTenant {
		apiResource := "https://" + cfg.App.ApiDomain
		accessTokenAudience = apiResource
		scopes = append(scopes, orgScope)
		resourceOptions = append(resourceOptions, oauth2.SetAuthURLParam("resource", apiResource))
	}

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.HttpServer.Auth.Oidc.ClientID,
		ClientSecret: cfg.HttpServer.Auth.Oidc.ClientSecret,
		Scopes:       scopes,
		RedirectURL:  oauthRedirectUrl,
	}

	return &oidcHandler{
		codec:           codec,
		oauthCfg:        oauthCfg,
		oidcIssuer:      cfg.HttpServer.Auth.Oidc.Issuer,
		multiTenant:     multiTenant,
		accessTokenCfg:  &oidc.Config{ClientID: accessTokenAudience},
		idTokenCfg:      &oidc.Config{ClientID: oauthCfg.ClientID},
		resourceOptions: resourceOptions,
	}, nil
}

func (h *oidcHandler) ensureProvider(ctx context.Context) error {
	if h.provider == nil {
		prov, provErr := oidc.NewProvider(ctx, h.oidcIssuer)
		if provErr != nil {
			return fmt.Errorf("create oidc provider: %w", provErr)
		}
		h.provider = prov
		// Basic only: the library otherwise retries a refused request with the secret in the body, which hides the
		// provider's original error.
		endpoint := h.provider.Endpoint()
		endpoint.AuthStyle = oauth2.AuthStyleInHeader
		h.oauthCfg.Endpoint = endpoint
		h.accessTokenVerifier = h.provider.VerifierContext(ctx, h.accessTokenCfg)
		h.idTokenVerifier = h.provider.VerifierContext(ctx, h.idTokenCfg)
	}
	return nil
}

type AuthFlowState struct {
	State        string `json:"state"`
	Nonce        string `json:"nonce"`
	CodeVerifier string `json:"code_verifier"`
	ReturnTo     string `json:"return_to"`
}

func createRandomValue() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

var authFlowWindow = 10 * time.Minute

func (h *oidcHandler) createAuthRedirect(w http.ResponseWriter, r *http.Request) (string, error) {
	if cfgErr := h.ensureProvider(r.Context()); cfgErr != nil {
		return "", cfgErr
	}

	state := createRandomValue()
	nonce := createRandomValue()

	q := r.URL.Query()
	returnTo := q.Get("return_to")
	if returnTo == "" {
		returnTo = "/"
	}
	if !strings.HasPrefix(returnTo, "/") || strings.HasPrefix(returnTo, "//") {
		return "", fmt.Errorf("invalid return_to")
	}

	verifier := oauth2.GenerateVerifier()
	vs := &AuthFlowState{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: verifier,
		ReturnTo:     returnTo,
	}

	encState, encErr := h.codec.encode(vs)
	if encErr != nil {
		slog.Debug("Failed to encode auth state cookie value", "error", encErr)
		return "", errWriteAuthState
	}
	h.setAuthFlowCookie(w, encState, int(authFlowWindow.Seconds()))

	opts := []oauth2.AuthCodeOption{
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(verifier),
	}
	opts = append(opts, h.resourceOptions...)
	return h.oauthCfg.AuthCodeURL(state, opts...), nil
}

type callbackExchangeResult struct {
	Session  *rez.UserAuthProviderSession
	ReturnTo string
}

func (h *oidcHandler) doCallbackExchange(w http.ResponseWriter, r *http.Request) (*callbackExchangeResult, error) {
	as, stateErr := h.readAndClearAuthFlowCookie(w, r)
	if stateErr != nil {
		return nil, errReadAuthState
	}

	q := r.URL.Query()
	code := q.Get("code")
	if code == "" {
		return nil, fmt.Errorf("missing code")
	}
	if q.Get("state") != as.State {
		return nil, fmt.Errorf("invalid state")
	}

	ctx := r.Context()
	if cfgErr := h.ensureProvider(ctx); cfgErr != nil {
		return nil, cfgErr
	}

	exchangeOpts := append([]oauth2.AuthCodeOption{oauth2.VerifierOption(as.CodeVerifier)}, h.resourceOptions...)
	token, exchangeErr := h.oauthCfg.Exchange(ctx, code, exchangeOpts...)
	if exchangeErr != nil {
		return nil, fmt.Errorf("token exchange failed: %w", exchangeErr)
	}

	vc, vcErr := h.extractVerifiedClaims(ctx, token, as.Nonce)
	if vcErr != nil {
		return nil, fmt.Errorf("token session: %w", vcErr)
	}

	sess, sessErr := vc.makeSession(h.multiTenant)
	if sessErr != nil {
		return nil, sessErr
	}
	return &callbackExchangeResult{Session: sess, ReturnTo: as.ReturnTo}, nil
}

const authFlowCookieName = "rez_auth_flow"

func (h *oidcHandler) setAuthFlowCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     authFlowCookieName,
		Value:    value,
		Path:     h.authStateCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *oidcHandler) readAndClearAuthFlowCookie(w http.ResponseWriter, r *http.Request) (*AuthFlowState, error) {
	var as AuthFlowState
	stateCookie, readCookieErr := r.Cookie(authFlowCookieName)
	if stateCookie == nil || readCookieErr != nil {
		return nil, errReadAuthState
	}
	if decodeErr := h.codec.decode(stateCookie.Value, &as); decodeErr != nil {
		return nil, fmt.Errorf("decode state: %w", decodeErr)
	}
	h.setAuthFlowCookie(w, "", -1)
	return &as, nil
}

func (h *oidcHandler) refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if cfgErr := h.ensureProvider(ctx); cfgErr != nil {
		return nil, cfgErr
	}
	expiredTokenSource := h.oauthCfg.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Second), // set expired to force a refresh request
	})
	return expiredTokenSource.Token()
}

func (h *oidcHandler) doLogout(ctx context.Context) error {
	if cfgErr := h.ensureProvider(ctx); cfgErr != nil {
		return cfgErr
	}
	// TODO: backchannel provider logout
	return nil
}

func (h *oidcHandler) extractVerifiedClaims(ctx context.Context, t *oauth2.Token, nonce string) (*VerifiedProviderAuthClaims, error) {
	if !t.Valid() {
		return nil, fmt.Errorf("invalid token")
	}

	at, atErr := h.accessTokenVerifier.Verify(ctx, t.AccessToken)
	if atErr != nil {
		return nil, fmt.Errorf("verify token: %w", atErr)
	}

	claims := VerifiedProviderAuthClaims{
		ExpiresAt: at.Expiry,
	}

	if atClaimsErr := at.Claims(&claims.AccessClaims); atClaimsErr != nil {
		return nil, fmt.Errorf("parse claims: %w", atClaimsErr)
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

	if claimsErr := id.Claims(&claims.IdentityClaims); claimsErr != nil {
		return nil, errs.ErrAuthSessionInvalid
	}

	return &claims, nil
}

type (
	ProviderAccessTokenClaims struct {
		OrgId   string    `json:"org_id"`
		OrgName string    `json:"org_name"`
		OrgRole user.Role `json:"org_role"`
	}

	ProviderIdentityTokenClaims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	VerifiedProviderAuthClaims struct {
		ExpiresAt      time.Time
		AccessClaims   ProviderAccessTokenClaims
		IdentityClaims ProviderIdentityTokenClaims
	}
)

// makeSession builds the provider session. In multi-tenant mode the access token must name the organization and
// the user's role in it; otherwise the organization comes from configuration and the role is left to sign-in.
func (s *VerifiedProviderAuthClaims) makeSession(multiTenant bool) (*rez.UserAuthProviderSession, error) {
	var role user.Role
	if multiTenant {
		if s.AccessClaims.OrgId == "" {
			return nil, fmt.Errorf("access token has no org_id claim")
		}
		if roleErr := user.RoleValidator(s.AccessClaims.OrgRole); roleErr != nil {
			return nil, fmt.Errorf("access token org_role claim: %w", roleErr)
		}
		role = s.AccessClaims.OrgRole
	}
	return &rez.UserAuthProviderSession{
		User: ent.User{
			AuthProviderID: s.IdentityClaims.Sub,
			Email:          s.IdentityClaims.Email,
			Name:           s.IdentityClaims.Name,
		},
		Org: ent.Organization{
			AuthProviderID: s.AccessClaims.OrgId,
			Name:           s.AccessClaims.OrgName,
		},
		Role:      role,
		ExpiresAt: s.ExpiresAt,
	}, nil
}
