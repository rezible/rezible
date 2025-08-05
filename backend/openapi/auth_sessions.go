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

	GetCurrentUserAuthSession(context.Context, *GetCurrentUserAuthSessionRequest) (*GetCurrentUserAuthSessionResponse, error)

	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
}

func (o operations) RegisterAuthSessions(api huma.API) {
	huma.Register(api, GetAuthSessionsConfig, o.GetAuthSessionsConfig)

	huma.Register(api, GetCurrentUserAuthSession, o.GetCurrentUserAuthSession)

	huma.Register(api, ListNotifications, o.ListNotifications)
	huma.Register(api, DeleteNotification, o.DeleteNotification)
}

type (
	AuthSessionsConfig struct {
		ProviderName string `json:"providerName"`
	}

	UserAuthSession struct {
		ExpiresAt    time.Time    `json:"expiresAt"`
		Organization Organization `json:"organization"`
		User         User         `json:"user"`
	}

	Organization struct {
		Id   uuid.UUID `json:"id"`
		Name string    `json:"name"`
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
	Security:    []map[string][]string{},
}

type GetAuthSessionsConfigRequest EmptyRequest
type GetAuthSessionsConfigResponse ItemResponse[AuthSessionsConfig]

var GetCurrentUserAuthSession = huma.Operation{
	OperationID: "get-current-user-auth-session",
	Method:      http.MethodGet,
	Path:        "/auth_session/user",
	Summary:     "Get the Auth Session for the Current User",
	Tags:        authSessionsTags,
	Errors:      errorCodes(),
}

type GetCurrentUserAuthSessionRequest EmptyRequest
type GetCurrentUserAuthSessionResponse ItemResponse[UserAuthSession]

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
