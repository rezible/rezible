package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"
)

type SystemAnalysisHandler interface {
	ListSystemComponents(context.Context, *ListSystemComponentsRequest) (*ListSystemComponentsResponse, error)
	CreateSystemComponent(context.Context, *CreateSystemComponentRequest) (*CreateSystemComponentResponse, error)
	GetSystemComponent(context.Context, *GetSystemComponentRequest) (*GetSystemComponentResponse, error)
	UpdateSystemComponent(context.Context, *UpdateSystemComponentRequest) (*UpdateSystemComponentResponse, error)
	ArchiveSystemComponent(context.Context, *ArchiveSystemComponentRequest) (*ArchiveSystemComponentResponse, error)

	GetSystemAnalysis(context.Context, *GetSystemAnalysisRequest) (*GetSystemAnalysisResponse, error)

	ListSystemAnalysisComponents(context.Context, *ListSystemAnalysisComponentsRequest) (*ListSystemAnalysisComponentsResponse, error)
	AddSystemAnalysisComponent(context.Context, *AddSystemAnalysisComponentRequest) (*AddSystemAnalysisComponentResponse, error)
	GetSystemAnalysisComponent(context.Context, *GetSystemAnalysisComponentRequest) (*GetSystemAnalysisComponentResponse, error)
	UpdateSystemAnalysisComponent(context.Context, *UpdateSystemAnalysisComponentRequest) (*UpdateSystemAnalysisComponentResponse, error)
	DeleteSystemAnalysisComponent(context.Context, *DeleteSystemAnalysisComponentRequest) (*DeleteSystemAnalysisComponentResponse, error)

	ListSystemAnalysisRelationships(context.Context, *ListSystemAnalysisRelationshipsRequest) (*ListSystemAnalysisRelationshipsResponse, error)
	CreateSystemAnalysisRelationship(context.Context, *CreateSystemAnalysisRelationshipRequest) (*CreateSystemAnalysisRelationshipResponse, error)
	GetSystemAnalysisRelationship(context.Context, *GetSystemAnalysisRelationshipRequest) (*GetSystemAnalysisRelationshipResponse, error)
	UpdateSystemAnalysisRelationship(context.Context, *UpdateSystemAnalysisRelationshipRequest) (*UpdateSystemAnalysisRelationshipResponse, error)
	DeleteSystemAnalysisRelationship(context.Context, *DeleteSystemAnalysisRelationshipRequest) (*DeleteSystemAnalysisRelationshipResponse, error)
}

func (o operations) RegisterSystemAnalysis(api huma.API) {
	huma.Register(api, ListSystemComponents, o.ListSystemComponents)
	huma.Register(api, CreateSystemComponent, o.CreateSystemComponent)
	huma.Register(api, GetSystemComponent, o.GetSystemComponent)
	huma.Register(api, UpdateSystemComponent, o.UpdateSystemComponent)
	huma.Register(api, ArchiveSystemComponent, o.ArchiveSystemComponent)

	huma.Register(api, GetSystemAnalysis, o.GetSystemAnalysis)

	huma.Register(api, ListSystemAnalysisComponents, o.ListSystemAnalysisComponents)
	huma.Register(api, AddSystemAnalysisComponent, o.AddSystemAnalysisComponent)
	huma.Register(api, GetSystemAnalysisComponent, o.GetSystemAnalysisComponent)
	huma.Register(api, UpdateSystemAnalysisComponent, o.UpdateSystemAnalysisComponent)
	huma.Register(api, DeleteSystemAnalysisComponent, o.DeleteSystemAnalysisComponent)

	huma.Register(api, ListSystemAnalysisRelationships, o.ListSystemAnalysisRelationships)
	huma.Register(api, GetSystemAnalysisRelationship, o.GetSystemAnalysisRelationship)
	huma.Register(api, CreateSystemAnalysisRelationship, o.CreateSystemAnalysisRelationship)
	huma.Register(api, UpdateSystemAnalysisRelationship, o.UpdateSystemAnalysisRelationship)
	huma.Register(api, DeleteSystemAnalysisRelationship, o.DeleteSystemAnalysisRelationship)
}

type (
	SystemComponentSignal struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes SystemComponentSignalAttributes `json:"attributes"`
	}
	SystemComponentSignalAttributes struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}

	SystemComponentControl struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes SystemComponentControlAttributes `json:"attributes"`
	}
	SystemComponentControlAttributes struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}

	SystemComponentConstraint struct {
		Id         uuid.UUID                           `json:"id"`
		Attributes SystemComponentConstraintAttributes `json:"attributes"`
	}
	SystemComponentConstraintAttributes struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}

	SystemComponent struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes SystemComponentAttributes `json:"attributes"`
	}
	SystemComponentAttributes struct {
		Name        string                      `json:"name"`
		Kind        string                      `json:"kind" enum:"service"`
		Description string                      `json:"description"`
		Properties  map[string]any              `json:"properties"`
		Constraints []SystemComponentConstraint `json:"constraints"`
		Signals     []SystemComponentSignal     `json:"signals"`
		Controls    []SystemComponentControl    `json:"controls"`
	}

	ScopedSystemAnalysis struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes ScopedSystemAnalysisAttributes `json:"attributes"`
	}

	ScopedSystemAnalysisAttributes struct {
		Components    []SystemAnalysisComponent    `json:"components"`
		Relationships []SystemAnalysisRelationship `json:"relationships"`
	}

	SystemAnalysisComponent struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemAnalysisRelationshipAttributes `json:"attributes"`
	}
	SystemAnalysisComponentAttributes struct {
		ComponentId uuid.UUID `json:"component_id"`
		Role        string    `json:"role"`
	}

	SystemAnalysisRelationship struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemAnalysisRelationshipAttributes `json:"attributes"`
	}
	SystemAnalysisRelationshipAttributes struct {
		SourceId          uuid.UUID   `json:"source_id"`
		TargetId          uuid.UUID   `json:"target_id"`
		Description       string      `json:"description"`
		FeedbackSignalIds []uuid.UUID `json:"feedback_signals"` // IDs of SystemComponentSignal from target to source
		ControlActionIds  []uuid.UUID `json:"control_actions"`  // IDs of SystemComponentControl from source to target
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

var systemAnalysisTags = []string{"System Analysis"}

// System Components

var ListSystemComponents = huma.Operation{
	OperationID: "list-system-components",
	Method:      http.MethodGet,
	Path:        "/system_components",
	Summary:     "List System Components",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type ListSystemComponentsRequest struct {
	ListRequest
}
type ListSystemComponentsResponse PaginatedResponse[SystemComponent]

var CreateSystemComponent = huma.Operation{
	OperationID: "create-system-component",
	Method:      http.MethodPost,
	Path:        "/system_components",
	Summary:     "Create a System Component",
	Tags:        systemAnalysisTags,
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
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemComponentRequest GetIdRequest
type GetSystemComponentResponse ItemResponse[SystemComponent]

var UpdateSystemComponent = huma.Operation{
	OperationID: "update-system-component",
	Method:      http.MethodPatch,
	Path:        "/system_components/{id}",
	Summary:     "Update a System Component",
	Tags:        systemAnalysisTags,
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
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentRequest ArchiveIdRequest
type ArchiveSystemComponentResponse EmptyResponse

var GetSystemAnalysis = huma.Operation{
	OperationID: "get-system-analysis",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}",
	Summary:     "Get System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemAnalysisRequest GetIdRequest
type GetSystemAnalysisResponse ItemResponse[ScopedSystemAnalysis]

// analysis components

var AddSystemAnalysisComponent = huma.Operation{
	OperationID: "add-system-analysis-component",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/components",
	Summary:     "Add a Component to a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type AddSystemAnalysisComponentAttributes struct {
	ComponentId uuid.UUID `json:"component_id"`
	Role        string    `json:"role"`
}
type AddSystemAnalysisComponentRequest CreateIdRequest[AddSystemAnalysisComponentAttributes]
type AddSystemAnalysisComponentResponse ItemResponse[SystemAnalysisComponent]

var ListSystemAnalysisComponents = huma.Operation{
	OperationID: "list-system-analysis-components",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/components",
	Summary:     "List components in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type ListSystemAnalysisComponentsRequest ListIdRequest
type ListSystemAnalysisComponentsResponse PaginatedResponse[SystemAnalysisComponent]

type SystemAnalysisEntityRequest struct {
	AnalysisId uuid.UUID `path:"analysis_id"`
	EntityId   uuid.UUID `path:"entity_id"`
}

var GetSystemAnalysisComponent = huma.Operation{
	OperationID: "get-system-analysis-component",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{analysis_id}/components/{entity_id}",
	Summary:     "Get a component in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemAnalysisComponentRequest SystemAnalysisEntityRequest
type GetSystemAnalysisComponentResponse ItemResponse[SystemAnalysisComponent]

var UpdateSystemAnalysisComponent = huma.Operation{
	OperationID: "update-system-analysis-component",
	Method:      http.MethodPatch,
	Path:        "/system_analysis/{analysis_id}/components/{entity_id}",
	Summary:     "Update a System Analysis Component",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type UpdateSystemAnalysisComponentAttributes struct {
	Role *string `json:"role,omitempty"`
}
type UpdateSystemAnalysisComponentRequest struct {
	SystemAnalysisEntityRequest
	RequestWithBodyAttributes[UpdateSystemAnalysisComponentAttributes]
}
type UpdateSystemAnalysisComponentResponse ItemResponse[SystemAnalysisComponent]

var DeleteSystemAnalysisComponent = huma.Operation{
	OperationID: "delete-system-analysis-component",
	Method:      http.MethodDelete,
	Path:        "/system_analysis/{analysis_id}/components/{entity_id}",
	Summary:     "Delete a Component from a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type DeleteSystemAnalysisComponentRequest SystemAnalysisEntityRequest
type DeleteSystemAnalysisComponentResponse EmptyResponse

// analysis relationships

var CreateSystemAnalysisRelationship = huma.Operation{
	OperationID: "create-system-analysis-relationship",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/relationships",
	Summary:     "Create a Relationship in a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type CreateSystemAnalysisRelationshipAttributes struct {
	// TODO: relationship attributes
}
type CreateSystemAnalysisRelationshipRequest CreateIdRequest[CreateSystemAnalysisRelationshipAttributes]
type CreateSystemAnalysisRelationshipResponse ItemResponse[SystemAnalysisRelationship]

var ListSystemAnalysisRelationships = huma.Operation{
	OperationID: "list-system-analysis-relationships",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/relationships",
	Summary:     "List relationships in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type ListSystemAnalysisRelationshipsRequest ListIdRequest
type ListSystemAnalysisRelationshipsResponse PaginatedResponse[SystemAnalysisRelationship]

var GetSystemAnalysisRelationship = huma.Operation{
	OperationID: "get-system-analysis-relationship",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{analysis_id}/relationships/{entity_id}",
	Summary:     "Get a relationship in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemAnalysisRelationshipRequest SystemAnalysisEntityRequest
type GetSystemAnalysisRelationshipResponse ItemResponse[SystemAnalysisRelationship]

var UpdateSystemAnalysisRelationship = huma.Operation{
	OperationID: "update-system-analysis-relationship",
	Method:      http.MethodPatch,
	Path:        "/system_analysis/{analysis_id}/relationships/{entity_id}",
	Summary:     "Update a System Analysis Relationship",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type UpdateSystemAnalysisRelationshipAttributes struct {
	// TODO
}
type UpdateSystemAnalysisRelationshipRequest struct {
	SystemAnalysisEntityRequest
	RequestWithBodyAttributes[UpdateSystemAnalysisRelationshipAttributes]
}
type UpdateSystemAnalysisRelationshipResponse ItemResponse[SystemAnalysisRelationship]

var DeleteSystemAnalysisRelationship = huma.Operation{
	OperationID: "delete-system-analysis-relationship",
	Method:      http.MethodDelete,
	Path:        "/system_analysis/{analysis_id}/relationships/{entity_id}",
	Summary:     "Delete a Relationship from a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type DeleteSystemAnalysisRelationshipRequest SystemAnalysisEntityRequest
type DeleteSystemAnalysisRelationshipResponse EmptyResponse
