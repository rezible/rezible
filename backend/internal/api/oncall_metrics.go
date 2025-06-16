package api

import (
	"context"
	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallMetricsHandler struct {
}

func newOncallMetricsHandler() *oncallMetricsHandler {
	return &oncallMetricsHandler{}
}

func (h *oncallMetricsHandler) GetOncallRosterMetrics(ctx context.Context, request *oapi.GetOncallRosterMetricsRequest) (*oapi.GetOncallRosterMetricsResponse, error) {
	var resp oapi.GetOncallRosterMetricsResponse

	return &resp, nil
}

func (h *oncallMetricsHandler) GetOncallShiftMetrics(ctx context.Context, request *oapi.GetOncallShiftMetricsRequest) (*oapi.GetOncallShiftMetricsResponse, error) {
	var resp oapi.GetOncallShiftMetricsResponse

	shiftMetrics := oapi.OncallShiftMetrics{
		Burden: oapi.OncallShiftMetricsBurden{
			FinalScore:           6.4,
			EventFrequency:       4.8,
			LifeImpact:           7.8,
			TimeImpact:           6.4,
			ResponseRequirements: 7.2,
			Isolation:            4.4,
		},
		Incidents: oapi.OncallShiftMetricsIncidents{
			Total:               4,
			ResponseTimeMinutes: 168,
		},
		Alerts: oapi.OncallShiftMetricsAlerts{
			Total:                 24,
			CountOffHours:         5,
			CountNight:            3,
			IncidentRate:          .1,
			TotalWithFeedback:     15,
			ActionabilityFeedback: .4,
			AccuracyFeedback:      .6,
			DocumentationFeedback: .6,
		},
	}
	compMetrics := oapi.OncallShiftMetrics{
		Burden: oapi.OncallShiftMetricsBurden{
			FinalScore:           5.9,
			EventFrequency:       4.3,
			LifeImpact:           4.5,
			TimeImpact:           4.2,
			ResponseRequirements: 3.0,
			Isolation:            3.4,
		},
		Incidents: oapi.OncallShiftMetricsIncidents{
			Total:               1.1,
			ResponseTimeMinutes: 33,
		},
		Alerts: oapi.OncallShiftMetricsAlerts{
			Total:                 19,
			CountOffHours:         7,
			CountNight:            4,
			IncidentRate:          .1,
			TotalWithFeedback:     13,
			ActionabilityFeedback: .33,
			AccuracyFeedback:      .52,
			DocumentationFeedback: .7,
		},
	}
	if request.ShiftId == uuid.Nil {
		resp.Body.Data = compMetrics
	} else {
		resp.Body.Data = shiftMetrics
	}

	return &resp, nil
}

func (h *oncallMetricsHandler) GetOncallShiftBurdenMetricWeights(context.Context, *oapi.GetOncallShiftBurdenMetricWeightsRequest) (*oapi.GetOncallShiftBurdenMetricWeightsResponse, error) {
	var resp oapi.GetOncallShiftBurdenMetricWeightsResponse

	resp.Body.Data = oapi.OncallShiftBurdenMetricWeights{}

	return &resp, nil
}
