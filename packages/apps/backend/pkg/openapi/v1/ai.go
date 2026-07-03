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

type AiHandler interface {
	ListAiAgentRuns(context.Context, *ListAiAgentRunsRequest) (*ListAiAgentRunsResponse, error)
	RequestAiAgentRun(context.Context, *RequestAiAgentRunRequest) (*RequestAiAgentRunResponse, error)
	GetAiAgentRun(context.Context, *GetAiAgentRunRequest) (*GetAiAgentRunResponse, error)
}

func (o operations) RegisterAi(api huma.API) {
	huma.Register(api, ListAiAgentRuns, o.ListAiAgentRuns)
	huma.Register(api, RequestAiAgentRun, o.RequestAiAgentRun)
	huma.Register(api, GetAiAgentRun, o.GetAiAgentRun)
}

type (
	AiAgentRun struct {
		Id         uuid.UUID            `json:"id"`
		Attributes AiAgentRunAttributes `json:"attributes"`
	}

	AiAgentRunAttributes struct {
		AgentName        string               `json:"agentname"`
		OwnerUserId      uuid.UUID            `json:"ownerUserId"`
		PermissionScopes []string             `json:"permissionScopes"`
		CreatedAt        time.Time            `json:"createdAt"`
		StartedAt        *time.Time           `json:"startedAt,omitempty"`
		Snapshots        []AiAgentRunSnapshot `json:"latestSnapshot"`
	}

	AiAgentRunSnapshot struct {
		Id         uuid.UUID                    `json:"id"`
		Attributes AiAgentRunSnapshotAttributes `json:"attributes"`
	}

	AiAgentRunSnapshotAttributes struct {
		Status       string                   `json:"status"`
		FinishReason string                   `json:"finish_reason"`
		ParentID     *uuid.UUID               `json:"parent_id"`
		HeartbeatAt  *time.Time               `json:"heartbeat_at"`
		CreatedAt    time.Time                `json:"created_at"`
		Error        *string                  `json:"error,omitempty"`
		State        *AiAgentRunSnapshotState `json:"state,omitempty"`
	}

	AiAgentRunSnapshotState struct {
		Artifacts []AiAgentRunSnapshotStateArtifact `json:"artifacts"`
		Messages  []*ai.Message                     `json:"messages"`
		Custom    map[string]any                    `json:"custom"`
	}

	AiAgentRunSnapshotStateArtifact struct {
		Name     string         `json:"name,omitempty"`
		Parts    []*ai.Part     `json:"parts"`
		Metadata map[string]any `json:"metadata,omitempty"`
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

func AiAgentRunFromEnt(run *ent.AiAgentRun) AiAgentRun {
	attrs := AiAgentRunAttributes{
		AgentName:        run.AgentName,
		OwnerUserId:      run.OwnerUserID,
		PermissionScopes: run.Scopes,
		CreatedAt:        run.CreatedAt,
		StartedAt:        run.StartedAt,
		Snapshots:        make([]AiAgentRunSnapshot, len(run.Edges.Snapshots)),
	}
	if run.Edges.Snapshots != nil {
		for i, snapshot := range run.Edges.Snapshots {
			attrs.Snapshots[i] = AiAgentRunSnapshotFromEnt(snapshot)
		}
	}
	return AiAgentRun{Id: run.ID, Attributes: attrs}
}

func AiAgentRunSnapshotFromEnt(s *ent.AiAgentRunSnapshot) AiAgentRunSnapshot {
	attrs := AiAgentRunSnapshotAttributes{
		Status:       s.Status.String(),
		FinishReason: s.FinishReason,
		ParentID:     s.ParentID,
		HeartbeatAt:  s.HeartbeatAt,
		CreatedAt:    s.CreatedAt,
		Error:        nil,
	}
	if s.Error != nil {
		attrs.Error = new("error")
	}
	if s.State != nil {
		var state AiAgentRunSnapshotState
		if jsonErr := json.Unmarshal(*s.State, &state); jsonErr != nil {
			slog.Error("failed to unmarshal state", "err", jsonErr.Error())
		} else {
			attrs.State = &state
		}
	}
	return AiAgentRunSnapshot{Id: s.ID, Attributes: attrs}
}

func AiAgentRunFindingFromEnt(f *ent.AiAgentRunFinding) AgentRunFinding {
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
				Citation:    AiAgentRunCitationFromEnt(fc.Edges.Citation),
			}
		}
	}
	return AgentRunFinding{Id: f.ID, Attributes: attrs}
}

func AiAgentRunCitationFromEnt(c *ent.AiAgentRunCitation) AgentRunCitation {
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

var aiTags = []string{"AI"}

var RequestAiAgentRun = openapi.Operation{
	OperationID: "request-ai-agent-run",
	Method:      http.MethodPost,
	Path:        "/ai/agents/runs",
	Summary:     "Request Agent Run",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type RequestAgentRunRequestAttributes struct {
	Workflow string         `json:"workflow"`
	Input    map[string]any `json:"input"`
}

type RequestAiAgentRunRequest RequestWithBodyAttributes[RequestAgentRunRequestAttributes]
type RequestAiAgentRunResponse ItemResponse[AiAgentRun]

var ListAiAgentRuns = openapi.Operation{
	OperationID: "list-ai-agent-runs",
	Method:      http.MethodGet,
	Path:        "/ai/agents/runs",
	Summary:     "List Agent Runs",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAiAgentRunsRequest struct {
	ListRequest
	AgentTaskId uuid.UUID           `query:"agentTaskId" required:"false"`
	Name        string              `query:"name" required:"false"`
	Resulted    OptionalParam[bool] `query:"resulted" required:"false"`
}
type ListAiAgentRunsResponse ListResponse[AiAgentRun]

var GetAiAgentRun = openapi.Operation{
	OperationID: "get-ai-agent-run",
	Method:      http.MethodGet,
	Path:        "/ai/agents/runs/{id}",
	Summary:     "Get an Agent Run",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type GetAiAgentRunRequest EmptyIdRequest
type GetAiAgentRunResponse ItemResponse[AiAgentRun]
