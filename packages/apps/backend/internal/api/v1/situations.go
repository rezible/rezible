package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/user"
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
	var response oapi.ListSituationsResponse
	params := rez.ListSituationsParams{
		ListParams: request.ListParams(),
		Status:     request.Status,
	}
	if !request.OpenedAfter.IsZero() {
		params.OpenedAfter = &request.OpenedAfter
	}
	params.Search = request.Search
	result, listErr := h.situations.ListSituations(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list situations", listErr)
	}
	response.Body = oapi.ConvertPaginatedResultBody(result, oapi.SituationFromEnt)
	return &response, nil
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

	exec := execution.GetContext(ctx)
	userId, userOk := exec.UserID()
	if !userOk {
		return nil, rez.ErrAuthSessionMissing
	}
	user, userErr := h.users.Get(ctx, user.ID(userId))
	if userErr != nil {
		return nil, oapi.Error(ctx, "failed to get user", userErr)
	}

	attrs := request.Body.Attributes
	assessment, addErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    request.Id,
		SystemHazardID: attrs.SystemHazardId,
		Status:         attrs.Status,
		Summary:        attrs.Summary,
		UserID:         &user.ID,
	})
	if addErr != nil {
		return nil, oapi.Error(ctx, "failed to add situation hazard assessment", addErr)
	}
	response.Body.Data = oapi.SituationHazardAssessmentFromEnt(assessment)
	return &response, nil
}
