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

	SessionCookieName = "rez_session"
)

var (
	ErrNoSession      = ErrorUnauthorized("no_session")
	ErrSessionExpired = ErrorUnauthorized("session_expired")
	ErrMissingUser    = ErrorUnauthorized("missing_user")
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

		ctx := access.AnonymousContext(c.Context())
		if !isExplicitNoSecurity(opSecurity) {
			if opSecurity == nil {
				opSecurity = DefaultSecurity
			}

			var authErr error
			token, scopes := extractRequestTokenAndScopes(r, opSecurity)
			if token == "" {
				authErr = rez.ErrNoAuthSession
			} else {
				ctx, authErr = auth.CreateVerifiedApiAuthContext(ctx, token, scopes)
			}
			if authErr != nil {
				writeStatusError(w, authErr)
				return
			}
		}
		next(huma.WithContext(c, ctx))
	}
}

func writeStatusError(w http.ResponseWriter, err error) {
	var resp StatusError
	if errors.Is(err, rez.ErrNoAuthSession) {
		resp = ErrNoSession
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		resp = ErrSessionExpired
	} else if errors.Is(err, rez.ErrAuthSessionUserMissing) {
		resp = ErrMissingUser
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
