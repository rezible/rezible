package openapi

import (
	"context"
	"github.com/rezible/rezible/ent"
	"net/http"
	"reflect"

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

	ListSystemComponentRelationships(context.Context, *ListSystemComponentRelationshipsRequest) (*ListSystemComponentRelationshipsResponse, error)
	CreateSystemComponentRelationship(context.Context, *CreateSystemComponentRelationshipRequest) (*CreateSystemComponentRelationshipResponse, error)
	UpdateSystemComponentRelationship(context.Context, *UpdateSystemComponentRelationshipRequest) (*UpdateSystemComponentRelationshipResponse, error)
	ArchiveSystemComponentRelationship(context.Context, *ArchiveSystemComponentRelationshipRequest) (*ArchiveSystemComponentRelationshipResponse, error)
}

func (o operations) RegisterSystemComponents(api huma.API) {
	huma.Register(api, ListSystemComponents, o.ListSystemComponents)
	huma.Register(api, ListIncidentSystemComponents, o.ListIncidentSystemComponents)
	huma.Register(api, CreateSystemComponent, o.CreateSystemComponent)
	huma.Register(api, GetSystemComponent, o.GetSystemComponent)
	huma.Register(api, UpdateSystemComponent, o.UpdateSystemComponent)
	huma.Register(api, ArchiveSystemComponent, o.ArchiveSystemComponent)

	huma.Register(api, ListSystemComponentRelationships, o.ListSystemComponentRelationships)
	huma.Register(api, CreateSystemComponentRelationship, o.CreateSystemComponentRelationship)
	huma.Register(api, UpdateSystemComponentRelationship, o.UpdateSystemComponentRelationship)
	huma.Register(api, ArchiveSystemComponentRelationship, o.ArchiveSystemComponentRelationship)
}

type (
	SystemComponent struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes SystemComponentAttributes `json:"attributes"`
	}

	SystemComponentAttributes struct {
		Name          string                        `json:"name"`
		Kind          string                        `json:"kind" enum:"service"`
		Description   string                        `json:"description"`
		Properties    map[string]any                `json:"properties"`
		Relationships []SystemComponentRelationship `json:"relationships"`
	}

	SystemComponentRelationship struct {
		Id         uuid.UUID                             `json:"id"`
		Attributes SystemComponentRelationshipAttributes `json:"attributes"`
	}

	SystemComponentRelationshipAttributes struct {
		Kind    string `json:"kind" enum:"control,feedback"`
		Details any    `json:"details"`
	}

	SystemComponentControlRelationshipAttributes struct {
		Kind    string                                    `json:"kind" enum:"control"`
		Details SystemComponentControlRelationshipDetails `json:"details"`
	}

	SystemComponentFeedbackRelationshipAttributes struct {
		Kind    string                                     `json:"kind" enum:"feedback"`
		Details SystemComponentFeedbackRelationshipDetails `json:"details"`
	}

	SystemComponentControlRelationshipDetails struct {
		ControllerId uuid.UUID `json:"controller_id"`
		ControlledId uuid.UUID `json:"controlled_id"`
		Control      string    `json:"control"`
		Description  string    `json:"description"`
	}

	SystemComponentFeedbackRelationshipDetails struct {
		SourceId    uuid.UUID `json:"source_id"`
		TargetId    uuid.UUID `json:"target_id"`
		Feedback    string    `json:"feedback"`
		Description string    `json:"description"`
	}

	IncidentSystemComponent struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes IncidentSystemComponentAttributes `json:"attributes"`
	}

	IncidentSystemComponentAttributes struct {
		Role      string          `json:"role" enum:"primary,contributing,affected,mitigating"`
		Component SystemComponent `json:"component"`
	}
)

func (scr SystemComponentRelationshipAttributes) Schema(r huma.Registry) *huma.Schema {
	controlSchema := r.Schema(reflect.TypeOf(SystemComponentControlRelationshipAttributes{}), true, "Control")
	feedbackSchema := r.Schema(reflect.TypeOf(SystemComponentFeedbackRelationshipAttributes{}), true, "Feedback")

	return &huma.Schema{
		Type:        huma.TypeObject,
		Description: "Attributes specific to the kind of relationship",

		OneOf: []*huma.Schema{
			{Ref: controlSchema.Ref},
			{Ref: feedbackSchema.Ref},
		},
		Discriminator: &huma.Discriminator{
			PropertyName: "kind",
			Mapping: map[string]string{
				"control":  controlSchema.Ref,
				"feedback": feedbackSchema.Ref,
			},
		},
	}
}

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

func SystemComponentControlRelationshipFromEnt(r *ent.SystemComponentControlRelationship) SystemComponentRelationship {
	details := SystemComponentControlRelationshipDetails{
		ControllerId: r.ControllerID,
		ControlledId: r.ControlledID,
		Control:      r.Type,
		Description:  r.Description,
	}
	return SystemComponentRelationship{
		Id: r.ID,
		Attributes: SystemComponentRelationshipAttributes{
			Kind:    "control",
			Details: details,
		},
	}
}

func SystemComponentFeedbackRelationshipFromEnt(sc *ent.SystemComponentFeedbackRelationship) SystemComponentRelationship {
	details := SystemComponentFeedbackRelationshipDetails{
		SourceId:    sc.SourceID,
		TargetId:    sc.TargetID,
		Feedback:    sc.Type,
		Description: sc.Description,
	}
	return SystemComponentRelationship{
		Id: sc.ID,
		Attributes: SystemComponentRelationshipAttributes{
			Kind:    "feedback",
			Details: details,
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

type GetSystemComponentRequest GetIdRequest
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

var CreateSystemComponentRelationship = huma.Operation{
	OperationID: "create-system-component-relationship",
	Method:      http.MethodPost,
	Path:        "/system_components/{id}/relationships",
	Summary:     "Create a System Component Relationship",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type CreateSystemComponentRelationshipRequestAttributes struct {
	OtherId uuid.UUID `json:"otherId"`
}
type CreateSystemComponentRelationshipRequest CreateIdRequest[CreateSystemComponentRelationshipRequestAttributes]
type CreateSystemComponentRelationshipResponse ItemResponse[SystemComponentRelationship]

var ListSystemComponentRelationships = huma.Operation{
	OperationID: "list-system-component-relationships",
	Method:      http.MethodGet,
	Path:        "/system_components/{id}/relationships",
	Summary:     "List relationships for a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ListSystemComponentRelationshipsRequest ListIdRequest
type ListSystemComponentRelationshipsResponse PaginatedResponse[SystemComponentRelationship]

var UpdateSystemComponentRelationship = huma.Operation{
	OperationID: "update-system-component-relationship",
	Method:      http.MethodPatch,
	Path:        "/system_component_relationships/{id}",
	Summary:     "Update a System Component Relationship",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type UpdateSystemComponentRelationshipAttributes struct {
}
type UpdateSystemComponentRelationshipRequest UpdateIdRequest[UpdateSystemComponentRelationshipAttributes]
type UpdateSystemComponentRelationshipResponse ItemResponse[SystemComponentRelationship]

var ArchiveSystemComponentRelationship = huma.Operation{
	OperationID: "archive-system-component-relationship",
	Method:      http.MethodDelete,
	Path:        "/system_component_relationships/{id}",
	Summary:     "Archive a System Component Relationship",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentRelationshipRequest ArchiveIdRequest
type ArchiveSystemComponentRelationshipResponse EmptyResponse
