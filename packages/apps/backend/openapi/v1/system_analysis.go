package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
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
		TopologySnapshot *SystemTopologySnapshot `json:"topologySnapshot,omitempty"`
		Nodes            []SystemAnalysisNode    `json:"nodes"`
		Edges            []SystemAnalysisEdge    `json:"edges"`
	}

	SystemAnalysisNode struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes SystemAnalysisNodeAttributes `json:"attributes"`
	}
	SystemAnalysisNodeAttributes struct {
		SnapshotEntity SystemTopologySnapshotEntity  `json:"snapshotEntity"`
		Position       SystemAnalysisDiagramPosition `json:"position"`
		Description    string                        `json:"description"`
	}

	SystemAnalysisDiagramPosition struct {
		X float64  `json:"x"`
		Y float64  `json:"y"`
		Z *float64 `json:"z,omitempty"`
	}

	SystemAnalysisEdge struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemAnalysisTopologyEdgeAttributes `json:"attributes"`
	}
	SystemAnalysisTopologyEdgeAttributes struct {
		SnapshotRelationship SystemTopologySnapshotRelationship `json:"snapshotRelationship"`
		Description          string                             `json:"description"`
	}
)

func SystemAnalysisFromEnt(sc *ent.SystemAnalysis) SystemAnalysis {
	attr := SystemAnalysisAttributes{}

	if snapshot, err := sc.Edges.TopologySnapshotOrErr(); err == nil {
		attr.TopologySnapshot = new(SystemTopologySnapshotFromEnt(snapshot))
	}

	attr.Nodes = make([]SystemAnalysisNode, len(sc.Edges.AnalysisNodes))
	for i, node := range sc.Edges.AnalysisNodes {
		attr.Nodes[i] = SystemAnalysisNodeFromEnt(node)
	}

	attr.Edges = make([]SystemAnalysisEdge, len(sc.Edges.AnalysisEdges))
	for i, edge := range sc.Edges.AnalysisEdges {
		attr.Edges[i] = SystemAnalysisEdgeFromEnt(edge)
	}

	return SystemAnalysis{Id: sc.ID, Attributes: attr}
}

func SystemAnalysisNodeFromEnt(node *ent.SystemAnalysisTopologyNode) SystemAnalysisNode {
	attr := SystemAnalysisNodeAttributes{
		Position: SystemAnalysisDiagramPosition{
			X: node.PosX,
			Y: node.PosY,
			Z: nil,
		},
		Description: node.Description,
	}

	if snapshotEntity, err := node.Edges.SnapshotEntityOrErr(); err == nil {
		attr.SnapshotEntity = SystemTopologySnapshotEntityFromEnt(snapshotEntity)
	}

	return SystemAnalysisNode{Id: node.ID, Attributes: attr}
}

func SystemAnalysisEdgeFromEnt(edge *ent.SystemAnalysisTopologyEdge) SystemAnalysisEdge {
	attr := SystemAnalysisTopologyEdgeAttributes{
		Description: edge.Description,
	}
	if snapshotRelationship, err := edge.Edges.SnapshotRelationshipOrErr(); err == nil {
		attr.SnapshotRelationship = SystemTopologySnapshotRelationshipFromEnt(snapshotRelationship)
	}

	return SystemAnalysisEdge{Id: edge.ID, Attributes: attr}
}

var systemAnalysisTags = []string{"System Analysis"}

var GetSystemAnalysis = huma.Operation{
	OperationID: "get-system-analysis",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}",
	Summary:     "Get System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisRequest EmptyIdRequest
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
	SnapshotEntityId  *uuid.UUID                    `json:"snapshotEntityId,omitempty"`
	KnowledgeEntityId *uuid.UUID                    `json:"knowledgeEntityId,omitempty"`
	Position          SystemAnalysisDiagramPosition `json:"position"`
	Description       string                        `json:"description"`
}
type AddSystemAnalysisNodeRequest IdRequest[AddSystemAnalysisNodeAttributes]
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
type ListSystemAnalysisNodesResponse PaginatedResponse[SystemAnalysisNode]

var GetSystemAnalysisNode = huma.Operation{
	OperationID: "get-system-analysis-node",
	Method:      http.MethodGet,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Get a node in a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisNodeRequest EmptyIdRequest
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
type UpdateSystemAnalysisNodeRequest IdRequest[UpdateSystemAnalysisNodeAttributes]
type UpdateSystemAnalysisNodeResponse ItemResponse[SystemAnalysisNode]

var DeleteSystemAnalysisNode = huma.Operation{
	OperationID: "delete-system-analysis-node",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Delete a node from a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisNodeRequest EmptyIdRequest
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
type ListSystemAnalysisEdgesResponse PaginatedResponse[SystemAnalysisEdge]

var AddSystemAnalysisEdge = huma.Operation{
	OperationID: "add-system-analysis-edge",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/edges",
	Summary:     "Add an edge to a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisEdgeAttributes struct {
	SnapshotRelationshipId uuid.UUID `json:"snapshotRelationshipId"`
	Description            string    `json:"description"`
}
type AddSystemAnalysisEdgeRequest IdRequest[AddSystemAnalysisEdgeAttributes]
type AddSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var GetSystemAnalysisEdge = huma.Operation{
	OperationID: "get-system-analysis-edge",
	Method:      http.MethodGet,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Get an edge in a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisEdgeRequest EmptyIdRequest
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
type UpdateSystemAnalysisEdgeRequest IdRequest[UpdateSystemAnalysisEdgeAttributes]
type UpdateSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var DeleteSystemAnalysisEdge = huma.Operation{
	OperationID: "delete-system-analysis-edge",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Delete an edge from a system analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEdgeRequest EmptyIdRequest
type DeleteSystemAnalysisEdgeResponse EmptyResponse
