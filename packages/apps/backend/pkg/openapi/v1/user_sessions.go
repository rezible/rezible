package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type UserSessionsHandler interface {
	GetUserSession(context.Context, *GetUserSessionRequest) (*GetUserSessionResponse, error)
	GetUserSessionPreferences(context.Context, *GetUserSessionPreferencesRequest) (*GetUserSessionPreferencesResponse, error)
	UpdateUserSessionPreferences(context.Context, *UpdateUserSessionPreferencesRequest) (*UpdateUserSessionPreferencesResponse, error)

	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
}

func (o operations) RegisterUserSessions(api huma.API) {
	huma.Register(api, GetUserSession, o.GetUserSession)
	huma.Register(api, GetUserSessionPreferences, o.GetUserSessionPreferences)
	huma.Register(api, UpdateUserSessionPreferences, o.UpdateUserSessionPreferences)

	huma.Register(api, ListNotifications, o.ListNotifications)
	huma.Register(api, DeleteNotification, o.DeleteNotification)
}

type (
	UserSession struct {
		User             User         `json:"user"`
		Organization     Organization `json:"organization"`
		OrganizationRole string       `json:"organizationRole" enum:"admin,member"`
		ExpiresAt        time.Time    `json:"expiresAt"`
	}

	UserNotification struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes UserNotificationAttributes `json:"attributes"`
	}

	UserNotificationAttributes struct {
		Text string `json:"text"`
	}

	UserSessionPreferences struct {
		Profile       UserSessionPreferencesProfile      `json:"profile"`
		Notifications UserSessionNotificationPreferences `json:"notifications"`
	}

	UserSessionPreferencesProfile struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Timezone string `json:"timezone"`
	}

	UserSessionNotificationPreferences struct {
		IncidentUpdates         bool `json:"incidentUpdates"`
		IncidentRoleAssignments bool `json:"incidentRoleAssignments"`
		AgentRunResults         bool `json:"agentRunResults"`
		IntegrationSyncFailures bool `json:"integrationSyncFailures"`
	}
)

func UserSessionPreferencesFromEnt(user *ent.User) UserSessionPreferences {
	notifications := map[string]bool{
		"incidentUpdates":         true,
		"incidentRoleAssignments": true,
		"agentRunResults":         true,
		"integrationSyncFailures": true,
	}
	for key, value := range user.NotificationPreferences {
		notifications[key] = value
	}
	return UserSessionPreferences{
		Profile: UserSessionPreferencesProfile{
			Name:     user.Name,
			Email:    user.Email,
			Timezone: user.Timezone,
		},
		Notifications: UserSessionNotificationPreferences{
			IncidentUpdates:         notifications["incidentUpdates"],
			IncidentRoleAssignments: notifications["incidentRoleAssignments"],
			AgentRunResults:         notifications["agentRunResults"],
			IntegrationSyncFailures: notifications["integrationSyncFailures"],
		},
	}
}

// Operations

var userSessionsTags = []string{"User Session"}

var GetUserSession = huma.Operation{
	OperationID: "get-user-session",
	Method:      http.MethodGet,
	Path:        "/user_session",
	Summary:     "Get the current User Session",
	Tags:        userSessionsTags,
	Errors:      ErrorCodes(),
}

type GetUserSessionRequest EmptyRequest
type GetUserSessionResponse ItemResponse[UserSession]

var GetUserSessionPreferences = huma.Operation{
	OperationID: "get-user-session-preferences",
	Method:      http.MethodGet,
	Path:        "/user_session/preferences",
	Summary:     "Get current user preferences",
	Tags:        userSessionsTags,
	Errors:      ErrorCodes(),
}

type GetUserSessionPreferencesRequest EmptyRequest
type GetUserSessionPreferencesResponse ItemResponse[UserSessionPreferences]

var UpdateUserSessionPreferences = huma.Operation{
	OperationID: "update-user-session-preferences",
	Method:      http.MethodPatch,
	Path:        "/user_session/preferences",
	Summary:     "Update current user preferences",
	Tags:        userSessionsTags,
	Errors:      ErrorCodes(),
}

type UpdateUserSessionPreferencesAttributes struct {
	Name                    *string `json:"name,omitempty"`
	Timezone                *string `json:"timezone,omitempty"`
	IncidentUpdates         *bool   `json:"incidentUpdates,omitempty"`
	IncidentRoleAssignments *bool   `json:"incidentRoleAssignments,omitempty"`
	AgentRunResults         *bool   `json:"agentRunResults,omitempty"`
	IntegrationSyncFailures *bool   `json:"integrationSyncFailures,omitempty"`
}
type UpdateUserSessionPreferencesRequest RequestWithBodyAttributes[UpdateUserSessionPreferencesAttributes]
type UpdateUserSessionPreferencesResponse ItemResponse[UserSessionPreferences]

var ListNotifications = huma.Operation{
	OperationID: "list-user-notifications",
	Method:      http.MethodGet,
	Path:        "/auth_session/notifications",
	Summary:     "List Notifications for the Current User",
	Tags:        userSessionsTags,
	Errors:      ErrorCodes(),
}

type ListNotificationsRequest ListRequest
type ListNotificationsResponse ListResponse[UserNotification]

var DeleteNotification = huma.Operation{
	OperationID: "delete-user-notification",
	Method:      http.MethodDelete,
	Path:        "/user_session/notifications/{id}",
	Summary:     "Delete a Notification for the Current User",
	Tags:        userSessionsTags,
	Errors:      ErrorCodes(),
}

type DeleteNotificationRequest IdRequest
type DeleteNotificationResponse EmptyResponse
