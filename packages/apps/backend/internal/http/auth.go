package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
)

type appAuthSessionCookie struct {
	name string
	path string
}

func newAppAuthSessionCookie(path string) *appAuthSessionCookie {
	return &appAuthSessionCookie{name: oapiv1.AppCookieName, path: path}
}

func (c *appAuthSessionCookie) Set(w http.ResponseWriter, sess *ent.UserAuthSession) {
	c.set(w, sess.ID.String(), int(time.Until(sess.ExpiresAt).Seconds()))
}

func (c *appAuthSessionCookie) Get(r *http.Request) (uuid.UUID, error) {
	if cookie, cookieErr := r.Cookie(c.name); cookieErr == nil {
		return uuid.Parse(cookie.Value)
	}
	return uuid.Nil, nil
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

type requestAuthValidator struct {
	devSessionOverride bool
	sessions           rez.AuthSessionService
	cookie             *appAuthSessionCookie
}

func newRequestAuthValidator(sess rez.AuthSessionService, asc *appAuthSessionCookie) *requestAuthValidator {
	return &requestAuthValidator{sessions: sess, cookie: asc}
}

func (v *requestAuthValidator) AuthSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessCtx, sessErr := v.createAuthSessionContext(w, r)
		if sessErr != nil {
			apiErr := oapiv1.ConvertAuthStatusError(sessErr)
			w.WriteHeader(apiErr.GetStatus())
			respErr := json.NewEncoder(w).Encode(apiErr)
			if respErr != nil {
				slog.Warn("failed to write api error response", "error", respErr)
			}
			return
		}
		next.ServeHTTP(w, r.WithContext(sessCtx))
	})
}

func (v *requestAuthValidator) createAuthSessionContext(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	sess, sessErr := v.extractRequestSession(w, r)
	if sessErr != nil {
		return nil, sessErr
	} else if sess == nil {
		return nil, rez.ErrAuthSessionMissing
	}

	if sess.ExpiresAt.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	return execution.NewUserContext(r.Context(), sess), nil
}

func (v *requestAuthValidator) extractRequestSession(w http.ResponseWriter, r *http.Request) (*ent.UserAuthSession, error) {
	cookieId, cookieErr := v.cookie.Get(r)
	if cookieId == uuid.Nil && v.devSessionOverride {
		return v.setDevSessionOverride(w, r)
	}
	if cookieErr != nil {
		return nil, rez.ErrAuthSessionInvalid
	}
	if cookieId != uuid.Nil {
		return v.sessions.LookupSession(r.Context(), cookieId)
	}

	var apiToken string
	if split := strings.Split(r.Header.Get("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
		apiToken = split[1]
	}
	if apiToken != "" {
		return v.sessions.CreateForToken(r.Context(), apiToken)
	}

	return nil, rez.ErrAuthSessionMissing
}

func (v *requestAuthValidator) setDevSessionOverride(w http.ResponseWriter, r *http.Request) (*ent.UserAuthSession, error) {
	devSess := &rez.UserAuthProviderSession{
		User: ent.User{
			Email:          "test@dev.rezible.com",
			Name:           "Dev User",
			AuthProviderID: "dev-user",
		},
		Org: ent.Organization{
			Name:           "Dev Org",
			AuthProviderID: "dev-org",
		},
		ExpiresAt: time.Now().Add(time.Hour),
	}
	slog.Warn("Authenticating with development override")
	sess, sessErr := v.sessions.CreateFromUserAuthResponse(r.Context(), devSess)
	if sessErr != nil {
		return nil, sessErr
	}
	v.cookie.Set(w, sess)
	return sess, nil
}
