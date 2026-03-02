package v1

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/openapi"
)

func extractRequestTokenAndScopes(r *http.Request, opSecurity openapi.SecurityMethods) (string, []string) {
	apiToken := openapi.GetRequestBearerToken(r)
	cookieToken := openapi.GetRequestSessionCookieToken(r)
	for _, methodScopes := range opSecurity {
		apiTokenScopes, apiTokenAllowed := methodScopes[openapi.SecurityMethodApiToken]
		if apiTokenAllowed && apiToken != "" {
			return apiToken, apiTokenScopes
		}
		cookieTokenScopes, cookieTokenAllowed := methodScopes[openapi.SecurityMethodSessionCookie]
		if cookieTokenAllowed && cookieToken != "" {
			return cookieToken, cookieTokenScopes
		}
	}
	return "", nil
}

func MakeSecurityMiddleware(auth rez.AuthService) openapi.Middleware {
	return func(c openapi.Context, next func(openapi.Context)) {
		r, w := humago.Unwrap(c)
		opSecurity := c.Operation().Security

		ctx := c.Context()
		if !openapi.IsExplicitNoSecurity(opSecurity) {
			if opSecurity == nil {
				opSecurity = openapi.DefaultSecurity
			}
			token, methodScopes := extractRequestTokenAndScopes(r, opSecurity)

			var scopes rez.AuthSessionScopes
			if len(methodScopes) > 0 {
				scopes["api"] = methodScopes
			}

			sess, sessErr := auth.VerifyAuthSessionToken(token, scopes)
			if sessErr != nil {
				writeAuthStatusError(w, sessErr)
				return
			}

			authCtx, authCtxErr := auth.CreateAuthContext(ctx, sess)
			if authCtxErr != nil {
				writeAuthStatusError(w, authCtxErr)
				return
			}

			ctx = authCtx
		} else {
			ctx = access.AnonymousContext(ctx)
		}

		next(huma.WithContext(c, ctx))
	}
}

func writeAuthStatusError(w http.ResponseWriter, err error) {
	var resp openapi.StatusError
	if errors.Is(err, rez.ErrNoAuthSession) {
		resp = openapi.ErrNoSession
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		resp = openapi.ErrSessionExpired
	} else if errors.Is(err, rez.ErrInvalidUser) {
		resp = openapi.ErrInvalidUser
	} else if errors.Is(err, rez.ErrInvalidTenant) {
		resp = openapi.ErrInvalidTenant
	} else if errors.Is(err, rez.ErrUnauthorized) {
		resp = openapi.ErrUnauthorized
	} else {
		resp = openapi.ErrUnknown
	}

	respBody, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.GetStatus())
	if _, writeErr := w.Write(respBody); writeErr != nil {
		// TODO: log?
	}
}
