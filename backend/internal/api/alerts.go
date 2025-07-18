package api

import (
	"context"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type alertsHandler struct {
	db *ent.Client
}

func newAlertsHandler(db *ent.Client) *alertsHandler {
	return &alertsHandler{db: db}
}

func (h *alertsHandler) ListAlerts(ctx context.Context, request *oapi.ListAlertsRequest) (*oapi.ListAlertsResponse, error) {
	var resp oapi.ListAlertsResponse

	return &resp, nil
}

func (h *alertsHandler) GetAlert(ctx context.Context, request *oapi.GetAlertRequest) (*oapi.GetAlertResponse, error) {
	var resp oapi.GetAlertResponse

	return &resp, nil
}
