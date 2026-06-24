package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentcase"
	"github.com/rezible/rezible/ent/agentrun"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type agentsHandler struct {
	agents rez.AgentService
}

func newAgentsHandler(agents rez.AgentService) *agentsHandler {
	return &agentsHandler{agents: agents}
}

func (h *agentsHandler) CreateAgentCase(ctx context.Context, req *oapi.CreateAgentCaseRequest) (*oapi.CreateAgentCaseResponse, error) {
	var resp oapi.CreateAgentCaseResponse
	attr := req.Body.Attributes
	c, createErr := h.agents.CreateCase(ctx, rez.AgentCaseRequest{
		Title:           attr.Title,
		Query:           attr.Query,
		WorkflowKind:    agentrun.WorkflowKind(attr.WorkflowKind),
		SubjectKind:     attr.SubjectKind,
		SubjectID:       attr.SubjectId,
		TriggerMetadata: attr.TriggerMetadata,
	})
	if createErr != nil {
		return nil, oapi.Error(ctx, "create agent case", createErr)
	}
	resp.Body.Data = oapi.AgentCaseFromEnt(c)
	return &resp, nil
}

func (h *agentsHandler) ListAgentCases(ctx context.Context, req *oapi.ListAgentCasesRequest) (*oapi.ListAgentCasesResponse, error) {
	var resp oapi.ListAgentCasesResponse
	cases, listErr := h.agents.ListCases(ctx, rez.ListAgentCasesParams{
		ListParams:   req.ListParams(),
		Status:       agentcase.Status(req.Status),
		WorkflowKind: agentrun.WorkflowKind(req.WorkflowKind),
		SubjectKind:  req.SubjectKind,
		SubjectID:    req.SubjectId,
	})
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent cases", listErr)
	}
	resp.Body.Data = make([]oapi.AgentCase, len(cases.Data))
	for i, c := range cases.Data {
		resp.Body.Data[i] = oapi.AgentCaseFromEnt(c)
	}
	resp.Body.Pagination.Total = cases.Count
	return &resp, nil
}

func (h *agentsHandler) GetAgentCase(ctx context.Context, req *oapi.GetAgentCaseRequest) (*oapi.GetAgentCaseResponse, error) {
	var resp oapi.GetAgentCaseResponse
	c, getErr := h.agents.GetCase(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent case", getErr)
	}
	resp.Body.Data = oapi.AgentCaseFromEnt(c)
	return &resp, nil
}

func (h *agentsHandler) ListAgentCaseSteps(ctx context.Context, req *oapi.ListAgentCaseStepsRequest) (*oapi.ListAgentCaseStepsResponse, error) {
	var resp oapi.ListAgentCaseStepsResponse
	steps, listErr := h.agents.ListCaseSteps(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent case steps", listErr)
	}
	resp.Body.Data = make([]oapi.AgentCaseStep, len(steps))
	for i, step := range steps {
		resp.Body.Data[i] = oapi.AgentCaseStepFromEnt(step)
	}
	resp.Body.Pagination.Total = len(steps)
	return &resp, nil
}

func (h *agentsHandler) ListAgentCaseArtifacts(ctx context.Context, req *oapi.ListAgentCaseArtifactsRequest) (*oapi.ListAgentCaseArtifactsResponse, error) {
	var resp oapi.ListAgentCaseArtifactsResponse
	artifacts, listErr := h.agents.ListCaseArtifacts(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent case artifacts", listErr)
	}
	resp.Body.Data = make([]oapi.AgentCaseArtifact, len(artifacts))
	for i, artifact := range artifacts {
		resp.Body.Data[i] = oapi.AgentCaseArtifactFromEnt(artifact)
	}
	resp.Body.Pagination.Total = len(artifacts)
	return &resp, nil
}

func (h *agentsHandler) ListAgentCaseConclusions(ctx context.Context, req *oapi.ListAgentCaseConclusionsRequest) (*oapi.ListAgentCaseConclusionsResponse, error) {
	var resp oapi.ListAgentCaseConclusionsResponse
	conclusions, listErr := h.agents.ListCaseConclusions(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent case conclusions", listErr)
	}
	resp.Body.Data = make([]oapi.AgentCaseConclusion, len(conclusions))
	for i, conclusion := range conclusions {
		resp.Body.Data[i] = oapi.AgentCaseConclusionFromEnt(conclusion)
	}
	resp.Body.Pagination.Total = len(conclusions)
	return &resp, nil
}

func (h *agentsHandler) RequestAgentCaseRun(ctx context.Context, req *oapi.RequestAgentCaseRunRequest) (*oapi.RequestAgentCaseRunResponse, error) {
	var resp oapi.RequestAgentCaseRunResponse
	attr := req.Body.Attributes
	run, runErr := h.agents.RequestCaseRun(ctx, rez.AgentCaseRunRequest{
		AgentCaseID:    req.Id,
		IdempotencyKey: attr.IdempotencyKey,
		Metadata:       attr.Metadata,
	})
	if runErr != nil {
		return nil, oapi.Error(ctx, "request agent case run", runErr)
	}
	resp.Body.Data = oapi.AgentRunFromEnt(run)
	return &resp, nil
}

func (h *agentsHandler) ListAgentRuns(ctx context.Context, req *oapi.ListAgentRunsRequest) (*oapi.ListAgentRunsResponse, error) {
	var resp oapi.ListAgentRunsResponse

	params := rez.ListAgentRunsParams{
		ListParams:   req.ListParams(),
		WorkflowKind: agentrun.WorkflowKind(req.WorkflowKind),
		Status:       agentrun.Status(req.Status),
		AgentCaseID:  req.AgentCaseId,
		SubjectKind:  req.SubjectKind,
		SubjectID:    req.SubjectId,
	}
	runs, listErr := h.agents.ListRuns(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent runs", listErr)
	}
	resp.Body.Data = make([]oapi.AgentRun, len(runs.Data))
	for i, run := range runs.Data {
		resp.Body.Data[i] = oapi.AgentRunFromEnt(run)
	}
	resp.Body.Pagination.Total = runs.Count
	return &resp, nil
}

func (h *agentsHandler) GetAgentRun(ctx context.Context, req *oapi.GetAgentRunRequest) (*oapi.GetAgentRunResponse, error) {
	var resp oapi.GetAgentRunResponse

	run, getErr := h.agents.GetRun(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent run", getErr)
	}
	resp.Body.Data = oapi.AgentRunFromEnt(run)
	return &resp, nil
}
