package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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
		Subject: rez.KnowledgeSubjectRef{Entity: &alertEntityRef},
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
			Subject: rez.KnowledgeSubjectRef{Entity: &observedEntityRef},
		}
		relationshipRef := rez.KnowledgeRelationshipRef{
			Predicate:           knr.PredicateObserves,
			ProviderResourceRef: projections.DerivedRelationshipRef(knr.PredicateObserves, alertResourceRef, observed.Ref),
			Source:              alertEntityRef,
			Target:              observedEntityRef,
		}
		relationshipEvidence := rez.KnowledgeEvidenceRef{
			Kind:        evidenceKind,
			Assertion:   knowledgeAssertionAlertObservesEntity,
			EffectiveAt: event.OccurredAt,
			Subject:     rez.KnowledgeSubjectRef{Relationship: &relationshipRef},
			SubjectState: schematypes.KnowledgeGraphSubjectState{
				DisplayName: "Observes " + observed.DisplayName,
			},
		}
		supportingEvidence = append(supportingEvidence, entityEvidence, relationshipEvidence)
	}

	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, _ *ent.Client) ([]rez.ProjectedEntityRef, error) {
		subj, ingestErr := s.ingestSubjectEvidence(ctx, event, alertEntityEvidence, supportingEvidence...)
		if ingestErr != nil {
			return nil, fmt.Errorf("alert knowledge evidence: %w", ingestErr)
		} else if subj.EntityID == nil {
			return nil, fmt.Errorf("nil subject entity")
		}

		recordParams := rez.RecordAlertInstanceParams{
			Event: event,
			Definition: rez.AlertDefinitionValues{
				KnowledgeEntityID:        *subj.EntityID,
				Title:                    attrs.Title,
				Description:              attrs.Description,
				Definition:               attrs.Definition,
				ResolutionTimeoutSeconds: attrs.ResolutionTimeoutSeconds,
				IdentityGroupLabels:      attrs.IdentityGroupLabels,
			},
			Instance: rez.AlertInstanceValues{
				InstanceID: attrs.InstanceID,
				Labels:     attrs.Labels,
				Summary:    attrs.Summary,
				Severity:   schematypes.SignalSeverity(attrs.Severity),
				Firing:     attrs.State == projections.AlertStateFiring,
				StartedAt:  attrs.StartedAt,
				EndedAt:    attrs.EndedAt,
			},
		}
		definition, recordErr := s.alerts.RecordAlertInstance(ctx, recordParams)
		if recordErr != nil {
			return nil, fmt.Errorf("record alert instance: %w", recordErr)
		}

		projected := []rez.ProjectedEntityRef{{
			Kind: knowledgeEntityKindAlert,
			Id:   definition.ID,
		}}
		return projected, nil
	})
}
