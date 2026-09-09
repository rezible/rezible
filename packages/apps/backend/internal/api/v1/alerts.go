package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type alertsHandler struct {
	alerts rez.AlertService
}

func newAlertsHandler(alerts rez.AlertService) *alertsHandler {
	return &alertsHandler{alerts: alerts}
}

func (h *alertsHandler) ListAlertDefinitions(ctx context.Context, req *oapi.ListAlertDefinitionsRequest) (*oapi.ListAlertDefinitionsResponse, error) {
	var resp oapi.ListAlertDefinitionsResponse

	params := rez.ListAlertsParams{
		ListParams: req.ListParams(),
	}
	params.Search = req.Search
	alerts, alertsErr := h.alerts.ListAlerts(ctx, params)
	if alertsErr != nil {
		return nil, oapi.Error(ctx, "failed to list alerts", alertsErr)
	}

	resp.Body = oapi.ConvertPaginatedResultBody(alerts, oapi.AlertDefinitionFromEnt)

	return &resp, nil
}

func (h *alertsHandler) GetAlertDefinition(ctx context.Context, req *oapi.GetAlertDefinitionRequest) (*oapi.GetAlertDefinitionResponse, error) {
	var resp oapi.GetAlertDefinitionResponse

	alert, getErr := h.alerts.GetAlert(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get alert", getErr)
	}
	resp.Body.Data = oapi.AlertDefinitionFromEnt(alert)

	return &resp, nil
}

func (h *alertsHandler) GetAlertMetrics(ctx context.Context, req *oapi.GetAlertMetricsRequest) (*oapi.GetAlertMetricsResponse, error) {
	var resp oapi.GetAlertMetricsResponse

	dateFrom, dateTo, windowErr := oapi.GetCalendarDateWindow(req.From, req.To)
	if windowErr != nil {
		return nil, oapi.Error(ctx, "invalid date window", windowErr)
	}

	params := rez.GetAlertMetricsParams{
		AlertId:  req.Id,
		RosterId: req.RosterId,
		From:     dateFrom,
		To:       dateTo,
	}
	metrics, getErr := h.alerts.GetAlertMetrics(ctx, params)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get alert metrics", getErr)
	}
	resp.Body.Data = oapi.AlertMetricsFromEnt(metrics)

	return &resp, nil
}

func (h *alertsHandler) ListAlertIncidentLinks(context.Context, *oapi.ListAlertIncidentLinksRequest) (*oapi.ListAlertIncidentLinksResponse, error) {
	var resp oapi.ListAlertIncidentLinksResponse

	resp.Body.Data = make([]oapi.AlertIncidentLink, 0)

	return &resp, nil
}
