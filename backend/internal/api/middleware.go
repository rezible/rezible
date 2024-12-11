package api

import (
	rez "github.com/twohundreds/rezible"
	oapi "github.com/twohundreds/rezible/openapi"
)

type middlewareHandler struct {
	auth rez.AuthService
}

func newMiddlewareHandler(auth rez.AuthService) *middlewareHandler {
	return &middlewareHandler{auth}
}

func (h *middlewareHandler) GetMiddleware() []oapi.Middleware {
	return []oapi.Middleware{}
}

func (h *middlewareHandler) ensureUserContext(ctx oapi.Context) (oapi.Context, error) {
	sess, sessErr := h.auth.GetSession(ctx.Context())
	if sessErr != nil {
		return nil, oapi.ErrorInternal("failed to get session", sessErr)
	}

	return oapi.WrapContextWithValue(ctx, "auth_session", sess), nil
}
