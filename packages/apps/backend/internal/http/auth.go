package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type requestAuthenticator struct {
	appCookie *oapiv1.AppCookie
	authSess  rez.AuthSessionService
}

func newRequestAuthenticator(authSess rez.AuthSessionService, appCookie *oapiv1.AppCookie) *requestAuthenticator {
	return &requestAuthenticator{
		appCookie: appCookie,
		authSess:  authSess,
	}
}

func (ra *requestAuthenticator) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authedReq, authErr := ra.authenticateRequest(r)
		if authedReq != nil {
			next.ServeHTTP(w, authedReq)
		}
		if authErr != nil {
			apiErr := oapiv1.ConvertAuthStatusError(authErr)

			w.WriteHeader(apiErr.GetStatus())
			respErr := json.NewEncoder(w).Encode(apiErr)
			if respErr != nil {
				slog.Warn("failed to write api error response", "error", respErr)
			}

		}
	})
}

func (ra *requestAuthenticator) authenticateRequest(r *http.Request) (*http.Request, error) {
	ctx := r.Context()

	if ac := ra.appCookie.Get(r); ac != nil {
		sessId, idErr := uuid.Parse(ac.Value)
		if idErr != nil {
			return nil, rez.ErrAuthSessionInvalid
		}
		sess, sessErr := ra.authSess.LookupSession(ctx, sessId)
		if sessErr != nil {
			return nil, rez.ErrAuthSessionInvalid
		}
		authCtx, cookieAuthErr := ra.authSess.CreateExecutionContext(r.Context(), sess)
		if authCtx != nil && cookieAuthErr == nil {
			return r.WithContext(authCtx), nil
		}
		return nil, cookieAuthErr
	}

	var bearerToken string
	if split := strings.Split(r.Header.Get("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
		bearerToken = split[1]
	}
	if bearerToken != "" {
		// TODO: handle api bearer token
		return nil, rez.ErrAuthSessionInvalid
	}

	return nil, rez.ErrAuthSessionMissing
}
