package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	sa "github.com/rezible/rezible/ent/systemanalysis"
	sate "github.com/rezible/rezible/ent/systemanalysistopologyedge"
	satn "github.com/rezible/rezible/ent/systemanalysistopologynode"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type systemAnalysisHandler struct {
	db rez.Database
}

func newSystemAnalysisHandler(db rez.Database) *systemAnalysisHandler {
	return &systemAnalysisHandler{db: db}
}

func (s *systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	var resp oapi.GetSystemAnalysisResponse

	analysis, queryErr := s.db.Client(ctx).SystemAnalysis.Query().
		Where(sa.ID(request.Id)).
		WithTopologySnapshot(func(q *ent.SystemTopologySnapshotQuery) {
			q.WithEntities()
			q.WithRelationships()
		}).
		WithAnalysisNodes(func(q *ent.SystemAnalysisTopologyNodeQuery) {
			q.WithSnapshotEntity()
		}).
		WithAnalysisEdges(func(q *ent.SystemAnalysisTopologyEdgeQuery) {
			q.WithSnapshotRelationship()
		}).
		Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get system analysis", queryErr)
	}
	resp.Body.Data = oapi.SystemAnalysisFromEnt(analysis)

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisNodes(ctx context.Context, request *oapi.ListSystemAnalysisNodesRequest) (*oapi.ListSystemAnalysisNodesResponse, error) {
	var resp oapi.ListSystemAnalysisNodesResponse

	nodes, queryErr := s.db.Client(ctx).SystemAnalysisTopologyNode.Query().
		Where(satn.AnalysisID(request.Id)).
		WithSnapshotEntity().
		All(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to query system analysis nodes", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemAnalysisNode, len(nodes))
	for i, node := range nodes {
		resp.Body.Data[i] = oapi.SystemAnalysisNodeFromEnt(node)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) AddSystemAnalysisNode(ctx context.Context, request *oapi.AddSystemAnalysisNodeRequest) (*oapi.AddSystemAnalysisNodeResponse, error) {
	var resp oapi.AddSystemAnalysisNodeResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisNode(ctx context.Context, request *oapi.GetSystemAnalysisNodeRequest) (*oapi.GetSystemAnalysisNodeResponse, error) {
	var resp oapi.GetSystemAnalysisNodeResponse

	node, getErr := s.db.Client(ctx).SystemAnalysisTopologyNode.Query().
		Where(satn.ID(request.Id)).
		WithSnapshotEntity().
		Only(ctx)
	if getErr != nil {
		return nil, oapi.Error(ctx, "failed to get system analysis node", getErr)
	}
	resp.Body.Data = oapi.SystemAnalysisNodeFromEnt(node)

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisNode(ctx context.Context, request *oapi.UpdateSystemAnalysisNodeRequest) (*oapi.UpdateSystemAnalysisNodeResponse, error) {
	var resp oapi.UpdateSystemAnalysisNodeResponse

	attr := request.Body.Attributes
	update := s.db.Client(ctx).SystemAnalysisTopologyNode.UpdateOneID(request.Id).
		SetNillableDescription(attr.Description)
	if attr.Position != nil {
		update.SetPosX(attr.Position.X)
		update.SetPosY(attr.Position.Y)
	}

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to update system analysis node", updateErr)
	}
	updated.Edges.SnapshotEntity, _ = updated.QuerySnapshotEntity().Only(ctx)
	resp.Body.Data = oapi.SystemAnalysisNodeFromEnt(updated)

	return &resp, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisNode(ctx context.Context, request *oapi.DeleteSystemAnalysisNodeRequest) (*oapi.DeleteSystemAnalysisNodeResponse, error) {
	var resp oapi.DeleteSystemAnalysisNodeResponse

	if delErr := s.db.Client(ctx).SystemAnalysisTopologyNode.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, oapi.Error(ctx, "failed to delete system analysis node", delErr)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisEdges(ctx context.Context, request *oapi.ListSystemAnalysisEdgesRequest) (*oapi.ListSystemAnalysisEdgesResponse, error) {
	var resp oapi.ListSystemAnalysisEdgesResponse

	edges, queryErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.Query().
		Where(sate.AnalysisID(request.Id)).
		WithSnapshotRelationship().
		All(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to query system analysis edges", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemAnalysisEdge, len(edges))
	for i, edge := range edges {
		resp.Body.Data[i] = oapi.SystemAnalysisEdgeFromEnt(edge)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) AddSystemAnalysisEdge(ctx context.Context, request *oapi.AddSystemAnalysisEdgeRequest) (*oapi.AddSystemAnalysisEdgeResponse, error) {
	var resp oapi.AddSystemAnalysisEdgeResponse

	attr := request.Body.Attributes
	created, createErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.Create().
		SetAnalysisID(request.Id).
		SetSnapshotRelationshipID(attr.SnapshotRelationshipId).
		SetDescription(attr.Description).
		Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to add system analysis edge", createErr)
	}
	created.Edges.SnapshotRelationship, _ = s.db.Client(ctx).SystemTopologySnapshotRelationship.Get(ctx, attr.SnapshotRelationshipId)
	resp.Body.Data = oapi.SystemAnalysisEdgeFromEnt(created)

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisEdge(ctx context.Context, request *oapi.GetSystemAnalysisEdgeRequest) (*oapi.GetSystemAnalysisEdgeResponse, error) {
	var resp oapi.GetSystemAnalysisEdgeResponse

	edge, getErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.Query().
		Where(sate.ID(request.Id)).
		WithSnapshotRelationship().
		Only(ctx)
	if getErr != nil {
		return nil, oapi.Error(ctx, "failed to get system analysis edge", getErr)
	}
	resp.Body.Data = oapi.SystemAnalysisEdgeFromEnt(edge)

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisEdge(ctx context.Context, request *oapi.UpdateSystemAnalysisEdgeRequest) (*oapi.UpdateSystemAnalysisEdgeResponse, error) {
	var resp oapi.UpdateSystemAnalysisEdgeResponse

	updated, updateErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.UpdateOneID(request.Id).
		SetNillableDescription(request.Body.Attributes.Description).
		Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to update system analysis edge", updateErr)
	}
	updated.Edges.SnapshotRelationship, _ = updated.QuerySnapshotRelationship().Only(ctx)
	resp.Body.Data = oapi.SystemAnalysisEdgeFromEnt(updated)

	return &resp, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisEdge(ctx context.Context, request *oapi.DeleteSystemAnalysisEdgeRequest) (*oapi.DeleteSystemAnalysisEdgeResponse, error) {
	var resp oapi.DeleteSystemAnalysisEdgeResponse

	if delErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, oapi.Error(ctx, "failed to delete system analysis edge", delErr)
	}

	return &resp, nil
}
