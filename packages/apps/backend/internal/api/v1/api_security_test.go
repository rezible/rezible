package apiv1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/userauthsession"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

// storedSessions holds the only sessions LookupSession finds.
type storedSessions struct {
	rez.AuthSessionService
	sessions map[uuid.UUID]*ent.UserAuthSession
}

func (s storedSessions) LookupSession(_ context.Context, id uuid.UUID) (*ent.UserAuthSession, error) {
	if sess, found := s.sessions[id]; found {
		return sess, nil
	}
	return nil, &ent.NotFoundError{}
}

// securedHandler implements no operations; the security middleware rejects each request before one runs.
type securedHandler struct {
	oapi.Handler
	provider *SecurityProvider
}

func (h securedHandler) CreateRequestSecurityContext(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	return h.provider.CreateRequestSecurityContext(w, r)
}

func (h securedHandler) VerifyRequestSecurity(ctx context.Context, opts oapi.OperationSecurityOptions) error {
	return h.provider.VerifyRequestSecurity(ctx, opts)
}

func TestSecurityProviderRejectedSessions(t *testing.T) {
	expired := &ent.UserAuthSession{ID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}
	sessions := storedSessions{sessions: map[uuid.UUID]*ent.UserAuthSession{expired.ID: expired}}
	provider := NewRequestSecurityProvider(sessions, oapi.NewAppAuthSessionCookie("/api"))
	api := oapi.MakeApi(securedHandler{provider: provider})

	for _, tc := range []struct {
		name          string
		cookie        string
		clearsCookies bool
	}{
		{"expired session", expired.ID.String(), true},
		{"unknown session", uuid.NewString(), true},
		{"malformed cookie", "not-a-session", true},
		{"no cookie", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, oapi.VersionPrefix+"/user_session", nil)
			if tc.cookie != "" {
				request.AddCookie(&http.Cookie{Name: oapi.AppAuthSessionCookieName, Value: tc.cookie})
			}
			recorder := httptest.NewRecorder()

			api.Adapter().ServeHTTP(recorder, request)

			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			var body struct {
				Code   string `json:"code"`
				Detail string `json:"detail"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body), recorder.Body.String())
			require.Equal(t, "unauthenticated", body.Code)
			require.Equal(t, "Sign in to continue.", body.Detail)

			cookies := recorder.Result().Cookies()
			if !tc.clearsCookies {
				require.Empty(t, cookies)
				return
			}
			require.Len(t, cookies, 1)
			require.Equal(t, oapi.AppAuthSessionCookieName, cookies[0].Name)
			require.Empty(t, cookies[0].Value)
			require.Negative(t, cookies[0].MaxAge)
		})
	}
}

// scopedHandler answers an admin-only operation and a default one once the security middleware lets them through.
type scopedHandler struct {
	securedHandler
}

func (h scopedHandler) CompleteOrgSetup(context.Context, *oapi.CompleteOrgSetupRequest) (*oapi.CompleteOrgSetupResponse, error) {
	return &oapi.CompleteOrgSetupResponse{}, nil
}

func (h scopedHandler) GetDocumentSession(context.Context, *oapi.GetDocumentSessionRequest) (*oapi.GetDocumentSessionResponse, error) {
	return &oapi.GetDocumentSessionResponse{}, nil
}

func TestSecurityProviderOperationScopes(t *testing.T) {
	tenantID := 1
	session := func(role userauthsession.Role) *ent.UserAuthSession {
		return &ent.UserAuthSession{ID: uuid.New(), TenantID: tenantID, UserID: uuid.New(), Role: role, ExpiresAt: time.Now().Add(time.Hour)}
	}
	admin, member := session(userauthsession.RoleAdmin), session(userauthsession.RoleMember)
	sessions := storedSessions{sessions: map[uuid.UUID]*ent.UserAuthSession{admin.ID: admin, member.ID: member}}
	provider := NewRequestSecurityProvider(sessions, oapi.NewAppAuthSessionCookie("/api"))
	api := oapi.MakeApi(scopedHandler{securedHandler{provider: provider}})

	setupPath := oapi.VersionPrefix + "/organizations/" + uuid.NewString() + "/setup"
	documentSessionPath := oapi.VersionPrefix + "/documents/" + uuid.NewString() + "/session"
	for _, tc := range []struct {
		name    string
		session *ent.UserAuthSession
		method  string
		path    string
		status  int
	}{
		{"member on admin operation", member, http.MethodPost, setupPath, http.StatusForbidden},
		{"admin on admin operation", admin, http.MethodPost, setupPath, http.StatusOK},
		{"member on document session", member, http.MethodGet, documentSessionPath, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"attributes":{"name":"Org"}}`))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(&http.Cookie{Name: oapi.AppAuthSessionCookieName, Value: tc.session.ID.String()})
			recorder := httptest.NewRecorder()

			api.Adapter().ServeHTTP(recorder, request)

			require.Equal(t, tc.status, recorder.Code, recorder.Body.String())
		})
	}
}
