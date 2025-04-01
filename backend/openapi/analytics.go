package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type AnalyticsHandler interface {
	GetOncallRosterMetrics(context.Context, *GetOncallRosterMetricsRequest) (*GetOncallRosterMetricsResponse, error)
	GetOncallShiftMetrics(context.Context, *GetOncallShiftMetricsRequest) (*GetOncallShiftMetricsResponse, error)
}

func (o operations) RegisterAnalytics(api huma.API) {
	huma.Register(api, GetOncallRosterMetrics, o.GetOncallRosterMetrics)
	huma.Register(api, GetOncallShiftMetrics, o.GetOncallShiftMetrics)
}

type (
	OncallRosterMetrics struct {
	}

	OncallShiftMetrics struct {
		TotalAlerts          float64                           `json:"totalAlerts"`
		TotalIncidents       float64                           `json:"totalIncidents"`
		IncidentActivity     []OncallShiftIncidentResponseTime `json:"incidentActivity"`
		NightAlerts          float64                           `json:"nightAlerts"`
		AlertIncidentRate    float64                           `json:"alertIncidentRate"`
		BusinessHoursAlerts  float64                           `json:"businessHoursAlerts"`
		OffHoursAlerts       float64                           `json:"offHoursAlerts"`
		OffHoursActivityTime float64                           `json:"offHoursActivityTime"`
		SleepDisruptionScore float64                           `json:"sleepDisruptionScore"`
		WorkloadScore        float64                           `json:"workloadScore"`
		BurdenScore          float64                           `json:"burdenScore"`
	}

	OncallShiftIncidentResponseTime struct {
		IncidentId uuid.UUID `json:"incidentId"`
		Minutes    float64   `json:"minutes"`
	}
)

var analyticsTags = []string{"Analytics"}

var GetOncallRosterMetrics = huma.Operation{
	OperationID: "get-oncall-roster-metrics",
	Method:      http.MethodGet,
	Path:        "/analytics/oncall_rosters",
	Summary:     "Get Metrics for an Oncall Roster",
	Tags:        analyticsTags,
	Errors:      errorCodes(),
}

type GetOncallRosterMetricsRequest struct {
	RosterId uuid.UUID `query:"rosterId"`
}
type GetOncallRosterMetricsResponse ItemResponse[OncallRosterMetrics]

var GetOncallShiftMetrics = huma.Operation{
	OperationID: "get-oncall-shift-metrics",
	Method:      http.MethodGet,
	Path:        "/analytics/oncall_shifts",
	Summary:     "Get Metrics for an Oncall Shift",
	Tags:        analyticsTags,
	Errors:      errorCodes(),
}

type GetOncallShiftMetricsRequest struct {
	ShiftId uuid.UUID `query:"shiftId"`
}
type GetOncallShiftMetricsResponse ItemResponse[OncallShiftMetrics]
