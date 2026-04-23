package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type IncidentMetadataHandler interface {
	GetIncidentMetadata(context.Context, *GetIncidentMetadataRequest) (*GetIncidentMetadataResponse, error)

	ListIncidentSeverities(context.Context, *ListIncidentSeveritiesRequest) (*ListIncidentSeveritiesResponse, error)
	CreateIncidentSeverity(context.Context, *CreateIncidentSeverityRequest) (*CreateIncidentSeverityResponse, error)
	GetIncidentSeverity(context.Context, *GetIncidentSeverityRequest) (*GetIncidentSeverityResponse, error)
	UpdateIncidentSeverity(context.Context, *UpdateIncidentSeverityRequest) (*UpdateIncidentSeverityResponse, error)
	ArchiveIncidentSeverity(context.Context, *ArchiveIncidentSeverityRequest) (*ArchiveIncidentSeverityResponse, error)

	ListIncidentTypes(context.Context, *ListIncidentTypesRequest) (*ListIncidentTypesResponse, error)
	CreateIncidentType(context.Context, *CreateIncidentTypeRequest) (*CreateIncidentTypeResponse, error)
	GetIncidentType(context.Context, *GetIncidentTypeRequest) (*GetIncidentTypeResponse, error)
	UpdateIncidentType(context.Context, *UpdateIncidentTypeRequest) (*UpdateIncidentTypeResponse, error)
	ArchiveIncidentType(context.Context, *ArchiveIncidentTypeRequest) (*ArchiveIncidentTypeResponse, error)

	ListIncidentRoles(context.Context, *ListIncidentRolesRequest) (*ListIncidentRolesResponse, error)
	CreateIncidentRole(context.Context, *CreateIncidentRoleRequest) (*CreateIncidentRoleResponse, error)
	GetIncidentRole(context.Context, *GetIncidentRoleRequest) (*GetIncidentRoleResponse, error)
	UpdateIncidentRole(context.Context, *UpdateIncidentRoleRequest) (*UpdateIncidentRoleResponse, error)
	ArchiveIncidentRole(context.Context, *ArchiveIncidentRoleRequest) (*ArchiveIncidentRoleResponse, error)

	ListIncidentTags(context.Context, *ListIncidentTagsRequest) (*ListIncidentTagsResponse, error)
	CreateIncidentTag(context.Context, *CreateIncidentTagRequest) (*CreateIncidentTagResponse, error)
	GetIncidentTag(context.Context, *GetIncidentTagRequest) (*GetIncidentTagResponse, error)
	UpdateIncidentTag(context.Context, *UpdateIncidentTagRequest) (*UpdateIncidentTagResponse, error)
	ArchiveIncidentTag(context.Context, *ArchiveIncidentTagRequest) (*ArchiveIncidentTagResponse, error)

	ListIncidentFields(context.Context, *ListIncidentFieldsRequest) (*ListIncidentFieldsResponse, error)
	CreateIncidentField(context.Context, *CreateIncidentFieldRequest) (*CreateIncidentFieldResponse, error)
	GetIncidentField(context.Context, *GetIncidentFieldRequest) (*GetIncidentFieldResponse, error)
	UpdateIncidentField(context.Context, *UpdateIncidentFieldRequest) (*UpdateIncidentFieldResponse, error)
	ArchiveIncidentField(context.Context, *ArchiveIncidentFieldRequest) (*ArchiveIncidentFieldResponse, error)
}

func (o operations) RegisterIncidentMetadata(api huma.API) {
	huma.Register(api, GetIncidentMetadata, o.GetIncidentMetadata)

	huma.Register(api, ListIncidentSeverities, o.ListIncidentSeverities)
	huma.Register(api, CreateIncidentSeverity, o.CreateIncidentSeverity)
	huma.Register(api, GetIncidentSeverity, o.GetIncidentSeverity)
	huma.Register(api, UpdateIncidentSeverity, o.UpdateIncidentSeverity)
	huma.Register(api, ArchiveIncidentSeverity, o.ArchiveIncidentSeverity)

	huma.Register(api, ListIncidentRoles, o.ListIncidentRoles)
	huma.Register(api, CreateIncidentRole, o.CreateIncidentRole)
	huma.Register(api, GetIncidentRole, o.GetIncidentRole)
	huma.Register(api, UpdateIncidentRole, o.UpdateIncidentRole)
	huma.Register(api, ArchiveIncidentRole, o.ArchiveIncidentRole)

	huma.Register(api, ListIncidentTypes, o.ListIncidentTypes)
	huma.Register(api, CreateIncidentType, o.CreateIncidentType)
	huma.Register(api, GetIncidentType, o.GetIncidentType)
	huma.Register(api, UpdateIncidentType, o.UpdateIncidentType)
	huma.Register(api, ArchiveIncidentType, o.ArchiveIncidentType)

	huma.Register(api, ListIncidentTags, o.ListIncidentTags)
	huma.Register(api, CreateIncidentTag, o.CreateIncidentTag)
	huma.Register(api, GetIncidentTag, o.GetIncidentTag)
	huma.Register(api, UpdateIncidentTag, o.UpdateIncidentTag)
	huma.Register(api, ArchiveIncidentTag, o.ArchiveIncidentTag)

	huma.Register(api, ListIncidentFields, o.ListIncidentFields)
	huma.Register(api, CreateIncidentField, o.CreateIncidentField)
	huma.Register(api, GetIncidentField, o.GetIncidentField)
	huma.Register(api, UpdateIncidentField, o.UpdateIncidentField)
	huma.Register(api, ArchiveIncidentField, o.ArchiveIncidentField)
}

type (
	IncidentMetadata struct {
		Severities []IncidentSeverity `json:"severities"`
		Roles      []IncidentRole     `json:"roles"`
		Types      []IncidentType     `json:"types"`
		Tags       []IncidentTag      `json:"tags"`
		Fields     []IncidentField    `json:"fields"`
	}

	IncidentSeverity struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes IncidentSeverityAttributes `json:"attributes"`
	}
	IncidentSeverityAttributes struct {
		Name        string `json:"name"`
		Rank        int    `json:"rank"`
		Color       string `json:"color"`
		Archived    bool   `json:"archived"`
		Description string `json:"description"`
	}

	IncidentType struct {
		Id         uuid.UUID              `json:"id"`
		Attributes IncidentTypeAttributes `json:"attributes"`
	}
	IncidentTypeAttributes struct {
		Name        string `json:"name"`
		Archived    bool   `json:"archived"`
		Description string `json:"description"`
	}

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

	IncidentTag struct {
		Id         uuid.UUID             `json:"id"`
		Attributes IncidentTagAttributes `json:"attributes"`
	}
	IncidentTagAttributes struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		Archived    bool   `json:"archived"`
		Description string `json:"description"`
	}

	IncidentField struct {
		Id         uuid.UUID               `json:"id"`
		Attributes IncidentFieldAttributes `json:"attributes"`
	}
	IncidentFieldAttributes struct {
		Name         string                `json:"name"`
		Archived     bool                  `json:"archived"`
		Description  string                `json:"description"`
		Required     bool                  `json:"required"`
		IncidentType *IncidentType         `json:"incidentType"`
		Options      []IncidentFieldOption `json:"options"`
	}
	IncidentFieldOption struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes IncidentFieldOptionAttributes `json:"attributes"`
	}
	IncidentFieldOptionAttributes struct {
		FieldOptionType string `json:"optionType" enum:"custom,derived"`
		Value           string `json:"value"`
		Archived        bool   `json:"archived"`
	}
)

func convert[A any, B any](items []A, fn func(A) B) []B {
	result := make([]B, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

func IncidentMetadataFromRez(md *rez.IncidentMetadata) IncidentMetadata {
	return IncidentMetadata{
		Severities: convert(md.Severities, IncidentSeverityFromEnt),
		Roles:      convert(md.Roles, IncidentRoleFromEnt),
		Types:      convert(md.Types, IncidentTypeFromEnt),
		Fields:     convert(md.Fields, IncidentFieldFromEnt),
		Tags:       convert(md.Tags, IncidentTagFromEnt),
	}
}

func IncidentSeverityFromEnt(sev *ent.IncidentSeverity) IncidentSeverity {
	return IncidentSeverity{
		Id: sev.ID,
		Attributes: IncidentSeverityAttributes{
			Name:        sev.Name,
			Rank:        sev.Rank,
			Color:       sev.Color,
			Description: sev.Description,
			Archived:    !sev.ArchiveTime.IsZero(),
		},
	}
}

func IncidentTypeFromEnt(t *ent.IncidentType) IncidentType {
	return IncidentType{
		Id: t.ID,
		Attributes: IncidentTypeAttributes{
			Name:     t.Name,
			Archived: !t.ArchiveTime.IsZero(),
		},
	}
}

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

func IncidentTagFromEnt(tag *ent.IncidentTag) IncidentTag {
	return IncidentTag{
		Id: tag.ID,
		Attributes: IncidentTagAttributes{
			Key:      tag.Key,
			Value:    tag.Value,
			Archived: !tag.ArchiveTime.IsZero(),
		},
	}
}

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

var GetIncidentMetadata = huma.Operation{
	OperationID: "get-incident-metadata",
	Method:      http.MethodGet,
	Path:        "/incident_metadata",
	Summary:     "Get available incident metadata options",
	Tags:        incidentsTags,
	Errors:      ErrorCodes(),
}

type GetIncidentMetadataRequest EmptyRequest
type GetIncidentMetadataResponse ItemResponse[IncidentMetadata]

// Incident Severities
var incidentSeveritiesTags = []string{"Incident Severities"}

var ListIncidentSeverities = huma.Operation{
	OperationID: "list-incident-severities",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/severities",
	Summary:     "List Severities",
	Tags:        incidentSeveritiesTags,
	Errors:      ErrorCodes(),
}

type ListIncidentSeveritiesRequest ListRequest
type ListIncidentSeveritiesResponse ListResponse[IncidentSeverity]

var GetIncidentSeverity = huma.Operation{
	OperationID: "get-incident-severity",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/severities/{id}",
	Summary:     "Get a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      ErrorCodes(),
}

type GetIncidentSeverityRequest GetIdRequest
type GetIncidentSeverityResponse ItemResponse[IncidentSeverity]

var CreateIncidentSeverity = huma.Operation{
	OperationID: "create-incident-severity",
	Method:      http.MethodPost,
	Path:        "/incident_metadata/severities",
	Summary:     "Create a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      ErrorCodes(),
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
	Path:        "/incident_metadata/severities/{id}",
	Summary:     "Update a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      ErrorCodes(),
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
	Path:        "/incident_metadata/severities/{id}",
	Summary:     "Archive a Severity",
	Tags:        incidentSeveritiesTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentSeverityRequest ArchiveIdRequest
type ArchiveIncidentSeverityResponse EmptyResponse

// Incident Types

var incidentTypesTags = []string{"Incident Types"}

var ListIncidentTypes = huma.Operation{
	OperationID: "list-incident-types",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/types",
	Summary:     "List Types",
	Tags:        incidentTypesTags,
	Errors:      ErrorCodes(),
}

type ListIncidentTypesRequest ListRequest
type ListIncidentTypesResponse ListResponse[IncidentType]

var GetIncidentType = huma.Operation{
	OperationID: "get-incident-type",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/types/{id}",
	Summary:     "Get a Severity",
	Tags:        incidentTypesTags,
	Errors:      ErrorCodes(),
}

type GetIncidentTypeRequest GetIdRequest
type GetIncidentTypeResponse ItemResponse[IncidentType]

var CreateIncidentType = huma.Operation{
	OperationID: "create-incident-type",
	Method:      http.MethodPost,
	Path:        "/incident_metadata/types",
	Summary:     "Create an Incident Type",
	Tags:        incidentTypesTags,
	Errors:      ErrorCodes(),
}

type CreateIncidentTypeAttributes struct {
	Name string `json:"name"`
}
type CreateIncidentTypeRequest RequestWithBodyAttributes[CreateIncidentTypeAttributes]
type CreateIncidentTypeResponse ItemResponse[IncidentType]

var UpdateIncidentType = huma.Operation{
	OperationID: "update-incident-type",
	Method:      http.MethodPatch,
	Path:        "/incident_metadata/types/{id}",
	Summary:     "Update an Incident Type",
	Tags:        incidentTypesTags,
	Errors:      ErrorCodes(),
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
	Path:        "/incident_metadata/types/{id}",
	Summary:     "Archive an Incident Type",
	Tags:        incidentTypesTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentTypeRequest ArchiveIdRequest
type ArchiveIncidentTypeResponse EmptyResponse

// Incident Roles
var incidentRolesTags = []string{"Incident Roles"}

var ListIncidentRoles = huma.Operation{
	OperationID: "list-incident-roles",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/roles",
	Summary:     "List Incident Roles",
	Tags:        incidentRolesTags,
	Errors:      ErrorCodes(),
}

type ListIncidentRolesRequest ListRequest
type ListIncidentRolesResponse ListResponse[IncidentRole]

var GetIncidentRole = huma.Operation{
	OperationID: "get-incident-role",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/roles/{id}",
	Summary:     "Get an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      ErrorCodes(),
}

type GetIncidentRoleRequest GetIdRequest
type GetIncidentRoleResponse ItemResponse[IncidentRole]

var CreateIncidentRole = huma.Operation{
	OperationID: "create-incident-role",
	Method:      http.MethodPost,
	Path:        "/incident_metadata/roles",
	Summary:     "Create an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      ErrorCodes(),
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
	Path:        "/incident_metadata/roles/{id}",
	Summary:     "Update an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      ErrorCodes(),
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
	Path:        "/incident_metadata/roles/{id}",
	Summary:     "Archive an Incident Role",
	Tags:        incidentRolesTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentRoleRequest ArchiveIdRequest
type ArchiveIncidentRoleResponse EmptyResponse

// Incident Tags

var incidentTagsTags = []string{"Incident Tags"}

var ListIncidentTags = huma.Operation{
	OperationID: "list-incident-tags",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/tags",
	Summary:     "List Incident Tags",
	Tags:        incidentTagsTags,
	Errors:      ErrorCodes(),
}

type ListIncidentTagsRequest ListRequest
type ListIncidentTagsResponse ListResponse[IncidentTag]

var GetIncidentTag = huma.Operation{
	OperationID: "get-incident-tag",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/tags/{id}",
	Summary:     "Get an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      ErrorCodes(),
}

type GetIncidentTagRequest GetIdRequest
type GetIncidentTagResponse ItemResponse[IncidentTag]

var CreateIncidentTag = huma.Operation{
	OperationID: "create-incident-tag",
	Method:      http.MethodPost,
	Path:        "/incident_metadata/tags",
	Summary:     "Create an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      ErrorCodes(),
}

type CreateIncidentTagAttributes struct {
	Value string `json:"value"`
}
type CreateIncidentTagRequest RequestWithBodyAttributes[CreateIncidentTagAttributes]
type CreateIncidentTagResponse ItemResponse[IncidentTag]

var UpdateIncidentTag = huma.Operation{
	OperationID: "update-incident-tag",
	Method:      http.MethodPatch,
	Path:        "/incident_metadata/tags/{id}",
	Summary:     "Update an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      ErrorCodes(),
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
	Path:        "/incident_metadata/tags/{id}",
	Summary:     "Archive an Incident Tag",
	Tags:        incidentTagsTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentTagRequest ArchiveIdRequest
type ArchiveIncidentTagResponse EmptyResponse

// Incident Fields

var incidentFieldsTags = []string{"Incident Fields"}

var ListIncidentFields = huma.Operation{
	OperationID: "list-incident-fields",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/fields",
	Summary:     "List Incident Fields",
	Tags:        incidentFieldsTags,
	Errors:      ErrorCodes(),
}

type ListIncidentFieldsRequest ListRequest
type ListIncidentFieldsResponse ListResponse[IncidentField]

var GetIncidentField = huma.Operation{
	OperationID: "get-incident-field",
	Method:      http.MethodGet,
	Path:        "/incident_metadata/fields/{id}",
	Summary:     "Get an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      ErrorCodes(),
}

type GetIncidentFieldRequest GetIdRequest
type GetIncidentFieldResponse ItemResponse[IncidentField]

var CreateIncidentField = huma.Operation{
	OperationID: "create-incident-field",
	Method:      http.MethodPost,
	Path:        "/incident_metadata/fields",
	Summary:     "Create an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      ErrorCodes(),
}

type CreateIncidentFieldAttributes struct {
	Name         string                                `json:"name"`
	Required     bool                                  `json:"required"`
	Options      []CreateIncidentFieldOptionAttributes `json:"options" min:"1"`
	IncidentType *string                               `json:"incidentType,omitempty"`
}
type CreateIncidentFieldOptionAttributes struct {
	FieldOptionType string `json:"fieldOptionType" enum:"custom,derived"`
	Value           string `json:"value"`
}
type CreateIncidentFieldRequest RequestWithBodyAttributes[CreateIncidentFieldAttributes]
type CreateIncidentFieldResponse ItemResponse[IncidentField]

var UpdateIncidentField = huma.Operation{
	OperationID: "update-incident-field",
	Method:      http.MethodPatch,
	Path:        "/incident_metadata/fields/{id}",
	Summary:     "Update an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      ErrorCodes(),
}

type UpdateIncidentFieldAttributes struct {
	Name         *string                                `json:"name,omitempty"`
	Archived     *bool                                  `json:"archived,omitempty"`
	Required     *bool                                  `json:"required,omitempty"`
	IncidentType *string                                `json:"incidentType,omitempty"`
	Options      *[]UpdateIncidentFieldOptionAttributes `json:"options,omitempty"`
}
type UpdateIncidentFieldOptionAttributes struct {
	Id              *string `json:"id,omitempty"`
	FieldOptionType string  `json:"fieldOptionType" enum:"custom,derived"`
	Value           string  `json:"value"`
	Archived        bool    `json:"archived"`
}
type UpdateIncidentFieldRequest UpdateIdRequest[UpdateIncidentFieldAttributes]
type UpdateIncidentFieldResponse ItemResponse[IncidentField]

var ArchiveIncidentField = huma.Operation{
	OperationID: "archive-incident-field",
	Method:      http.MethodDelete,
	Path:        "/incident_metadata/fields/{id}",
	Summary:     "Archive an Incident Field",
	Tags:        incidentFieldsTags,
	Errors:      ErrorCodes(),
}

type ArchiveIncidentFieldRequest ArchiveIdRequest
type ArchiveIncidentFieldResponse EmptyResponse
