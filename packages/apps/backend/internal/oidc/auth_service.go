package oidc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	orgs  rez.OrganizationService
	users rez.UserService

	cfg      Config
	provider *tokenProvider
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (*AuthService, error) {
	cfg, cfgErr := loadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	s := &AuthService{
		orgs:  orgs,
		users: users,
		cfg:   *cfg,
	}

	return s, nil
}

func (s *AuthService) getTokenProvider(ctx context.Context) (*tokenProvider, error) {
	if s.provider == nil {
		prov, provErr := makeTokenProvider(ctx, s.cfg.Oidc)
		if provErr != nil {
			return nil, fmt.Errorf("make token provider: %w", provErr)
		}
		s.provider = prov
	}
	return s.provider, nil
}

func (s *AuthService) GetAuthSessionConfig() rez.AuthSessionConfig {
	return rez.AuthSessionConfig{
		Issuer:          s.cfg.Oidc.Issuer,
		AppClientId:     s.cfg.Oidc.AppClient.Id,
		AppClientScopes: s.cfg.Oidc.AppClient.Scopes,
	}
}

func (s *AuthService) CompleteClientAuthSessionFlow(ctx context.Context, code string, verifier string) ([]http.Cookie, error) {
	prov, provErr := s.getTokenProvider(ctx)
	if provErr != nil {
		return nil, provErr
	}
	token, exchangeErr := prov.exchangeAppClientAuthCode(ctx, code, verifier)
	if exchangeErr != nil {
		return nil, fmt.Errorf("exchange token: %w", exchangeErr)
	}
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		log.Debug().Msg("token does not contain id_token")
		return nil, rez.ErrAuthSessionInvalid
	}
	log.Debug().
		Str("token", token.AccessToken).
		Interface("id", token.Extra("id_token")).
		Msg("exchanged token")
	id, verifyErr := prov.verifyIdToken(ctx, idTokenStr)
	if verifyErr != nil {
		return nil, verifyErr
	}
	usr, usrErr := s.users.SyncFromAuthProvider(ctx, id.getOrganization(), id.getUser())
	if usrErr != nil {
		return nil, fmt.Errorf("match auth user: %w", usrErr)
	}
	authCtx := s.makeAuthSessionContext(ctx, usr, id)
	return oapiv1.MakeAuthSessionCookies(authCtx, *token), nil
}

func (s *AuthService) RefreshClientAuthSession(ctx context.Context, refreshToken string) ([]http.Cookie, error) {
	if refreshToken == "" {
		return nil, rez.ErrAuthSessionMissing
	}
	prov, provErr := s.getTokenProvider(ctx)
	if provErr != nil {
		return nil, provErr
	}
	freshToken, refreshErr := prov.refreshToken(ctx, refreshToken)
	if refreshErr != nil {
		return nil, fmt.Errorf("fetch refreshed token: %w", refreshErr)
	}
	return oapiv1.MakeAuthSessionCookies(ctx, *freshToken), nil
}

func (s *AuthService) ClearClientAuthSession() ([]http.Cookie, error) {
	return oapiv1.MakeLogoutAuthSessionCookies(), nil
}

type authUserSessionContextKey struct{}

func (s *AuthService) CreateAuthSessionContext(ctx context.Context, tokenStr string) (context.Context, error) {
	prov, provErr := s.getTokenProvider(ctx)
	if provErr != nil {
		return nil, provErr
	}
	token, verifyErr := prov.verifyIdToken(ctx, tokenStr)
	if verifyErr != nil {
		log.Debug().Err(verifyErr).Msg("verify auth token")
		return nil, verifyErr
	}
	usr, usrErr := s.users.Get(access.SystemContext(ctx), user.AuthProviderID(token.getUser().AuthProviderID))
	if usrErr != nil {
		log.Debug().Err(usrErr).Msg("get user")
		return nil, rez.ErrAuthSessionInvalid
	}
	return s.makeAuthSessionContext(ctx, usr, token), nil
}

func (s *AuthService) makeAuthSessionContext(ctx context.Context, u *ent.User, token *verifiedSessionToken) context.Context {
	return context.WithValue(access.WithUser(ctx, u), authUserSessionContextKey{}, newAuthSession(u, token))
}

func (s *AuthService) GetAuthSession(ctx context.Context) rez.AuthSession {
	if sess, ok := ctx.Value(authUserSessionContextKey{}).(AuthSession); ok {
		return sess
	}
	return AuthSession{}
}

type AuthSession struct {
	userId    uuid.UUID
	scopes    []string
	expiresAt time.Time
}

func newAuthSession(u *ent.User, t *verifiedSessionToken) AuthSession {
	return AuthSession{
		userId:    u.ID,
		scopes:    t.claims.Scopes,
		expiresAt: t.id.Expiry,
	}
}

func (a AuthSession) UserId() uuid.UUID {
	return a.userId
}

func (a AuthSession) Scopes() []string {
	return a.scopes
}

func (a AuthSession) ExpiresAt() time.Time {
	return a.expiresAt
}
