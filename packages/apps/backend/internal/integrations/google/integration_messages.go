package google

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/pkg/messages"
)

type eventHandler struct {
	intg *Integration
}

func (i *Integration) MessageHandlers() []rez.MessageEventHandler {
	mh := &eventHandler{intg: i}
	return []rez.MessageEventHandler{
		messages.NewEventHandler("Google.OnIncidentUpdate", mh.onIncidentUpdate),
	}
}

func (h *eventHandler) withInstallation(ctx context.Context, fn func(*InstalledIntegration) error) error {
	rows, listErr := h.intg.installations.ListInstallations(ctx, integration.Name(integrationName))
	if listErr != nil {
		return fmt.Errorf("list google installations: %w", listErr)
	}
	if len(rows) == 0 {
		return nil
	}
	if len(rows) > 1 {
		return fmt.Errorf("found multiple installations for integration %q", integrationName)
	}
	ii, installedErr := h.intg.newInstalledIntegration(rows[0])
	if installedErr != nil {
		return fmt.Errorf("load google installation: %w", installedErr)
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
