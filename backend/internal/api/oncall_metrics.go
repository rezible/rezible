package api

import (
	"context"
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

	metrics := oapi.OncallShiftMetrics{}
	resp.Body.Data = metrics

	return &resp, nil
}

func (h *oncallMetricsHandler) GetOncallShiftBurdenMetricWeights(context.Context, *oapi.GetOncallShiftBurdenMetricWeightsRequest) (*oapi.GetOncallShiftBurdenMetricWeightsResponse, error) {
	var resp oapi.GetOncallShiftBurdenMetricWeightsResponse

	return &resp, nil
}
