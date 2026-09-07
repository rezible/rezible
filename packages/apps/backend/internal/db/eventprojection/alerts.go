package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ad "github.com/rezible/rezible/ent/alertdefinition"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeAssertionAlertDefinitionObserved = "alert_definition_observed"
	knowledgeAssertionAlertEntityObserved     = "alert_related_entity_observed"
	knowledgeAssertionAlertObservesEntity     = "alert_observes_entity"
)

func (s *ProjectionService) handleAlertInstanceEvent(ctx context.Context, event *projections.AlertInstanceEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	alertResourceRef := rez.ProviderResourceRef{
		Provider:          event.Event.Provider,
		ProviderNamespace: event.Event.ProviderNamespace,
		ResourceRef:       event.Event.ProviderResourceRef,
	}
	alertEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategorySignal,
		Kind:                knowledgeEntityKindAlert,
		ProviderResourceRef: alertResourceRef,
	}
	alertEntityEvidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionAlertDefinitionObserved,
		EffectiveAt: event.Event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attributes.Title,
			Description: attributes.Description,
			Properties: map[string]any{
				"definition": attributes.Definition,
			},
		},
		SubjectEntity: &alertEntityRef,
	}

	evidenceKind := projectionEvidenceKind(event.Event)
	supportingEvidence := make([]rez.KnowledgeEvidenceRef, 0, len(attributes.ObservedEntities)*2)
	for _, observed := range projections.SortEntityObservations(attributes.ObservedEntities) {
		observedEntityRef := rez.KnowledgeEntityRef{
			Category:            observed.Category,
			Kind:                observed.Kind,
			ProviderResourceRef: observed.Ref,
		}
		entityEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionAlertEntityObserved,
			EffectiveAt: event.Event.OccurredAt,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: observed.DisplayName,
				Description: observed.Description,
				Properties:  observed.Properties,
			},
			SubjectEntity: &observedEntityRef,
		}
		relationshipRef := rez.KnowledgeRelationshipRef{
			Predicate:           knr.PredicateObserves,
			ProviderResourceRef: projections.DerivedRelationshipRef(knr.PredicateObserves, alertResourceRef, observed.Ref),
			Source:              alertEntityRef,
			Target:              observedEntityRef,
		}
		relationshipEvidence := rez.KnowledgeEvidenceRef{
			Kind:                evidenceKind,
			Assertion:           knowledgeAssertionAlertObservesEntity,
			EffectiveAt:         event.Event.OccurredAt,
			SubjectRelationship: &relationshipRef,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: "Observes " + observed.DisplayName,
			},
		}
		supportingEvidence = append(supportingEvidence, entityEvidence, relationshipEvidence)
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, event.Event, alertEntityEvidence, supportingEvidence...)
		if ingestErr != nil {
			return fmt.Errorf("alert knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return fmt.Errorf("nil subject entity")
		}

		upsertAlert := tx.AlertDefinition.Create().
			SetKnowledgeEntityID(*subj.EntityID).
			SetTitle(attributes.Title).
			SetDescription(attributes.Description).
			SetDefinition(attributes.Definition).
			OnConflictColumns(ad.FieldTenantID, ad.FieldKnowledgeEntityID).
			UpdateNewValues()
		alertID, alertErr := upsertAlert.ID(ctx)
		if alertErr != nil {
			return fmt.Errorf("upsert alert: %w", alertErr)
		}

		if _, eventErr := s.alerts.RecordAlertEvent(ctx, alertID, event.Event); eventErr != nil {
			return fmt.Errorf("contribute alert event: %w", eventErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntityKindAlert,
			Id:   alertID,
		})

		return nil
	})
}
