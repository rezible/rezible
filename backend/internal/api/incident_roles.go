package api

import (
	"context"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema"
	oapi "github.com/rezible/rezible/openapi"
)

type incidentRolesHandler struct {
	roles *ent.IncidentRoleClient
}

func newincidentRolesHandler(roles *ent.IncidentRoleClient) *incidentRolesHandler {
	return &incidentRolesHandler{roles}
}

func (h *incidentRolesHandler) ListIncidentRoles(ctx context.Context, request *oapi.ListIncidentRolesRequest) (*oapi.ListIncidentRolesResponse, error) {
	var resp oapi.ListIncidentRolesResponse

	query := h.roles.Query()

	if true {
		ctx = schema.IncludeArchived(ctx)
	}

	res, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, apiError("Failed to query incident roles", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentRole, len(res))
	for i, role := range res {
		resp.Body.Data[i] = oapi.IncidentRoleFromEnt(role)
	}

	return &resp, nil
}

func (h *incidentRolesHandler) CreateIncidentRole(ctx context.Context, request *oapi.CreateIncidentRoleRequest) (*oapi.CreateIncidentRoleResponse, error) {
	var resp oapi.CreateIncidentRoleResponse

	attr := request.Body.Attributes

	query := h.roles.Create().
		SetName(attr.Name).
		SetRequired(attr.Required)

	role, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, apiError("Failed to create incident role", createErr)
	}
	resp.Body.Data = oapi.IncidentRoleFromEnt(role)

	return &resp, nil
}

func (h *incidentRolesHandler) GetIncidentRole(ctx context.Context, request *oapi.GetIncidentRoleRequest) (*oapi.GetIncidentRoleResponse, error) {
	var resp oapi.GetIncidentRoleResponse

	ctx = schema.IncludeArchived(ctx)
	role, queryErr := h.roles.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("Failed to get incident role", queryErr)
	}
	resp.Body.Data = oapi.IncidentRoleFromEnt(role)

	return &resp, nil
}

func (h *incidentRolesHandler) UpdateIncidentRole(ctx context.Context, request *oapi.UpdateIncidentRoleRequest) (*oapi.UpdateIncidentRoleResponse, error) {
	var resp oapi.UpdateIncidentRoleResponse

	attr := request.Body.Attributes
	query := h.roles.UpdateOneID(request.Id).
		SetNillableName(attr.Name).
		SetNillableRequired(attr.Required)

	if attr.Archived != nil && (*attr.Archived == false) {
		query.ClearArchiveTime()
	}

	role, saveErr := query.Save(ctx)
	if saveErr != nil {
		return nil, apiError("Failed to update incident role", saveErr)
	}
	resp.Body.Data = oapi.IncidentRoleFromEnt(role)

	return &resp, nil
}

func (h *incidentRolesHandler) ArchiveIncidentRole(ctx context.Context, request *oapi.ArchiveIncidentRoleRequest) (*oapi.ArchiveIncidentRoleResponse, error) {
	var resp oapi.ArchiveIncidentRoleResponse

	delErr := h.roles.DeleteOneID(request.Id).Exec(ctx)
	if delErr != nil {
		return nil, apiError("Failed to archive incident role", delErr)
	}

	return &resp, nil
}
