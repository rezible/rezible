package google

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/pkg/messages"
)

type eventHandler struct {
	integrations rez.IntegrationService
	messages     rez.MessageService
	incidents    rez.IncidentService
}

func (i *Integration) registerMessageHandlers() error {
	mh := &eventHandler{
		integrations: i.integrations,
		messages:     i.messages,
		incidents:    i.incidents,
	}
	eventsErr := i.messages.AddHandlers(
		messages.NewEventHandler("Google.OnIncidentUpdate", mh.onIncidentUpdate))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	return nil
}

func (h *eventHandler) withInstallation(ctx context.Context, fn func(*InstalledIntegration) error) error {
	intgs, lookupErr := h.integrations.ListAllInstalled(ctx, integration.Name(integrationName))
	if lookupErr != nil && !ent.IsNotFound(lookupErr) {
		return fmt.Errorf("error looking up Integration: %w", lookupErr)
	}
	if len(intgs) == 0 {
		return nil
	}
	if len(intgs) > 1 {
		return fmt.Errorf("found multiple Integrations with name %q", integrationName)
	}
	ii, ok := intgs[0].(*InstalledIntegration)
	if !ok {
		return fmt.Errorf("invalid configured Integration: %w", lookupErr)
	}
	return fn(ii)
}

func (h *eventHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	if ev.Created {
		return h.withInstallation(ctx, func(ii *InstalledIntegration) error {
			if !ii.isVideoConferenceEnabled() {
				// TODO: create video conference?
			}
			return nil
		})
	}
	return nil
}
