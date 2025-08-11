package openapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type OncallMetricsHandler interface {
	GetOncallRosterMetrics(context.Context, *GetOncallRosterMetricsRequest) (*GetOncallRosterMetricsResponse, error)
	GetOncallShiftMetrics(context.Context, *GetOncallShiftMetricsRequest) (*GetOncallShiftMetricsResponse, error)
	GetOncallShiftBurdenMetricWeights(context.Context, *GetOncallShiftBurdenMetricWeightsRequest) (*GetOncallShiftBurdenMetricWeightsResponse, error)
}

func (o operations) RegisterOncallMetrics(api huma.API) {
	huma.Register(api, GetOncallRosterMetrics, o.GetOncallRosterMetrics)
	huma.Register(api, GetOncallShiftMetrics, o.GetOncallShiftMetrics)
	huma.Register(api, GetOncallShiftBurdenMetricWeights, o.GetOncallShiftBurdenMetricWeights)
}

type (
	OncallRosterMetrics struct {
		ShiftMetrics       []OncallShiftMetrics `json:"shiftMetrics"`
		HandoverCompletion float32              `json:"handoverCompletion"`
		BacklogBurnRate    float32              `json:"backlogBurnRate"`
	}

	OncallShiftBurdenMetricWeights struct {
	}

	OncallShiftMetrics struct {
		Burden OncallShiftMetricsBurden `json:"burden"`
		Events OncallShiftMetricsEvents `json:"events"`
	}

	OncallShiftMetricsBurden struct {
		FinalScore           float32 `json:"finalScore"`
		EventFrequency       float32 `json:"eventFrequency"`
		LifeImpact           float32 `json:"lifeImpact"`
		TimeImpact           float32 `json:"timeImpact"`
		ResponseRequirements float32 `json:"responseRequirements"`
		Isolation            float32 `json:"isolation"`
	}

	OncallShiftMetricsEvents struct {
		Total float32 `json:"total"`

		TotalInterrupts         float32 `json:"totalInterrupts"`
		InterruptsBusinessHours float32 `json:"interruptsBusinessHours"`
		InterruptsNight         float32 `json:"interruptsNight"`
		InterruptResponseTime   float32 `json:"interruptResponseTime"`

		TotalIncidents float32 `json:"totalIncidents"`
		IncidentTime   float32 `json:"incidentTime"`

		TotalAlerts           float32 `json:"totalAlerts"`
		AlertIncidentRate     float32 `json:"alertIncidentRate"`
		AlertFeedback         float32 `json:"alertFeedbackCount"`
		ActionabilityFeedback float32 `json:"alertActionability"`
		AccuracyFeedback      float32 `json:"alertAccuracy"`
		DocumentationFeedback float32 `json:"alertDocumentation"`
	}

	OncallShiftMetricsAlertInstance struct {
		AlertId          uuid.UUID              `json:"alertId"`
		Timestamp        time.Time              `json:"timestamp"`
		ResponseMinutes  float32                `json:"responseMinutes"`
		LinkedIncidentID *uuid.UUID             `json:"linkedIncidentId,omitempty"`
		Feedback         *AlertFeedbackInstance `json:"feedback"`
	}
)

func OncallShiftMetricsFromEnt(m *ent.OncallShiftMetrics) OncallShiftMetrics {
	return OncallShiftMetrics{
		Burden: OncallShiftMetricsBurden{
			FinalScore:           m.BurdenScore,
			EventFrequency:       m.EventFrequency,
			LifeImpact:           m.LifeImpact,
			TimeImpact:           m.TimeImpact,
			ResponseRequirements: m.ResponseRequirements,
			Isolation:            m.Isolation,
		},
		Events: OncallShiftMetricsEvents{
			Total:                   m.EventsTotal,
			TotalInterrupts:         m.InterruptsTotal,
			InterruptsBusinessHours: m.InterruptsBusinessHours,
			InterruptsNight:         m.InterruptsNight,
			InterruptResponseTime:   0,
			TotalIncidents:          m.IncidentsTotal,
			IncidentTime:            m.IncidentResponseTime,
			TotalAlerts:             m.AlertsTotal,
			AlertIncidentRate:       0,
			AlertFeedback:           0,
			ActionabilityFeedback:   0,
			AccuracyFeedback:        0,
			DocumentationFeedback:   0,
		},
	}
}

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

var GetOncallShiftBurdenMetricWeights = huma.Operation{
	OperationID: "get-oncall-shift-burden-metric-weights",
	Method:      http.MethodGet,
	Path:        "/oncall_metrics/burden_weights",
	Summary:     "Get Weights for Calculating Burden",
	Tags:        oncallMetricsTags,
	Errors:      errorCodes(),
}

type GetOncallShiftBurdenMetricWeightsRequest EmptyRequest
type GetOncallShiftBurdenMetricWeightsResponse ItemResponse[OncallShiftBurdenMetricWeights]
