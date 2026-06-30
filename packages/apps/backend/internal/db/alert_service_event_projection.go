package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionAlertDefinitionObserved = "alert_definition_observed"
	assertionAlertRelatedEntity      = "alert_related_entity"
	knowledgeEntityKindAlert         = "alert"
)

func (s *AlertService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	if projections.SubjectKindAlert.Matches(event) {
		observed, validationErr := projections.DecodeAlertEvent(event)
		if validationErr != nil || observed == nil {
			return nil, fmt.Errorf("invalid event: %w", validationErr)
		}
		return s.handleAlertEventProjection(ctx, observed)
	}
	return nil, nil
}

func (s *AlertService) handleAlertEventProjection(ctx context.Context, ae *projections.AlertEvent) (map[string][]uuid.UUID, error) {
	projIds := make(map[string][]uuid.UUID)
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
		return nil, fmt.Errorf("save projected entity: %w", saveKnowledgeErr)
	}

	// TODO: use regular alert service update flow here instead

	upsert := s.db.Client(ctx).Alert.Create().
		SetKnowledgeEntityID(keId).
		SetTitle(attrs.Title).
		SetDescription(attrs.Description).
		SetDefinition(attrs.Definition).
		OnConflictColumns(alert.FieldTenantID, alert.FieldKnowledgeEntityID).
		UpdateNewValues()
	alertId, saveErr := upsert.ID(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("upsert alert: %w", saveErr)
	}
	projIds["alert"] = append(projIds["alert"], alertId)

	alertAlias := ae.Event.MakeEntityAliasRef()

	for _, related := range projections.SortRelatedEntityRefs(attrs.RelatedEntities) {
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
			return nil, fmt.Errorf("resolve related entity: %w", entErr)
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
			return nil, fmt.Errorf("resolve related relationship: %w", relErr)
		}
	}

	return projIds, nil
}
