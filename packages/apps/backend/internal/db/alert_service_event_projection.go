package db

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/integrations/projections"
)

const (
	assertionAlertDefinitionObserved = "alert_definition_observed"
	knowledgeKindAlert               = "alert"
)

func (s *AlertService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) error {
	if !projections.SubjectKindAlert.Matches(event) {
		return nil
	}
	observed, validationErr := projections.DecodeAlertEvent(event)
	if validationErr != nil || observed == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}

	attrs := observed.Attributes
	entityParams := rez.ResolveKnowledgeEntityParams{
		Event:             event,
		EvidenceAssertion: assertionAlertDefinitionObserved,
		Entity: &ent.KnowledgeEntity{
			Kind:        knowledgeKindAlert,
			DisplayName: attrs.Title,
		},
		Aliases: eventKnowledgeEntityAliases(event),
	}
	knowledgeEntity, saveKnowledgeErr := s.knowledge.ResolveEntity(ctx, entityParams)
	if saveKnowledgeErr != nil {
		return fmt.Errorf("save projected entity: %w", saveKnowledgeErr)
	}

	upsert := s.client.Alert.Create().
		SetKnowledgeEntityID(knowledgeEntity.ID).
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
