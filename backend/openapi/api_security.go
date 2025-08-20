package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
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
	SecurityMethodSessionCookie    = "session-cookie"
	SecurityMethodApiToken         = "api-token"
	SecurityMethodAuthSessionToken = "session-token"

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
			Type: "apiKey",
			In:   "cookie",
			Name: SessionCookieName,
		},
		SecurityMethodApiToken: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
		SecurityMethodAuthSessionToken: {
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

func GetRequestSessionCookieToken(r *http.Request) (string, error) {
	authCookie, cookieErr := r.Cookie(SessionCookieName)
	if cookieErr != nil {
		if errors.Is(cookieErr, http.ErrNoCookie) {
			return "", nil
		}
		return "", cookieErr
	}
	return authCookie.Value, nil
}

func GetRequestBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return "", nil
	}
	authType := split[0]
	token := split[1]
	if authType != "Bearer" {
		return "", fmt.Errorf("invalid Authorization type %s", authType)
	}
	return token, nil
}

var securityMethodTokenFuncs = map[string]func(r *http.Request) (string, error){
	SecurityMethodSessionCookie:    GetRequestSessionCookieToken,
	SecurityMethodApiToken:         GetRequestBearerToken,
	SecurityMethodAuthSessionToken: GetRequestBearerToken,
}

func getRequestSecurityTokenAndScopes(sec oapiSecurity, r *http.Request) (string, []string) {
	for _, methods := range sec {
		for method, reqScopes := range methods {
			if tokenFn, ok := securityMethodTokenFuncs[method]; ok {
				if token, tokenErr := tokenFn(r); tokenErr == nil && token != "" {
					return token, reqScopes
				}
			}
		}
	}
	return "", nil
}

func MakeSecurityMiddleware(auth rez.AuthService) Middleware {
	return func(c Context, next func(Context)) {
		r, w := humago.Unwrap(c)

		security := c.Operation().Security

		authCtx := c.Context()
		if isExplicitNoSecurity(security) {
			authCtx = access.AnonymousContext(authCtx)
		} else {
			if security == nil {
				security = DefaultSecurity
			}
			token, requiredScopes := getRequestSecurityTokenAndScopes(security, r)

			var authErr error
			authCtx, authErr = auth.CreateVerifiedApiAuthContext(c.Context(), token, requiredScopes)
			if authErr != nil {
				writeStatusError(w, authErr)
				return
			}
		}
		next(huma.WithContext(c, authCtx))
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
