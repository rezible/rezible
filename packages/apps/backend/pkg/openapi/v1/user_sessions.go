package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type UserSessionsHandler interface {
	GetUserSession(context.Context, *GetUserSessionRequest) (*GetUserSessionResponse, error)

	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
}

func (o operations) RegisterUserSessions(api huma.API) {
	huma.Register(api, GetUserSession, o.GetUserSession)

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
)

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

type DeleteNotificationRequest EmptyIdRequest
type DeleteNotificationResponse EmptyResponse
