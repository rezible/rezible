package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/privacy"
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

const (
	defaultSessionDuration = time.Hour
)

type AuthSessionService struct {
	users         rez.UserService
	sessProvider  rez.AuthSessionProvider
	sessionSecret []byte
}

var _ rez.AuthSessionService = (*AuthSessionService)(nil)

func NewAuthSessionService(users rez.UserService, sessProv rez.AuthSessionProvider, sessionSecretKey string) (*AuthSessionService, error) {
	return &AuthSessionService{
		users:         users,
		sessProvider:  sessProv,
		sessionSecret: []byte(sessionSecretKey),
	}, nil
}

func (s *AuthSessionService) ProviderName() string {
	return s.sessProvider.Name()
}

type authUserSessionContextKey struct{}

func newUserAuthSession(userId uuid.UUID, expiresAt time.Time) *rez.UserAuthSession {
	return &rez.UserAuthSession{UserId: userId, ExpiresAt: expiresAt}
}

func (s *AuthSessionService) CreateUserAuthContext(ctx context.Context, sess *rez.UserAuthSession) (context.Context, error) {
	user, userErr := s.users.GetById(privacy.DecisionContext(ctx, privacy.Allow), sess.UserId)
	if userErr != nil {
		return ctx, fmt.Errorf("get user by id: %w", userErr)
	}
	accessCtx := access.TenantContext(ctx, access.RoleUser, user.TenantID)

	authCtx := context.WithValue(accessCtx, authUserSessionContextKey{}, sess)
	return authCtx, nil
}

func (s *AuthSessionService) GetUserAuthSession(ctx context.Context) (*rez.UserAuthSession, error) {
	sess, ok := ctx.Value(authUserSessionContextKey{}).(*rez.UserAuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

func isRedirectableError(err error) bool {
	return errors.Is(err, rez.ErrAuthSessionExpired) || errors.Is(err, rez.ErrNoAuthSession)
}

func (s *AuthSessionService) wrapUserAuthRequest(r *http.Request, sess *rez.UserAuthSession) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), authUserSessionContextKey{}, sess))
}

func (s *AuthSessionService) FrontendMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/favicon.ico" {
				next.ServeHTTP(w, r)
				return
			}

			_, sessErr := s.getVerifiedUserAuthSession(r)
			if sessErr != nil {
				if isRedirectableError(sessErr) {
					s.sessProvider.StartAuthFlow(w, r)
				} else {
					http.Error(w, sessErr.Error(), http.StatusInternalServerError)
				}
				return
			}
			next.ServeHTTP(w, r) //s.wrapUserAuthRequest(r, sess))
		})
	}
}

func (s *AuthSessionService) MCPServerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, authErr := s.getMCPUserSession(r)
			if authErr != nil {
				// w.Header().Add("WWW-Authenticate", `Bearer resource_metadata="/.well-known/oauth-protected-resource"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, s.wrapUserAuthRequest(r, sess))
		})
	}
}

func (s *AuthSessionService) AuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/logout" {
			s.handleLogoutRequest(w, r)
			return
		}

		if providerHandled := s.providerAuthFlow(w, r); providerHandled {
			return
		}

		_, sessErr := s.getVerifiedUserAuthSession(r)
		if sessErr != nil && isRedirectableError(sessErr) {
			s.sessProvider.StartAuthFlow(w, r)
		} else {
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
		}
	})
}

func (s *AuthSessionService) handleLogoutRequest(w http.ResponseWriter, r *http.Request) {
	s.sessProvider.ClearSession(w, r)
	http.SetCookie(w, oapi.MakeSessionCookie(r, "", time.Now(), -1))
	http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
}

func (s *AuthSessionService) providerAuthFlow(w http.ResponseWriter, r *http.Request) bool {
	var redirectUrl string
	var createSessionErr error

	ctx := r.Context()
	onUserSessionCreated := func(provUser *ent.User, expiresAt time.Time, redirect string) {
		redirectUrl = redirect
		if expiresAt.IsZero() {
			expiresAt = time.Now().Add(defaultSessionDuration)
		}

		userId, matchIdErr := s.lookupProviderUser(ctx, provUser)
		if matchIdErr != nil {
			createSessionErr = fmt.Errorf("failed to match user id from provider details: %w", matchIdErr)
			return
		}
		if userId == uuid.Nil {
			// TODO: handle this
			log.Debug().Msg("no internal user exists for auth provider supplied details")
		}
		token, tokenErr := s.IssueUserAuthSessionToken(newUserAuthSession(userId, expiresAt))
		if tokenErr != nil {
			createSessionErr = fmt.Errorf("failed to issue user session token: %w", tokenErr)
		} else {
			http.SetCookie(w, oapi.MakeSessionCookie(r, token, expiresAt, 0))
		}
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

func (s *AuthSessionService) lookupProviderUser(ctx context.Context, provUser *ent.User) (uuid.UUID, error) {
	// TODO: use provider mapping to match user details, not just by email
	email := provUser.Email
	if rez.DebugMode && os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL") != "" {
		email = os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL")
		log.Debug().Str("email", email).Msg("using debug auth email")
	}

	allowQueryCtx := privacy.DecisionContext(ctx, privacy.Allow)
	user, lookupErr := s.users.GetByEmail(allowQueryCtx, email)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("users.GetByEmail: %w", lookupErr)
	}
	return user.ID, nil
}

func (s *AuthSessionService) getMCPUserSession(r *http.Request) (*rez.UserAuthSession, error) {
	bearerToken, tokenErr := oapi.GetRequestApiBearerToken(r)
	if tokenErr != nil {
		return nil, tokenErr
	}
	// TODO: a lot
	// return s.VerifyUserSessionToken(bearerToken)
	log.Debug().Str("bearer", bearerToken).Msg("skipping mcp auth verification")
	return newUserAuthSession(uuid.New(), time.Now().Add(time.Hour)), nil
}

func (s *AuthSessionService) getVerifiedUserAuthSession(r *http.Request) (*rez.UserAuthSession, error) {
	cookieToken, cookieErr := oapi.GetRequestSessionCookieToken(r)
	if cookieErr != nil {
		return nil, fmt.Errorf("getting token from cookie: %w", cookieErr)
	}
	return s.VerifyUserAuthSessionToken(cookieToken)
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	UserId uuid.UUID `json:"userId"`
}

func (s *AuthSessionService) IssueUserAuthSessionToken(sess *rez.UserAuthSession) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": sess.UserId,
		"exp":    jwt.NewNumericDate(sess.ExpiresAt),
	})

	signedToken, signErr := token.SignedString(s.sessionSecret)
	if signErr != nil {
		return "", fmt.Errorf("failed to sign token: %w", signErr)
	}

	return signedToken, nil
}

func (s *AuthSessionService) VerifyUserAuthSessionToken(tokenStr string) (*rez.UserAuthSession, error) {
	if tokenStr == "" {
		return nil, rez.ErrNoAuthSession
	}

	claims, parseErr := s.parseSessionTokenClaims(tokenStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse token: %w", parseErr)
	}

	if claims.UserId == uuid.Nil {
		return nil, rez.ErrAuthSessionUserMissing
	}

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	return newUserAuthSession(claims.UserId, exp.Time), nil
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
