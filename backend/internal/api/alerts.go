package api

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type alertsHandler struct {
	alerts rez.AlertService
	events rez.OncallEventsService
}

func newAlertsHandler(alerts rez.AlertService, events rez.OncallEventsService) *alertsHandler {
	return &alertsHandler{alerts: alerts, events: events}
}

func (h *alertsHandler) ListAlerts(ctx context.Context, req *oapi.ListAlertsRequest) (*oapi.ListAlertsResponse, error) {
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

func (h *alertsHandler) GetAlert(ctx context.Context, req *oapi.GetAlertRequest) (*oapi.GetAlertResponse, error) {
	var resp oapi.GetAlertResponse

	alert, getErr := h.alerts.GetAlert(ctx, req.Id)
	if getErr != nil {
		return nil, detailError("get alert", getErr)
	}
	resp.Body.Data = oapi.AlertFromEnt(alert)

	return &resp, nil
}

func (h *alertsHandler) GetAlertMetrics(ctx context.Context, req *oapi.GetAlertMetricsRequest) (*oapi.GetAlertMetricsResponse, error) {
	var resp oapi.GetAlertMetricsResponse

	params := rez.GetAlertMetricsParams{
		AlertId:  req.Id,
		RosterId: req.RosterId,
		From:     req.From,
		To:       req.To,
	}
	metrics, getErr := h.alerts.GetAlertMetrics(ctx, params)
	if getErr != nil {
		return nil, detailError("get alert metrics", getErr)
	}
	resp.Body.Data = oapi.AlertMetricsFromEnt(metrics)

	return &resp, nil
}
