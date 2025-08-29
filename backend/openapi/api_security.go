package openapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
)

type oapiSecurity = []map[string][]string

func isExplicitNoSecurity(s oapiSecurity) bool {
	return s != nil && len(s) == 0
}

const (
	SecurityMethodSessionCookie = "session-cookie"
	SecurityMethodApiToken      = "api-token"

	SessionCookieName = "rezible_auth"
)

var (
	ErrNoSession      = ErrorUnauthorized("no_session")
	ErrSessionExpired = ErrorUnauthorized("session_expired")
	ErrInvalidUser    = ErrorUnauthorized("invalid_user")
	ErrInvalidTenant  = ErrorUnauthorized("invalid_tenant")
	ErrUnauthorized   = ErrorUnauthorized("unauthorized")
	ErrUnknown        = ErrorUnauthorized("unknown")
)

var (
	DefaultSecuritySchemes = map[string]*huma.SecurityScheme{
		SecurityMethodSessionCookie: {
			Name: SessionCookieName,
			Type: "apiKey",
			In:   "cookie",
		},
		SecurityMethodApiToken: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	DefaultSecurity = oapiSecurity{
		{SecurityMethodSessionCookie: {}},
		{SecurityMethodApiToken: {}},
	}
	ExplicitNoSecurity = oapiSecurity{}
)

func GetRequestSessionCookieToken(r *http.Request) string {
	authCookie, cookieErr := r.Cookie(SessionCookieName)
	if cookieErr != nil {
		return ""
	}
	return authCookie.Value
}

func GetRequestBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return ""
	}
	authType := split[0]
	token := split[1]
	if authType != "Bearer" {
		return ""
	}
	return token
}

func extractRequestTokenAndScopes(r *http.Request, opSecurity oapiSecurity) (string, []string) {
	apiToken := GetRequestBearerToken(r)
	cookieToken := GetRequestSessionCookieToken(r)
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

func MakeSecurityMiddleware(auth rez.AuthService) Middleware {
	return func(c Context, next func(Context)) {
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

func writeAuthStatusError(w http.ResponseWriter, err error) {
	var resp StatusError
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
