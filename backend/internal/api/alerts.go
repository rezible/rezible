package api

import (
	"context"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type alertsHandler struct {
	alerts rez.AlertService
}

func newAlertsHandler(alerts rez.AlertService) *alertsHandler {
	return &alertsHandler{alerts: alerts}
}

func (h *alertsHandler) ListAlerts(ctx context.Context, request *oapi.ListAlertsRequest) (*oapi.ListAlertsResponse, error) {
	var resp oapi.ListAlertsResponse

	alerts, count, alertsErr := h.alerts.ListAlerts(ctx, rez.ListAlertsParams{})
	if alertsErr != nil {
		return nil, detailError("failed to list alerts", alertsErr)
	}

	resp.Body.Data = make([]oapi.Alert, len(alerts))
	for i, a := range alerts {
		resp.Body.Data[i] = oapi.AlertFromEnt(a)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: count,
	}

	return &resp, nil
}

func (h *alertsHandler) GetAlert(ctx context.Context, request *oapi.GetAlertRequest) (*oapi.GetAlertResponse, error) {
	var resp oapi.GetAlertResponse

	alert, getErr := h.alerts.GetAlert(ctx, request.Id)
	if getErr != nil {
		return nil, detailError("get alert", getErr)
	}

	resp.Body.Data = oapi.AlertFromEnt(alert)

	return &resp, nil
}
