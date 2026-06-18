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

type (
	RequestSecurityMethods struct {
		ApiToken  MethodSecurityOption
		AppCookie MethodSecurityOption
	}
	MethodSecurityOption struct {
		Allowed           bool
		RequiredScopeSets [][]string
	}
	MethodSecurityCheckFn func(context.Context, RequestSecurityMethods) error
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
			// explicitly no security required
			next(c)
			return
		}

		if opSecurity == nil {
			// default security methods
			opSecurity = ApiSecurityMethods
		}

		var sec RequestSecurityMethods
		for _, methodScopes := range opSecurity {
			if scopes, allowed := methodScopes[SecurityMethodApiToken]; allowed {
				sec.ApiToken.Allowed = true
				sec.ApiToken.RequiredScopeSets = append(sec.ApiToken.RequiredScopeSets, scopes)
			}
			if scopes, allowed := methodScopes[SecurityMethodAppCookie]; allowed {
				sec.AppCookie.Allowed = true
				sec.AppCookie.RequiredScopeSets = append(sec.AppCookie.RequiredScopeSets, scopes)
			}
		}

		if checkErr := checkFn(c.Context(), sec); checkErr != nil {
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
