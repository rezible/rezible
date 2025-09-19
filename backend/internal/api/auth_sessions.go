package api

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type authSessionsHandler struct {
	auth  rez.AuthService
	orgs  rez.OrganizationService
	users rez.UserService
}

func newAuthSessionsHandler(auth rez.AuthService, orgs rez.OrganizationService, users rez.UserService) *authSessionsHandler {
	return &authSessionsHandler{auth: auth, orgs: orgs, users: users}
}

func (h *authSessionsHandler) GetAuthSessionsConfig(ctx context.Context, req *oapi.GetAuthSessionsConfigRequest) (*oapi.GetAuthSessionsConfigResponse, error) {
	var resp oapi.GetAuthSessionsConfigResponse

	providers := h.auth.Providers()
	configs := make([]oapi.AuthSessionProviderConfig, len(providers))
	for i, prov := range providers {
		configs[i] = oapi.AuthSessionProviderConfig{
			Name:              prov.DisplayName(),
			StartFlowEndpoint: h.auth.GetProviderStartFlowPath(prov),
			Enabled:           true,
		}
	}

	resp.Body.Data = oapi.AuthSessionsConfig{
		Providers: configs,
	}

	return &resp, nil
}

func (h *authSessionsHandler) GetCurrentAuthSession(ctx context.Context, input *oapi.GetCurrentAuthSessionRequest) (*oapi.GetCurrentAuthSessionResponse, error) {
	var resp oapi.GetCurrentAuthSessionResponse

	sess := getRequestAuthSession(ctx, h.auth)

	user, userErr := h.users.GetById(ctx, sess.UserId)
	if userErr != nil {
		return nil, apiError("failed to get user", userErr)
	}

	org, orgErr := h.orgs.GetCurrent(ctx)
	if orgErr != nil {
		return nil, apiError("failed to get organization", orgErr)
	}

	resp.Body.Data = oapi.AuthSession{
		User: oapi.UserFromEnt(user),
		Organization: oapi.Organization{
			Id:                   org.ID,
			Name:                 org.Name,
			RequiresInitialSetup: org.InitialSetupAt.IsZero(),
		},
		ExpiresAt: sess.ExpiresAt,
	}

	return &resp, nil
}

func (h *authSessionsHandler) ListNotifications(ctx context.Context, request *oapi.ListNotificationsRequest) (*oapi.ListNotificationsResponse, error) {
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

func (h *authSessionsHandler) DeleteNotification(ctx context.Context, request *oapi.DeleteNotificationRequest) (*oapi.DeleteNotificationResponse, error) {
	var resp oapi.DeleteNotificationResponse

	// TODO: delete from db

	return &resp, nil
}
