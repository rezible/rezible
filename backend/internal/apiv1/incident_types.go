package apiv1

import (
	"context"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidenttype"
	"github.com/rezible/rezible/ent/schema"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type incidentTypesHandler struct {
	types *ent.IncidentTypeClient
}

func newIncidentTypesHandler(types *ent.IncidentTypeClient) *incidentTypesHandler {
	return &incidentTypesHandler{types}
}

func (h *incidentTypesHandler) ListIncidentTypes(ctx context.Context, request *oapi.ListIncidentTypesRequest) (*oapi.ListIncidentTypesResponse, error) {
	var resp oapi.ListIncidentTypesResponse

	query := h.types.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}
	if len(request.Search) > 0 {
		query = query.Where(incidenttype.NameContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(incidenttype.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, apiError("Failed to query incident types", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentType, len(res))
	for i, t := range res {
		resp.Body.Data[i] = oapi.IncidentTypeFromEnt(t)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, apiError("Failed to query incident type count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *incidentTypesHandler) CreateIncidentType(ctx context.Context, request *oapi.CreateIncidentTypeRequest) (*oapi.CreateIncidentTypeResponse, error) {
	var resp oapi.CreateIncidentTypeResponse

	attr := request.Body.Attributes
	query := h.types.Create().SetName(attr.Name)
	t, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, apiError("Failed to create incident type", createErr)
	}
	resp.Body.Data = oapi.IncidentTypeFromEnt(t)

	return &resp, nil
}

func (h *incidentTypesHandler) GetIncidentType(ctx context.Context, request *oapi.GetIncidentTypeRequest) (*oapi.GetIncidentTypeResponse, error) {
	var resp oapi.GetIncidentTypeResponse

	t, queryErr := h.types.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("Failed to get incident type", queryErr)
	}
	resp.Body.Data = oapi.IncidentTypeFromEnt(t)

	return &resp, nil
}

func (h *incidentTypesHandler) UpdateIncidentType(ctx context.Context, request *oapi.UpdateIncidentTypeRequest) (*oapi.UpdateIncidentTypeResponse, error) {
	var resp oapi.UpdateIncidentTypeResponse

	attr := request.Body.Attributes
	query := h.types.UpdateOneID(request.Id).
		SetNillableName(attr.Name)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	t, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, apiError("Failed to update incident type", updateErr)
	}
	resp.Body.Data = oapi.IncidentTypeFromEnt(t)

	return &resp, nil
}

func (h *incidentTypesHandler) ArchiveIncidentType(ctx context.Context, request *oapi.ArchiveIncidentTypeRequest) (*oapi.ArchiveIncidentTypeResponse, error) {
	var resp oapi.ArchiveIncidentTypeResponse

	delErr := h.types.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, apiError("Failed to archive incident type", delErr)
	}

	return &resp, nil
}
