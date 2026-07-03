package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	aar "github.com/rezible/rezible/ent/aiagentrun"
	"github.com/rezible/rezible/ent/predicate"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type agentsHandler struct {
	agents rez.AiSessionService
}

func newAgentsHandler(agents rez.AiSessionService) *agentsHandler {
	return &agentsHandler{agents: agents}
}

func (h *agentsHandler) RequestAiAgentRun(ctx context.Context, req *oapi.RequestAiAgentRunRequest) (*oapi.RequestAiAgentRunResponse, error) {
	var resp oapi.RequestAiAgentRunResponse
	attr := req.Body.Attributes
	params := rez.CreateAgentRunParams{
		AgentName: attr.Workflow,
		Input:     attr.Input,
	}
	run, createErr := h.agents.CreateAgentRun(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create agent task", createErr)
	}
	resp.Body.Data = oapi.AiAgentRunFromEnt(run)
	return &resp, nil
}

func (h *agentsHandler) ListAiAgentRuns(ctx context.Context, req *oapi.ListAiAgentRunsRequest) (*oapi.ListAiAgentRunsResponse, error) {
	var resp oapi.ListAiAgentRunsResponse
	var predicates []predicate.AiAgentRun
	if req.Name != "" {
		predicates = append(predicates, aar.AgentName(req.Name))
	}
	if req.Resulted.IsSet {
		p := aar.HasResult()
		if req.Resulted.Value {
			p = aar.Not(p)
		}
		predicates = append(predicates, p)
	}
	params := rez.ListAgentRunsParams{
		ListParams: req.ListParams(),
		Predicates: predicates,
	}
	runs, listErr := h.agents.ListAgentRuns(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent runs", listErr)
	}
	resp.Body.Data = make([]oapi.AiAgentRun, len(runs.Data))
	for i, run := range runs.Data {
		resp.Body.Data[i] = oapi.AiAgentRunFromEnt(run)
	}
	resp.Body.Pagination.Total = runs.Count
	return &resp, nil
}

func (h *agentsHandler) GetAiAgentRun(ctx context.Context, req *oapi.GetAiAgentRunRequest) (*oapi.GetAiAgentRunResponse, error) {
	var resp oapi.GetAiAgentRunResponse
	run, getErr := h.agents.GetAgentRun(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent run", getErr)
	}
	resp.Body.Data = oapi.AiAgentRunFromEnt(run)
	return &resp, nil
}
