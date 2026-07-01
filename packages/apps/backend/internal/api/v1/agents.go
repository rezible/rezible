package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/predicate"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type agentsHandler struct {
	agents rez.AgentService
}

func newAgentsHandler(agents rez.AgentService) *agentsHandler {
	return &agentsHandler{agents: agents}
}

func (h *agentsHandler) RequestAgentRun(ctx context.Context, req *oapi.RequestAgentRunRequest) (*oapi.RequestAgentRunResponse, error) {
	var resp oapi.RequestAgentRunResponse
	attr := req.Body.Attributes
	params := rez.CreateAgentRunParams{
		Workflow:    attr.Workflow,
		Input:       attr.Input,
		TriggerKind: "manual",
	}
	run, createErr := h.agents.CreateRun(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create agent task", createErr)
	}
	resp.Body.Data = oapi.AgentRunFromEnt(run)
	return &resp, nil
}

func (h *agentsHandler) ListAgentRuns(ctx context.Context, req *oapi.ListAgentRunsRequest) (*oapi.ListAgentRunsResponse, error) {
	var resp oapi.ListAgentRunsResponse
	var predicates []predicate.AgentRun
	if req.Workflow != "" {
		predicates = append(predicates, agentrun.Workflow(req.Workflow))
	}
	if req.Resulted.IsSet {
		p := agentrun.HasResult()
		if req.Resulted.Value {
			p = agentrun.Not(p)
		}
		predicates = append(predicates, p)
	}
	params := rez.ListAgentRunsParams{
		ListParams: req.ListParams(),
		Predicates: predicates,
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
