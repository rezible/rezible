package openapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type SecurityScheme = huma.SecurityScheme

type SecurityMethods = []map[string][]string

func IsExplicitNoSecurity(s SecurityMethods) bool {
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
	DefaultSecuritySchemes = map[string]*SecurityScheme{
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
	DefaultSecurity = SecurityMethods{
		{SecurityMethodSessionCookie: {}},
		{SecurityMethodApiToken: {}},
	}
	ExplicitNoSecurity = SecurityMethods{}
)

func MakeSessionCookie(token string, expires time.Time, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
	return cookie
}

func MakeRemoveSessionCookie() *http.Cookie {
	return MakeSessionCookie("", time.Now(), -1)
}

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
