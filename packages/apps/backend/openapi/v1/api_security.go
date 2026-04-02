package v1

import (
	"context"
	"errors"
	"fmt"
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
	ErrAuthSessionMissing = huma.Error401Unauthorized("auth_session_missing")
	ErrAuthSessionExpired = huma.Error401Unauthorized("auth_session_expired")
	ErrAuthSessionInvalid = huma.Error401Unauthorized("auth_session_invalid")
	ErrForbidden          = huma.Error403Forbidden("forbidden")
	ErrDomainNotAllowed   = huma.Error403Forbidden("domain_not_allowed")
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
			statusErr := convertAuthStatusError(api, c, authCtxErr)
			if respErr := huma.WriteErr(api, c, statusErr.GetStatus(), statusErr.Error()); respErr != nil {
				log.Error().Err(respErr).Msg("failed to write api error response")
			}
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

	token, methodScopes := extractRequestTokenAndMethodScopes(c, opSecurity)
	authCtx, authCtxErr := auth.CreateAuthSessionContext(ctx, token)
	if authCtxErr != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", authCtxErr)
	}
	if len(methodScopes) > 0 {
		log.Debug().Strs("scopes", methodScopes).Msg("TODO: verify scopes")
	}
	return huma.WithContext(c, authCtx), nil
}

func extractRequestTokenAndMethodScopes(c huma.Context, opSecurity SecurityMethods) (string, []string) {
	r, _ := humago.Unwrap(c)
	cookieToken := GetRequestAuthCookieToken(r)
	bearerToken := GetRequestAuthBearerToken(r)

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

func GetRequestAuthCookieToken(r *http.Request) string {
	if cookie, cookieErr := r.Cookie(accessTokenCookieName); cookieErr == nil {
		return cookie.Value
	}
	return ""
}

func GetRequestAuthBearerToken(r *http.Request) string {
	if split := strings.Split(r.Header.Get("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
		return split[1]
	}
	return ""
}

func convertAuthStatusError(api huma.API, c huma.Context, err error) huma.StatusError {
	if errors.Is(err, rez.ErrAuthSessionMissing) {
		return ErrAuthSessionMissing
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		return ErrAuthSessionExpired
	} else if errors.Is(err, rez.ErrAuthSessionInvalid) {
		return ErrAuthSessionInvalid
	} else if errors.Is(err, rez.ErrInvalidUser) {
		return ErrAuthSessionInvalid
	} else if errors.Is(err, rez.ErrInvalidTenant) {
		return ErrAuthSessionInvalid
	} else if errors.Is(err, rez.ErrDomainNotAllowed) {
		return ErrDomainNotAllowed
	}
	log.Warn().Err(err).Msg("unknown auth status error")
	return ErrAuthSessionInvalid
}
