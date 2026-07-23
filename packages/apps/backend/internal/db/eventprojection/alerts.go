package eventprojection

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/ent/alertinstance"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	knowledgeEntityKindAlert         = "alert"
	knowledgeEntityKindAlertInstance = "alert_instance"

	knowledgeRelationshipKindAlertInstanceOf = "alert_instance_of"
	knowledgeRelationshipKindAlertRelatedTo  = "alert_related_to"

	knowledgeAssertionAlertDefinitionObserved = "alert_definition_observed"
	knowledgeAssertionAlertInstanceObserved   = "alert_instance_observed"
	knowledgeAssertionAlertRelatedEntity      = "alert_related_entity"
	knowledgeAssertionAlertInstanceOf         = "alert_instance_of_definition"
)

func (s *ProjectionService) handleAlertEventProjection(ctx context.Context, normalizedEvent *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
	event, decodeErr := projections.DecodeAlertInstanceEvent(normalizedEvent)
	if decodeErr != nil {
		return nil, fmt.Errorf("invalid alert event: %w", decodeErr)
	}
	attributes := event.Attributes
	definitionRef := ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindAlert,
		Reference:   attributes.ExternalRef,
		DisplayName: attributes.Title,
		Description: attributes.Description,
		Properties:  map[string]any{"definition": attributes.Definition},
	}
	definitionAlias := ent.KnowledgeSubjectAliasRef{
		Kind:               ksa.SubjectKindEntity,
		Provider:           event.Event.Provider,
		ProviderSubjectRef: "alert_definition:" + attributes.ExternalRef,
		Description:        "Alert definition",
		SubjectEntityRef:   &definitionRef,
	}
	definitionEvidence := ent.KnowledgeEvidenceRef{
		Kind:            projectionEvidenceKind(event.Event),
		Assertion:       knowledgeAssertionAlertDefinitionObserved,
		EffectiveAt:     event.Event.OccurredAt,
		Properties:      definitionRef.Properties,
		SubjectAliasRef: definitionAlias,
	}

	instanceReference := attributes.InstanceExternalRef
	if instanceReference == "" {
		instanceReference = event.Event.ProviderSubjectRef
	}
	instanceRef := ent.KnowledgeEntityRef{
		Kind:        knowledgeEntityKindAlertInstance,
		Reference:   instanceReference,
		DisplayName: attributes.Title,
		Description: attributes.Description,
	}
	instanceAlias := event.Event.MakeSubjectAliasRef(ksa.SubjectKindEntity, "Alert instance")
	instanceAlias.SubjectEntityRef = &instanceRef
	instanceEvidence := ent.KnowledgeEvidenceRef{
		Kind:            projectionEvidenceKind(event.Event),
		Assertion:       knowledgeAssertionAlertInstanceObserved,
		EffectiveAt:     event.Event.OccurredAt,
		SubjectAliasRef: instanceAlias,
	}

	var projected []rez.ProjectedEntityRef
	projectTxFn := func(txCtx context.Context, tx *ent.Client) error {
		definitionEntityID, definitionErr := s.ingestDomainEntityEvidence(txCtx, event.Event, definitionEvidence)
		if definitionErr != nil {
			return fmt.Errorf("ingest alert definition: %w", definitionErr)
		}
		instanceEntityID, instanceErr := s.ingestDomainEntityEvidence(txCtx, event.Event, instanceEvidence)
		if instanceErr != nil {
			return fmt.Errorf("ingest alert instance: %w", instanceErr)
		}

		upsertAlert := tx.Alert.Create().
			SetKnowledgeEntityID(definitionEntityID).
			SetTitle(attributes.Title).
			SetDescription(attributes.Description).
			SetDefinition(attributes.Definition).
			OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
			UpdateNewValues()
		alertID, alertErr := upsertAlert.ID(txCtx)
		if alertErr != nil {
			return fmt.Errorf("upsert alert: %w", alertErr)
		}

		upsertInstance := tx.AlertInstance.Create().
			SetAlertID(alertID).
			SetKnowledgeEntityID(instanceEntityID).
			OnConflictColumns(alertinstance.FieldTenantID, alertinstance.FieldKnowledgeEntityID).
			UpdateAlertID()
		instanceID, saveInstanceErr := upsertInstance.ID(txCtx)
		if saveInstanceErr != nil {
			return fmt.Errorf("upsert alert instance: %w", saveInstanceErr)
		}

		relationshipEvidence := make([]ent.KnowledgeEvidenceRef, 0, len(attributes.RelatedEntities)*2+1)
		instanceOfRef := &ent.KnowledgeRelationshipRef{
			Kind:       knowledgeRelationshipKindAlertInstanceOf,
			EntityRefs: [2]ent.KnowledgeEntityRef{instanceRef, definitionRef},
		}
		instanceOfAlias := ent.KnowledgeSubjectAliasRef{
			Kind:                   ksa.SubjectKindRelationship,
			Provider:               event.Event.Provider,
			ProviderSubjectRef:     event.Event.ProviderSubjectRef + ":instance_of",
			Description:            "Alert instance belongs to definition",
			SubjectRelationshipRef: instanceOfRef,
		}
		relationshipEvidence = append(relationshipEvidence, ent.KnowledgeEvidenceRef{
			Kind:            projectionEvidenceKind(event.Event),
			Assertion:       knowledgeAssertionAlertInstanceOf,
			EffectiveAt:     event.Event.OccurredAt,
			SubjectAliasRef: instanceOfAlias,
		})

		for _, related := range attributes.RelatedEntities {
			componentRef := systemComponentRef(related.Kind, related.ExternalRef, related.DisplayName, "", nil)
			componentAlias := ent.KnowledgeSubjectAliasRef{
				Kind:               ksa.SubjectKindEntity,
				Provider:           event.Event.Provider,
				ProviderSubjectRef: related.ExternalRef,
				Description:        "Alert related component",
				SubjectEntityRef:   &componentRef,
			}
			relationshipRef := &ent.KnowledgeRelationshipRef{
				Kind:       knowledgeRelationshipKindAlertRelatedTo,
				EntityRefs: [2]ent.KnowledgeEntityRef{definitionRef, componentRef},
			}
			relationshipAlias := ent.KnowledgeSubjectAliasRef{
				Kind:                   ksa.SubjectKindRelationship,
				Provider:               event.Event.Provider,
				ProviderSubjectRef:     fmt.Sprintf("%s:alert_related_to:%s", event.Event.ProviderSubjectRef, related.ExternalRef),
				Description:            "Alert definition relates to component",
				SubjectRelationshipRef: relationshipRef,
			}
			relationshipEvidence = append(relationshipEvidence,
				ent.KnowledgeEvidenceRef{
					Kind:            projectionEvidenceKind(event.Event),
					Assertion:       knowledgeAssertionSystemComponentObserved,
					EffectiveAt:     event.Event.OccurredAt,
					Properties:      componentRef.Properties,
					SubjectAliasRef: componentAlias,
				},
				ent.KnowledgeEvidenceRef{
					Kind:            projectionEvidenceKind(event.Event),
					Assertion:       knowledgeAssertionAlertRelatedEntity,
					EffectiveAt:     event.Event.OccurredAt,
					SubjectAliasRef: relationshipAlias,
				},
			)
		}
		if ingestErr := s.ingestProjectedEvidence(txCtx, event.Event, relationshipEvidence...); ingestErr != nil {
			return fmt.Errorf("ingest alert relationships: %w", ingestErr)
		}

		projected = append(projected, rez.ProjectedEntityRef{Kind: knowledgeEntityKindAlertInstance, Id: instanceID})
		return nil
	}
	return projected, s.db.WithTx(ctx, projectTxFn)
}
