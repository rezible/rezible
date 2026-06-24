package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type SystemTopologyHandler interface {
	ListSystemTopologyEntities(context.Context, *ListSystemTopologyEntitiesRequest) (*ListSystemTopologyEntitiesResponse, error)
	GetSystemTopologyEntity(context.Context, *GetSystemTopologyEntityRequest) (*GetSystemTopologyEntityResponse, error)
	GetSystemTopologyEntityNeighborhood(context.Context, *GetSystemTopologyEntityNeighborhoodRequest) (*GetSystemTopologyEntityNeighborhoodResponse, error)
	ListSystemTopologyRelationships(context.Context, *ListSystemTopologyRelationshipsRequest) (*ListSystemTopologyRelationshipsResponse, error)
	CreateSystemTopologySnapshot(context.Context, *CreateSystemTopologySnapshotRequest) (*CreateSystemTopologySnapshotResponse, error)
	GetSystemTopologySnapshot(context.Context, *GetSystemTopologySnapshotRequest) (*GetSystemTopologySnapshotResponse, error)
}

func (o operations) RegisterSystemTopology(api huma.API) {
	huma.Register(api, ListSystemTopologyEntities, o.ListSystemTopologyEntities)
	huma.Register(api, GetSystemTopologyEntity, o.GetSystemTopologyEntity)
	huma.Register(api, GetSystemTopologyEntityNeighborhood, o.GetSystemTopologyEntityNeighborhood)
	huma.Register(api, ListSystemTopologyRelationships, o.ListSystemTopologyRelationships)
	huma.Register(api, CreateSystemTopologySnapshot, o.CreateSystemTopologySnapshot)
	huma.Register(api, GetSystemTopologySnapshot, o.GetSystemTopologySnapshot)
}

type (
	SystemTopologyEntity struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes SystemTopologyEntityAttributes `json:"attributes"`
	}
	SystemTopologyEntityAttributes struct {
		Kind          string                       `json:"kind"`
		DisplayName   string                       `json:"displayName"`
		Description   string                       `json:"description"`
		Properties    map[string]any               `json:"properties"`
		Aliases       []SystemTopologyEntityAlias  `json:"aliases"`
		Relationships []SystemTopologyRelationship `json:"relationships,omitempty"`
		CreatedAt     time.Time                    `json:"createdAt"`
		UpdatedAt     time.Time                    `json:"updatedAt"`
	}
	SystemTopologyEntityAlias struct {
		Id             uuid.UUID `json:"id"`
		Provider       string    `json:"provider"`
		ProviderSource string    `json:"providerSource"`
		SubjectKind    string    `json:"subjectKind"`
		SubjectRef     string    `json:"subjectRef"`
		FirstSeenAt    time.Time `json:"firstSeenAt"`
		LastSeenAt     time.Time `json:"lastSeenAt"`
	}

	SystemTopologyRelationship struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes SystemTopologyRelationshipAttributes `json:"attributes"`
	}
	SystemTopologyRelationshipAttributes struct {
		SourceEntityId uuid.UUID             `json:"sourceEntityId"`
		TargetEntityId uuid.UUID             `json:"targetEntityId"`
		Kind           string                `json:"kind"`
		DisplayName    string                `json:"displayName"`
		Description    string                `json:"description"`
		Properties     map[string]any        `json:"properties"`
		Source         *SystemTopologyEntity `json:"source,omitempty"`
		Target         *SystemTopologyEntity `json:"target,omitempty"`
		FirstSeenAt    time.Time             `json:"firstSeenAt"`
		LastSeenAt     time.Time             `json:"lastSeenAt"`
		CreatedAt      time.Time             `json:"createdAt"`
		UpdatedAt      time.Time             `json:"updatedAt"`
	}

	SystemTopologyGraph struct {
		Entities      []SystemTopologyEntity       `json:"entities"`
		Relationships []SystemTopologyRelationship `json:"relationships"`
	}

	SystemTopologySnapshot struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes SystemTopologySnapshotAttributes `json:"attributes"`
	}
	SystemTopologySnapshotAttributes struct {
		Name            string                               `json:"name"`
		AsOf            time.Time                            `json:"asOf"`
		Scope           string                               `json:"scope"`
		ScopeProperties map[string]any                       `json:"scopeProperties"`
		Entities        []SystemTopologySnapshotEntity       `json:"entities"`
		Relationships   []SystemTopologySnapshotRelationship `json:"relationships"`
		CreatedAt       time.Time                            `json:"createdAt"`
	}
	SystemTopologySnapshotEntity struct {
		Id         uuid.UUID                              `json:"id"`
		Attributes SystemTopologySnapshotEntityAttributes `json:"attributes"`
	}
	SystemTopologySnapshotEntityAttributes struct {
		KnowledgeEntityId *uuid.UUID       `json:"knowledgeEntityId,omitempty"`
		Kind              string           `json:"kind"`
		DisplayName       string           `json:"displayName"`
		Description       string           `json:"description"`
		Properties        map[string]any   `json:"properties"`
		Aliases           []map[string]any `json:"aliases"`
	}
	SystemTopologySnapshotRelationship struct {
		Id         uuid.UUID                                    `json:"id"`
		Attributes SystemTopologySnapshotRelationshipAttributes `json:"attributes"`
	}
	SystemTopologySnapshotRelationshipAttributes struct {
		KnowledgeRelationshipId *uuid.UUID     `json:"knowledgeRelationshipId,omitempty"`
		SourceSnapshotEntityId  uuid.UUID      `json:"sourceSnapshotEntityId"`
		TargetSnapshotEntityId  uuid.UUID      `json:"targetSnapshotEntityId"`
		Kind                    string         `json:"kind"`
		DisplayName             string         `json:"displayName"`
		Description             string         `json:"description"`
		Properties              map[string]any `json:"properties"`
	}
)

func SystemTopologyEntityFromEnt(entity *ent.KnowledgeEntity) SystemTopologyEntity {
	attr := SystemTopologyEntityAttributes{
		Kind:        entity.Kind,
		DisplayName: entity.DisplayName,
		Description: entity.Description,
		Properties:  entity.Properties,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}

	attr.Aliases = make([]SystemTopologyEntityAlias, len(entity.Edges.Aliases))
	for i, alias := range entity.Edges.Aliases {
		attr.Aliases[i] = SystemTopologyEntityAliasFromEnt(alias)
	}

	numSourceRels := len(entity.Edges.SourceRelationships)
	attr.Relationships = make([]SystemTopologyRelationship, numSourceRels+len(entity.Edges.TargetRelationships))
	for i, rel := range entity.Edges.SourceRelationships {
		attr.Relationships[i] = SystemTopologyRelationshipFromEnt(rel)
	}
	for i, rel := range entity.Edges.TargetRelationships {
		attr.Relationships[numSourceRels+i] = SystemTopologyRelationshipFromEnt(rel)
	}

	return SystemTopologyEntity{Id: entity.ID, Attributes: attr}
}

func SystemTopologyEntityAliasFromEnt(alias *ent.KnowledgeEntityAlias) SystemTopologyEntityAlias {
	return SystemTopologyEntityAlias{
		Id: alias.ID,
		//Provider:       alias.Provider,
		//ProviderSource: alias.ProviderSource,
		//SubjectKind:    alias.SubjectKind,
		//SubjectRef:     alias.SubjectRef,
		//FirstSeenAt:    alias.FirstSeenAt,
		//LastSeenAt:     alias.LastSeenAt,
	}
}

func SystemTopologyRelationshipFromEnt(rel *ent.KnowledgeRelationship) SystemTopologyRelationship {
	attr := SystemTopologyRelationshipAttributes{
		SourceEntityId: rel.SourceEntityID,
		TargetEntityId: rel.TargetEntityID,
		Kind:           rel.Kind,
		DisplayName:    rel.DisplayName,
		Description:    rel.Description,
		Properties:     rel.Properties,
		CreatedAt:      rel.CreatedAt,
		UpdatedAt:      rel.UpdatedAt,
	}
	if source, err := rel.Edges.SourceEntityOrErr(); err == nil {
		attr.Source = new(SystemTopologyEntityFromEnt(source))
	}
	if target, err := rel.Edges.TargetEntityOrErr(); err == nil {
		attr.Target = new(SystemTopologyEntityFromEnt(target))
	}
	return SystemTopologyRelationship{Id: rel.ID, Attributes: attr}
}

func SystemTopologySnapshotFromEnt(snapshot *ent.SystemTopologySnapshot) SystemTopologySnapshot {
	attr := SystemTopologySnapshotAttributes{
		Name:            snapshot.Name,
		AsOf:            snapshot.AsOf,
		Scope:           snapshot.Scope.String(),
		ScopeProperties: snapshot.ScopeProperties,
		CreatedAt:       snapshot.CreatedAt,
	}
	attr.Entities = make([]SystemTopologySnapshotEntity, len(snapshot.Edges.Entities))
	for i, entity := range snapshot.Edges.Entities {
		attr.Entities[i] = SystemTopologySnapshotEntityFromEnt(entity)
	}
	attr.Relationships = make([]SystemTopologySnapshotRelationship, len(snapshot.Edges.Relationships))
	for i, rel := range snapshot.Edges.Relationships {
		attr.Relationships[i] = SystemTopologySnapshotRelationshipFromEnt(rel)
	}
	return SystemTopologySnapshot{Id: snapshot.ID, Attributes: attr}
}

func SystemTopologySnapshotEntityFromEnt(entity *ent.SystemTopologySnapshotEntity) SystemTopologySnapshotEntity {
	return SystemTopologySnapshotEntity{
		Id: entity.ID,
		Attributes: SystemTopologySnapshotEntityAttributes{
			KnowledgeEntityId: entity.KnowledgeEntityID,
			Kind:              entity.EntityKind,
			DisplayName:       entity.DisplayName,
			Description:       entity.Description,
			Properties:        entity.Properties,
			Aliases:           entity.Aliases,
		},
	}
}

func SystemTopologySnapshotRelationshipFromEnt(rel *ent.SystemTopologySnapshotRelationship) SystemTopologySnapshotRelationship {
	return SystemTopologySnapshotRelationship{
		Id: rel.ID,
		Attributes: SystemTopologySnapshotRelationshipAttributes{
			KnowledgeRelationshipId: rel.KnowledgeRelationshipID,
			SourceSnapshotEntityId:  rel.SourceSnapshotEntityID,
			TargetSnapshotEntityId:  rel.TargetSnapshotEntityID,
			Kind:                    rel.RelationshipKind,
			DisplayName:             rel.DisplayName,
			Description:             rel.Description,
			Properties:              rel.Properties,
		},
	}
}

var topologyTags = []string{"SystemTopology"}

var ListSystemTopologyEntities = huma.Operation{
	OperationID: "list-system-topology-entities",
	Method:      http.MethodGet,
	Path:        "/system_topology/entities",
	Summary:     "List System Topology Entities",
	Tags:        topologyTags,
	Errors:      ErrorCodes(),
}

type ListSystemTopologyEntitiesRequest struct {
	ListRequest
	Kind           []string `query:"kind" required:"false"`
	Provider       string   `query:"provider" required:"false"`
	ProviderSource string   `query:"providerSource" required:"false"`
	SubjectKind    string   `query:"subjectKind" required:"false"`
}
type ListSystemTopologyEntitiesResponse ListResponse[SystemTopologyEntity]

var GetSystemTopologyEntity = huma.Operation{
	OperationID: "get-system-topology-entity",
	Method:      http.MethodGet,
	Path:        "/system_topology/entities/{id}",
	Summary:     "Get System Topology Entity",
	Tags:        topologyTags,
	Errors:      ErrorCodes(),
}

type GetSystemTopologyEntityRequest EmptyIdRequest
type GetSystemTopologyEntityResponse ItemResponse[SystemTopologyEntity]

var GetSystemTopologyEntityNeighborhood = huma.Operation{
	OperationID: "get-system-topology-entity-neighborhood",
	Method:      http.MethodGet,
	Path:        "/system_topology/entities/{id}/neighborhood",
	Summary:     "Get System Topology Entity Neighborhood",
	Tags:        topologyTags,
	Errors:      ErrorCodes(),
}

type GetSystemTopologyEntityNeighborhoodRequest struct {
	Id               uuid.UUID `path:"id"`
	Depth            int       `query:"depth" default:"1" minimum:"1" maximum:"4" required:"false"`
	RelationshipKind []string  `query:"relationshipKind" required:"false"`
}
type GetSystemTopologyEntityNeighborhoodResponse ItemResponse[SystemTopologyGraph]

var ListSystemTopologyRelationships = huma.Operation{
	OperationID: "list-system-topology-relationships",
	Method:      http.MethodGet,
	Path:        "/system_topology/relationships",
	Summary:     "List System Topology Relationships",
	Tags:        topologyTags,
	Errors:      ErrorCodes(),
}

type ListSystemTopologyRelationshipsRequest struct {
	ListRequest
	Kind           []string  `query:"kind" required:"false"`
	EntityId       uuid.UUID `query:"entityId" required:"false"`
	SourceEntityId uuid.UUID `query:"sourceEntityId" required:"false"`
	TargetEntityId uuid.UUID `query:"targetEntityId" required:"false"`
}
type ListSystemTopologyRelationshipsResponse ListResponse[SystemTopologyRelationship]

var CreateSystemTopologySnapshot = huma.Operation{
	OperationID: "create-system-topology-snapshot",
	Method:      http.MethodPost,
	Path:        "/system_topology/snapshots",
	Summary:     "Create System Topology Snapshot",
	Tags:        topologyTags,
	Errors:      ErrorCodes(),
}

type CreateSystemTopologySnapshotAttributes struct {
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
type CreateSystemTopologySnapshotRequest RequestWithBodyAttributes[CreateSystemTopologySnapshotAttributes]
type CreateSystemTopologySnapshotResponse ItemResponse[SystemTopologySnapshot]

var GetSystemTopologySnapshot = huma.Operation{
	OperationID: "get-system-topology-snapshot",
	Method:      http.MethodGet,
	Path:        "/system_topology/snapshots/{id}",
	Summary:     "Get System Topology Snapshot",
	Tags:        topologyTags,
	Errors:      ErrorCodes(),
}

type GetSystemTopologySnapshotRequest EmptyIdRequest
type GetSystemTopologySnapshotResponse ItemResponse[SystemTopologySnapshot]

func CreateSystemTopologySnapshotParamsFromAttributes(attr CreateSystemTopologySnapshotAttributes) rez.CreateSystemTopologySnapshotParams {
	var asOf time.Time
	if attr.AsOf != nil {
		asOf = *attr.AsOf
	}
	return rez.CreateSystemTopologySnapshotParams{
		Name:              attr.Name,
		AsOf:              asOf,
		Scope:             attr.Scope,
		ScopeProperties:   attr.ScopeProperties,
		EntityIDs:         attr.EntityIds,
		RootEntityIDs:     attr.RootEntityIds,
		Depth:             attr.Depth,
		EntityKinds:       attr.EntityKinds,
		RelationshipKinds: attr.RelationshipKinds,
		IncludeIncidents:  attr.IncludeIncidents,
		IncludeChanges:    attr.IncludeChanges,
		IncludeAlerts:     attr.IncludeAlerts,
	}
}
