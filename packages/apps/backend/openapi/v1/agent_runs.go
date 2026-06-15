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

type AgentRunsHandler interface {
	ListAgentRuns(context.Context, *ListAgentRunsRequest) (*ListAgentRunsResponse, error)
	GetAgentRun(context.Context, *GetAgentRunRequest) (*GetAgentRunResponse, error)
	ListAgentRunArtifacts(context.Context, *ListAgentRunArtifactsRequest) (*ListAgentRunArtifactsResponse, error)
}

func (o operations) RegisterAgentRuns(api huma.API) {
	huma.Register(api, ListAgentRuns, o.ListAgentRuns)
	huma.Register(api, GetAgentRun, o.GetAgentRun)
	huma.Register(api, ListAgentRunArtifacts, o.ListAgentRunArtifacts)
}

type (
	AgentRun struct {
		Id         uuid.UUID          `json:"id"`
		Attributes AgentRunAttributes `json:"attributes"`
	}

	AgentRunAttributes struct {
		WorkflowKind    string         `json:"workflowKind"`
		Status          string         `json:"status"`
		IdempotencyKey  string         `json:"idempotencyKey"`
		SubjectKind     string         `json:"subjectKind,omitempty"`
		SubjectId       *uuid.UUID     `json:"subjectId,omitempty"`
		TriggerMetadata map[string]any `json:"triggerMetadata,omitempty"`
		ModelMetadata   map[string]any `json:"modelMetadata,omitempty"`
		ErrorCode       string         `json:"errorCode,omitempty"`
		ErrorMessage    string         `json:"errorMessage,omitempty"`
		QueuedAt        time.Time      `json:"queuedAt"`
		StartedAt       *time.Time     `json:"startedAt,omitempty"`
		CompletedAt     *time.Time     `json:"completedAt,omitempty"`
		FailedAt        *time.Time     `json:"failedAt,omitempty"`
	}

	AgentRunArtifact struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunArtifactAttributes `json:"attributes"`
	}

	AgentRunArtifactAttributes struct {
		Kind      string         `json:"kind"`
		Name      string         `json:"name"`
		Payload   map[string]any `json:"payload"`
		Redacted  bool           `json:"redacted"`
		CreatedAt time.Time      `json:"createdAt"`
	}
)

func AgentRunFromEnt(run *ent.AgentRun) AgentRun {
	return AgentRun{
		Id: run.ID,
		Attributes: AgentRunAttributes{
			WorkflowKind:    run.WorkflowKind.String(),
			Status:          run.Status.String(),
			IdempotencyKey:  run.IdempotencyKey,
			SubjectKind:     run.SubjectKind,
			SubjectId:       run.SubjectID,
			TriggerMetadata: run.TriggerMetadata,
			ModelMetadata:   run.ModelMetadata,
			ErrorCode:       run.ErrorCode,
			ErrorMessage:    run.ErrorMessage,
			QueuedAt:        run.QueuedAt,
			StartedAt:       run.StartedAt,
			CompletedAt:     run.CompletedAt,
			FailedAt:        run.FailedAt,
		},
	}
}

func AgentRunArtifactFromEnt(artifact *ent.AgentRunArtifact) AgentRunArtifact {
	return AgentRunArtifact{
		Id: artifact.ID,
		Attributes: AgentRunArtifactAttributes{
			Kind:      artifact.Kind.String(),
			Name:      artifact.Name,
			Payload:   artifact.Payload,
			Redacted:  artifact.Redacted,
			CreatedAt: artifact.CreatedAt,
		},
	}
}

var agentRunsTags = []string{"Agent Runs"}

var ListAgentRuns = openapi.Operation{
	OperationID: "list-agent-runs",
	Method:      http.MethodGet,
	Path:        "/agent-runs",
	Summary:     "List Agent Runs",
	Tags:        agentRunsTags,
	Errors:      ErrorCodes(),
}

type ListAgentRunsRequest struct {
	ListRequest
	WorkflowKind string    `query:"workflowKind" required:"false"`
	Status       string    `query:"status" required:"false"`
	SubjectKind  string    `query:"subjectKind" required:"false"`
	SubjectId    uuid.UUID `query:"subjectId" required:"false"`
}
type ListAgentRunsResponse ListResponse[AgentRun]

var GetAgentRun = openapi.Operation{
	OperationID: "get-agent-run",
	Method:      http.MethodGet,
	Path:        "/agent-runs/{id}",
	Summary:     "Get Agent Run",
	Tags:        agentRunsTags,
	Errors:      ErrorCodes(),
}

type GetAgentRunRequest EmptyIdRequest
type GetAgentRunResponse ItemResponse[AgentRun]

var ListAgentRunArtifacts = openapi.Operation{
	OperationID: "list-agent-run-artifacts",
	Method:      http.MethodGet,
	Path:        "/agent-runs/{id}/artifacts",
	Summary:     "List Agent Run Artifacts",
	Tags:        agentRunsTags,
	Errors:      ErrorCodes(),
}

type ListAgentRunArtifactsRequest EmptyIdRequest
type ListAgentRunArtifactsResponse ListResponse[AgentRunArtifact]
