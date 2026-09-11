package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	saent "github.com/rezible/rezible/ent/systemanalysisentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	sarel "github.com/rezible/rezible/ent/systemanalysisrelationship"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type systemAnalysisHandler struct {
	analysis rez.SystemAnalysisService
}

func newSystemAnalysisHandler(analysis rez.SystemAnalysisService) *systemAnalysisHandler {
	return &systemAnalysisHandler{analysis: analysis}
}

func (h *systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	analysis, queryErr := h.analysis.GetSystemAnalysis(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "get system analysis", queryErr)
	}

	var response oapi.GetSystemAnalysisResponse
	response.Body.Data = oapi.SystemAnalysisFromEnt(analysis)
	return &response, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysis(ctx context.Context, request *oapi.UpdateSystemAnalysisRequest) (*oapi.UpdateSystemAnalysisResponse, error) {
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisMutation) {
		if attrs.ScopeEntityId != nil {
			m.SetScopeEntityID(*attrs.ScopeEntityId)
		}
		if attrs.SubjectEntityId != nil {
			m.SetSubjectEntityID(*attrs.SubjectEntityId)
		}
		if attrs.ReferenceTime != nil {
			m.SetReferenceTime(*attrs.ReferenceTime)
		}
	}
	analysis, updateErr := h.analysis.SetSystemAnalysis(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis", updateErr)
	}

	var response oapi.UpdateSystemAnalysisResponse
	response.Body.Data = oapi.SystemAnalysisFromEnt(analysis)
	return &response, nil
}

func (h *systemAnalysisHandler) ListSystemAnalysisNodes(ctx context.Context, request *oapi.ListSystemAnalysisNodesRequest) (*oapi.ListSystemAnalysisNodesResponse, error) {
	params := rez.ListSystemAnalysisEntitiesParams{
		ListParams: request.ListParams(),
		Predicates: []predicate.SystemAnalysisEntity{saent.AnalysisID(request.Id)},
	}
	nodes, listErr := h.analysis.ListSystemAnalysisEntities(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list system analysis nodes", listErr)
	}

	body, bodyErr := oapi.MaybeConvertPaginatedResultBody(nodes, oapi.SystemAnalysisNodeFromEnt)
	if bodyErr != nil {
		return nil, oapi.Error(ctx, "list system analysis nodes", bodyErr)
	}
	return &oapi.ListSystemAnalysisNodesResponse{Body: *body}, nil
}

func (h *systemAnalysisHandler) AddSystemAnalysisNode(ctx context.Context, request *oapi.AddSystemAnalysisNodeRequest) (*oapi.AddSystemAnalysisNodeResponse, error) {
	var resp oapi.AddSystemAnalysisNodeResponse
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisEntityMutation) {
		m.SetAnalysisID(request.Id)
		m.SetKnowledgeEntityID(attrs.KnowledgeEntityId)
		m.SetPosX(attrs.Position.X)
		m.SetPosY(attrs.Position.Y)
		if attrs.DescriptionOverride != nil {
			m.SetDescriptionOverride(*attrs.DescriptionOverride)
		}
		if attrs.Hidden != nil {
			m.SetHidden(*attrs.Hidden)
		}
		if attrs.LabelOverride != nil {
			m.SetLabelOverride(*attrs.LabelOverride)
		}
	}
	node, createErr := h.analysis.SetSystemAnalysisEntity(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis node", createErr)
	}

	data, dataErr := oapi.SystemAnalysisNodeFromEnt(node)
	if dataErr != nil {
		return nil, oapi.Error(ctx, "add system analysis node", dataErr)
	}
	resp.Body.Data = *data
	return &resp, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisNode(ctx context.Context, request *oapi.UpdateSystemAnalysisNodeRequest) (*oapi.UpdateSystemAnalysisNodeResponse, error) {
	var resp oapi.UpdateSystemAnalysisNodeResponse
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisEntityMutation) {
		if attrs.Position != nil {
			m.SetPosX(attrs.Position.X)
			m.SetPosY(attrs.Position.Y)
		}
		if attrs.DescriptionOverride != nil {
			m.SetDescriptionOverride(*attrs.DescriptionOverride)
		}
		if attrs.Hidden != nil {
			m.SetHidden(*attrs.Hidden)
		}
		if attrs.LabelOverride != nil {
			m.SetLabelOverride(*attrs.LabelOverride)
		}
	}
	node, updateErr := h.analysis.SetSystemAnalysisEntity(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis node", updateErr)
	}

	data, dataErr := oapi.SystemAnalysisNodeFromEnt(node)
	if dataErr != nil {
		return nil, oapi.Error(ctx, "add system analysis node", dataErr)
	}
	resp.Body.Data = *data
	return &resp, nil
}

func (h *systemAnalysisHandler) DeleteSystemAnalysisNode(ctx context.Context, request *oapi.DeleteSystemAnalysisNodeRequest) (*oapi.DeleteSystemAnalysisNodeResponse, error) {
	if deleteErr := h.analysis.DeleteSystemAnalysisEntity(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis node", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisNodeResponse{}, nil
}

func (h *systemAnalysisHandler) ListSystemAnalysisEdges(ctx context.Context, request *oapi.ListSystemAnalysisEdgesRequest) (*oapi.ListSystemAnalysisEdgesResponse, error) {
	params := rez.ListSystemAnalysisRelationshipsParams{
		ListParams: request.ListParams(),
		Predicates: []predicate.SystemAnalysisRelationship{sarel.AnalysisID(request.Id)},
	}
	edges, listErr := h.analysis.ListSystemAnalysisRelationships(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list system analysis edges", listErr)
	}

	var resp oapi.ListSystemAnalysisEdgesResponse
	body, bodyErr := oapi.MaybeConvertPaginatedResultBody(edges, oapi.SystemAnalysisEdgeFromEnt)
	if bodyErr != nil {
		return nil, oapi.Error(ctx, "list system analysis edges", bodyErr)
	}
	resp.Body = *body
	return &resp, nil
}

func (h *systemAnalysisHandler) AddSystemAnalysisEdge(ctx context.Context, request *oapi.AddSystemAnalysisEdgeRequest) (*oapi.AddSystemAnalysisEdgeResponse, error) {
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisRelationshipMutation) {
		m.SetAnalysisID(request.Id)
		m.SetKnowledgeRelationshipID(attrs.KnowledgeRelationshipId)
		if attrs.DescriptionOverride != nil {
			m.SetDescriptionOverride(*attrs.DescriptionOverride)
		}
		if attrs.Hidden != nil {
			m.SetHidden(*attrs.Hidden)
		}
		if attrs.LabelOverride != nil {
			m.SetLabelOverride(*attrs.LabelOverride)
		}
	}
	edge, createErr := h.analysis.SetSystemAnalysisRelationship(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis edge", createErr)
	}

	var resp oapi.AddSystemAnalysisEdgeResponse
	data, dataErr := oapi.SystemAnalysisEdgeFromEnt(edge)
	if dataErr != nil {
		return nil, oapi.Error(ctx, "add system analysis edge", dataErr)
	}
	resp.Body.Data = *data
	return &resp, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisEdge(ctx context.Context, request *oapi.UpdateSystemAnalysisEdgeRequest) (*oapi.UpdateSystemAnalysisEdgeResponse, error) {
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisRelationshipMutation) {
		if attrs.DescriptionOverride != nil {
			m.SetDescriptionOverride(*attrs.DescriptionOverride)
		}
		if attrs.Hidden != nil {
			m.SetHidden(*attrs.Hidden)
		}
		if attrs.LabelOverride != nil {
			m.SetLabelOverride(*attrs.LabelOverride)
		}
	}
	edge, updateErr := h.analysis.SetSystemAnalysisRelationship(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis edge", updateErr)
	}

	var resp oapi.UpdateSystemAnalysisEdgeResponse
	data, dataErr := oapi.SystemAnalysisEdgeFromEnt(edge)
	if dataErr != nil {
		return nil, oapi.Error(ctx, "convert edge", dataErr)
	}
	resp.Body.Data = *data
	return &resp, nil
}

func (h *systemAnalysisHandler) DeleteSystemAnalysisEdge(ctx context.Context, request *oapi.DeleteSystemAnalysisEdgeRequest) (*oapi.DeleteSystemAnalysisEdgeResponse, error) {
	if deleteErr := h.analysis.DeleteSystemAnalysisRelationship(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis edge", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisEdgeResponse{}, nil
}

func (h *systemAnalysisHandler) ListSystemAnalysisEntries(ctx context.Context, request *oapi.ListSystemAnalysisEntriesRequest) (*oapi.ListSystemAnalysisEntriesResponse, error) {
	var resp oapi.ListSystemAnalysisEntriesResponse
	params := rez.ListSystemAnalysisEntriesParams{
		ListParams: request.ListParams(),
		Predicates: []predicate.SystemAnalysisEntry{sae.AnalysisID(request.Id)},
	}
	entries, listErr := h.analysis.ListSystemAnalysisEntries(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list system analysis entries", listErr)
	}
	resp.Body = oapi.ConvertPaginatedResultBody(entries, oapi.SystemAnalysisEntryFromEnt)
	return &resp, nil
}

func (h *systemAnalysisHandler) GetSystemAnalysisEntry(ctx context.Context, request *oapi.GetSystemAnalysisEntryRequest) (*oapi.GetSystemAnalysisEntryResponse, error) {
	entry, getErr := h.analysis.LookupSystemAnalysisEntry(ctx, sae.ID(request.Id))
	if getErr != nil {
		return nil, oapi.Error(ctx, "get system analysis entry", getErr)
	}
	var response oapi.GetSystemAnalysisEntryResponse
	response.Body.Data = oapi.SystemAnalysisEntryFromEnt(entry)
	return &response, nil
}

func (h *systemAnalysisHandler) CreateSystemAnalysisEntry(ctx context.Context, request *oapi.CreateSystemAnalysisEntryRequest) (*oapi.CreateSystemAnalysisEntryResponse, error) {
	var resp oapi.CreateSystemAnalysisEntryResponse
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(request.Id)
		if attrs.Reference != nil {
			m.SetReference(*attrs.Reference)
		}
		m.SetKind(sae.Kind(attrs.Kind))
		if attrs.OccurredAt != nil {
			m.SetOccurredAt(*attrs.OccurredAt)
		}
		m.SetTitle(attrs.Title)
		if attrs.Body != nil {
			m.SetBody(*attrs.Body)
		}
		if attrs.Properties != nil {
			m.SetProperties(attrs.Properties)
		}
	}
	entry, createErr := h.analysis.SetSystemAnalysisEntry(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create system analysis entry", createErr)
	}

	resp.Body.Data = oapi.SystemAnalysisEntryFromEnt(entry)
	return &resp, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisEntry(ctx context.Context, request *oapi.UpdateSystemAnalysisEntryRequest) (*oapi.UpdateSystemAnalysisEntryResponse, error) {
	var resp oapi.UpdateSystemAnalysisEntryResponse
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisEntryMutation) {
		if attrs.Kind != nil {
			m.SetKind(sae.Kind(*attrs.Kind))
		}
		if attrs.OccurredAt != nil {
			m.SetOccurredAt(*attrs.OccurredAt)
		}
		if attrs.Title != nil {
			m.SetTitle(*attrs.Title)
		}
		if attrs.Body != nil {
			m.SetBody(*attrs.Body)
		}
		if attrs.Sequence != nil {
			m.SetSequence(*attrs.Sequence)
		}
		if attrs.Properties != nil {
			m.SetProperties(attrs.Properties)
		}
	}
	entry, updateErr := h.analysis.SetSystemAnalysisEntry(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis entry", updateErr)
	}

	resp.Body.Data = oapi.SystemAnalysisEntryFromEnt(entry)
	return &resp, nil
}

func (h *systemAnalysisHandler) DeleteSystemAnalysisEntry(ctx context.Context, request *oapi.DeleteSystemAnalysisEntryRequest) (*oapi.DeleteSystemAnalysisEntryResponse, error) {
	if deleteErr := h.analysis.DeleteSystemAnalysisEntry(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis entry", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisEntryResponse{}, nil
}

func (h *systemAnalysisHandler) AddSystemAnalysisEntrySubject(ctx context.Context, request *oapi.AddSystemAnalysisEntrySubjectRequest) (*oapi.AddSystemAnalysisEntrySubjectResponse, error) {
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetEntryID(request.Id)
		m.SetRole(attrs.Role)
		if attrs.KnowledgeEntityId != nil {
			m.SetKnowledgeEntityID(*attrs.KnowledgeEntityId)
		}
		if attrs.KnowledgeRelationshipId != nil {
			m.SetKnowledgeRelationshipID(*attrs.KnowledgeRelationshipId)
		}
		if attrs.KnowledgeEvidenceId != nil {
			m.SetKnowledgeEvidenceID(*attrs.KnowledgeEvidenceId)
		}
		if attrs.NormalizedEventId != nil {
			m.SetNormalizedEventID(*attrs.NormalizedEventId)
		}
	}
	subject, createErr := h.analysis.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis entry subject", createErr)
	}

	var resp oapi.AddSystemAnalysisEntrySubjectResponse
	resp.Body.Data = oapi.SystemAnalysisEntrySubjectFromEnt(subject)
	return &resp, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisEntrySubject(ctx context.Context, request *oapi.UpdateSystemAnalysisEntrySubjectRequest) (*oapi.UpdateSystemAnalysisEntrySubjectResponse, error) {
	setFn := func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetRole(request.Body.Attributes.Role)
	}
	subject, updateErr := h.analysis.SetSystemAnalysisEntrySubject(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis entry subject", updateErr)
	}

	var resp oapi.UpdateSystemAnalysisEntrySubjectResponse
	resp.Body.Data = oapi.SystemAnalysisEntrySubjectFromEnt(subject)
	return &resp, nil
}

func (h *systemAnalysisHandler) DeleteSystemAnalysisEntrySubject(ctx context.Context, request *oapi.DeleteSystemAnalysisEntrySubjectRequest) (*oapi.DeleteSystemAnalysisEntrySubjectResponse, error) {
	if deleteErr := h.analysis.DeleteSystemAnalysisEntrySubject(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis entry subject", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisEntrySubjectResponse{}, nil
}
