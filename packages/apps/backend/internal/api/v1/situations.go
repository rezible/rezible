package apiv1

import (
	"context"
	"sort"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/situationinvestigation"
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
	response.Body.Data = make([]oapi.Situation, 0, len(result.Data))
	for _, value := range result.Data {
		converted, conversionErr := oapi.SituationFromEnt(value)
		if conversionErr != nil {
			return nil, oapi.Error(ctx, "convert situation", conversionErr)
		}
		response.Body.Data = append(response.Body.Data, converted)
	}
	response.Body.Pagination = oapi.Pagination{Page: result.Page, PageSize: result.PageSize, Total: result.Total}
	return &response, nil
}

func (h *situationsHandler) GetSituation(ctx context.Context, request *oapi.GetSituationRequest) (*oapi.GetSituationResponse, error) {
	var response oapi.GetSituationResponse
	s, getErr := h.situations.GetSituation(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "failed to get situation", getErr)
	}
	converted, conversionErr := oapi.SituationFromEnt(s)
	if conversionErr != nil {
		return nil, oapi.Error(ctx, "convert situation", conversionErr)
	}
	response.Body.Data = converted
	return &response, nil
}

func (h *situationsHandler) ListSituationInvestigations(ctx context.Context, request *oapi.ListSituationInvestigationsRequest) (*oapi.ListSituationInvestigationsResponse, error) {
	var response oapi.ListSituationInvestigationsResponse
	situation, getErr := h.situations.GetSituation(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get situation", getErr)
	}
	result := make([]oapi.SituationInvestigation, 0, len(situation.Edges.Investigations))
	sort.Slice(situation.Edges.Investigations, func(i, j int) bool {
		return situation.Edges.Investigations[i].ID.String() < situation.Edges.Investigations[j].ID.String()
	})
	for _, investigation := range situation.Edges.Investigations {
		converted, conversionErr := oapi.SituationInvestigationFromEnt(investigation, situation.EvidenceRevision)
		if conversionErr != nil {
			return nil, oapi.Error(ctx, "convert investigation", conversionErr)
		}
		result = append(result, *converted)
	}
	page, pageSize := request.Page, request.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	start := (page - 1) * pageSize
	if start > len(result) {
		start = len(result)
	}
	end := start + pageSize
	if end > len(result) {
		end = len(result)
	}
	response.Body.Data = result[start:end]
	response.Body.Pagination = oapi.Pagination{Page: page, PageSize: pageSize, Total: len(result)}
	return &response, nil
}

func (h *situationsHandler) GetSituationInvestigation(ctx context.Context, request *oapi.GetSituationInvestigationRequest) (*oapi.GetSituationInvestigationResponse, error) {
	var response oapi.GetSituationInvestigationResponse
	investigation, getErr := h.situations.LookupSituationInvestigation(ctx, situationinvestigation.ID(request.Id))
	if getErr != nil {
		return nil, oapi.Error(ctx, "get situation investigation", getErr)
	}
	situation, situationErr := h.situations.GetSituation(ctx, investigation.SituationID)
	if situationErr != nil {
		return nil, oapi.Error(ctx, "get investigation situation", situationErr)
	}
	converted, conversionErr := oapi.SituationInvestigationFromEnt(investigation, situation.EvidenceRevision)
	if conversionErr != nil {
		return nil, oapi.Error(ctx, "convert investigation", conversionErr)
	}
	response.Body.Data = *converted
	return &response, nil
}

func (h *situationsHandler) StartSituationInvestigation(ctx context.Context, request *oapi.StartSituationInvestigationRequest) (*oapi.StartSituationInvestigationResponse, error) {
	var response oapi.StartSituationInvestigationResponse
	params := rez.CreateSituationInvestigationParams{SituationID: request.Id, Query: request.Body.Attributes.Query}
	investigation, createErr := h.situations.CreateSituationInvestigation(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "start situation investigation", createErr)
	}
	investigation, lookupErr := h.situations.LookupSituationInvestigation(ctx, situationinvestigation.ID(investigation.ID))
	if lookupErr != nil {
		return nil, oapi.Error(ctx, "load started investigation", lookupErr)
	}
	situation, situationErr := h.situations.GetSituation(ctx, request.Id)
	if situationErr != nil {
		return nil, oapi.Error(ctx, "get investigation situation", situationErr)
	}
	converted, conversionErr := oapi.SituationInvestigationFromEnt(investigation, situation.EvidenceRevision)
	if conversionErr != nil {
		return nil, oapi.Error(ctx, "convert investigation", conversionErr)
	}
	response.Body.Data = *converted
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
		return nil, oapi.Error(ctx, "add situation hazard assessment", rez.ErrAuthSessionMissing)
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
