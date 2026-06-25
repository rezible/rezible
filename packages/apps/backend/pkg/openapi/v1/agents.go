package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/openapi"
)

type AgentsHandler interface {
	CreateAgentTask(context.Context, *CreateAgentTaskRequest) (*CreateAgentTaskResponse, error)
	ListAgentTasks(context.Context, *ListAgentTasksRequest) (*ListAgentTasksResponse, error)
	GetAgentTask(context.Context, *GetAgentTaskRequest) (*GetAgentTaskResponse, error)
	RequestAgentTaskRun(context.Context, *RequestAgentTaskRunRequest) (*RequestAgentTaskRunResponse, error)
	ListAgentRuns(context.Context, *ListAgentRunsRequest) (*ListAgentRunsResponse, error)
	GetAgentRun(context.Context, *GetAgentRunRequest) (*GetAgentRunResponse, error)
	ListAgentRunCitations(context.Context, *ListAgentRunCitationsRequest) (*ListAgentRunCitationsResponse, error)
	ListAgentRunFindings(context.Context, *ListAgentRunFindingsRequest) (*ListAgentRunFindingsResponse, error)
	GetAgentRunResult(context.Context, *GetAgentRunResultRequest) (*GetAgentRunResultResponse, error)
	ListAgentRunToolCalls(context.Context, *ListAgentRunToolCallsRequest) (*ListAgentRunToolCallsResponse, error)
}

func (o operations) RegisterAgents(api huma.API) {
	huma.Register(api, CreateAgentTask, o.CreateAgentTask)
	huma.Register(api, ListAgentTasks, o.ListAgentTasks)
	huma.Register(api, GetAgentTask, o.GetAgentTask)
	huma.Register(api, RequestAgentTaskRun, o.RequestAgentTaskRun)
	huma.Register(api, ListAgentRuns, o.ListAgentRuns)
	huma.Register(api, GetAgentRun, o.GetAgentRun)
	huma.Register(api, ListAgentRunCitations, o.ListAgentRunCitations)
	huma.Register(api, ListAgentRunFindings, o.ListAgentRunFindings)
	huma.Register(api, GetAgentRunResult, o.GetAgentRunResult)
	huma.Register(api, ListAgentRunToolCalls, o.ListAgentRunToolCalls)
}

type (
	AgentTask struct {
		Id         uuid.UUID           `json:"id"`
		Attributes AgentTaskAttributes `json:"attributes"`
	}

	AgentTaskAttributes struct {
		OwnerUserId     uuid.UUID      `json:"ownerUserId"`
		Workflow        string         `json:"workflow"`
		Input           map[string]any `json:"input"`
		TriggerKind     string         `json:"triggerKind"`
		TriggerMetadata map[string]any `json:"triggerMetadata"`
		CreatedAt       time.Time      `json:"createdAt"`
		UpdatedAt       time.Time      `json:"updatedAt"`
	}

	AgentRun struct {
		Id         uuid.UUID          `json:"id"`
		Attributes AgentRunAttributes `json:"attributes"`
	}

	AgentRunAttributes struct {
		AgentTaskId uuid.UUID       `json:"agentTaskId"`
		Attempt     int             `json:"attempt"`
		CreatedAt   time.Time       `json:"createdAt"`
		UpdatedAt   time.Time       `json:"updatedAt"`
		StartedAt   *time.Time      `json:"startedAt,omitempty"`
		Result      *AgentRunResult `json:"result,omitempty"`
	}

	AgentRunCitation struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunCitationAttributes `json:"attributes"`
	}

	AgentRunCitationAttributes struct {
		AgentRunId              uuid.UUID      `json:"agentRunId"`
		CitationKind            string         `json:"citationKind"`
		Summary                 string         `json:"summary"`
		KnowledgeEntityId       *uuid.UUID     `json:"knowledgeEntityId,omitempty"`
		KnowledgeRelationshipId *uuid.UUID     `json:"knowledgeRelationshipId,omitempty"`
		KnowledgeEvidenceId     *uuid.UUID     `json:"knowledgeEvidenceId,omitempty"`
		AgentTaskId             *uuid.UUID     `json:"agentTaskId,omitempty"`
		AgentRunToolCallId      *uuid.UUID     `json:"agentRunToolCallId,omitempty"`
		DomainEntityType        string         `json:"domainEntityType,omitempty"`
		DomainEntityId          *uuid.UUID     `json:"domainEntityId,omitempty"`
		DomainEntitySnapshot    map[string]any `json:"snapshot,omitempty"`
		CreatedAt               time.Time      `json:"createdAt"`
		UpdatedAt               time.Time      `json:"updatedAt"`
	}

	AgentRunFinding struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes AgentRunFindingAttributes `json:"attributes"`
	}

	AgentRunFindingAttributes struct {
		AgentRunId  uuid.UUID `json:"agentRunId"`
		Sequence    int       `json:"sequence"`
		FindingKind string    `json:"findingKind"`
		Content     string    `json:"content"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}

	AgentRunResult struct {
		Id         uuid.UUID                `json:"id"`
		Attributes AgentRunResultAttributes `json:"attributes"`
	}

	AgentRunResultAttributes struct {
		AgentRunId   uuid.UUID      `json:"agentRunId"`
		Output       map[string]any `json:"output,omitempty"`
		ErrorMessage string         `json:"errorMessage,omitempty"`
		CreatedAt    time.Time      `json:"createdAt"`
		UpdatedAt    time.Time      `json:"updatedAt"`
	}

	AgentRunToolCall struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunToolCallAttributes `json:"attributes"`
	}

	AgentRunToolCallAttributes struct {
		AgentRunId   uuid.UUID      `json:"agentRunId"`
		ToolId       string         `json:"toolId"`
		Status       string         `json:"status"`
		ToolParams   map[string]any `json:"toolParams,omitempty"`
		Result       map[string]any `json:"result,omitempty"`
		ErrorMessage string         `json:"errorMessage,omitempty"`
		StartedAt    *time.Time     `json:"startedAt,omitempty"`
		FinishedAt   *time.Time     `json:"finishedAt,omitempty"`
		CreatedAt    time.Time      `json:"createdAt"`
		UpdatedAt    time.Time      `json:"updatedAt"`
	}
)

func AgentTaskFromEnt(task *ent.AgentTask) AgentTask {
	attrs := AgentTaskAttributes{
		OwnerUserId:     task.OwnerUserID,
		Workflow:        task.Workflow,
		TriggerKind:     task.TriggerKind,
		TriggerMetadata: task.TriggerMetadata,
		CreatedAt:       task.CreatedAt,
		UpdatedAt:       task.UpdatedAt,
	}
	if jsonErr := json.Unmarshal(task.Input, &attrs.Input); jsonErr != nil {
		slog.Error("failed to unmarshal task input", "error", jsonErr.Error())
	}
	return AgentTask{Id: task.ID, Attributes: attrs}
}

func AgentRunFromEnt(run *ent.AgentRun) AgentRun {
	attrs := AgentRunAttributes{
		AgentTaskId: run.AgentTaskID,
		Attempt:     run.Attempt,
		StartedAt:   run.StartedAt,
		CreatedAt:   run.CreatedAt,
		UpdatedAt:   run.UpdatedAt,
	}
	if run.Edges.Result != nil {
		attrs.Result = new(AgentRunResultFromEnt(run.Edges.Result))
	}
	return AgentRun{Id: run.ID, Attributes: attrs}
}

func AgentRunToolCallFromEnt(c *ent.AgentRunToolCall) AgentRunToolCall {
	attrs := AgentRunToolCallAttributes{
		AgentRunId:   c.AgentRunID,
		ToolId:       c.ToolID,
		Status:       c.Status.String(),
		ToolParams:   c.ToolParams,
		Result:       c.Result,
		ErrorMessage: c.ErrorMessage,
		StartedAt:    c.StartedAt,
		FinishedAt:   c.FinishedAt,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	return AgentRunToolCall{Id: c.ID, Attributes: attrs}
}

func AgentRunCitationFromEnt(c *ent.AgentRunCitation) AgentRunCitation {
	attrs := AgentRunCitationAttributes{
		AgentRunId:              c.AgentRunID,
		CitationKind:            c.Kind,
		Summary:                 c.Summary,
		KnowledgeEntityId:       c.KnowledgeEntityID,
		KnowledgeRelationshipId: c.KnowledgeRelationshipID,
		KnowledgeEvidenceId:     c.KnowledgeEvidenceID,
		AgentTaskId:             c.AgentTaskID,
		AgentRunToolCallId:      c.AgentRunToolCallID,
		DomainEntityType:        c.DomainEntityType,
		DomainEntityId:          c.DomainEntityID,
		DomainEntitySnapshot:    c.DomainEntitySnapshot,
		CreatedAt:               c.CreatedAt,
		UpdatedAt:               c.UpdatedAt,
	}
	return AgentRunCitation{Id: c.ID, Attributes: attrs}
}

func AgentRunFindingFromEnt(f *ent.AgentRunFinding) AgentRunFinding {
	attrs := AgentRunFindingAttributes{
		AgentRunId:  f.AgentRunID,
		Sequence:    f.Sequence,
		FindingKind: f.FindingKind,
		Content:     f.Content,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
	return AgentRunFinding{Id: f.ID, Attributes: attrs}
}

func AgentRunResultFromEnt(result *ent.AgentRunResult) AgentRunResult {
	attrs := AgentRunResultAttributes{
		AgentRunId:   result.AgentRunID,
		ErrorMessage: result.ErrorMessage,
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
	}
	if jsonErr := json.Unmarshal(result.Output, &attrs.Output); jsonErr != nil {
		slog.Error("failed to unmarshal result output", "error", jsonErr.Error())
	}
	return AgentRunResult{Id: result.ID, Attributes: attrs}
}

var agentsTags = []string{"Agents"}

var CreateAgentTask = openapi.Operation{
	OperationID: "create-agent-task",
	Method:      http.MethodPost,
	Path:        "/agents/tasks",
	Summary:     "Create Agent Task",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type CreateAgentTaskAttributes struct {
	Workflow       string         `json:"workflow"`
	Input          map[string]any `json:"input"`
	TriggerKind    string         `json:"triggerKind,omitempty"`
	TriggerPayload map[string]any `json:"triggerPayload,omitempty"`
}
type CreateAgentTaskRequest RequestWithBodyAttributes[CreateAgentTaskAttributes]
type CreateAgentTaskResponse ItemResponse[AgentTask]

var ListAgentTasks = openapi.Operation{
	OperationID: "list-agent-tasks",
	Method:      http.MethodGet,
	Path:        "/agents/tasks",
	Summary:     "List Agent Tasks",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentTasksRequest struct {
	ListRequest
	Workflow       string    `query:"workflow" required:"false"`
	TriggerKind    string    `query:"triggerKind" required:"false"`
	SubjectKind    string    `query:"subjectKind" required:"false"`
	DomainEntityId uuid.UUID `query:"domainEntityId" required:"false"`
}
type ListAgentTasksResponse ListResponse[AgentTask]

var GetAgentTask = openapi.Operation{
	OperationID: "get-agent-task",
	Method:      http.MethodGet,
	Path:        "/agents/tasks/{id}",
	Summary:     "Get Agent Task",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type GetAgentTaskRequest EmptyIdRequest
type GetAgentTaskResponse ItemResponse[AgentTask]

var RequestAgentTaskRun = openapi.Operation{
	OperationID: "request-agent-task-run",
	Method:      http.MethodPost,
	Path:        "/agents/tasks/{id}/runs",
	Summary:     "Request Agent Task Run",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type RequestAgentTaskRunRequest EmptyIdRequest
type RequestAgentTaskRunResponse ItemResponse[AgentRun]

var ListAgentRuns = openapi.Operation{
	OperationID: "list-agent-runs",
	Method:      http.MethodGet,
	Path:        "/agents/runs",
	Summary:     "List Agent Runs",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentRunsRequest struct {
	ListRequest
	AgentTaskId uuid.UUID           `query:"agentTaskId" required:"false"`
	Workflow    string              `query:"workflow" required:"false"`
	Started     OptionalParam[bool] `query:"status" required:"false"`
	Resulted    OptionalParam[bool] `query:"resulted" required:"false"`
}
type ListAgentRunsResponse ListResponse[AgentRun]

var GetAgentRun = openapi.Operation{
	OperationID: "get-agent-run",
	Method:      http.MethodGet,
	Path:        "/agents/runs/{id}",
	Summary:     "Get Agent Run",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type GetAgentRunRequest EmptyIdRequest
type GetAgentRunResponse ItemResponse[AgentRun]

var ListAgentRunCitations = openapi.Operation{
	OperationID: "list-agent-run-citations",
	Method:      http.MethodGet,
	Path:        "/agents/runs/{id}/citations",
	Summary:     "List Agent Run Citations",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentRunCitationsRequest EmptyIdRequest
type ListAgentRunCitationsResponse ListResponse[AgentRunCitation]

var ListAgentRunFindings = openapi.Operation{
	OperationID: "list-agent-run-findings",
	Method:      http.MethodGet,
	Path:        "/agents/runs/{id}/findings",
	Summary:     "List Agent Run Findings",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentRunFindingsRequest EmptyIdRequest
type ListAgentRunFindingsResponse ListResponse[AgentRunFinding]

var GetAgentRunResult = openapi.Operation{
	OperationID: "get-agent-run-result",
	Method:      http.MethodGet,
	Path:        "/agents/runs/{id}/result",
	Summary:     "Get Agent Run Result",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type GetAgentRunResultRequest EmptyIdRequest
type GetAgentRunResultResponse ItemResponse[AgentRunResult]

var ListAgentRunToolCalls = openapi.Operation{
	OperationID: "list-agent-run-tool-calls",
	Method:      http.MethodGet,
	Path:        "/agents/runs/{id}/tool-calls",
	Summary:     "List Agent Run Tool Calls",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentRunToolCallsRequest EmptyIdRequest
type ListAgentRunToolCallsResponse ListResponse[AgentRunToolCall]
