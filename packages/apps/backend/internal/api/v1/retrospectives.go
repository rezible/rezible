package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
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
	return nil, oapi.Error(ctx, "retrospective metadata updates are not implemented", rez.ErrNotImplemented)
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	var resp oapi.GetRetrospectiveResponse

	retro, retroErr := h.retros.Get(ctx, retrospective.ID(input.Id))
	if retroErr != nil {
		return nil, oapi.Error(ctx, "failed to get retrospective", retroErr)
	}
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)

	return &resp, nil
}
