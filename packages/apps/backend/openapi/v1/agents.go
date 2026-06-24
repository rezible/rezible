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
	CreateAgentCase(context.Context, *CreateAgentCaseRequest) (*CreateAgentCaseResponse, error)
	ListAgentCases(context.Context, *ListAgentCasesRequest) (*ListAgentCasesResponse, error)
	GetAgentCase(context.Context, *GetAgentCaseRequest) (*GetAgentCaseResponse, error)
	ListAgentCaseSteps(context.Context, *ListAgentCaseStepsRequest) (*ListAgentCaseStepsResponse, error)
	ListAgentCaseArtifacts(context.Context, *ListAgentCaseArtifactsRequest) (*ListAgentCaseArtifactsResponse, error)
	ListAgentCaseConclusions(context.Context, *ListAgentCaseConclusionsRequest) (*ListAgentCaseConclusionsResponse, error)

	RequestAgentCaseRun(context.Context, *RequestAgentCaseRunRequest) (*RequestAgentCaseRunResponse, error)
	ListAgentRuns(context.Context, *ListAgentRunsRequest) (*ListAgentRunsResponse, error)
	GetAgentRun(context.Context, *GetAgentRunRequest) (*GetAgentRunResponse, error)
}

func (o operations) RegisterAgents(api huma.API) {
	huma.Register(api, CreateAgentCase, o.CreateAgentCase)
	huma.Register(api, ListAgentCases, o.ListAgentCases)
	huma.Register(api, GetAgentCase, o.GetAgentCase)
	huma.Register(api, ListAgentCaseSteps, o.ListAgentCaseSteps)
	huma.Register(api, ListAgentCaseArtifacts, o.ListAgentCaseArtifacts)
	huma.Register(api, ListAgentCaseConclusions, o.ListAgentCaseConclusions)
	huma.Register(api, RequestAgentCaseRun, o.RequestAgentCaseRun)
	huma.Register(api, ListAgentRuns, o.ListAgentRuns)
	huma.Register(api, GetAgentRun, o.GetAgentRun)
}

type (
	AgentCase struct {
		Id         uuid.UUID           `json:"id"`
		Attributes AgentCaseAttributes `json:"attributes"`
	}

	AgentCaseAttributes struct {
		Status          string         `json:"status"`
		Title           string         `json:"title"`
		Query           string         `json:"query,omitempty"`
		WorkflowKind    string         `json:"workflowKind,omitempty"`
		SubjectKind     string         `json:"subjectKind,omitempty"`
		SubjectId       *uuid.UUID     `json:"subjectId,omitempty"`
		TriggerMetadata map[string]any `json:"triggerMetadata,omitempty"`
		Summary         string         `json:"summary,omitempty"`
		ErrorCode       string         `json:"errorCode,omitempty"`
		ErrorMessage    string         `json:"errorMessage,omitempty"`
		CreatedAt       time.Time      `json:"createdAt"`
		UpdatedAt       time.Time      `json:"updatedAt"`
	}

	AgentCaseStep struct {
		Id         uuid.UUID               `json:"id"`
		Attributes AgentCaseStepAttributes `json:"attributes"`
	}

	AgentCaseStepAttributes struct {
		AgentCaseId uuid.UUID      `json:"agentCaseId"`
		AgentRunId  *uuid.UUID     `json:"agentRunId,omitempty"`
		Sequence    int            `json:"sequence"`
		Kind        string         `json:"kind"`
		Title       string         `json:"title"`
		Summary     string         `json:"summary,omitempty"`
		Input       map[string]any `json:"input,omitempty"`
		Output      map[string]any `json:"output,omitempty"`
		StartedAt   *time.Time     `json:"startedAt,omitempty"`
		CompletedAt *time.Time     `json:"completedAt,omitempty"`
		CreatedAt   time.Time      `json:"createdAt"`
	}

	AgentCaseArtifact struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes AgentCaseArtifactAttributes `json:"attributes"`
	}

	AgentCaseArtifactAttributes struct {
		AgentCaseId     uuid.UUID      `json:"agentCaseId"`
		AgentCaseStepId *uuid.UUID     `json:"agentCaseStepId,omitempty"`
		AgentRunId      *uuid.UUID     `json:"agentRunId,omitempty"`
		Kind            string         `json:"kind"`
		Role            string         `json:"role,omitempty"`
		Name            string         `json:"name"`
		Payload         map[string]any `json:"payload"`
		Redacted        bool           `json:"redacted"`
		CreatedAt       time.Time      `json:"createdAt"`
	}

	AgentCaseConclusion struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes AgentCaseConclusionAttributes `json:"attributes"`
	}

	AgentCaseConclusionAttributes struct {
		AgentCaseId        uuid.UUID      `json:"agentCaseId"`
		AgentCaseStepId    *uuid.UUID     `json:"agentCaseStepId,omitempty"`
		AgentRunId         *uuid.UUID     `json:"agentRunId,omitempty"`
		Kind               string         `json:"kind"`
		Summary            string         `json:"summary,omitempty"`
		Confidence         string         `json:"confidence,omitempty"`
		RecommendedActions []string       `json:"recommendedActions,omitempty"`
		Limitations        []string       `json:"limitations,omitempty"`
		Payload            map[string]any `json:"payload"`
		CreatedAt          time.Time      `json:"createdAt"`
	}

	AgentRun struct {
		Id         uuid.UUID          `json:"id"`
		Attributes AgentRunAttributes `json:"attributes"`
	}

	AgentRunAttributes struct {
		AgentCaseId     *uuid.UUID     `json:"agentCaseId,omitempty"`
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
)

func AgentCaseFromEnt(c *ent.AgentCase) AgentCase {
	return AgentCase{
		Id: c.ID,
		Attributes: AgentCaseAttributes{
			Status:          c.Status.String(),
			Title:           c.Title,
			Query:           c.Query,
			WorkflowKind:    c.WorkflowKind.String(),
			SubjectKind:     c.SubjectKind,
			SubjectId:       c.SubjectID,
			TriggerMetadata: c.TriggerMetadata,
			Summary:         c.Summary,
			ErrorCode:       c.ErrorCode,
			ErrorMessage:    c.ErrorMessage,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
		},
	}
}

func AgentCaseStepFromEnt(step *ent.AgentCaseStep) AgentCaseStep {
	return AgentCaseStep{
		Id: step.ID,
		Attributes: AgentCaseStepAttributes{
			AgentCaseId: step.AgentCaseID,
			AgentRunId:  step.AgentRunID,
			Sequence:    step.Sequence,
			Kind:        step.Kind.String(),
			Title:       step.Title,
			Summary:     step.Summary,
			Input:       step.Input,
			Output:      step.Output,
			StartedAt:   step.StartedAt,
			CompletedAt: step.CompletedAt,
			CreatedAt:   step.CreatedAt,
		},
	}
}

func AgentCaseArtifactFromEnt(artifact *ent.AgentCaseArtifact) AgentCaseArtifact {
	return AgentCaseArtifact{
		Id: artifact.ID,
		Attributes: AgentCaseArtifactAttributes{
			AgentCaseId:     artifact.AgentCaseID,
			AgentCaseStepId: artifact.AgentCaseStepID,
			AgentRunId:      artifact.AgentRunID,
			Kind:            artifact.Kind.String(),
			Role:            artifact.Role,
			Name:            artifact.Name,
			Payload:         artifact.Payload,
			Redacted:        artifact.Redacted,
			CreatedAt:       artifact.CreatedAt,
		},
	}
}

func AgentCaseConclusionFromEnt(conclusion *ent.AgentCaseConclusion) AgentCaseConclusion {
	return AgentCaseConclusion{
		Id: conclusion.ID,
		Attributes: AgentCaseConclusionAttributes{
			AgentCaseId:        conclusion.AgentCaseID,
			AgentCaseStepId:    conclusion.AgentCaseStepID,
			AgentRunId:         conclusion.AgentRunID,
			Kind:               conclusion.Kind,
			Summary:            conclusion.Summary,
			Confidence:         conclusion.Confidence,
			RecommendedActions: conclusion.RecommendedActions,
			Limitations:        conclusion.Limitations,
			Payload:            conclusion.Payload,
			CreatedAt:          conclusion.CreatedAt,
		},
	}
}

func AgentRunFromEnt(run *ent.AgentRun) AgentRun {
	return AgentRun{
		Id: run.ID,
		Attributes: AgentRunAttributes{
			AgentCaseId:     run.AgentCaseID,
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

var agentsTags = []string{"Agents"}

var CreateAgentCase = openapi.Operation{
	OperationID: "create-agent-case",
	Method:      http.MethodPost,
	Path:        "/agents/cases",
	Summary:     "Create Agent Case",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type CreateAgentCaseAttributes struct {
	Title           string         `json:"title,omitempty"`
	Query           string         `json:"query,omitempty"`
	WorkflowKind    string         `json:"workflowKind"`
	SubjectKind     string         `json:"subjectKind,omitempty"`
	SubjectId       uuid.UUID      `json:"subjectId,omitempty"`
	TriggerMetadata map[string]any `json:"triggerMetadata,omitempty"`
}
type CreateAgentCaseRequest RequestWithBodyAttributes[CreateAgentCaseAttributes]
type CreateAgentCaseResponse ItemResponse[AgentCase]

var ListAgentCases = openapi.Operation{
	OperationID: "list-agent-cases",
	Method:      http.MethodGet,
	Path:        "/agents/cases",
	Summary:     "List Agent Cases",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentCasesRequest struct {
	ListRequest
	Status       string    `query:"status" required:"false"`
	WorkflowKind string    `query:"workflowKind" required:"false"`
	SubjectKind  string    `query:"subjectKind" required:"false"`
	SubjectId    uuid.UUID `query:"subjectId" required:"false"`
}
type ListAgentCasesResponse ListResponse[AgentCase]

var GetAgentCase = openapi.Operation{
	OperationID: "get-agent-case",
	Method:      http.MethodGet,
	Path:        "/agents/cases/{id}",
	Summary:     "Get Agent Case",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type GetAgentCaseRequest EmptyIdRequest
type GetAgentCaseResponse ItemResponse[AgentCase]

var ListAgentCaseSteps = openapi.Operation{
	OperationID: "list-agent-case-steps",
	Method:      http.MethodGet,
	Path:        "/agents/cases/{id}/steps",
	Summary:     "List Agent Case Steps",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentCaseStepsRequest EmptyIdRequest
type ListAgentCaseStepsResponse ListResponse[AgentCaseStep]

var ListAgentCaseArtifacts = openapi.Operation{
	OperationID: "list-agent-case-artifacts",
	Method:      http.MethodGet,
	Path:        "/agents/cases/{id}/artifacts",
	Summary:     "List Agent Case Artifacts",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentCaseArtifactsRequest EmptyIdRequest
type ListAgentCaseArtifactsResponse ListResponse[AgentCaseArtifact]

var ListAgentCaseConclusions = openapi.Operation{
	OperationID: "list-agent-case-conclusions",
	Method:      http.MethodGet,
	Path:        "/agents/cases/{id}/conclusions",
	Summary:     "List Agent Case Conclusions",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type ListAgentCaseConclusionsRequest EmptyIdRequest
type ListAgentCaseConclusionsResponse ListResponse[AgentCaseConclusion]

var RequestAgentCaseRun = openapi.Operation{
	OperationID: "request-agent-case-run",
	Method:      http.MethodPost,
	Path:        "/agents/cases/{id}/runs",
	Summary:     "Request Agent Case Run",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type RequestAgentCaseRunAttributes struct {
	IdempotencyKey string         `json:"idempotencyKey,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}
type RequestAgentCaseRunRequest struct {
	EmptyIdRequest
	RequestWithBodyAttributes[RequestAgentCaseRunAttributes]
}
type RequestAgentCaseRunResponse ItemResponse[AgentRun]

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
	WorkflowKind string    `query:"workflowKind" required:"false"`
	Status       string    `query:"status" required:"false"`
	AgentCaseId  uuid.UUID `query:"agentCaseId" required:"false"`
	SubjectKind  string    `query:"subjectKind" required:"false"`
	SubjectId    uuid.UUID `query:"subjectId" required:"false"`
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
