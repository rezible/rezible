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

	body, bodyErr := oapi.MaybeConvertPaginatedResultBody(edges, oapi.SystemAnalysisEdgeFromEnt)
	if bodyErr != nil {
		return nil, oapi.Error(ctx, "list system analysis edges", bodyErr)
	}
	var resp oapi.ListSystemAnalysisEdgesResponse
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

	data, dataErr := oapi.SystemAnalysisEdgeFromEnt(edge)
	if dataErr != nil {
		return nil, oapi.Error(ctx, "add system analysis edge", dataErr)
	}
	var resp oapi.AddSystemAnalysisEdgeResponse
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

	data, dataErr := oapi.SystemAnalysisEdgeFromEnt(edge)
	if dataErr != nil {
		return nil, oapi.Error(ctx, "convert edge", dataErr)
	}
	var resp oapi.UpdateSystemAnalysisEdgeResponse
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
	listParams := request.ListParams()
	if len(request.SelectedEntryIDs) > 0 {
		listParams.Page = 1
		listParams.PageSize = len(request.SelectedEntryIDs)
	}
	preds := []predicate.SystemAnalysisEntry{sae.AnalysisID(request.Id)}
	if len(request.SelectedEntryIDs) > 0 {
		// TODO
	}
	params := rez.ListSystemAnalysisEntriesParams{
		ListParams: listParams,
		Predicates: preds,
		Kinds:      request.Kind,
	}
	entries, listErr := h.analysis.ListSystemAnalysisEntries(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list system analysis entries", listErr)
	}
	var resp oapi.ListSystemAnalysisEntriesResponse
	resp.Body = oapi.ConvertPaginatedResultBody(entries, oapi.SystemAnalysisEntryFromEnt)
	return &resp, nil
}

func (h *systemAnalysisHandler) GetSystemAnalysisEntry(ctx context.Context, request *oapi.GetSystemAnalysisEntryRequest) (*oapi.GetSystemAnalysisEntryResponse, error) {
	entry, getErr := h.analysis.LookupSystemAnalysisEntry(ctx, sae.ID(request.Id))
	if getErr != nil {
		return nil, oapi.Error(ctx, "get system analysis entry", getErr)
	}
	var resp oapi.GetSystemAnalysisEntryResponse
	resp.Body.Data = oapi.SystemAnalysisEntryFromEnt(entry)
	return &resp, nil
}

func (h *systemAnalysisHandler) CreateSystemAnalysisEntry(ctx context.Context, request *oapi.CreateSystemAnalysisEntryRequest) (*oapi.CreateSystemAnalysisEntryResponse, error) {
	attrs := request.Body.Attributes
	entry, createErr := h.analysis.SetSystemAnalysisEntry(ctx, uuid.Nil, rez.SetSystemAnalysisEntryParams{
		AnalysisID:  request.Id,
		Kind:        attrs.Kind,
		Title:       attrs.Title,
		Body:        attrs.Body,
		OccurredAt:  attrs.OccurredAt,
		SetSubjects: h.systemAnalysisEntrySubjectInputs(attrs.Subjects),
	})
	if createErr != nil {
		return nil, oapi.Error(ctx, "create system analysis entry", createErr)
	}

	var resp oapi.CreateSystemAnalysisEntryResponse
	resp.Body.Data = oapi.SystemAnalysisEntryFromEnt(entry)
	return &resp, nil
}

func (h *systemAnalysisHandler) systemAnalysisEntrySubjectInputs(subjects []oapi.SystemAnalysisEntrySubjectInput) []rez.SetSystemAnalysisEntrySubjectParams {
	inputs := make([]rez.SetSystemAnalysisEntrySubjectParams, 0, len(subjects))
	for _, subject := range subjects {
		inputs = append(inputs, rez.SetSystemAnalysisEntrySubjectParams{
			Role:                    subject.Role,
			KnowledgeEntityID:       subject.KnowledgeEntityID,
			KnowledgeRelationshipID: subject.KnowledgeRelationshipID,
			KnowledgeEvidenceID:     subject.KnowledgeEvidenceID,
		})
	}
	return inputs
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisEntry(ctx context.Context, request *oapi.UpdateSystemAnalysisEntryRequest) (*oapi.UpdateSystemAnalysisEntryResponse, error) {
	attrs := request.Body.Attributes
	params := rez.SetSystemAnalysisEntryParams{
		Kind:        attrs.Kind,
		Title:       attrs.Title,
		Body:        attrs.Body,
		OccurredAt:  attrs.OccurredAt,
		SetSubjects: h.systemAnalysisEntrySubjectInputs(attrs.Subjects),
	}
	entry, updateErr := h.analysis.SetSystemAnalysisEntry(ctx, request.Id, params)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis entry", updateErr)
	}
	var resp oapi.UpdateSystemAnalysisEntryResponse
	resp.Body.Data = oapi.SystemAnalysisEntryFromEnt(entry)
	return &resp, nil
}

func (h *systemAnalysisHandler) DeleteSystemAnalysisEntry(ctx context.Context, request *oapi.DeleteSystemAnalysisEntryRequest) (*oapi.DeleteSystemAnalysisEntryResponse, error) {
	if deleteErr := h.analysis.DeleteSystemAnalysisEntry(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis entry", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisEntryResponse{}, nil
}

func (h *systemAnalysisHandler) CreateSystemAnalysisEntrySubject(ctx context.Context, request *oapi.CreateSystemAnalysisEntrySubjectRequest) (*oapi.CreateSystemAnalysisEntrySubjectResponse, error) {
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
	}
	subject, createErr := h.analysis.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis entry subject", createErr)
	}
	var resp oapi.CreateSystemAnalysisEntrySubjectResponse
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
