package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type appAuthSessionCookie struct {
	name string
	path string
}

func newAppAuthSessionCookie(name string, path string) *appAuthSessionCookie {
	return &appAuthSessionCookie{name: name, path: path}
}

func (c *appAuthSessionCookie) Set(w http.ResponseWriter, sess *ent.UserAuthSession) {
	c.set(w, sess.ID.String(), int(time.Until(sess.ExpiresAt).Seconds()))
}

func (c *appAuthSessionCookie) Get(r *http.Request) string {
	if cookie, cookieErr := r.Cookie(c.name); cookieErr == nil {
		return cookie.Value
	}
	return ""
}

func (c *appAuthSessionCookie) Clear(w http.ResponseWriter) {
	c.set(w, "", -1)
}

func (c *appAuthSessionCookie) set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Path:     c.path,
		Value:    value,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) makeApiRequestAuthenticator(authSess rez.AuthSessionService, ac *appAuthSessionCookie) func(http.Handler) http.Handler {
	getRequestAuthSession := func(r *http.Request) (*ent.UserAuthSession, error) {
		if rawSessId := ac.Get(r); rawSessId != "" {
			sessId, idErr := uuid.Parse(rawSessId)
			if idErr != nil {
				return nil, rez.ErrAuthSessionInvalid
			}
			return authSess.Get(r.Context(), sessId)
		}

		if split := strings.Split(r.Header.Get("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
			return authSess.CreateFromToken(r.Context(), split[1])
		}

		return nil, rez.ErrAuthSessionMissing
	}

	validateAuthSession := func(s *ent.UserAuthSession) error {
		if s == nil {
			return rez.ErrAuthSessionMissing
		}
		if s.ExpiresAt.Before(time.Now()) {
			return rez.ErrAuthSessionExpired
		}
		return nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, sessErr := getRequestAuthSession(r)

			if sessErr == nil {
				sessErr = validateAuthSession(sess)
			}

			if sessErr != nil {
				apiErr := oapiv1.ConvertAuthStatusError(sessErr)
				w.WriteHeader(apiErr.GetStatus())
				respErr := json.NewEncoder(w).Encode(apiErr)
				if respErr != nil {
					slog.Warn("failed to write api error response", "error", respErr)
				}
				return
			}

			authReq := r.WithContext(execution.NewUserContext(r.Context(), sess))
			next.ServeHTTP(w, authReq)
		})
	}
}
