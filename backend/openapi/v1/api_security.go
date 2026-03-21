package v1

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/openapi"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

var (
	ErrNoSession      = huma.Error401Unauthorized("no_session")
	ErrSessionExpired = huma.Error401Unauthorized("session_expired")
	ErrMissingScopes  = huma.Error401Unauthorized("missing_scopes")
	ErrInvalidUser    = huma.Error401Unauthorized("invalid_user")
	ErrInvalidTenant  = huma.Error401Unauthorized("invalid_tenant")
	ErrUnknown        = huma.Error401Unauthorized("unknown")
)

type SecurityScheme = huma.SecurityScheme

type SecurityMethods = []map[string][]string

const (
	SecurityMethodCookieToken = "cookie-token"
	SecurityMethodBearerToken = "bearer-token"

	accessTokenCookieName  = "rez_access_token"
	refreshTokenCookieName = "rez_refresh_token"
)

type RequestWithRefreshTokenCookie struct {
	Cookie http.Cookie `cookie:"rez_refresh_token"`
}

var (
	ApiSecurityMethods = SecurityMethods{
		{SecurityMethodCookieToken: {}},
		{SecurityMethodBearerToken: {}},
	}
	SecurityMethodCookieOnly = SecurityMethods{
		{SecurityMethodCookieToken: {}},
	}
	ExplicitNoSecurity = SecurityMethods{}
)

func GetDefaultSecuritySchemes() map[string]*SecurityScheme {
	cookieTokenSecurityScheme := &SecurityScheme{
		Name: accessTokenCookieName,
		Type: "apiKey",
		In:   "cookie",
	}
	apiTokenSecurityScheme := &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	return map[string]*SecurityScheme{
		SecurityMethodCookieToken: cookieTokenSecurityScheme,
		SecurityMethodBearerToken: apiTokenSecurityScheme,
	}
}

func MakeAuthSessionCookies(ctx context.Context, token oauth2.Token) []http.Cookie {
	return []http.Cookie{
		{
			Name:     accessTokenCookieName,
			Value:    token.AccessToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   int(token.ExpiresIn),
		},
		{
			Name:     refreshTokenCookieName,
			Value:    token.RefreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     RefreshAuthSession.Path,
			MaxAge:   60 * 60 * 24 * 30,
		},
	}
}

func MakeLogoutAuthSessionCookies() []http.Cookie {
	return []http.Cookie{
		{
			Name:     accessTokenCookieName,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   -1,
		},
		{
			Name:     refreshTokenCookieName,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Path:     RefreshAuthSession.Path,
			MaxAge:   -1,
		},
	}
}

func MakeSecurityMiddleware(api openapi.API, auth rez.AuthService) openapi.Middleware {
	return func(c openapi.Context, next func(openapi.Context)) {
		authCtx, authCtxErr := CreateAuthContext(c, auth)
		if authCtxErr != nil {
			writeAuthStatusError(api, c, authCtxErr)
			return
		}
		next(authCtx)
	}
}

func CreateAuthContext(c huma.Context, auth rez.AuthService) (huma.Context, error) {
	ctx := access.AnonymousContext(c.Context())

	opSecurity := c.Operation().Security
	isExplicitNoSecurity := opSecurity != nil && len(opSecurity) == 0
	if isExplicitNoSecurity {
		return huma.WithContext(c, ctx), nil
	}

	if opSecurity == nil {
		opSecurity = ApiSecurityMethods
	}

	cookieToken, bearerToken := getRequestTokens(c)
	token, methodScopes := extractRequestTokenAndMethodScopes(opSecurity, cookieToken, bearerToken)
	authCtx, authCtxErr := auth.CreateAuthSessionContext(ctx, token)
	if authCtxErr != nil {
		return nil, authCtxErr
	}
	if len(methodScopes) > 0 {
		log.Debug().Strs("scopes", methodScopes).Msg("TODO: verify scopes")
	}
	return huma.WithContext(c, authCtx), nil
}

func getRequestTokens(c huma.Context) (cookieToken string, bearerToken string) {
	if cookie, cookieErr := huma.ReadCookie(c, accessTokenCookieName); cookieErr == nil {
		cookieToken = cookie.Value
	}
	if split := strings.Split(c.Header("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
		bearerToken = split[1]
	}
	return cookieToken, bearerToken
}

func extractRequestTokenAndMethodScopes(opSecurity SecurityMethods, cookieToken, bearerToken string) (string, []string) {
	for _, methodScopes := range opSecurity {
		cookieTokenScopes, cookieTokenAllowed := methodScopes[SecurityMethodCookieToken]
		if cookieToken != "" && cookieTokenAllowed {
			return cookieToken, cookieTokenScopes
		}
		bearerTokenScopes, bearerTokenAllowed := methodScopes[SecurityMethodBearerToken]
		if bearerToken != "" && bearerTokenAllowed {
			return bearerToken, bearerTokenScopes
		}
	}
	return "", nil
}

func writeAuthStatusError(api huma.API, c huma.Context, err error) {
	var resp huma.StatusError
	if errors.Is(err, rez.ErrAuthSessionMissing) {
		resp = ErrNoSession
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		resp = ErrSessionExpired
	} else if errors.Is(err, rez.ErrAuthSessionInvalid) {
		resp = ErrInvalidUser
	} else if errors.Is(err, rez.ErrInvalidUser) {
		resp = ErrInvalidUser
	} else if errors.Is(err, rez.ErrInvalidTenant) {
		resp = ErrInvalidTenant
	} else {
		resp = ErrUnknown
	}

	_, w := humago.Unwrap(c)
	for _, cookie := range MakeLogoutAuthSessionCookies() {
		http.SetCookie(w, &cookie)
	}

	if respErr := huma.WriteErr(api, c, resp.GetStatus(), resp.Error()); respErr != nil {
		log.Error().Err(respErr).Msg("failed to write api error response")
	}
}
