package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/integrations/projections"
)

const (
	assertionAlertDefinitionObserved = "alert_definition_observed"
)

func handleAlertEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	if projections.SubjectKindAlert.Matches(event) {
		observed, validationErr := projections.DecodeAlertEvent(event)
		if validationErr != nil || observed == nil {
			return fmt.Errorf("invalid event: %w", validationErr)
		}
		ap := &alertEventProjectionHandler{
			client:   client,
			observed: observed,
			kp:       newKnowledgeEntityEventProjector(event, client),
		}
		return ap.handle(ctx)
	}

	return nil
}

type alertEventProjectionHandler struct {
	client   *ent.Client
	observed *projections.AlertEvent
	kp       *knowledgeEntityEventProjector
}

func (h *alertEventProjectionHandler) handle(ctx context.Context) error {
	attrs := h.observed.Attributes
	projEntity := ProjectedKnowledgeEntity{
		Kind:          "alert_definition",
		AssertionKind: assertionAlertDefinitionObserved,
		DisplayName:   attrs.Title,
		Properties:    h.observed.Event.Attributes,
		Aliases:       []EntityAliasRef{h.kp.makeEntityRef(h.observed.Event, "")},
	}
	savedKnowledge, saveKnowledgeErr := h.kp.saveProjectedEntity(ctx, projEntity)
	if saveKnowledgeErr != nil {
		return fmt.Errorf("save projected entity: %w", saveKnowledgeErr)
	}
	if len(savedKnowledge.Aliases) == 0 {
		return fmt.Errorf("alert knowledge entity has no aliases")
	}
	if evidenceErr := h.kp.addEntityEvidence(ctx, savedKnowledge); evidenceErr != nil {
		return fmt.Errorf("add entity evidence: %w", evidenceErr)
	}

	upsert := h.client.Alert.Create().
		SetKnowledgeEntityID(savedKnowledge.Entity.ID).
		SetTitle(attrs.Title).
		SetDescription(attrs.Description).
		SetDefinition(attrs.Definition).
		OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
		UpdateNewValues()
	alertId, saveErr := upsert.ID(ctx)
	if saveErr != nil {
		return fmt.Errorf("upsert alert: %w", saveErr)
	}
	slog.Debug("saved alert", slog.String("id", alertId.String()))

	return nil
}

func (h *alertEventProjectionHandler) alertProjectionChanged(existing *ent.Alert, attrs projections.AlertSubjectAttributes) bool {
	return existing.Title != attrs.Title ||
		existing.Description != attrs.Description ||
		existing.Definition != attrs.Definition
}
