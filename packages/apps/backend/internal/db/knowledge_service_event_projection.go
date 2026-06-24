package db

import (
	"context"
	"fmt"
	"slices"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	assertionCodeRepositoryExists        = "code_repository_exists"
	assertionCodeChangeObserved          = "code_change_observed"
	assertionCodeChangeTouchedRepository = "code_change_touched_repository"
	assertionCodeChangeRelatedEntity     = "code_change_related_entity"
	assertionChatMessageObserved         = "chat_message_observed"
	assertionChatMessageRelatedEntity    = "chat_message_related_entity"
	assertionSystemComponentExists       = "system_component_exists"
	assertionSystemRelationshipExists    = "system_relationship_exists"

	knowledgeKindCodeRepository = "code_repository"
	knowledgeKindCodeChange     = "code_change"
	knowledgeKindChatMessage    = "chat_message"
	relationshipKindTouched     = "touched_repository"
	relationshipKindRelatedTo   = "related_to"
)

func (s *KnowledgeService) HandleEventProjection(ctx context.Context, ev *ent.NormalizedEvent) error {
	proj := newKnowledgeEntityEventProjector(ev, s)

	result, eventErr := proj.projectEvent(ev)
	if eventErr != nil {
		return fmt.Errorf("project event: %w", eventErr)
	}
	if result != nil {
		if saveErr := proj.saveProjection(ctx, result); saveErr != nil {
			return fmt.Errorf("save projection result: %w", saveErr)
		}
	}
	return nil
}

type knowledgeEntityEventProjector struct {
	event     *ent.NormalizedEvent
	knowledge rez.KnowledgeService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, knowledge rez.KnowledgeService) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{
		event:     ev,
		knowledge: knowledge,
	}
}

func (kp *knowledgeEntityEventProjector) projectEvent(ev *ent.NormalizedEvent) (*KnowledgeProjection, error) {
	var decErr error
	switch projections.SubjectKind(ev.SubjectKind) {
	case projections.SubjectKindChatMessage:
		{
			var obs *projections.ChatMessage
			if obs, decErr = projections.DecodeChatMessageEvent(ev); decErr == nil {
				return kp.projectChatMessageEvent(obs), decErr
			}
		}
	case projections.SubjectKindCodeForge:
		{
			var obs *projections.CodeForgeEvent
			if obs, decErr = projections.DecodeCodeForgeEvent(ev); decErr == nil {
				return kp.projectCodeForgeEvent(obs), decErr
			}
		}
	case projections.SubjectKindCodeChange:
		{
			var obs *projections.CodeChangeEvent
			if obs, decErr = projections.DecodeCodeChangeEvent(ev); decErr == nil {
				return kp.projectCodeChangeEvent(obs), decErr
			}
		}
	case projections.SubjectKindSystemComponent:
		{
			var obs *projections.SystemComponentEvent
			if obs, decErr = projections.DecodeSystemComponentEvent(ev); decErr == nil {
				return kp.projectSystemComponentEvent(obs), decErr
			}
		}
	case projections.SubjectKindSystemRelationship:
		{
			var obs *projections.SystemRelationshipEvent
			if obs, decErr = projections.DecodeSystemRelationshipEvent(ev); decErr == nil {
				return kp.projectSystemRelationshipEvent(obs), decErr
			}
		}
	}
	return nil, decErr
}

func (kp *knowledgeEntityEventProjector) saveProjection(ctx context.Context, result *KnowledgeProjection) error {
	for _, projEnt := range sortProjectedKnowledgeEntities(result.Entities) {
		_, saveErr := kp.knowledge.ResolveProjectedEntity(ctx, kp.event, projEnt)
		if saveErr != nil {
			return fmt.Errorf("save projected entity: %w", saveErr)
		}
	}

	for _, projRel := range sortProjectedKnowledgeRelationships(result.Relationships) {
		_, saveErr := kp.knowledge.ResolveProjectedRelationship(ctx, kp.event, projRel)
		if saveErr != nil {
			return fmt.Errorf("save projected relationship: %w", saveErr)
		}
	}
	return nil
}

func sortProjectedKnowledgeEntities(entities []rez.ProjectedKnowledgeEntity) []rez.ProjectedKnowledgeEntity {
	sortedEntities := append([]rez.ProjectedKnowledgeEntity(nil), entities...)
	slices.SortStableFunc(sortedEntities, func(left, right rez.ProjectedKnowledgeEntity) int {
		leftKey := knowledgeEntitySortKey(left)
		rightKey := knowledgeEntitySortKey(right)
		if leftKey != rightKey {
			return strings.Compare(leftKey, rightKey)
		}
		if left.Kind != right.Kind {
			return strings.Compare(left.Kind, right.Kind)
		}
		return strings.Compare(left.DisplayName, right.DisplayName)
	})
	return sortedEntities
}

func sortProjectedKnowledgeRelationships(relationships []rez.ProjectedKnowledgeRelationship) []rez.ProjectedKnowledgeRelationship {
	sortedRelationships := append([]rez.ProjectedKnowledgeRelationship(nil), relationships...)
	slices.SortStableFunc(sortedRelationships, func(left, right rez.ProjectedKnowledgeRelationship) int {
		leftKey := left.FromAliasRef.SortKey() + "\x1f" + left.ToAliasRef.SortKey()
		rightKey := right.FromAliasRef.SortKey() + "\x1f" + right.ToAliasRef.SortKey()
		if leftKey != rightKey {
			return strings.Compare(leftKey, rightKey)
		}
		return strings.Compare(left.Kind, right.Kind)
	})
	return sortedRelationships
}

// Event projections

type KnowledgeProjection struct {
	Entities      []rez.ProjectedKnowledgeEntity
	Relationships []rez.ProjectedKnowledgeRelationship
}

func (kp *knowledgeEntityEventProjector) makeExternalEntityAliasRef(ev *ent.NormalizedEvent, externalRef string) ent.KnowledgeEntityAliasRef {
	ref := ev.MakeEntityAliasRef()
	if externalRef != "" {
		ref.ProviderSubjectRef = externalRef
	}
	return ref
}

func (kp *knowledgeEntityEventProjector) projectCodeForgeEvent(pe *projections.CodeForgeEvent) *KnowledgeProjection {
	repoEntity := rez.ProjectedKnowledgeEntity{
		Kind:              knowledgeKindCodeRepository,
		EvidenceAssertion: assertionCodeRepositoryExists,
		DisplayName:       pe.Attributes.DisplayName,
		Properties:        pe.Event.Attributes,
		AliasRefs:         []ent.KnowledgeEntityAliasRef{pe.Event.MakeEntityAliasRef()},
	}
	return &KnowledgeProjection{
		Entities: []rez.ProjectedKnowledgeEntity{repoEntity},
	}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEvent(pe *projections.CodeChangeEvent) *KnowledgeProjection {
	attrs := pe.Attributes
	changeEventAliasRef := pe.Event.MakeEntityAliasRef()
	codeChangedEntity := rez.ProjectedKnowledgeEntity{
		Kind:              knowledgeKindCodeChange,
		EvidenceAssertion: assertionCodeChangeObserved,
		DisplayName:       attrs.DisplayName,
		Properties:        pe.Event.Attributes,
		AliasRefs:         []ent.KnowledgeEntityAliasRef{changeEventAliasRef},
	}

	entities := []rez.ProjectedKnowledgeEntity{codeChangedEntity}
	relationships := make([]rez.ProjectedKnowledgeRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := kp.makeExternalEntityAliasRef(pe.Event, attrs.RepositoryExternalRef)
		entities = append(entities, rez.ProjectedKnowledgeEntity{
			Kind:              knowledgeKindCodeRepository,
			EvidenceAssertion: assertionCodeRepositoryExists,
			DisplayName:       attrs.RepositoryExternalRef,
			Properties:        map[string]any{"external_ref": attrs.RepositoryExternalRef},
			AliasRefs:         []ent.KnowledgeEntityAliasRef{repoAlias},
			IsPlaceholder:     true,
		})

		repoChangeRelationship := rez.ProjectedKnowledgeRelationship{
			Kind:              relationshipKindTouched,
			EvidenceAssertion: assertionCodeChangeTouchedRepository,
			DisplayName:       "code change touched repository",
			Properties: map[string]any{
				"repository_external_ref": attrs.RepositoryExternalRef,
			},
			FromAliasRef: changeEventAliasRef,
			ToAliasRef:   repoAlias,
		}
		relationships = append(relationships, repoChangeRelationship)
	}
	for _, related := range attrs.RelatedEntities {
		relatedAlias := kp.makeExternalEntityAliasRef(pe.Event, related.ExternalRef)
		entities = append(entities, rez.ProjectedKnowledgeEntity{
			Kind:              related.Kind,
			EvidenceAssertion: assertionSystemComponentExists,
			DisplayName:       related.DisplayName,
			Properties:        map[string]any{"external_ref": related.ExternalRef},
			AliasRefs:         []ent.KnowledgeEntityAliasRef{relatedAlias},
			IsPlaceholder:     true,
		})
		relationships = append(relationships, rez.ProjectedKnowledgeRelationship{
			Kind:              relationshipKindRelatedTo,
			EvidenceAssertion: assertionCodeChangeRelatedEntity,
			DisplayName:       "code change related to " + related.DisplayName,
			Properties:        map[string]any{"related_external_ref": related.ExternalRef},
			FromAliasRef:      changeEventAliasRef,
			ToAliasRef:        relatedAlias,
		})
	}

	return &KnowledgeProjection{
		Entities:      entities,
		Relationships: relationships,
	}
}

func (kp *knowledgeEntityEventProjector) projectChatMessageEvent(pe *projections.ChatMessage) *KnowledgeProjection {
	attrs := pe.Attributes
	messageAlias := pe.Event.MakeEntityAliasRef()
	messageEntity := rez.ProjectedKnowledgeEntity{
		Kind:              knowledgeKindChatMessage,
		EvidenceAssertion: assertionChatMessageObserved,
		DisplayName:       attrs.Body,
		Properties:        pe.Event.Attributes,
		AliasRefs:         []ent.KnowledgeEntityAliasRef{messageAlias},
	}
	entities := []rez.ProjectedKnowledgeEntity{messageEntity}
	relationships := make([]rez.ProjectedKnowledgeRelationship, 0, len(attrs.RelatedEntities))
	for _, related := range attrs.RelatedEntities {
		relatedAlias := kp.makeExternalEntityAliasRef(pe.Event, related.ExternalRef)
		entities = append(entities, rez.ProjectedKnowledgeEntity{
			Kind:              related.Kind,
			EvidenceAssertion: assertionSystemComponentExists,
			DisplayName:       related.DisplayName,
			Properties:        map[string]any{"external_ref": related.ExternalRef},
			AliasRefs:         []ent.KnowledgeEntityAliasRef{relatedAlias},
			IsPlaceholder:     true,
		})
		relationships = append(relationships, rez.ProjectedKnowledgeRelationship{
			Kind:              relationshipKindRelatedTo,
			EvidenceAssertion: assertionChatMessageRelatedEntity,
			DisplayName:       "chat message related to " + related.DisplayName,
			Properties: map[string]any{
				"conversation_external_ref": attrs.ConversationExternalRef,
				"sender_external_ref":       attrs.SenderExternalRef,
				"related_external_ref":      related.ExternalRef,
			},
			FromAliasRef: messageAlias,
			ToAliasRef:   relatedAlias,
		})
	}
	return &KnowledgeProjection{
		Entities:      entities,
		Relationships: relationships,
	}
}

func (kp *knowledgeEntityEventProjector) projectSystemComponentEvent(pe *projections.SystemComponentEvent) *KnowledgeProjection {
	attrs := pe.Attributes
	componentEntity := rez.ProjectedKnowledgeEntity{
		Kind:              attrs.Kind,
		DisplayName:       attrs.DisplayName,
		Description:       attrs.Description,
		Properties:        attrs.Properties,
		EvidenceAssertion: assertionSystemComponentExists,
		AliasRefs:         []ent.KnowledgeEntityAliasRef{kp.makeExternalEntityAliasRef(pe.Event, attrs.ExternalRef)},
	}
	return &KnowledgeProjection{
		Entities: []rez.ProjectedKnowledgeEntity{componentEntity},
	}
}

func (kp *knowledgeEntityEventProjector) projectSystemRelationshipEvent(pe *projections.SystemRelationshipEvent) *KnowledgeProjection {
	attrs := pe.Attributes

	sourceAliasRef := kp.makeExternalEntityAliasRef(pe.Event, attrs.SourceExternalRef)
	sourceEntity := rez.ProjectedKnowledgeEntity{
		IsPlaceholder:     true,
		EvidenceAssertion: assertionSystemComponentExists,
		Kind:              attrs.SourceKind,
		DisplayName:       attrs.SourceDisplayName,
		Properties:        map[string]any{"external_ref": attrs.SourceExternalRef},
		AliasRefs:         []ent.KnowledgeEntityAliasRef{sourceAliasRef},
	}

	targetAliasRef := kp.makeExternalEntityAliasRef(pe.Event, attrs.TargetExternalRef)
	targetEntity := rez.ProjectedKnowledgeEntity{
		IsPlaceholder:     true,
		EvidenceAssertion: assertionSystemComponentExists,
		Kind:              attrs.TargetKind,
		DisplayName:       attrs.TargetDisplayName,
		Properties:        map[string]any{"external_ref": attrs.TargetExternalRef},
		AliasRefs:         []ent.KnowledgeEntityAliasRef{targetAliasRef},
	}

	relationship := rez.ProjectedKnowledgeRelationship{
		EvidenceAssertion: assertionSystemRelationshipExists,
		Kind:              attrs.Kind,
		DisplayName:       attrs.DisplayName,
		Description:       attrs.Description,
		Properties:        attrs.Properties,
		FromAliasRef:      sourceAliasRef,
		ToAliasRef:        targetAliasRef,
	}

	return &KnowledgeProjection{
		Entities:      []rez.ProjectedKnowledgeEntity{sourceEntity, targetEntity},
		Relationships: []rez.ProjectedKnowledgeRelationship{relationship},
	}
}
