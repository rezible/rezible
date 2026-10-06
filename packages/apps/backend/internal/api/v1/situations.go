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
	params.Search = request.Search
	for _, stage := range request.Stage {
		params.Stages = append(params.Stages, rez.SituationStage(stage))
	}
	if request.Muted.IsSet {
		params.Muted = &request.Muted.Value
	}
	if !request.OpenedAfter.IsZero() {
		params.OpenedAfter = &request.OpenedAfter
	}
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

func (h *situationsHandler) RaiseSituation(ctx context.Context, request *oapi.RaiseSituationRequest) (*oapi.RaiseSituationResponse, error) {
	params := rez.RaiseSituationParams{StartInvestigation: true}
	raised, raiseErr := h.situations.RaiseSituation(ctx, request.Id, params)
	if raiseErr != nil {
		return nil, oapi.Error(ctx, "failed to raise situation", raiseErr)
	}
	var response oapi.RaiseSituationResponse
	response.Body.Data = oapi.SituationFromEnt(raised)
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

func (h *situationsHandler) SetSituationMute(ctx context.Context, request *oapi.SetSituationMuteRequest) (*oapi.SetSituationMuteResponse, error) {
	updated, changeErr := h.situations.SetSituationMute(ctx, request.Id, &rez.SituationMute{Reason: request.Body.Attributes.Reason})
	if changeErr != nil {
		return nil, oapi.Error(ctx, "set situation mute", changeErr)
	}
	var response oapi.SetSituationMuteResponse
	response.Body.Data = oapi.SituationFromEnt(updated)
	return &response, nil
}

func (h *situationsHandler) ClearSituationMute(ctx context.Context, request *oapi.ClearSituationMuteRequest) (*oapi.ClearSituationMuteResponse, error) {
	updated, changeErr := h.situations.SetSituationMute(ctx, request.Id, nil)
	if changeErr != nil {
		return nil, oapi.Error(ctx, "clear situation mute", changeErr)
	}
	var response oapi.ClearSituationMuteResponse
	response.Body.Data = oapi.SituationFromEnt(updated)
	return &response, nil
}

func (h *situationsHandler) SetSituationHold(ctx context.Context, request *oapi.SetSituationHoldRequest) (*oapi.SetSituationHoldResponse, error) {
	updated, changeErr := h.situations.SetSituationHold(ctx, request.Id, &rez.SituationHold{Until: request.Body.Attributes.Until})
	if changeErr != nil {
		return nil, oapi.Error(ctx, "set situation hold", changeErr)
	}
	var response oapi.SetSituationHoldResponse
	response.Body.Data = oapi.SituationFromEnt(updated)
	return &response, nil
}

func (h *situationsHandler) ClearSituationHold(ctx context.Context, request *oapi.ClearSituationHoldRequest) (*oapi.ClearSituationHoldResponse, error) {
	updated, changeErr := h.situations.SetSituationHold(ctx, request.Id, nil)
	if changeErr != nil {
		return nil, oapi.Error(ctx, "clear situation hold", changeErr)
	}
	var response oapi.ClearSituationHoldResponse
	response.Body.Data = oapi.SituationFromEnt(updated)
	return &response, nil
}

func (h *situationsHandler) CloseSituation(ctx context.Context, request *oapi.CloseSituationRequest) (*oapi.CloseSituationResponse, error) {
	params := rez.CloseSituationParams{Note: request.Body.Attributes.Note}
	updated, changeErr := h.situations.CloseSituation(ctx, request.Id, params)
	if changeErr != nil {
		return nil, oapi.Error(ctx, "close situation", changeErr)
	}
	var response oapi.CloseSituationResponse
	response.Body.Data = oapi.SituationFromEnt(updated)
	return &response, nil
}

func (h *situationsHandler) MergeSituation(ctx context.Context, request *oapi.MergeSituationRequest) (*oapi.MergeSituationResponse, error) {
	params := rez.MergeSituationsParams{
		SourceID:    request.Id,
		TargetID:    request.Body.Attributes.TargetId,
		Explanation: request.Body.Attributes.Explanation,
	}
	updated, changeErr := h.situations.MergeSituations(ctx, params)
	if changeErr != nil {
		return nil, oapi.Error(ctx, "merge situation", changeErr)
	}
	var response oapi.MergeSituationResponse
	response.Body.Data = oapi.SituationFromEnt(updated)
	return &response, nil
}

func (h *situationsHandler) ListSituationJudgments(ctx context.Context, request *oapi.ListSituationJudgmentsRequest) (*oapi.ListSituationJudgmentsResponse, error) {
	params := rez.ListSituationJudgmentsParams{ListParams: request.ListParams(), SituationID: request.Id}
	result, listErr := h.situations.ListSituationJudgments(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list situation judgments", listErr)
	}
	return &oapi.ListSituationJudgmentsResponse{Body: oapi.ConvertPaginatedResultBody(result, oapi.SituationJudgmentHistoryItemFromEnt)}, nil
}
