package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type alertsHandler struct {
	alerts rez.AlertService
}

func newAlertsHandler(alerts rez.AlertService) *alertsHandler {
	return &alertsHandler{alerts: alerts}
}

func (h *alertsHandler) ListAlerts(ctx context.Context, req *oapi.ListAlertsRequest) (*oapi.ListAlertsResponse, error) {
	var resp oapi.ListAlertsResponse

	alerts, count, alertsErr := h.alerts.ListAlerts(ctx, rez.ListAlertsParams{})
	if alertsErr != nil {
		return nil, apiError("failed to list alerts", alertsErr)
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
		return nil, apiError("get alert", getErr)
	}
	resp.Body.Data = oapi.AlertFromEnt(alert)

	return &resp, nil
}

func (h *alertsHandler) GetAlertMetrics(ctx context.Context, req *oapi.GetAlertMetricsRequest) (*oapi.GetAlertMetricsResponse, error) {
	var resp oapi.GetAlertMetricsResponse

	dateFrom, dateTo, windowErr := oapi.GetCalendarDateWindow(req.From, req.To)
	if windowErr != nil {
		return nil, apiError("invalid date window", windowErr)
	}

	params := rez.GetAlertMetricsParams{
		AlertId:  req.Id,
		RosterId: req.RosterId,
		From:     dateFrom,
		To:       dateTo,
	}
	metrics, getErr := h.alerts.GetAlertMetrics(ctx, params)
	if getErr != nil {
		return nil, apiError("get alert metrics", getErr)
	}
	resp.Body.Data = oapi.AlertMetricsFromEnt(metrics)

	return &resp, nil
}

func (h *alertsHandler) ListAlertIncidentLinks(context.Context, *oapi.ListAlertIncidentLinksRequest) (*oapi.ListAlertIncidentLinksResponse, error) {
	var resp oapi.ListAlertIncidentLinksResponse

	resp.Body.Data = make([]oapi.AlertIncidentLink, 0)

	return &resp, nil
}
