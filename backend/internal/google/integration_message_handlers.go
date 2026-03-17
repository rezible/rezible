package google

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
)

type messageHandler struct {
	services *rez.Services
}

func (i *integration) registerMessageHandlers() error {
	mh := &messageHandler{services: i.services}
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

func (mh *messageHandler) lookupConfiguredIntegration(ctx context.Context) (*ConfiguredIntegration, error) {
	intg, lookupErr := mh.services.Integrations.GetConfigured(ctx, integrationName)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("error looking up integration: %w", lookupErr)
	}
	ci, ok := intg.(*ConfiguredIntegration)
	if !ok {
		return nil, fmt.Errorf("invalid configured integration: %w", lookupErr)
	}
	return ci, nil
}

func (mh *messageHandler) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	if ev.Created {
		return mh.onIncidentCreated(ctx, ev.IncidentId)
	}
	return nil
}

func (mh *messageHandler) onIncidentCreated(ctx context.Context, id uuid.UUID) error {
	ci, ciErr := mh.lookupConfiguredIntegration(ctx)
	if ciErr != nil || ci == nil {
		log.Debug().Err(ciErr).Msg("google integration not configured")
		return ciErr
	}
	if ci.isVideoConferenceEnabled() {
		cmd := &cmdCreateIncidentVideoConference{IncidentId: id}
		if cmdErr := mh.services.Messages.SendCommand(ctx, cmd); cmdErr != nil {
			return fmt.Errorf("failed to send command: %w", cmdErr)
		}
	}
	return nil
}

type cmdCreateIncidentVideoConference struct {
	IncidentId uuid.UUID `json:"incident_id"`
}

func (mh *messageHandler) createIncidentVideoConference(ctx context.Context, cmd *cmdCreateIncidentVideoConference) error {
	ci, ciErr := mh.lookupConfiguredIntegration(ctx)
	if ciErr != nil || ci == nil {
		return ciErr
	}
	ms, msErr := ci.VideoConferenceIntegration(ctx)
	if msErr != nil {
		return fmt.Errorf("getting configured integration: %w", msErr)
	}
	inc, incErr := mh.services.Incidents.Get(ctx, cmd.IncidentId)
	if incErr != nil {
		return fmt.Errorf("get incident: %w", incErr)
	}
	return ms.CreateIncidentVideoConference(ctx, inc)
}
