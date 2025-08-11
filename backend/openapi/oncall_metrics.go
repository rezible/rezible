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
		Burden     OncallShiftMetricsBurden     `json:"burden"`
		Incidents  OncallShiftMetricsIncidents  `json:"incidents"`
		Interrupts OncallShiftMetricsInterrupts `json:"interrupts"`
	}

	OncallShiftMetricsBurden struct {
		FinalScore           float32 `json:"finalScore"`
		EventFrequency       float32 `json:"eventFrequency"`
		LifeImpact           float32 `json:"lifeImpact"`
		TimeImpact           float32 `json:"timeImpact"`
		ResponseRequirements float32 `json:"responseRequirements"`
		Isolation            float32 `json:"isolation"`
	}

	OncallShiftMetricsIncidents struct {
		Total               float32 `json:"total"`
		ResponseTimeMinutes float32 `json:"responseTimeMinutes"`
	}

	OncallShiftMetricsInterrupts struct {
		Total                 float32 `json:"total"`
		TotalAlerts           float32 `json:"totalAlerts"`
		CountOffHours         float32 `json:"countOffHours"`
		CountNight            float32 `json:"countNight"`
		IncidentRate          float32 `json:"incidentRate"`
		TotalWithFeedback     float32 `json:"totalWithFeedback"`
		ActionabilityFeedback float32 `json:"actionabilityFeedback"`
		AccuracyFeedback      float32 `json:"accuracyFeedback"`
		DocumentationFeedback float32 `json:"documentationFeedback"`
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
	offHoursInterrupts := m.InterruptsTotal - m.InterruptsBusinessHours - m.InterruptsNight

	return OncallShiftMetrics{
		Burden: OncallShiftMetricsBurden{
			FinalScore:           m.BurdenScore,
			EventFrequency:       m.EventFrequency,
			LifeImpact:           m.LifeImpact,
			TimeImpact:           m.TimeImpact,
			ResponseRequirements: m.ResponseRequirements,
			Isolation:            m.Isolation,
		},
		Incidents: OncallShiftMetricsIncidents{
			Total:               m.IncidentsTotal,
			ResponseTimeMinutes: m.IncidentResponseTime,
		},
		Interrupts: OncallShiftMetricsInterrupts{
			Total:         m.InterruptsTotal,
			TotalAlerts:   m.InterruptsAlerts,
			CountOffHours: offHoursInterrupts,
			CountNight:    m.InterruptsNight,
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
