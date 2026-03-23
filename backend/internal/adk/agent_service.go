package adk

import (
	"context"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentdebriefmessage"
)

type AgentService struct {
}

func NewAgentService() (*AgentService, error) {
	return &AgentService{}, nil
}

func (s *AgentService) GenerateDebriefResponse(ctx context.Context, debrief *ent.IncidentDebrief) (*ent.IncidentDebriefMessage, error) {
	// TODO: generate
	return &ent.IncidentDebriefMessage{
		DebriefID: debrief.ID,
		CreatedAt: time.Now(),
		Type:      incidentdebriefmessage.TypeAssistant,
		Body:      "faked message",
	}, nil
}
