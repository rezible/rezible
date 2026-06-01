package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations/projections"
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
	event      *ent.NormalizedEvent
	observedAt time.Time
	knowledge  rez.KnowledgeService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, knowledge rez.KnowledgeService) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{
		event:      ev,
		observedAt: observedAtForEvent(ev),
		knowledge:  knowledge,
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

type entityAliasRef struct {
	Provider           string
	ProviderSubjectRef string
}

func (r entityAliasRef) asAlias() *ent.KnowledgeEntityAlias {
	return &ent.KnowledgeEntityAlias{Provider: r.Provider, ProviderSubjectRef: r.ProviderSubjectRef}
}

func (kp *knowledgeEntityEventProjector) saveProjection(ctx context.Context, result *KnowledgeProjection) error {
	aliasRefLookup := make(map[entityAliasRef]uuid.UUID)
	for _, proj := range result.Entities {
		params := rez.ResolveKnowledgeEntityParams{
			IsPlaceholder: proj.IsPlaceholder,
			Event:         kp.event,
			Entity: &ent.KnowledgeEntity{
				Kind:        proj.Kind,
				DisplayName: proj.DisplayName,
				Description: proj.Description,
				Properties:  proj.Properties,
			},
			Aliases:           make([]*ent.KnowledgeEntityAlias, len(proj.Aliases)),
			EvidenceAssertion: proj.Assertion,
		}
		for i, ref := range proj.Aliases {
			params.Aliases[i] = ref.asAlias()
		}
		saved, saveErr := kp.knowledge.ResolveEntity(ctx, params)
		if saveErr != nil {
			return fmt.Errorf("save projected entity: %w", saveErr)
		}
		for _, aliasRef := range proj.Aliases {
			aliasRefLookup[aliasRef] = saved.ID
		}
	}
	for _, projRel := range result.Relationships {
		if saveErr := kp.saveProjectedRelationship(ctx, projRel, aliasRefLookup); saveErr != nil {
			return fmt.Errorf("save projected relationship: %w", saveErr)
		}
	}
	return nil
}

func (kp *knowledgeEntityEventProjector) resolveEntityFromAlias(ctx context.Context, ref entityAliasRef) (*ent.KnowledgeEntity, error) {
	resolvedAliases, resolveAliasesErr := kp.knowledge.ResolveEntityAliases(ctx, ref.asAlias())
	if resolveAliasesErr != nil {
		return nil, fmt.Errorf("resolve existing entity aliases: %w", resolveAliasesErr)
	}
	for _, res := range resolvedAliases {
		if res.Edges.Entity != nil {
			return res.Edges.Entity, nil
		}
	}
	return nil, nil
}

func (kp *knowledgeEntityEventProjector) saveProjectedRelationship(ctx context.Context, proj ProjectedKnowledgeRelationship, refLookup map[entityAliasRef]uuid.UUID) error {
	fromId, fromAliasWasSaved := refLookup[proj.FromAlias]
	if !fromAliasWasSaved || fromId == uuid.Nil {
		fromEntity, resolveEntityErr := kp.resolveEntityFromAlias(ctx, proj.FromAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("FromAlias: %w", resolveEntityErr)
		} else if fromEntity != nil {
			fromId = fromEntity.ID
		}
	}
	toId, toAliasWasSaved := refLookup[proj.ToAlias]
	if !toAliasWasSaved || toId == uuid.Nil {
		toEntity, resolveEntityErr := kp.resolveEntityFromAlias(ctx, proj.ToAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("ToAlias: %w", resolveEntityErr)
		} else if toEntity != nil {
			toId = toEntity.ID
		}
	}

	if fromId == uuid.Nil || toId == uuid.Nil {
		return fmt.Errorf("alias resolved to nil entity id")
	}

	rel := &ent.KnowledgeRelationship{
		SourceEntityID:  fromId,
		TargetEntityID:  toId,
		Kind:            proj.Kind,
		DisplayName:     proj.DisplayName,
		Description:     proj.Description,
		FirstObservedAt: &kp.observedAt,
		Properties:      proj.Properties,
	}
	rel.Edges.Evidence = []*ent.KnowledgeEvidence{
		{
			EventID:    kp.event.ID,
			Assertion:  proj.Assertion,
			ObservedAt: kp.observedAt,
		},
	}
	_, saveErr := kp.knowledge.ResolveRelationship(ctx, rel)
	if saveErr != nil {
		return fmt.Errorf("resolve knowledge relationship: %w", saveErr)
	}
	return nil
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
	Properties    map[string]any
	Aliases       []entityAliasRef
}

type ProjectedKnowledgeRelationship struct {
	Kind        string
	Assertion   string
	DisplayName string
	Description string
	Properties  map[string]any
	FromAlias   entityAliasRef
	ToAlias     entityAliasRef
}

func (kp *knowledgeEntityEventProjector) makeEntityRef(ev *ent.NormalizedEvent, providerSubjectRef string) entityAliasRef {
	if providerSubjectRef == "" {
		providerSubjectRef = ev.ProviderSubjectRef
	}
	return entityAliasRef{Provider: ev.Provider, ProviderSubjectRef: providerSubjectRef}
}

func (kp *knowledgeEntityEventProjector) projectCodeForgeEvent(pe *projections.CodeForgeEvent) *KnowledgeProjection {
	repoEntity := ProjectedKnowledgeEntity{
		Kind:        knowledgeKindCodeRepository,
		Assertion:   assertionCodeRepositoryExists,
		DisplayName: pe.Attributes.DisplayName,
		Properties:  pe.Event.Attributes,
		Aliases:     []entityAliasRef{kp.makeEntityRef(pe.Event, "")},
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
		Properties:  pe.Event.Attributes,
		Aliases:     []entityAliasRef{changeEventAlias},
	}

	entities := []ProjectedKnowledgeEntity{codeChangedEntity}
	relationships := make([]ProjectedKnowledgeRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := kp.makeEntityRef(pe.Event, attrs.RepositoryExternalRef)
		entities = append(entities, ProjectedKnowledgeEntity{
			Kind:          knowledgeKindCodeRepository,
			Assertion:     assertionCodeRepositoryExists,
			DisplayName:   attrs.RepositoryExternalRef,
			Properties:    map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Aliases:       []entityAliasRef{repoAlias},
			IsPlaceholder: true,
		})

		repoChangeRelationship := ProjectedKnowledgeRelationship{
			Kind:        relationshipKindTouched,
			Assertion:   assertionCodeChangeTouchedRepository,
			DisplayName: "code change touched repository",
			Properties: map[string]any{
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
		Properties:  attrs.Properties,
		Aliases:     []entityAliasRef{kp.makeEntityRef(pe.Event, attrs.ExternalRef)},
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
		Properties:    map[string]any{"external_ref": attrs.SourceExternalRef},
		Aliases:       []entityAliasRef{sourceAlias},
	}

	targetAlias := kp.makeEntityRef(pe.Event, attrs.TargetExternalRef)
	targetEntity := ProjectedKnowledgeEntity{
		IsPlaceholder: true,
		Kind:          attrs.TargetKind,
		Assertion:     assertionSystemComponentExists,
		DisplayName:   attrs.TargetDisplayName,
		Properties:    map[string]any{"external_ref": attrs.TargetExternalRef},
		Aliases:       []entityAliasRef{targetAlias},
	}

	relationship := ProjectedKnowledgeRelationship{
		Kind:        attrs.Kind,
		Assertion:   assertionSystemRelationshipExists,
		DisplayName: attrs.DisplayName,
		Description: attrs.Description,
		Properties:  attrs.Properties,
		FromAlias:   sourceAlias,
		ToAlias:     targetAlias,
	}

	return &KnowledgeProjection{
		Entities:      []ProjectedKnowledgeEntity{sourceEntity, targetEntity},
		Relationships: []ProjectedKnowledgeRelationship{relationship},
	}
}
