package openapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type AuthSessionsHandler interface {
	GetAuthSessionsConfig(context.Context, *GetAuthSessionsConfigRequest) (*GetAuthSessionsConfigResponse, error)

	GetCurrentAuthSession(context.Context, *GetCurrentAuthSessionRequest) (*GetCurrentAuthSessionResponse, error)

	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
}

func (o operations) RegisterAuthSessions(api huma.API) {
	huma.Register(api, GetAuthSessionsConfig, o.GetAuthSessionsConfig)

	huma.Register(api, GetCurrentAuthSession, o.GetCurrentAuthSession)

	huma.Register(api, ListNotifications, o.ListNotifications)
	huma.Register(api, DeleteNotification, o.DeleteNotification)
}

type (
	AuthSessionsConfig struct {
		Providers []AuthSessionProviderConfig `json:"providers"`
	}

	AuthSessionProviderConfig struct {
		Name              string `json:"name"`
		Enabled           bool   `json:"enabled"`
		StartFlowEndpoint string `json:"startFlowEndpoint"`
	}

	AuthSession struct {
		ExpiresAt    time.Time    `json:"expiresAt"`
		Organization Organization `json:"organization"`
		User         User         `json:"user"`
	}

	Organization struct {
		Id                   uuid.UUID `json:"id"`
		Name                 string    `json:"name"`
		RequiresInitialSetup bool      `json:"requiresInitialSetup"`
	}

	UserNotification struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes UserNotificationAttributes `json:"attributes"`
	}

	UserNotificationAttributes struct {
		Text string `json:"text"`
	}

	UserAssignment struct {
		ItemId   uuid.UUID `json:"itemId"`
		ItemType string    `json:"itemType"`
		Title    string    `json:"title"`
		Deadline time.Time `json:"deadline"`
		Role     string    `json:"role"`
	}
)

// Operations

var authSessionsTags = []string{"Auth Sessions"}

var GetAuthSessionsConfig = huma.Operation{
	OperationID: "get-auth-session-config",
	Method:      http.MethodGet,
	Path:        "/auth_session/config",
	Summary:     "Get the Auth Session config",
	Tags:        authSessionsTags,
	Errors:      errorCodes(),
	Security:    ExplicitNoSecurity,
}

type GetAuthSessionsConfigRequest struct {
	Email string `query:"email" required:"false"`
}
type GetAuthSessionsConfigResponse ItemResponse[AuthSessionsConfig]

var GetCurrentAuthSession = huma.Operation{
	OperationID: "get-current-auth-session",
	Method:      http.MethodGet,
	Path:        "/auth_session",
	Summary:     "Get the Auth Session for the Current User",
	Tags:        authSessionsTags,
	Errors:      errorCodes(),
}

type GetCurrentAuthSessionRequest EmptyRequest
type GetCurrentAuthSessionResponse ItemResponse[AuthSession]

var ListNotifications = huma.Operation{
	OperationID: "list-user-notifications",
	Method:      http.MethodGet,
	Path:        "/auth_session/user/notifications",
	Summary:     "List Notifications for the Current User",
	Tags:        authSessionsTags,
	Errors:      errorCodes(),
}

type ListNotificationsRequest ListRequest
type ListNotificationsResponse PaginatedResponse[UserNotification]

var DeleteNotification = huma.Operation{
	OperationID: "delete-user-notification",
	Method:      http.MethodDelete,
	Path:        "/user_session/notifications/{id}",
	Summary:     "Delete a Notification for the Current User",
	Tags:        authSessionsTags,
	Errors:      errorCodes(),
}

type DeleteNotificationRequest DeleteIdRequest
type DeleteNotificationResponse EmptyResponse
