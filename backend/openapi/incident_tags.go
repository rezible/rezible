package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"

	"github.com/twohundreds/rezible/ent"
)

type IncidentTagsHandler interface {
	ListIncidentTags(context.Context, *ListIncidentTagsRequest) (*ListIncidentTagsResponse, error)
	CreateIncidentTag(context.Context, *CreateIncidentTagRequest) (*CreateIncidentTagResponse, error)
	GetIncidentTag(context.Context, *GetIncidentTagRequest) (*GetIncidentTagResponse, error)
	UpdateIncidentTag(context.Context, *UpdateIncidentTagRequest) (*UpdateIncidentTagResponse, error)
	ArchiveIncidentTag(context.Context, *ArchiveIncidentTagRequest) (*ArchiveIncidentTagResponse, error)
}

func (o operations) RegisterIncidentTags(api huma.API) {
	huma.Register(api, ListIncidentTags, o.ListIncidentTags)
	huma.Register(api, CreateIncidentTag, o.CreateIncidentTag)
	huma.Register(api, GetIncidentTag, o.GetIncidentTag)
	huma.Register(api, UpdateIncidentTag, o.UpdateIncidentTag)
	huma.Register(api, ArchiveIncidentTag, o.ArchiveIncidentTag)
}

type (
	IncidentTag struct {
		Id         uuid.UUID             `json:"id"`
		Attributes IncidentTagAttributes `json:"attributes"`
	}

	IncidentTagAttributes struct {
		Value       string `json:"value"`
		Archived    bool   `json:"archived"`
		Description string `json:"description"`
	}
)

func IncidentTagFromEnt(tag *ent.IncidentTag) IncidentTag {
	return IncidentTag{
		Id:         tag.ID,
		Attributes: IncidentTagAttributes{},
	}
}

var incidentTagsTags = []string{"Incident Tags"}

// ops

var ListIncidentTags = huma.Operation{
	OperationID: "list-incident-tags",
	Method:      http.MethodGet,
	Path:        "/incident_tags",
	Summary:     "List Incident Tags",
	Tags:        incidentTagsTags,
	Errors:      errorCodes(),
}

type ListIncidentTagsRequest ListRequest
type ListIncidentTagsResponse PaginatedResponse[IncidentTag]

var GetIncidentTag = huma.Operation{
	OperationID: "get-incident-tag",
	Method:      http.MethodGet,
	Path:        "/incident_tags/{id}",
	Summary:     "Get an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      errorCodes(),
}

type GetIncidentTagRequest GetIdRequest
type GetIncidentTagResponse ItemResponse[IncidentTag]

var CreateIncidentTag = huma.Operation{
	OperationID: "create-incident-tag",
	Method:      http.MethodPost,
	Path:        "/incident_tags",
	Summary:     "Create an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      errorCodes(),
}

type CreateIncidentTagAttributes struct {
	Value string `json:"value"`
}
type CreateIncidentTagRequest RequestWithBodyAttributes[CreateIncidentTagAttributes]
type CreateIncidentTagResponse ItemResponse[IncidentTag]

var UpdateIncidentTag = huma.Operation{
	OperationID: "update-incident-tag",
	Method:      http.MethodPatch,
	Path:        "/incident_tags/{id}",
	Summary:     "Update an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentTagAttributes struct {
	Value    *string `json:"value,omitempty"`
	Archived *bool   `json:"archived,omitempty"`
}
type UpdateIncidentTagRequest UpdateIdRequest[UpdateIncidentTagAttributes]
type UpdateIncidentTagResponse ItemResponse[IncidentTag]

var ArchiveIncidentTag = huma.Operation{
	OperationID: "archive-incident-tag",
	Method:      http.MethodDelete,
	Path:        "/incident_tags/{id}",
	Summary:     "Archive an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentTagRequest ArchiveIdRequest
type ArchiveIncidentTagResponse EmptyResponse
