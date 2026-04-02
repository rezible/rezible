package apiv1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/user"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type authSessionsHandler struct {
	auth  rez.AuthService
	orgs  rez.OrganizationService
	users rez.UserService
}

func newAuthSessionsHandler(auth rez.AuthService, orgs rez.OrganizationService, users rez.UserService) *authSessionsHandler {
	return &authSessionsHandler{auth: auth, orgs: orgs, users: users}
}

func (h *authSessionsHandler) CompleteAuthSessionFlow(ctx context.Context, req *oapi.CompleteAuthSessionFlowRequest) (*oapi.CompleteAuthSessionFlowResponse, error) {
	var resp oapi.CompleteAuthSessionFlowResponse

	attr := req.Body.Attributes
	cookies, flowErr := h.auth.CompleteClientAuthSessionFlow(ctx, attr.Code, attr.Verifier)
	if flowErr != nil {
		if errors.Is(flowErr, rez.ErrDomainNotAllowed) {
			return nil, oapi.ErrDomainNotAllowed
		}
		return nil, oapi.Error("failed to complete auth session flow", flowErr)
	}
	resp.SetCookie = cookies

	return &resp, nil
}

func (h *authSessionsHandler) RefreshAuthSession(ctx context.Context, req *oapi.RefreshAuthSessionRequest) (*oapi.RefreshAuthSessionResponse, error) {
	var resp oapi.RefreshAuthSessionResponse

	refreshToken := req.Cookie.Value
	if refreshToken == "" {
		return nil, oapi.Error("no refresh cookie", oapi.ErrAuthSessionInvalid)
	}
	cookies, cookiesErr := h.auth.RefreshClientAuthSession(ctx, refreshToken)
	if cookiesErr != nil {
		return nil, oapi.Error("failed to refresh session cookies", cookiesErr)
	}
	resp.SetCookie = cookies

	return &resp, nil
}

func (h *authSessionsHandler) ClearAuthSession(ctx context.Context, req *oapi.ClearAuthSessionRequest) (*oapi.ClearAuthSessionResponse, error) {
	var resp oapi.ClearAuthSessionResponse

	cookies, cookiesErr := h.auth.ClearClientAuthSession()
	if cookiesErr != nil {
		return nil, oapi.Error("failed to clear auth session", cookiesErr)
	}
	resp.SetCookie = cookies

	return &resp, nil
}

func (h *authSessionsHandler) GetCurrentAuthSession(ctx context.Context, input *oapi.GetCurrentAuthSessionRequest) (*oapi.GetCurrentAuthSessionResponse, error) {
	var resp oapi.GetCurrentAuthSessionResponse

	sess := h.auth.GetAuthSession(ctx)

	u, userErr := h.users.Get(ctx, user.ID(sess.UserId()))
	if userErr != nil {
		return nil, oapi.Error("failed to get user", userErr)
	}

	org, orgErr := h.orgs.GetCurrent(ctx)
	if orgErr != nil {
		return nil, oapi.Error("failed to get organization", orgErr)
	}

	resp.Body.Data = oapi.AuthSession{
		ExpiresAt:    sess.ExpiresAt(),
		User:         oapi.UserFromEnt(u),
		Organization: oapi.OrganizationFromEnt(org),
	}

	return &resp, nil
}

func (h *authSessionsHandler) ListNotifications(ctx context.Context, request *oapi.ListNotificationsRequest) (*oapi.ListNotificationsResponse, error) {
	var resp oapi.ListNotificationsResponse

	// TODO: fetch from db
	notifs := []oapi.UserNotification{
		{
			Id: uuid.New(),
			Attributes: oapi.UserNotificationAttributes{
				Text: "bleh",
			},
		},
	}

	resp.Body.Data = make([]oapi.UserNotification, len(notifs))
	for i, notif := range notifs {
		resp.Body.Data[i] = notif
	}

	return &resp, nil
}

func (h *authSessionsHandler) DeleteNotification(ctx context.Context, request *oapi.DeleteNotificationRequest) (*oapi.DeleteNotificationResponse, error) {
	var resp oapi.DeleteNotificationResponse

	// TODO: delete from db

	return &resp, nil
}
