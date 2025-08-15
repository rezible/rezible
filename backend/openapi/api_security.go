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
	SecurityMethodSessionCookie    = "session-cookie"
	SecurityMethodApiToken         = "api-token"
	SecurityMethodAuthSessionToken = "session-token"

	sessionCookieName = "rez_session"
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
			Name: sessionCookieName,
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
	SecurityMethodSessionCookie:    GetRequestSessionCookieToken,
	SecurityMethodApiToken:         GetRequestApiBearerToken,
	SecurityMethodAuthSessionToken: GetRequestApiBearerToken,
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

func MakeSecurityMiddleware(auth rez.AuthSessionService, users rez.UserService) Middleware {
	return func(c Context, next func(Context)) {
		security := c.Operation().Security
		explicitNoAuth := security != nil && len(security) == 0
		if security == nil {
			security = DefaultSecurity
		}

		ctx := c.Context()
		if !explicitNoAuth {
			r, w := humago.Unwrap(c)

			token, requiredScopes := getRequestSecurityTokenAndScopes(security, r)

			sess, verifyErr := auth.VerifyAuthSessionToken(token, nil)
			if verifyErr != nil {
				log.Debug().Err(verifyErr).Msg("failed to verify session token")
				writeStatusError(w, verifyErr)
				return
			}

			for _, scope := range requiredScopes {
				log.Debug().Str("scope", scope).Msg("check scope")
			}

			userCtx, userErr := users.CreateUserContext(ctx, sess.UserId)
			if userErr != nil {
				log.Debug().Err(userErr).Msg("failed to create user auth context")
				writeStatusError(w, userErr)
				return
			}
			ctx = auth.SetAuthSessionContext(userCtx, sess)
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
