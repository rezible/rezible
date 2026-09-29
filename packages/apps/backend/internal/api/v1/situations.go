package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type situationsHandler struct {
	*baseHandler
	situations rez.SituationService
}

func newSituationsHandler(bh *baseHandler, situations rez.SituationService) *situationsHandler {
	return &situationsHandler{bh, situations}
}

func (h *situationsHandler) ListSituations(ctx context.Context, request *oapi.ListSituationsRequest) (*oapi.ListSituationsResponse, error) {
	params := rez.ListSituationsParams{ListParams: request.ListParams()}
	if request.Status != "" {
		if request.Status == "active" {
			params.Active = new(true)
		} else if request.Status == "investigating" {
			params.HasInvestigation = new(true)
		} else if request.Status == "closed" {
			params.Active = new(false)
		}
	}
	if !request.OpenedAfter.IsZero() {
		params.OpenedAfter = &request.OpenedAfter
	}
	params.Search = request.Search
	result, listErr := h.situations.ListSituations(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list situations", listErr)
	}
	return &oapi.ListSituationsResponse{Body: oapi.ConvertPaginatedResultBody(result, oapi.SituationFromEnt)}, nil
}

func (h *situationsHandler) GetSituation(ctx context.Context, request *oapi.GetSituationRequest) (*oapi.GetSituationResponse, error) {
	var response oapi.GetSituationResponse
	s, getErr := h.situations.GetSituation(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "failed to get situation", getErr)
	}
	response.Body.Data = oapi.SituationFromEnt(s)
	return &response, nil
}

func (h *situationsHandler) RequestSituationInvestigation(ctx context.Context, request *oapi.RequestSituationInvestigationRequest) (*oapi.RequestSituationInvestigationResponse, error) {
	_, invErr := h.situations.RequestSituationInvestigation(ctx, request.Id)
	if invErr != nil {
		return nil, oapi.Error(ctx, "failed to request situation investigation", invErr)
	}
	situ, situErr := h.situations.GetSituation(ctx, request.Id)
	if situErr != nil {
		return nil, oapi.Error(ctx, "failed to get situation", situErr)
	}
	var response oapi.RequestSituationInvestigationResponse
	response.Body.Data = oapi.SituationFromEnt(situ)
	return &response, nil
}

func (h *situationsHandler) ListSituationHazardAssessments(ctx context.Context, request *oapi.ListSituationHazardAssessmentsRequest) (*oapi.ListSituationHazardAssessmentsResponse, error) {
	var response oapi.ListSituationHazardAssessmentsResponse
	result, listErr := h.situations.ListSituationHazardAssessments(ctx, rez.ListSituationHazardAssessmentsParams{
		ListParams:  request.ListParams(),
		SituationID: request.Id,
	})
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list situation hazard assessments", listErr)
	}
	response.Body = oapi.ConvertPaginatedResultBody(result, oapi.SituationHazardAssessmentFromEnt)
	return &response, nil
}

func (h *situationsHandler) AddSituationHazardAssessment(ctx context.Context, request *oapi.AddSituationHazardAssessmentRequest) (*oapi.AddSituationHazardAssessmentResponse, error) {
	var response oapi.AddSituationHazardAssessmentResponse

	attrs := request.Body.Attributes
	assessment, addErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    request.Id,
		SystemHazardID: attrs.SystemHazardId,
		Status:         attrs.Status,
		Summary:        attrs.Summary,
		UserID:         new(h.mustUserID(ctx)),
	})
	if addErr != nil {
		return nil, oapi.Error(ctx, "failed to add situation hazard assessment", addErr)
	}
	response.Body.Data = oapi.SituationHazardAssessmentFromEnt(assessment)
	return &response, nil
}
