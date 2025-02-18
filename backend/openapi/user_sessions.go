package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type UserSessionsHandler interface {
	GetUserSession(context.Context, *GetUserSessionRequest) (*GetUserSessionResponse, error)
	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
	ListUserAssignments(context.Context, *ListUserAssignmentsRequest) (*ListUserAssignmentsResponse, error)
}

func (o operations) RegisterUserSessions(api huma.API) {
	huma.Register(api, GetUserSession, o.GetUserSession)
	huma.Register(api, ListNotifications, o.ListNotifications)
	huma.Register(api, DeleteNotification, o.DeleteNotification)

	huma.Register(api, ListUserAssignments, o.ListUserAssignments)
}

type (
	UserSession struct {
		ExpiresAt time.Time `json:"expiresAt"`
		User      User      `json:"user"`
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

var sessionTags = []string{"User Session"}

var GetUserSession = huma.Operation{
	OperationID: "get-current-user-session",
	Method:      http.MethodGet,
	Path:        "/user_session",
	Summary:     "Get the Auth Session for the Current User",
	Tags:        sessionTags,
	Errors:      errorCodes(),
}

type GetUserSessionRequest EmptyRequest
type GetUserSessionResponse ItemResponse[UserSession]

var ListNotifications = huma.Operation{
	OperationID: "list-user-notifications",
	Method:      http.MethodGet,
	Path:        "/user_session/notifications",
	Summary:     "List Notifications for the Current User",
	Tags:        sessionTags,
	Errors:      errorCodes(),
}

type ListNotificationsRequest ListRequest
type ListNotificationsResponse PaginatedResponse[UserNotification]

var DeleteNotification = huma.Operation{
	OperationID: "delete-user-notification",
	Method:      http.MethodDelete,
	Path:        "/user_session/notifications/{id}",
	Summary:     "Delete a Notification for the Current User",
	Tags:        sessionTags,
	Errors:      errorCodes(),
}

type DeleteNotificationRequest DeleteIdRequest
type DeleteNotificationResponse EmptyResponse

var ListUserAssignments = huma.Operation{
	OperationID: "list-user-assignments",
	Method:      http.MethodGet,
	Path:        "/user_session/assignments",
	Summary:     "List Assignments for the Current User",
	Tags:        sessionTags,
	Errors:      errorCodes(),
}

type ListUserAssignmentsRequest ListRequest
type ListUserAssignmentsResponse PaginatedResponse[UserAssignment]
