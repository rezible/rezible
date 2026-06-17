package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	mapset "github.com/deckarep/golang-set/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
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

type AppCookie struct {
	path string
}

func NewAppCookie(path string) *AppCookie {
	return &AppCookie{path: path}
}

func (cw *AppCookie) Set(w http.ResponseWriter, sess *ent.UserAuthSession) {
	cw.set(w, sess.ID.String(), int(time.Until(sess.ExpiresAt).Seconds()))
}

func (cw *AppCookie) Get(r *http.Request) *http.Cookie {
	if cookie, cookieErr := r.Cookie(AppCookieName); cookieErr == nil {
		return cookie
	}
	return nil
}

func (cw *AppCookie) Clear(w http.ResponseWriter) {
	cw.set(w, "", -1)
}

func (cw *AppCookie) set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     AppCookieName,
		Path:     cw.path,
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func MakeAuthContextVerifier() func(c huma.Context) bool {
	writeAuthError := makeAuthErrorWriter()

	return func(c huma.Context) bool {
		opSecurity := c.Operation().Security

		if opSecurity != nil && len(opSecurity) == 0 {
			// explicitly no security required
			return true
		}

		if opSecurity == nil {
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
			writeAuthError(c, rez.ErrAuthSessionMissing)
			return false
		}

		var requiredScopes mapset.Set[string]
		if exec.Auth.TokenID != nil { // authed from token
			if !apiTokenAllowed {
				writeAuthError(c, rez.ErrAuthSessionInvalid)
				return false
			}
			if exec.Auth.UserID == nil {
				slog.WarnContext(ctx, "request auth token id set, no user set?")
				writeAuthError(c, rez.ErrAuthSessionInvalid)
				return false
			}
			if !strings.HasPrefix(c.URL().Host, "api") {
				// api tokens only allowed for api host
				writeAuthError(c, rez.ErrAuthSessionInvalid)
				return false
			}
			requiredScopes = apiTokenScopes
		} else if exec.Auth.UserID != nil { // user exists, authed by cookie
			if !appCookieAllowed {
				writeAuthError(c, rez.ErrAuthSessionInvalid)
				return false
			}
			requiredScopes = appCookieScopes
		} else { // no auth context
			slog.WarnContext(ctx, "request missing execution auth ctx?")
			writeAuthError(c, rez.ErrAuthSessionMissing)
			return false
		}

		if requiredScopes.Cardinality() > 0 {
			slog.DebugContext(ctx, "TODO: verify scopes",
				"scopes", requiredScopes.ToSlice())
		}

		return true
	}
}

func makeAuthErrorWriter() func(c huma.Context, authErr error) {
	api := makeUnhandledApi()
	return func(c huma.Context, authErr error) {
		statusErr := ConvertAuthStatusError(authErr)
		if writeErr := huma.WriteErr(api, c, statusErr.GetStatus(), statusErr.Error()); writeErr != nil {
			slog.Error("failed to write api error response", "error", writeErr)
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
