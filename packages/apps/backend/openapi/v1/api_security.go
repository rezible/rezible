package v1

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	rez "github.com/rezible/rezible"
)

var (
	ErrAuthSessionMissing = huma.Error401Unauthorized("auth_session_missing")
	ErrAuthSessionExpired = huma.Error401Unauthorized("auth_session_expired")
	ErrAuthSessionInvalid = huma.Error401Unauthorized("auth_session_invalid")
	ErrForbidden          = huma.Error403Forbidden("forbidden")
	ErrDomainNotAllowed   = huma.Error403Forbidden("domain_not_allowed")
)

type (
	SecurityScheme        = huma.SecurityScheme
	SecurityMethodOptions = []map[string][]string
)

const (
	SecurityMethodAppCookie          = "app-cookie"
	SecurityMethodApiToken           = "api-token"
	SecurityMethodScopedSessionToken = "session-token"

	AppCookieName = "rez_auth_session"
)

var (
	DefaultSecurityMethods = SecurityMethodOptions{
		{SecurityMethodAppCookie: {}},
		{SecurityMethodApiToken: {}},
	}
)

func MethodSecuritySchemes() map[string]*SecurityScheme {
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
	sessionTokenSecurityScheme := &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "paseto",
	}
	return map[string]*SecurityScheme{
		SecurityMethodAppCookie:          appCookieSecurityScheme,
		SecurityMethodApiToken:           apiTokenSecurityScheme,
		SecurityMethodScopedSessionToken: sessionTokenSecurityScheme,
	}
}

type (
	MethodSecurityCheckFn func(context.Context, SecurityMethodOptions) error
)

func MakeRequestMethodSecurityMiddleware(checkFn MethodSecurityCheckFn) func(c huma.Context, next func(huma.Context)) {
	api := makeUnhandledApi()
	writeAuthError := func(c huma.Context, authErr error) {
		statusErr := ConvertAuthStatusError(authErr)
		if writeErr := huma.WriteErr(api, c, statusErr.GetStatus(), statusErr.Error()); writeErr != nil {
			slog.Error("failed to write api error response", "error", writeErr)
		}
	}
	return func(c huma.Context, next func(huma.Context)) {
		opSecurity := c.Operation().Security

		if opSecurity != nil && len(opSecurity) == 0 {
			next(c)
			return
		}
		if opSecurity == nil {
			opSecurity = DefaultSecurityMethods
		}

		if checkErr := checkFn(c.Context(), opSecurity); checkErr != nil {
			writeAuthError(c, checkErr)
		} else {
			next(c)
		}
	}
}

func ConvertAuthStatusError(err error) huma.StatusError {
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
