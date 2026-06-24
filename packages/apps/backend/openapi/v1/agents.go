package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/openapi"
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
		OwnerUserId    uuid.UUID      `json:"ownerUserId"`
		WorkflowKind   string         `json:"workflowKind"`
		WorkflowInput  map[string]any `json:"workflowInput"`
		TriggerKind    string         `json:"triggerKind"`
		TriggerPayload map[string]any `json:"triggerPayload"`
		CreatedAt      time.Time      `json:"createdAt"`
		UpdatedAt      time.Time      `json:"updatedAt"`
	}

	AgentRun struct {
		Id         uuid.UUID          `json:"id"`
		Attributes AgentRunAttributes `json:"attributes"`
	}

	AgentRunAttributes struct {
		AgentTaskId  uuid.UUID  `json:"agentTaskId"`
		WorkflowKind string     `json:"workflowKind"`
		Attempt      int        `json:"attempt"`
		Status       string     `json:"status"`
		ErrorMessage string     `json:"errorMessage,omitempty"`
		StartedAt    *time.Time `json:"startedAt,omitempty"`
		FinishedAt   *time.Time `json:"finishedAt,omitempty"`
		CreatedAt    time.Time  `json:"createdAt"`
		UpdatedAt    time.Time  `json:"updatedAt"`
	}

	AgentRunCitation struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunCitationAttributes `json:"attributes"`
	}

	AgentRunCitationAttributes struct {
		AgentRunId              uuid.UUID      `json:"agentRunId"`
		CitationKind            string         `json:"citationKind"`
		DomainEntityType        string         `json:"domainEntityType,omitempty"`
		DomainEntityId          *uuid.UUID     `json:"domainEntityId,omitempty"`
		KnowledgeEntityId       *uuid.UUID     `json:"knowledgeEntityId,omitempty"`
		KnowledgeRelationshipId *uuid.UUID     `json:"knowledgeRelationshipId,omitempty"`
		KnowledgeEvidenceId     *uuid.UUID     `json:"knowledgeEvidenceId,omitempty"`
		AgentTaskId             *uuid.UUID     `json:"agentTaskId,omitempty"`
		AgentRunToolCallId      *uuid.UUID     `json:"agentRunToolCallId,omitempty"`
		Summary                 string         `json:"summary"`
		Snapshot                map[string]any `json:"snapshot,omitempty"`
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
		AgentRunId uuid.UUID      `json:"agentRunId"`
		Content    string         `json:"content"`
		Data       map[string]any `json:"data,omitempty"`
		CreatedAt  time.Time      `json:"createdAt"`
		UpdatedAt  time.Time      `json:"updatedAt"`
	}

	AgentRunToolCall struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunToolCallAttributes `json:"attributes"`
	}

	AgentRunToolCallAttributes struct {
		AgentRunId   uuid.UUID      `json:"agentRunId"`
		ToolName     string         `json:"toolName"`
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
	return AgentTask{
		Id: task.ID,
		Attributes: AgentTaskAttributes{
			OwnerUserId:    task.OwnerUserID,
			WorkflowKind:   task.WorkflowKind,
			WorkflowInput:  task.WorkflowInput,
			TriggerKind:    task.TriggerKind,
			TriggerPayload: task.TriggerPayload,
			CreatedAt:      task.CreatedAt,
			UpdatedAt:      task.UpdatedAt,
		},
	}
}

func AgentRunFromEnt(run *ent.AgentRun) AgentRun {
	workflowKind := ""
	if run.Edges.AgentTask != nil {
		workflowKind = run.Edges.AgentTask.WorkflowKind
	}
	return AgentRun{
		Id: run.ID,
		Attributes: AgentRunAttributes{
			AgentTaskId:  run.AgentTaskID,
			WorkflowKind: workflowKind,
			Attempt:      run.Attempt,
			Status:       run.Status.String(),
			ErrorMessage: run.ErrorMessage,
			StartedAt:    run.StartedAt,
			FinishedAt:   run.FinishedAt,
			CreatedAt:    run.CreatedAt,
			UpdatedAt:    run.UpdatedAt,
		},
	}
}

func AgentRunCitationFromEnt(citation *ent.AgentRunCitation) AgentRunCitation {
	return AgentRunCitation{
		Id: citation.ID,
		Attributes: AgentRunCitationAttributes{
			AgentRunId:              citation.AgentRunID,
			CitationKind:            citation.CitationKind,
			DomainEntityType:        citation.DomainEntityType,
			DomainEntityId:          citation.DomainEntityID,
			KnowledgeEntityId:       citation.KnowledgeEntityID,
			KnowledgeRelationshipId: citation.KnowledgeRelationshipID,
			KnowledgeEvidenceId:     citation.KnowledgeEvidenceID,
			AgentTaskId:             citation.AgentTaskID,
			AgentRunToolCallId:      citation.AgentRunToolCallID,
			Summary:                 citation.Summary,
			Snapshot:                citation.Snapshot,
			CreatedAt:               citation.CreatedAt,
			UpdatedAt:               citation.UpdatedAt,
		},
	}
}

func AgentRunFindingFromEnt(finding *ent.AgentRunFinding) AgentRunFinding {
	return AgentRunFinding{
		Id: finding.ID,
		Attributes: AgentRunFindingAttributes{
			AgentRunId:  finding.AgentRunID,
			Sequence:    finding.Sequence,
			FindingKind: finding.FindingKind,
			Content:     finding.Content,
			CreatedAt:   finding.CreatedAt,
			UpdatedAt:   finding.UpdatedAt,
		},
	}
}

func AgentRunResultFromEnt(result *ent.AgentRunResult) AgentRunResult {
	return AgentRunResult{
		Id: result.ID,
		Attributes: AgentRunResultAttributes{
			AgentRunId: result.AgentRunID,
			Content:    result.Content,
			Data:       result.Data,
			CreatedAt:  result.CreatedAt,
			UpdatedAt:  result.UpdatedAt,
		},
	}
}

func AgentRunToolCallFromEnt(toolCall *ent.AgentRunToolCall) AgentRunToolCall {
	return AgentRunToolCall{
		Id: toolCall.ID,
		Attributes: AgentRunToolCallAttributes{
			AgentRunId:   toolCall.AgentRunID,
			ToolName:     toolCall.ToolName,
			Status:       toolCall.Status.String(),
			ToolParams:   toolCall.ToolParams,
			Result:       toolCall.Result,
			ErrorMessage: toolCall.ErrorMessage,
			StartedAt:    toolCall.StartedAt,
			FinishedAt:   toolCall.FinishedAt,
			CreatedAt:    toolCall.CreatedAt,
			UpdatedAt:    toolCall.UpdatedAt,
		},
	}
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
	WorkflowKind   string         `json:"workflowKind"`
	WorkflowInput  map[string]any `json:"workflowInput"`
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
	WorkflowKind string    `query:"workflowKind" required:"false"`
	TriggerKind  string    `query:"triggerKind" required:"false"`
	SubjectType  string    `query:"subjectType" required:"false"`
	SubjectId    uuid.UUID `query:"subjectId" required:"false"`
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

type RequestAgentTaskRunRequest struct {
	EmptyIdRequest
}
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
	AgentTaskId  uuid.UUID `query:"agentTaskId" required:"false"`
	WorkflowKind string    `query:"workflowKind" required:"false"`
	Status       string    `query:"status" required:"false"`
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
