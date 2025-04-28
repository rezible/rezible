package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type OncallMetricsHandler interface {
	GetOncallRosterMetrics(context.Context, *GetOncallRosterMetricsRequest) (*GetOncallRosterMetricsResponse, error)
	GetOncallShiftMetrics(context.Context, *GetOncallShiftMetricsRequest) (*GetOncallShiftMetricsResponse, error)
}

func (o operations) RegisterOncallMetrics(api huma.API) {
	huma.Register(api, GetOncallRosterMetrics, o.GetOncallRosterMetrics)
	huma.Register(api, GetOncallShiftMetrics, o.GetOncallShiftMetrics)
}

type (
	OncallRosterMetrics struct {
		HealthScore        float32 `json:"healthScore"`
		Incidents          int     `json:"incidents"`
		Alerts             int     `json:"alerts"`
		NightAlerts        int     `json:"nightAlerts"`
		OutOfHoursAlerts   int     `json:"outOfHoursAlerts"`
		AlertActionability float32 `json:"alertActionability"`
		HandoverCompletion float32 `json:"handoverCompletion"`
		BacklogBurnRate    float32 `json:"backlogBurnRate"`
	}

	OncallShiftMetrics struct {
		BurdenScore          float32                           `json:"burdenScore"`
		Incidents            int                               `json:"incidents"`
		IncidentActivity     []OncallShiftIncidentResponseTime `json:"incidentActivity"`
		Alerts               int                               `json:"alerts"`
		NightAlerts          int                               `json:"nightAlerts"`
		AlertActionability   float32                           `json:"alertActionability"`
		AlertIncidentRate    float32                           `json:"alertIncidentRate"`
		OffHoursAlerts       int                               `json:"offHoursAlerts"`
		OffHoursActivityTime float32                           `json:"offHoursActivityTime"`
		SleepDisruptionScore float32                           `json:"sleepDisruptionScore"`
		WorkloadScore        float32                           `json:"workloadScore"`
	}

	OncallShiftIncidentResponseTime struct {
		IncidentId uuid.UUID `json:"incidentId"`
		Minutes    float32   `json:"minutes"`
	}
)

var oncallMetricsTags = []string{"Oncall Metrics"}

var GetOncallRosterMetrics = huma.Operation{
	OperationID: "get-oncall-roster-metrics",
	Method:      http.MethodGet,
	Path:        "/oncall_metrics/rosters",
	Summary:     "Get Metrics for an Oncall Roster",
	Tags:        oncallMetricsTags,
	Errors:      errorCodes(),
}

type GetOncallRosterMetricsRequest struct {
	RosterId uuid.UUID `query:"rosterId"`
}
type GetOncallRosterMetricsResponse ItemResponse[OncallRosterMetrics]

var GetOncallShiftMetrics = huma.Operation{
	OperationID: "get-oncall-shift-metrics",
	Method:      http.MethodGet,
	Path:        "/oncall_metrics/shifts",
	Summary:     "Get Metrics for an Oncall Shift",
	Tags:        oncallMetricsTags,
	Errors:      errorCodes(),
}

type GetOncallShiftMetricsRequest struct {
	ShiftId uuid.UUID `query:"shiftId"`
}
type GetOncallShiftMetricsResponse ItemResponse[OncallShiftMetrics]
