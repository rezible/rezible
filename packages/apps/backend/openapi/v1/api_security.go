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
	SecurityMethodAppCookie = "app-cookie"
	SecurityMethodApiToken  = "api-token"

	AppCookieName = "rez_auth_session"
)

var (
	ApiSecurityMethods = SecurityMethods{
		{SecurityMethodAppCookie: {}},
		{SecurityMethodApiToken: {}},
	}
	SecurityMethodCookieOnly = SecurityMethods{
		{SecurityMethodAppCookie: {}},
	}
	ExplicitNoSecurity = SecurityMethods{}
)

func GetDefaultSecuritySchemes() map[string]*SecurityScheme {
	cookieTokenSecurityScheme := &SecurityScheme{
		Name: AppCookieName,
		Type: "openIdConnect",
		In:   "cookie",
	}
	apiTokenSecurityScheme := &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	return map[string]*SecurityScheme{
		SecurityMethodAppCookie: cookieTokenSecurityScheme,
		SecurityMethodApiToken:  apiTokenSecurityScheme,
	}
}

func MakeSecurityMiddleware(api openapi.API, auth rez.AuthService) openapi.Middleware {
	return func(c openapi.Context, next func(openapi.Context)) {
		authCtx, scopes, authCtxErr := extractRequestAuthContext(auth, c)
		if authCtxErr != nil {
			statusErr := convertAuthStatusError(api, c, authCtxErr)
			if respErr := huma.WriteErr(api, c, statusErr.GetStatus(), statusErr.Error()); respErr != nil {
				log.Error().Err(respErr).Msg("failed to write api error response")
			}
			return
		}

		if len(scopes) > 0 {
			log.Debug().Strs("scopes", scopes).Msg("TODO: verify scopes")
		}

		next(huma.WithContext(c, authCtx))
	}
}

func extractRequestAuthContext(auth rez.AuthService, c huma.Context) (context.Context, []string, error) {
	ctx := access.AnonymousContext(c.Context())

	opSecurity := c.Operation().Security
	isExplicitNoSecurity := opSecurity != nil && len(opSecurity) == 0
	if isExplicitNoSecurity {
		return ctx, nil, nil
	}
	if opSecurity == nil {
		opSecurity = ApiSecurityMethods
	}

	r, _ := humago.Unwrap(c)

	log.Debug().
		Str("host", r.Host).
		Str("origin", r.Header.Get("Origin")).
		Msg("get token")

	appCookie := GetRequestAppCookieValue(r)
	canUseAppCookie := appCookie != ""

	apiToken := getRequestApiToken(r)
	canUseApiToken := apiToken != ""

	for _, methodScopes := range opSecurity {
		appCookieScopes, methodAllowsAppCookie := methodScopes[SecurityMethodAppCookie]
		if canUseAppCookie && methodAllowsAppCookie {
			authCtx, authErr := auth.SetAuthSessionContextFromAppCookie(ctx, appCookie)
			return authCtx, appCookieScopes, authErr
		}

		apiTokenScopes, methodAllowsApiToken := methodScopes[SecurityMethodApiToken]
		if canUseApiToken && methodAllowsApiToken {
			authCtx, authErr := auth.SetAuthSessionContextFromApiToken(ctx, apiToken)
			return authCtx, apiTokenScopes, authErr
		}
	}

	return nil, nil, rez.ErrAuthSessionMissing
}

func GetRequestAppCookieValue(r *http.Request) string {
	if cookie, cookieErr := r.Cookie(AppCookieName); cookieErr == nil {
		return cookie.Value
	}
	return ""
}

func getRequestApiToken(r *http.Request) string {
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
