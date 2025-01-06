package openapi

import (
	"context"
	"github.com/rezible/rezible/ent"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type SystemComponentsHandler interface {
	ListSystemComponents(context.Context, *ListSystemComponentsRequest) (*ListSystemComponentsResponse, error)
	CreateSystemComponent(context.Context, *CreateSystemComponentRequest) (*CreateSystemComponentResponse, error)
	GetSystemComponent(context.Context, *GetSystemComponentRequest) (*GetSystemComponentResponse, error)
	UpdateSystemComponent(context.Context, *UpdateSystemComponentRequest) (*UpdateSystemComponentResponse, error)
	ArchiveSystemComponent(context.Context, *ArchiveSystemComponentRequest) (*ArchiveSystemComponentResponse, error)

	ListIncidentSystemComponents(context.Context, *ListIncidentSystemComponentsRequest) (*ListIncidentSystemComponentsResponse, error)
}

func (o operations) RegisterSystemComponents(api huma.API) {
	huma.Register(api, ListSystemComponents, o.ListSystemComponents)
	huma.Register(api, CreateSystemComponent, o.CreateSystemComponent)
	huma.Register(api, GetSystemComponent, o.GetSystemComponent)
	huma.Register(api, UpdateSystemComponent, o.UpdateSystemComponent)
	huma.Register(api, ArchiveSystemComponent, o.ArchiveSystemComponent)

	huma.Register(api, ListIncidentSystemComponents, o.ListIncidentSystemComponents)
}

type (
	SystemComponent struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes SystemComponentAttributes `json:"attributes"`
	}

	SystemComponentAttributes struct {
		Name string `json:"name"`
	}

	IncidentSystemComponent struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes IncidentSystemComponentAttributes `json:"attributes"`
	}

	IncidentSystemComponentAttributes struct {
		Role string `json:"role"`
	}
)

func SystemComponentFromEnt(sc *ent.SystemComponent) SystemComponent {
	return SystemComponent{
		Id: sc.ID,
		Attributes: SystemComponentAttributes{
			Name: sc.Name,
		},
	}
}

func IncidentSystemComponentFromEnt(isc *ent.IncidentSystemComponent) IncidentSystemComponent {
	return IncidentSystemComponent{
		Id: isc.ID,
		Attributes: IncidentSystemComponentAttributes{
			Role: isc.Role.String(),
		},
	}
}

var systemComponentsTags = []string{"System Components"}

// Operations

var ListSystemComponents = huma.Operation{
	OperationID: "list-system-components",
	Method:      http.MethodGet,
	Path:        "/system_components",
	Summary:     "List System Components",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ListSystemComponentsRequest struct {
	ListRequest
}
type ListSystemComponentsResponse PaginatedResponse[SystemComponent]

var ListIncidentSystemComponents = huma.Operation{
	OperationID: "list-incident-system-components",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/components",
	Summary:     "List System Components for Incident",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ListIncidentSystemComponentsRequest ListIdRequest
type ListIncidentSystemComponentsResponse PaginatedResponse[IncidentSystemComponent]

var CreateSystemComponent = huma.Operation{
	OperationID: "create-system-component",
	Method:      http.MethodPost,
	Path:        "/system_components",
	Summary:     "Create a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type CreateSystemComponentAttributes struct {
	Name string `json:"name"`
}
type CreateSystemComponentRequest RequestWithBodyAttributes[CreateSystemComponentAttributes]
type CreateSystemComponentResponse ItemResponse[SystemComponent]

var GetSystemComponent = huma.Operation{
	OperationID: "get-system-component",
	Method:      http.MethodGet,
	Path:        "/system_components/{id}",
	Summary:     "Get a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type GetSystemComponentRequest GetFlexibleIdRequest
type GetSystemComponentResponse ItemResponse[SystemComponent]

var UpdateSystemComponent = huma.Operation{
	OperationID: "update-system-component",
	Method:      http.MethodPatch,
	Path:        "/system_components/{id}",
	Summary:     "Update a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type UpdateSystemComponentAttributes struct {
}
type UpdateSystemComponentRequest UpdateIdRequest[UpdateSystemComponentAttributes]
type UpdateSystemComponentResponse ItemResponse[SystemComponent]

var ArchiveSystemComponent = huma.Operation{
	OperationID: "archive-system-component",
	Method:      http.MethodDelete,
	Path:        "/system_components/{id}",
	Summary:     "Archive a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentRequest ArchiveIdRequest
type ArchiveSystemComponentResponse EmptyResponse
