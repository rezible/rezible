package google

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/predicate"
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
	listParams := rez.ListIntegrationsParams{
		Predicates: []predicate.Integration{integration.IntegrationName(integrationName)},
	}
	intgs, lookupErr := h.integrations.ListInstalled(ctx, listParams)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil
		}
		return fmt.Errorf("error looking up Integration: %w", lookupErr)
	}
	if len(intgs) == 0 {
		return nil
	} else if len(intgs) > 1 {
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
				return nil
			}
			return h.messages.SendCommand(ctx, &cmdCreateIncidentVideoConference{IncidentId: ev.IncidentId})
		})
	}
	return nil
}

type cmdCreateIncidentVideoConference struct {
	IncidentId uuid.UUID `json:"incident_id"`
}

func (h *eventHandler) createIncidentVideoConference(ctx context.Context, cmd *cmdCreateIncidentVideoConference) error {
	//return h.withInstallation(ctx, func(ii *InstalledIntegration) error {
	//	inc, incErr := h.incidents.Get(ctx, incident.ID(cmd.IncidentId))
	//	if incErr != nil {
	//		return fmt.Errorf("get incident: %w", incErr)
	//	}
	//	return newMeetService(ii).CreateIncidentVideoConference(ctx, inc)
	//})
	return nil
}
