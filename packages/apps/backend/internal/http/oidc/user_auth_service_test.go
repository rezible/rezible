package oidc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

// deletedSessions records the sessions DeleteSession is asked to delete, none of which exist.
type deletedSessions struct {
	rez.AuthSessionService
	deleted []uuid.UUID
}

func (s *deletedSessions) DeleteSession(_ context.Context, id uuid.UUID) error {
	s.deleted = append(s.deleted, id)
	return &ent.NotFoundError{}
}

func TestLogoutDeletesSession(t *testing.T) {
	sessionID := uuid.New()
	for _, tc := range []struct {
		name    string
		cookie  string
		deleted []uuid.UUID
	}{
		{"session cookie", sessionID.String(), []uuid.UUID{sessionID}},
		{"malformed cookie", "not-a-session", nil},
		{"no cookie", "", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sessions := &deletedSessions{}
			provider := &UserAuthProvider{
				sess:   sessions,
				cookie: oapi.NewAppAuthSessionCookie("/api"),
				oidc:   &oidcHandler{},
			}
			request := httptest.NewRequest(http.MethodGet, "/logout", nil)
			if tc.cookie != "" {
				request.AddCookie(&http.Cookie{Name: oapi.AppAuthSessionCookieName, Value: tc.cookie})
			}
			recorder := httptest.NewRecorder()

			provider.MakeAuthHandler().ServeHTTP(recorder, request)

			require.Equal(t, tc.deleted, sessions.deleted)
			require.Equal(t, http.StatusFound, recorder.Code)
			require.Equal(t, "/login", recorder.Header().Get("Location"))
			cookies := recorder.Result().Cookies()
			require.Len(t, cookies, 1)
			require.Equal(t, oapi.AppAuthSessionCookieName, cookies[0].Name)
			require.Negative(t, cookies[0].MaxAge)
		})
	}
}
