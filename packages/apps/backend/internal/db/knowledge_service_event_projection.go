package db

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/projections"
)

const (
	assertionCodeRepositoryExists        = "code_repository_exists"
	assertionCodeChangeObserved          = "code_change_observed"
	assertionCodeChangeTouchedRepository = "code_change_touched_repository"
	assertionSystemComponentExists       = "system_component_exists"
	assertionSystemRelationshipExists    = "system_relationship_exists"

	knowledgeKindCodeRepository = "code_repository"
	knowledgeKindCodeChange     = "code_change"
	relationshipKindTouched     = "touched_repository"
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
	for _, projEnt := range result.Entities {
		_, saveErr := kp.knowledge.ResolveProjectedEntity(ctx, kp.event, projEnt)
		if saveErr != nil {
			return fmt.Errorf("save projected entity: %w", saveErr)
		}
	}
	for _, projRel := range result.Relationships {
		_, saveErr := kp.knowledge.ResolveProjectedRelationship(ctx, kp.event, projRel)
		if saveErr != nil {
			return fmt.Errorf("save projected relationship: %w", saveErr)
		}
	}
	return nil
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
