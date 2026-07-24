package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindAlert                  = "alert"
	knowledgeRelationshipKindAlertRelatedTo   = "alert_related_to"
	knowledgeAssertionAlertDefinitionObserved = "alert_definition_observed"
)

func (s *ProjectionService) handleAlertInstanceEvent(ctx context.Context, event *projections.AlertInstanceEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	alertEntityRef := ent.KnowledgeEntityRef{
		Kind:  knowledgeEntityKindAlert,
		Alias: event.Event.KnowledgeAliasRef(),
	}
	alertEntityEvidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionAlertDefinitionObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeEvidenceSubjectState{
			DisplayName: attributes.Title,
			Description: attributes.Description,
			Properties: map[string]any{
				"definition":  attributes.Definition,
				"external_id": attributes.ExternalRef,
			},
		},
		SubjectEntity: &alertEntityRef,
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		alertKe, definitionErr := s.knowledge.IngestEntityEvidence(ctx, event.Event, alertEntityEvidence)
		if definitionErr != nil {
			return fmt.Errorf("alert knowledge evidence: %w", definitionErr)
		}

		upsertAlert := tx.Alert.Create().
			SetKnowledgeEntity(alertKe).
			SetTitle(attributes.Title).
			SetDescription(attributes.Description).
			SetDefinition(attributes.Definition).
			OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
			UpdateNewValues()
		alertID, alertErr := upsertAlert.ID(ctx)
		if alertErr != nil {
			return fmt.Errorf("upsert alert: %w", alertErr)
		}
		projected = append(projected, rez.ProjectedEntityRef{Kind: knowledgeEntityKindAlert, Id: alertID})

		return nil
	})
}
