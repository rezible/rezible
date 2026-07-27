package apiv1

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	kr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/predicate"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type knowledgeGraphHandler struct {
	knowledge rez.KnowledgeGraphService
}

func newKnowledgeGraphHandler(knowledge rez.KnowledgeGraphService) *knowledgeGraphHandler {
	return &knowledgeGraphHandler{knowledge: knowledge}
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphEntities(ctx context.Context, request *oapi.ListKnowledgeGraphEntitiesRequest) (*oapi.ListKnowledgeGraphEntitiesResponse, error) {
	var response oapi.ListKnowledgeGraphEntitiesResponse
	var preds []predicate.KnowledgeEntity
	if len(request.Kind) > 0 {
		preds = append(preds, kne.KindIn(request.Kind...))
	}
	if request.Provider != "" {
		preds = append(preds, kne.HasAliasesWith(ksa.Provider(request.Provider)))
	}
	if request.ProviderSource != "" {
		preds = append(preds, kne.HasAliasesWith(ksa.HasEvidenceWith(knev.HasEventWith(ne.ProviderSource(request.ProviderSource)))))
	}
	if request.SubjectKind != "" {
		preds = append(preds, kne.HasAliasesWith(ksa.HasEvidenceWith(knev.HasEventWith(ne.SubjectKind(request.SubjectKind)))))
	}
	params := rez.ListKnowledgeGraphEntitiesParams{
		ListParams: request.ListParams(),
		Predicates: preds,
	}
	result, queryErr := h.knowledge.ListEntities(ctx, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list knowledge graph entities", queryErr)
	}
	response.Body.Data = make([]oapi.KnowledgeGraphEntity, len(result.Data))
	for i, entity := range result.Data {
		response.Body.Data[i] = oapi.KnowledgeGraphEntityFromEnt(entity)
	}
	response.Body.Pagination = oapi.ResponsePagination{Total: result.Count}
	return &response, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphView(ctx context.Context, request *oapi.GetKnowledgeGraphViewRequest) (*oapi.GetKnowledgeGraphViewResponse, error) {
	params := rez.GetKnowledgeGraphViewParams{
		EntityID:          request.EntityId,
		Depth:             request.Depth,
		RelationshipKinds: request.RelationshipKind,
	}
	view, viewErr := h.knowledge.GetView(ctx, params)
	if viewErr != nil {
		return nil, oapi.Error(ctx, "failed to get knowledge graph view", viewErr)
	}
	var response oapi.GetKnowledgeGraphViewResponse
	response.Body.Data = oapi.KnowledgeGraphViewFromRez(view)
	return &response, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphEntity(ctx context.Context, request *oapi.GetKnowledgeGraphEntityRequest) (*oapi.GetKnowledgeGraphEntityResponse, error) {
	var response oapi.GetKnowledgeGraphEntityResponse
	entity, queryErr := h.knowledge.GetEntity(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get knowledge graph entity", queryErr)
	}
	response.Body.Data = oapi.KnowledgeGraphEntityFromEnt(entity)
	return &response, nil
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphRelationships(ctx context.Context, request *oapi.ListKnowledgeGraphRelationshipsRequest) (*oapi.ListKnowledgeGraphRelationshipsResponse, error) {
	var response oapi.ListKnowledgeGraphRelationshipsResponse
	var preds []predicate.KnowledgeRelationship
	if len(request.Kind) > 0 {
		preds = append(preds, kr.KindIn(request.Kind...))
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
	params := rez.ListKnowledgeGraphRelationshipsParams{
		ListParams: request.ListParams(),
		Predicates: preds,
	}
	result, queryErr := h.knowledge.ListRelationships(ctx, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list knowledge graph relationships", queryErr)
	}
	response.Body.Data = make([]oapi.KnowledgeGraphRelationship, len(result.Data))
	for i, rel := range result.Data {
		response.Body.Data[i] = oapi.KnowledgeGraphRelationshipFromEnt(rel)
	}
	response.Body.Pagination = oapi.ResponsePagination{Total: result.Count}
	return &response, nil
}
