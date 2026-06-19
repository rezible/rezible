package db

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/projections"
)

const (
	assertionAlertDefinitionObserved = "alert_definition_observed"
	assertionAlertRelatedEntity      = "alert_related_entity"
	knowledgeEntityKindAlert         = "alert"
)

func (s *AlertService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) error {
	if projections.SubjectKindAlert.Matches(event) {
		observed, validationErr := projections.DecodeAlertEvent(event)
		if validationErr != nil || observed == nil {
			return fmt.Errorf("invalid event: %w", validationErr)
		}
		return s.handleAlertEventProjection(ctx, observed)
	}
	return nil
}

func (s *AlertService) handleAlertEventProjection(ctx context.Context, ae *projections.AlertEvent) error {
	attrs := ae.Attributes
	projKnowledgeEntity := rez.ProjectedKnowledgeEntity{
		EvidenceAssertion: assertionAlertDefinitionObserved,
		Kind:              knowledgeEntityKindAlert,
		DisplayName:       attrs.Title,
		AliasRefs: []ent.KnowledgeEntityAliasRef{
			{Provider: ae.Event.Provider, ProviderSubjectRef: ae.Event.ProviderSubjectRef},
		},
	}
	keId, saveKnowledgeErr := s.knowledge.ResolveProjectedEntity(ctx, ae.Event, projKnowledgeEntity)
	if saveKnowledgeErr != nil {
		return fmt.Errorf("save projected entity: %w", saveKnowledgeErr)
	}

	// TODO: use regular alert service update flow here instead

	upsert := s.db.Client(ctx).Alert.Create().
		SetKnowledgeEntityID(keId).
		SetTitle(attrs.Title).
		SetDescription(attrs.Description).
		SetDefinition(attrs.Definition).
		OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
		UpdateNewValues()
	if _, saveErr := upsert.ID(ctx); saveErr != nil {
		return fmt.Errorf("upsert alert: %w", saveErr)
	}

	alertAlias := ae.Event.MakeEntityAliasRef()

	for _, related := range attrs.RelatedEntities {
		relatedAlias := ent.KnowledgeEntityAliasRef{
			Provider:           ae.Event.Provider,
			ProviderSubjectRef: related.ExternalRef,
		}
		projRelatedEnt := rez.ProjectedKnowledgeEntity{
			EvidenceAssertion: assertionSystemComponentExists,
			Kind:              related.Kind,
			DisplayName:       related.DisplayName,
			Properties:        map[string]any{"external_ref": related.ExternalRef},
			AliasRefs:         []ent.KnowledgeEntityAliasRef{relatedAlias},
			IsPlaceholder:     true,
		}
		if _, entErr := s.knowledge.ResolveProjectedEntity(ctx, ae.Event, projRelatedEnt); entErr != nil {
			return fmt.Errorf("resolve related entity: %w", entErr)
		}
		projRelatedRel := rez.ProjectedKnowledgeRelationship{
			Kind:              relationshipKindRelatedTo,
			EvidenceAssertion: assertionAlertRelatedEntity,
			DisplayName:       "alert related to " + related.DisplayName,
			Properties: map[string]any{
				"related_external_ref": related.ExternalRef,
			},
			FromAliasRef: alertAlias,
			ToAliasRef:   relatedAlias,
		}
		if _, relErr := s.knowledge.ResolveProjectedRelationship(ctx, ae.Event, projRelatedRel); relErr != nil {
			return fmt.Errorf("resolve related relationship: %w", relErr)
		}
	}

	return nil
}
