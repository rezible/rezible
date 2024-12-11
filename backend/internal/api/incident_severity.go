package api

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/incidentseverity"
	"github.com/twohundreds/rezible/ent/schema"
	oapi "github.com/twohundreds/rezible/openapi"
)

type incidentSeverityHandler struct {
	severities *ent.IncidentSeverityClient
}

func newIncidentSeverityHandler(severities *ent.IncidentSeverityClient) *incidentSeverityHandler {
	return &incidentSeverityHandler{severities}
}

func (h *incidentSeverityHandler) ListIncidentSeverities(ctx context.Context, request *oapi.ListIncidentSeveritiesRequest) (*oapi.ListIncidentSeveritiesResponse, error) {
	var resp oapi.ListIncidentSeveritiesResponse

	query := h.severities.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}
	if len(request.Search) > 0 {
		query = query.Where(incidentseverity.NameContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(incidentseverity.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, detailError("Failed to query incident severities", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentSeverity, len(res))
	for i, tag := range res {
		resp.Body.Data[i] = oapi.IncidentSeverityFromEnt(tag)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, detailError("Failed to query incident severity count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *incidentSeverityHandler) CreateIncidentSeverity(ctx context.Context, request *oapi.CreateIncidentSeverityRequest) (*oapi.CreateIncidentSeverityResponse, error) {
	var resp oapi.CreateIncidentSeverityResponse

	attr := request.Body.Attributes
	query := h.severities.Create().
		SetName(attr.Name)

	sev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("Failed to create incident severity", createErr)
	}
	resp.Body.Data = oapi.IncidentSeverityFromEnt(sev)

	return &resp, nil
}

func (h *incidentSeverityHandler) GetIncidentSeverity(ctx context.Context, request *oapi.GetIncidentSeverityRequest) (*oapi.GetIncidentSeverityResponse, error) {
	var resp oapi.GetIncidentSeverityResponse

	sev, queryErr := h.severities.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("Failed to get incident tag", queryErr)
	}
	resp.Body.Data = oapi.IncidentSeverityFromEnt(sev)

	return &resp, nil
}

func (h *incidentSeverityHandler) UpdateIncidentSeverity(ctx context.Context, request *oapi.UpdateIncidentSeverityRequest) (*oapi.UpdateIncidentSeverityResponse, error) {
	var resp oapi.UpdateIncidentSeverityResponse

	attr := request.Body.Attributes
	query := h.severities.UpdateOneID(request.Id).
		SetNillableName(attr.Name)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	sev, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, detailError("Failed to update incident severity", updateErr)
	}
	resp.Body.Data = oapi.IncidentSeverityFromEnt(sev)

	return &resp, nil
}

func (h *incidentSeverityHandler) ArchiveIncidentSeverity(ctx context.Context, request *oapi.ArchiveIncidentSeverityRequest) (*oapi.ArchiveIncidentSeverityResponse, error) {
	var resp oapi.ArchiveIncidentSeverityResponse

	delErr := h.severities.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, detailError("Failed to archive incident severity", delErr)
	}

	return &resp, nil
}
