package apiv1

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	kr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type knowledgeGraphHandler struct {
	knowledge rez.KnowledgeGraphQueryService
}

func newKnowledgeGraphHandler(knowledge rez.KnowledgeGraphQueryService) *knowledgeGraphHandler {
	return &knowledgeGraphHandler{knowledge: knowledge}
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphEntities(ctx context.Context, request *oapi.ListKnowledgeGraphEntitiesRequest) (*oapi.ListKnowledgeGraphEntitiesResponse, error) {
	var response oapi.ListKnowledgeGraphEntitiesResponse
	var preds []predicate.KnowledgeEntity
	if len(request.Category) > 0 {
		categories := make([]kne.Category, len(request.Category))
		for i, value := range request.Category {
			categories[i] = kne.Category(value)
			if categoryErr := kne.CategoryValidator(categories[i]); categoryErr != nil {
				return nil, oapi.Error(ctx, "invalid knowledge graph entity category", categoryErr)
			}
		}
		preds = append(preds, kne.CategoryIn(categories...))
	}
	if len(request.Kind) > 0 {
		preds = append(preds, kne.KindIn(request.Kind...))
	}
	if request.Provider != "" {
		preds = append(preds, kne.HasAliasesWith(ksa.Provider(request.Provider)))
	}
	if request.ProviderNamespace != "" {
		preds = append(preds, kne.HasAliasesWith(ksa.ProviderNamespace(request.ProviderNamespace)))
	}
	listParams := request.ListParams()
	listParams.Search = request.Search
	params := rez.ListKnowledgeEntitiesParams{
		ListParams: listParams,
		Predicates: preds,
	}
	result, listErr := h.knowledge.ListEntities(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list knowledge graph entities", listErr)
	}
	response.Body = oapi.ConvertPaginatedResultBody(result, oapi.KnowledgeGraphEntityFromEnt)
	return &response, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphEntity(ctx context.Context, request *oapi.GetKnowledgeGraphEntityRequest) (*oapi.GetKnowledgeGraphEntityResponse, error) {
	var response oapi.GetKnowledgeGraphEntityResponse
	entity, entityErr := h.knowledge.GetEntity(ctx, request.Id)
	if entityErr != nil {
		return nil, oapi.Error(ctx, "failed to get knowledge graph entity", entityErr)
	}
	response.Body.Data = oapi.KnowledgeGraphEntityFromEnt(entity)
	return &response, nil
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphRelationships(ctx context.Context, request *oapi.ListKnowledgeGraphRelationshipsRequest) (*oapi.ListKnowledgeGraphRelationshipsResponse, error) {
	var response oapi.ListKnowledgeGraphRelationshipsResponse
	var preds []predicate.KnowledgeRelationship
	if len(request.Predicate) > 0 {
		predicates := make([]kr.Predicate, len(request.Predicate))
		for i, value := range request.Predicate {
			predicates[i] = kr.Predicate(value)
			if predicateErr := kr.PredicateValidator(predicates[i]); predicateErr != nil {
				return nil, oapi.Error(ctx, "invalid knowledge graph relationship predicate", predicateErr)
			}
		}
		preds = append(preds, kr.PredicateIn(predicates...))
	}
	if request.EntityId != uuid.Nil {
		preds = append(preds, kr.Or(kr.SourceEntityID(request.EntityId), kr.TargetEntityID(request.EntityId)))
	}
	if request.SourceEntityId != uuid.Nil {
		preds = append(preds, kr.SourceEntityID(request.SourceEntityId))
	}
	if request.TargetEntityId != uuid.Nil {
		preds = append(preds, kr.TargetEntityID(request.TargetEntityId))
	}
	params := rez.ListKnowledgeRelationshipsParams{
		ListParams: request.ListParams(),
		Predicates: preds,
	}
	result, queryErr := h.knowledge.ListRelationships(ctx, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list knowledge graph relationships", queryErr)
	}
	response.Body = oapi.ConvertPaginatedResultBody(result, oapi.KnowledgeGraphRelationshipFromEnt)
	return &response, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphRelationship(ctx context.Context, request *oapi.GetKnowledgeGraphRelationshipRequest) (*oapi.GetKnowledgeGraphRelationshipResponse, error) {
	var response oapi.GetKnowledgeGraphRelationshipResponse
	rel, relErr := h.knowledge.GetRelationship(ctx, request.Id)
	if relErr != nil {
		return nil, oapi.Error(ctx, "failed to get knowledge graph relationship", relErr)
	}
	response.Body.Data = oapi.KnowledgeGraphRelationshipFromEnt(rel)
	return &response, nil
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphSubjectAliases(ctx context.Context, request *oapi.ListKnowledgeGraphSubjectAliasesRequest) (*oapi.ListKnowledgeGraphSubjectAliasesResponse, error) {
	var response oapi.ListKnowledgeGraphSubjectAliasesResponse
	var preds []predicate.KnowledgeSubjectAlias

	params := rez.ListKnowledgeSubjectAliasesParams{
		ListParams: request.ListParams(),
		Predicates: preds,
	}
	result, queryErr := h.knowledge.ListSubjectAliases(ctx, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list knowledge graph relationships", queryErr)
	}
	response.Body = oapi.ConvertPaginatedResultBody(result, oapi.KnowledgeGraphSubjectAliasFromEnt)
	return &response, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphSubjectAliases(ctx context.Context, request *oapi.GetKnowledgeGraphSubjectAliasRequest) (*oapi.GetKnowledgeGraphSubjectAliasResponse, error) {
	var response oapi.GetKnowledgeGraphSubjectAliasResponse
	alias, aliasErr := h.knowledge.GetSubjectAlias(ctx, request.Id)
	if aliasErr != nil {
		return nil, oapi.Error(ctx, "failed to get subject alias", aliasErr)
	}
	response.Body.Data = oapi.KnowledgeGraphSubjectAliasFromEnt(alias)
	return &response, nil
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphEvidence(ctx context.Context, request *oapi.ListKnowledgeGraphEvidenceRequest) (*oapi.ListKnowledgeGraphEvidenceResponse, error) {
	var predicates []predicate.KnowledgeEvidence
	if request.SubjectAliasId != uuid.Nil {
		predicates = append(predicates, ke.SubjectAliasID(request.SubjectAliasId))
	}
	if request.EntityId != uuid.Nil {
		predicates = append(predicates, ke.HasSubjectAliasWith(ksa.EntityID(request.EntityId)))
	}
	if request.RelationshipId != uuid.Nil {
		predicates = append(predicates, ke.HasSubjectAliasWith(ksa.RelationshipID(request.RelationshipId)))
	}
	result, resultErr := h.knowledge.ListEvidence(ctx, rez.ListKnowledgeEvidenceParams{
		ListParams: request.ListParams(),
		Predicates: predicates,
	})
	if resultErr != nil {
		return nil, oapi.Error(ctx, "failed to list knowledge graph evidence", resultErr)
	}
	data := make([]oapi.KnowledgeGraphEvidence, len(result.Data))
	for i, evidence := range result.Data {
		data[i] = *oapi.KnowledgeGraphEvidenceFromEnt(evidence)
	}
	return &oapi.ListKnowledgeGraphEvidenceResponse{Body: oapi.PaginatedResponseBody[oapi.KnowledgeGraphEvidence]{
		Data:       data,
		Pagination: oapi.ConvertResultPagination(result),
	}}, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphEvidence(ctx context.Context, request *oapi.GetKnowledgeGraphEvidenceRequest) (*oapi.GetKnowledgeGraphEvidenceResponse, error) {
	evidence, evidenceErr := h.knowledge.GetEvidence(ctx, request.Id)
	if evidenceErr != nil {
		return nil, oapi.Error(ctx, "failed to get knowledge graph evidence", evidenceErr)
	}
	var response oapi.GetKnowledgeGraphEvidenceResponse
	response.Body.Data = *oapi.KnowledgeGraphEvidenceFromEnt(evidence)
	return &response, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphView(ctx context.Context, request *oapi.GetKnowledgeGraphViewRequest) (*oapi.GetKnowledgeGraphViewResponse, error) {
	var response oapi.GetKnowledgeGraphViewResponse
	response.Body.Data = oapi.KnowledgeGraphView{
		RootId:        uuid.UUID{},
		Entities:      make([]oapi.KnowledgeGraphEntity, 0),
		Relationships: make([]oapi.KnowledgeGraphRelationship, 0),
		Truncated:     false,
	}
	return &response, nil
}
