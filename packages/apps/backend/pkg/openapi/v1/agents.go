package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/openapi"
)

type AgentsHandler interface {
	ListAgentRuns(context.Context, *ListAgentRunsRequest) (*ListAgentRunsResponse, error)
	RequestAgentRun(context.Context, *RequestAgentRunRequest) (*RequestAgentRunResponse, error)
	GetAgentRun(context.Context, *GetAgentRunRequest) (*GetAgentRunResponse, error)
}

func (o operations) RegisterAgents(api huma.API) {
	huma.Register(api, ListAgentRuns, o.ListAgentRuns)
	huma.Register(api, RequestAgentRun, o.RequestAgentRun)
	huma.Register(api, GetAgentRun, o.GetAgentRun)
}

type (
	AgentRun struct {
		Id         uuid.UUID          `json:"id"`
		Attributes AgentRunAttributes `json:"attributes"`
	}

	AgentRunAttributes struct {
		OwnerUserId uuid.UUID          `json:"ownerUserId"`
		Workflow    string             `json:"workflow"`
		TriggerKind string             `json:"triggerKind"`
		CreatedAt   time.Time          `json:"createdAt"`
		UpdatedAt   time.Time          `json:"updatedAt"`
		Snapshots   []AgentRunSnapshot `json:"latestSnapshot"`
	}

	AgentRunSnapshot struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunSnapshotAttributes `json:"attributes"`
	}

	AgentRunSnapshotAttributes struct {
		Status       string                 `json:"status"`
		FinishReason string                 `json:"finish_reason"`
		ParentID     *uuid.UUID             `json:"parent_id"`
		HeartbeatAt  *time.Time             `json:"heartbeat_at"`
		CreatedAt    time.Time              `json:"created_at"`
		UpdatedAt    time.Time              `json:"updated_at"`
		Error        *string                `json:"error,omitempty"`
		State        *AgentRunSnapshotState `json:"state,omitempty"`
	}

	AgentRunSnapshotState struct {
		Artifacts []AgentRunSnapshotStateArtifact `json:"artifacts"`
		Messages  []*ai.Message                   `json:"messages"`
		Custom    map[string]any                  `json:"custom"`
	}

	AgentRunSnapshotStateArtifact struct {
		Metadata map[string]any `json:"metadata,omitempty"`
		Name     string         `json:"name,omitempty"`
		Parts    []*ai.Part     `json:"parts"`
	}

	AgentRunResult struct {
		Id         uuid.UUID                `json:"id"`
		Attributes AgentRunResultAttributes `json:"attributes"`
	}

	AgentRunResultAttributes struct {
		Output       map[string]any `json:"output,omitempty"`
		ErrorMessage string         `json:"errorMessage,omitempty"`
		CreatedAt    time.Time      `json:"createdAt"`
		UpdatedAt    time.Time      `json:"updatedAt"`
		//Findings     []AgentRunFinding `json:"findings"`
	}

	AgentRunFinding struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes AgentRunFindingAttributes `json:"attributes"`
	}

	AgentRunFindingAttributes struct {
		AgentRunId  uuid.UUID                 `json:"agentRunId"`
		FindingKind string                    `json:"findingKind"`
		Content     string                    `json:"content"`
		CreatedAt   time.Time                 `json:"createdAt"`
		UpdatedAt   time.Time                 `json:"updatedAt"`
		Citations   []AgentRunFindingCitation `json:"citations"`
	}

	AgentRunFindingCitation struct {
		SupportKind string           `json:"supportKind"`
		Citation    AgentRunCitation `json:"citation"`
	}

	AgentRunCitation struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes AgentRunCitationAttributes `json:"attributes"`
	}

	AgentRunCitationAttributes struct {
		CitationKind            string         `json:"citationKind"`
		Summary                 string         `json:"summary"`
		KnowledgeEntityId       *uuid.UUID     `json:"knowledgeEntityId,omitempty"`
		KnowledgeRelationshipId *uuid.UUID     `json:"knowledgeRelationshipId,omitempty"`
		KnowledgeEvidenceId     *uuid.UUID     `json:"knowledgeEvidenceId,omitempty"`
		DomainEntityType        string         `json:"domainEntityType,omitempty"`
		DomainEntityId          *uuid.UUID     `json:"domainEntityId,omitempty"`
		DomainEntitySnapshot    map[string]any `json:"snapshot,omitempty"`
		CreatedAt               time.Time      `json:"createdAt"`
		UpdatedAt               time.Time      `json:"updatedAt"`
	}
)

func AgentRunFromEnt(run *ent.AgentRun) AgentRun {
	attrs := AgentRunAttributes{
		OwnerUserId: run.OwnerUserID,
		Workflow:    run.Workflow,
		TriggerKind: run.TriggerKind.String(),
		CreatedAt:   run.CreatedAt,
		UpdatedAt:   run.UpdatedAt,
	}
	if run.Edges.Snapshots != nil {
		attrs.Snapshots = make([]AgentRunSnapshot, len(run.Edges.Snapshots))
		for i, snapshot := range run.Edges.Snapshots {
			attrs.Snapshots[i] = AgentRunSnapshotFromEnt(snapshot)
		}
	}
	//if run.Edges.Result != nil {
	//	attrs.Result = new(AgentRunResultFromEnt(run.Edges.Result))
	//}
	return AgentRun{Id: run.ID, Attributes: attrs}
}

func AgentRunSnapshotFromEnt(s *ent.AgentRunSnapshot) AgentRunSnapshot {
	attrs := AgentRunSnapshotAttributes{
		Status:       s.Status.String(),
		FinishReason: s.FinishReason,
		ParentID:     s.ParentID,
		HeartbeatAt:  s.HeartbeatAt,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
		Error:        nil,
	}
	if s.Error != nil {
		attrs.Error = new("error")
	}
	if s.State != nil {
		var state AgentRunSnapshotState
		if jsonErr := json.Unmarshal(*s.State, &state); jsonErr != nil {
			slog.Error("failed to unmarshal state", "err", jsonErr.Error())
		} else {
			attrs.State = &state
		}
	}
	return AgentRunSnapshot{Id: s.ID, Attributes: attrs}
}

func AgentRunResultFromEnt(result *ent.AgentRunResult) AgentRunResult {
	attrs := AgentRunResultAttributes{
		Output:       nil,
		ErrorMessage: "",
		CreatedAt:    result.CreatedAt,
		UpdatedAt:    result.UpdatedAt,
	}
	if jsonErr := json.Unmarshal(result.Output, &attrs.Output); jsonErr != nil {
		slog.Error("failed to unmarshal result output", "error", jsonErr.Error())
	}
	//if len(result.Edges.Findings) > 0 {
	//	attrs.Findings = make([]AgentRunFinding, len(result.Edges.Findings))
	//	for i, finding := range result.Edges.Findings {
	//		attrs.Findings[i] = AgentRunFindingFromEnt(finding)
	//	}
	//}
	return AgentRunResult{Id: result.ID, Attributes: attrs}
}

func AgentRunFindingFromEnt(f *ent.AgentRunFinding) AgentRunFinding {
	attrs := AgentRunFindingAttributes{
		FindingKind: f.FindingKind,
		Content:     f.Content,
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
		Citations:   nil,
	}
	if len(f.Edges.FindingCitations) > 0 {
		attrs.Citations = make([]AgentRunFindingCitation, len(f.Edges.FindingCitations))
		for i, fc := range f.Edges.FindingCitations {
			attrs.Citations[i] = AgentRunFindingCitation{
				SupportKind: fc.SupportKind,
				Citation:    AgentRunCitationFromEnt(fc.Edges.Citation),
			}
		}
	}
	return AgentRunFinding{Id: f.ID, Attributes: attrs}
}

func AgentRunCitationFromEnt(c *ent.AgentRunCitation) AgentRunCitation {
	attrs := AgentRunCitationAttributes{
		CitationKind:            c.Kind,
		Summary:                 c.Summary,
		KnowledgeEntityId:       c.KnowledgeEntityID,
		KnowledgeRelationshipId: c.KnowledgeRelationshipID,
		KnowledgeEvidenceId:     c.KnowledgeEvidenceID,
		DomainEntityType:        c.DomainEntityType,
		DomainEntityId:          c.DomainEntityID,
		DomainEntitySnapshot:    c.DomainEntitySnapshot,
		CreatedAt:               c.CreatedAt,
		UpdatedAt:               c.UpdatedAt,
	}
	return AgentRunCitation{Id: c.ID, Attributes: attrs}
}

var agentsTags = []string{"Agents"}

var RequestAgentRun = openapi.Operation{
	OperationID: "request-agent-run",
	Method:      http.MethodPost,
	Path:        "/agents/runs",
	Summary:     "Request Agent Run",
	Tags:        agentsTags,
	Errors:      ErrorCodes(),
}

type RequestAgentRunRequestAttributes struct {
	Workflow string `json:"workflow"`
	Input    []byte `json:"input"`
}

type RequestAgentRunRequest RequestWithBodyAttributes[RequestAgentRunRequestAttributes]
type RequestAgentRunResponse ItemResponse[AgentRun]

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
