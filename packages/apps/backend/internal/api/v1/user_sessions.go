package apiv1

import (
	"context"
	"maps"
	"sort"
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
		return nil, oapi.Error(ctx, "get user session", rez.ErrAuthSessionMissing)
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
		return nil, oapi.Error(ctx, "get current user", rez.ErrAuthSessionMissing)
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

var inboxItemExamples = []oapi.InboxItem{
	{
		Id: uuid.MustParse("77000000-0000-4000-8000-000000000001"),
		Attributes: oapi.InboxItemAttributes{
			Kind:        "question",
			Reason:      "Confirm deploy correlation.",
			RecipientId: new(uuid.MustParse("55555555-5555-4555-8555-555555555555")),
			TargetKind:  "discussion-thread",
			TargetId:    uuid.MustParse("88000000-0000-4000-8000-000000000001"),
			State:       "open",
			Context:     "Checkout investigation",
			AnalysisId:  new(uuid.MustParse("22000000-0000-4000-8000-000000000001")),
			TeamId:      new(uuid.MustParse("33333333-3333-4333-8333-333333333333")),
		},
	},
	{
		Id: uuid.MustParse("77000000-0000-4000-8000-000000000002"),
		Attributes: oapi.InboxItemAttributes{
			Kind:        "annotation",
			Reason:      "Annotate rollback evidence.",
			RecipientId: new(uuid.MustParse("55555555-5555-4555-8555-555555555555")),
			TargetKind:  "normalized-event",
			TargetId:    uuid.MustParse("30000000-0000-4000-8000-000000000004"),
			State:       "open",
			Context:     "Checkout analysis",
			AnalysisId:  new(uuid.MustParse("22000000-0000-4000-8000-000000000001")),
			TeamId:      new(uuid.MustParse("33333333-3333-4333-8333-333333333333")),
		},
	},
	{
		Id: uuid.MustParse("77000000-0000-4000-8000-000000000003"),
		Attributes: oapi.InboxItemAttributes{
			Kind:        "task",
			Reason:      "Validate pool limit.",
			RecipientId: new(uuid.MustParse("55555555-5555-4555-8555-555555555555")),
			TargetKind:  "task",
			TargetId:    uuid.MustParse("99000000-0000-4000-8000-000000000001"),
			State:       "completed",
			Context:     "Checkout incident",
			IncidentId:  new(uuid.MustParse("44000000-0000-4000-8000-000000000001")),
			TeamId:      new(uuid.MustParse("33333333-3333-4333-8333-333333333333")),
		},
	},
	{
		Id: uuid.MustParse("77000000-0000-4000-8000-000000000004"),
		Attributes: oapi.InboxItemAttributes{
			Kind:           "maintenance",
			Reason:         "Review runbook update.",
			RecipientId:    new(uuid.MustParse("55555555-5555-4555-8555-555555555555")),
			TargetKind:     "maintenance-request",
			TargetId:       uuid.MustParse("aa000000-0000-4000-8000-000000000001"),
			State:          "open",
			Context:        "Reliability maintenance",
			ProposedChange: "Add rollback verification and pool saturation checks to the checkout runbook.",
			TeamId:         new(uuid.MustParse("33333333-3333-4333-8333-333333333333")),
		},
	},
}

func (h *userSessionsHandler) ListInboxItems(ctx context.Context, request *oapi.ListInboxItemsRequest) (*oapi.ListInboxItemsResponse, error) {
	viewer, ok := execution.GetContext(ctx).UserID()
	if request.Scope == "mine" && !ok {
		return nil, oapi.Error(ctx, "authenticated viewer is required for Mine", rez.ErrAuthSessionMissing)
	}
	if request.Scope == "team" && request.TeamId == uuid.Nil {
		return nil, oapi.Error(ctx, "teamId is required for Team", rez.ErrInvalidInput)
	}
	items := make([]oapi.InboxItem, 0)
	for _, inboxItem := range inboxItemExamples {
		if request.RecipientId != uuid.Nil && (inboxItem.Attributes.RecipientId == nil || *inboxItem.Attributes.RecipientId != request.RecipientId) {
			continue
		}
		if request.Kind != "" && request.Kind != inboxItem.Attributes.Kind {
			continue
		}
		if request.State != "" && request.State != inboxItem.Attributes.State {
			continue
		}
		if request.Scope == "mine" && (inboxItem.Attributes.RecipientId == nil || *inboxItem.Attributes.RecipientId != viewer) {
			continue
		}
		if request.TeamId != uuid.Nil && (inboxItem.Attributes.TeamId == nil || *inboxItem.Attributes.TeamId != request.TeamId) {
			continue
		}
		items = append(items, inboxItem)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Id.String() < items[j].Id.String() })
	page, pageSize := request.Page, request.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	start := min((page-1)*pageSize, len(items))
	end := min(start+pageSize, len(items))
	var response oapi.ListInboxItemsResponse
	response.Body.Data = items[start:end]
	response.Body.Pagination = oapi.Pagination{Page: page, PageSize: pageSize, Total: len(items)}
	return &response, nil
}

func (*userSessionsHandler) GetInboxItem(ctx context.Context, request *oapi.GetInboxItemRequest) (*oapi.GetInboxItemResponse, error) {
	for _, item := range inboxItemExamples {
		if item.Id == request.Id {
			response := &oapi.GetInboxItemResponse{}
			response.Body.Data = item
			return response, nil
		}
	}
	return nil, oapi.Error(ctx, "inbox item not found", rez.ErrNotFound)
}
