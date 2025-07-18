package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
)

type oapiSecurity = []map[string][]string

const (
	securityMethodSessionCookie    = "session-cookie"
	securityMethodApiToken         = "api-token"
	securityMethodAuthSessionToken = "session-token"

	sessionCookieName = "rez_session"
)

var (
	DefaultSecuritySchemes = map[string]*huma.SecurityScheme{
		securityMethodSessionCookie: {
			Type: "apiKey",
			In:   "cookie",
			Name: sessionCookieName,
		},
		securityMethodApiToken: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
		securityMethodAuthSessionToken: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	DefaultSecurity = oapiSecurity{
		{securityMethodSessionCookie: {}},
		{securityMethodApiToken: {}},
	}
)

func MakeSessionCookie(r *http.Request, value string, expires time.Time, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Domain:   r.Host,
		Path:     "/",
		Expires:  expires,
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		MaxAge:   maxAge,
		// SameSite: http.SameSiteLaxMode,
	}
	if domain, _, splitErr := net.SplitHostPort(r.Host); splitErr == nil {
		cookie.Domain = domain
	}
	return cookie
}

func GetRequestSessionCookieToken(r *http.Request) (string, error) {
	authCookie, cookieErr := r.Cookie(sessionCookieName)
	if cookieErr != nil {
		if errors.Is(cookieErr, http.ErrNoCookie) {
			return "", nil
		}
		return "", cookieErr
	}
	return authCookie.Value, nil
}

func GetRequestApiBearerToken(r *http.Request) (string, error) {
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
	securityMethodSessionCookie:    GetRequestSessionCookieToken,
	securityMethodApiToken:         GetRequestApiBearerToken,
	securityMethodAuthSessionToken: GetRequestApiBearerToken,
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

func MakeSecurityMiddleware(auth rez.AuthSessionService) Middleware {
	return func(c Context, next func(Context)) {
		security := c.Operation().Security
		explicitNoAuth := security != nil && len(security) == 0
		if explicitNoAuth {
			next(c)
			return
		} else if security == nil {
			security = DefaultSecurity
		}

		r, w := humago.Unwrap(c)
		token, requiredScopes := getRequestSecurityTokenAndScopes(security, r)
		sess, authErr := auth.VerifySessionToken(token)
		if authErr != nil {
			log.Debug().Err(authErr).Msg("failed to verify session token")
			writeAuthSessionError(w, authErr)
			return
		}

		// TODO: check scopes
		for _, scope := range requiredScopes {
			log.Warn().Str("scope", scope).Msg("TODO: verify request security scopes")
		}

		authCtx := auth.CreateSessionContext(r.Context(), sess)
		next(huma.WithContext(c, authCtx))
	}
}

func writeAuthSessionError(w http.ResponseWriter, authErr error) {
	var resp StatusError
	if errors.Is(authErr, rez.ErrNoAuthSession) {
		resp = ErrorUnauthorized("no_session")
	} else if errors.Is(authErr, rez.ErrAuthSessionExpired) {
		resp = ErrorUnauthorized("session_expired")
	} else if errors.Is(authErr, rez.ErrAuthSessionUserMissing) {
		resp = ErrorUnauthorized("missing_user")
	} else {
		resp = ErrorUnauthorized("unknown")
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
