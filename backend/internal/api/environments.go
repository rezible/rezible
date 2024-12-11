package api

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/environment"
	"github.com/twohundreds/rezible/ent/schema"
	oapi "github.com/twohundreds/rezible/openapi"
)

type environmentsHandler struct {
	envs *ent.EnvironmentClient
}

func newEnvironmentsHandler(envs *ent.EnvironmentClient) *environmentsHandler {
	return &environmentsHandler{envs}
}

func (h *environmentsHandler) ListEnvironments(ctx context.Context, request *oapi.ListEnvironmentsRequest) (*oapi.ListEnvironmentsResponse, error) {
	var resp oapi.ListEnvironmentsResponse

	query := h.envs.Query()

	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}

	if len(request.Search) > 0 {
		query = query.Where(environment.NameContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(environment.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, detailError("Failed to query environments", queryErr)
	}

	resp.Body.Data = make([]oapi.Environment, len(res))
	for i, env := range res {
		resp.Body.Data[i] = oapi.EnvironmentFromEnt(env)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, detailError("Failed to query environments count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *environmentsHandler) CreateEnvironment(ctx context.Context, request *oapi.CreateEnvironmentRequest) (*oapi.CreateEnvironmentResponse, error) {
	var resp oapi.CreateEnvironmentResponse

	attr := request.Body.Attributes
	query := h.envs.Create().SetName(attr.Name)
	env, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("Failed to create Environment", createErr)
	}
	resp.Body.Data = oapi.EnvironmentFromEnt(env)

	return &resp, nil
}

func (h *environmentsHandler) GetEnvironment(ctx context.Context, request *oapi.GetEnvironmentRequest) (*oapi.GetEnvironmentResponse, error) {
	var resp oapi.GetEnvironmentResponse

	env, queryErr := h.envs.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("Failed to query Environment", queryErr)
	}
	resp.Body.Data = oapi.EnvironmentFromEnt(env)

	return &resp, nil
}

func (h *environmentsHandler) UpdateEnvironment(ctx context.Context, request *oapi.UpdateEnvironmentRequest) (*oapi.UpdateEnvironmentResponse, error) {
	var resp oapi.UpdateEnvironmentResponse

	attr := request.Body.Attributes
	query := h.envs.UpdateOneID(request.Id).
		SetNillableName(attr.Name)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	env, updateErr := query.Save(ctx)
	if updateErr != nil {
		return nil, detailError("Failed to update Environment", updateErr)
	}
	resp.Body.Data = oapi.EnvironmentFromEnt(env)

	return &resp, nil
}

func (h *environmentsHandler) ArchiveEnvironment(ctx context.Context, request *oapi.ArchiveEnvironmentRequest) (*oapi.ArchiveEnvironmentResponse, error) {
	var resp oapi.ArchiveEnvironmentResponse

	delErr := h.envs.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, detailError("Failed to delete environment", delErr)
	}

	return &resp, nil
}
