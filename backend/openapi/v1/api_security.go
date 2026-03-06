package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/openapi"
)

type SecurityScheme = huma.SecurityScheme

type SecurityMethods = []map[string][]string

const (
	SecurityMethodSessionCookie = "session-cookie"
	SecurityMethodApiToken      = "api-token"
)

var (
	ErrNoSession      = openapi.ErrorUnauthorized("no_session")
	ErrSessionExpired = openapi.ErrorUnauthorized("session_expired")
	ErrInvalidUser    = openapi.ErrorUnauthorized("invalid_user")
	ErrInvalidTenant  = openapi.ErrorUnauthorized("invalid_tenant")
	ErrUnauthorized   = openapi.ErrorUnauthorized("unauthorized")
	ErrUnknown        = openapi.ErrorUnauthorized("unknown")
)

var (
	DefaultSecurity = SecurityMethods{
		{SecurityMethodSessionCookie: {}},
		{SecurityMethodApiToken: {}},
	}
	ExplicitNoSecurity = SecurityMethods{}
)

func GetDefaultSecuritySchemes() map[string]*SecurityScheme {
	sessionCookieSecurityScheme := &SecurityScheme{
		Name: rez.Config.AuthSessionCookieName(),
		Type: "apiKey",
		In:   "cookie",
	}
	apiTokenSecurityScheme := &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	return map[string]*SecurityScheme{
		SecurityMethodSessionCookie: sessionCookieSecurityScheme,
		SecurityMethodApiToken:      apiTokenSecurityScheme,
	}
}

func isExplicitNoSecurity(s SecurityMethods) bool {
	return s != nil && len(s) == 0
}

func MakeSecurityMiddleware(auth rez.AuthService) openapi.Middleware {
	return func(c openapi.Context, next func(openapi.Context)) {
		r, w := humago.Unwrap(c)
		opSecurity := c.Operation().Security

		ctx := c.Context()
		if !isExplicitNoSecurity(opSecurity) {
			if opSecurity == nil {
				opSecurity = DefaultSecurity
			}
			token, methodScopes := extractRequestTokenAndScopes(r, opSecurity)

			var scopes rez.AuthSessionScopes
			if len(methodScopes) > 0 {
				scopes["api"] = methodScopes
			}

			sess, sessErr := auth.VerifyAuthSessionToken(token, scopes)
			if sessErr != nil {
				writeAuthStatusError(w, sessErr)
				return
			}

			authCtx, authCtxErr := auth.CreateAuthContext(ctx, sess)
			if authCtxErr != nil {
				writeAuthStatusError(w, authCtxErr)
				return
			}

			ctx = authCtx
		} else {
			ctx = access.AnonymousContext(ctx)
		}

		next(huma.WithContext(c, ctx))
	}
}

func getRequestBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if split := strings.Split(header, " "); len(split) == 2 && split[0] == "Bearer" {
		return split[1]
	}
	return ""
}

func getRequestSessionCookieToken(r *http.Request) string {
	authCookie, cookieErr := r.Cookie(rez.Config.AuthSessionCookieName())
	if cookieErr != nil {
		return ""
	}
	return authCookie.Value
}

func extractRequestTokenAndScopes(r *http.Request, opSecurity SecurityMethods) (string, []string) {
	apiToken := getRequestBearerToken(r)
	cookieToken := getRequestSessionCookieToken(r)
	for _, methodScopes := range opSecurity {
		apiTokenScopes, apiTokenAllowed := methodScopes[SecurityMethodApiToken]
		if apiTokenAllowed && apiToken != "" {
			return apiToken, apiTokenScopes
		}
		cookieTokenScopes, cookieTokenAllowed := methodScopes[SecurityMethodSessionCookie]
		if cookieTokenAllowed && cookieToken != "" {
			return cookieToken, cookieTokenScopes
		}
	}
	return "", nil
}

func writeAuthStatusError(w http.ResponseWriter, err error) {
	var resp openapi.StatusError
	if errors.Is(err, rez.ErrNoAuthSession) {
		resp = ErrNoSession
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		resp = ErrSessionExpired
	} else if errors.Is(err, rez.ErrInvalidUser) {
		resp = ErrInvalidUser
	} else if errors.Is(err, rez.ErrInvalidTenant) {
		resp = ErrInvalidTenant
	} else if errors.Is(err, rez.ErrUnauthorized) {
		resp = ErrUnauthorized
	} else {
		resp = ErrUnknown
	}

	respBody, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.GetStatus())
	if _, writeErr := w.Write(respBody); writeErr != nil {
		// TODO: log?
	}
}
