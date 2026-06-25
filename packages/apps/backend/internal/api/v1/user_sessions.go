package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationrole"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type userSessionsHandler struct {
	orgs  rez.OrganizationService
	users rez.UserService
}

func newUserSessionsHandler(orgs rez.OrganizationService, users rez.UserService) *userSessionsHandler {
	return &userSessionsHandler{orgs: orgs, users: users}
}

func (h *userSessionsHandler) GetUserSession(ctx context.Context, req *oapi.GetUserSessionRequest) (*oapi.GetUserSessionResponse, error) {
	var resp oapi.GetUserSessionResponse

	exec := execution.GetContext(ctx)
	userId, userOk := exec.UserID()
	if !userOk {
		return nil, rez.ErrAuthSessionMissing
	}

	u, userErr := h.users.Get(ctx, user.ID(userId))
	if userErr != nil {
		return nil, oapi.Error(ctx, "failed to get user", userErr)
	}

	org, orgErr := h.orgs.Get(ctx, organization.TenantID(u.TenantID))
	if orgErr != nil {
		return nil, oapi.Error(ctx, "failed to get organization", orgErr)
	}

	orgRole := organizationrole.RoleMember.String()
	role, roleErr := u.QueryOrganizationRole().Only(ctx)
	if roleErr != nil && !ent.IsNotFound(roleErr) {
		return nil, oapi.Error(ctx, "failed to get organization role", roleErr)
	}
	if roleErr == nil && role.OrganizationID == org.ID && role.Role == organizationrole.RoleAdmin {
		orgRole = organizationrole.RoleAdmin.String()
	}

	resp.Body.Data = oapi.UserSession{
		User:             oapi.UserFromEnt(u),
		Organization:     oapi.OrganizationFromEnt(org),
		OrganizationRole: orgRole,
		ExpiresAt:        exec.Auth.ExpiresAt,
	}

	return &resp, nil
}

func (h *userSessionsHandler) ListNotifications(ctx context.Context, req *oapi.ListNotificationsRequest) (*oapi.ListNotificationsResponse, error) {
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

func (h *userSessionsHandler) DeleteNotification(ctx context.Context, req *oapi.DeleteNotificationRequest) (*oapi.DeleteNotificationResponse, error) {
	var resp oapi.DeleteNotificationResponse

	// TODO: delete from db

	return &resp, nil
}
