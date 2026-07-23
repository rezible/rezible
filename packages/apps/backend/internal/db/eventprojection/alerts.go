package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindAlert                  = "alert"
	knowledgeEntityKindAlertInstance          = "alert_instance"
	knowledgeRelationshipKindAlertRelatedTo   = "alert_related_to"
	knowledgeAssertionAlertDefinitionObserved = "alert_definition_observed"
)

func (s *ProjectionService) handleAlertInstanceEvent(ctx context.Context, event *projections.AlertInstanceEvent) ([]rez.ProjectedEntityRef, error) {
	attributes := event.Attributes

	alertEntityRef := ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindAlert,
		Reference:   attributes.ExternalRef,
		DisplayName: attributes.Title,
		Description: attributes.Description,
		Properties:  map[string]any{"definition": attributes.Definition},
	}
	alertEntityEvidence := ent.KnowledgeEvidenceRef{
		Kind:        projectionEvidenceKind(event.Event),
		Assertion:   knowledgeAssertionAlertDefinitionObserved,
		EffectiveAt: event.Event.OccurredAt,
		Properties:  alertEntityRef.Properties,
		SubjectAliasRef: ent.KnowledgeSubjectAliasRef{
			Kind:               ksa.SubjectKindEntity,
			Provider:           event.Event.Provider,
			ProviderSubjectRef: "alert_definition:" + attributes.ExternalRef,
			Description:        "Alert definition",
			SubjectEntityRef:   &alertEntityRef,
		},
	}

	var projected []rez.ProjectedEntityRef
	return projected, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		alertSubj, definitionErr := s.knowledge.IngestDomainEntityEvidence(ctx, event.Event, alertEntityEvidence)
		if definitionErr != nil {
			return fmt.Errorf("alert knowledge evidence: %w", definitionErr)
		}

		upsertAlert := tx.Alert.Create().
			SetKnowledgeEntityID(alertSubj.EntityID).
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

		//createInstance := tx.AlertInstance.Create().
		//	SetAlertID(alertID)
		//instance, saveInstanceErr := createInstance.Save(ctx)
		//if saveInstanceErr != nil {
		//	return fmt.Errorf("save alert instance: %w", saveInstanceErr)
		//}
		//projected = append(projected, rez.ProjectedEntityRef{Kind: knowledgeEntityKindAlertInstance, Id: instance.ID})

		relationshipEvidence := make([]ent.KnowledgeEvidenceRef, 0, len(attributes.RelatedEntities)*2)

		for _, related := range attributes.RelatedEntities {
			componentRef := systemComponentRef(related.Kind, related.ExternalRef, related.DisplayName, "", nil)
			componentAlias := ent.KnowledgeSubjectAliasRef{
				Kind:               ksa.SubjectKindEntity,
				Provider:           event.Event.Provider,
				ProviderSubjectRef: related.ExternalRef,
				Description:        "Alert related component",
				SubjectEntityRef:   &componentRef,
			}

			relationshipAlias := ent.KnowledgeSubjectAliasRef{
				Kind:               ksa.SubjectKindRelationship,
				Provider:           event.Event.Provider,
				ProviderSubjectRef: fmt.Sprintf("%s:alert_related_to:%s", event.Event.ProviderSubjectRef, related.ExternalRef),
				Description:        "Alert definition relates to component",
				SubjectRelationshipRef: &ent.KnowledgeRelationshipRef{
					Kind:       knowledgeRelationshipKindAlertRelatedTo,
					EntityRefs: [2]ent.KnowledgeEntityRef{alertEntityRef, componentRef},
				},
			}

			relationshipEvidence = append(relationshipEvidence,
				ent.KnowledgeEvidenceRef{
					Kind:            projectionEvidenceKind(event.Event),
					Assertion:       knowledgeAssertionSystemComponentExists,
					EffectiveAt:     event.Event.OccurredAt,
					Properties:      componentRef.Properties,
					SubjectAliasRef: componentAlias,
				},
				ent.KnowledgeEvidenceRef{
					Kind:            projectionEvidenceKind(event.Event),
					Assertion:       "alert_relates_to",
					EffectiveAt:     event.Event.OccurredAt,
					SubjectAliasRef: relationshipAlias,
				},
			)
		}

		if ingestErr := s.knowledge.IngestEvidence(ctx, event.Event, relationshipEvidence...); ingestErr != nil {
			return fmt.Errorf("ingest alert relationships: %w", ingestErr)
		}

		return nil
	})
}
