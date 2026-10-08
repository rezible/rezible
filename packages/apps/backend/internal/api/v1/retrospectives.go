package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/pkg/errs"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type retrospectivesHandler struct {
	users     rez.UserService
	incidents rez.IncidentService
	retros    rez.RetrospectiveService
	documents rez.DocumentsService
}

func newRetrospectivesHandler(users rez.UserService, incidents rez.IncidentService, retros rez.RetrospectiveService, documents rez.DocumentsService) *retrospectivesHandler {
	return &retrospectivesHandler{users, incidents, retros, documents}
}

func (h *retrospectivesHandler) ListRetrospectives(ctx context.Context, input *oapi.ListRetrospectivesRequest) (*oapi.ListRetrospectivesResponse, error) {
	var resp oapi.ListRetrospectivesResponse

	results := &ent.ListResult[ent.Retrospective]{
		Data:     make(ent.Retrospectives, 0),
		Page:     input.Page,
		PageSize: input.PageSize,
	}
	resp.Body = oapi.ConvertPaginatedResultBody(results, oapi.RetrospectiveFromEnt)

	return &resp, nil
}

func (h *retrospectivesHandler) UpdateRetrospective(ctx context.Context, req *oapi.UpdateRetrospectiveRequest) (*oapi.UpdateRetrospectiveResponse, error) {
	return nil, oapi.Error(ctx, "retrospective metadata updates are not implemented", errs.ErrNotImplemented)
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	retro, retroErr := h.retros.Get(ctx, retrospective.ID(input.Id))
	if retroErr != nil {
		return nil, oapi.Error(ctx, "failed to get retrospective", retroErr)
	}
	var resp oapi.GetRetrospectiveResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) RequestRetrospectiveReview(ctx context.Context, input *oapi.RequestRetrospectiveReviewRequest) (*oapi.RequestRetrospectiveReviewResponse, error) {
	return nil, oapi.Error(ctx, "not implemented", errs.ErrNotImplemented)
}

func (h *retrospectivesHandler) GetRetrospectiveReportComposition(ctx context.Context, input *oapi.GetRetrospectiveReportCompositionRequest) (*oapi.GetRetrospectiveReportCompositionResponse, error) {
	composition, queryErr := h.retros.GetReportComposition(ctx, input.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get report composition", queryErr)
	}
	var resp oapi.GetRetrospectiveReportCompositionResponse
	resp.Body.Data = oapi.RetrospectiveReportCompositionFromRez(composition)
	return &resp, nil
}

func (h *retrospectivesHandler) SetRetrospectiveReportFindingSelection(ctx context.Context, input *oapi.SetRetrospectiveReportFindingSelectionRequest) (*oapi.SetRetrospectiveReportFindingSelectionResponse, error) {
	return nil, oapi.Error(ctx, "not implemented", errs.ErrNotImplemented)
}

func (h *retrospectivesHandler) CreateIncidentRetrospective(ctx context.Context, request *oapi.CreateIncidentRetrospectiveRequest) (*oapi.CreateIncidentRetrospectiveResponse, error) {
	retro, createErr := h.retros.CreateForIncident(ctx, request.Id)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to start retrospective", createErr)
	}
	var response oapi.CreateIncidentRetrospectiveResponse
	response.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &response, nil
}
