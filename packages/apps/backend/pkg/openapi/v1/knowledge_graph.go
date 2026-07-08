package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type KnowledgeGraphHandler interface {
	ListKnowledgeGraphEntities(context.Context, *ListKnowledgeGraphEntitiesRequest) (*ListKnowledgeGraphEntitiesResponse, error)
	ListKnowledgeGraphRelationships(context.Context, *ListKnowledgeGraphRelationshipsRequest) (*ListKnowledgeGraphRelationshipsResponse, error)
	GetKnowledgeGraphEntity(context.Context, *GetKnowledgeGraphEntityRequest) (*GetKnowledgeGraphEntityResponse, error)
	CreateKnowledgeGraphSnapshot(context.Context, *CreateKnowledgeGraphSnapshotRequest) (*CreateKnowledgeGraphSnapshotResponse, error)
	GetKnowledgeGraphSnapshot(context.Context, *GetKnowledgeGraphSnapshotRequest) (*GetKnowledgeGraphSnapshotResponse, error)
}

func (o operations) RegisterKnowledgeGraph(api huma.API) {
	huma.Register(api, ListKnowledgeGraphEntities, o.ListKnowledgeGraphEntities)
	huma.Register(api, ListKnowledgeGraphRelationships, o.ListKnowledgeGraphRelationships)
	huma.Register(api, GetKnowledgeGraphEntity, o.GetKnowledgeGraphEntity)
	huma.Register(api, CreateKnowledgeGraphSnapshot, o.CreateKnowledgeGraphSnapshot)
	huma.Register(api, GetKnowledgeGraphSnapshot, o.GetKnowledgeGraphSnapshot)
}

type (
	KnowledgeGraphView struct {
		Entities      []KnowledgeGraphEntity       `json:"entities"`
		Relationships []KnowledgeGraphRelationship `json:"relationships"`
	}

	KnowledgeGraphEntity struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes KnowledgeGraphEntityAttributes `json:"attributes"`
	}
	KnowledgeGraphEntityAttributes struct {
		Kind        string                       `json:"kind"`
		DisplayName string                       `json:"displayName"`
		Description string                       `json:"description"`
		Properties  map[string]any               `json:"properties"`
		Aliases     []KnowledgeGraphSubjectAlias `json:"aliases"`
		CreatedAt   time.Time                    `json:"createdAt"`
		UpdatedAt   time.Time                    `json:"updatedAt"`
	}

	KnowledgeGraphRelationship struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes KnowledgeGraphRelationshipAttributes `json:"attributes"`
	}
	KnowledgeGraphRelationshipAttributes struct {
		Source      Expandable[KnowledgeGraphEntityAttributes] `json:"source"`
		Target      Expandable[KnowledgeGraphEntityAttributes] `json:"target"`
		Kind        string                                     `json:"kind"`
		DisplayName string                                     `json:"displayName"`
		Description string                                     `json:"description"`
		Properties  map[string]any                             `json:"properties"`
		FirstSeenAt time.Time                                  `json:"firstSeenAt"`
		LastSeenAt  time.Time                                  `json:"lastSeenAt"`
		CreatedAt   time.Time                                  `json:"createdAt"`
		UpdatedAt   time.Time                                  `json:"updatedAt"`
	}

	KnowledgeGraphSubjectAlias struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes KnowledgeGraphSubjectAliasAttributes `json:"attributes"`
	}

	KnowledgeGraphSubjectAliasAttributes struct {
		Kind               string `json:"kind" enum:"entity,relationship"`
		Provider           string `json:"provider"`
		ProviderSubjectRef string `json:"providerSubjectRef"`
		Description        string `json:"displayName"`
	}

	KnowledgeGraphSnapshot struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes KnowledgeGraphSnapshotAttributes `json:"attributes"`
	}
	KnowledgeGraphSnapshotAttributes struct {
		AsOf            time.Time                            `json:"asOf"`
		CreatedAt       time.Time                            `json:"createdAt"`
		Scope           string                               `json:"scope"`
		ScopeProperties map[string]any                       `json:"scopeProperties"`
		Entities        []KnowledgeGraphSnapshotEntity       `json:"entities"`
		Relationships   []KnowledgeGraphSnapshotRelationship `json:"relationships"`
	}
	KnowledgeGraphSnapshotEntity struct {
		Id         uuid.UUID                              `json:"id"`
		Attributes KnowledgeGraphSnapshotEntityAttributes `json:"attributes"`
	}
	KnowledgeGraphSnapshotEntityAttributes struct {
		EntityId    *uuid.UUID     `json:"entityId,omitempty"`
		Kind        string         `json:"kind"`
		DisplayName string         `json:"displayName"`
		Description string         `json:"description"`
		Properties  map[string]any `json:"properties"`
	}
	KnowledgeGraphSnapshotRelationship struct {
		Id         uuid.UUID                                    `json:"id"`
		Attributes KnowledgeGraphSnapshotRelationshipAttributes `json:"attributes"`
	}
	KnowledgeGraphSnapshotRelationshipAttributes struct {
		RelationshipId         *uuid.UUID     `json:"relationshipId,omitempty"`
		SourceSnapshotEntityId uuid.UUID      `json:"sourceSnapshotEntityId"`
		TargetSnapshotEntityId uuid.UUID      `json:"targetSnapshotEntityId"`
		Kind                   string         `json:"kind"`
		DisplayName            string         `json:"displayName"`
		Description            string         `json:"description"`
		Properties             map[string]any `json:"properties"`
	}
)

func KnowledgeGraphEntityFromEnt(entity *ent.KnowledgeEntity) KnowledgeGraphEntity {
	attr := KnowledgeGraphEntityAttributes{
		Kind:        entity.Kind,
		DisplayName: entity.DisplayName,
		Description: entity.Description,
		Properties:  entity.LiveProperties,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}

	attr.Aliases = make([]KnowledgeGraphSubjectAlias, len(entity.Edges.Aliases))
	for i, alias := range entity.Edges.Aliases {
		attr.Aliases[i] = KnowledgeGraphSubjectAliasFromEnt(alias)
	}

	return KnowledgeGraphEntity{Id: entity.ID, Attributes: attr}
}

func KnowledgeGraphSubjectAliasFromEnt(alias *ent.KnowledgeSubjectAlias) KnowledgeGraphSubjectAlias {
	attrs := KnowledgeGraphSubjectAliasAttributes{
		Description:        alias.Description,
		Provider:           alias.Provider,
		ProviderSubjectRef: alias.ProviderSubjectRef,
	}
	return KnowledgeGraphSubjectAlias{Id: alias.ID, Attributes: attrs}
}

func KnowledgeGraphRelationshipFromEnt(rel *ent.KnowledgeRelationship) KnowledgeGraphRelationship {
	attr := KnowledgeGraphRelationshipAttributes{
		Kind:        rel.Kind,
		Description: rel.Description,
		Properties:  rel.Properties,
		CreatedAt:   rel.CreatedAt,
		UpdatedAt:   rel.UpdatedAt,
		Source:      Expandable[KnowledgeGraphEntityAttributes]{Id: rel.SourceEntityID},
		Target:      Expandable[KnowledgeGraphEntityAttributes]{Id: rel.TargetEntityID},
	}
	if source, err := rel.Edges.SourceEntityOrErr(); err == nil {
		s := KnowledgeGraphEntityFromEnt(source)
		attr.Source.Attributes = &s.Attributes
	}
	if target, err := rel.Edges.TargetEntityOrErr(); err == nil {
		t := KnowledgeGraphEntityFromEnt(target)
		attr.Target.Attributes = &t.Attributes
	}
	return KnowledgeGraphRelationship{Id: rel.ID, Attributes: attr}
}

func KnowledgeGraphSnapshotFromEnt(snapshot *ent.KnowledgeGraphSnapshot) KnowledgeGraphSnapshot {
	attr := KnowledgeGraphSnapshotAttributes{
		Scope:           snapshot.ScopeKind,
		ScopeProperties: snapshot.ScopeProperties,
		AsOf:            snapshot.AsOf,
		CreatedAt:       snapshot.CreatedAt,
	}
	attr.Entities = make([]KnowledgeGraphSnapshotEntity, len(snapshot.Edges.Entities))
	for i, entity := range snapshot.Edges.Entities {
		attr.Entities[i] = KnowledgeGraphSnapshotEntityFromEnt(entity)
	}
	attr.Relationships = make([]KnowledgeGraphSnapshotRelationship, len(snapshot.Edges.Relationships))
	for i, rel := range snapshot.Edges.Relationships {
		attr.Relationships[i] = KnowledgeGraphSnapshotRelationshipFromEnt(rel)
	}
	return KnowledgeGraphSnapshot{Id: snapshot.ID, Attributes: attr}
}

func KnowledgeGraphSnapshotEntityFromEnt(entity *ent.KnowledgeGraphSnapshotEntity) KnowledgeGraphSnapshotEntity {
	attrs := KnowledgeGraphSnapshotEntityAttributes{
		EntityId:    entity.KnowledgeEntityID,
		Kind:        entity.EntityKind,
		DisplayName: entity.DisplayName,
		Description: entity.Description,
		Properties:  entity.Properties,
	}
	return KnowledgeGraphSnapshotEntity{Id: entity.ID, Attributes: attrs}
}

func KnowledgeGraphSnapshotRelationshipFromEnt(rel *ent.KnowledgeGraphSnapshotRelationship) KnowledgeGraphSnapshotRelationship {
	attrs := KnowledgeGraphSnapshotRelationshipAttributes{
		RelationshipId:         rel.KnowledgeRelationshipID,
		SourceSnapshotEntityId: rel.SourceSnapshotEntityID,
		TargetSnapshotEntityId: rel.TargetSnapshotEntityID,
		Kind:                   rel.RelationshipKind,
		DisplayName:            rel.DisplayName,
		Description:            rel.Description,
		Properties:             rel.Properties,
	}
	return KnowledgeGraphSnapshotRelationship{Id: rel.ID, Attributes: attrs}
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

var GetKnowledgeGraphView = huma.Operation{
	OperationID: "get-knowledge-graph-view",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/view",
	Summary:     "Get Knowledge Graph View",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphViewRequest struct {
	EntityId         uuid.UUID `path:"entityId"`
	Depth            int       `query:"depth" default:"1" minimum:"1" maximum:"4" required:"false"`
	RelationshipKind []string  `query:"relationshipKind" required:"false"`
}
type GetKnowledgeGraphViewResponse ItemResponse[KnowledgeGraphView]

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

var CreateKnowledgeGraphSnapshot = huma.Operation{
	OperationID: "create-knowledge-graph-snapshot",
	Method:      http.MethodPost,
	Path:        "/knowledge_graph/snapshots",
	Summary:     "Create Knowledge Graph Snapshot",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type CreateKnowledgeGraphSnapshotRequestAttributes struct {
	Name              string         `json:"name"`
	AsOf              *time.Time     `json:"asOf,omitempty"`
	Scope             string         `json:"scope" enum:"explicit_entities,root_entities,incident,retrospective,search,analysis"`
	ScopeProperties   map[string]any `json:"scopeProperties"`
	EntityIds         []uuid.UUID    `json:"entityIds"`
	RootEntityIds     []uuid.UUID    `json:"rootEntityIds"`
	Depth             int            `json:"depth"`
	EntityKinds       []string       `json:"entityKinds"`
	RelationshipKinds []string       `json:"relationshipKinds"`
	IncludeIncidents  bool           `json:"includeIncidents"`
	IncludeChanges    bool           `json:"includeChanges"`
	IncludeAlerts     bool           `json:"includeAlerts"`
}
type CreateKnowledgeGraphSnapshotRequest RequestWithBodyAttributes[CreateKnowledgeGraphSnapshotRequestAttributes]
type CreateKnowledgeGraphSnapshotResponse ItemResponse[KnowledgeGraphSnapshot]

var GetKnowledgeGraphSnapshot = huma.Operation{
	OperationID: "get-knowledge-graph-snapshot",
	Method:      http.MethodGet,
	Path:        "/knowledge_graph/snapshots/{id}",
	Summary:     "Get Knowledge Graph Snapshot",
	Tags:        knowledgeGraphTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphSnapshotRequest IdRequest
type GetKnowledgeGraphSnapshotResponse ItemResponse[KnowledgeGraphSnapshot]
