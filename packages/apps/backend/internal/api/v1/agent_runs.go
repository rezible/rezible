package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentrun"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type agentsHandler struct {
	agents rez.AgentService
}

func newAgentsHandler(agents rez.AgentService) *agentsHandler {
	return &agentsHandler{agents: agents}
}

func (h *agentsHandler) CreateAgentTask(ctx context.Context, req *oapi.CreateAgentTaskRequest) (*oapi.CreateAgentTaskResponse, error) {
	var resp oapi.CreateAgentTaskResponse
	attr := req.Body.Attributes
	task, createErr := h.agents.CreateTask(ctx, rez.CreateAgentTaskRequest{
		WorkflowKind:   rez.AgentWorkflowKind(attr.WorkflowKind),
		WorkflowInput:  attr.WorkflowInput,
		TriggerKind:    attr.TriggerKind,
		TriggerPayload: attr.TriggerPayload,
	})
	if createErr != nil {
		return nil, oapi.Error(ctx, "create agent task", createErr)
	}
	resp.Body.Data = oapi.AgentTaskFromEnt(task)
	return &resp, nil
}

func (h *agentsHandler) ListAgentTasks(ctx context.Context, req *oapi.ListAgentTasksRequest) (*oapi.ListAgentTasksResponse, error) {
	var resp oapi.ListAgentTasksResponse
	tasks, listErr := h.agents.ListTasks(ctx, rez.ListAgentTasksParams{
		ListParams:   req.ListParams(),
		WorkflowKind: rez.AgentWorkflowKind(req.WorkflowKind),
		TriggerKind:  req.TriggerKind,
		SubjectType:  req.SubjectType,
		SubjectID:    req.SubjectId,
	})
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent tasks", listErr)
	}
	resp.Body.Data = make([]oapi.AgentTask, len(tasks.Data))
	for i, task := range tasks.Data {
		resp.Body.Data[i] = oapi.AgentTaskFromEnt(task)
	}
	resp.Body.Pagination.Total = tasks.Count
	return &resp, nil
}

func (h *agentsHandler) GetAgentTask(ctx context.Context, req *oapi.GetAgentTaskRequest) (*oapi.GetAgentTaskResponse, error) {
	var resp oapi.GetAgentTaskResponse
	task, getErr := h.agents.GetTask(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent task", getErr)
	}
	resp.Body.Data = oapi.AgentTaskFromEnt(task)
	return &resp, nil
}

func (h *agentsHandler) RequestAgentTaskRun(ctx context.Context, req *oapi.RequestAgentTaskRunRequest) (*oapi.RequestAgentTaskRunResponse, error) {
	var resp oapi.RequestAgentTaskRunResponse
	run, runErr := h.agents.RequestTaskRun(ctx, req.Id)
	if runErr != nil {
		return nil, oapi.Error(ctx, "request agent task run", runErr)
	}
	resp.Body.Data = oapi.AgentRunFromEnt(run)
	return &resp, nil
}

func (h *agentsHandler) ListAgentRuns(ctx context.Context, req *oapi.ListAgentRunsRequest) (*oapi.ListAgentRunsResponse, error) {
	var resp oapi.ListAgentRunsResponse
	runs, listErr := h.agents.ListRuns(ctx, rez.ListAgentRunsParams{
		ListParams:   req.ListParams(),
		AgentTaskID:  req.AgentTaskId,
		WorkflowKind: rez.AgentWorkflowKind(req.WorkflowKind),
		Status:       agentrun.Status(req.Status),
	})
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

func (h *agentsHandler) ListAgentRunCitations(ctx context.Context, req *oapi.ListAgentRunCitationsRequest) (*oapi.ListAgentRunCitationsResponse, error) {
	var resp oapi.ListAgentRunCitationsResponse
	citations, listErr := h.agents.ListRunCitations(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent run citations", listErr)
	}
	resp.Body.Data = make([]oapi.AgentRunCitation, len(citations))
	for i, citation := range citations {
		resp.Body.Data[i] = oapi.AgentRunCitationFromEnt(citation)
	}
	resp.Body.Pagination.Total = len(citations)
	return &resp, nil
}

func (h *agentsHandler) ListAgentRunFindings(ctx context.Context, req *oapi.ListAgentRunFindingsRequest) (*oapi.ListAgentRunFindingsResponse, error) {
	var resp oapi.ListAgentRunFindingsResponse
	findings, listErr := h.agents.ListRunFindings(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent run findings", listErr)
	}
	resp.Body.Data = make([]oapi.AgentRunFinding, len(findings))
	for i, finding := range findings {
		resp.Body.Data[i] = oapi.AgentRunFindingFromEnt(finding)
	}
	resp.Body.Pagination.Total = len(findings)
	return &resp, nil
}

func (h *agentsHandler) GetAgentRunResult(ctx context.Context, req *oapi.GetAgentRunResultRequest) (*oapi.GetAgentRunResultResponse, error) {
	var resp oapi.GetAgentRunResultResponse
	result, getErr := h.agents.GetRunResult(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent run result", getErr)
	}
	resp.Body.Data = oapi.AgentRunResultFromEnt(result)
	return &resp, nil
}

func (h *agentsHandler) ListAgentRunToolCalls(ctx context.Context, req *oapi.ListAgentRunToolCallsRequest) (*oapi.ListAgentRunToolCallsResponse, error) {
	var resp oapi.ListAgentRunToolCallsResponse
	toolCalls, listErr := h.agents.ListRunToolCalls(ctx, req.Id)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent run tool calls", listErr)
	}
	resp.Body.Data = make([]oapi.AgentRunToolCall, len(toolCalls))
	for i, toolCall := range toolCalls {
		resp.Body.Data[i] = oapi.AgentRunToolCallFromEnt(toolCall)
	}
	resp.Body.Pagination.Total = len(toolCalls)
	return &resp, nil
}
