package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
)

type KnowledgeGraphHandler interface {
	ListKnowledgeGraphEntities(context.Context, *ListKnowledgeGraphEntitiesRequest) (*ListKnowledgeGraphEntitiesResponse, error)
	GetKnowledgeGraphEntity(context.Context, *GetKnowledgeGraphEntityRequest) (*GetKnowledgeGraphEntityResponse, error)

	ListKnowledgeGraphRelationships(context.Context, *ListKnowledgeGraphRelationshipsRequest) (*ListKnowledgeGraphRelationshipsResponse, error)
	GetKnowledgeGraphRelationship(context.Context, *GetKnowledgeGraphRelationshipRequest) (*GetKnowledgeGraphRelationshipResponse, error)

	ListKnowledgeGraphSubjectAliases(context.Context, *ListKnowledgeGraphSubjectAliasesRequest) (*ListKnowledgeGraphSubjectAliasesResponse, error)
	GetKnowledgeGraphSubjectAliases(context.Context, *GetKnowledgeGraphSubjectAliasRequest) (*GetKnowledgeGraphSubjectAliasResponse, error)

	ListKnowledgeGraphEvidence(context.Context, *ListKnowledgeGraphEvidenceRequest) (*ListKnowledgeGraphEvidenceResponse, error)
	GetKnowledgeGraphEvidence(context.Context, *GetKnowledgeGraphEvidenceRequest) (*GetKnowledgeGraphEvidenceResponse, error)

	SelectKnowledgeGraphEntities(context.Context, *SelectKnowledgeGraphEntitiesRequest) (*SelectKnowledgeGraphEntitiesResponse, error)
	ExpandKnowledgeGraphRelationships(context.Context, *ExpandKnowledgeGraphRelationshipsRequest) (*ExpandKnowledgeGraphRelationshipsResponse, error)
}

func (o operations) RegisterKnowledgeGraph(api huma.API) {
	huma.Register(api, ListKnowledgeGraphEntities, o.ListKnowledgeGraphEntities)
	huma.Register(api, GetKnowledgeGraphEntity, o.GetKnowledgeGraphEntity)

	huma.Register(api, ListKnowledgeGraphRelationships, o.ListKnowledgeGraphRelationships)
	huma.Register(api, GetKnowledgeGraphRelationship, o.GetKnowledgeGraphRelationship)

	huma.Register(api, ListKnowledgeGraphSubjectAliases, o.ListKnowledgeGraphSubjectAliases)
	huma.Register(api, GetKnowledgeGraphSubjectAlias, o.GetKnowledgeGraphSubjectAliases)

	huma.Register(api, ListKnowledgeGraphEvidence, o.ListKnowledgeGraphEvidence)
	huma.Register(api, GetKnowledgeGraphEvidence, o.GetKnowledgeGraphEvidence)

	huma.Register(api, SelectKnowledgeGraphEntities, o.SelectKnowledgeGraphEntities)
	huma.Register(api, ExpandKnowledgeGraphRelationships, o.ExpandKnowledgeGraphRelationships)
}

type (
	KnowledgeGraphEntity struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes KnowledgeGraphEntityAttributes `json:"attributes"`
	}
	KnowledgeGraphEntityAttributes struct {
		Category    kne.Category                 `json:"category"`
		Kind        string                       `json:"kind"`
		Aliases     []KnowledgeGraphSubjectAlias `json:"aliases"`
		LatestState *KnowledgeGraphSubjectState  `json:"latestState,omitempty"`
		CreatedAt   time.Time                    `json:"createdAt"`
		UpdatedAt   time.Time                    `json:"updatedAt"`
	}

	KnowledgeGraphRelationship struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes KnowledgeGraphRelationshipAttributes `json:"attributes"`
	}
	KnowledgeGraphRelationshipAttributes struct {
		Predicate      string                       `json:"predicate"`
		SourceEntityId uuid.UUID                    `json:"sourceEntityId"`
		TargetEntityId uuid.UUID                    `json:"targetEntityId"`
		Aliases        []KnowledgeGraphSubjectAlias `json:"aliases"`
		LatestState    *KnowledgeGraphSubjectState  `json:"latestState,omitempty"`
		CreatedAt      time.Time                    `json:"createdAt"`
		UpdatedAt      time.Time                    `json:"updatedAt"`
	}

	KnowledgeGraphSubjectAlias struct {
		Id         uuid.UUID                            `json:"id"`
		Attributes KnowledgeGraphSubjectAliasAttributes `json:"attributes"`
	}
	KnowledgeGraphSubjectAliasAttributes struct {
		Kind        string              `json:"kind" enum:"entity,relationship"`
		ResourceRef ProviderResourceRef `json:"resourceRef"`
	}

	KnowledgeGraphEvidence struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes KnowledgeGraphEvidenceAttributes `json:"attributes"`
	}

	KnowledgeGraphEvidenceAttributes struct {
		Kind           string                     `json:"kind" enum:"observed,deleted"`
		EffectiveAt    time.Time                  `json:"effectiveAt"`
		CreatedAt      time.Time                  `json:"createdAt"`
		EventId        uuid.UUID                  `json:"eventId"`
		SubjectAliasId uuid.UUID                  `json:"subjectAliasId"`
		SubjectState   KnowledgeGraphSubjectState `json:"subjectState"`
	}

	KnowledgeGraphSubjectState struct {
		DisplayName string         `json:"displayName"`
		Description string         `json:"description"`
		Properties  map[string]any `json:"properties"`
	}

	KnowledgeGraphEntitiesPage struct {
		Entities []KnowledgeGraphEntitySummary `json:"entities"`
	}

	KnowledgeGraphEntitySummary struct {
		ID       uuid.UUID    `json:"id"`
		Category kne.Category `json:"category"`
		Kind     string       `json:"kind"`
	}

	KnowledgeGraphRelationshipsPage struct {
		Relationships []KnowledgeGraphRelationshipSummary `json:"relationships"`
	}

	KnowledgeGraphRelationshipSummary struct {
		ID        uuid.UUID     `json:"id"`
		SourceID  uuid.UUID     `json:"sourceId"`
		TargetID  uuid.UUID     `json:"targetId"`
		Predicate knr.Predicate `json:"predicate"`
	}
)

func (o operations) RegisterKnowledgeGraphEnums(api huma.API) {
	registerEnumAlias[kne.Category, knowledgeEntityCategorySchema](api)
	registerEnumAlias[knr.Predicate, knowledgeRelationshipPredicateSchema](api)
}

type knowledgeEntityCategorySchema kne.Category

func (knowledgeEntityCategorySchema) Schema(huma.Registry) *huma.Schema {
	return makeEnumStringSchema(kne.CategoryValues)
}

type knowledgeRelationshipPredicateSchema knr.Predicate

func (knowledgeRelationshipPredicateSchema) Schema(huma.Registry) *huma.Schema {
	return makeEnumStringSchema(knr.PredicateValues)
}

func KnowledgeGraphEntityFromEnt(e *ent.KnowledgeEntity) KnowledgeGraphEntity {
	attr := KnowledgeGraphEntityAttributes{
		Category:  e.Category,
		Kind:      e.Kind,
		Aliases:   make([]KnowledgeGraphSubjectAlias, len(e.Edges.Aliases)),
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
	if latestEv := e.LatestEvidence(); latestEv != nil {
		attr.LatestState = new(KnowledgeGraphSubjectStateFromEnt(latestEv.SubjectState))
	}

	for i, alias := range e.Edges.Aliases {
		attr.Aliases[i] = KnowledgeGraphSubjectAliasFromEnt(alias)
	}

	return KnowledgeGraphEntity{Id: e.ID, Attributes: attr}
}

func KnowledgeGraphRelationshipFromEnt(rel *ent.KnowledgeRelationship) KnowledgeGraphRelationship {
	attr := KnowledgeGraphRelationshipAttributes{
		Predicate:      rel.Predicate.String(),
		SourceEntityId: rel.SourceEntityID,
		TargetEntityId: rel.TargetEntityID,
		CreatedAt:      rel.CreatedAt,
		UpdatedAt:      rel.UpdatedAt,
		Aliases:        make([]KnowledgeGraphSubjectAlias, len(rel.Edges.Aliases)),
	}
	if latestEv := rel.LatestEvidence(); latestEv != nil {
		attr.LatestState = new(KnowledgeGraphSubjectStateFromEnt(latestEv.SubjectState))
	}
	for i, alias := range rel.Edges.Aliases {
		attr.Aliases[i] = KnowledgeGraphSubjectAliasFromEnt(alias)
	}
	return KnowledgeGraphRelationship{Id: rel.ID, Attributes: attr}
}

func KnowledgeGraphSubjectAliasFromEnt(alias *ent.KnowledgeSubjectAlias) KnowledgeGraphSubjectAlias {
	attrs := KnowledgeGraphSubjectAliasAttributes{
		Kind: alias.SubjectKind.String(),
		ResourceRef: ProviderResourceRefFromRez(rez.ProviderResourceRef{
			Provider:          alias.Provider,
			ProviderNamespace: alias.ProviderNamespace,
			ResourceRef:       alias.ProviderResourceRef,
		}),
	}
	return KnowledgeGraphSubjectAlias{Id: alias.ID, Attributes: attrs}
}

func KnowledgeGraphEvidenceFromEnt(ev *ent.KnowledgeEvidence) *KnowledgeGraphEvidence {
	attrs := KnowledgeGraphEvidenceAttributes{
		Kind:           ev.Kind.String(),
		EffectiveAt:    ev.EffectiveAt,
		CreatedAt:      ev.CreatedAt,
		EventId:        ev.EventID,
		SubjectAliasId: ev.SubjectAliasID,
		SubjectState:   KnowledgeGraphSubjectStateFromEnt(ev.SubjectState),
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

var knowledgeTags = []string{"Knowledge Graph"}

var ListKnowledgeGraphEntities = huma.Operation{
	OperationID: "list-knowledge-graph-entities",
	Method:      http.MethodGet,
	Path:        "/knowledge/entities",
	Summary:     "List Knowledge Graph Entities",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeGraphEntitiesRequest struct {
	PaginationRequest
	Search            string         `query:"search" required:"false" nullable:"false"`
	Category          []kne.Category `query:"category" required:"false"`
	Kind              []string       `query:"kind" required:"false"`
	Provider          string         `query:"provider" required:"false"`
	ProviderNamespace string         `query:"providerNamespace" required:"false"`
}
type ListKnowledgeGraphEntitiesResponse PaginatedResponse[KnowledgeGraphEntity]

var GetKnowledgeGraphEntity = huma.Operation{
	OperationID: "get-knowledge-graph-entity",
	Method:      http.MethodGet,
	Path:        "/knowledge/entities/{id}",
	Summary:     "Get Knowledge Graph Entity",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphEntityRequest IdRequest
type GetKnowledgeGraphEntityResponse ItemResponse[KnowledgeGraphEntity]

var ListKnowledgeGraphRelationships = huma.Operation{
	OperationID: "list-knowledge-graph-relationships",
	Method:      http.MethodGet,
	Path:        "/knowledge/relationships",
	Summary:     "List Knowledge Graph Relationships",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeGraphRelationshipsRequest struct {
	PaginationRequest
	Predicate      []string  `query:"predicate" required:"false"`
	EntityId       uuid.UUID `query:"entityId" required:"false"`
	SourceEntityId uuid.UUID `query:"sourceEntityId" required:"false"`
	TargetEntityId uuid.UUID `query:"targetEntityId" required:"false"`
}
type ListKnowledgeGraphRelationshipsResponse PaginatedResponse[KnowledgeGraphRelationship]

var GetKnowledgeGraphRelationship = huma.Operation{
	OperationID: "get-knowledge-graph-relationship",
	Method:      http.MethodGet,
	Path:        "/knowledge/relationships/{id}",
	Summary:     "Get Knowledge Graph Relationship",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphRelationshipRequest IdRequest
type GetKnowledgeGraphRelationshipResponse ItemResponse[KnowledgeGraphRelationship]

var ListKnowledgeGraphSubjectAliases = huma.Operation{
	OperationID: "list-knowledge-graph-aliases",
	Method:      http.MethodGet,
	Path:        "/knowledge/aliases",
	Summary:     "List Knowledge Graph Subject Aliases",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeGraphSubjectAliasesRequest struct {
	PaginationRequest
}
type ListKnowledgeGraphSubjectAliasesResponse PaginatedResponse[KnowledgeGraphSubjectAlias]

var GetKnowledgeGraphSubjectAlias = huma.Operation{
	OperationID: "get-knowledge-graph-alias",
	Method:      http.MethodGet,
	Path:        "/knowledge/aliases/{id}",
	Summary:     "Get Knowledge Graph Subject Alias",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphSubjectAliasRequest IdRequest
type GetKnowledgeGraphSubjectAliasResponse ItemResponse[KnowledgeGraphSubjectAlias]

var ListKnowledgeGraphEvidence = huma.Operation{
	OperationID: "list-knowledge-graph-evidence",
	Method:      http.MethodGet,
	Path:        "/knowledge/evidence",
	Summary:     "List Knowledge Graph Evidence",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeGraphEvidenceRequest struct {
	PaginationRequest
	SubjectAliasId uuid.UUID `query:"subjectAliasId" required:"false"`
	EntityId       uuid.UUID `query:"entityId" required:"false"`
	RelationshipId uuid.UUID `query:"relationshipId" required:"false"`
}
type ListKnowledgeGraphEvidenceResponse PaginatedResponse[KnowledgeGraphEvidence]

var GetKnowledgeGraphEvidence = huma.Operation{
	OperationID: "get-knowledge-graph-evidence",
	Method:      http.MethodGet,
	Path:        "/knowledge/evidence/{id}",
	Summary:     "Get Knowledge Graph Evidence",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeGraphEvidenceRequest IdRequest
type GetKnowledgeGraphEvidenceResponse ItemResponse[KnowledgeGraphEvidence]

var SelectKnowledgeGraphEntities = huma.Operation{
	OperationID: "select-knowledge-graph-entities",
	Method:      http.MethodGet,
	Path:        "/knowledge/graph/entities",
	Summary:     "Select Knowledge Graph Entities",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type SelectKnowledgeGraphEntitiesRequest struct {
}
type SelectKnowledgeGraphEntitiesResponse ItemResponse[KnowledgeGraphEntitiesPage]

var ExpandKnowledgeGraphRelationships = huma.Operation{
	OperationID: "expand-knowledge-graph-relationships",
	Method:      http.MethodGet,
	Path:        "/knowledge/graph/relationships",
	Summary:     "Expand Knowledge Graph Relationships",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type ExpandKnowledgeGraphRelationshipsRequest struct {
}
type ExpandKnowledgeGraphRelationshipsResponse ItemResponse[KnowledgeGraphRelationshipsPage]
