package slackagent

import (
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
)

type app struct {
	appCfg     rez.AppConfig
	db         rez.Database
	messages   rez.MessageService
	eventAnnos rez.EventAnnotationsService
}

func makeApp(cfg rez.Config, db rez.Database, msgs rez.MessageService, eventAnnos rez.EventAnnotationsService) (*app, error) {
	h := &app{
		appCfg:     cfg.App,
		db:         db,
		messages:   msgs,
		eventAnnos: eventAnnos,
	}
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (a *app) registerMessageHandlers() error {
	return errors.Join()
}

func (a *app) EventsApiHandler() slackintegration.EventsApiHandler {
	return a.handleEventsApiEvent
}

func (a *app) SlashCommandHandlers() map[string]slackintegration.SlashCommandHandler {
	return map[string]slackintegration.SlashCommandHandler{}
}

func (a *app) InteractionCallbackHandlers() map[slack.InteractionType]slackintegration.InteractionCallbackHandler {
	return map[slack.InteractionType]slackintegration.InteractionCallbackHandler{
		slack.InteractionTypeMessageAction:  a.handleMessageActionInteraction,
		slack.InteractionTypeBlockActions:   a.handleBlockActionsInteraction,
		slack.InteractionTypeViewSubmission: a.handleViewSubmissionInteraction,
	}
}
