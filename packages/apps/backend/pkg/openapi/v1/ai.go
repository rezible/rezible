package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/openapi"
)

type AiHandler interface {
	ListAiAgents(context.Context, *ListAiAgentsRequest) (*ListAiAgentsResponse, error)
	CreateAgentSession(context.Context, *CreateAgentSessionRequest) (*CreateAgentSessionResponse, error)
	ListAgentSessions(context.Context, *ListAgentSessionsRequest) (*ListAgentSessionsResponse, error)
	GetAgentSession(context.Context, *GetAgentSessionRequest) (*GetAgentSessionResponse, error)
	ListAgentTurns(context.Context, *ListAgentTurnsRequest) (*ListAgentTurnsResponse, error)
	RequestAgentTurn(context.Context, *RequestAgentTurnRequest) (*RequestAgentTurnResponse, error)
	AbortAgentTurn(context.Context, *AgentTurnActionRequest) (*AgentTurnActionResponse, error)
	RetryAgentTurn(context.Context, *AgentTurnActionRequest) (*AgentTurnActionResponse, error)
}

func (o operations) RegisterAi(api huma.API) {
	huma.Register(api, ListAiAgents, o.ListAiAgents)
	huma.Register(api, CreateAgentSession, o.CreateAgentSession)
	huma.Register(api, ListAgentSessions, o.ListAgentSessions)
	huma.Register(api, GetAgentSession, o.GetAgentSession)
	huma.Register(api, ListAgentTurns, o.ListAgentTurns)
	huma.Register(api, RequestAgentTurn, o.RequestAgentTurn)
	huma.Register(api, AbortAgentTurn, o.AbortAgentTurn)
	huma.Register(api, RetryAgentTurn, o.RetryAgentTurn)
}

type (
	AgentSession struct {
		Id         uuid.UUID              `json:"id"`
		Attributes AgentSessionAttributes `json:"attributes"`
	}

	AgentSessionAttributes struct {
		AgentName        string     `json:"agentName"`
		OwnerUserId      uuid.UUID  `json:"ownerUserId"`
		PermissionScopes []string   `json:"permissionScopes"`
		CreatedAt        time.Time  `json:"createdAt"`
		InitialTurn      *AgentTurn `json:"initialTurn,omitempty"`
		LatestTurn       *AgentTurn `json:"latestTurn,omitempty"`
	}

	AgentTurn struct {
		Id         uuid.UUID           `json:"id"`
		Attributes AgentTurnAttributes `json:"attributes"`
	}

	AgentTurnAttributes struct {
		ParentTurnId *uuid.UUID                         `json:"parentTurnId,omitempty"`
		Status       string                             `json:"status"`
		CreatedAt    time.Time                          `json:"createdAt"`
		UpdatedAt    time.Time                          `json:"updatedAt"`
		StartedAt    *time.Time                         `json:"startedAt,omitempty"`
		FinishedAt   *time.Time                         `json:"finishedAt,omitempty"`
		FinishReason string                             `json:"finishReason,omitempty"`
		Error        *AgentTurnError                    `json:"error,omitempty"`
		State        *aix.SessionState[json.RawMessage] `json:"state,omitempty"`
	}

	AgentTurnError struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	}

	AiAgentConfig struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Model       string `json:"model"`
	}
)

func AiAgentConfigFromRez(cfg rez.AiAgentConfig) AiAgentConfig {
	return AiAgentConfig{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Model:       cfg.Model,
	}
}

func AgentSessionFromEnt(session *ent.AgentSession) AgentSession {
	attrs := AgentSessionAttributes{
		AgentName:        session.AgentName,
		OwnerUserId:      session.OwnerUserID,
		PermissionScopes: session.DefaultScopes,
		CreatedAt:        session.CreatedAt,
	}
	if numTurns := len(session.Edges.Turns); numTurns > 0 {
		attrs.InitialTurn = new(AgentTurnFromEnt(session.Edges.Turns[0]))
		if numTurns > 1 {
			attrs.LatestTurn = new(AgentTurnFromEnt(session.Edges.Turns[numTurns-1]))
		}
	}
	return AgentSession{Id: session.ID, Attributes: attrs}
}

func AgentTurnFromEnt(turn *ent.AgentTurn) AgentTurn {
	attrs := AgentTurnAttributes{
		ParentTurnId: turn.ParentID,
		Status:       turn.Status.String(),
		CreatedAt:    turn.CreatedAt,
		UpdatedAt:    turn.UpdatedAt,
		StartedAt:    turn.StartedAt,
		FinishedAt:   turn.FinishedAt,
		FinishReason: turn.FinishReason,
	}
	if turn.Error != nil {
		attrs.Error = &AgentTurnError{
			Code:    "TODO",
			Message: "error",
		}
	}
	if turn.State != nil {
		if jsonErr := json.Unmarshal(turn.State, &attrs.State); jsonErr != nil {
			slog.Error("failed to unmarshal state")
		}
	}
	return AgentTurn{Id: turn.ID, Attributes: attrs}
}

var aiTags = []string{"AI"}

var ListAiAgents = openapi.Operation{
	OperationID: "list-ai-agents",
	Method:      http.MethodGet,
	Path:        "/ai/agents",
	Summary:     "List AI Agents",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAiAgentsRequest EmptyRequest
type ListAiAgentsResponse ListResponse[AiAgentConfig]

var CreateAgentSession = openapi.Operation{
	OperationID:   "create-agent-session",
	Method:        http.MethodPost,
	Path:          "/ai/agent_sessions",
	Summary:       "Create Agent Session",
	Tags:          aiTags,
	Errors:        ErrorCodes(http.StatusConflict),
	DefaultStatus: http.StatusCreated,
}

type CreateAgentSessionAttributes struct {
	AgentName string         `json:"agentName" minLength:"1"`
	Input     map[string]any `json:"input"`
}

type CreateAgentSessionRequest RequestWithBodyAttributes[CreateAgentSessionAttributes]
type CreateAgentSessionResponse ItemResponse[AgentSession]

var ListAgentSessions = openapi.Operation{
	OperationID: "list-agent-sessions",
	Method:      http.MethodGet,
	Path:        "/ai/agent_sessions",
	Summary:     "List Agent Sessions",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAgentSessionsRequest struct {
	ListRequest
}
type ListAgentSessionsResponse ListResponse[AgentSession]

var GetAgentSession = openapi.Operation{
	OperationID: "get-agent-session",
	Method:      http.MethodGet,
	Path:        "/ai/agent_sessions/{id}",
	Summary:     "Get Agent Session",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type GetAgentSessionRequest IdRequest
type GetAgentSessionResponse ItemResponse[AgentSession]

var ListAgentTurns = openapi.Operation{
	OperationID: "list-agent-turns",
	Method:      http.MethodGet,
	Path:        "/ai/agent_sessions/{id}/turns",
	Summary:     "List Agent Turns",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAgentTurnsRequest struct {
	ListIdRequest
}
type ListAgentTurnsResponse ListResponse[AgentTurn]

type AgentTurnResume struct {
	Respond []*ai.Part `json:"respond,omitempty"`
	Restart []*ai.Part `json:"restart,omitempty"`
}

var RequestAgentTurn = openapi.Operation{
	OperationID:   "request-agent-turn",
	Method:        http.MethodPost,
	Path:          "/ai/agent_sessions/{id}/turns",
	Summary:       "Request Agent Turn",
	Tags:          aiTags,
	Errors:        ErrorCodes(http.StatusConflict),
	DefaultStatus: http.StatusAccepted,
}

type RequestAgentTurnRequestAttributes struct {
	Message *string          `json:"message,omitempty"`
	Resume  *AgentTurnResume `json:"resume,omitempty"`
}

type RequestAgentTurnRequest IdRequestWithBody[RequestAgentTurnRequestAttributes]
type RequestAgentTurnResponse ItemResponse[AgentTurn]

type AgentTurnActionRequest IdRequest
type AgentTurnActionResponse ItemResponse[AgentTurn]

var AbortAgentTurn = openapi.Operation{
	OperationID: "abort-agent-turn",
	Method:      http.MethodPost,
	Path:        "/ai/agent_turns/{id}/abort",
	Summary:     "Abort Agent Turn",
	Tags:        aiTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

var RetryAgentTurn = openapi.Operation{
	OperationID:   "retry-agent-turn",
	Method:        http.MethodPost,
	Path:          "/ai/agent_turns/{id}/retry",
	Summary:       "Retry Agent Turn",
	Tags:          aiTags,
	Errors:        ErrorCodes(http.StatusConflict),
	DefaultStatus: http.StatusAccepted,
}
