package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/openapi"
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
	SecurityMethodApiTokenOnly = SecurityMethods{
		{SecurityMethodApiToken: {}},
	}
	ExplicitNoSecurity = SecurityMethods{}
)

func GetDefaultSecuritySchemes() map[string]*SecurityScheme {
	appCookieSecurityScheme := &SecurityScheme{
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
		SecurityMethodAppCookie: appCookieSecurityScheme,
		SecurityMethodApiToken:  apiTokenSecurityScheme,
	}
}

func MakeSecurityMiddleware(api openapi.API, auth rez.AuthSessionService) openapi.Middleware {
	return func(c openapi.Context, next func(openapi.Context)) {
		authCtx, authCtxErr := createRequestAuthContext(c, auth)
		if authCtxErr != nil {
			writeAuthStatusError(api, c, authCtxErr)
			return
		}
		next(authCtx)
	}
}

func createRequestAuthContext(c huma.Context, auth rez.AuthSessionService) (huma.Context, error) {
	opSecurity := c.Operation().Security
	isExplicitNoSecurity := opSecurity != nil && len(opSecurity) == 0
	if isExplicitNoSecurity {
		return c, nil
	}
	token, scopes := extractRequestAuth(c, opSecurity)
	if token == "" {
		return nil, rez.ErrAuthSessionMissing
	}

	if len(scopes) > 0 {
		slog.Debug("TODO: verify scopes", "scopes", scopes)
	}

	authCtx, ctxErr := auth.Authenticate(c.Context(), token)
	return huma.WithContext(c, authCtx), ctxErr
}

func extractRequestAuth(c huma.Context, opSec SecurityMethods) (string, []string) {
	r, _ := humago.Unwrap(c)
	if opSec == nil {
		opSec = ApiSecurityMethods
	}

	appCookie := GetRequestAppCookieValue(r)
	apiToken := GetRequestApiTokenValue(r)

	isApiRequest := strings.HasPrefix(r.Host, "api")
	for _, methodScopes := range opSec {
		if isApiRequest {
			scopes, allowed := methodScopes[SecurityMethodApiToken]
			if allowed && apiToken != "" {
				return apiToken, scopes
			}
		} else {
			scopes, allowed := methodScopes[SecurityMethodAppCookie]
			if allowed && appCookie != "" {
				return appCookie, scopes
			}
		}
	}
	return "", nil
}

func GetRequestAppCookieValue(r *http.Request) string {
	if cookie, cookieErr := r.Cookie(AppCookieName); cookieErr == nil {
		return cookie.Value
	}
	return ""
}

func GetRequestApiTokenValue(r *http.Request) string {
	if split := strings.Split(r.Header.Get("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
		return split[1]
	}
	return ""
}

func writeAuthStatusError(api huma.API, c huma.Context, err error) {
	authErr := convertAuthStatusError(err)
	if writeErr := huma.WriteErr(api, c, authErr.GetStatus(), authErr.Error()); writeErr != nil {
		slog.Error("failed to write api error response", "error", writeErr)
	}
}

func convertAuthStatusError(err error) huma.StatusError {
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
	slog.Warn("unknown auth status error", "error", err)
	return ErrAuthSessionInvalid
}
