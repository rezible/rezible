package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"

	"github.com/rezible/rezible/ent"
)

type IncidentTypesHandler interface {
	ListIncidentTypes(context.Context, *ListIncidentTypesRequest) (*ListIncidentTypesResponse, error)
	CreateIncidentType(context.Context, *CreateIncidentTypeRequest) (*CreateIncidentTypeResponse, error)
	GetIncidentType(context.Context, *GetIncidentTypeRequest) (*GetIncidentTypeResponse, error)
	UpdateIncidentType(context.Context, *UpdateIncidentTypeRequest) (*UpdateIncidentTypeResponse, error)
	ArchiveIncidentType(context.Context, *ArchiveIncidentTypeRequest) (*ArchiveIncidentTypeResponse, error)
}

func (o operations) RegisterIncidentTypes(api huma.API) {
	huma.Register(api, ListIncidentTypes, o.ListIncidentTypes)
	huma.Register(api, CreateIncidentType, o.CreateIncidentType)
	huma.Register(api, GetIncidentType, o.GetIncidentType)
	huma.Register(api, UpdateIncidentType, o.UpdateIncidentType)
	huma.Register(api, ArchiveIncidentType, o.ArchiveIncidentType)
}

type (
	IncidentType struct {
		Id         uuid.UUID              `json:"id"`
		Attributes IncidentTypeAttributes `json:"attributes"`
	}

	IncidentTypeAttributes struct {
		Name        string `json:"name"`
		Archived    bool   `json:"archived"`
		Description string `json:"description"`
	}
)

func IncidentTypeFromEnt(t *ent.IncidentType) IncidentType {
	return IncidentType{
		Id: t.ID,
		Attributes: IncidentTypeAttributes{
			Name:     t.Name,
			Archived: !t.ArchiveTime.IsZero(),
		},
	}
}

var incidentTypesTags = []string{"Incident Types"}

// ops

var ListIncidentTypes = huma.Operation{
	OperationID: "list-incident-types",
	Method:      http.MethodGet,
	Path:        "/incident_types",
	Summary:     "List Types",
	Tags:        incidentTypesTags,
	Errors:      errorCodes(),
}

type ListIncidentTypesRequest ListRequest
type ListIncidentTypesResponse PaginatedResponse[IncidentType]

var GetIncidentType = huma.Operation{
	OperationID: "get-incident-type",
	Method:      http.MethodGet,
	Path:        "/incident_types/{id}",
	Summary:     "Get a Severity",
	Tags:        incidentTypesTags,
	Errors:      errorCodes(),
}

type GetIncidentTypeRequest GetIdRequest
type GetIncidentTypeResponse ItemResponse[IncidentType]

var CreateIncidentType = huma.Operation{
	OperationID: "create-incident-type",
	Method:      http.MethodPost,
	Path:        "/incident_types",
	Summary:     "Create an Incident Type",
	Tags:        incidentTypesTags,
	Errors:      errorCodes(),
}

type CreateIncidentTypeAttributes struct {
	Name string `json:"name"`
}
type CreateIncidentTypeRequest RequestWithBodyAttributes[CreateIncidentTypeAttributes]
type CreateIncidentTypeResponse ItemResponse[IncidentType]

var UpdateIncidentType = huma.Operation{
	OperationID: "update-incident-type",
	Method:      http.MethodPatch,
	Path:        "/incident_types/{id}",
	Summary:     "Update an Incident Type",
	Tags:        incidentTypesTags,
	Errors:      errorCodes(),
}

type UpdateIncidentTypeAttributes struct {
	Name     *string `json:"name,omitempty"`
	Archived *bool   `json:"archived,omitempty"`
}
type UpdateIncidentTypeRequest UpdateIdRequest[UpdateIncidentTypeAttributes]
type UpdateIncidentTypeResponse ItemResponse[IncidentType]

var ArchiveIncidentType = huma.Operation{
	OperationID: "archive-incident-type",
	Method:      http.MethodDelete,
	Path:        "/incident_types/{id}",
	Summary:     "Archive an Incident Type",
	Tags:        incidentTypesTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentTypeRequest ArchiveIdRequest
type ArchiveIncidentTypeResponse EmptyResponse
