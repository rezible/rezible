package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/openapi"
)

type AiHandler interface {
	ListAiAgents(context.Context, *ListAiAgentsRequest) (*ListAiAgentsResponse, error)

	CreateAgentSession(context.Context, *CreateAgentSessionRequest) (*CreateAgentSessionResponse, error)
	ListAgentSessions(context.Context, *ListAgentSessionsRequest) (*ListAgentSessionsResponse, error)
	GetAgentSession(context.Context, *GetAgentSessionRequest) (*GetAgentSessionResponse, error)
	ListAgentMessages(context.Context, *ListAgentMessagesRequest) (*ListAgentMessagesResponse, error)
	ListAgentArtifacts(context.Context, *ListAgentArtifactsRequest) (*ListAgentArtifactsResponse, error)

	ListAgentTurns(context.Context, *ListAgentTurnsRequest) (*ListAgentTurnsResponse, error)
	RequestAgentTurn(context.Context, *RequestAgentTurnRequest) (*RequestAgentTurnResponse, error)
	AbortAgentTurn(context.Context, *AgentTurnActionRequest) (*AgentTurnActionResponse, error)
	RetryAgentTurn(context.Context, *AgentTurnActionRequest) (*AgentTurnActionResponse, error)

	StreamAgentSessionEvents(context.Context, *StreamAgentSessionEventsRequest, sse.Sender)
}

func (o operations) RegisterAi(api huma.API) {
	huma.Register(api, ListAiAgents, o.ListAiAgents)

	huma.Register(api, CreateAgentSession, o.CreateAgentSession)
	huma.Register(api, ListAgentSessions, o.ListAgentSessions)
	huma.Register(api, GetAgentSession, o.GetAgentSession)
	huma.Register(api, ListAgentMessages, o.ListAgentMessages)
	huma.Register(api, ListAgentArtifacts, o.ListAgentArtifacts)

	huma.Register(api, ListAgentTurns, o.ListAgentTurns)
	huma.Register(api, RequestAgentTurn, o.RequestAgentTurn)
	huma.Register(api, AbortAgentTurn, o.AbortAgentTurn)
	huma.Register(api, RetryAgentTurn, o.RetryAgentTurn)

	sse.Register(api, StreamAgentSessionEvents, StreamAgentSessionEventsTypes, o.StreamAgentSessionEvents)
}

type (
	AgentSession struct {
		Id         uuid.UUID              `json:"id"`
		Attributes AgentSessionAttributes `json:"attributes"`
	}

	AgentSessionAttributes struct {
		AgentName        string     `json:"agentName"`
		PermissionScopes []string   `json:"permissionScopes"`
		CreatedAt        time.Time  `json:"createdAt"`
		UpdatedAt        time.Time  `json:"updatedAt"`
		SystemAnalysisId *uuid.UUID `json:"systemAnalysisId,omitempty"`
	}

	AgentMessage struct {
		Id         uuid.UUID              `json:"id"`
		Attributes AgentMessageAttributes `json:"attributes"`
	}
	AgentMessageAttributes struct {
		AgentTurnId uuid.UUID          `json:"agentTurnId"`
		Sequence    int                `json:"sequence"`
		Role        string             `json:"role"`
		Parts       []AgentMessagePart `json:"parts"`
		Visible     bool               `json:"visible"`
		CreatedAt   time.Time          `json:"createdAt"`
		UpdatedAt   time.Time          `json:"updatedAt"`
	}

	AgentArtifact struct {
		Id         uuid.UUID               `json:"id"`
		Attributes AgentArtifactAttributes `json:"attributes"`
	}
	AgentArtifactAttributes struct {
		LastAgentTurnId *uuid.UUID         `json:"lastAgentTurnId,omitempty"`
		Name            string             `json:"name"`
		Parts           []AgentMessagePart `json:"parts"`
		CreatedAt       time.Time          `json:"createdAt"`
		UpdatedAt       time.Time          `json:"updatedAt"`
	}

	AgentMessagePart struct {
		Kind        string             `json:"kind"`
		Text        *string            `json:"text,omitempty"`
		ContentType *string            `json:"contentType,omitempty"`
		Url         *string            `json:"url,omitempty"`
		Data        *string            `json:"data,omitempty"`
		Ref         *string            `json:"ref,omitempty"`
		Name        *string            `json:"name,omitempty"`
		Input       any                `json:"input,omitempty"`
		Partial     *bool              `json:"partial,omitempty"`
		Output      any                `json:"output,omitempty"`
		Content     []AgentMessagePart `json:"content,omitempty"`
		Uri         *string            `json:"uri,omitempty"`
		Custom      map[string]any     `json:"custom,omitempty"`
	}

	// AgentProviderRead is a provider read an agent made: for which service and range, the integration that ran
	// it, the native query, a link opening it in the provider and the items the agent highlighted. A failed read
	// has no query or link.
	AgentProviderRead struct {
		Id         uuid.UUID                   `json:"id"`
		Attributes AgentProviderReadAttributes `json:"attributes"`
	}

	AgentProviderReadAttributes struct {
		ToolName string `json:"toolName"`
		// IntegrationProvider and IntegrationDisplayName are empty once the integration is deleted.
		IntegrationProvider    string                  `json:"integrationProvider"`
		IntegrationDisplayName string                  `json:"integrationDisplayName"`
		ServiceName            string                  `json:"serviceName"`
		Start                  time.Time               `json:"start"`
		End                    time.Time               `json:"end"`
		Query                  string                  `json:"query"`
		Limits                 []string                `json:"limits"`
		SourceUrl              string                  `json:"sourceUrl"`
		Status                 string                  `json:"status" enum:"running,succeeded,failed"`
		Highlight              []AgentProviderReadItem `json:"highlight" doc:"Items the agent cited, newest citation first"`
	}

	// AgentProviderReadItem is a log line with its text, or a counted point with its value.
	AgentProviderReadItem struct {
		Ref    string            `json:"ref"`
		Time   time.Time         `json:"time" doc:"The newest repeat's time when count is above one"`
		Text   string            `json:"text,omitempty"`
		Value  *float64          `json:"value,omitempty"`
		Labels map[string]string `json:"labels"`
		Count  int               `json:"count" doc:"How many lines this entry stands for, identical or differing only in identifiers"`
		// Earliest is the oldest repeat's time; set when Count is above one.
		Earliest *time.Time `json:"earliest,omitempty"`
	}

	AgentTurn struct {
		Id         uuid.UUID           `json:"id"`
		Attributes AgentTurnAttributes `json:"attributes"`
	}

	AgentTurnAttributes struct {
		Status       string          `json:"status"`
		Sequence     int             `json:"sequence"`
		CreatedAt    time.Time       `json:"createdAt"`
		UpdatedAt    time.Time       `json:"updatedAt"`
		StartedAt    *time.Time      `json:"startedAt,omitempty"`
		FinishedAt   *time.Time      `json:"finishedAt,omitempty"`
		FinishReason string          `json:"finishReason,omitempty"`
		Error        *AgentTurnError `json:"error,omitempty"`
	}

	AgentTurnError struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	}

	AgentTurnStatusOverview struct {
		Id         uuid.UUID        `json:"id"`
		Status     agentturn.Status `json:"status"`
		FinishedAt *time.Time       `json:"finishedAt,omitempty"`
	}

	AgentTurnUpdatedEvent struct {
		StartedAt    *time.Time `json:"startedAt,omitempty"`
		SessionId    uuid.UUID  `json:"sessionId"`
		TurnId       uuid.UUID  `json:"turnId"`
		Status       string     `json:"status"`
		FinishReason string     `json:"finishReason,omitempty"`
	}

	AgentTurnChunkEvent struct {
		StartedAt *time.Time          `json:"startedAt,omitempty"`
		SessionId uuid.UUID           `json:"sessionId"`
		TurnId    uuid.UUID           `json:"turnId"`
		Model     *AgentModelChunk    `json:"model,omitempty"`
		Artifact  *AgentArtifactChunk `json:"artifact,omitempty"`
		TurnEnd   *AgentTurnEndChunk  `json:"turnEnd,omitempty"`
	}
	AgentModelChunk struct {
		Aggregated bool               `json:"aggregated"`
		Index      int                `json:"index"`
		Role       string             `json:"role"`
		Parts      []AgentMessagePart `json:"parts"`
	}
	AgentArtifactChunk struct {
		Name  string             `json:"name"`
		Parts []AgentMessagePart `json:"parts"`
	}
	AgentTurnEndChunk struct {
		FinishReason string `json:"finishReason"`
	}

	AiAgentConfig struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Model       string `json:"model"`
	}
)

func (o operations) RegisterAiEnums(api huma.API) {
	registerEnumAlias[agentturn.Status, agentTurnStatusSchema](api)
}

type agentTurnStatusSchema agentturn.Status

func (agentTurnStatusSchema) Schema(huma.Registry) *huma.Schema {
	return makeEnumStringSchema(agentturn.StatusValues)
}

func AgentTurnChunkEventFromRez(sessionId, turnId uuid.UUID, startedAt *time.Time, chunk rez.AiAgentTurnChunk) (*AgentTurnChunkEvent, error) {
	event := &AgentTurnChunkEvent{
		StartedAt: startedAt,
		SessionId: sessionId,
		TurnId:    turnId,
	}
	if chunk.ModelChunk != nil {
		parts, partsErr := MaybeConvertSlice(chunk.ModelChunk.Content, AgentMessagePartFromGenkit)
		if partsErr != nil {
			return nil, fmt.Errorf("model chunk: %w", partsErr)
		}
		event.Model = &AgentModelChunk{
			Aggregated: chunk.ModelChunk.Aggregated,
			Index:      chunk.ModelChunk.Index,
			Role:       string(chunk.ModelChunk.Role),
			Parts:      parts,
		}
	}
	if chunk.Artifact != nil {
		parts, partsErr := MaybeConvertSlice(chunk.Artifact.Parts, AgentMessagePartFromGenkit)
		if partsErr != nil {
			return nil, fmt.Errorf("artifact chunk: %w", partsErr)
		}
		event.Artifact = &AgentArtifactChunk{
			Name:  chunk.Artifact.Name,
			Parts: parts,
		}
	}
	if chunk.TurnEndFinishReason != nil {
		event.TurnEnd = &AgentTurnEndChunk{
			FinishReason: string(*chunk.TurnEndFinishReason),
		}
	}
	return event, nil
}

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
		PermissionScopes: session.Scopes,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
	}
	return AgentSession{Id: session.ID, Attributes: attrs}
}

func AgentTurnStatusOverviewFromDetail(t *ent.AgentTurn) *AgentTurnStatusOverview {
	if t == nil {
		return nil
	}
	return &AgentTurnStatusOverview{Id: t.ID, Status: t.Status, FinishedAt: t.FinishedAt}
}

func AgentMessagePartFromGenkit(part *ai.Part) (*AgentMessagePart, error) {
	if part == nil {
		return nil, fmt.Errorf("agent message part is nil")
	}

	switch part.Kind {
	case ai.PartText:
		return &AgentMessagePart{Kind: "text", Text: &part.Text}, nil
	case ai.PartReasoning:
		return &AgentMessagePart{Kind: "reasoning", Text: &part.Text}, nil
	case ai.PartMedia:
		return &AgentMessagePart{Kind: "media", ContentType: &part.ContentType, Url: &part.Text}, nil
	case ai.PartData:
		return &AgentMessagePart{Kind: "data", Data: &part.Text}, nil
	case ai.PartToolRequest:
		if req := part.ToolRequest; req != nil {
			return &AgentMessagePart{
				Kind:    "tool-request",
				Ref:     &req.Ref,
				Name:    &req.Name,
				Input:   req.Input,
				Partial: &req.Partial,
			}, nil
		}
		return nil, fmt.Errorf("tool request payload is nil")
	case ai.PartToolResponse:
		if part.ToolResponse == nil {
			return nil, fmt.Errorf("tool response payload is nil")
		}
		content, contentErr := MaybeConvertSlice(part.ToolResponse.Content, AgentMessagePartFromGenkit)
		if contentErr != nil {
			return nil, fmt.Errorf("tool response content: %w", contentErr)
		}
		return &AgentMessagePart{
			Kind:    "tool-response",
			Ref:     &part.ToolResponse.Ref,
			Name:    &part.ToolResponse.Name,
			Output:  part.ToolResponse.Output,
			Content: content,
		}, nil
	case ai.PartResource:
		if part.Resource == nil {
			return nil, fmt.Errorf("resource payload is nil")
		}
		return &AgentMessagePart{Kind: "resource", Uri: &part.Resource.Uri}, nil
	case ai.PartCustom:
		return &AgentMessagePart{Kind: "custom", Custom: part.Custom}, nil
	default:
		return nil, fmt.Errorf("unrecognized agent message part kind %d", part.Kind)
	}
}

func AgentMessageFromEnt(msg *ent.AgentMessage) (*AgentMessage, error) {
	parts, partsErr := MaybeConvertSlice(msg.Content, AgentMessagePartFromGenkit)
	if partsErr != nil {
		return nil, partsErr
	}
	attrs := AgentMessageAttributes{
		AgentTurnId: msg.AgentTurnID,
		Sequence:    msg.Sequence,
		Role:        msg.Role.String(),
		Parts:       parts,
		Visible:     msg.Visible,
		CreatedAt:   msg.CreatedAt,
		UpdatedAt:   msg.UpdatedAt,
	}
	return &AgentMessage{Id: msg.ID, Attributes: attrs}, nil
}

func AgentArtifactFromEnt(artifact *ent.AgentArtifact) (*AgentArtifact, error) {
	parts, partsErr := MaybeConvertSlice(artifact.Parts, AgentMessagePartFromGenkit)
	if partsErr != nil {
		return nil, partsErr
	}
	attrs := AgentArtifactAttributes{
		LastAgentTurnId: artifact.LastAgentTurnID,
		Name:            artifact.Name,
		Parts:           parts,
		CreatedAt:       artifact.CreatedAt,
		UpdatedAt:       artifact.UpdatedAt,
	}
	return &AgentArtifact{Id: artifact.ID, Attributes: attrs}, nil
}

// AgentProviderReadFromEnt takes the range from the read's input, which a read that failed also has, the query
// and link from the read that ran, and the integration from its edge.
func AgentProviderReadFromEnt(read *ent.AgentProviderRead) AgentProviderRead {
	attrs := AgentProviderReadAttributes{
		ToolName:  read.ToolName,
		Limits:    []string{},
		Status:    string(read.Status),
		Highlight: make([]AgentProviderReadItem, len(read.Highlight)),
	}
	if service := read.Edges.ServiceEntity; service != nil {
		attrs.ServiceName = service.State.DisplayName
	}
	if integration := read.Edges.Integration; integration != nil {
		attrs.IntegrationProvider = integration.Provider
		attrs.IntegrationDisplayName = integration.DisplayName
	}
	var input struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}
	if json.Unmarshal(read.Input, &input) == nil {
		attrs.Start, attrs.End = input.Start, input.End
	}
	var record rezai.ProviderReadRecord
	if len(read.Read) > 0 && json.Unmarshal(read.Read, &record) == nil {
		attrs.Query = record.Query
		attrs.SourceUrl = record.SourceURL
		if record.Limits != nil {
			attrs.Limits = record.Limits
		}
	}
	for index, item := range read.Highlight {
		converted := AgentProviderReadItem{
			Ref:      item.Ref,
			Time:     item.Time,
			Text:     item.Text,
			Value:    item.Value,
			Labels:   item.Labels,
			Count:    max(item.Count, 1),
			Earliest: item.Earliest,
		}
		if converted.Labels == nil {
			converted.Labels = map[string]string{}
		}
		attrs.Highlight[index] = converted
	}
	return AgentProviderRead{Id: read.ID, Attributes: attrs}
}

func AgentTurnFromEnt(turn *ent.AgentTurn) AgentTurn {
	attrs := AgentTurnAttributes{
		Status:       turn.Status.String(),
		Sequence:     turn.Sequence,
		CreatedAt:    turn.CreatedAt,
		UpdatedAt:    turn.UpdatedAt,
		StartedAt:    turn.StartedAt,
		FinishedAt:   turn.FinishedAt,
		FinishReason: turn.FinishReason,
	}
	if turn.Error != nil {
		attrs.Error = &AgentTurnError{
			Code:    turn.Status.String(),
			Message: *turn.Error,
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
type ListAiAgentsResponse CollectionResponse[AiAgentConfig]

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
	AgentName string `json:"agentName" minLength:"1"`
	Input     []byte `json:"input"`
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
	PaginationRequest
}
type ListAgentSessionsResponse PaginatedResponse[AgentSession]

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

var ListAgentMessages = openapi.Operation{
	OperationID: "list-agent-messages",
	Method:      http.MethodGet,
	Path:        "/ai/agent_sessions/{id}/messages",
	Summary:     "List Agent Session Messages",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAgentMessagesRequest PaginatedIdRequest
type ListAgentMessagesResponse PaginatedResponse[AgentMessage]

var ListAgentArtifacts = openapi.Operation{
	OperationID: "list-agent-artifacts",
	Method:      http.MethodGet,
	Path:        "/ai/agent_sessions/{id}/artifacts",
	Summary:     "List Agent Session Artifacts",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAgentArtifactsRequest PaginatedIdRequest
type ListAgentArtifactsResponse PaginatedResponse[AgentArtifact]

var ListAgentTurns = openapi.Operation{
	OperationID: "list-agent-turns",
	Method:      http.MethodGet,
	Path:        "/ai/agent_sessions/{id}/turns",
	Summary:     "List Agent Turns",
	Tags:        aiTags,
	Errors:      ErrorCodes(),
}

type ListAgentTurnsRequest struct {
	PaginatedIdRequest
}
type ListAgentTurnsResponse PaginatedResponse[AgentTurn]

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

var (
	StreamAgentSessionEvents = openapi.Operation{
		OperationID: "stream-agent-session-events",
		Method:      http.MethodGet,
		Path:        "/ai/agent_sessions/{id}/events",
		Summary:     "Stream Agent Session Events",
		Tags:        aiTags,
		Errors:      ErrorCodes(),
	}
	StreamAgentSessionEventsTypes = map[string]any{
		"turn-chunk":   AgentTurnChunkEvent{},
		"turn-updated": AgentTurnUpdatedEvent{},
	}
)

type StreamAgentSessionEventsRequest IdRequest
