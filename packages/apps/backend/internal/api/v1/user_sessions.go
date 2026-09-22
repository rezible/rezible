package apiv1

import (
	"context"
	"maps"
	"strings"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationrole"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type userSessionsHandler struct {
	*baseHandler
	orgs rez.OrganizationService

	inboxExamples []oapi.InboxItem
}

func newUserSessionsHandler(bh *baseHandler, orgs rez.OrganizationService) *userSessionsHandler {
	return &userSessionsHandler{baseHandler: bh, orgs: orgs}
}

func (h *userSessionsHandler) GetUserSession(ctx context.Context, req *oapi.GetUserSessionRequest) (*oapi.GetUserSessionResponse, error) {
	var resp oapi.GetUserSessionResponse

	auth := h.mustAuth(ctx)

	var expiresAt time.Time
	if auth.ExpiresAt != nil {
		expiresAt = *auth.ExpiresAt
	}

	u, userErr := h.currentUser(ctx)
	if userErr != nil {
		return nil, oapi.Error(ctx, "failed to get user", userErr)
	}

	org, orgErr := h.orgs.Get(ctx, organization.TenantID(u.TenantID))
	if orgErr != nil {
		return nil, oapi.Error(ctx, "failed to get organization", orgErr)
	}

	userOrgRole := organizationrole.RoleMember
	orgRole, queryOrgRoleErr := u.QueryOrganizationRole().Only(ctx)
	if queryOrgRoleErr != nil && !ent.IsNotFound(queryOrgRoleErr) {
		return nil, oapi.Error(ctx, "failed to query user org role", queryOrgRoleErr)
	} else if orgRole != nil {
		userOrgRole = orgRole.Role
	}

	resp.Body.Data = oapi.UserSession{
		User:             oapi.UserFromEnt(u),
		Organization:     oapi.OrganizationFromEnt(org),
		OrganizationRole: userOrgRole,
		ExpiresAt:        expiresAt,
	}

	return &resp, nil
}

func (h *userSessionsHandler) GetUserSessionPreferences(ctx context.Context, req *oapi.GetUserSessionPreferencesRequest) (*oapi.GetUserSessionPreferencesResponse, error) {
	var resp oapi.GetUserSessionPreferencesResponse

	u, userErr := h.currentUser(ctx)
	if userErr != nil {
		return nil, userErr
	}

	resp.Body.Data = oapi.UserSessionPreferencesFromEnt(u)
	return &resp, nil
}

func (h *userSessionsHandler) UpdateUserSessionPreferences(ctx context.Context, req *oapi.UpdateUserSessionPreferencesRequest) (*oapi.UpdateUserSessionPreferencesResponse, error) {
	var resp oapi.UpdateUserSessionPreferencesResponse

	curr, currErr := h.currentUser(ctx)
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
	maps.Copy(notificationPrefs, curr.NotificationPreferences)
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

	resp.Body.Data = make([]oapi.UserNotification, 0)
	resp.Body.Pagination = oapi.Pagination{Page: req.Page, PageSize: req.PageSize, Total: 0}

	return &resp, nil
}

func (h *userSessionsHandler) DeleteNotification(ctx context.Context, req *oapi.DeleteNotificationRequest) (*oapi.DeleteNotificationResponse, error) {
	var resp oapi.DeleteNotificationResponse

	// TODO: delete from db

	return &resp, nil
}

func (h *userSessionsHandler) ListInboxItems(ctx context.Context, request *oapi.ListInboxItemsRequest) (*oapi.ListInboxItemsResponse, error) {
	var response oapi.ListInboxItemsResponse
	response.Body.Data = make([]oapi.InboxItem, 0)
	response.Body.Pagination = oapi.Pagination{Page: 1, PageSize: 25, Total: 0}
	return &response, nil
}

func (h *userSessionsHandler) GetInboxItem(ctx context.Context, request *oapi.GetInboxItemRequest) (*oapi.GetInboxItemResponse, error) {
	return nil, oapi.Error(ctx, "inbox item not found", rez.ErrNotFound)
}
