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
)

const (
	orgScope = "organization"
)

type oauthHandler struct {
	cfg         rez.HttpAuthOidcConfig
	redirectUrl string

	authStateCookiePath string
	codec               *cookieCodec

	apiAudience     string
	singleTenantOrg *ent.Organization

	resourceOption      oauth2.AuthCodeOption
	provider            *oidc.Provider
	oauthCfg            *oauth2.Config
	accessTokenVerifier *oidc.IDTokenVerifier
	idTokenVerifier     *oidc.IDTokenVerifier
}

func newOAuthHandler(cfg rez.Config) (*oauthHandler, error) {
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

	apiAudience := cfg.App.ApiDomain
	if apiAudience == "" {
		return nil, fmt.Errorf("no api url configured, can't verify token audience")
	}
	h := &oauthHandler{
		cfg:            cfg.HttpServer.Auth.Oidc,
		redirectUrl:    oauthRedirectUrl,
		codec:          codec,
		apiAudience:    apiAudience,
		resourceOption: oauth2.SetAuthURLParam("resource", apiAudience),
	}
	if cfg.App.SingleTenant.Enabled {
		h.singleTenantOrg = &ent.Organization{
			AuthProviderID: "default",
			Name:           cfg.App.SingleTenant.OrgName,
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
		accessTokenAudience := h.apiAudience
		if h.singleTenantOrg != nil {
			accessTokenAudience = h.cfg.ClientID
		}
		h.oauthCfg = &oauth2.Config{
			ClientID:     h.cfg.ClientID,
			ClientSecret: h.cfg.ClientSecret,
			Scopes:       h.scopes(),
			RedirectURL:  h.redirectUrl,
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

func createRandomValue() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

var authFlowWindow = 10 * time.Minute

func (h *oauthHandler) createAuthRedirect(w http.ResponseWriter, r *http.Request) (string, error) {
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
	h.setAuthStateCookie(w, encState, int(authFlowWindow.Seconds()))

	opts := []oauth2.AuthCodeOption{
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(verifier),
		h.resourceOption,
	}
	return h.oauthCfg.AuthCodeURL(state, opts...), nil
}

func (h *oauthHandler) doCallbackExchange(w http.ResponseWriter, r *http.Request) (*rez.UserAuthProviderSession, string, error) {
	as, stateErr := h.readAndClearAuthStateCookie(w, r)
	if stateErr != nil {
		return nil, "", errReadAuthState
	}

	q := r.URL.Query()
	code := q.Get("code")
	if code == "" {
		return nil, "", fmt.Errorf("missing code")
	}
	if q.Get("state") != as.State {
		return nil, "", fmt.Errorf("invalid state")
	}

	ctx := r.Context()
	if cfgErr := h.ensureProvider(ctx); cfgErr != nil {
		return nil, "", cfgErr
	}
	token, exchangeErr := h.oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(as.CodeVerifier), h.resourceOption)
	if exchangeErr != nil {
		return nil, "", fmt.Errorf("token exchange failed: %w", exchangeErr)
	}
	if !token.Valid() {
		return nil, "", fmt.Errorf("invalid token")
	}

	sess, sessErr := h.extractSessionClaims(ctx, token, as.Nonce)
	if sessErr != nil {
		return nil, "", fmt.Errorf("sess: %w", sessErr)
	}
	return sess, as.ReturnTo, nil
}

const authStateCookieName = "rez_auth_state"

func (h *oauthHandler) setAuthStateCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     authStateCookieName,
		Value:    value,
		Path:     h.authStateCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *oauthHandler) readAndClearAuthStateCookie(w http.ResponseWriter, r *http.Request) (*AuthFlowState, error) {
	var as AuthFlowState
	stateCookie, readCookieErr := r.Cookie(authStateCookieName)
	if stateCookie == nil || readCookieErr != nil {
		return nil, errReadAuthState
	}
	if decodeErr := h.codec.decode(stateCookie.Value, &as); decodeErr != nil {
		return nil, fmt.Errorf("decode state: %w", decodeErr)
	}
	h.setAuthStateCookie(w, "", -1)
	return &as, nil
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

func (h *oauthHandler) extractSessionClaims(ctx context.Context, t *oauth2.Token, nonce string) (*rez.UserAuthProviderSession, error) {
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

	sess := &rez.UserAuthProviderSession{
		ExpiresAt: at.Expiry,
		User: ent.User{
			AuthProviderID: idClaims.Sub,
			Email:          idClaims.Email,
			Name:           idClaims.Name,
		},
		Org: ent.Organization{
			AuthProviderID: atClaims.OrganizationId,
			Name:           atClaims.OrganizationName,
		},
	}

	if h.singleTenantOrg != nil {
		slog.Debug("using single tenant organization")
		sess.Org = *h.singleTenantOrg
	}

	return sess, nil
}
