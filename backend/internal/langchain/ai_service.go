package langchain

import (
	"context"
	"errors"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/tmc/langchaingo/llms"
	"strings"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentdebriefmessage"
)

type AiService struct {
	prov rez.AiModelProvider
}

func NewAiService(ctx context.Context, pl rez.ProviderLoader) (*AiService, error) {
	prov, aiProvErr := pl.LoadAiModelProvider(ctx)
	if aiProvErr != nil {
		return nil, fmt.Errorf("failed to create AI model provider: %w", aiProvErr)
	}

	return &AiService{prov: prov}, nil
}

var (
	incidentDebriefSystemPrompt = llms.TextContent{
		Text: `You are an AI model conducting an incident debrief. Details about the incident`,
	}
	debriefResponseTools = []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        incidentdebriefmessage.RequestedToolRating.String(),
				Description: "Get a structured rating from the user out of 5",
				Parameters:  nil,
			},
		},
	}
)

func (s *AiService) GenerateDebriefResponse(ctx context.Context, debrief *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error) {
	debriefMessages, msgErr := debrief.Edges.MessagesOrErr()
	if msgErr != nil {
		return nil, msgErr
	}

	systemMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{incidentDebriefSystemPrompt},
	}
	thread := append([]llms.MessageContent{systemMessage}, convertDebriefMessagesToThread(debriefMessages)...)

	model := s.prov.GetModel()
	content, genErr := model.GenerateContent(ctx, thread, llms.WithTools(debriefResponseTools))
	if genErr != nil {
		return nil, fmt.Errorf("generate content: %w", genErr)
	}

	return convertDebriefResponseToMessage(content.Choices)
}

func convertDebriefMessagesToThread(messages []*ent.IncidentDebriefMessage) []llms.MessageContent {
	var thread []llms.MessageContent

	var toolRequestMsg *ent.IncidentDebriefMessage
	for _, debriefMsg := range messages {
		msg := debriefMsg

		var role llms.ChatMessageType
		var parts []llms.ContentPart

		if msg.Type == incidentdebriefmessage.TypeUser {
			role = llms.ChatMessageTypeHuman
			var part llms.ContentPart
			if toolRequestMsg != nil {
				part = llms.ToolCallResponse{
					ToolCallID: toolRequestMsg.ID.String(),
					Name:       toolRequestMsg.RequestedTool.String(),
					Content:    msg.Body,
				}
				toolRequestMsg = nil
			} else {
				part = llms.TextPart(msg.Body)
			}
			parts = []llms.ContentPart{part}
		} else {
			role = llms.ChatMessageTypeAI
			parts = []llms.ContentPart{llms.TextPart(msg.Body)}
			toolRequest := msg.RequestedTool.String()
			if toolRequest != "" {
				parts = append(parts, llms.ToolCall{
					ID: msg.ID.String(),
					FunctionCall: &llms.FunctionCall{
						Name: toolRequest,
					},
				})
				toolRequestMsg = msg
			}
		}

		thread = append(thread, llms.MessageContent{Role: role, Parts: parts})
	}

	return thread
}

func convertDebriefResponseToMessage(choices []*llms.ContentChoice) (*ent.IncidentDebriefMessage, error) {
	if len(choices) == 0 {
		return nil, errors.New("no content returned")
	}

	msg := &ent.IncidentDebriefMessage{
		Type: incidentdebriefmessage.TypeAssistant,
	}
	bodies := make([]string, len(choices))
	for i, choice := range choices {
		bodies[i] = choice.Content
		if choice.FuncCall != nil {
			name := choice.FuncCall.Name
			if isValidDebriefToolRequest(name) {
				msg.RequestedTool = incidentdebriefmessage.RequestedTool(name)
			}
		}
	}
	msg.Body = strings.Join(bodies, "\n")

	return msg, nil
}

func isValidDebriefToolRequest(callName string) bool {
	if callName == incidentdebriefmessage.RequestedToolRating.String() {
		return true
	}
	return false
}
