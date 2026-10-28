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
	kp "github.com/rezible/rezible/ent/knowledgeproposal"
	kpd "github.com/rezible/rezible/ent/knowledgeproposaldecision"
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

	ListKnowledgeProposals(context.Context, *ListKnowledgeProposalsRequest) (*ListKnowledgeProposalsResponse, error)
	GetKnowledgeProposal(context.Context, *GetKnowledgeProposalRequest) (*GetKnowledgeProposalResponse, error)
	ConfirmKnowledgeProposal(context.Context, *ConfirmKnowledgeProposalRequest) (*ConfirmKnowledgeProposalResponse, error)
	RejectKnowledgeProposal(context.Context, *RejectKnowledgeProposalRequest) (*RejectKnowledgeProposalResponse, error)
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

	huma.Register(api, ListKnowledgeProposals, o.ListKnowledgeProposals)
	huma.Register(api, GetKnowledgeProposal, o.GetKnowledgeProposal)
	huma.Register(api, ConfirmKnowledgeProposal, o.ConfirmKnowledgeProposal)
	huma.Register(api, RejectKnowledgeProposal, o.RejectKnowledgeProposal)
}

type (
	KnowledgeGraphEntity struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes KnowledgeGraphEntityAttributes `json:"attributes"`
	}
	KnowledgeGraphEntityAttributes struct {
		Category    kne.Category                `json:"category"`
		Kind        string                      `json:"kind"`
		DisplayName string                      `json:"displayName"`
		LatestState *KnowledgeGraphSubjectState `json:"latestState,omitempty"`
		// LinkingAttributes and the optional alias fields are filled by the entity list and get operations.
		LinkingAttributes []KnowledgeGraphLinkingAttribute `json:"linkingAttributes"`
		Aliases           []KnowledgeGraphEntityAlias      `json:"aliases"`
		CreatedAt         time.Time                        `json:"createdAt"`
		UpdatedAt         time.Time                        `json:"updatedAt"`
	}

	KnowledgeGraphLinkingAttribute struct {
		Attribute string `json:"attribute"`
		Value     string `json:"value"`
	}

	// KnowledgeGraphEntityAlias is one of an entity's aliases with its provider's name and when Rezible
	// last recorded evidence from it.
	KnowledgeGraphEntityAlias struct {
		KnowledgeGraphSubjectAlias
		IntegrationDisplayName string `json:"integrationDisplayName,omitempty"`
		Provider               string `json:"provider,omitempty"`
		// Name is the display name the alias's newest evidence states, or its resource ref when it states none.
		Name string `json:"name,omitempty"`
		// LastRecordedAt is when Rezible last stored evidence from the alias, not when the alias last reported
		// the name.
		LastRecordedAt *time.Time `json:"lastRecordedAt,omitempty"`
		// MergedByProposal is set when the alias joined the entity through a confirmed match.
		MergedByProposal *KnowledgeGraphAliasMerge `json:"mergedByProposal,omitempty"`
	}

	KnowledgeGraphAliasMerge struct {
		ProposalId  uuid.UUID                `json:"proposalId"`
		ConfirmedBy *KnowledgeProposalPerson `json:"confirmedBy,omitempty"`
		ConfirmedAt *time.Time               `json:"confirmedAt,omitempty"`
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
		Assertion      string                     `json:"assertion"`
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
		Entities           []KnowledgeGraphEntitySummary `json:"entities"`
		EntitySelectionRef string                        `json:"entitySelectionRef"`
		NextCursor         *string                       `json:"nextCursor,omitempty"`
	}

	KnowledgeGraphEntitySummary struct {
		ID          uuid.UUID    `json:"id"`
		Category    kne.Category `json:"category"`
		Kind        string       `json:"kind"`
		DisplayName string       `json:"displayName"`
	}

	KnowledgeGraphRelationshipsPage struct {
		Relationships []KnowledgeGraphRelationshipSummary `json:"relationships"`
		NextCursor    *string                             `json:"nextCursor,omitempty"`
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
		Category:          e.Category,
		Kind:              e.Kind,
		DisplayName:       e.State.DisplayName,
		LinkingAttributes: make([]KnowledgeGraphLinkingAttribute, len(e.Edges.LinkingAttributes)),
		Aliases:           make([]KnowledgeGraphEntityAlias, len(e.Edges.Aliases)),
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
	if latestEv := e.LatestEvidence(); latestEv != nil {
		attr.LatestState = new(KnowledgeGraphSubjectStateFromEnt(latestEv.SubjectState))
	}

	for i, alias := range e.Edges.Aliases {
		attr.Aliases[i] = KnowledgeGraphEntityAlias{KnowledgeGraphSubjectAlias: KnowledgeGraphSubjectAliasFromEnt(alias)}
	}
	for i, linkingAttribute := range e.Edges.LinkingAttributes {
		attr.LinkingAttributes[i] = KnowledgeGraphLinkingAttribute{
			Attribute: linkingAttribute.Attribute,
			Value:     linkingAttribute.Value,
		}
	}

	return KnowledgeGraphEntity{Id: e.ID, Attributes: attr}
}

// KnowledgeGraphEntityWithAliasesFromRez converts an entity with its loaded linking attributes and its aliases.
func KnowledgeGraphEntityWithAliasesFromRez(e *rez.KnowledgeEntityWithAliases) KnowledgeGraphEntity {
	entity := KnowledgeGraphEntityFromEnt(e.Entity)
	entity.Attributes.Aliases = make([]KnowledgeGraphEntityAlias, 0, len(e.Aliases))
	for _, alias := range e.Aliases {
		entity.Attributes.Aliases = append(entity.Attributes.Aliases, KnowledgeGraphEntityAliasFromRez(alias))
	}
	return entity
}

func KnowledgeGraphEntityAliasFromRez(alias rez.KnowledgeEntityAlias) KnowledgeGraphEntityAlias {
	converted := KnowledgeGraphEntityAlias{
		KnowledgeGraphSubjectAlias: KnowledgeGraphSubjectAliasFromEnt(alias.Alias),
		IntegrationDisplayName:     alias.IntegrationDisplayName,
		Provider:                   alias.Alias.Provider,
		Name:                       alias.Name,
	}
	if !alias.LastRecordedAt.IsZero() {
		converted.LastRecordedAt = new(alias.LastRecordedAt)
	}
	if proposal := alias.Alias.Edges.MergedByProposal; proposal != nil {
		merge := &KnowledgeGraphAliasMerge{ProposalId: proposal.ID}
		for _, decision := range proposal.Edges.Decisions {
			if decision.Status == kpd.StatusConfirmed {
				merge.ConfirmedBy = KnowledgeProposalPersonFromEnt(decision.Edges.User)
				merge.ConfirmedAt = new(decision.CreatedAt)
			}
		}
		converted.MergedByProposal = merge
	}
	return converted
}

func KnowledgeGraphEntitySummaryFromEnt(entity *ent.KnowledgeEntity) KnowledgeGraphEntitySummary {
	return KnowledgeGraphEntitySummary{
		ID:          entity.ID,
		Category:    entity.Category,
		Kind:        entity.Kind,
		DisplayName: entity.State.DisplayName,
	}
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

func KnowledgeGraphRelationshipSummaryFromEnt(relationship *ent.KnowledgeRelationship) KnowledgeGraphRelationshipSummary {
	return KnowledgeGraphRelationshipSummary{
		ID:        relationship.ID,
		SourceID:  relationship.SourceEntityID,
		TargetID:  relationship.TargetEntityID,
		Predicate: relationship.Predicate,
	}
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
		Assertion:      ev.Assertion,
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
	Category []kne.Category        `query:"category" required:"false"`
	Kind     []string              `query:"kind" required:"false"`
	Cursor   OptionalParam[string] `query:"cursor" required:"false"`
	Limit    OptionalParam[int]    `query:"limit" required:"false"`
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
	EntitySelectionRef string                `query:"entitySelectionRef" required:"true"`
	Predicate          []knr.Predicate       `query:"predicate" required:"false"`
	Cursor             OptionalParam[string] `query:"cursor" required:"false"`
	Limit              OptionalParam[int]    `query:"limit" required:"false"`
}

type ExpandKnowledgeGraphRelationshipsResponse ItemResponse[KnowledgeGraphRelationshipsPage]

type (
	KnowledgeProposal struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes KnowledgeProposalAttributes `json:"attributes"`
	}

	// KnowledgeProposalAttributes describe a match Rezible suggests: two values of one linking attribute that may
	// name one entity.
	KnowledgeProposalAttributes struct {
		Category  kne.Category `json:"category"`
		Kind      string       `json:"kind"`
		Attribute string       `json:"attribute"`
		// Values are the two values, in order, each with the entity holding it now.
		Values []KnowledgeProposalValue `json:"values"`
		Status string                   `json:"status" enum:"pending,confirmed,rejected"`
		// Origin is why Rezible proposed it: a similar name, or an alias naming an entity differently.
		Origin string `json:"origin" enum:"name_similarity,identity_conflict"`
		// Rationale is Rezible's suggestion, in its own words.
		Rationale string                      `json:"rationale"`
		Evidence  KnowledgeProposalEvidence   `json:"evidence"`
		Decisions []KnowledgeProposalDecision `json:"decisions"`
		CreatedAt time.Time                   `json:"createdAt"`
		// Implies, set by get-knowledge-proposal, lists the other proposals confirming this one would settle.
		Implies []KnowledgeProposalImplication `json:"implies,omitempty"`
	}

	KnowledgeProposalValue struct {
		Value string `json:"value"`
		// Entity is the entity holding the value, absent when none does.
		Entity *KnowledgeProposalValueEntity `json:"entity,omitempty"`
	}

	KnowledgeProposalValueEntity struct {
		Id          uuid.UUID                   `json:"id"`
		DisplayName string                      `json:"displayName"`
		Aliases     []KnowledgeGraphEntityAlias `json:"aliases"`
	}

	// KnowledgeProposalEvidence is the evidence that brought the value which caused the proposal.
	KnowledgeProposalEvidence struct {
		Id uuid.UUID `json:"id"`
		// SubjectEntityId is the entity the evidence is recorded on now.
		SubjectEntityId        *uuid.UUID `json:"subjectEntityId,omitempty"`
		IntegrationDisplayName string     `json:"integrationDisplayName"`
		Provider               string     `json:"provider"`
		ResourceRef            string     `json:"resourceRef"`
		Assertion              string     `json:"assertion"`
		EffectiveAt            time.Time  `json:"effectiveAt"`
		// DisplayName is the name the evidence states.
		DisplayName string `json:"displayName"`
	}

	KnowledgeProposalDecision struct {
		Status    string                  `json:"status" enum:"confirmed,rejected"`
		User      KnowledgeProposalPerson `json:"user"`
		CreatedAt time.Time               `json:"createdAt"`
		// ImpliedByProposalId is set when the decision followed from confirming that proposal, whose merge put
		// both values on one entity.
		ImpliedByProposalId *uuid.UUID `json:"impliedByProposalId,omitempty"`
	}

	KnowledgeProposalPerson struct {
		Id   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	// KnowledgeProposalImplication is a proposal that confirming another would confirm with it.
	KnowledgeProposalImplication struct {
		Id     uuid.UUID `json:"id"`
		Values []string  `json:"values"`
		Status string    `json:"status" enum:"pending,rejected"`
		// RejectedBy and RejectedAt are the latest rejection, when it is rejected.
		RejectedBy *KnowledgeProposalPerson `json:"rejectedBy,omitempty"`
		RejectedAt *time.Time               `json:"rejectedAt,omitempty"`
	}
)

func KnowledgeProposalPersonFromEnt(user *ent.User) *KnowledgeProposalPerson {
	if user == nil {
		return nil
	}
	return &KnowledgeProposalPerson{Id: user.ID, Name: user.Name}
}

// KnowledgeProposalFromRez converts a proposal with the entities holding its values.
func KnowledgeProposalFromRez(p *rez.KnowledgeProposalWithValues) KnowledgeProposal {
	proposal := p.Proposal
	attrs := KnowledgeProposalAttributes{
		Category:  kne.Category(proposal.Category),
		Kind:      proposal.Kind,
		Attribute: proposal.Attribute,
		Status:    proposal.Status.String(),
		Origin:    proposal.Origin.String(),
		Rationale: proposal.Rationale,
		Decisions: make([]KnowledgeProposalDecision, len(proposal.Edges.Decisions)),
		CreatedAt: proposal.CreatedAt,
	}
	attrs.Values = []KnowledgeProposalValue{
		KnowledgeProposalValueFromRez(p.ValueA),
		KnowledgeProposalValueFromRez(p.ValueB),
	}
	if evidence := proposal.Edges.Evidence; evidence != nil {
		attrs.Evidence = KnowledgeProposalEvidence{
			Id:          evidence.ID,
			Assertion:   evidence.Assertion,
			EffectiveAt: evidence.EffectiveAt,
			DisplayName: evidence.SubjectState.DisplayName,
		}
		if alias := evidence.Edges.SubjectAlias; alias != nil {
			attrs.Evidence.SubjectEntityId = alias.EntityID
			attrs.Evidence.IntegrationDisplayName = p.EvidenceIntegrationDisplayName
			attrs.Evidence.Provider = alias.Provider
			attrs.Evidence.ResourceRef = alias.ProviderResourceRef
		}
	}
	for i, decision := range proposal.Edges.Decisions {
		attrs.Decisions[i] = KnowledgeProposalDecision{
			Status:              decision.Status.String(),
			CreatedAt:           decision.CreatedAt,
			ImpliedByProposalId: decision.ImpliedByProposalID,
		}
		if person := KnowledgeProposalPersonFromEnt(decision.Edges.User); person != nil {
			attrs.Decisions[i].User = *person
		}
	}
	return KnowledgeProposal{Id: proposal.ID, Attributes: attrs}
}

// KnowledgeProposalValueFromRez converts a proposal's value with the entity holding it, if any.
func KnowledgeProposalValueFromRez(value rez.KnowledgeProposalValue) KnowledgeProposalValue {
	converted := KnowledgeProposalValue{Value: value.Value}
	if value.Holder == nil {
		return converted
	}
	converted.Entity = &KnowledgeProposalValueEntity{
		Id:          value.Holder.Entity.ID,
		DisplayName: value.Holder.Entity.State.DisplayName,
		Aliases:     make([]KnowledgeGraphEntityAlias, 0),
	}
	for _, alias := range value.Holder.Aliases {
		converted.Entity.Aliases = append(converted.Entity.Aliases, KnowledgeGraphEntityAliasFromRez(alias))
	}
	return converted
}

func KnowledgeProposalImplicationFromEnt(proposal *ent.KnowledgeProposal) KnowledgeProposalImplication {
	implication := KnowledgeProposalImplication{
		Id:     proposal.ID,
		Values: []string{proposal.ValueA, proposal.ValueB},
		Status: proposal.Status.String(),
	}
	if proposal.Status == kp.StatusRejected {
		for _, decision := range proposal.Edges.Decisions {
			if decision.Status == kpd.StatusRejected {
				implication.RejectedBy = KnowledgeProposalPersonFromEnt(decision.Edges.User)
				implication.RejectedAt = new(decision.CreatedAt)
			}
		}
	}
	return implication
}

var ListKnowledgeProposals = huma.Operation{
	OperationID: "list-knowledge-proposals",
	Method:      http.MethodGet,
	Path:        "/knowledge/proposals",
	Summary:     "List Knowledge Proposals",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type ListKnowledgeProposalsRequest struct {
	PaginationRequest
	Status []string `query:"status" required:"false" enum:"pending,confirmed,rejected"`
	// EntityId keeps proposals naming a value one of the entities holds now.
	EntityId []uuid.UUID `query:"entityId" required:"false"`
}
type ListKnowledgeProposalsResponse PaginatedResponse[KnowledgeProposal]

var GetKnowledgeProposal = huma.Operation{
	OperationID: "get-knowledge-proposal",
	Method:      http.MethodGet,
	Path:        "/knowledge/proposals/{id}",
	Summary:     "Get Knowledge Proposal",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(),
}

type GetKnowledgeProposalRequest IdRequest
type GetKnowledgeProposalResponse ItemResponse[KnowledgeProposal]

var ConfirmKnowledgeProposal = huma.Operation{
	OperationID: "confirm-knowledge-proposal",
	Method:      http.MethodPost,
	Path:        "/knowledge/proposals/{id}/confirm",
	Summary:     "Confirm Knowledge Proposal",
	Description: "Confirms that the proposal's two values name one entity, and merges their entities. This can't be undone.",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type ConfirmKnowledgeProposalRequest IdRequest
type ConfirmKnowledgeProposalResponse ItemResponse[KnowledgeProposal]

var RejectKnowledgeProposal = huma.Operation{
	OperationID: "reject-knowledge-proposal",
	Method:      http.MethodPost,
	Path:        "/knowledge/proposals/{id}/reject",
	Summary:     "Reject Knowledge Proposal",
	Tags:        knowledgeTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type RejectKnowledgeProposalRequest IdRequest
type RejectKnowledgeProposalResponse ItemResponse[KnowledgeProposal]
