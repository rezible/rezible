package apiv1

import (
	"context"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/errs"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type investigationsHandler struct {
	investigations rez.InvestigationService
	outputs        rez.InvestigationOutputService
}

func newInvestigationsHandler(investigations rez.InvestigationService, outputs rez.InvestigationOutputService) *investigationsHandler {
	return &investigationsHandler{investigations: investigations, outputs: outputs}
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

func (h *investigationsHandler) UpdateInvestigation(ctx context.Context, req *oapi.UpdateInvestigationRequest) (*oapi.UpdateInvestigationResponse, error) {
	if updateErr := h.investigations.UpdateInvestigation(ctx, req.Id); updateErr != nil {
		return nil, oapi.Error(ctx, "update investigation", updateErr)
	}
	detail, getErr := h.investigations.ReadInvestigationDetail(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get updated investigation", getErr)
	}
	var resp oapi.UpdateInvestigationResponse
	resp.Body.Data = oapi.InvestigationFromDetail(detail)
	return &resp, nil
}

func (h *investigationsHandler) GetInvestigationReport(ctx context.Context, req *oapi.GetInvestigationReportRequest) (*oapi.GetInvestigationReportResponse, error) {
	params := rez.ReadInvestigationReportParams{
		InvestigationID: req.Id,
		Selection:       rez.InvestigationReportSelection(strings.TrimSpace(req.Selection)),
	}
	report, readErr := h.outputs.ReadInvestigationReport(ctx, params)
	if readErr != nil {
		return nil, oapi.Error(ctx, "get investigation report", readErr)
	} else if report == nil {
		return nil, oapi.Error(ctx, "get investigation report", errs.ErrNotFound)
	}
	var resp oapi.GetInvestigationReportResponse
	resp.Body.Data = oapi.InvestigationReportFromEnt(report)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationFindings(ctx context.Context, request *oapi.ListInvestigationFindingsRequest) (*oapi.ListInvestigationFindingsResponse, error) {
	params := rez.ListInvestigationFindingsParams{
		ListParams:      request.ListParams(),
		InvestigationID: request.Id,
	}
	result, listErr := h.outputs.ListInvestigationFindings(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation findings", listErr)
	}
	var resp oapi.ListInvestigationFindingsResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationFindingFromEnt)
	return &resp, nil
}

func (h *investigationsHandler) GetInvestigationFinding(ctx context.Context, request *oapi.GetInvestigationFindingRequest) (*oapi.GetInvestigationFindingResponse, error) {
	finding, readErr := h.outputs.GetInvestigationFindingVersion(ctx, request.Id, request.VersionId)
	if readErr != nil {
		return nil, oapi.Error(ctx, "get investigation finding", readErr)
	}
	var resp oapi.GetInvestigationFindingResponse
	resp.Body.Data = oapi.InvestigationFindingFromEnt(finding)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationHypotheses(ctx context.Context, request *oapi.ListInvestigationHypothesesRequest) (*oapi.ListInvestigationHypothesesResponse, error) {
	params := rez.ListInvestigationHypothesesParams{
		ListParams:      request.ListParams(),
		InvestigationID: request.Id,
	}
	result, listErr := h.outputs.ListInvestigationHypotheses(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation hypotheses", listErr)
	}
	var resp oapi.ListInvestigationHypothesesResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationHypothesisFromEnt)
	return &resp, nil
}

func (h *investigationsHandler) GetInvestigationHypothesis(ctx context.Context, request *oapi.GetInvestigationHypothesisRequest) (*oapi.GetInvestigationHypothesisResponse, error) {
	hypothesis, readErr := h.outputs.GetInvestigationHypothesisVersion(ctx, request.Id, request.VersionId)
	if readErr != nil {
		return nil, oapi.Error(ctx, "get investigation hypothesis", readErr)
	}
	var resp oapi.GetInvestigationHypothesisResponse
	resp.Body.Data = oapi.InvestigationHypothesisFromEnt(hypothesis)
	return &resp, nil
}

func (h *investigationsHandler) SubmitInvestigationUserInput(ctx context.Context, request *oapi.SubmitInvestigationUserInputRequest) (*oapi.SubmitInvestigationUserInputResponse, error) {
	params := rez.SubmitInvestigationUserInputParams{
		InvestigationID: request.Id,
		Text:            strings.TrimSpace(request.Body.Text),
		SubmissionKey:   strings.TrimSpace(request.Body.SubmissionKey),
	}
	if params.Text == "" || params.SubmissionKey == "" {
		return nil, oapi.Error(ctx, "submit investigation user input", errs.ErrUnprocessableInput)
	}
	input, submitErr := h.investigations.SubmitInvestigationUserInput(ctx, params)
	if submitErr != nil {
		return nil, oapi.Error(ctx, "submit investigation user input", submitErr)
	}
	var resp oapi.SubmitInvestigationUserInputResponse
	resp.Body.Data = oapi.InvestigationUserInputFromEnt(input)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationUserInputs(ctx context.Context, request *oapi.ListInvestigationUserInputsRequest) (*oapi.ListInvestigationUserInputsResponse, error) {
	params := rez.ListInvestigationUserInputsParams{
		ListParams:      request.ListParams(),
		InvestigationID: request.Id,
	}
	result, listErr := h.investigations.ListInvestigationUserInputs(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation user inputs", listErr)
	}
	var resp oapi.ListInvestigationUserInputsResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationUserInputFromEnt)
	return &resp, nil
}

func (h *investigationsHandler) ListInvestigationEvidenceRevisions(ctx context.Context, request *oapi.ListInvestigationEvidenceRevisionsRequest) (*oapi.ListInvestigationEvidenceRevisionsResponse, error) {
	params := rez.ListInvestigationEvidenceRevisionsParams{
		ListParams:      request.ListParams(),
		InvestigationID: request.Id,
	}
	result, listErr := h.investigations.ListInvestigationEvidenceRevisions(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list investigation evidence revisions", listErr)
	}
	var resp oapi.ListInvestigationEvidenceRevisionsResponse
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.InvestigationEvidenceRevisionFromEnt)
	return &resp, nil
}
