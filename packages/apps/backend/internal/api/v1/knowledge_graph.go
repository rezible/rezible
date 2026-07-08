package apiv1

import (
	"context"
	"time"

	rez "github.com/rezible/rezible"
	ke "github.com/rezible/rezible/ent/knowledgeentity"
	kr "github.com/rezible/rezible/ent/knowledgerelationship"
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
	var resp oapi.ListKnowledgeGraphEntitiesResponse
	var preds []predicate.KnowledgeEntity
	if len(request.Kind) > 0 {
		preds = append(preds, ke.KindIn(request.Kind...))
	}
	result, queryErr := h.knowledge.ListEntities(ctx, rez.ListKnowledgeGraphEntitiesParams{
		ListParams: request.ListParams(),
		Predicates: preds,
	})
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list topology entities", queryErr)
	}
	resp.Body.Data = make([]oapi.KnowledgeGraphEntity, len(result.Data))
	for i, entity := range result.Data {
		resp.Body.Data[i] = oapi.KnowledgeGraphEntityFromEnt(entity)
	}
	resp.Body.Pagination = oapi.ResponsePagination{Total: result.Count}
	return &resp, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphEntity(ctx context.Context, request *oapi.GetKnowledgeGraphEntityRequest) (*oapi.GetKnowledgeGraphEntityResponse, error) {
	var resp oapi.GetKnowledgeGraphEntityResponse
	entity, queryErr := h.knowledge.GetEntity(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get topology entity", queryErr)
	}
	resp.Body.Data = oapi.KnowledgeGraphEntityFromEnt(entity)
	return &resp, nil
}

func (h *knowledgeGraphHandler) ListKnowledgeGraphRelationships(ctx context.Context, request *oapi.ListKnowledgeGraphRelationshipsRequest) (*oapi.ListKnowledgeGraphRelationshipsResponse, error) {
	var resp oapi.ListKnowledgeGraphRelationshipsResponse
	var preds []predicate.KnowledgeRelationship
	if len(request.Kind) > 0 {
		preds = append(preds, kr.KindIn(request.Kind...))
	}
	result, queryErr := h.knowledge.ListRelationships(ctx, rez.ListKnowledgeGraphRelationshipsParams{
		ListParams: request.ListParams(),
		Predicates: preds,
	})
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list topology relationships", queryErr)
	}
	resp.Body.Data = make([]oapi.KnowledgeGraphRelationship, len(result.Data))
	for i, rel := range result.Data {
		resp.Body.Data[i] = oapi.KnowledgeGraphRelationshipFromEnt(rel)
	}
	resp.Body.Pagination = oapi.ResponsePagination{Total: result.Count}
	return &resp, nil
}

func (h *knowledgeGraphHandler) CreateKnowledgeGraphSnapshot(ctx context.Context, req *oapi.CreateKnowledgeGraphSnapshotRequest) (*oapi.CreateKnowledgeGraphSnapshotResponse, error) {
	var resp oapi.CreateKnowledgeGraphSnapshotResponse
	attrs := req.Body.Attributes
	// TODO
	params := rez.CreateKnowledgeGraphSnapshotParams{
		Name:              attrs.Name,
		AsOf:              time.Time{},
		Scope:             "",
		ScopeProperties:   nil,
		EntityIDs:         nil,
		RootEntityIDs:     nil,
		Depth:             0,
		EntityKinds:       nil,
		RelationshipKinds: nil,
	}
	snapshot, createErr := h.knowledge.CreateSnapshot(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to create topology snapshot", createErr)
	}
	resp.Body.Data = oapi.KnowledgeGraphSnapshotFromEnt(snapshot)
	return &resp, nil
}

func (h *knowledgeGraphHandler) GetKnowledgeGraphSnapshot(ctx context.Context, request *oapi.GetKnowledgeGraphSnapshotRequest) (*oapi.GetKnowledgeGraphSnapshotResponse, error) {
	var resp oapi.GetKnowledgeGraphSnapshotResponse
	snapshot, queryErr := h.knowledge.GetSnapshot(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get topology snapshot", queryErr)
	}
	resp.Body.Data = oapi.KnowledgeGraphSnapshotFromEnt(snapshot)
	return &resp, nil
}
