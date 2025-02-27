package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
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
		SourceId        uuid.UUID                                  `json:"sourceId"`
		TargetId        uuid.UUID                                  `json:"targetId"`
		Description     string                                     `json:"description"`
		FeedbackSignals []SystemAnalysisRelationshipFeedbackSignal `json:"feedbackSignals"`
		ControlActions  []SystemAnalysisRelationshipControlAction  `json:"controlActions"`
	}

	SystemAnalysisRelationshipControlAction struct {
		Id         uuid.UUID                                         `json:"id"`
		Attributes SystemAnalysisRelationshipControlActionAttributes `json:"attributes"`
	}
	SystemAnalysisRelationshipControlActionAttributes struct {
		ControlId   uuid.UUID `json:"controlId"`
		Description string    `json:"description"`
	}

	SystemAnalysisRelationshipFeedbackSignal struct {
		Id         uuid.UUID                                          `json:"id"`
		Attributes SystemAnalysisRelationshipFeedbackSignalAttributes `json:"attributes"`
	}
	SystemAnalysisRelationshipFeedbackSignalAttributes struct {
		SignalId    uuid.UUID `json:"signalId"`
		Description string    `json:"description"`
	}

	// TODO: System Hazard support
	SystemHazard struct {
		Id         uuid.UUID              `json:"id"`
		Attributes SystemHazardAttributes `json:"attributes"`
	}

	SystemHazardAttributes struct {
		Label          string      `json:"label"`
		Severity       string      `json:"severity"`
		Likelihood     string      `json:"likelihood"`
		Constraints    []uuid.UUID `json:"constraintIds"`
		Feedbacks      []uuid.UUID `json:"signalIds"`
		ControlActions []uuid.UUID `json:"controlIds"`
	}
)

func SystemAnalysisFromEnt(sc *ent.SystemAnalysis) SystemAnalysis {
	attr := SystemAnalysisAttributes{}

	// TODO

	return SystemAnalysis{
		Id:         sc.ID,
		Attributes: attr,
	}
}

func SystemAnalysisComponentFromEnt(sc *ent.SystemAnalysisComponent) SystemAnalysisComponent {
	return SystemAnalysisComponent{
		Id:         sc.ID,
		Attributes: SystemAnalysisComponentAttributes{},
	}
}

func SystemAnalysisRelationshipFromEnt(sc *ent.SystemAnalysisRelationship) SystemAnalysisRelationship {
	return SystemAnalysisRelationship{
		Id:         sc.ID,
		Attributes: SystemAnalysisRelationshipAttributes{},
	}
}

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
	ComponentId uuid.UUID                     `json:"componentId"`
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

var GetSystemAnalysisComponent = huma.Operation{
	OperationID: "get-system-analysis-component",
	Method:      http.MethodGet,
	Path:        "/system_analysis_components/{id}",
	Summary:     "Get a component in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemAnalysisComponentRequest GetIdRequest
type GetSystemAnalysisComponentResponse ItemResponse[SystemAnalysisComponent]

var UpdateSystemAnalysisComponent = huma.Operation{
	OperationID: "update-system-analysis-component",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_components/{id}",
	Summary:     "Update a System Analysis Component",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type UpdateSystemAnalysisComponentAttributes struct {
	Position *SystemAnalysisDiagramPosition `json:"position,omitempty"`
}
type UpdateSystemAnalysisComponentRequest UpdateIdRequest[UpdateSystemAnalysisComponentAttributes]
type UpdateSystemAnalysisComponentResponse ItemResponse[SystemAnalysisComponent]

var DeleteSystemAnalysisComponent = huma.Operation{
	OperationID: "delete-system-analysis-component",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_components/{id}",
	Summary:     "Delete a Component from a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type DeleteSystemAnalysisComponentRequest DeleteIdRequest
type DeleteSystemAnalysisComponentResponse EmptyResponse

// analysis relationships

var ListSystemAnalysisRelationships = huma.Operation{
	OperationID: "list-system-analysis-relationships",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/relationships",
	Summary:     "List relationships in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type ListSystemAnalysisRelationshipsRequest struct {
	ListIdRequest
	AnalysisComponentId uuid.UUID `query:"analysisComponentId"`
}
type ListSystemAnalysisRelationshipsResponse PaginatedResponse[SystemAnalysisRelationship]

var GetSystemAnalysisRelationship = huma.Operation{
	OperationID: "get-system-analysis-relationship",
	Method:      http.MethodGet,
	Path:        "/system_analysis_relationships/{id}",
	Summary:     "Get a relationship in a System analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type GetSystemAnalysisRelationshipRequest GetIdRequest
type GetSystemAnalysisRelationshipResponse ItemResponse[SystemAnalysisRelationship]

var CreateSystemAnalysisRelationship = huma.Operation{
	OperationID: "create-system-analysis-relationship",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/relationships",
	Summary:     "Create a Relationship in a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type CreateSystemAnalysisRelationshipAttributes struct {
	SourceId        uuid.UUID                                            `json:"sourceId"`
	TargetId        uuid.UUID                                            `json:"targetId"`
	Description     string                                               `json:"description"`
	FeedbackSignals []SystemAnalysisRelationshipFeedbackSignalAttributes `json:"feedbackSignals"`
	ControlActions  []SystemAnalysisRelationshipControlActionAttributes  `json:"controlActions"`
}
type CreateSystemAnalysisRelationshipRequest CreateIdRequest[CreateSystemAnalysisRelationshipAttributes]
type CreateSystemAnalysisRelationshipResponse ItemResponse[SystemAnalysisRelationship]

var UpdateSystemAnalysisRelationship = huma.Operation{
	OperationID: "update-system-analysis-relationship",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_relationships/{id}",
	Summary:     "Update a System Analysis Relationship",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type UpdateSystemAnalysisRelationshipAttributes struct {
	Description     *string                                               `json:"description,omitempty"`
	FeedbackSignals *[]SystemAnalysisRelationshipFeedbackSignalAttributes `json:"feedbackSignals,omitempty"`
	ControlActions  *[]SystemAnalysisRelationshipControlActionAttributes  `json:"controlActions,omitempty"`
}
type UpdateSystemAnalysisRelationshipRequest UpdateIdRequest[UpdateSystemAnalysisRelationshipAttributes]
type UpdateSystemAnalysisRelationshipResponse ItemResponse[SystemAnalysisRelationship]

var DeleteSystemAnalysisRelationship = huma.Operation{
	OperationID: "delete-system-analysis-relationship",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_relationships/{id}",
	Summary:     "Delete a Relationship from a System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      errorCodes(),
}

type DeleteSystemAnalysisRelationshipRequest DeleteIdRequest
type DeleteSystemAnalysisRelationshipResponse EmptyResponse
