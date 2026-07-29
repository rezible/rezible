package apiv1

import (
	"context"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type aiHandler struct {
	ai     rez.AiService
	agents rez.AgentSessionService
}

func newAiHandler(ai rez.AiService, agents rez.AgentSessionService) *aiHandler {
	return &aiHandler{ai: ai, agents: agents}
}

func (h *aiHandler) ListAiAgents(ctx context.Context, req *oapi.ListAiAgentsRequest) (*oapi.ListAiAgentsResponse, error) {
	var resp oapi.ListAiAgentsResponse
	resp.Body.Data = oapi.ConvertSlice(h.ai.GetAgents(), oapi.AiAgentConfigFromRez)
	return &resp, nil
}

func (h *aiHandler) CreateAgentSession(ctx context.Context, req *oapi.CreateAgentSessionRequest) (*oapi.CreateAgentSessionResponse, error) {
	var resp oapi.CreateAgentSessionResponse
	ownerId, ownerIdOk := execution.GetContext(ctx).UserID()
	if !ownerIdOk {
		return nil, oapi.ErrAuthSessionMissing
	}
	attrs := req.Body.Attributes
	params := rez.CreateAgentSessionParams{
		AgentName:   attrs.AgentName,
		OwnerUserID: ownerId,
		Input:       attrs.Input,
	}
	session, createErr := h.agents.CreateAgentSession(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create agent session", createErr)
	}
	resp.Body.Data = oapi.AgentSessionFromEnt(session)
	return &resp, nil
}

func (h *aiHandler) ListAgentSessions(ctx context.Context, req *oapi.ListAgentSessionsRequest) (*oapi.ListAgentSessionsResponse, error) {
	if _, ownerIdOk := execution.GetContext(ctx).UserID(); !ownerIdOk {
		return nil, oapi.ErrAuthSessionMissing
	}

	params := rez.ListAgentSessionsParams{
		ListParams: req.ListParams(),
	}
	sessions, listErr := h.agents.ListAgentSessions(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent sessions", listErr)
	}

	var resp oapi.ListAgentSessionsResponse
	resp.Body.Data = make([]oapi.AgentSession, len(sessions.Data))
	for i, session := range sessions.Data {
		resp.Body.Data[i] = oapi.AgentSessionFromEnt(session)
	}
	resp.Body.Pagination.Total = sessions.Count
	return &resp, nil
}

func (h *aiHandler) GetAgentSession(ctx context.Context, req *oapi.GetAgentSessionRequest) (*oapi.GetAgentSessionResponse, error) {
	var resp oapi.GetAgentSessionResponse
	session, getErr := h.agents.GetAgentSession(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get agent session", getErr)
	}
	resp.Body.Data = oapi.AgentSessionFromEnt(session)
	return &resp, nil
}

func (h *aiHandler) ListAgentTurns(ctx context.Context, req *oapi.ListAgentTurnsRequest) (*oapi.ListAgentTurnsResponse, error) {
	var resp oapi.ListAgentTurnsResponse
	params := rez.ListAgentTurnsParams{
		ListParams: req.ListParams(),
		Predicates: []predicate.AgentTurn{agentturn.AgentSessionID(req.Id)},
	}
	turns, listErr := h.agents.ListAgentTurns(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent turns", listErr)
	}
	resp.Body.Data = make([]oapi.AgentTurn, len(turns.Data))
	for i, turn := range turns.Data {
		resp.Body.Data[i] = oapi.AgentTurnFromEnt(turn)
	}
	resp.Body.Pagination.Total = turns.Count
	return &resp, nil
}

func (h *aiHandler) makeTurnRequestInput(attrs oapi.RequestAgentTurnRequestAttributes) (*rez.AgentTurnInput, error) {
	var message *ai.Message
	if attrs.Message != nil {
		trimmed := strings.TrimSpace(*attrs.Message)
		if trimmed == "" {
			return nil, fmt.Errorf("%w: message must not be blank", rez.ErrInvalidInput)
		}
		message = ai.NewUserTextMessage(trimmed)
	}
	var resume *ai.GenerateActionResume
	if attrs.Resume != nil && len(attrs.Resume.Respond)+len(attrs.Resume.Restart) > 0 {
		resume = &ai.GenerateActionResume{Respond: attrs.Resume.Respond, Restart: attrs.Resume.Restart}
	}
	return &rez.AgentTurnInput{Message: message, Resume: resume}, nil
}

func (h *aiHandler) RequestAgentTurn(ctx context.Context, req *oapi.RequestAgentTurnRequest) (*oapi.RequestAgentTurnResponse, error) {
	var resp oapi.RequestAgentTurnResponse
	input, inputErr := h.makeTurnRequestInput(req.Body.Attributes)
	if inputErr != nil {
		return nil, oapi.Error(ctx, "invalid agent turn input", inputErr)
	}
	params := &rez.RequestAgentTurnParams{
		Input:        input,
		ParentTurnID: nil,
	}
	turn, requestErr := h.agents.RequestAgentTurn(ctx, req.Id, params)
	if requestErr != nil {
		return nil, oapi.Error(ctx, "request agent turn", requestErr)
	}
	resp.Body.Data = oapi.AgentTurnFromEnt(turn)
	return &resp, nil
}

func (h *aiHandler) AbortAgentTurn(ctx context.Context, req *oapi.AgentTurnActionRequest) (*oapi.AgentTurnActionResponse, error) {
	var resp oapi.AgentTurnActionResponse
	turn, abortErr := h.agents.AbortAgentTurn(ctx, req.Id)
	if abortErr != nil {
		return nil, oapi.Error(ctx, "abort agent turn", abortErr)
	}
	resp.Body.Data = oapi.AgentTurnFromEnt(turn)
	return &resp, nil
}

func (h *aiHandler) RetryAgentTurn(ctx context.Context, req *oapi.AgentTurnActionRequest) (*oapi.AgentTurnActionResponse, error) {
	var resp oapi.AgentTurnActionResponse
	turn, retryErr := h.agents.RetryAgentTurn(ctx, req.Id)
	if retryErr != nil {
		return nil, oapi.Error(ctx, "retry agent turn", retryErr)
	}
	resp.Body.Data = oapi.AgentTurnFromEnt(turn)
	return &resp, nil
}
