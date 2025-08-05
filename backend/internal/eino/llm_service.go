package eino

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/cloudwego/eino/schema"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentdebriefmessage"
)

type LanguageModelService struct {
	prov rez.LanguageModelProvider
}

func NewLanguageModelService(prov rez.LanguageModelProvider) (*LanguageModelService, error) {
	return &LanguageModelService{prov: prov}, nil
}

var (
	incidentDebriefSystemPrompt = schema.SystemMessage(
		`You are an AI model conducting an incident debrief. Details about the incident`)
)

func (s *LanguageModelService) GenerateDebriefResponse(ctx context.Context, debrief *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error) {
	if rez.DebugMode {
		log.Debug().Msg("TODO: faking ai response")
		return &ent.IncidentDebriefMessage{
			DebriefID: debrief.ID,
			CreatedAt: time.Now(),
			Type:      incidentdebriefmessage.TypeAssistant,
			Body:      "faked message",
		}, nil
	}

	debriefMessages, msgErr := debrief.Edges.MessagesOrErr()
	if msgErr != nil {
		return nil, msgErr
	}

	model := s.prov.Model()
	resp, genErr := model.Generate(ctx, createDebriefThread(debriefMessages))
	if genErr != nil {
		return nil, fmt.Errorf("generate content: %w", genErr)
	}

	respMsg := &ent.IncidentDebriefMessage{
		Type: incidentdebriefmessage.TypeAssistant,
		Body: resp.Content,
	}

	return respMsg, nil
}

func createDebriefThread(messages []*ent.IncidentDebriefMessage) []*schema.Message {
	thread := []*schema.Message{incidentDebriefSystemPrompt}

	for _, debriefMsg := range messages {
		var msg *schema.Message

		if debriefMsg.Type == incidentdebriefmessage.TypeUser {
			msg = schema.UserMessage(debriefMsg.Body)
		} else {
			msg = schema.AssistantMessage(debriefMsg.Body, nil)
		}

		if msg == nil {
			log.Warn().Str("id", debriefMsg.ID.String()).Msg("empty debrief message?")
			continue
		}
		thread = append(thread, msg)
	}

	return thread
}
