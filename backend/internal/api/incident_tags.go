package api

import (
	"context"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidenttag"
	"github.com/rezible/rezible/ent/schema"
	oapi "github.com/rezible/rezible/openapi"
)

type incidentTagsHandler struct {
	tags *ent.IncidentTagClient
}

func newIncidentTagsHandler(tags *ent.IncidentTagClient) *incidentTagsHandler {
	return &incidentTagsHandler{tags}
}

func (h *incidentTagsHandler) ListIncidentTags(ctx context.Context, request *oapi.ListIncidentTagsRequest) (*oapi.ListIncidentTagsResponse, error) {
	var resp oapi.ListIncidentTagsResponse

	query := h.tags.Query()
	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}
	if len(request.Search) > 0 {
		query = query.Where(incidenttag.ValueContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(incidenttag.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, detailError("Failed to query incident tags", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentTag, len(res))
	for i, tag := range res {
		resp.Body.Data[i] = oapi.IncidentTagFromEnt(tag)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, detailError("Failed to query incident tag count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *incidentTagsHandler) CreateIncidentTag(ctx context.Context, request *oapi.CreateIncidentTagRequest) (*oapi.CreateIncidentTagResponse, error) {
	var resp oapi.CreateIncidentTagResponse

	attr := request.Body.Attributes
	query := h.tags.Create().SetValue(attr.Value)
	tag, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("Failed to create incident tag", createErr)
	}
	resp.Body.Data = oapi.IncidentTagFromEnt(tag)

	return &resp, nil
}

func (h *incidentTagsHandler) GetIncidentTag(ctx context.Context, request *oapi.GetIncidentTagRequest) (*oapi.GetIncidentTagResponse, error) {
	var resp oapi.GetIncidentTagResponse

	tag, queryErr := h.tags.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("Failed to get incident tag", queryErr)
	}
	resp.Body.Data = oapi.IncidentTagFromEnt(tag)

	return &resp, nil
}

func (h *incidentTagsHandler) UpdateIncidentTag(ctx context.Context, request *oapi.UpdateIncidentTagRequest) (*oapi.UpdateIncidentTagResponse, error) {
	var resp oapi.UpdateIncidentTagResponse

	attr := request.Body.Attributes
	query := h.tags.UpdateOneID(request.Id).
		SetNillableValue(attr.Value)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	tag, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, detailError("Failed to update incident tag", updateErr)
	}
	resp.Body.Data = oapi.IncidentTagFromEnt(tag)

	return &resp, nil
}

func (h *incidentTagsHandler) ArchiveIncidentTag(ctx context.Context, request *oapi.ArchiveIncidentTagRequest) (*oapi.ArchiveIncidentTagResponse, error) {
	var resp oapi.ArchiveIncidentTagResponse

	delErr := h.tags.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, detailError("Failed to archive incident tag", delErr)
	}

	return &resp, nil
}
