package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/user"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type authSessionsHandler struct {
	auth  rez.AuthSessionService
	orgs  rez.OrganizationService
	users rez.UserService
}

func newAuthSessionsHandler(auth rez.AuthSessionService, orgs rez.OrganizationService, users rez.UserService) *authSessionsHandler {
	return &authSessionsHandler{auth: auth, orgs: orgs, users: users}
}

func (h *authSessionsHandler) GetCurrentAuthSession(ctx context.Context, input *oapi.GetCurrentAuthSessionRequest) (*oapi.GetCurrentAuthSessionResponse, error) {
	var resp oapi.GetCurrentAuthSessionResponse

	sess := h.auth.GetAuthSession(ctx)

	u, userErr := h.users.Get(ctx, user.ID(sess.UserId))
	if userErr != nil {
		return nil, oapi.Error("failed to get user", userErr)
	}

	org, orgErr := h.orgs.Get(ctx, organization.TenantID(u.TenantID))
	if orgErr != nil {
		return nil, oapi.Error("failed to get organization", orgErr)
	}

	resp.Body.Data = oapi.AuthSession{
		ExpiresAt:    sess.ExpiresAt,
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
