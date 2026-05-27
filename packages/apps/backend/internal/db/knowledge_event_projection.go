package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations/projections"
)

const (
	evidenceSourceNormalizedEventProjection = "normalized_event_projection"

	assertionCodeRepositoryExists        = "code_repository_exists"
	assertionCodeChangeObserved          = "code_change_observed"
	assertionCodeChangeTouchedRepository = "code_change_touched_repository"
	assertionSystemComponentExists       = "system_component_exists"
	assertionSystemRelationshipExists    = "system_relationship_exists"

	knowledgeKindCodeRepository = "code_repository"
	knowledgeKindCodeChange     = "code_change"
	relationshipKindTouched     = "touched_repository"
)

func HandleKnowledgeEntityEventProjection(ctx context.Context, client *ent.Client, ev *ent.NormalizedEvent) error {
	proj := newKnowledgeEntityEventProjector(ev, client)

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
	event      *ent.NormalizedEvent
	observedAt time.Time
	knowledge  *KnowledgeService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, client *ent.Client) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{
		event:      ev,
		observedAt: observedAtForEvent(ev),
		knowledge:  newKnowledgeService(client),
	}
}

func observedAtForEvent(ev *ent.NormalizedEvent) time.Time {
	observedAt := ev.OccurredAt
	if observedAt.IsZero() {
		if !ev.ReceivedAt.IsZero() {
			observedAt = ev.ReceivedAt
		} else {
			observedAt = time.Now()
		}
	}
	return observedAt
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

type EntityAliasRef struct {
	Provider           string
	ProviderSubjectRef string
}

func (kp *knowledgeEntityEventProjector) saveProjection(ctx context.Context, result *KnowledgeProjection) error {
	savedAliasRefLookup := make(map[EntityAliasRef]uuid.UUID)
	for _, projEntity := range result.Entities {
		saved, saveErr := kp.saveProjectedEntity(ctx, projEntity)
		if saveErr != nil {
			return fmt.Errorf("save projected entity: %w", saveErr)
		}
		if evidenceErr := kp.addEntityEvidence(ctx, saved, projEntity.Assertion); evidenceErr != nil {
			return fmt.Errorf("add entity evidence: %w", evidenceErr)
		}
		for _, aliasRef := range projEntity.Aliases {
			savedAliasRefLookup[aliasRef] = saved.Entity.ID
		}
	}
	for _, projRel := range result.Relationships {
		if saveErr := kp.saveProjectedRelationship(ctx, projRel, savedAliasRefLookup); saveErr != nil {
			return fmt.Errorf("save projected relationship: %w", saveErr)
		}
	}
	return nil
}

func (kp *knowledgeEntityEventProjector) saveProjectedEntity(ctx context.Context, proj ProjectedKnowledgeEntity) (*ResolvedKnowledgeEntity, error) {
	return kp.knowledge.ResolveEntityByAliases(ctx, ResolveKnowledgeEntityParams{
		Event:         kp.event,
		Kind:          proj.Kind,
		DisplayName:   proj.DisplayName,
		Description:   proj.Description,
		Aliases:       proj.Aliases,
		IsPlaceholder: proj.IsPlaceholder,
	})
}

func (kp *knowledgeEntityEventProjector) addEntityEvidence(ctx context.Context, saved *ResolvedKnowledgeEntity, assertion string) error {
	return kp.knowledge.RecordEntityEvidence(ctx, RecordKnowledgeEntityEvidenceParams{
		Event:        kp.event,
		EntityID:     saved.Entity.ID,
		Aliases:      saved.Aliases,
		EvidenceKind: saved.EvidenceKind,
		Assertion:    assertion,
	})
}

func (kp *knowledgeEntityEventProjector) saveProjectedRelationship(ctx context.Context, rel ProjectedKnowledgeRelationship, refLookup map[EntityAliasRef]uuid.UUID) error {
	var resolveEntityErr error
	fromId, fromAliasWasSaved := refLookup[rel.FromAlias]
	if !fromAliasWasSaved || fromId == uuid.Nil {
		fromId, resolveEntityErr = kp.knowledge.resolveExistingEntityIDByAliases(ctx, rel.FromAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("FromAlias: %w", resolveEntityErr)
		}
	}
	toId, toAliasWasSaved := refLookup[rel.ToAlias]
	if !toAliasWasSaved || toId == uuid.Nil {
		toId, resolveEntityErr = kp.knowledge.resolveExistingEntityIDByAliases(ctx, rel.ToAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("ToAlias: %w", resolveEntityErr)
		}
	}
	if fromId == uuid.Nil || toId == uuid.Nil {
		return fmt.Errorf("alias resolved to nil entity id")
	}

	relationshipParams := ResolveKnowledgeRelationshipParams{
		Kind:           rel.Kind,
		DisplayName:    rel.DisplayName,
		Description:    rel.Description,
		SourceEntityID: fromId,
		TargetEntityID: toId,
		Attributes:     rel.Attributes,
		ObservedAt:     kp.observedAt,
	}
	savedRel, saveErr := kp.knowledge.ResolveRelationship(ctx, relationshipParams)
	if saveErr != nil {
		return fmt.Errorf("resolve knowledge relationship: %w", saveErr)
	}

	evidenceParams := RecordKnowledgeRelationshipEvidenceParams{
		RelationshipID:    savedRel.Relationship.ID,
		NormalizedEventID: kp.event.ID,
		Assertion:         rel.Assertion,
		EvidenceKind:      savedRel.EvidenceKind,
		ObservedAt:        kp.observedAt,
		Attributes:        rel.Attributes,
	}
	return kp.knowledge.RecordRelationshipEvidence(ctx, evidenceParams)
}

// Event projections

type KnowledgeProjection struct {
	Entities      []ProjectedKnowledgeEntity
	Relationships []ProjectedKnowledgeRelationship
}

type ProjectedKnowledgeEntity struct {
	IsPlaceholder bool
	Kind          string
	Assertion     string
	DisplayName   string
	Description   string
	Attributes    map[string]any
	Aliases       []EntityAliasRef
}

type ProjectedKnowledgeRelationship struct {
	Kind        string
	Assertion   string
	DisplayName string
	Description string
	Attributes  map[string]any
	FromAlias   EntityAliasRef
	ToAlias     EntityAliasRef
}

func (kp *knowledgeEntityEventProjector) makeEntityRef(ev *ent.NormalizedEvent, ProviderSubjectRef string) EntityAliasRef {
	return entityAliasRefForEvent(ev, ProviderSubjectRef)
}

func entityAliasRefForEvent(ev *ent.NormalizedEvent, providerSubjectRef string) EntityAliasRef {
	if providerSubjectRef == "" {
		providerSubjectRef = ev.ProviderSubjectRef
	}
	return EntityAliasRef{Provider: ev.Provider, ProviderSubjectRef: providerSubjectRef}
}

func (kp *knowledgeEntityEventProjector) projectCodeForgeEvent(pe *projections.CodeForgeEvent) *KnowledgeProjection {
	repoEntity := ProjectedKnowledgeEntity{
		Kind:        knowledgeKindCodeRepository,
		Assertion:   assertionCodeRepositoryExists,
		DisplayName: pe.Attributes.DisplayName,
		Attributes:  pe.Event.Attributes,
		Aliases:     []EntityAliasRef{kp.makeEntityRef(pe.Event, "")},
	}
	return &KnowledgeProjection{Entities: []ProjectedKnowledgeEntity{repoEntity}}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEvent(pe *projections.CodeChangeEvent) *KnowledgeProjection {
	attrs := pe.Attributes
	changeEventAlias := kp.makeEntityRef(pe.Event, "")
	codeChangedEntity := ProjectedKnowledgeEntity{
		Kind:        knowledgeKindCodeChange,
		Assertion:   assertionCodeChangeObserved,
		DisplayName: attrs.DisplayName,
		Attributes:  pe.Event.Attributes,
		Aliases:     []EntityAliasRef{changeEventAlias},
	}

	entities := []ProjectedKnowledgeEntity{codeChangedEntity}
	relationships := make([]ProjectedKnowledgeRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := kp.makeEntityRef(pe.Event, attrs.RepositoryExternalRef)
		entities = append(entities, ProjectedKnowledgeEntity{
			Kind:          knowledgeKindCodeRepository,
			Assertion:     assertionCodeRepositoryExists,
			DisplayName:   attrs.RepositoryExternalRef,
			Attributes:    map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Aliases:       []EntityAliasRef{repoAlias},
			IsPlaceholder: true,
		})

		repoChangeRelationship := ProjectedKnowledgeRelationship{
			Kind:        relationshipKindTouched,
			Assertion:   assertionCodeChangeTouchedRepository,
			DisplayName: "code change touched repository",
			Attributes: map[string]any{
				"repository_external_ref": attrs.RepositoryExternalRef,
			},
			FromAlias: changeEventAlias,
			ToAlias:   repoAlias,
		}
		relationships = append(relationships, repoChangeRelationship)
	}

	return &KnowledgeProjection{Entities: entities, Relationships: relationships}
}

func (kp *knowledgeEntityEventProjector) projectSystemComponentEvent(pe *projections.SystemComponentEvent) *KnowledgeProjection {
	attrs := pe.Attributes
	componentEntity := ProjectedKnowledgeEntity{
		Assertion:   assertionSystemComponentExists,
		Kind:        attrs.Kind,
		DisplayName: attrs.DisplayName,
		Description: attrs.Description,
		Attributes:  attrs.Properties,
		Aliases:     []EntityAliasRef{kp.makeEntityRef(pe.Event, attrs.ExternalRef)},
	}
	return &KnowledgeProjection{
		Entities: []ProjectedKnowledgeEntity{componentEntity},
	}
}

func (kp *knowledgeEntityEventProjector) projectSystemRelationshipEvent(pe *projections.SystemRelationshipEvent) *KnowledgeProjection {
	attrs := pe.Attributes

	sourceAlias := kp.makeEntityRef(pe.Event, attrs.SourceExternalRef)
	sourceEntity := ProjectedKnowledgeEntity{
		IsPlaceholder: true,
		Kind:          attrs.SourceKind,
		Assertion:     assertionSystemComponentExists,
		DisplayName:   attrs.SourceDisplayName,
		Attributes:    map[string]any{"external_ref": attrs.SourceExternalRef},
		Aliases:       []EntityAliasRef{sourceAlias},
	}

	targetAlias := kp.makeEntityRef(pe.Event, attrs.TargetExternalRef)
	targetEntity := ProjectedKnowledgeEntity{
		IsPlaceholder: true,
		Kind:          attrs.TargetKind,
		Assertion:     assertionSystemComponentExists,
		DisplayName:   attrs.TargetDisplayName,
		Attributes:    map[string]any{"external_ref": attrs.TargetExternalRef},
		Aliases:       []EntityAliasRef{targetAlias},
	}

	relationship := ProjectedKnowledgeRelationship{
		Kind:        attrs.Kind,
		Assertion:   assertionSystemRelationshipExists,
		DisplayName: attrs.DisplayName,
		Description: attrs.Description,
		Attributes:  attrs.Properties,
		FromAlias:   sourceAlias,
		ToAlias:     targetAlias,
	}

	return &KnowledgeProjection{
		Entities:      []ProjectedKnowledgeEntity{sourceEntity, targetEntity},
		Relationships: []ProjectedKnowledgeRelationship{relationship},
	}
}
