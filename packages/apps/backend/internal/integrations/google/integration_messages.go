package google

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
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
	eventsErr := i.messages.AddEventHandlers(
		rez.NewEventHandler("Google.OnIncidentUpdate", mh.onIncidentUpdate))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	cmdsErr := i.messages.AddCommandHandlers(
		rez.NewCommandHandler("Google.CreateIncidentVideoConference", mh.createIncidentVideoConference))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	return nil
}

func (h *eventHandler) withInstallation(ctx context.Context, fn func(*InstalledIntegration) error) error {
	listParams := rez.ListIntegrationsParams{Providers: []string{integrationName}}
	intgs, lookupErr := h.integrations.ListInstalled(ctx, listParams)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil
		}
		return fmt.Errorf("error looking up Integration: %w", lookupErr)
	}
	if len(intgs) == 0 {
		return nil
	}
	// TODO: handle multiple installations
	if ci, ok := intgs[0].(*InstalledIntegration); ok {
		return fn(ci)
	}
	return fmt.Errorf("invalid configured Integration: %w", lookupErr)
}

func (h *eventHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	if ev.Created {
		return h.onIncidentCreated(ctx, ev.IncidentId)
	}
	return nil
}

func (h *eventHandler) onIncidentCreated(ctx context.Context, id uuid.UUID) error {
	return h.withInstallation(ctx, func(ci *InstalledIntegration) error {
		if !ci.isVideoConferenceEnabled() {
			return nil
		}
		return h.messages.SendCommand(ctx, &cmdCreateIncidentVideoConference{IncidentId: id})
	})
}

type cmdCreateIncidentVideoConference struct {
	IncidentId uuid.UUID `json:"incident_id"`
}

func (h *eventHandler) createIncidentVideoConference(ctx context.Context, cmd *cmdCreateIncidentVideoConference) error {
	return h.withInstallation(ctx, func(ci *InstalledIntegration) error {
		inc, incErr := h.incidents.Get(ctx, incident.ID(cmd.IncidentId))
		if incErr != nil {
			return fmt.Errorf("get incident: %w", incErr)
		}
		return newMeetService(ci).CreateIncidentVideoConference(ctx, inc)
	})
}
