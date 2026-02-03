package google

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog/log"
)

func (i *integration) registerMessageHandlers() error {
	eventsErr := i.services.Messages.AddEventHandlers(
		rez.NewEventHandler("GoogleOnIncidentUpdate", i.onIncidentUpdate))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	cmdsErr := i.services.Messages.AddCommandHandlers(
		rez.NewCommandHandler("Google.CreateIncidentVideoConference", i.createIncidentVideoConference))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	return nil
}

func (i *integration) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	if ev.Created {
		return i.onIncidentCreated(ctx, ev.IncidentId)
	}
	return nil
}

func (i *integration) onIncidentCreated(ctx context.Context, id uuid.UUID) error {
	ci, ciErr := i.lookupConfiguredIntegration(ctx)
	if ciErr != nil || ci == nil {
		log.Debug().Err(ciErr).Msg("google integration not configured")
		return ciErr
	}
	if ci.isVideoConferenceEnabled() {
		cmd := &cmdCreateIncidentVideoConference{IncidentId: id}
		if cmdErr := i.services.Messages.SendCommand(ctx, cmd); cmdErr != nil {
			return fmt.Errorf("failed to send command: %w", cmdErr)
		}
	}
	return nil
}

type cmdCreateIncidentVideoConference struct {
	IncidentId uuid.UUID `json:"incident_id"`
}

func (i *integration) createIncidentVideoConference(ctx context.Context, cmd *cmdCreateIncidentVideoConference) error {
	ci, ciErr := i.lookupConfiguredIntegration(ctx)
	if ciErr != nil || ci == nil {
		return ciErr
	}
	ms, msErr := ci.VideoConferenceIntegration(ctx)
	if msErr != nil {
		return fmt.Errorf("getting configured integration: %w", msErr)
	}
	inc, incErr := i.services.Incidents.Get(ctx, cmd.IncidentId)
	if incErr != nil {
		return fmt.Errorf("get incident: %w", incErr)
	}
	return ms.CreateIncidentVideoConference(ctx, inc)
}
