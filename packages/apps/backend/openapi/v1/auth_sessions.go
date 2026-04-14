package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type AuthSessionsHandler interface {
	GetAuthSessionConfig(context.Context, *GetAuthSessionConfigRequest) (*GetAuthSessionConfigResponse, error)
	GetCurrentAuthSession(context.Context, *GetCurrentAuthSessionRequest) (*GetCurrentAuthSessionResponse, error)
	CompleteAuthSessionFlow(context.Context, *CompleteAuthSessionFlowRequest) (*CompleteAuthSessionFlowResponse, error)
	RefreshAuthSession(context.Context, *RefreshAuthSessionRequest) (*RefreshAuthSessionResponse, error)
	ClearAuthSession(context.Context, *ClearAuthSessionRequest) (*ClearAuthSessionResponse, error)

	ListNotifications(context.Context, *ListNotificationsRequest) (*ListNotificationsResponse, error)
	DeleteNotification(context.Context, *DeleteNotificationRequest) (*DeleteNotificationResponse, error)
}

func (o operations) RegisterAuthSessions(api huma.API) {
	huma.Register(api, GetAuthSessionConfig, o.GetAuthSessionConfig)

	huma.Register(api, GetCurrentAuthSession, o.GetCurrentAuthSession)
	huma.Register(api, CompleteAuthSessionFlow, o.CompleteAuthSessionFlow)
	huma.Register(api, RefreshAuthSession, o.RefreshAuthSession)
	huma.Register(api, ClearAuthSession, o.ClearAuthSession)

	huma.Register(api, ListNotifications, o.ListNotifications)
	huma.Register(api, DeleteNotification, o.DeleteNotification)
}

type (
	AuthSessionConfig struct {
		Issuer          string   `json:"issuer"`
		AppClientId     string   `json:"app_client_id"`
		AppClientScopes []string `json:"app_client_scopes"`
	}

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

var GetAuthSessionConfig = huma.Operation{
	OperationID: "get-auth-session-config",
	Method:      http.MethodGet,
	Path:        "/auth_session/config",
	Summary:     "Get auth session info and config",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
	Security:    ExplicitNoSecurity,
}

type GetAuthSessionConfigRequest EmptyRequest
type GetAuthSessionConfigResponse ItemResponse[AuthSessionConfig]

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

var CompleteAuthSessionFlow = huma.Operation{
	OperationID: "complete-auth-session-flow",
	Method:      http.MethodPost,
	Path:        "/auth_session/create",
	Summary:     "Complete an Auth Session flow",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
	Security:    ExplicitNoSecurity,
}

type CompleteAuthSessionFlowRequestAttributes struct {
	Code     string `json:"code"`
	Verifier string `json:"verifier"`
}
type CompleteAuthSessionFlowRequest RequestWithBodyAttributes[CompleteAuthSessionFlowRequestAttributes]
type CompleteAuthSessionFlowResponse SetCookieResponse

var RefreshAuthSession = huma.Operation{
	OperationID: "refresh-auth-session",
	Method:      http.MethodPost,
	Path:        "/auth_session/refresh",
	Summary:     "Refresh an active Auth Session",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
	Security:    SecurityMethodCookieOnly,
}

type RefreshAuthSessionRequest RequestWithRefreshTokenCookie
type RefreshAuthSessionResponse SetCookieResponse

var ClearAuthSession = huma.Operation{
	OperationID: "clear-auth-session",
	Method:      http.MethodPost,
	Path:        "/auth_session/clear",
	Summary:     "Clear an active Auth Session",
	Tags:        authSessionsTags,
	Errors:      ErrorCodes(),
	Security:    ExplicitNoSecurity,
}

type ClearAuthSessionRequest EmptyRequest
type ClearAuthSessionResponse SetCookieResponse

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
