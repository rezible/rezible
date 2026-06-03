package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	mapset "github.com/deckarep/golang-set/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
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

func MakeSecurityMiddleware() openapi.Middleware {
	api := makeUnhandledApi()
	return func(c openapi.Context, next func(openapi.Context)) {
		if authErr := verifyRequestAuthContext(c); authErr != nil {
			statusErr := ConvertAuthStatusError(authErr)
			if writeErr := huma.WriteErr(api, c, statusErr.GetStatus(), statusErr.Error()); writeErr != nil {
				slog.Error("failed to write api error response", "error", writeErr)
			}
			return
		}
		next(c)
	}
}

func verifyRequestAuthContext(c huma.Context) error {
	opSecurity := c.Operation().Security

	if opSecurity != nil && len(opSecurity) == 0 {
		// explicitly no security required
		return nil
	} else if opSecurity == nil {
		// default security methods
		opSecurity = ApiSecurityMethods
	}

	apiTokenAllowed := false
	apiTokenScopes := mapset.NewSet[string]()

	appCookieAllowed := false
	appCookieScopes := mapset.NewSet[string]()
	for _, methodScopes := range opSecurity {
		if scopes, allowed := methodScopes[SecurityMethodApiToken]; allowed {
			apiTokenAllowed = true
			apiTokenScopes.Append(scopes...)
		}
		if scopes, allowed := methodScopes[SecurityMethodAppCookie]; allowed {
			appCookieAllowed = true
			appCookieScopes.Append(scopes...)
		}
	}

	ctx := c.Context()
	exec := execution.GetContext(ctx)
	if exec.IsAnonymous() {
		return rez.ErrAuthSessionMissing
	}

	var requiredScopes mapset.Set[string]
	if exec.Auth.TokenID != nil { // authed from token
		if !apiTokenAllowed {
			return rez.ErrAuthSessionInvalid
		}
		if exec.Auth.UserID == nil {
			slog.WarnContext(ctx, "request auth token id set, no user set?")
			return rez.ErrAuthSessionInvalid
		}
		if !strings.HasPrefix(c.URL().Host, "api") {
			// api tokens only allowed for api host
			return rez.ErrAuthSessionInvalid
		}
		requiredScopes = apiTokenScopes
	} else if exec.Auth.UserID != nil { // user exists, authed by cookie
		if !appCookieAllowed {
			return rez.ErrAuthSessionInvalid
		}
		requiredScopes = appCookieScopes
	} else { // no auth context
		slog.WarnContext(ctx, "request missing execution auth ctx?")
		return rez.ErrAuthSessionMissing
	}

	if requiredScopes.Cardinality() > 0 {
		slog.DebugContext(ctx, "TODO: verify scopes",
			"scopes", requiredScopes.ToSlice())
	}

	return nil
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
