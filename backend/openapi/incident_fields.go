package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"

	"github.com/rezible/rezible/ent"
)

type IncidentFieldsHandler interface {
	ListIncidentFields(context.Context, *ListIncidentFieldsRequest) (*ListIncidentFieldsResponse, error)
	CreateIncidentField(context.Context, *CreateIncidentFieldRequest) (*CreateIncidentFieldResponse, error)
	GetIncidentField(context.Context, *GetIncidentFieldRequest) (*GetIncidentFieldResponse, error)
	UpdateIncidentField(context.Context, *UpdateIncidentFieldRequest) (*UpdateIncidentFieldResponse, error)
	ArchiveIncidentField(context.Context, *ArchiveIncidentFieldRequest) (*ArchiveIncidentFieldResponse, error)
}

func (o operations) RegisterIncidentFields(api huma.API) {
	huma.Register(api, ListIncidentFields, o.ListIncidentFields)
	huma.Register(api, CreateIncidentField, o.CreateIncidentField)
	huma.Register(api, GetIncidentField, o.GetIncidentField)
	huma.Register(api, UpdateIncidentField, o.UpdateIncidentField)
	huma.Register(api, ArchiveIncidentField, o.ArchiveIncidentField)
}

type (
	IncidentField struct {
		Id         uuid.UUID               `json:"id"`
		Attributes IncidentFieldAttributes `json:"attributes"`
	}

	IncidentFieldAttributes struct {
		Name         string                `json:"name"`
		Archived     bool                  `json:"archived"`
		Description  string                `json:"description"`
		Required     bool                  `json:"required"`
		IncidentType *IncidentType         `json:"incident_type"`
		Options      []IncidentFieldOption `json:"options"`
	}

	IncidentFieldOption struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes IncidentFieldOptionAttributes `json:"attributes"`
	}

	IncidentFieldOptionAttributes struct {
		FieldOptionType string `json:"option_type" enum:"custom,derived"`
		Value           string `json:"value"`
		Archived        bool   `json:"archived"`
	}
)

func IncidentFieldFromEnt(field *ent.IncidentField) IncidentField {
	opts := make([]IncidentFieldOption, len(field.Edges.Options))
	for i, o := range field.Edges.Options {
		opts[i] = IncidentFieldOptionFromEnt(o)
	}
	f := IncidentField{
		Id: field.ID,
		Attributes: IncidentFieldAttributes{
			Name:         field.Name,
			Archived:     !field.ArchiveTime.IsZero(),
			Description:  "",
			Required:     false,
			IncidentType: nil,
			Options:      opts,
		},
	}
	return f
}

func IncidentFieldOptionFromEnt(opt *ent.IncidentFieldOption) IncidentFieldOption {
	return IncidentFieldOption{
		Id: opt.ID,
		Attributes: IncidentFieldOptionAttributes{
			Value:           opt.Value,
			FieldOptionType: opt.Type.String(),
			Archived:        !opt.ArchiveTime.IsZero(),
		},
	}
}

var incidentFieldsTags = []string{"Incident Fields"}

// ops

var ListIncidentFields = huma.Operation{
	OperationID: "list-incident-fields",
	Method:      http.MethodGet,
	Path:        "/incident_fields",
	Summary:     "List Incident Fields",
	Tags:        incidentFieldsTags,
	Errors:      errorCodes(),
}

type ListIncidentFieldsRequest ListRequest
type ListIncidentFieldsResponse PaginatedResponse[IncidentField]

var GetIncidentField = huma.Operation{
	OperationID: "get-incident-field",
	Method:      http.MethodGet,
	Path:        "/incident_fields/{id}",
	Summary:     "Get an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      errorCodes(),
}

type GetIncidentFieldRequest GetIdRequest
type GetIncidentFieldResponse ItemResponse[IncidentField]

var CreateIncidentField = huma.Operation{
	OperationID: "create-incident-field",
	Method:      http.MethodPost,
	Path:        "/incident_fields",
	Summary:     "Create an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      errorCodes(),
}

type CreateIncidentFieldAttributes struct {
	Name         string                                `json:"name"`
	Required     bool                                  `json:"required"`
	Options      []CreateIncidentFieldOptionAttributes `json:"options" min:"1"`
	IncidentType *string                               `json:"incident_type,omitempty"`
}
type CreateIncidentFieldOptionAttributes struct {
	FieldOptionType string `json:"field_option_type" enum:"custom,derived"`
	Value           string `json:"value"`
}
type CreateIncidentFieldRequest RequestWithBodyAttributes[CreateIncidentFieldAttributes]
type CreateIncidentFieldResponse ItemResponse[IncidentField]

var UpdateIncidentField = huma.Operation{
	OperationID: "update-incident-field",
	Method:      http.MethodPatch,
	Path:        "/incident_fields/{id}",
	Summary:     "Update an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentFieldAttributes struct {
	Name         *string                                `json:"name,omitempty"`
	Archived     *bool                                  `json:"archived,omitempty"`
	Required     *bool                                  `json:"required,omitempty"`
	IncidentType *string                                `json:"incident_type,omitempty"`
	Options      *[]UpdateIncidentFieldOptionAttributes `json:"options,omitempty"`
}
type UpdateIncidentFieldOptionAttributes struct {
	Id              *string `json:"id,omitempty"`
	FieldOptionType string  `json:"field_option_type" enum:"custom,derived"`
	Value           string  `json:"value"`
	Archived        bool    `json:"archived"`
}
type UpdateIncidentFieldRequest UpdateIdRequest[UpdateIncidentFieldAttributes]
type UpdateIncidentFieldResponse ItemResponse[IncidentField]

var ArchiveIncidentField = huma.Operation{
	OperationID: "archive-incident-field",
	Method:      http.MethodDelete,
	Path:        "/incident_fields/{id}",
	Summary:     "Archive an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentFieldRequest ArchiveIdRequest
type ArchiveIncidentFieldResponse EmptyResponse
