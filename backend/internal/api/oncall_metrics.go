package api

import (
	"context"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"

	oapi "github.com/rezible/rezible/openapi"
)

type oncallMetricsHandler struct {
	oncall rez.OncallService
}

func newOncallMetricsHandler(oncall rez.OncallService) *oncallMetricsHandler {
	return &oncallMetricsHandler{oncall: oncall}
}

func (h *oncallMetricsHandler) GetOncallRosterMetrics(ctx context.Context, request *oapi.GetOncallRosterMetricsRequest) (*oapi.GetOncallRosterMetricsResponse, error) {
	var resp oapi.GetOncallRosterMetricsResponse

	return &resp, nil
}

func (h *oncallMetricsHandler) GetOncallShiftMetrics(ctx context.Context, request *oapi.GetOncallShiftMetricsRequest) (*oapi.GetOncallShiftMetricsResponse, error) {
	var resp oapi.GetOncallShiftMetricsResponse

	var metrics *ent.OncallShiftMetrics
	var metricsErr error
	if request.ShiftId != uuid.Nil {
		metrics, metricsErr = h.oncall.GetShiftMetrics(ctx, request.ShiftId)
	} else {
		// TODO: include in request
		from := time.Now().Add(-time.Hour)
		to := time.Now().Add(time.Hour)
		metrics, metricsErr = h.oncall.GetComparisonShiftMetrics(ctx, from, to)
	}
	if metricsErr != nil {
		return nil, detailError("failed to get oncall shift metrics", metricsErr)
	}
	resp.Body.Data = oapi.OncallShiftMetricsFromEnt(metrics)

	return &resp, nil
}

func (h *oncallMetricsHandler) GetOncallShiftBurdenMetricWeights(context.Context, *oapi.GetOncallShiftBurdenMetricWeightsRequest) (*oapi.GetOncallShiftBurdenMetricWeightsResponse, error) {
	var resp oapi.GetOncallShiftBurdenMetricWeightsResponse

	resp.Body.Data = oapi.OncallShiftBurdenMetricWeights{}

	return &resp, nil
}
