package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type systemTopologyHandler struct {
	topology rez.SystemTopologyService
}

func newSystemTopologyHandler(topology rez.SystemTopologyService) *systemTopologyHandler {
	return &systemTopologyHandler{topology: topology}
}

func (h *systemTopologyHandler) ListSystemTopologyEntities(ctx context.Context, request *oapi.ListSystemTopologyEntitiesRequest) (*oapi.ListSystemTopologyEntitiesResponse, error) {
	var resp oapi.ListSystemTopologyEntitiesResponse
	result, queryErr := h.topology.ListEntities(ctx, rez.ListSystemTopologyEntitiesParams{
		ListParams: request.ListParams(),
		Kinds:      request.Kind,
	})
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list topology entities", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemTopologyEntity, len(result.Data))
	for i, entity := range result.Data {
		resp.Body.Data[i] = oapi.SystemTopologyEntityFromEnt(entity)
	}
	resp.Body.Pagination = oapi.ResponsePagination{Total: result.Count}
	return &resp, nil
}

func (h *systemTopologyHandler) GetSystemTopologyEntity(ctx context.Context, request *oapi.GetSystemTopologyEntityRequest) (*oapi.GetSystemTopologyEntityResponse, error) {
	var resp oapi.GetSystemTopologyEntityResponse
	entity, queryErr := h.topology.GetEntity(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get topology entity", queryErr)
	}
	resp.Body.Data = oapi.SystemTopologyEntityFromEnt(entity)
	return &resp, nil
}

func (h *systemTopologyHandler) GetSystemTopologyEntityNeighborhood(ctx context.Context, request *oapi.GetSystemTopologyEntityNeighborhoodRequest) (*oapi.GetSystemTopologyEntityNeighborhoodResponse, error) {
	var resp oapi.GetSystemTopologyEntityNeighborhoodResponse
	graph, queryErr := h.topology.GetNeighborhood(ctx, request.Id, rez.SystemTopologyNeighborhoodParams{
		Depth:             request.Depth,
		RelationshipKinds: request.RelationshipKind,
	})
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get topology neighborhood", queryErr)
	}
	resp.Body.Data.Entities = make([]oapi.SystemTopologyEntity, len(graph.Entities))
	for i, entity := range graph.Entities {
		resp.Body.Data.Entities[i] = oapi.SystemTopologyEntityFromEnt(entity)
	}
	resp.Body.Data.Relationships = make([]oapi.SystemTopologyRelationship, len(graph.Relationships))
	for i, rel := range graph.Relationships {
		resp.Body.Data.Relationships[i] = oapi.SystemTopologyRelationshipFromEnt(rel)
	}
	return &resp, nil
}

func (h *systemTopologyHandler) ListSystemTopologyRelationships(ctx context.Context, request *oapi.ListSystemTopologyRelationshipsRequest) (*oapi.ListSystemTopologyRelationshipsResponse, error) {
	var resp oapi.ListSystemTopologyRelationshipsResponse
	result, queryErr := h.topology.ListRelationships(ctx, rez.ListSystemTopologyRelationshipsParams{
		ListParams: request.ListParams(),
		Kinds:      request.Kind,
	})
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list topology relationships", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemTopologyRelationship, len(result.Data))
	for i, rel := range result.Data {
		resp.Body.Data[i] = oapi.SystemTopologyRelationshipFromEnt(rel)
	}
	resp.Body.Pagination = oapi.ResponsePagination{Total: result.Count}
	return &resp, nil
}

func (h *systemTopologyHandler) CreateSystemTopologySnapshot(ctx context.Context, request *oapi.CreateSystemTopologySnapshotRequest) (*oapi.CreateSystemTopologySnapshotResponse, error) {
	var resp oapi.CreateSystemTopologySnapshotResponse
	snapshot, createErr := h.topology.CreateSnapshot(ctx, oapi.CreateSystemTopologySnapshotParamsFromAttributes(request.Body.Attributes))
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to create topology snapshot", createErr)
	}
	resp.Body.Data = oapi.SystemTopologySnapshotFromEnt(snapshot)
	return &resp, nil
}

func (h *systemTopologyHandler) GetSystemTopologySnapshot(ctx context.Context, request *oapi.GetSystemTopologySnapshotRequest) (*oapi.GetSystemTopologySnapshotResponse, error) {
	var resp oapi.GetSystemTopologySnapshotResponse
	snapshot, queryErr := h.topology.GetSnapshot(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get topology snapshot", queryErr)
	}
	resp.Body.Data = oapi.SystemTopologySnapshotFromEnt(snapshot)
	return &resp, nil
}
