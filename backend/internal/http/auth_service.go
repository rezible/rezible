package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	authSessionContextKey = "user_auth_session"
	authSessionCookieName = "rez_auth"
	authMountPath         = "/auth"

	errInvalidAuthToken = errors.New("invalid bearer token")
)

// AutoCreateUsers TODO: don't auto-create users
const AutoCreateUsers = true

type AuthService struct {
	users         rez.UserService
	sessProvider  rez.AuthSessionProvider
	sessionSecret []byte
}

func NewAuthService(ctx context.Context, users rez.UserService, pl rez.ProviderLoader, sessionSecretKey string) (*AuthService, error) {
	sessProv, sessErr := pl.LoadAuthSessionProvider(ctx)
	if sessErr != nil {
		return nil, sessErr
	}

	return &AuthService{
		users:         users,
		sessProvider:  sessProv,
		sessionSecret: []byte(sessionSecretKey),
	}, nil
}

func (s *AuthService) MakeRequireAuthMiddleware(redirect bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, sessErr := s.loadAuthSession(r)
			if sessErr == nil {
				authedReq := r.WithContext(context.WithValue(r.Context(), authSessionContextKey, sess))
				next.ServeHTTP(w, authedReq)
				return
			}

			if !errors.Is(sessErr, rez.ErrNoAuthSession) {
				log.Error().Err(sessErr).Msgf("failed to check user authentication")
			} else if redirect {
				log.Debug().Msg("no session, redirecting from mw")
				s.sessProvider.StartAuthFlow(w, r)
				return
			}

			w.WriteHeader(http.StatusUnauthorized)
		})
	}
}

func (s *AuthService) MakeAuthHandler() http.Handler {
	redirect := http.RedirectHandler("/", http.StatusFound)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		onSessionCreated := s.OnAuthSessionCreatedCallback(w, r)
		handledByProvider := s.sessProvider.HandleAuthFlowRequest(w, r, onSessionCreated)
		if handledByProvider {
			return
		}

		_, sessErr := s.loadAuthSession(r)
		if sessErr != nil {
			s.sessProvider.StartAuthFlow(w, r)
			return
		}

		redirect.ServeHTTP(w, r)
	})
}

func (s *AuthService) OnAuthSessionCreatedCallback(w http.ResponseWriter, r *http.Request) func(*rez.AuthSession) {
	return func(sess *rez.AuthSession) {
		provUser := &sess.User
		user, userErr := s.lookupProviderUser(r.Context(), provUser)
		if userErr != nil || user == nil {
			log.Debug().Msg("failed to lookup user from provider details")
			return
		}
		sess.User = *user
		if storeErr := s.storeAuthSession(w, r, sess); storeErr != nil {
			log.Error().Err(storeErr).Msg("failed to store auth session")
		}
	}
}

func (s *AuthService) lookupProviderUser(ctx context.Context, usr *ent.User) (*ent.User, error) {
	// TODO: use provider mapping to match user details
	email := usr.Email
	usr, userErr := s.users.GetByEmail(ctx, email)
	if userErr == nil && usr != nil {
		return usr, nil
	}
	if ent.IsNotFound(userErr) {
		// TODO: user doesn't exist, create?
		log.Debug().Str("email", email).Msg("user doesn't exist")
	} else {
		log.Error().Err(userErr).Str("email", email).Msg("failed to lookup user")
	}
	return nil, userErr
}

func (s *AuthService) storeAuthSession(w http.ResponseWriter, r *http.Request, sess *rez.AuthSession) error {
	token, tokenErr := s.IssueSessionToken(sess)
	if tokenErr != nil {
		return tokenErr
	}

	cookie := &http.Cookie{
		Name:     authSessionCookieName,
		Value:    token,
		Domain:   r.Host,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		// SameSite: http.SameSiteLaxMode,
	}
	if domain, _, splitErr := net.SplitHostPort(r.Host); splitErr == nil {
		cookie.Domain = domain
	}

	fmt.Printf("set cookie!\n")

	http.SetCookie(w, cookie)
	return nil
}

func (s *AuthService) loadAuthSession(r *http.Request) (*rez.AuthSession, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		split := strings.Split(authHeader, " ")
		if len(split) != 2 {
			return nil, errInvalidAuthToken
		}
		authType := split[0]
		token := split[1]
		if authType != "Bearer" {
			log.Debug().Str("authType", authType).Msg("invalid auth type")
			return nil, fmt.Errorf("invalid auth type %s", authType)
		}
		return s.VerifySessionToken(token)
	}

	authCookie, cookieErr := r.Cookie(authSessionCookieName)
	if cookieErr != nil {
		if errors.Is(cookieErr, http.ErrNoCookie) {
			return nil, rez.ErrNoAuthSession
		}
		return nil, cookieErr
	}

	return s.VerifySessionToken(authCookie.Value)
}

func (s *AuthService) GetSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authSessionContextKey).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Session rez.AuthSession `json:"session"`
}

func (s *AuthService) IssueSessionToken(session *rez.AuthSession) (string, error) {
	if session == nil {
		return "", errors.New("nil session")
	}

	claims := jwt.MapClaims{
		"session": *session,
		"exp":     jwt.NewNumericDate(session.ExpiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, signErr := token.SignedString(s.sessionSecret)
	if signErr != nil {
		return "", fmt.Errorf("failed to sign token: %w", signErr)
	}

	return signedToken, nil
}

func (s *AuthService) VerifySessionToken(tokenStr string) (*rez.AuthSession, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
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

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, fmt.Errorf("claims expired: %s", exp)
	}

	return &claims.Session, nil
}
