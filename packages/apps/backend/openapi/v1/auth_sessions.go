package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type AuthSessionsHandler interface {
	GetCurrentAuthSession(context.Context, *GetCurrentAuthSessionRequest) (*GetCurrentAuthSessionResponse, error)

	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
}

func (o operations) RegisterAuthSessions(api huma.API) {
	huma.Register(api, GetCurrentAuthSession, o.GetCurrentAuthSession)

	huma.Register(api, ListNotifications, o.ListNotifications)
	huma.Register(api, DeleteNotification, o.DeleteNotification)
}

type (
	AuthSession struct {
		ExpiresAt    time.Time    `json:"expiresAt"`
		Organization Organization `json:"organization"`
		User         User         `json:"user"`
	}

	UserNotification struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes UserNotificationAttributes `json:"attributes"`
	}

	UserNotificationAttributes struct {
		Text string `json:"text"`
	}
)

// Operations

var authSessionsTags = []string{"Auth Sessions"}

var GetCurrentAuthSession = huma.Operation{
	OperationID: "get-current-auth-session",
	Method:      http.MethodGet,
	Path:        "/auth_session/current",
	Summary:     "Get the current Auth Session",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
}

type GetCurrentAuthSessionRequest EmptyRequest
type GetCurrentAuthSessionResponse ItemResponse[AuthSession]

var ListNotifications = huma.Operation{
	OperationID: "list-user-notifications",
	Method:      http.MethodGet,
	Path:        "/auth_session/notifications",
	Summary:     "List Notifications for the Current User",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
}

type ListNotificationsRequest ListRequest
type ListNotificationsResponse ListResponse[UserNotification]

var DeleteNotification = huma.Operation{
	OperationID: "delete-user-notification",
	Method:      http.MethodDelete,
	Path:        "/user_session/notifications/{id}",
	Summary:     "Delete a Notification for the Current User",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
}

type DeleteNotificationRequest DeleteIdRequest
type DeleteNotificationResponse EmptyResponse
