package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/schema/schematypes"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type KnowledgeGraphHandler interface {
	ListKnowledgeGraphEntities(context.Context, *ListKnowledgeGraphEntitiesRequest) (*ListKnowledgeGraphEntitiesResponse, error)
	GetKnowledgeGraphEntity(context.Context, *GetKnowledgeGraphEntityRequest) (*GetKnowledgeGraphEntityResponse, error)

	ListKnowledgeGraphRelationships(context.Context, *ListKnowledgeGraphRelationshipsRequest) (*ListKnowledgeGraphRelationshipsResponse, error)
	GetKnowledgeGraphRelationship(context.Context, *GetKnowledgeGraphRelationshipRequest) (*GetKnowledgeGraphRelationshipResponse, error)

	GetKnowledgeGraphView(context.Context, *GetKnowledgeGraphViewRequest) (*GetKnowledgeGraphViewResponse, error)
}

func (o operations) RegisterKnowledgeGraph(api huma.API) {
	huma.Register(api, ListKnowledgeGraphEntities, o.ListKnowledgeGraphEntities)
	huma.Register(api, GetKnowledgeGraphEntity, o.GetKnowledgeGraphEntity)

	huma.Register(api, ListKnowledgeGraphRelationships, o.ListKnowledgeGraphRelationships)
	huma.Register(api, GetKnowledgeGraphRelationship, o.GetKnowledgeGraphRelationship)

	huma.Register(api, GetKnowledgeGraphView, o.GetKnowledgeGraphView)
}

type (
	KnowledgeGraphEvidence struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes KnowledgeGraphEvidenceAttributes `json:"attributes"`
	}

	KnowledgeGraphEvidenceAttributes struct {
		Kind         string                     `json:"kind" enum:"observed,deleted"`
		EffectiveAt  time.Time                  `json:"effectiveAt"`
		SubjectState KnowledgeGraphSubjectState `json:"subjectState"`
	}

	KnowledgeGraphSubjectState struct {
		DisplayName string         `json:"displayName"`
		Description string         `json:"description"`
		Properties  map[string]any `json:"properties"`
	}

	KnowledgeGraphEntity struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes KnowledgeGraphEntityAttributes `json:"attributes"`
	}
	KnowledgeGraphEntityAttributes struct {
		Kind           string                       `json:"kind"`
		LatestEvidence *KnowledgeGraphEvidence      `json:"latestEvidence,omitempty"`
		Aliases        []KnowledgeGraphSubjectAlias `json:"aliases"`
		CreatedAt      time.Time                    `json:"createdAt"`
		UpdatedAt      time.Time                    `json:"updatedAt"`
	}

	KnowledgeGraphRelationship struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes KnowledgeGraphRelationshipAttributes `json:"attributes"`
	}
	KnowledgeGraphRelationshipAttributes struct {
		Kind           string                       `json:"kind"`
		SourceEntityId uuid.UUID                    `json:"sourceEntityId"`
		TargetEntityId uuid.UUID                    `json:"targetEntityId"`
		LatestEvidence *KnowledgeGraphEvidence      `json:"latestEvidence,omitempty"`
		Aliases        []KnowledgeGraphSubjectAlias `json:"aliases"`
		CreatedAt      time.Time                    `json:"createdAt"`
		UpdatedAt      time.Time                    `json:"updatedAt"`
	}

	KnowledgeGraphSubjectAlias struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes KnowledgeGraphSubjectAliasAttributes `json:"attributes"`
	}

	KnowledgeGraphSubjectAliasAttributes struct {
		Kind               string `json:"kind" enum:"entity,relationship"`
		Provider           string `json:"provider"`
		ProviderSource     string `json:"providerSource"`
		ProviderSubjectRef string `json:"providerSubjectRef"`
	}

	KnowledgeGraphView struct {
		RootId        uuid.UUID                    `json:"rootId"`
		Entities      []KnowledgeGraphEntity       `json:"entities"`
		Relationships []KnowledgeGraphRelationship `json:"relationships"`
		Truncated     bool                         `json:"truncated"`
	}
)

func KnowledgeGraphEntityFromEnt(e *ent.KnowledgeEntity) KnowledgeGraphEntity {
	attr := KnowledgeGraphEntityAttributes{
		Kind:      e.Kind,
		Aliases:   nil,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
	if latestEv := e.LatestEvidence(); latestEv != nil {
		attr.LatestEvidence = KnowledgeGraphEvidenceFromEnt(latestEv)
	}

	attr.Aliases = make([]KnowledgeGraphSubjectAlias, len(e.Edges.Aliases))
	for i, alias := range e.Edges.Aliases {
		attr.Aliases[i] = KnowledgeGraphSubjectAliasFromEnt(alias)
	}

	return KnowledgeGraphEntity{Id: e.ID, Attributes: attr}
}

func KnowledgeGraphRelationshipFromEnt(rel *ent.KnowledgeRelationship) KnowledgeGraphRelationship {
	attr := KnowledgeGraphRelationshipAttributes{
		Kind:           rel.Kind,
		SourceEntityId: rel.SourceEntityID,
		TargetEntityId: rel.TargetEntityID,
		CreatedAt:      rel.CreatedAt,
		UpdatedAt:      rel.UpdatedAt,
		Aliases:        nil,
	}
	if latestEv := rel.LatestEvidence(); latestEv != nil {
		attr.LatestEvidence = KnowledgeGraphEvidenceFromEnt(latestEv)
	}

	attr.Aliases = make([]KnowledgeGraphSubjectAlias, len(rel.Edges.Aliases))
	for i, alias := range rel.Edges.Aliases {
		attr.Aliases[i] = KnowledgeGraphSubjectAliasFromEnt(alias)
	}
	return KnowledgeGraphRelationship{Id: rel.ID, Attributes: attr}
}

func KnowledgeGraphSubjectAliasFromEnt(alias *ent.KnowledgeSubjectAlias) KnowledgeGraphSubjectAlias {
	attrs := KnowledgeGraphSubjectAliasAttributes{
		Kind:               alias.SubjectKind.String(),
		Provider:           alias.Provider,
		ProviderSource:     alias.ProviderSource,
		ProviderSubjectRef: alias.ProviderSubjectRef,
	}
	return KnowledgeGraphSubjectAlias{Id: alias.ID, Attributes: attrs}
}

func KnowledgeGraphEvidenceFromEnt(ev *ent.KnowledgeEvidence) *KnowledgeGraphEvidence {
	if ev == nil {
		return nil
	}
	attrs := KnowledgeGraphEvidenceAttributes{
		Kind:         ev.Kind.String(),
		EffectiveAt:  ev.EffectiveAt,
		SubjectState: KnowledgeGraphSubjectStateFromEnt(ev.SubjectState),
	}
	return &KnowledgeGraphEvidence{Id: ev.ID, Attributes: attrs}
}

func KnowledgeGraphSubjectStateFromEnt(s schematypes.KnowledgeGraphSubjectState) KnowledgeGraphSubjectState {
	return KnowledgeGraphSubjectState{
		DisplayName: s.DisplayName,
		Description: s.Description,
		Properties:  s.Properties,
	}
}

func KnowledgeGraphViewFromRez(view *rez.KnowledgeGraphView) KnowledgeGraphView {
	result := KnowledgeGraphView{RootId: view.RootID, Truncated: view.Truncated}
	result.Entities = make([]KnowledgeGraphEntity, len(view.Entities))
	for i, entity := range view.Entities {
		result.Entities[i] = KnowledgeGraphEntityFromEnt(entity)
	}
	result.Relationships = make([]KnowledgeGraphRelationship, len(view.Relationships))
	for i, rel := range view.Relationships {
		result.Relationships[i] = KnowledgeGraphRelationshipFromEnt(rel)
	}
	return result
}

var knowledgeGraphTags = []string{"Knowledge Graph"}

var ListKnowledgeGraphEntities = huma.Operation{
	OperationID: "list-knowledge-graph-entities",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/entities",
	Summary:     "List Knowledge Graph Entities",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeGraphEntitiesRequest struct {
	ListRequest
	Kind           []string `query:"kind" required:"false"`
	Provider       string   `query:"provider" required:"false"`
	ProviderSource string   `query:"providerSource" required:"false"`
	SubjectKind    string   `query:"subjectKind" required:"false"`
}
type ListKnowledgeGraphEntitiesResponse ListResponse[KnowledgeGraphEntity]

var GetKnowledgeGraphEntity = huma.Operation{
	OperationID: "get-knowledge-graph-entity",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/entities/{id}",
	Summary:     "Get Knowledge Graph Entity",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphEntityRequest IdRequest
type GetKnowledgeGraphEntityResponse ItemResponse[KnowledgeGraphEntity]

var ListKnowledgeGraphRelationships = huma.Operation{
	OperationID: "list-knowledge-graph-relationships",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/relationships",
	Summary:     "List Knowledge Graph Relationships",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeGraphRelationshipsRequest struct {
	ListRequest
	Kind           []string  `query:"kind" required:"false"`
	EntityId       uuid.UUID `query:"entityId" required:"false"`
	SourceEntityId uuid.UUID `query:"sourceEntityId" required:"false"`
	TargetEntityId uuid.UUID `query:"targetEntityId" required:"false"`
}
type ListKnowledgeGraphRelationshipsResponse ListResponse[KnowledgeGraphRelationship]

var GetKnowledgeGraphRelationship = huma.Operation{
	OperationID: "get-knowledge-graph-relationship",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/relationships/{id}",
	Summary:     "Get Knowledge Graph Relationship",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphRelationshipRequest IdRequest
type GetKnowledgeGraphRelationshipResponse ItemResponse[KnowledgeGraphRelationship]

var GetKnowledgeGraphView = huma.Operation{
	OperationID: "get-knowledge-graph-view",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/view",
	Summary:     "Get Knowledge Graph View",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphViewRequest struct {
	EntityId         uuid.UUID `query:"entityId" required:"false"`
	Depth            int       `query:"depth" default:"1" minimum:"1" maximum:"4" required:"false"`
	RelationshipKind []string  `query:"relationshipKind" required:"false"`
}
type GetKnowledgeGraphViewResponse ItemResponse[KnowledgeGraphView]
