package apiv1

import (
	"context"
	"strings"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type investigationsHandler struct {
	investigations rez.InvestigationService
}

func newInvestigationsHandler(investigations rez.InvestigationService) *investigationsHandler {
	return &investigationsHandler{investigations: investigations}
}

func (h *investigationsHandler) GetInvestigation(ctx context.Context, req *oapi.GetInvestigationRequest) (*oapi.GetInvestigationResponse, error) {
	detail, getErr := h.investigations.ReadInvestigationDetail(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get investigation", getErr)
	}
	var resp oapi.GetInvestigationResponse
	resp.Body.Data = oapi.InvestigationFromDetail(detail)
	return &resp, nil
}

func (h *investigationsHandler) getReportSelection(sel string) (rez.InvestigationReportSelection, error) {
	switch strings.TrimSpace(sel) {
	case "", string(rez.InvestigationReportSelectionLatest):
		return rez.InvestigationReportSelectionLatest, nil
	case string(rez.InvestigationReportSelectionCompleted):
		return rez.InvestigationReportSelectionCompleted, nil
	}
	return "", oapi.Err422InvalidInput(rez.ErrInvalidInput)
}

func (h *investigationsHandler) GetInvestigationReport(ctx context.Context, req *oapi.GetInvestigationReportRequest) (*oapi.GetInvestigationReportResponse, error) {
	selection, selectionErr := h.getReportSelection(req.Selection)
	if selectionErr != nil {
		return nil, oapi.Error(ctx, "get investigation report", selectionErr)
	}
	params := rez.ReadInvestigationReportParams{Selection: selection}
	report, readErr := h.investigations.ReadInvestigationReport(ctx, req.Id, params)
	if readErr != nil {
		return nil, oapi.Error(ctx, "get investigation report", readErr)
	} else if report == nil {
		return nil, oapi.Error(ctx, "get investigation report", oapi.Err404NotFound())
	}
	var resp oapi.GetInvestigationReportResponse
	resp.Body.Data = oapi.InvestigationReportFromResult(report)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationFindings(ctx context.Context, request *oapi.ListInvestigationFindingsRequest) (*oapi.ListInvestigationFindingsResponse, error) {
	result, listErr := h.investigations.ListInvestigationFindings(ctx, request.Id, request.ListParams())
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation findings", listErr)
	}
	var resp oapi.ListInvestigationFindingsResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationFindingFromResult)
	return &resp, nil
}

func (h *investigationsHandler) GetInvestigationFinding(ctx context.Context, request *oapi.GetInvestigationFindingRequest) (*oapi.GetInvestigationFindingResponse, error) {
	finding, readErr := h.investigations.GetInvestigationFindingVersion(ctx, request.Id, request.VersionId)
	if readErr != nil {
		return nil, oapi.Error(ctx, "get investigation finding", readErr)
	}
	var resp oapi.GetInvestigationFindingResponse
	resp.Body.Data = oapi.InvestigationFindingFromResult(finding)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationHypotheses(ctx context.Context, request *oapi.ListInvestigationHypothesesRequest) (*oapi.ListInvestigationHypothesesResponse, error) {
	result, listErr := h.investigations.ListInvestigationHypotheses(ctx, request.Id, request.ListParams())
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation hypotheses", listErr)
	}
	var resp oapi.ListInvestigationHypothesesResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationHypothesisFromResult)
	return &resp, nil
}

func (h *investigationsHandler) GetInvestigationHypothesis(ctx context.Context, request *oapi.GetInvestigationHypothesisRequest) (*oapi.GetInvestigationHypothesisResponse, error) {
	hypothesis, readErr := h.investigations.GetInvestigationHypothesisVersion(ctx, request.Id, request.VersionId)
	if readErr != nil {
		return nil, oapi.Error(ctx, "get investigation hypothesis", readErr)
	}
	var resp oapi.GetInvestigationHypothesisResponse
	resp.Body.Data = oapi.InvestigationHypothesisFromResult(hypothesis)
	return &resp, nil
}

func (h *investigationsHandler) SubmitInvestigationUserInput(ctx context.Context, request *oapi.SubmitInvestigationUserInputRequest) (*oapi.SubmitInvestigationUserInputResponse, error) {
	params := rez.SubmitInvestigationUserInputParams{
		InvestigationID: request.Id,
		Text:            strings.TrimSpace(request.Body.Text),
		SubmissionKey:   strings.TrimSpace(request.Body.SubmissionKey),
	}
	if params.Text == "" || params.SubmissionKey == "" {
		return nil, oapi.Error(ctx, "submit investigation user input", oapi.Err422InvalidInput(rez.ErrInvalidInput))
	}
	input, submitErr := h.investigations.SubmitInvestigationUserInput(ctx, params)
	if submitErr != nil {
		return nil, oapi.Error(ctx, "submit investigation user input", submitErr)
	}
	var resp oapi.SubmitInvestigationUserInputResponse
	resp.Body.Data = oapi.InvestigationUserInputFromRez(input)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationUserInputs(ctx context.Context, request *oapi.ListInvestigationUserInputsRequest) (*oapi.ListInvestigationUserInputsResponse, error) {
	result, listErr := h.investigations.ListInvestigationUserInputs(ctx, request.Id, request.ListParams())
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation user inputs", listErr)
	}
	var resp oapi.ListInvestigationUserInputsResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationUserInputFromRez)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationEvidenceRevisions(ctx context.Context, request *oapi.ListInvestigationEvidenceRevisionsRequest) (*oapi.ListInvestigationEvidenceRevisionsResponse, error) {
	result, listErr := h.investigations.ListInvestigationEvidenceRevisions(ctx, request.Id, request.ListParams())
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation evidence revisions", listErr)
	}
	var resp oapi.ListInvestigationEvidenceRevisionsResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationEvidenceRevisionFromEnt)
	return &resp, nil
}
