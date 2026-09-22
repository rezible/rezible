package v1

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/openapi"
)

type (
	SecurityScheme = huma.SecurityScheme

	SecurityMethodKind           = string
	SecurityMethodRequiredScopes = []string

	OperationSecurityMethods = map[SecurityMethodKind]SecurityMethodRequiredScopes
	OperationSecurityOptions = []OperationSecurityMethods
)

const (
	SecurityMethodAppCookie          SecurityMethodKind = "app-cookie"
	SecurityMethodApiToken           SecurityMethodKind = "api-token"
	SecurityMethodScopedSessionToken SecurityMethodKind = "scoped-session-token"

	AppAuthSessionCookieName = "rez_auth_session"
)

var DefaultOperationSecurityMethodOptions = OperationSecurityOptions{
	{SecurityMethodAppCookie: {}},
	{SecurityMethodApiToken: {}},
}

func MethodSecuritySchemes() map[string]*SecurityScheme {
	return map[string]*SecurityScheme{
		SecurityMethodAppCookie: {
			Name: AppAuthSessionCookieName,
			Type: "openIdConnect",
			In:   "cookie",
		},
		SecurityMethodApiToken: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
		SecurityMethodScopedSessionToken: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "paseto",
		},
	}
}

type authSessionCookie struct {
	path string
}

func NewAppAuthSessionCookie(path string) rez.AppAuthSessionCookie {
	return &authSessionCookie{path: path}
}

func (c *authSessionCookie) Set(w http.ResponseWriter, sess *ent.UserAuthSession) {
	c.set(w, sess.ID.String(), int(time.Until(sess.ExpiresAt).Seconds()))
}

func (c *authSessionCookie) Get(r *http.Request) (uuid.UUID, error) {
	if cookie, cookieErr := r.Cookie(AppAuthSessionCookieName); cookieErr == nil {
		return uuid.Parse(cookie.Value)
	}
	return uuid.Nil, nil
}

func (c *authSessionCookie) Clear(w http.ResponseWriter) {
	c.set(w, "", -1)
}

func (c *authSessionCookie) set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     AppAuthSessionCookieName,
		Path:     c.path,
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

type SecurityProvider interface {
	CreateRequestSecurityContext(http.ResponseWriter, *http.Request) (context.Context, error)
	VerifyRequestSecurity(context.Context, OperationSecurityOptions) error
}

type contextUnwrapperFunc func(huma.Context) (*http.Request, http.ResponseWriter)

func makeRequestMethodSecurityMiddleware(api openapi.API, p SecurityProvider, unwrapCtx contextUnwrapperFunc) openapi.Middleware {
	defaultSecurityOpts := api.OpenAPI().Security
	getOperationSecurityOptions := func(c huma.Context) OperationSecurityOptions {
		if opts := c.Operation().Security; opts != nil {
			return opts
		}
		return defaultSecurityOpts
	}
	writeSecurityError := func(c huma.Context, err error) {
		statusErr := ConvertStatusError("verify request security", err)
		if writeErr := huma.WriteErr(api, c, statusErr.GetStatus(), statusErr.Error()); writeErr != nil {
			slog.Error("failed to write api error response", "error", writeErr)
		}
	}
	return func(c huma.Context, next func(huma.Context)) {
		r, w := unwrapCtx(c)
		secCtx, sessErr := p.CreateRequestSecurityContext(w, r)
		if sessErr != nil {
			writeSecurityError(c, sessErr)
			return
		}
		c = huma.WithContext(c, secCtx)

		secOpts := getOperationSecurityOptions(c)
		if secOpts != nil && len(secOpts) == 0 {
			slog.Debug("no security options for operation")
		}

		if secErr := p.VerifyRequestSecurity(c.Context(), secOpts); secErr != nil {
			writeSecurityError(c, secErr)
			return
		}

		next(c)
	}
}
