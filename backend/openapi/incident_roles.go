package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"
)

type IncidentRolesHandler interface {
	ListIncidentRoles(context.Context, *ListIncidentRolesRequest) (*ListIncidentRolesResponse, error)
	CreateIncidentRole(context.Context, *CreateIncidentRoleRequest) (*CreateIncidentRoleResponse, error)
	GetIncidentRole(context.Context, *GetIncidentRoleRequest) (*GetIncidentRoleResponse, error)
	UpdateIncidentRole(context.Context, *UpdateIncidentRoleRequest) (*UpdateIncidentRoleResponse, error)
	ArchiveIncidentRole(context.Context, *ArchiveIncidentRoleRequest) (*ArchiveIncidentRoleResponse, error)
}

func (o operations) RegisterIncidentRoles(api huma.API) {
	huma.Register(api, ListIncidentRoles, o.ListIncidentRoles)
	huma.Register(api, CreateIncidentRole, o.CreateIncidentRole)
	huma.Register(api, GetIncidentRole, o.GetIncidentRole)
	huma.Register(api, UpdateIncidentRole, o.UpdateIncidentRole)
	huma.Register(api, ArchiveIncidentRole, o.ArchiveIncidentRole)
}

type (
	IncidentRole struct {
		Id         uuid.UUID              `json:"id"`
		Attributes IncidentRoleAttributes `json:"attributes"`
	}

	IncidentRoleAttributes struct {
		Name        string `json:"name"`
		Archived    bool   `json:"archived"`
		Required    bool   `json:"required"`
		Description string `json:"description"`
	}
)

func IncidentRoleFromEnt(role *ent.IncidentRole) IncidentRole {
	return IncidentRole{
		Id: role.ID,
		Attributes: IncidentRoleAttributes{
			Name:        role.Name,
			Archived:    !role.ArchiveTime.IsZero(),
			Required:    role.Required,
			Description: "",
		},
	}
}

var incidentRolesTags = []string{"Incident Roles"}

// ops

var ListIncidentRoles = huma.Operation{
	OperationID: "list-incident-roles",
	Method:      http.MethodGet,
	Path:        "/incident_roles",
	Summary:     "List Incident Roles",
	Tags:        incidentRolesTags,
	Errors:      errorCodes(),
}

type ListIncidentRolesRequest ListRequest
type ListIncidentRolesResponse PaginatedResponse[IncidentRole]

var GetIncidentRole = huma.Operation{
	OperationID: "get-incident-role",
	Method:      http.MethodGet,
	Path:        "/incident_roles/{id}",
	Summary:     "Get an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      errorCodes(),
}

type GetIncidentRoleRequest GetIdRequest
type GetIncidentRoleResponse ItemResponse[IncidentRole]

var CreateIncidentRole = huma.Operation{
	OperationID: "create-incident-role",
	Method:      http.MethodPost,
	Path:        "/incident_roles",
	Summary:     "Create an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      errorCodes(),
}

type CreateIncidentRoleAttributes struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}
type CreateIncidentRoleRequest RequestWithBodyAttributes[CreateIncidentRoleAttributes]
type CreateIncidentRoleResponse ItemResponse[IncidentRole]

var UpdateIncidentRole = huma.Operation{
	OperationID: "update-incident-role",
	Method:      http.MethodPatch,
	Path:        "/incident_roles/{id}",
	Summary:     "Update an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      errorCodes(),
}

type UpdateIncidentRoleAttributes struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Required    *bool   `json:"required,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
}
type UpdateIncidentRoleRequest UpdateIdRequest[UpdateIncidentRoleAttributes]
type UpdateIncidentRoleResponse ItemResponse[IncidentRole]

var ArchiveIncidentRole = huma.Operation{
	OperationID: "archive-incident-role",
	Method:      http.MethodDelete,
	Path:        "/incident_roles/{id}",
	Summary:     "Archive an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentRoleRequest ArchiveIdRequest
type ArchiveIncidentRoleResponse EmptyResponse
