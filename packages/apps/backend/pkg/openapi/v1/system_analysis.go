package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type SystemAnalysisHandler interface {
	GetSystemAnalysis(context.Context, *GetSystemAnalysisRequest) (*GetSystemAnalysisResponse, error)

	ListSystemAnalysisNodes(context.Context, *ListSystemAnalysisNodesRequest) (*ListSystemAnalysisNodesResponse, error)
	AddSystemAnalysisNode(context.Context, *AddSystemAnalysisNodeRequest) (*AddSystemAnalysisNodeResponse, error)
	GetSystemAnalysisNode(context.Context, *GetSystemAnalysisNodeRequest) (*GetSystemAnalysisNodeResponse, error)
	UpdateSystemAnalysisNode(context.Context, *UpdateSystemAnalysisNodeRequest) (*UpdateSystemAnalysisNodeResponse, error)
	DeleteSystemAnalysisNode(context.Context, *DeleteSystemAnalysisNodeRequest) (*DeleteSystemAnalysisNodeResponse, error)

	ListSystemAnalysisEdges(context.Context, *ListSystemAnalysisEdgesRequest) (*ListSystemAnalysisEdgesResponse, error)
	AddSystemAnalysisEdge(context.Context, *AddSystemAnalysisEdgeRequest) (*AddSystemAnalysisEdgeResponse, error)
	GetSystemAnalysisEdge(context.Context, *GetSystemAnalysisEdgeRequest) (*GetSystemAnalysisEdgeResponse, error)
	UpdateSystemAnalysisEdge(context.Context, *UpdateSystemAnalysisEdgeRequest) (*UpdateSystemAnalysisEdgeResponse, error)
	DeleteSystemAnalysisEdge(context.Context, *DeleteSystemAnalysisEdgeRequest) (*DeleteSystemAnalysisEdgeResponse, error)
}

func (o operations) RegisterSystemAnalysis(api huma.API) {
	huma.Register(api, GetSystemAnalysis, o.GetSystemAnalysis)
	huma.Register(api, ListSystemAnalysisNodes, o.ListSystemAnalysisNodes)
	huma.Register(api, AddSystemAnalysisNode, o.AddSystemAnalysisNode)
	huma.Register(api, GetSystemAnalysisNode, o.GetSystemAnalysisNode)
	huma.Register(api, UpdateSystemAnalysisNode, o.UpdateSystemAnalysisNode)
	huma.Register(api, DeleteSystemAnalysisNode, o.DeleteSystemAnalysisNode)
	huma.Register(api, ListSystemAnalysisEdges, o.ListSystemAnalysisEdges)
	huma.Register(api, AddSystemAnalysisEdge, o.AddSystemAnalysisEdge)
	huma.Register(api, GetSystemAnalysisEdge, o.GetSystemAnalysisEdge)
	huma.Register(api, UpdateSystemAnalysisEdge, o.UpdateSystemAnalysisEdge)
	huma.Register(api, DeleteSystemAnalysisEdge, o.DeleteSystemAnalysisEdge)
}

type (
	SystemAnalysis struct {
		Id         uuid.UUID                `json:"id"`
		Attributes SystemAnalysisAttributes `json:"attributes"`
	}
	SystemAnalysisAttributes struct {
		Nodes []SystemAnalysisNode `json:"nodes"`
		Edges []SystemAnalysisEdge `json:"edges"`
	}

	SystemAnalysisNode struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes SystemAnalysisNodeAttributes `json:"attributes"`
	}
	SystemAnalysisNodeAttributes struct {
		KnowledgeEntity KnowledgeGraphEntity          `json:"knowledgeEntity"`
		ReferencedAt    time.Time                     `json:"referencedAt"`
		Position        SystemAnalysisDiagramPosition `json:"position"`
		Description     string                        `json:"description"`
	}

	SystemAnalysisDiagramPosition struct {
		X float64  `json:"x"`
		Y float64  `json:"y"`
		Z *float64 `json:"z,omitempty"`
	}

	SystemAnalysisEdge struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes SystemAnalysisEdgeAttributes `json:"attributes"`
	}
	SystemAnalysisEdgeAttributes struct {
		KnowledgeRelationship KnowledgeGraphRelationship `json:"knowledgeRelationship"`
		ReferencedAt          time.Time                  `json:"referencedAt"`
		Description           string                     `json:"description"`
	}
)

var systemAnalysisTags = []string{"System Analysis"}

var GetSystemAnalysis = huma.Operation{
	OperationID: "get-system-analysis",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}",
	Summary:     "Get System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisRequest IdRequest
type GetSystemAnalysisResponse ItemResponse[SystemAnalysis]

var AddSystemAnalysisNode = huma.Operation{
	OperationID: "add-system-analysis-node",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/nodes",
	Summary:     "Add a node to a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisNodeAttributes struct {
	KnowledgeEntityId uuid.UUID                     `json:"knowledgeEntityId"`
	ReferencedAt      *time.Time                    `json:"referencedAt,omitempty"`
	Position          SystemAnalysisDiagramPosition `json:"position"`
	Description       string                        `json:"description"`
}
type AddSystemAnalysisNodeRequest IdRequestWithBody[AddSystemAnalysisNodeAttributes]
type AddSystemAnalysisNodeResponse ItemResponse[SystemAnalysisNode]

var ListSystemAnalysisNodes = huma.Operation{
	OperationID: "list-system-analysis-nodes",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/nodes",
	Summary:     "List nodes in a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type ListSystemAnalysisNodesRequest ListIdRequest
type ListSystemAnalysisNodesResponse ListResponse[SystemAnalysisNode]

var GetSystemAnalysisNode = huma.Operation{
	OperationID: "get-system-analysis-node",
	Method:      http.MethodGet,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Get a system analysis node",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisNodeRequest IdRequest
type GetSystemAnalysisNodeResponse ItemResponse[SystemAnalysisNode]

var UpdateSystemAnalysisNode = huma.Operation{
	OperationID: "update-system-analysis-node",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Update a system analysis node",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisNodeAttributes struct {
	Position    *SystemAnalysisDiagramPosition `json:"position,omitempty"`
	Description *string                        `json:"description,omitempty"`
}
type UpdateSystemAnalysisNodeRequest IdRequestWithBody[UpdateSystemAnalysisNodeAttributes]
type UpdateSystemAnalysisNodeResponse ItemResponse[SystemAnalysisNode]

var DeleteSystemAnalysisNode = huma.Operation{
	OperationID: "delete-system-analysis-node",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Delete a system analysis node",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisNodeRequest IdRequest
type DeleteSystemAnalysisNodeResponse EmptyResponse

var ListSystemAnalysisEdges = huma.Operation{
	OperationID: "list-system-analysis-edges",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/edges",
	Summary:     "List edges in a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type ListSystemAnalysisEdgesRequest ListIdRequest
type ListSystemAnalysisEdgesResponse ListResponse[SystemAnalysisEdge]

var AddSystemAnalysisEdge = huma.Operation{
	OperationID: "add-system-analysis-edge",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/edges",
	Summary:     "Add an edge to a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisEdgeAttributes struct {
	KnowledgeRelationshipId uuid.UUID  `json:"knowledgeRelationshipId"`
	ReferencedAt            *time.Time `json:"referencedAt,omitempty"`
	Description             string     `json:"description"`
}
type AddSystemAnalysisEdgeRequest IdRequestWithBody[AddSystemAnalysisEdgeAttributes]
type AddSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var GetSystemAnalysisEdge = huma.Operation{
	OperationID: "get-system-analysis-edge",
	Method:      http.MethodGet,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Get a system analysis edge",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisEdgeRequest IdRequest
type GetSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var UpdateSystemAnalysisEdge = huma.Operation{
	OperationID: "update-system-analysis-edge",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Update a system analysis edge",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisEdgeAttributes struct {
	Description *string `json:"description,omitempty"`
}
type UpdateSystemAnalysisEdgeRequest IdRequestWithBody[UpdateSystemAnalysisEdgeAttributes]
type UpdateSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var DeleteSystemAnalysisEdge = huma.Operation{
	OperationID: "delete-system-analysis-edge",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Delete a system analysis edge",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEdgeRequest IdRequest
type DeleteSystemAnalysisEdgeResponse EmptyResponse
