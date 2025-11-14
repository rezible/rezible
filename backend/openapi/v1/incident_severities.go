package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"

	"github.com/rezible/rezible/ent"
)

type IncidentSeveritiesHandler interface {
	ListIncidentSeverities(context.Context, *ListIncidentSeveritiesRequest) (*ListIncidentSeveritiesResponse, error)
	CreateIncidentSeverity(context.Context, *CreateIncidentSeverityRequest) (*CreateIncidentSeverityResponse, error)
	GetIncidentSeverity(context.Context, *GetIncidentSeverityRequest) (*GetIncidentSeverityResponse, error)
	UpdateIncidentSeverity(context.Context, *UpdateIncidentSeverityRequest) (*UpdateIncidentSeverityResponse, error)
	ArchiveIncidentSeverity(context.Context, *ArchiveIncidentSeverityRequest) (*ArchiveIncidentSeverityResponse, error)
}

func (o operations) RegisterIncidentSeverities(api huma.API) {
	huma.Register(api, ListIncidentSeverities, o.ListIncidentSeverities)
	huma.Register(api, CreateIncidentSeverity, o.CreateIncidentSeverity)
	huma.Register(api, GetIncidentSeverity, o.GetIncidentSeverity)
	huma.Register(api, UpdateIncidentSeverity, o.UpdateIncidentSeverity)
	huma.Register(api, ArchiveIncidentSeverity, o.ArchiveIncidentSeverity)
}

type (
	IncidentSeverity struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes IncidentSeverityAttributes `json:"attributes"`
	}

	IncidentSeverityAttributes struct {
		Name        string `json:"name"`
		Rank        int    `json:"rank"`
		Archived    bool   `json:"archived"`
		Description string `json:"description"`
	}
)

func IncidentSeverityFromEnt(sev *ent.IncidentSeverity) IncidentSeverity {
	return IncidentSeverity{
		Id: sev.ID,
		Attributes: IncidentSeverityAttributes{
			Name:        sev.Name,
			Rank:        sev.Rank,
			Description: sev.Description,
			Archived:    !sev.ArchiveTime.IsZero(),
		},
	}
}

var incidentSeveritiesTags = []string{"Incident Severities"}

// ops

var ListIncidentSeverities = huma.Operation{
	OperationID: "list-incident-severities",
	Method:      http.MethodGet,
	Path:        "/incident_severities",
	Summary:     "List Severities",
	Tags:        incidentSeveritiesTags,
	Errors:      errorCodes(),
}

type ListIncidentSeveritiesRequest ListRequest
type ListIncidentSeveritiesResponse PaginatedResponse[IncidentSeverity]

var GetIncidentSeverity = huma.Operation{
	OperationID: "get-incident-severity",
	Method:      http.MethodGet,
	Path:        "/incident_severities/{id}",
	Summary:     "Get a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      errorCodes(),
}

type GetIncidentSeverityRequest GetIdRequest
type GetIncidentSeverityResponse ItemResponse[IncidentSeverity]

var CreateIncidentSeverity = huma.Operation{
	OperationID: "create-incident-severity",
	Method:      http.MethodPost,
	Path:        "/incident_severities",
	Summary:     "Create a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      errorCodes(),
}

type CreateIncidentSeverityAttributes struct {
	Name string `json:"title"`
	Rank int    `json:"rank"`
}
type CreateIncidentSeverityRequest RequestWithBodyAttributes[CreateIncidentSeverityAttributes]
type CreateIncidentSeverityResponse ItemResponse[IncidentSeverity]

var UpdateIncidentSeverity = huma.Operation{
	OperationID: "update-incident-severity",
	Method:      http.MethodPatch,
	Path:        "/incident_severities/{id}",
	Summary:     "Update a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      errorCodes(),
}

type UpdateIncidentSeverityAttributes struct {
	Name     *string `json:"name,omitempty"`
	Rank     *int    `json:"rank,omitempty"`
	Archived *bool   `json:"archived,omitempty"`
}
type UpdateIncidentSeverityRequest UpdateIdRequest[UpdateIncidentSeverityAttributes]
type UpdateIncidentSeverityResponse ItemResponse[IncidentSeverity]

var ArchiveIncidentSeverity = huma.Operation{
	OperationID: "archive-incident-severity",
	Method:      http.MethodDelete,
	Path:        "/incident_severities/{id}",
	Summary:     "Archive a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentSeverityRequest ArchiveIdRequest
type ArchiveIncidentSeverityResponse EmptyResponse
