package google

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
)

type serviceLoader struct{}

type incidentEventHandler struct {
	msgs      rez.MessageService
	incidents rez.IncidentService
	svcLoader *serviceLoader
}

func newIncidentEventHandler(sl *serviceLoader, msgs rez.MessageService, incidents rez.IncidentService) (*incidentEventHandler, error) {
	h := &incidentEventHandler{svcLoader: sl, msgs: msgs, incidents: incidents}
	if hErr := h.registerHandlers(); hErr != nil {
		return nil, fmt.Errorf("registering message handlers: %w", hErr)
	}
	return h, nil
}

func (h *incidentEventHandler) registerHandlers() error {
	cmdsErr := h.msgs.AddCommandHandlers()
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	eventsErr := h.msgs.AddEventHandlers(
		rez.NewEventHandler("GoogleOnIncidentUpdate", h.onIncidentUpdate))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}

	return nil
}

func (h *incidentEventHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	
	return nil
}
