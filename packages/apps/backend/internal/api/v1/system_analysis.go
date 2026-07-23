package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	isc "github.com/rezible/rezible/ent/incidenttimelineeventsystemcontext"
	sa "github.com/rezible/rezible/ent/systemanalysis"
	sate "github.com/rezible/rezible/ent/systemanalysistopologyedge"
	satn "github.com/rezible/rezible/ent/systemanalysistopologynode"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type systemAnalysisHandler struct {
	db        rez.Database
	knowledge rez.KnowledgeGraphService
}

func newSystemAnalysisHandler(db rez.Database, knowledge rez.KnowledgeGraphService) *systemAnalysisHandler {
	return &systemAnalysisHandler{db: db, knowledge: knowledge}
}

func (s *systemAnalysisHandler) nodeFromEnt(ctx context.Context, node *ent.SystemAnalysisTopologyNode) (oapi.SystemAnalysisNode, error) {
	entity, entityErr := s.knowledge.GetEntityAt(ctx, node.KnowledgeEntityID, node.ReferencedAt)
	if entityErr != nil {
		return oapi.SystemAnalysisNode{}, entityErr
	}
	return oapi.SystemAnalysisNode{
		Id: node.ID,
		Attributes: oapi.SystemAnalysisNodeAttributes{
			KnowledgeEntity: oapi.KnowledgeGraphEntityFromEnt(entity),
			ReferencedAt:    node.ReferencedAt,
			Position:        oapi.SystemAnalysisDiagramPosition{X: node.PosX, Y: node.PosY},
			Description:     node.Description,
		},
	}, nil
}

func (s *systemAnalysisHandler) edgeFromEnt(ctx context.Context, edge *ent.SystemAnalysisTopologyEdge) (oapi.SystemAnalysisEdge, error) {
	relationship, relationshipErr := s.knowledge.GetRelationshipAt(ctx, edge.KnowledgeRelationshipID, edge.ReferencedAt)
	if relationshipErr != nil {
		return oapi.SystemAnalysisEdge{}, relationshipErr
	}
	return oapi.SystemAnalysisEdge{
		Id: edge.ID,
		Attributes: oapi.SystemAnalysisEdgeAttributes{
			KnowledgeRelationship: oapi.KnowledgeGraphRelationshipFromEnt(relationship),
			ReferencedAt:          edge.ReferencedAt,
			Description:           edge.Description,
		},
	}, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	analysis, queryErr := s.db.Client(ctx).SystemAnalysis.Query().
		Where(sa.ID(request.Id)).
		WithAnalysisNodes().
		WithAnalysisEdges().
		Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "get system analysis", queryErr)
	}

	result := oapi.SystemAnalysis{Id: analysis.ID}
	result.Attributes.Nodes = make([]oapi.SystemAnalysisNode, len(analysis.Edges.AnalysisNodes))
	for i, node := range analysis.Edges.AnalysisNodes {
		converted, convertErr := s.nodeFromEnt(ctx, node)
		if convertErr != nil {
			return nil, oapi.Error(ctx, "get system analysis node", convertErr)
		}
		result.Attributes.Nodes[i] = converted
	}
	result.Attributes.Edges = make([]oapi.SystemAnalysisEdge, len(analysis.Edges.AnalysisEdges))
	for i, edge := range analysis.Edges.AnalysisEdges {
		converted, convertErr := s.edgeFromEnt(ctx, edge)
		if convertErr != nil {
			return nil, oapi.Error(ctx, "get system analysis edge", convertErr)
		}
		result.Attributes.Edges[i] = converted
	}
	var response oapi.GetSystemAnalysisResponse
	response.Body.Data = result
	return &response, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisNodes(ctx context.Context, request *oapi.ListSystemAnalysisNodesRequest) (*oapi.ListSystemAnalysisNodesResponse, error) {
	nodes, queryErr := s.db.Client(ctx).SystemAnalysisTopologyNode.Query().Where(satn.AnalysisID(request.Id)).All(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "list system analysis nodes", queryErr)
	}
	var response oapi.ListSystemAnalysisNodesResponse
	response.Body.Data = make([]oapi.SystemAnalysisNode, len(nodes))
	for i, node := range nodes {
		converted, convertErr := s.nodeFromEnt(ctx, node)
		if convertErr != nil {
			return nil, oapi.Error(ctx, "get system analysis node", convertErr)
		}
		response.Body.Data[i] = converted
	}
	response.Body.Pagination.Total = len(nodes)
	return &response, nil
}

func (s *systemAnalysisHandler) AddSystemAnalysisNode(ctx context.Context, request *oapi.AddSystemAnalysisNodeRequest) (*oapi.AddSystemAnalysisNodeResponse, error) {
	attributes := request.Body.Attributes
	created, createErr := s.db.Client(ctx).SystemAnalysisTopologyNode.Create().
		SetAnalysisID(request.Id).
		SetKnowledgeEntityID(attributes.KnowledgeEntityId).
		SetNillableReferencedAt(attributes.ReferencedAt).
		SetPosX(attributes.Position.X).
		SetPosY(attributes.Position.Y).
		SetDescription(attributes.Description).
		Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis node", createErr)
	}
	converted, convertErr := s.nodeFromEnt(ctx, created)
	if convertErr != nil {
		return nil, oapi.Error(ctx, "get system analysis node", convertErr)
	}
	var response oapi.AddSystemAnalysisNodeResponse
	response.Body.Data = converted
	return &response, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisNode(ctx context.Context, request *oapi.GetSystemAnalysisNodeRequest) (*oapi.GetSystemAnalysisNodeResponse, error) {
	node, getErr := s.db.Client(ctx).SystemAnalysisTopologyNode.Get(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get system analysis node", getErr)
	}
	converted, convertErr := s.nodeFromEnt(ctx, node)
	if convertErr != nil {
		return nil, oapi.Error(ctx, "resolve system analysis node", convertErr)
	}
	var response oapi.GetSystemAnalysisNodeResponse
	response.Body.Data = converted
	return &response, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisNode(ctx context.Context, request *oapi.UpdateSystemAnalysisNodeRequest) (*oapi.UpdateSystemAnalysisNodeResponse, error) {
	attributes := request.Body.Attributes
	update := s.db.Client(ctx).SystemAnalysisTopologyNode.UpdateOneID(request.Id).SetNillableDescription(attributes.Description)
	if attributes.Position != nil {
		update.SetPosX(attributes.Position.X).SetPosY(attributes.Position.Y)
	}
	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis node", updateErr)
	}
	converted, convertErr := s.nodeFromEnt(ctx, updated)
	if convertErr != nil {
		return nil, oapi.Error(ctx, "resolve system analysis node", convertErr)
	}
	var response oapi.UpdateSystemAnalysisNodeResponse
	response.Body.Data = converted
	return &response, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisNode(ctx context.Context, request *oapi.DeleteSystemAnalysisNodeRequest) (*oapi.DeleteSystemAnalysisNodeResponse, error) {
	deleteErr := s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		if _, contextErr := tx.IncidentTimelineEventSystemContext.Delete().
			Where(isc.SystemAnalysisNodeID(request.Id)).
			Exec(txCtx); contextErr != nil {
			return contextErr
		}
		return tx.SystemAnalysisTopologyNode.DeleteOneID(request.Id).Exec(txCtx)
	})
	if deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis node", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisNodeResponse{}, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisEdges(ctx context.Context, request *oapi.ListSystemAnalysisEdgesRequest) (*oapi.ListSystemAnalysisEdgesResponse, error) {
	edges, queryErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.Query().Where(sate.AnalysisID(request.Id)).All(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "list system analysis edges", queryErr)
	}
	var response oapi.ListSystemAnalysisEdgesResponse
	response.Body.Data = make([]oapi.SystemAnalysisEdge, len(edges))
	for i, edge := range edges {
		converted, convertErr := s.edgeFromEnt(ctx, edge)
		if convertErr != nil {
			return nil, oapi.Error(ctx, "get system analysis edge", convertErr)
		}
		response.Body.Data[i] = converted
	}
	response.Body.Pagination.Total = len(edges)
	return &response, nil
}

func (s *systemAnalysisHandler) AddSystemAnalysisEdge(ctx context.Context, request *oapi.AddSystemAnalysisEdgeRequest) (*oapi.AddSystemAnalysisEdgeResponse, error) {
	attributes := request.Body.Attributes
	created, createErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.Create().
		SetAnalysisID(request.Id).
		SetKnowledgeRelationshipID(attributes.KnowledgeRelationshipId).
		SetNillableReferencedAt(attributes.ReferencedAt).
		SetDescription(attributes.Description).
		Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis edge", createErr)
	}
	converted, convertErr := s.edgeFromEnt(ctx, created)
	if convertErr != nil {
		return nil, oapi.Error(ctx, "resolve system analysis edge", convertErr)
	}
	var response oapi.AddSystemAnalysisEdgeResponse
	response.Body.Data = converted
	return &response, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisEdge(ctx context.Context, request *oapi.GetSystemAnalysisEdgeRequest) (*oapi.GetSystemAnalysisEdgeResponse, error) {
	edge, getErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.Get(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get system analysis edge", getErr)
	}
	converted, convertErr := s.edgeFromEnt(ctx, edge)
	if convertErr != nil {
		return nil, oapi.Error(ctx, "resolve system analysis edge", convertErr)
	}
	var response oapi.GetSystemAnalysisEdgeResponse
	response.Body.Data = converted
	return &response, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisEdge(ctx context.Context, request *oapi.UpdateSystemAnalysisEdgeRequest) (*oapi.UpdateSystemAnalysisEdgeResponse, error) {
	updated, updateErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.UpdateOneID(request.Id).
		SetNillableDescription(request.Body.Attributes.Description).
		Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis edge", updateErr)
	}
	converted, convertErr := s.edgeFromEnt(ctx, updated)
	if convertErr != nil {
		return nil, oapi.Error(ctx, "resolve system analysis edge", convertErr)
	}
	var response oapi.UpdateSystemAnalysisEdgeResponse
	response.Body.Data = converted
	return &response, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisEdge(ctx context.Context, request *oapi.DeleteSystemAnalysisEdgeRequest) (*oapi.DeleteSystemAnalysisEdgeResponse, error) {
	if deleteErr := s.db.Client(ctx).SystemAnalysisTopologyEdge.DeleteOneID(request.Id).Exec(ctx); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis edge", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisEdgeResponse{}, nil
}
