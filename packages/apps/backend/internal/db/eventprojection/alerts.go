package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionAlertDefinitionObserved = "alert_definition_observed"
)

func (s *ProjectionService) handleAlertInstanceEvent(ctx context.Context, event *projections.AlertInstanceEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	alertEntityRef := ent.KnowledgeEntityRef{
		Kind:            kne.KindSignal,
		Subkind:         knowledgeEntitySubkindAlert,
		SubjectAliasRef: event.Event.KnowledgeSubjectAliasRef(),
	}
	alertEntityEvidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionAlertDefinitionObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
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
		subj, ingestErr := s.knowledge.IngestSubjectEvidence(ctx, event.Event, alertEntityEvidence)
		if ingestErr != nil {
			return fmt.Errorf("alert knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return fmt.Errorf("nil subject entity")
		}

		upsertAlert := tx.Alert.Create().
			SetKnowledgeEntityID(*subj.EntityID).
			SetTitle(attributes.Title).
			SetDescription(attributes.Description).
			SetDefinition(attributes.Definition).
			OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
			UpdateNewValues()
		alertID, alertErr := upsertAlert.ID(ctx)
		if alertErr != nil {
			return fmt.Errorf("upsert alert: %w", alertErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntitySubkindAlert,
			Id:   alertID,
		})

		return nil
	})
}
