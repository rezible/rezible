package apiv1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type SecurityProvider struct {
	sessions rez.AuthSessionService
	cookie   rez.AppAuthSessionCookie
}

func NewRequestSecurityProvider(sess rez.AuthSessionService, cookie rez.AppAuthSessionCookie) *SecurityProvider {
	return &SecurityProvider{sessions: sess, cookie: cookie}
}

func (v *SecurityProvider) CreateRequestSecurityContext(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	sess, sessErr := v.extractRequestSession(r)
	if sessErr != nil {
		return nil, sessErr
	}
	return v.makeVerifiedRequestContext(r.Context(), sess)
}

func (v *SecurityProvider) makeVerifiedRequestContext(ctx context.Context, sess *ent.UserAuthSession) (context.Context, error) {
	if verifyErr := v.verifySession(sess); verifyErr != nil {
		return nil, verifyErr
	}
	return execution.NewUserContext(ctx, sess), nil
}

func (v *SecurityProvider) verifySession(sess *ent.UserAuthSession) error {
	if sess == nil {
		return rez.ErrAuthSessionMissing
	}
	if sess.ExpiresAt.Before(time.Now()) {
		return rez.ErrAuthSessionExpired
	}
	return nil
}

func (v *SecurityProvider) extractRequestSession(r *http.Request) (*ent.UserAuthSession, error) {
	cookieId, cookieErr := v.cookie.Get(r)
	if cookieErr != nil {
		return nil, rez.ErrAuthSessionInvalid
	}
	if cookieId != uuid.Nil {
		sess, lookupErr := v.sessions.LookupSession(r.Context(), cookieId)
		if lookupErr != nil {
			if ent.IsNotFound(lookupErr) {
				return nil, rez.ErrAuthSessionInvalid
			}
			return nil, fmt.Errorf("lookup session: %w", lookupErr)
		}
		return sess, nil
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

func (v *SecurityProvider) VerifyRequestSecurity(ctx context.Context, opts oapi.OperationSecurityOptions) error {
	ec := execution.GetContext(ctx)

	if ec.IsAnonymous() {
		return rez.ErrAuthSessionMissing
	}

	if len(ec.Auth.Scopes) > 0 {
		var scopesSatisfied bool
		for _, methodOpts := range opts {
			for method, methodScopes := range methodOpts {
				if v.scopesSatisfied(ec.Auth.Scopes, methodScopes) {
					scopesSatisfied = true
					slog.Debug("method scopes satisfied", "method", method)
					break
				}
			}
		}
		if !scopesSatisfied {
			return rez.ErrAuthSessionInvalid
		}
	}

	return nil
}

func (v *SecurityProvider) scopesSatisfied(requiredScopes []string, authScopes []string) bool {
	// TODO
	return true
}

type DevelopmentSecurityProvider struct {
	sp       *SecurityProvider
	identity *rez.UserAuthProviderSession
}

func NewDevelopmentSecurityProvider(sp *SecurityProvider, identity *rez.UserAuthProviderSession) (*DevelopmentSecurityProvider, error) {
	return &DevelopmentSecurityProvider{sp: sp, identity: identity}, nil
}

func (p *DevelopmentSecurityProvider) CreateRequestSecurityContext(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	slog.Warn("Authenticating with development override")
	sess, sessErr := p.sp.extractRequestSession(r)
	if sessErr != nil && errors.Is(sessErr, rez.ErrAuthSessionMissing) {
		sess, sessErr = p.sp.sessions.CreateFromUserAuthResponse(r.Context(), p.identity)
		if sessErr != nil {
			return nil, sessErr
		}
		p.sp.cookie.Set(w, sess)
	}
	return p.sp.makeVerifiedRequestContext(r.Context(), sess)
}

func (p *DevelopmentSecurityProvider) VerifyRequestSecurity(ctx context.Context, opts oapi.OperationSecurityOptions) error {
	return p.sp.VerifyRequestSecurity(ctx, opts)
}
