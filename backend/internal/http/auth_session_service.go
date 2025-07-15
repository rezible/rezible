package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type authContextKey struct{}

var (
	authSessionContextKey = authContextKey{}
)

const (
	defaultSessionDuration = time.Hour
)

// TODO: this should eventually be backed by some kind of storage

type AuthSessionService struct {
	users         rez.UserService
	sessProvider  rez.AuthSessionProvider
	sessionSecret []byte
}

func NewAuthSessionService(ctx context.Context, users rez.UserService, sessProv rez.AuthSessionProvider, sessionSecretKey string) (*AuthSessionService, error) {
	return &AuthSessionService{
		users:         users,
		sessProvider:  sessProv,
		sessionSecret: []byte(sessionSecretKey),
	}, nil
}

func (s *AuthSessionService) ProviderName() string {
	return s.sessProvider.Name()
}

func (s *AuthSessionService) CreateSessionContext(ctx context.Context, sess *rez.AuthSession) context.Context {
	return context.WithValue(ctx, authSessionContextKey, sess)
}

func (s *AuthSessionService) GetSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authSessionContextKey).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

func isRedirectableError(err error) bool {
	return errors.Is(err, rez.ErrAuthSessionExpired) || errors.Is(err, rez.ErrNoAuthSession)
}

func (s *AuthSessionService) MakeFrontendAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/favicon.ico" {
				next.ServeHTTP(w, r)
				return
			}

			sess, sessErr := s.getVerifiedSessionFromRequestCookieToken(r)
			if sessErr == nil {
				next.ServeHTTP(w, r.WithContext(s.CreateSessionContext(r.Context(), sess)))
				return
			}

			if isRedirectableError(sessErr) {
				s.sessProvider.StartAuthFlow(w, r)
				return
			}
			http.Error(w, sessErr.Error(), http.StatusInternalServerError)
		})
	}
}

func (s *AuthSessionService) getMCPAuthSession(r *http.Request) (*rez.AuthSession, error) {
	token, tokenErr := oapi.GetRequestApiBearerToken(r)
	if tokenErr != nil {
		return nil, tokenErr
	} else if token == "" {
		return nil, rez.ErrNoAuthSession
	}
	// TODO: a lot
	fakeSess := &rez.AuthSession{
		ExpiresAt: time.Now().Add(time.Hour),
		UserId:    uuid.New(),
	}
	return fakeSess, nil
}

func (s *AuthSessionService) MakeMCPServerAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, authErr := s.getMCPAuthSession(r)
			if authErr != nil {
				// w.Header().Add("WWW-Authenticate", `Bearer resource_metadata="/.well-known/oauth-protected-resource"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(s.CreateSessionContext(r.Context(), sess)))
		})
	}
}

func (s *AuthSessionService) MakeUserAuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if providerHandled := s.providerAuthFlow(w, r); providerHandled {
			return
		}

		if r.URL.Path == "/auth/logout" {
			s.clearAuthSession(w, r)
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
			return
		}

		_, sessErr := s.getVerifiedSessionFromRequestCookieToken(r)
		if sessErr != nil && isRedirectableError(sessErr) {
			s.sessProvider.StartAuthFlow(w, r)
		} else {
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
		}
	})
}

func (s *AuthSessionService) providerAuthFlow(w http.ResponseWriter, r *http.Request) bool {
	var redirectUrl string
	var createSessionErr error

	ctx := r.Context()
	onUserSessionCreated := func(provUser *ent.User, expiresAt time.Time, redirect string) {
		redirectUrl = redirect
		expiry := expiresAt
		if expiresAt.IsZero() {
			expiry = time.Now().Add(defaultSessionDuration)
		}

		userId, matchIdErr := s.matchUserIdFromProvider(ctx, provUser)
		if matchIdErr != nil {
			createSessionErr = fmt.Errorf("failed to match user id from provider details: %w", matchIdErr)
			return
		}
		if userId == uuid.Nil {
			// TODO: handle this
			log.Debug().Msg("no internal user exists for auth provider supplied details")
		}
		createSessionErr = s.storeAuthSession(w, r, &rez.AuthSession{ExpiresAt: expiry, UserId: userId})
	}

	providerHandled := s.sessProvider.HandleAuthFlowRequest(w, r, onUserSessionCreated)
	if !providerHandled {
		return false
	}

	if createSessionErr != nil {
		log.Error().Err(createSessionErr).Msg("failed to create session")
		http.Error(w, createSessionErr.Error(), http.StatusInternalServerError)
	} else if redirectUrl != "" {
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}

	return true
}

func (s *AuthSessionService) matchUserIdFromProvider(ctx context.Context, provUser *ent.User) (uuid.UUID, error) {
	// TODO: use provider mapping to match user details, not just by email
	email := provUser.Email
	if rez.DebugMode && os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL") != "" {
		email = os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL")
		log.Debug().Str("email", email).Msg("using debug auth email")
	}

	user, lookupErr := s.users.GetByEmail(ctx, email)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return uuid.Nil, nil
		}
		return uuid.Nil, lookupErr
	}
	return user.ID, nil
}

func (s *AuthSessionService) storeAuthSession(w http.ResponseWriter, r *http.Request, sess *rez.AuthSession) error {
	token, tokenErr := s.IssueSessionToken(sess)
	if tokenErr != nil {
		return tokenErr
	}

	http.SetCookie(w, oapi.MakeSessionCookie(r, token, sess.ExpiresAt, 0))
	return nil
}

func (s *AuthSessionService) clearAuthSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, oapi.MakeSessionCookie(r, "", time.Now(), -1))
	s.sessProvider.ClearSession(w, r)
}

func (s *AuthSessionService) getVerifiedSessionFromRequestCookieToken(req *http.Request) (*rez.AuthSession, error) {
	cookieToken, cookieErr := oapi.GetRequestSessionCookieToken(req)
	if cookieErr != nil {
		return nil, fmt.Errorf("error getting token from cookie: %w", cookieErr)
	}
	return s.VerifySessionToken(cookieToken)
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Session rez.AuthSession `json:"session"`
}

func (s *AuthSessionService) IssueSessionToken(session *rez.AuthSession) (string, error) {
	if session == nil {
		return "", errors.New("nil session")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session": *session,
		"exp":     jwt.NewNumericDate(session.ExpiresAt),
	})

	signedToken, signErr := token.SignedString(s.sessionSecret)
	if signErr != nil {
		return "", fmt.Errorf("failed to sign token: %w", signErr)
	}

	return signedToken, nil
}

func (s *AuthSessionService) VerifySessionToken(tokenStr string) (*rez.AuthSession, error) {
	if tokenStr == "" {
		return nil, rez.ErrNoAuthSession
	}

	claims, parseErr := s.parseSessionTokenClaims(tokenStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse token: %w", parseErr)
	}

	if claims.Session.UserId == uuid.Nil {
		return nil, rez.ErrAuthSessionUserMissing
	}

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	return &claims.Session, nil
}

func (s *AuthSessionService) parseSessionTokenClaims(tokenStr string) (*authSessionTokenClaims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.sessionSecret, nil
	}

	token, parseErr := jwt.ParseWithClaims(tokenStr, &authSessionTokenClaims{}, keyFunc)
	if parseErr != nil {
		return nil, fmt.Errorf("parse: %w", parseErr)
	}

	claims, claimsOk := token.Claims.(*authSessionTokenClaims)
	if !claimsOk {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
