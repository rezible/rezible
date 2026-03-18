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
	services *rez.Services
}

func (i *integration) registerMessageHandlers() error {
	mh := &eventHandler{services: i.services}
	eventsErr := i.services.Messages.AddEventHandlers(
		rez.NewEventHandler("GoogleOnIncidentUpdate", mh.onIncidentUpdate))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	cmdsErr := i.services.Messages.AddCommandHandlers(
		rez.NewCommandHandler("Google.CreateIncidentVideoConference", mh.createIncidentVideoConference))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	return nil
}

func (h *eventHandler) withConfiguredIntegration(ctx context.Context, fn func(*ConfiguredIntegration) error) error {
	intg, lookupErr := h.services.Integrations.GetConfigured(ctx, integrationName)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil
		}
		return fmt.Errorf("error looking up integration: %w", lookupErr)
	}
	if ci, ok := intg.(*ConfiguredIntegration); ok {
		return fn(ci)
	}
	return fmt.Errorf("invalid configured integration: %w", lookupErr)
}

func (h *eventHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	if ev.Created {
		return h.onIncidentCreated(ctx, ev.IncidentId)
	}
	return nil
}

func (h *eventHandler) onIncidentCreated(ctx context.Context, id uuid.UUID) error {
	return h.withConfiguredIntegration(ctx, func(ci *ConfiguredIntegration) error {
		if !ci.isVideoConferenceEnabled() {
			return nil
		}
		return h.services.Messages.SendCommand(ctx, &cmdCreateIncidentVideoConference{IncidentId: id})
	})
}

type cmdCreateIncidentVideoConference struct {
	IncidentId uuid.UUID `json:"incident_id"`
}

func (h *eventHandler) createIncidentVideoConference(ctx context.Context, cmd *cmdCreateIncidentVideoConference) error {
	return h.withConfiguredIntegration(ctx, func(ci *ConfiguredIntegration) error {
		inc, incErr := h.services.Incidents.Get(ctx, incident.ID(cmd.IncidentId))
		if incErr != nil {
			return fmt.Errorf("get incident: %w", incErr)
		}
		return newMeetService(ci).CreateIncidentVideoConference(ctx, inc)
	})
}
