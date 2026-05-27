package db

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/integrations/projections"
)

const (
	assertionAlertDefinitionObserved = "alert_definition_observed"
	knowledgeKindAlert               = "alert"
)

func HandleAlertEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	if projections.SubjectKindAlert.Matches(event) {
		observed, validationErr := projections.DecodeAlertEvent(event)
		if validationErr != nil || observed == nil {
			return fmt.Errorf("invalid event: %w", validationErr)
		}
		ap := &alertEventProjectionHandler{
			client:    client,
			observed:  observed,
			knowledge: newKnowledgeService(client),
		}
		return ap.handle(ctx)
	}

	return nil
}

type alertEventProjectionHandler struct {
	client    *ent.Client
	observed  *projections.AlertEvent
	knowledge *KnowledgeService
}

func (h *alertEventProjectionHandler) handle(ctx context.Context) error {
	attrs := h.observed.Attributes
	ev := h.observed.Event
	params := ResolveKnowledgeEntityParams{
		Event:       ev,
		Kind:        knowledgeKindAlert,
		DisplayName: attrs.Title,
		Aliases:     []EntityAliasRef{entityAliasRefForEvent(ev, "")},
	}
	savedKnowledge, saveKnowledgeErr := h.knowledge.ResolveEntityWithAssertion(ctx, params, assertionAlertDefinitionObserved)
	if saveKnowledgeErr != nil {
		return fmt.Errorf("save projected entity: %w", saveKnowledgeErr)
	}

	upsert := h.client.Alert.Create().
		SetKnowledgeEntityID(savedKnowledge.Entity.ID).
		SetTitle(attrs.Title).
		SetDescription(attrs.Description).
		SetDefinition(attrs.Definition).
		OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
		UpdateNewValues()
	if _, saveErr := upsert.ID(ctx); saveErr != nil {
		return fmt.Errorf("upsert alert: %w", saveErr)
	}

	return nil
}
