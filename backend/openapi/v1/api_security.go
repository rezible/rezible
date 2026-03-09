package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/openapi"
	"github.com/rs/zerolog/log"
)

var (
	ErrNoSession      = openapi.ErrorUnauthorized("no_session")
	ErrSessionExpired = openapi.ErrorUnauthorized("session_expired")
	ErrMissingScopes  = openapi.ErrorUnauthorized("missing_scopes")
	ErrInvalidUser    = openapi.ErrorUnauthorized("invalid_user")
	ErrInvalidTenant  = openapi.ErrorUnauthorized("invalid_tenant")
	ErrUnknown        = openapi.ErrorUnauthorized("unknown")
)

type SecurityScheme = huma.SecurityScheme

type SecurityMethods = []map[string][]string

const (
	SecurityMethodSessionCookie = "session-cookie"
	SecurityMethodBearerToken   = "api-token"
)

var (
	DefaultSecurity = SecurityMethods{
		{SecurityMethodSessionCookie: {}},
		{SecurityMethodBearerToken: {}},
	}
	ExplicitNoSecurity = SecurityMethods{}
)

func GetDefaultSecuritySchemes() map[string]*SecurityScheme {
	sessionCookieSecurityScheme := &SecurityScheme{
		Name: rez.Config.AuthSessionCookieName(),
		Type: "apiKey",
		In:   "cookie",
	}
	apiTokenSecurityScheme := &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	}
	return map[string]*SecurityScheme{
		SecurityMethodSessionCookie: sessionCookieSecurityScheme,
		SecurityMethodBearerToken:   apiTokenSecurityScheme,
	}
}

func isExplicitNoSecurity(s SecurityMethods) bool {
	return s != nil && len(s) == 0
}

func MakeSecurityMiddleware(auth rez.AuthService) openapi.Middleware {
	return func(c openapi.Context, next func(openapi.Context)) {
		opSecurity := c.Operation().Security
		ctx := access.AnonymousContext(c.Context())
		if !isExplicitNoSecurity(opSecurity) {
			if opSecurity == nil {
				opSecurity = DefaultSecurity
			}
			sess, verifyErr := auth.VerifyAuthSessionToken(extractRequestTokenAndMethodScopes(c, opSecurity))
			if verifyErr != nil {
				writeAuthStatusError(c, verifyErr)
				return
			}
			authCtx, authCtxErr := auth.CreateAuthContext(ctx, sess)
			if authCtxErr != nil {
				writeAuthStatusError(c, authCtxErr)
				return
			}
			ctx = authCtx
		}

		next(huma.WithContext(c, ctx))
	}
}

func extractRequestTokenAndMethodScopes(c openapi.Context, opSecurity SecurityMethods) (string, []string) {
	var bearerToken string
	if split := strings.Split(c.Header("Authorization"), " "); len(split) == 2 && split[0] == "Bearer" {
		bearerToken = split[1]
	}
	var cookieToken string
	r, _ := humago.Unwrap(c)
	if authCookie, cookieErr := r.Cookie(rez.Config.AuthSessionCookieName()); cookieErr == nil {
		cookieToken = authCookie.Value
	}
	for _, methodScopes := range opSecurity {
		bearerTokenScopes, bearerTokenAllowed := methodScopes[SecurityMethodBearerToken]
		if bearerToken != "" && bearerTokenAllowed {
			return bearerToken, bearerTokenScopes
		}
		cookieTokenScopes, cookieTokenAllowed := methodScopes[SecurityMethodSessionCookie]
		if cookieToken != "" && cookieTokenAllowed {
			return cookieToken, cookieTokenScopes
		}
	}
	return "", nil
}

func writeAuthStatusError(c openapi.Context, err error) {
	var resp openapi.StatusError
	if errors.Is(err, rez.ErrNoAuthSession) {
		resp = ErrNoSession
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		resp = ErrSessionExpired
	} else if errors.Is(err, rez.ErrInvalidUser) {
		resp = ErrInvalidUser
	} else if errors.Is(err, rez.ErrInvalidTenant) {
		resp = ErrInvalidTenant
	} else {
		resp = ErrUnknown
	}

	status := resp.GetStatus()
	if jsonErr := json.NewEncoder(c.BodyWriter()).Encode(resp); jsonErr != nil {
		log.Error().Err(jsonErr).Msg("failed to write error body")
		status = http.StatusInternalServerError
	}
	c.SetStatus(status)
}
