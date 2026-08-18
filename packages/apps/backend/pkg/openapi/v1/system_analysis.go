package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type SystemAnalysisHandler interface {
	GetSystemAnalysis(context.Context, *GetSystemAnalysisRequest) (*GetSystemAnalysisResponse, error)
	UpdateSystemAnalysis(context.Context, *UpdateSystemAnalysisRequest) (*UpdateSystemAnalysisResponse, error)
	GetSystemAnalysisGraph(context.Context, *GetSystemAnalysisGraphRequest) (*GetSystemAnalysisGraphResponse, error)

	ListSystemAnalysisNodes(context.Context, *ListSystemAnalysisNodesRequest) (*ListSystemAnalysisNodesResponse, error)
	AddSystemAnalysisNode(context.Context, *AddSystemAnalysisNodeRequest) (*AddSystemAnalysisNodeResponse, error)
	UpdateSystemAnalysisNode(context.Context, *UpdateSystemAnalysisNodeRequest) (*UpdateSystemAnalysisNodeResponse, error)
	DeleteSystemAnalysisNode(context.Context, *DeleteSystemAnalysisNodeRequest) (*DeleteSystemAnalysisNodeResponse, error)

	ListSystemAnalysisEdges(context.Context, *ListSystemAnalysisEdgesRequest) (*ListSystemAnalysisEdgesResponse, error)
	AddSystemAnalysisEdge(context.Context, *AddSystemAnalysisEdgeRequest) (*AddSystemAnalysisEdgeResponse, error)
	UpdateSystemAnalysisEdge(context.Context, *UpdateSystemAnalysisEdgeRequest) (*UpdateSystemAnalysisEdgeResponse, error)
	DeleteSystemAnalysisEdge(context.Context, *DeleteSystemAnalysisEdgeRequest) (*DeleteSystemAnalysisEdgeResponse, error)

	ListSystemAnalysisEntries(context.Context, *ListSystemAnalysisEntriesRequest) (*ListSystemAnalysisEntriesResponse, error)
	CreateSystemAnalysisEntry(context.Context, *CreateSystemAnalysisEntryRequest) (*CreateSystemAnalysisEntryResponse, error)
	UpdateSystemAnalysisEntry(context.Context, *UpdateSystemAnalysisEntryRequest) (*UpdateSystemAnalysisEntryResponse, error)
	DeleteSystemAnalysisEntry(context.Context, *DeleteSystemAnalysisEntryRequest) (*DeleteSystemAnalysisEntryResponse, error)

	AddSystemAnalysisEntrySubject(context.Context, *AddSystemAnalysisEntrySubjectRequest) (*AddSystemAnalysisEntrySubjectResponse, error)
	UpdateSystemAnalysisEntrySubject(context.Context, *UpdateSystemAnalysisEntrySubjectRequest) (*UpdateSystemAnalysisEntrySubjectResponse, error)
	DeleteSystemAnalysisEntrySubject(context.Context, *DeleteSystemAnalysisEntrySubjectRequest) (*DeleteSystemAnalysisEntrySubjectResponse, error)
}

func (o operations) RegisterSystemAnalysis(api huma.API) {
	huma.Register(api, GetSystemAnalysis, o.GetSystemAnalysis)
	huma.Register(api, UpdateSystemAnalysis, o.UpdateSystemAnalysis)
	huma.Register(api, GetSystemAnalysisGraph, o.GetSystemAnalysisGraph)
	huma.Register(api, ListSystemAnalysisNodes, o.ListSystemAnalysisNodes)
	huma.Register(api, AddSystemAnalysisNode, o.AddSystemAnalysisNode)
	huma.Register(api, UpdateSystemAnalysisNode, o.UpdateSystemAnalysisNode)
	huma.Register(api, DeleteSystemAnalysisNode, o.DeleteSystemAnalysisNode)
	huma.Register(api, ListSystemAnalysisEdges, o.ListSystemAnalysisEdges)
	huma.Register(api, AddSystemAnalysisEdge, o.AddSystemAnalysisEdge)
	huma.Register(api, UpdateSystemAnalysisEdge, o.UpdateSystemAnalysisEdge)
	huma.Register(api, DeleteSystemAnalysisEdge, o.DeleteSystemAnalysisEdge)
	huma.Register(api, ListSystemAnalysisEntries, o.ListSystemAnalysisEntries)
	huma.Register(api, CreateSystemAnalysisEntry, o.CreateSystemAnalysisEntry)
	huma.Register(api, UpdateSystemAnalysisEntry, o.UpdateSystemAnalysisEntry)
	huma.Register(api, DeleteSystemAnalysisEntry, o.DeleteSystemAnalysisEntry)
	huma.Register(api, AddSystemAnalysisEntrySubject, o.AddSystemAnalysisEntrySubject)
	huma.Register(api, UpdateSystemAnalysisEntrySubject, o.UpdateSystemAnalysisEntrySubject)
	huma.Register(api, DeleteSystemAnalysisEntrySubject, o.DeleteSystemAnalysisEntrySubject)
}

type (
	SystemAnalysis struct {
		Id         uuid.UUID                `json:"id"`
		Attributes SystemAnalysisAttributes `json:"attributes"`
	}

	SystemAnalysisAttributes struct {
		ScopeEntityId   *uuid.UUID            `json:"scopeEntityId,omitempty"`
		SubjectEntityId *uuid.UUID            `json:"subjectEntityId,omitempty"`
		ReferenceTime   *time.Time            `json:"referenceTime,omitempty"`
		Nodes           []SystemAnalysisNode  `json:"nodes"`
		Edges           []SystemAnalysisEdge  `json:"edges"`
		Entries         []SystemAnalysisEntry `json:"entries"`
	}
)

type (
	SystemAnalysisNode struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes SystemAnalysisNodeAttributes `json:"attributes"`
	}

	SystemAnalysisNodeAttributes struct {
		KnowledgeEntity     KnowledgeGraphEntity          `json:"knowledgeEntity"`
		Position            SystemAnalysisDiagramPosition `json:"position"`
		Hidden              bool                          `json:"hidden"`
		LabelOverride       *string                       `json:"labelOverride,omitempty"`
		DescriptionOverride *string                       `json:"descriptionOverride,omitempty"`
	}

	SystemAnalysisDiagramPosition struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	SystemAnalysisEdge struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes SystemAnalysisEdgeAttributes `json:"attributes"`
	}

	SystemAnalysisEdgeAttributes struct {
		KnowledgeRelationship KnowledgeGraphRelationship `json:"knowledgeRelationship"`
		SourceNodeId          uuid.UUID                  `json:"sourceNodeId"`
		TargetNodeId          uuid.UUID                  `json:"targetNodeId"`
		Hidden                bool                       `json:"hidden"`
		LabelOverride         *string                    `json:"labelOverride,omitempty"`
		DescriptionOverride   *string                    `json:"descriptionOverride,omitempty"`
	}
)

type (
	SystemAnalysisEntry struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes SystemAnalysisEntryAttributes `json:"attributes"`
	}

	SystemAnalysisEntryAttributes struct {
		Kind       string                       `json:"kind" enum:"observation,context,decision,action,finding,recommendation"`
		OccurredAt *time.Time                   `json:"occurredAt,omitempty"`
		Sequence   int                          `json:"sequence"`
		Title      string                       `json:"title"`
		Body       string                       `json:"body,omitempty"`
		Properties map[string]any               `json:"properties"`
		Subjects   []SystemAnalysisEntrySubject `json:"subjects"`
	}

	SystemAnalysisEntrySubject struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemAnalysisEntrySubjectAttributes `json:"attributes"`
	}

	SystemAnalysisEntrySubjectAttributes struct {
		Role                  string                      `json:"role"`
		SubjectKind           string                      `json:"subjectKind" enum:"entity,relationship,evidence"`
		SubjectId             uuid.UUID                   `json:"subjectId"`
		KnowledgeEntity       *KnowledgeGraphEntity       `json:"knowledgeEntity,omitempty"`
		KnowledgeRelationship *KnowledgeGraphRelationship `json:"knowledgeRelationship,omitempty"`
		KnowledgeEvidence     *KnowledgeGraphEvidence     `json:"knowledgeEvidence,omitempty"`
	}
)

func SystemAnalysisFromEnt(analysis *ent.SystemAnalysis) SystemAnalysis {
	attrs := SystemAnalysisAttributes{
		ScopeEntityId:   analysis.ScopeEntityID,
		SubjectEntityId: analysis.SubjectEntityID,
		ReferenceTime:   analysis.ReferenceTime,
		Nodes:           ConvertSlice(analysis.Edges.AnalysisEntities, SystemAnalysisNodeFromEnt),
		Edges:           ConvertSlice(analysis.Edges.AnalysisRelationships, SystemAnalysisEdgeFromEnt),
		Entries:         ConvertSlice(analysis.Edges.Entries, SystemAnalysisEntryWithSubjectsFromEnt),
	}
	return SystemAnalysis{Id: analysis.ID, Attributes: attrs}
}

func SystemAnalysisNodeFromEnt(node *ent.SystemAnalysisEntity) SystemAnalysisNode {
	position := SystemAnalysisDiagramPosition{}
	if node.PosX != nil {
		position.X = *node.PosX
	}
	if node.PosY != nil {
		position.Y = *node.PosY
	}
	attrs := SystemAnalysisNodeAttributes{
		Position:            position,
		Hidden:              node.Hidden,
		LabelOverride:       node.LabelOverride,
		DescriptionOverride: node.DescriptionOverride,
	}
	if node.Edges.KnowledgeEntity != nil {
		attrs.KnowledgeEntity = KnowledgeGraphEntityFromEnt(node.Edges.KnowledgeEntity)
	}
	return SystemAnalysisNode{Id: node.ID, Attributes: attrs}
}

func SystemAnalysisEdgeFromEnt(edge *ent.SystemAnalysisRelationship) SystemAnalysisEdge {
	attrs := SystemAnalysisEdgeAttributes{
		SourceNodeId:        edge.SourceAnalysisEntityID,
		TargetNodeId:        edge.TargetAnalysisEntityID,
		Hidden:              edge.Hidden,
		LabelOverride:       edge.LabelOverride,
		DescriptionOverride: edge.DescriptionOverride,
	}
	if edge.Edges.KnowledgeRelationship != nil {
		attrs.KnowledgeRelationship = KnowledgeGraphRelationshipFromEnt(edge.Edges.KnowledgeRelationship)
	}
	return SystemAnalysisEdge{Id: edge.ID, Attributes: attrs}
}

func SystemAnalysisEntryFromEnt(entry *ent.SystemAnalysisEntry, subjects []SystemAnalysisEntrySubject) SystemAnalysisEntry {
	properties := entry.Properties
	if properties == nil {
		properties = map[string]any{}
	}
	attrs := SystemAnalysisEntryAttributes{
		Kind:       entry.Kind.String(),
		OccurredAt: entry.OccurredAt,
		Sequence:   entry.Sequence,
		Title:      entry.Title,
		Body:       entry.Body,
		Properties: properties,
		Subjects:   subjects,
	}
	return SystemAnalysisEntry{Id: entry.ID, Attributes: attrs}
}

func SystemAnalysisEntryWithSubjectsFromEnt(entry *ent.SystemAnalysisEntry) SystemAnalysisEntry {
	subjects := ConvertSlice(entry.Edges.Subjects, SystemAnalysisEntrySubjectFromEnt)
	return SystemAnalysisEntryFromEnt(entry, subjects)
}

func SystemAnalysisEntrySubjectFromEnt(subject *ent.SystemAnalysisEntrySubject) SystemAnalysisEntrySubject {
	attrs := SystemAnalysisEntrySubjectAttributes{
		Role: subject.Role,
	}

	if subject.KnowledgeEntityID != nil {
		attrs.SubjectKind = "entity"
		attrs.SubjectId = *subject.KnowledgeEntityID
		if subject.Edges.KnowledgeEntity != nil {
			converted := KnowledgeGraphEntityFromEnt(subject.Edges.KnowledgeEntity)
			attrs.KnowledgeEntity = &converted
		}
	}
	if subject.KnowledgeRelationshipID != nil {
		attrs.SubjectKind = "relationship"
		attrs.SubjectId = *subject.KnowledgeRelationshipID
		if subject.Edges.KnowledgeRelationship != nil {
			converted := KnowledgeGraphRelationshipFromEnt(subject.Edges.KnowledgeRelationship)
			attrs.KnowledgeRelationship = &converted
		}
	}
	if subject.KnowledgeEvidenceID != nil {
		attrs.SubjectKind = "evidence"
		attrs.SubjectId = *subject.KnowledgeEvidenceID
		if subject.Edges.KnowledgeEvidence != nil {
			attrs.KnowledgeEvidence = KnowledgeGraphEvidenceFromEnt(subject.Edges.KnowledgeEvidence)
		}
	}

	return SystemAnalysisEntrySubject{Id: subject.ID, Attributes: attrs}
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

type GetSystemAnalysisRequest IdRequest
type GetSystemAnalysisResponse ItemResponse[SystemAnalysis]

var UpdateSystemAnalysis = huma.Operation{
	OperationID: "update-system-analysis",
	Method:      http.MethodPatch,
	Path:        "/system_analysis/{id}",
	Summary:     "Update System Analysis",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisAttributes struct {
	ScopeEntityId   *uuid.UUID `json:"scopeEntityId,omitempty"`
	SubjectEntityId *uuid.UUID `json:"subjectEntityId,omitempty"`
	ReferenceTime   *time.Time `json:"referenceTime,omitempty"`
}
type UpdateSystemAnalysisRequest IdRequestWithBody[UpdateSystemAnalysisAttributes]
type UpdateSystemAnalysisResponse ItemResponse[SystemAnalysis]

var GetSystemAnalysisGraph = huma.Operation{
	OperationID: "get-system-analysis-graph",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/graph",
	Summary:     "Get System Analysis Graph",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type GetSystemAnalysisGraphRequest struct {
	Id               uuid.UUID `path:"id"`
	Depth            int       `query:"depth" required:"false"`
	RelationshipKind []string  `query:"relationshipKind" required:"false"`
}
type GetSystemAnalysisGraphResponse ItemResponse[KnowledgeGraphView]

var ListSystemAnalysisNodes = huma.Operation{
	OperationID: "list-system-analysis-nodes",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/nodes",
	Summary:     "List System Analysis Nodes",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type ListSystemAnalysisNodesRequest ListIdRequest
type ListSystemAnalysisNodesResponse ListResponse[SystemAnalysisNode]

var AddSystemAnalysisNode = huma.Operation{
	OperationID: "add-system-analysis-node",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/nodes",
	Summary:     "Add System Analysis Node",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisNodeAttributes struct {
	KnowledgeEntityId   uuid.UUID                     `json:"knowledgeEntityId"`
	Position            SystemAnalysisDiagramPosition `json:"position"`
	Hidden              *bool                         `json:"hidden,omitempty"`
	LabelOverride       *string                       `json:"labelOverride,omitempty"`
	DescriptionOverride *string                       `json:"descriptionOverride,omitempty"`
}
type AddSystemAnalysisNodeRequest IdRequestWithBody[AddSystemAnalysisNodeAttributes]
type AddSystemAnalysisNodeResponse ItemResponse[SystemAnalysisNode]

var UpdateSystemAnalysisNode = huma.Operation{
	OperationID: "update-system-analysis-node",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Update System Analysis Node",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisNodeAttributes struct {
	Position            *SystemAnalysisDiagramPosition `json:"position,omitempty"`
	Hidden              *bool                          `json:"hidden,omitempty"`
	LabelOverride       *string                        `json:"labelOverride,omitempty"`
	DescriptionOverride *string                        `json:"descriptionOverride,omitempty"`
}
type UpdateSystemAnalysisNodeRequest IdRequestWithBody[UpdateSystemAnalysisNodeAttributes]
type UpdateSystemAnalysisNodeResponse ItemResponse[SystemAnalysisNode]

var DeleteSystemAnalysisNode = huma.Operation{
	OperationID: "delete-system-analysis-node",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_nodes/{id}",
	Summary:     "Delete System Analysis Node",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisNodeRequest IdRequest
type DeleteSystemAnalysisNodeResponse EmptyResponse

var ListSystemAnalysisEdges = huma.Operation{
	OperationID: "list-system-analysis-edges",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/edges",
	Summary:     "List System Analysis Edges",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type ListSystemAnalysisEdgesRequest ListIdRequest
type ListSystemAnalysisEdgesResponse ListResponse[SystemAnalysisEdge]

var AddSystemAnalysisEdge = huma.Operation{
	OperationID: "add-system-analysis-edge",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/edges",
	Summary:     "Add System Analysis Edge",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisEdgeAttributes struct {
	KnowledgeRelationshipId uuid.UUID `json:"knowledgeRelationshipId"`
	Hidden                  *bool     `json:"hidden,omitempty"`
	LabelOverride           *string   `json:"labelOverride,omitempty"`
	DescriptionOverride     *string   `json:"descriptionOverride,omitempty"`
}
type AddSystemAnalysisEdgeRequest IdRequestWithBody[AddSystemAnalysisEdgeAttributes]
type AddSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var UpdateSystemAnalysisEdge = huma.Operation{
	OperationID: "update-system-analysis-edge",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Update System Analysis Edge",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisEdgeAttributes struct {
	Hidden              *bool   `json:"hidden,omitempty"`
	LabelOverride       *string `json:"labelOverride,omitempty"`
	DescriptionOverride *string `json:"descriptionOverride,omitempty"`
}
type UpdateSystemAnalysisEdgeRequest IdRequestWithBody[UpdateSystemAnalysisEdgeAttributes]
type UpdateSystemAnalysisEdgeResponse ItemResponse[SystemAnalysisEdge]

var DeleteSystemAnalysisEdge = huma.Operation{
	OperationID: "delete-system-analysis-edge",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_edges/{id}",
	Summary:     "Delete System Analysis Edge",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEdgeRequest IdRequest
type DeleteSystemAnalysisEdgeResponse EmptyResponse

var ListSystemAnalysisEntries = huma.Operation{
	OperationID: "list-system-analysis-entries",
	Method:      http.MethodGet,
	Path:        "/system_analysis/{id}/entries",
	Summary:     "List System Analysis Entries",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type ListSystemAnalysisEntriesRequest ListIdRequest
type ListSystemAnalysisEntriesResponse ListResponse[SystemAnalysisEntry]

var CreateSystemAnalysisEntry = huma.Operation{
	OperationID: "create-system-analysis-entry",
	Method:      http.MethodPost,
	Path:        "/system_analysis/{id}/entries",
	Summary:     "Create System Analysis Entry",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type CreateSystemAnalysisEntryAttributes struct {
	Kind       string         `json:"kind" enum:"observation,context,decision,action,finding,recommendation"`
	OccurredAt *time.Time     `json:"occurredAt,omitempty"`
	Sequence   int            `json:"sequence"`
	Title      string         `json:"title"`
	Body       *string        `json:"body,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}
type CreateSystemAnalysisEntryRequest IdRequestWithBody[CreateSystemAnalysisEntryAttributes]
type CreateSystemAnalysisEntryResponse ItemResponse[SystemAnalysisEntry]

var UpdateSystemAnalysisEntry = huma.Operation{
	OperationID: "update-system-analysis-entry",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_entries/{id}",
	Summary:     "Update System Analysis Entry",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisEntryAttributes struct {
	Kind       *string        `json:"kind,omitempty" enum:"observation,context,decision,action,finding,recommendation"`
	OccurredAt *time.Time     `json:"occurredAt,omitempty"`
	Sequence   *int           `json:"sequence,omitempty"`
	Title      *string        `json:"title,omitempty"`
	Body       *string        `json:"body,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}
type UpdateSystemAnalysisEntryRequest IdRequestWithBody[UpdateSystemAnalysisEntryAttributes]
type UpdateSystemAnalysisEntryResponse ItemResponse[SystemAnalysisEntry]

var DeleteSystemAnalysisEntry = huma.Operation{
	OperationID: "delete-system-analysis-entry",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_entries/{id}",
	Summary:     "Delete System Analysis Entry",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEntryRequest IdRequest
type DeleteSystemAnalysisEntryResponse EmptyResponse

var AddSystemAnalysisEntrySubject = huma.Operation{
	OperationID: "add-system-analysis-entry-subject",
	Method:      http.MethodPost,
	Path:        "/system_analysis_entries/{id}/subjects",
	Summary:     "Add System Analysis Entry Subject",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type AddSystemAnalysisEntrySubjectAttributes struct {
	Role                    string     `json:"role"`
	KnowledgeEntityId       *uuid.UUID `json:"knowledgeEntityId,omitempty"`
	KnowledgeRelationshipId *uuid.UUID `json:"knowledgeRelationshipId,omitempty"`
	KnowledgeEvidenceId     *uuid.UUID `json:"knowledgeEvidenceId,omitempty"`
}
type AddSystemAnalysisEntrySubjectRequest IdRequestWithBody[AddSystemAnalysisEntrySubjectAttributes]
type AddSystemAnalysisEntrySubjectResponse ItemResponse[SystemAnalysisEntrySubject]

var UpdateSystemAnalysisEntrySubject = huma.Operation{
	OperationID: "update-system-analysis-entry-subject",
	Method:      http.MethodPatch,
	Path:        "/system_analysis_entry_subjects/{id}",
	Summary:     "Update System Analysis Entry Subject",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type UpdateSystemAnalysisEntrySubjectAttributes struct {
	Role string `json:"role"`
}
type UpdateSystemAnalysisEntrySubjectRequest IdRequestWithBody[UpdateSystemAnalysisEntrySubjectAttributes]
type UpdateSystemAnalysisEntrySubjectResponse ItemResponse[SystemAnalysisEntrySubject]

var DeleteSystemAnalysisEntrySubject = huma.Operation{
	OperationID: "delete-system-analysis-entry-subject",
	Method:      http.MethodDelete,
	Path:        "/system_analysis_entry_subjects/{id}",
	Summary:     "Delete System Analysis Entry Subject",
	Tags:        systemAnalysisTags,
	Errors:      ErrorCodes(),
}

type DeleteSystemAnalysisEntrySubjectRequest IdRequest
type DeleteSystemAnalysisEntrySubjectResponse EmptyResponse
