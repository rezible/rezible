package apiv1

import (
	"context"
	"strings"

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

	resp.Body.Data = oapi.UserSession{
		User:         oapi.UserFromEnt(u),
		Organization: oapi.OrganizationFromEnt(org),
	}

	if exec.Auth.ExpiresAt != nil {
		resp.Body.Data.ExpiresAt = *exec.Auth.ExpiresAt
	}

	role, roleErr := u.QueryOrganizationRole().Only(ctx)
	if roleErr != nil && !ent.IsNotFound(roleErr) {
		return nil, oapi.Error(ctx, "failed to get organization role", roleErr)
	}

	resp.Body.Data.OrganizationRole = organizationrole.RoleMember.String()
	if roleErr == nil && role.OrganizationID == org.ID && role.Role == organizationrole.RoleAdmin {
		resp.Body.Data.OrganizationRole = organizationrole.RoleAdmin.String()
	}

	return &resp, nil
}

func (h *userSessionsHandler) getCurrentUser(ctx context.Context) (*ent.User, error) {
	exec := execution.GetContext(ctx)
	userId, userOk := exec.UserID()
	if !userOk {
		return nil, rez.ErrAuthSessionMissing
	}
	u, userErr := h.users.Get(ctx, user.ID(userId))
	if userErr != nil {
		return nil, oapi.Error(ctx, "failed to get user", userErr)
	}
	return u, nil
}

func (h *userSessionsHandler) GetUserSessionPreferences(ctx context.Context, req *oapi.GetUserSessionPreferencesRequest) (*oapi.GetUserSessionPreferencesResponse, error) {
	var resp oapi.GetUserSessionPreferencesResponse

	u, userErr := h.getCurrentUser(ctx)
	if userErr != nil {
		return nil, userErr
	}

	resp.Body.Data = oapi.UserSessionPreferencesFromEnt(u)
	return &resp, nil
}

func (h *userSessionsHandler) UpdateUserSessionPreferences(ctx context.Context, req *oapi.UpdateUserSessionPreferencesRequest) (*oapi.UpdateUserSessionPreferencesResponse, error) {
	var resp oapi.UpdateUserSessionPreferencesResponse

	curr, currErr := h.getCurrentUser(ctx)
	if currErr != nil {
		return nil, currErr
	}

	attrs := req.Body.Attributes

	reqPrefs := map[string]*bool{
		"incidentUpdates":         attrs.IncidentUpdates,
		"incidentRoleAssignments": attrs.IncidentRoleAssignments,
		"agentRunResults":         attrs.AgentRunResults,
		"integrationSyncFailures": attrs.IntegrationSyncFailures,
	}

	notificationPrefs := map[string]bool{}
	for key, value := range curr.NotificationPreferences {
		notificationPrefs[key] = value
	}
	for key, value := range reqPrefs {
		if value != nil {
			notificationPrefs[key] = *value
		}
	}

	u, updateErr := h.users.Set(ctx, curr.ID, func(m *ent.UserMutation) {
		if attrs.Name != nil {
			m.SetName(strings.TrimSpace(*attrs.Name))
		}
		if attrs.Timezone != nil {
			m.SetTimezone(strings.TrimSpace(*attrs.Timezone))
		}
		m.SetNotificationPreferences(notificationPrefs)
	})
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to update preferences", updateErr)
	}

	resp.Body.Data = oapi.UserSessionPreferencesFromEnt(u)
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
