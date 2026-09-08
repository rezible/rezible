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

func (s *ProjectionService) handleAlertInstanceEvent(ctx context.Context, e *projections.AlertInstanceEvent) ([]rez.ProjectedEntityRef, error) {
	event := e.Event
	attrs := e.Attributes

	alertResourceRef := rez.ProviderResourceRef{
		Provider:          event.Provider,
		ProviderNamespace: event.ProviderNamespace,
		ResourceRef:       event.ProviderResourceRef,
	}
	alertEntityRef := rez.KnowledgeEntityRef{
		Category:            kne.CategorySignal,
		Kind:                knowledgeEntityKindAlert,
		ProviderResourceRef: alertResourceRef,
	}
	alertEntityEvidence := rez.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event),
		Assertion:   knowledgeAssertionAlertDefinitionObserved,
		EffectiveAt: event.OccurredAt,
		SubjectState: schematypes.KnowledgeGraphSubjectState{
			DisplayName: attrs.Title,
			Description: attrs.Description,
			Properties: map[string]any{
				"definition": attrs.Definition,
			},
		},
		SubjectEntity: &alertEntityRef,
	}

	evidenceKind := projectionEvidenceKind(event)
	supportingEvidence := make([]rez.KnowledgeEvidenceRef, 0, len(attrs.ObservedEntities)*2)
	for _, observed := range projections.SortEntityObservations(attrs.ObservedEntities) {
		observedEntityRef := rez.KnowledgeEntityRef{
			Category:            observed.Category,
			Kind:                observed.Kind,
			ProviderResourceRef: observed.Ref,
		}
		entityEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionAlertEntityObserved,
			EffectiveAt: event.OccurredAt,
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
			EffectiveAt:         event.OccurredAt,
			SubjectRelationship: &relationshipRef,
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: "Observes " + observed.DisplayName,
			},
		}
		supportingEvidence = append(supportingEvidence, entityEvidence, relationshipEvidence)
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, event, alertEntityEvidence, supportingEvidence...)
		if ingestErr != nil {
			return fmt.Errorf("alert knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return fmt.Errorf("nil subject entity")
		}

		upsertDefinition := tx.AlertDefinition.Create().
			SetKnowledgeEntityID(*subj.EntityID).
			SetTitle(attrs.Title).
			SetDescription(attrs.Description).
			SetDefinition(attrs.Definition).
			OnConflictColumns(ad.FieldTenantID, ad.FieldKnowledgeEntityID).
			UpdateNewValues()
		definitionId, alertErr := upsertDefinition.ID(ctx)
		if alertErr != nil {
			return fmt.Errorf("upsert alert: %w", alertErr)
		}

		if _, eventErr := s.alerts.RecordAlertDefinitionInstance(ctx, definitionId, event); eventErr != nil {
			return fmt.Errorf("contribute alert event: %w", eventErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{
			Kind: knowledgeEntityKindAlert,
			Id:   definitionId,
		})

		return nil
	})
}
