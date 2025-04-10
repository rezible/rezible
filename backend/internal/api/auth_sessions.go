package api

import (
	"context"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type authSessionsHandler struct {
	auth  rez.AuthSessionService
	users rez.UserService
}

func newAuthSessionsHandler(auth rez.AuthSessionService, users rez.UserService) *authSessionsHandler {
	return &authSessionsHandler{auth: auth, users: users}
}

func (h *authSessionsHandler) GetAuthSessionsConfig(ctx context.Context, req *oapi.GetAuthSessionsConfigRequest) (*oapi.GetAuthSessionsConfigResponse, error) {
	var resp oapi.GetAuthSessionsConfigResponse

	resp.Body.Data = oapi.AuthSessionsConfig{
		ProviderName: h.auth.ProviderName(),
	}

	return &resp, nil
}

func (h *authSessionsHandler) GetCurrentUserAuthSession(ctx context.Context, input *oapi.GetCurrentUserAuthSessionRequest) (*oapi.GetCurrentUserAuthSessionResponse, error) {
	var resp oapi.GetCurrentUserAuthSessionResponse

	sess := mustGetAuthSession(ctx, h.auth)
	user, userErr := h.users.GetById(ctx, sess.UserId)
	if userErr != nil {
		return nil, detailError("failed to get user", userErr)
	}

	resp.Body.Data = oapi.UserAuthSession{
		ExpiresAt: sess.ExpiresAt,
		User:      oapi.UserFromEnt(user),
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
