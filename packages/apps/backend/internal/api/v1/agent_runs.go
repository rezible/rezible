package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentrun"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type agentRunsHandler struct {
	agents rez.AgentService
}

func newAgentRunsHandler(agents rez.AgentService) *agentRunsHandler {
	return &agentRunsHandler{agents: agents}
}

func (h *agentRunsHandler) ListAgentRuns(ctx context.Context, req *oapi.ListAgentRunsRequest) (*oapi.ListAgentRunsResponse, error) {
	var resp oapi.ListAgentRunsResponse

	params := rez.ListAgentRunsParams{
		ListParams:   req.ListParams(),
		WorkflowKind: agentrun.WorkflowKind(req.WorkflowKind),
		Status:       agentrun.Status(req.Status),
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

func (h *agentRunsHandler) GetAgentRun(ctx context.Context, req *oapi.GetAgentRunRequest) (*oapi.GetAgentRunResponse, error) {
	var resp oapi.GetAgentRunResponse

	run, getErr := h.agents.GetRun(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent run", getErr)
	}
	resp.Body.Data = oapi.AgentRunFromEnt(run)
	return &resp, nil
}

func (h *agentRunsHandler) ListAgentRunArtifacts(ctx context.Context, req *oapi.ListAgentRunArtifactsRequest) (*oapi.ListAgentRunArtifactsResponse, error) {
	var resp oapi.ListAgentRunArtifactsResponse

	artifacts, listErr := h.agents.ListRunArtifacts(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent run artifacts", listErr)
	}
	resp.Body.Data = make([]oapi.AgentRunArtifact, len(artifacts))
	for i, artifact := range artifacts {
		resp.Body.Data[i] = oapi.AgentRunArtifactFromEnt(artifact)
	}
	resp.Body.Pagination.Total = len(artifacts)
	return &resp, nil
}
