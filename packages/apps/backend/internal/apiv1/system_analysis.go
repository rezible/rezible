package apiv1

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/ent/systemanalysisnode"
	"github.com/rezible/rezible/ent/systemanalysistopologyedge"
	"github.com/rezible/rezible/ent/topologysnapshot"
	"github.com/rezible/rezible/ent/topologysnapshotentity"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type systemAnalysisHandler struct {
	db *ent.Client
}

func newSystemAnalysisHandler(db *ent.Client) *systemAnalysisHandler {
	return &systemAnalysisHandler{db: db}
}

func snapshotAliases(aliases []*ent.KnowledgeEntityAlias) []map[string]any {
	res := make([]map[string]any, len(aliases))
	for i, alias := range aliases {
		res[i] = map[string]any{
			"id":             alias.ID.String(),
			"provider":       alias.Provider,
			"providerSource": alias.ProviderSource,
			"subjectKind":    alias.SubjectKind,
			"subjectRef":     alias.SubjectRef,
			"firstSeenAt":    alias.FirstSeenAt,
			"lastSeenAt":     alias.LastSeenAt,
		}
	}
	return res
}

func ensureAnalysisSnapshot(ctx context.Context, tx *ent.Tx, analysis *ent.SystemAnalysis) (*ent.SystemAnalysis, error) {
	if analysis.TopologySnapshotID != nil {
		return analysis, nil
	}

	snapshot, createSnapshotErr := tx.TopologySnapshot.Create().
		SetScope(topologysnapshot.ScopeAnalysis).
		SetScopeProperties(map[string]any{
			"analysisId": analysis.ID.String(),
		}).
		Save(ctx)
	if createSnapshotErr != nil {
		return nil, fmt.Errorf("create topology snapshot: %w", createSnapshotErr)
	}

	updated, updateErr := tx.SystemAnalysis.UpdateOneID(analysis.ID).
		SetTopologySnapshotID(snapshot.ID).
		Save(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("attach topology snapshot: %w", updateErr)
	}
	return updated, nil
}

func copyKnowledgeEntityToSnapshot(ctx context.Context, tx *ent.Tx, snapshotID uuid.UUID, knowledgeEntityID uuid.UUID) (*ent.TopologySnapshotEntity, error) {
	existing, existingErr := tx.TopologySnapshotEntity.Query().
		Where(
			topologysnapshotentity.SnapshotID(snapshotID),
			topologysnapshotentity.KnowledgeEntityID(knowledgeEntityID),
		).
		Only(ctx)
	if existingErr == nil {
		return existing, nil
	}
	if !ent.IsNotFound(existingErr) {
		return nil, fmt.Errorf("query existing snapshot entity: %w", existingErr)
	}

	entity, entityErr := tx.KnowledgeEntity.Query().
		Where(knowledgeentity.ID(knowledgeEntityID)).
		WithAliases().
		Only(ctx)
	if entityErr != nil {
		return nil, fmt.Errorf("query knowledge entity: %w", entityErr)
	}

	snapshotEntity, createErr := tx.TopologySnapshotEntity.Create().
		SetSnapshotID(snapshotID).
		SetKnowledgeEntityID(entity.ID).
		SetEntityKind(entity.Kind).
		SetDisplayName(entity.DisplayName).
		SetDescription(entity.Description).
		SetProperties(entity.Properties).
		SetAliases(snapshotAliases(entity.Edges.Aliases)).
		Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create snapshot entity: %w", createErr)
	}
	return snapshotEntity, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	var resp oapi.GetSystemAnalysisResponse

	analysis, queryErr := s.db.SystemAnalysis.Query().
		Where(systemanalysis.ID(request.Id)).
		WithTopologySnapshot(func(q *ent.TopologySnapshotQuery) {
			q.WithEntities()
			q.WithRelationships()
		}).
		WithAnalysisNodes(func(q *ent.SystemAnalysisNodeQuery) {
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

	nodes, queryErr := s.db.SystemAnalysisNode.Query().
		Where(systemanalysisnode.AnalysisID(request.Id)).
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

	attr := request.Body.Attributes
	if attr.SnapshotEntityId == nil && attr.KnowledgeEntityId == nil {
		return nil, huma.Error400BadRequest("snapshotEntityId or knowledgeEntityId is required")
	}

	var added *ent.SystemAnalysisNode
	var addedSnapshotEntityID uuid.UUID
	txErr := ent.WithTx(ctx, s.db, func(tx *ent.Tx) error {
		snapshotEntityID := attr.SnapshotEntityId
		if snapshotEntityID == nil {
			analysis, analysisErr := tx.SystemAnalysis.Get(ctx, request.Id)
			if analysisErr != nil {
				return fmt.Errorf("query analysis: %w", analysisErr)
			}
			analysis, analysisErr = ensureAnalysisSnapshot(ctx, tx, analysis)
			if analysisErr != nil {
				return analysisErr
			}
			snapshotEntity, snapshotEntityErr := copyKnowledgeEntityToSnapshot(ctx, tx, *analysis.TopologySnapshotID, *attr.KnowledgeEntityId)
			if snapshotEntityErr != nil {
				return snapshotEntityErr
			}
			snapshotEntityID = &snapshotEntity.ID
		}

		var addErr error
		added, addErr = tx.SystemAnalysisNode.Create().
			SetAnalysisID(request.Id).
			SetSnapshotEntityID(*snapshotEntityID).
			SetPosY(attr.Position.Y).
			SetPosX(attr.Position.X).
			SetDescription(attr.Description).
			Save(ctx)
		if addErr != nil {
			return fmt.Errorf("create analysis node: %w", addErr)
		}
		addedSnapshotEntityID = *snapshotEntityID
		return nil
	})
	if txErr != nil {
		return nil, oapi.Error(ctx, "failed to add system analysis node", txErr)
	}
	added.Edges.SnapshotEntity, _ = s.db.TopologySnapshotEntity.Get(ctx, addedSnapshotEntityID)
	resp.Body.Data = oapi.SystemAnalysisNodeFromEnt(added)

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisNode(ctx context.Context, request *oapi.GetSystemAnalysisNodeRequest) (*oapi.GetSystemAnalysisNodeResponse, error) {
	var resp oapi.GetSystemAnalysisNodeResponse

	node, getErr := s.db.SystemAnalysisNode.Query().
		Where(systemanalysisnode.ID(request.Id)).
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
	update := s.db.SystemAnalysisNode.UpdateOneID(request.Id).
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

	if delErr := s.db.SystemAnalysisNode.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, oapi.Error(ctx, "failed to delete system analysis node", delErr)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisEdges(ctx context.Context, request *oapi.ListSystemAnalysisEdgesRequest) (*oapi.ListSystemAnalysisEdgesResponse, error) {
	var resp oapi.ListSystemAnalysisEdgesResponse

	edges, queryErr := s.db.SystemAnalysisTopologyEdge.Query().
		Where(systemanalysistopologyedge.AnalysisID(request.Id)).
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
	created, createErr := s.db.SystemAnalysisTopologyEdge.Create().
		SetAnalysisID(request.Id).
		SetSnapshotRelationshipID(attr.SnapshotRelationshipId).
		SetDescription(attr.Description).
		Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to add system analysis edge", createErr)
	}
	created.Edges.SnapshotRelationship, _ = s.db.TopologySnapshotRelationship.Get(ctx, attr.SnapshotRelationshipId)
	resp.Body.Data = oapi.SystemAnalysisEdgeFromEnt(created)

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisEdge(ctx context.Context, request *oapi.GetSystemAnalysisEdgeRequest) (*oapi.GetSystemAnalysisEdgeResponse, error) {
	var resp oapi.GetSystemAnalysisEdgeResponse

	edge, getErr := s.db.SystemAnalysisTopologyEdge.Query().
		Where(systemanalysistopologyedge.ID(request.Id)).
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

	updated, updateErr := s.db.SystemAnalysisTopologyEdge.UpdateOneID(request.Id).
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

	if delErr := s.db.SystemAnalysisTopologyEdge.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, oapi.Error(ctx, "failed to delete system analysis edge", delErr)
	}

	return &resp, nil
}
