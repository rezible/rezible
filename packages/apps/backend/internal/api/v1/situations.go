package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type situationsHandler struct {
	situations rez.SituationService
	users      rez.UserService
}

func newSituationsHandler(situations rez.SituationService, users rez.UserService) *situationsHandler {
	return &situationsHandler{situations: situations, users: users}
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

func (h *situationsHandler) StartSituationInvestigation(ctx context.Context, request *oapi.StartSituationInvestigationRequest) (*oapi.StartSituationInvestigationResponse, error) {
	var response oapi.StartSituationInvestigationResponse
	params := rez.CreateSituationInvestigationParams{
		SituationID: request.Id,
		Query:       request.Body.Attributes.Query,
	}
	sitEnv, createErr := h.situations.CreateSituationInvestigation(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "start situation investigation", createErr)
	}
	response.Body.Data = oapi.SituationInvestigationFromEnt(sitEnv)
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
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, oapi.Error(ctx, "add situation hazard assessment", rez.ErrAuthSessionMissing)
	}

	attrs := request.Body.Attributes
	assessment, addErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    request.Id,
		SystemHazardID: attrs.SystemHazardId,
		Status:         attrs.Status,
		Summary:        attrs.Summary,
		UserID:         &userId,
	})
	if addErr != nil {
		return nil, oapi.Error(ctx, "failed to add situation hazard assessment", addErr)
	}
	response.Body.Data = oapi.SituationHazardAssessmentFromEnt(assessment)
	return &response, nil
}
