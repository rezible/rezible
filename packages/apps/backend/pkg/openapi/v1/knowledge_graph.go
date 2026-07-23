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

type KnowledgeGraphHandler interface {
	ListKnowledgeGraphEntities(context.Context, *ListKnowledgeGraphEntitiesRequest) (*ListKnowledgeGraphEntitiesResponse, error)
	GetKnowledgeGraphEntity(context.Context, *GetKnowledgeGraphEntityRequest) (*GetKnowledgeGraphEntityResponse, error)
	GetKnowledgeGraphView(context.Context, *GetKnowledgeGraphViewRequest) (*GetKnowledgeGraphViewResponse, error)

	ListKnowledgeGraphRelationships(context.Context, *ListKnowledgeGraphRelationshipsRequest) (*ListKnowledgeGraphRelationshipsResponse, error)
}

func (o operations) RegisterKnowledgeGraph(api huma.API) {
	huma.Register(api, ListKnowledgeGraphEntities, o.ListKnowledgeGraphEntities)
	huma.Register(api, GetKnowledgeGraphEntity, o.GetKnowledgeGraphEntity)
	huma.Register(api, GetKnowledgeGraphView, o.GetKnowledgeGraphView)

	huma.Register(api, ListKnowledgeGraphRelationships, o.ListKnowledgeGraphRelationships)
}

type (
	KnowledgeGraphView struct {
		Entities      []KnowledgeGraphEntity       `json:"entities"`
		Relationships []KnowledgeGraphRelationship `json:"relationships"`
		Evidence      []KnowledgeGraphEvidence     `json:"evidence"`
		Truncated     bool                         `json:"truncated"`
		Warnings      []string                     `json:"warnings"`
	}

	KnowledgeGraphEvidence struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes KnowledgeGraphEvidenceAttributes `json:"attributes"`
	}
	KnowledgeGraphEvidenceAttributes struct {
		EventId        uuid.UUID      `json:"eventId"`
		Provider       string         `json:"provider"`
		ProviderSource string         `json:"providerSource"`
		Assertion      string         `json:"assertion"`
		EvidenceKind   string         `json:"evidenceKind"`
		EffectiveAt    time.Time      `json:"effectiveAt"`
		Properties     map[string]any `json:"properties"`
		EntityId       *uuid.UUID     `json:"entityId,omitempty"`
		RelationshipId *uuid.UUID     `json:"relationshipId,omitempty"`
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
		Description        string `json:"description"`
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
		Kind:               alias.SubjectKind.String(),
		Description:        alias.Description,
		Provider:           alias.Provider,
		ProviderSubjectRef: alias.ProviderSubjectRef,
	}
	return KnowledgeGraphSubjectAlias{Id: alias.ID, Attributes: attrs}
}

func KnowledgeGraphRelationshipFromEnt(rel *ent.KnowledgeRelationship) KnowledgeGraphRelationship {
	attr := KnowledgeGraphRelationshipAttributes{
		Kind:        rel.Kind,
		DisplayName: rel.Kind,
		Description: rel.Description,
		Properties:  rel.Properties,
		CreatedAt:   rel.CreatedAt,
		UpdatedAt:   rel.UpdatedAt,
		Source:      Expandable[KnowledgeGraphEntityAttributes]{Id: rel.SourceEntityID},
		Target:      Expandable[KnowledgeGraphEntityAttributes]{Id: rel.TargetEntityID},
	}
	for i, alias := range rel.Edges.Aliases {
		if i == 0 || alias.FirstObservedAt.Before(attr.FirstSeenAt) {
			attr.FirstSeenAt = alias.FirstObservedAt
		}
		if alias.LastObservedAt.After(attr.LastSeenAt) {
			attr.LastSeenAt = alias.LastObservedAt
		}
	}
	if attr.FirstSeenAt.IsZero() {
		attr.FirstSeenAt = rel.CreatedAt
		attr.LastSeenAt = rel.UpdatedAt
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

func KnowledgeGraphViewFromRez(view *rez.KnowledgeGraphView) KnowledgeGraphView {
	result := KnowledgeGraphView{Truncated: view.Truncated, Warnings: view.Warnings}
	result.Entities = make([]KnowledgeGraphEntity, len(view.Entities))
	for i, entity := range view.Entities {
		result.Entities[i] = KnowledgeGraphEntityFromEnt(entity)
	}
	result.Relationships = make([]KnowledgeGraphRelationship, len(view.Relationships))
	for i, relationship := range view.Relationships {
		result.Relationships[i] = KnowledgeGraphRelationshipFromEnt(relationship)
	}
	result.Evidence = make([]KnowledgeGraphEvidence, len(view.Evidence))
	for i, evidence := range view.Evidence {
		result.Evidence[i] = KnowledgeGraphEvidenceFromEnt(evidence)
	}
	return result
}

func KnowledgeGraphEvidenceFromEnt(evidence *ent.KnowledgeEvidence) KnowledgeGraphEvidence {
	attributes := KnowledgeGraphEvidenceAttributes{
		EventId: evidence.EventID, Assertion: evidence.Assertion,
		EvidenceKind: evidence.EvidenceKind.String(), EffectiveAt: evidence.EffectiveAt,
		Properties: evidence.Properties,
	}
	if event, err := evidence.Edges.EventOrErr(); err == nil {
		attributes.Provider = event.Provider
		attributes.ProviderSource = event.ProviderSource
	}
	if alias, err := evidence.Edges.AliasOrErr(); err == nil {
		if alias.EntityID != uuid.Nil {
			id := alias.EntityID
			attributes.EntityId = &id
		}
		if alias.RelationshipID != uuid.Nil {
			id := alias.RelationshipID
			attributes.RelationshipId = &id
		}
	}
	return KnowledgeGraphEvidence{Id: evidence.ID, Attributes: attributes}
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
	Path:        "/knowledge_graph/entities/{entityId}/view",
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
