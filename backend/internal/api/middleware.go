package api

import (
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type middlewareHandler struct {
	auth rez.AuthSessionService
}

func newMiddlewareHandler(auth rez.AuthSessionService) *middlewareHandler {
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
