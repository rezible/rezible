package apiv1

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentartifact"
	"github.com/rezible/rezible/ent/agentmessage"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/messages"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"golang.org/x/sync/errgroup"
)

type aiHandler struct {
	ai     rez.AiService
	agents rez.AgentSessionService
	msgs   rez.MessageService
}

func newAiHandler(ai rez.AiService, agents rez.AgentSessionService, msgs rez.MessageService) *aiHandler {
	return &aiHandler{
		ai:     ai,
		agents: agents,
		msgs:   msgs,
	}
}

func (h *aiHandler) ListAiAgents(ctx context.Context, req *oapi.ListAiAgentsRequest) (*oapi.ListAiAgentsResponse, error) {
	var resp oapi.ListAiAgentsResponse
	resp.Body.Data = oapi.ConvertSlice(h.ai.GetAgents(), oapi.AiAgentConfigFromRez)
	return &resp, nil
}

func (h *aiHandler) CreateAgentSession(ctx context.Context, req *oapi.CreateAgentSessionRequest) (*oapi.CreateAgentSessionResponse, error) {
	var resp oapi.CreateAgentSessionResponse
	attrs := req.Body.Attributes

	input, inputErr := h.ai.ValidateAgentSessionInput(attrs.AgentName, attrs.Input)
	if inputErr != nil {
		return nil, oapi.Error(ctx, "invalid input", inputErr)
	}

	params := rez.CreateAgentSessionParams{
		AgentName: attrs.AgentName,
		Input:     input,
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
	resp.Body = oapi.ConvertPaginatedResultBody(sessions, oapi.AgentSessionFromEnt)
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

func (h *aiHandler) ListAgentMessages(ctx context.Context, req *oapi.ListAgentMessagesRequest) (*oapi.ListAgentMessagesResponse, error) {
	params := rez.ListAgentMessagesParams{
		ListParams: req.ListParams(),
		Predicates: []predicate.AgentMessage{agentmessage.AgentSessionID(req.Id)},
	}
	msgs, listErr := h.agents.ListAgentMessages(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent session messages", listErr)
	}
	body, bodyErr := oapi.MaybeConvertPaginatedResultBody(msgs, oapi.AgentMessageFromEnt)
	if bodyErr != nil {
		return nil, oapi.Error(ctx, "convert agent session messages", bodyErr)
	}
	return &oapi.ListAgentMessagesResponse{Body: *body}, nil
}

func (h *aiHandler) ListAgentArtifacts(ctx context.Context, req *oapi.ListAgentArtifactsRequest) (*oapi.ListAgentArtifactsResponse, error) {
	params := rez.ListAgentArtifactsParams{
		ListParams: req.ListParams(),
		Predicates: []predicate.AgentArtifact{agentartifact.AgentSessionID(req.Id)},
	}
	artifacts, listErr := h.agents.ListAgentArtifacts(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent session artifacts", listErr)
	}
	body, bodyErr := oapi.MaybeConvertPaginatedResultBody(artifacts, oapi.AgentArtifactFromEnt)
	if bodyErr != nil {
		return nil, oapi.Error(ctx, "convert agent session artifacts", bodyErr)
	}
	return &oapi.ListAgentArtifactsResponse{Body: *body}, nil
}

func (h *aiHandler) ListAgentTurns(ctx context.Context, req *oapi.ListAgentTurnsRequest) (*oapi.ListAgentTurnsResponse, error) {
	params := rez.ListAgentTurnsParams{
		ListParams: req.ListParams(),
		Predicates: []predicate.AgentTurn{agentturn.AgentSessionID(req.Id)},
	}
	turns, listErr := h.agents.ListAgentTurns(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list agent turns", listErr)
	}
	body := oapi.ConvertPaginatedResultBody(turns, oapi.AgentTurnFromEnt)
	return &oapi.ListAgentTurnsResponse{Body: body}, nil
}

func (h *aiHandler) makeTurnRequestInput(attrs oapi.RequestAgentTurnRequestAttributes) (*rez.AiAgentTurnInput, error) {
	if attrs.Message != nil {
		if attrs.Resume != nil {
			return nil, fmt.Errorf("cannot specify message and resume")
		}
		msg := ai.NewUserTextMessage(strings.TrimSpace(*attrs.Message))
		if msg.Text() == "" {
			return nil, fmt.Errorf("message must not be blank")
		}
		return &rez.AiAgentTurnInput{Message: msg}, nil
	} else if attrs.Resume != nil {
		if len(attrs.Resume.Respond)+len(attrs.Resume.Restart) == 0 {
			return nil, fmt.Errorf("empty resume params")
		}
		resume := &aix.ToolResume{
			Respond: attrs.Resume.Respond,
			Restart: attrs.Resume.Restart,
		}
		return &rez.AiAgentTurnInput{Resume: resume}, nil
	}
	return nil, fmt.Errorf("exactly one of message or resume is required")
}

func (h *aiHandler) RequestAgentTurn(ctx context.Context, req *oapi.RequestAgentTurnRequest) (*oapi.RequestAgentTurnResponse, error) {
	var resp oapi.RequestAgentTurnResponse
	input, inputErr := h.makeTurnRequestInput(req.Body.Attributes)
	if inputErr != nil {
		return nil, oapi.Error(ctx, "invalid agent turn input", fmt.Errorf("%w: %w", rez.ErrInvalidInput, inputErr))
	}
	params := &rez.RequestAgentTurnParams{
		Input: input,
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

func (h *aiHandler) StreamAgentSessionEvents(ctx context.Context, req *oapi.StreamAgentSessionEventsRequest, send sse.Sender) {
	// The sender can end the whole stream when the client connection stops accepting writes.
	streamCtx, cancelStreamContext := context.WithCancel(ctx)
	defer cancelStreamContext()

	// A subscription failure cancels the sender, while a sender failure closes the subscriptions.
	g, subscriberCtx := errgroup.WithContext(streamCtx)

	// Subscription callbacks enqueue events so only the sender goroutine writes to the response.
	events := make(chan sse.Message, 8)
	eventHandlers := h.makeAgentSessionEventHandlers(func(ctx context.Context, data any) error {
		select {
		case events <- sse.Message{Data: data}:
			return nil
		case <-ctx.Done():
			return nil
		}
	})

	g.Go(func() error {
		for {
			select {
			case event := <-events:
				if sendErr := send(event); sendErr != nil {
					// Write failures normally mean the client disconnected, not a server failure.
					cancelStreamContext()
					return nil
				}
			case <-subscriberCtx.Done():
				return nil
			}
		}
	})

	subscribeOpts := &rez.MessageEventSubscriptionOpts{
		Scopes: []string{"agent_session:" + req.Id.String()},
	}
	g.Go(func() error {
		return h.msgs.Subscribe(subscriberCtx, subscribeOpts, eventHandlers...)
	})

	// Sender termination is clean, so any returned error came from a subscription.
	if subErr := g.Wait(); subErr != nil {
		slog.WarnContext(ctx, "agent session event subscription failed",
			"error", subErr,
			"sessionId", req.Id)
	}
}

func (h *aiHandler) makeAgentSessionEventHandlers(sendEvent func(context.Context, any) error) []rez.MessageEventHandler {
	onTurnChunkFn := func(ctx context.Context, event *rezai.EventOnAgentTurnChunk) error {
		converted, convertErr := oapi.AgentTurnChunkEventFromRez(event.AgentSessionId, event.AgentTurnId, event.Chunk)
		if convertErr != nil {
			return convertErr
		}
		return sendEvent(ctx, *converted)
	}
	onTurnUpdateFn := func(ctx context.Context, event *rezai.AgentTurnUpdated) error {
		return sendEvent(ctx, oapi.AgentTurnUpdatedEvent{
			SessionId:    event.AgentSessionId,
			TurnId:       event.AgentTurnId,
			Status:       event.Status.String(),
			FinishReason: event.FinishReason,
		})
	}
	return []rez.MessageEventHandler{
		messages.NewEventHandler("agent-session-sse-chunks", onTurnChunkFn),
		messages.NewEventHandler("agent-session-sse-updates", onTurnUpdateFn),
	}
}
