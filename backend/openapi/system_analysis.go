package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"
)

type SystemAnalysisHandler interface {
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
	SystemAnalysis struct {
		Id         uuid.UUID                `json:"id"`
		Attributes SystemAnalysisAttributes `json:"attributes"`
	}
	SystemAnalysisAttributes struct {
		Components    []SystemAnalysisComponent    `json:"components"`
		Relationships []SystemAnalysisRelationship `json:"relationships"`
	}

	SystemAnalysisComponent struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes SystemAnalysisComponentAttributes `json:"attributes"`
	}
	SystemAnalysisComponentAttributes struct {
		Component SystemComponent               `json:"component"`
		Position  SystemAnalysisDiagramPosition `json:"position"`
	}

	SystemAnalysisDiagramPosition struct {
		X float64  `json:"x"`
		Y float64  `json:"y"`
		Z *float64 `json:"z,omitempty"`
	}

	SystemAnalysisRelationship struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemAnalysisRelationshipAttributes `json:"attributes"`
	}
	SystemAnalysisRelationshipAttributes struct {
		SourceId        uuid.UUID                                  `json:"source_id"`
		TargetId        uuid.UUID                                  `json:"target_id"`
		Description     string                                     `json:"description"`
		FeedbackSignals []SystemAnalysisRelationshipFeedbackSignal `json:"feedback_signals"`
		ControlActions  []SystemAnalysisRelationshipControlAction  `json:"control_actions"`
	}

	SystemAnalysisRelationshipComponent struct {
		ComponentId uuid.UUID `json:"component_id"`
	}

	SystemAnalysisRelationshipControlAction struct {
		Id         uuid.UUID                                         `json:"id"`
		Attributes SystemAnalysisRelationshipControlActionAttributes `json:"attributes"`
	}
	SystemAnalysisRelationshipControlActionAttributes struct {
		ControlId   uuid.UUID `json:"control_id"`
		Description string    `json:"description"`
	}

	SystemAnalysisRelationshipFeedbackSignal struct {
		Id         uuid.UUID                                          `json:"id"`
		Attributes SystemAnalysisRelationshipFeedbackSignalAttributes `json:"attributes"`
	}
	SystemAnalysisRelationshipFeedbackSignalAttributes struct {
		SignalId    uuid.UUID `json:"signal_id"`
		Description string    `json:"description"`
	}
)

var systemAnalysisTags = []string{"System Analysis"}

var GetSystemAnalysis = huma.Operation{
	OperationID: "get-system-analysis",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}",
	Summary:     "Get System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemAnalysisRequest GetIdRequest
type GetSystemAnalysisResponse ItemResponse[SystemAnalysis]

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
	ComponentId uuid.UUID                     `json:"component_id"`
	Position    SystemAnalysisDiagramPosition `json:"position"`
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
