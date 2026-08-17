package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
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
	response.Body.Data = oapi.SystemAnalysisWithEntriesFromEnt(analysis)
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
	response.Body.Data = oapi.SystemAnalysisWithEntriesFromEnt(analysis)
	return &response, nil
}

func (h *systemAnalysisHandler) GetSystemAnalysisGraph(ctx context.Context, request *oapi.GetSystemAnalysisGraphRequest) (*oapi.GetSystemAnalysisGraphResponse, error) {
	params := rez.GetKnowledgeGraphViewParams{
		Depth:             request.Depth,
		RelationshipKinds: request.RelationshipKind,
	}
	view, viewErr := h.analysis.GetSystemAnalysisGraph(ctx, request.Id, params)
	if viewErr != nil {
		return nil, oapi.Error(ctx, "get system analysis graph", viewErr)
	}

	var response oapi.GetSystemAnalysisGraphResponse
	response.Body.Data = oapi.KnowledgeGraphViewFromRez(view)
	return &response, nil
}

func (h *systemAnalysisHandler) ListSystemAnalysisEntries(ctx context.Context, request *oapi.ListSystemAnalysisEntriesRequest) (*oapi.ListSystemAnalysisEntriesResponse, error) {
	entries, listErr := h.analysis.ListSystemAnalysisEntries(ctx, request.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list system analysis entries", listErr)
	}

	var response oapi.ListSystemAnalysisEntriesResponse
	response.Body.Data = oapi.ConvertSlice(entries, oapi.SystemAnalysisEntryWithSubjectsFromEnt)
	response.Body.Pagination.Total = len(response.Body.Data)
	return &response, nil
}

func (h *systemAnalysisHandler) CreateSystemAnalysisEntry(ctx context.Context, request *oapi.CreateSystemAnalysisEntryRequest) (*oapi.CreateSystemAnalysisEntryResponse, error) {
	attrs := request.Body.Attributes
	setFn := func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(request.Id)
		m.SetKind(sae.Kind(attrs.Kind))
		if attrs.OccurredAt != nil {
			m.SetOccurredAt(*attrs.OccurredAt)
		}
		m.SetSequence(attrs.Sequence)
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

	var response oapi.CreateSystemAnalysisEntryResponse
	response.Body.Data = oapi.SystemAnalysisEntryWithSubjectsFromEnt(entry)
	return &response, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisEntry(ctx context.Context, request *oapi.UpdateSystemAnalysisEntryRequest) (*oapi.UpdateSystemAnalysisEntryResponse, error) {
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

	var response oapi.UpdateSystemAnalysisEntryResponse
	response.Body.Data = oapi.SystemAnalysisEntryWithSubjectsFromEnt(entry)
	return &response, nil
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
	}
	subject, createErr := h.analysis.SetSystemAnalysisEntrySubject(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, oapi.Error(ctx, "add system analysis entry subject", createErr)
	}

	var response oapi.AddSystemAnalysisEntrySubjectResponse
	response.Body.Data = oapi.SystemAnalysisEntrySubjectFromEnt(subject)
	return &response, nil
}

func (h *systemAnalysisHandler) UpdateSystemAnalysisEntrySubject(ctx context.Context, request *oapi.UpdateSystemAnalysisEntrySubjectRequest) (*oapi.UpdateSystemAnalysisEntrySubjectResponse, error) {
	setFn := func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetRole(request.Body.Attributes.Role)
	}
	subject, updateErr := h.analysis.SetSystemAnalysisEntrySubject(ctx, request.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update system analysis entry subject", updateErr)
	}

	var response oapi.UpdateSystemAnalysisEntrySubjectResponse
	response.Body.Data = oapi.SystemAnalysisEntrySubjectFromEnt(subject)
	return &response, nil
}

func (h *systemAnalysisHandler) DeleteSystemAnalysisEntrySubject(ctx context.Context, request *oapi.DeleteSystemAnalysisEntrySubjectRequest) (*oapi.DeleteSystemAnalysisEntrySubjectResponse, error) {
	if deleteErr := h.analysis.DeleteSystemAnalysisEntrySubject(ctx, request.Id); deleteErr != nil {
		return nil, oapi.Error(ctx, "delete system analysis entry subject", deleteErr)
	}
	return &oapi.DeleteSystemAnalysisEntrySubjectResponse{}, nil
}
