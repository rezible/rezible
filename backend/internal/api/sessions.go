package api

import (
	"context"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type sessionsHandler struct {
	auth rez.AuthService
}

func newSessionsHandler(auth rez.AuthService) *sessionsHandler {
	return &sessionsHandler{auth}
}

func (h *sessionsHandler) GetUserSession(ctx context.Context, input *oapi.GetUserSessionRequest) (*oapi.GetUserSessionResponse, error) {
	var resp oapi.GetUserSessionResponse

	sess := mustGetAuthSession(ctx, h.auth)
	resp.Body.Data = oapi.UserSessionFromRez(sess)

	return &resp, nil
}

func (h *sessionsHandler) ListNotifications(ctx context.Context, request *oapi.ListNotificationsRequest) (*oapi.ListNotificationsResponse, error) {
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

func (h *sessionsHandler) DeleteNotification(ctx context.Context, request *oapi.DeleteNotificationRequest) (*oapi.DeleteNotificationResponse, error) {
	var resp oapi.DeleteNotificationResponse

	// TODO: delete from db

	return &resp, nil
}

func (h *sessionsHandler) ListUserAssignments(context.Context, *oapi.ListUserAssignmentsRequest) (*oapi.ListUserAssignmentsResponse, error) {
	var resp oapi.ListUserAssignmentsResponse

	return &resp, nil
}
