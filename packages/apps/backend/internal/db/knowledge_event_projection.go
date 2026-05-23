package db

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
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

func handleKnowledgeEntityEventProjection(ctx context.Context, client *ent.Client, ev *ent.NormalizedEvent) error {
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
	observedAt := ev.OccurredAt
	if observedAt.IsZero() {
		if !ev.ReceivedAt.IsZero() {
			observedAt = ev.ReceivedAt
		} else {
			observedAt = time.Now()
		}
	}
	return &knowledgeEntityEventProjector{
		event:      ev,
		observedAt: observedAt,
		knowledge:  newKnowledgeService(client),
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
		if evidenceErr := kp.addEntityEvidence(ctx, saved); evidenceErr != nil {
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

func (kp *knowledgeEntityEventProjector) resolveExistingProjectedEntityID(ctx context.Context, refs ...EntityAliasRef) (uuid.UUID, error) {
	// definitely a better way to do this
	var id uuid.UUID
	for _, ref := range refs {
		alias, queryErr := kp.knowledge.lookupEntityAliasRef(ctx, ref)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return uuid.Nil, fmt.Errorf("failed to lookup entity alias: %w", queryErr)
		}
		if alias == nil {
			continue
		}
		if id == uuid.Nil {
			id = alias.EntityID
			continue
		}
		if id != alias.EntityID {
			return uuid.Nil, fmt.Errorf("projected aliases resolve to different entities: %s -> %s (expected %s)",
				ref.ProviderSubjectRef, alias.EntityID, id)
		}
	}
	return id, nil
}

type savedProjectedKnowledgeEntity struct {
	Entity        *ent.KnowledgeEntity
	Aliases       []*ent.KnowledgeEntityAlias
	EvidenceKind  knev.EvidenceKind
	AssertionKind string
}

func (kp *knowledgeEntityEventProjector) saveProjectedEntity(ctx context.Context, proj ProjectedKnowledgeEntity) (*savedProjectedKnowledgeEntity, error) {
	// needed as knowledge entities do not have a stable identifier
	existingId, lookupErr := kp.resolveExistingProjectedEntityID(ctx, proj.Aliases...)
	if lookupErr != nil {
		return nil, fmt.Errorf("failed to resolve existing projected entity: %w", lookupErr)
	}

	var current *ent.KnowledgeEntity
	if existingId != uuid.Nil {
		var existingErr error
		current, existingErr = kp.knowledge.GetEntity(ctx, kne.ID(existingId))
		if existingErr != nil {
			return nil, fmt.Errorf("query existing projected entity: %w", existingErr)
		}
		if current.Kind != proj.Kind {
			return nil, fmt.Errorf("knowledge alias resolved to incompatible entity kind %q, expected %q", current.Kind, proj.Kind)
		}
	}

	mergedProperties := proj.Properties
	if mergedProperties == nil {
		mergedProperties = map[string]any{}
	}
	evidenceKind := knev.EvidenceKindObserved
	if current != nil {
		mergedProperties = proj.mergeProperties(current.Properties)
		propsChanged := !reflect.DeepEqual(current.Properties, mergedProperties)
		undeleted := current.DeletedAt != nil
		displayChanged := !proj.IsPlaceholder && (current.DisplayName != proj.DisplayName || current.Description != proj.Description)
		if propsChanged || displayChanged || undeleted {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	setEntity := func(m *ent.KnowledgeEntityMutation) {
		isFresh := current == nil || !proj.IsPlaceholder
		if isFresh {
			m.SetKind(proj.Kind)
		}
		if isFresh || current.DisplayName == "" {
			m.SetDisplayName(proj.DisplayName)
		}
		if isFresh || current.Description == "" {
			m.SetDescription(proj.Description)
		}
		if isFresh || current.FirstObservedAt == nil {
			m.SetFirstObservedAt(kp.observedAt)
		}
		m.SetProperties(mergedProperties)
		m.SetLastObservedAt(kp.observedAt)
		m.ClearDeletedAt()
	}
	savedEntity, entityErr := kp.knowledge.SetEntity(ctx, existingId, setEntity)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}

	savedAliases := make([]*ent.KnowledgeEntityAlias, 0, len(proj.Aliases))
	for _, alias := range proj.Aliases {
		existingAlias, aliasLookupErr := kp.knowledge.lookupEntityAliasRef(ctx, alias)
		if aliasLookupErr != nil && !ent.IsNotFound(aliasLookupErr) {
			return nil, fmt.Errorf("lookup existing alias: %w", aliasLookupErr)
		}
		var existingAliasId uuid.UUID
		if existingAlias != nil {
			if existingAlias.EntityID != savedEntity.ID {
				return nil, fmt.Errorf(
					"knowledge alias %s/%s already belongs to entity %s, cannot attach to entity %s",
					alias.Provider,
					alias.ProviderSubjectRef,
					existingAlias.EntityID,
					savedEntity.ID,
				)
			}
			existingAliasId = existingAlias.ID
		}

		setEntityAlias := func(m *ent.KnowledgeEntityAliasMutation) {
			m.SetEntityID(savedEntity.ID)
			m.SetProvider(alias.Provider)
			m.SetProviderSubjectRef(alias.ProviderSubjectRef)
		}
		savedAlias, aliasErr := kp.knowledge.SetEntityAlias(ctx, existingAliasId, setEntityAlias)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}

		savedAliases = append(savedAliases, savedAlias)
	}
	return &savedProjectedKnowledgeEntity{
		Entity:        savedEntity,
		Aliases:       savedAliases,
		EvidenceKind:  evidenceKind,
		AssertionKind: proj.AssertionKind,
	}, nil
}

func (kp *knowledgeEntityEventProjector) addEntityEvidence(ctx context.Context, saved *savedProjectedKnowledgeEntity) error {
	for _, alias := range saved.Aliases {
		_, evidenceErr := kp.knowledge.AddEvidence(ctx, func(m *ent.KnowledgeEvidenceMutation) {
			m.SetSubjectType(knev.SubjectTypeEntity)
			m.SetEntityID(saved.Entity.ID)
			m.SetAliasID(alias.ID)
			m.SetNormalizedEventID(kp.event.ID)
			m.SetAssertionKind(saved.AssertionKind)
			m.SetEvidenceKind(saved.EvidenceKind)
			m.SetObservedAt(kp.observedAt)
			m.SetSource(evidenceSourceNormalizedEventProjection)
			m.SetProperties(saved.Entity.Properties)
		})
		if evidenceErr != nil {
			return fmt.Errorf("entity alias evidence: %w", evidenceErr)
		}
	}
	return nil
}

func (kp *knowledgeEntityEventProjector) saveProjectedRelationship(ctx context.Context, rel ProjectedKnowledgeRelationship, refLookup map[EntityAliasRef]uuid.UUID) error {
	var resolveEntityErr error
	fromId, fromAliasWasSaved := refLookup[rel.FromAlias]
	if !fromAliasWasSaved || fromId == uuid.Nil {
		fromId, resolveEntityErr = kp.resolveExistingProjectedEntityID(ctx, rel.FromAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("FromAlias: %w", resolveEntityErr)
		}
	}
	toId, toAliasWasSaved := refLookup[rel.ToAlias]
	if !toAliasWasSaved || toId == uuid.Nil {
		toId, resolveEntityErr = kp.resolveExistingProjectedEntityID(ctx, rel.ToAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("ToAlias: %w", resolveEntityErr)
		}
	}
	if fromId == uuid.Nil || toId == uuid.Nil {
		return fmt.Errorf("alias resolved to nil entity id")
	}

	lookupExistingPred := knr.And(knr.Kind(rel.Kind), knr.SourceEntityID(fromId), knr.TargetEntityID(toId))
	existing, existingErr := kp.knowledge.GetRelationship(ctx, lookupExistingPred)
	if existingErr != nil && !ent.IsNotFound(existingErr) {
		return fmt.Errorf("query existing relationship: %w", existingErr)
	}

	evidenceKind := knev.EvidenceKindObserved
	var existingId uuid.UUID
	if existing != nil {
		existingId = existing.ID
		if !reflect.DeepEqual(existing.Properties, rel.Properties) {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	setRelationshipFn := func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(fromId)
		m.SetTargetEntityID(toId)
		m.SetKind(rel.Kind)
		m.SetDisplayName(rel.DisplayName)
		m.SetDescription(rel.Description)
		if existing == nil || existing.FirstObservedAt == nil {
			m.SetFirstObservedAt(kp.observedAt)
		}
		m.SetLastObservedAt(kp.observedAt)
		m.ClearDeletedAt()
		if rel.Properties != nil {
			m.SetProperties(rel.Properties)
		}
	}
	savedRel, saveErr := kp.knowledge.SetRelationship(ctx, existingId, setRelationshipFn)
	if saveErr != nil {
		return fmt.Errorf("upsert knowledge relationship: %w", saveErr)
	}

	addRelationshipEvidenceFn := func(m *ent.KnowledgeEvidenceMutation) {
		m.SetSubjectType(knev.SubjectTypeRelationship)
		m.SetRelationshipID(savedRel.ID)
		m.SetNormalizedEventID(kp.event.ID)
		m.SetAssertionKind(rel.AssertionKind)
		m.SetEvidenceKind(evidenceKind)
		m.SetObservedAt(kp.observedAt)
		m.SetSource(evidenceSourceNormalizedEventProjection)
		m.SetProperties(rel.Properties)
	}
	_, evidenceErr := kp.knowledge.AddEvidence(ctx, addRelationshipEvidenceFn)
	if evidenceErr != nil {
		return fmt.Errorf("record relationship evidence: %w", evidenceErr)
	}
	return nil
}

// Event projections

type KnowledgeProjection struct {
	Entities      []ProjectedKnowledgeEntity
	Relationships []ProjectedKnowledgeRelationship
}

type ProjectedKnowledgeEntity struct {
	Kind          string
	AssertionKind string
	DisplayName   string
	Description   string
	Properties    map[string]any
	Aliases       []EntityAliasRef
	IsPlaceholder bool
}

func (pe ProjectedKnowledgeEntity) mergeProperties(existing map[string]any) map[string]any {
	pp := pe.Properties
	merged := make(map[string]any, len(existing)+len(pp))
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range pp {
		if _, exists := merged[k]; exists && pe.IsPlaceholder {
			continue
		}
		merged[k] = v
	}
	return merged
}

type ProjectedKnowledgeRelationship struct {
	Kind          string
	AssertionKind string
	DisplayName   string
	Description   string
	Properties    map[string]any
	FromAlias     EntityAliasRef
	ToAlias       EntityAliasRef
}

func (kp *knowledgeEntityEventProjector) makeEntityRef(ev *ent.NormalizedEvent, ProviderSubjectRef string) EntityAliasRef {
	if ProviderSubjectRef == "" {
		ProviderSubjectRef = ev.ProviderSubjectRef
	}
	return EntityAliasRef{Provider: ev.Provider, ProviderSubjectRef: ProviderSubjectRef}
}

func (kp *knowledgeEntityEventProjector) projectCodeForgeEvent(pe *projections.CodeForgeEvent) *KnowledgeProjection {
	repoEntity := ProjectedKnowledgeEntity{
		Kind:          knowledgeKindCodeRepository,
		AssertionKind: assertionCodeRepositoryExists,
		DisplayName:   pe.Attributes.DisplayName,
		Properties:    pe.Event.Attributes,
		Aliases:       []EntityAliasRef{kp.makeEntityRef(pe.Event, "")},
	}
	return &KnowledgeProjection{Entities: []ProjectedKnowledgeEntity{repoEntity}}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEvent(pe *projections.CodeChangeEvent) *KnowledgeProjection {
	attrs := pe.Attributes
	changeEventAlias := kp.makeEntityRef(pe.Event, "")
	codeChangedEntity := ProjectedKnowledgeEntity{
		Kind:          knowledgeKindCodeChange,
		AssertionKind: assertionCodeChangeObserved,
		DisplayName:   attrs.DisplayName,
		Properties:    pe.Event.Attributes,
		Aliases:       []EntityAliasRef{changeEventAlias},
	}

	entities := []ProjectedKnowledgeEntity{codeChangedEntity}
	relationships := make([]ProjectedKnowledgeRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := kp.makeEntityRef(pe.Event, attrs.RepositoryExternalRef)
		entities = append(entities, ProjectedKnowledgeEntity{
			Kind:          knowledgeKindCodeRepository,
			AssertionKind: assertionCodeRepositoryExists,
			DisplayName:   attrs.RepositoryExternalRef,
			Properties:    map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Aliases:       []EntityAliasRef{repoAlias},
			IsPlaceholder: true,
		})

		repoChangeRelationship := ProjectedKnowledgeRelationship{
			Kind:          relationshipKindTouched,
			AssertionKind: assertionCodeChangeTouchedRepository,
			DisplayName:   "code change touched repository",
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
		AssertionKind: assertionSystemComponentExists,
		Kind:          attrs.Kind,
		DisplayName:   attrs.DisplayName,
		Description:   attrs.Description,
		Properties:    attrs.Properties,
		Aliases:       []EntityAliasRef{kp.makeEntityRef(pe.Event, attrs.ExternalRef)},
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
		AssertionKind: assertionSystemComponentExists,
		DisplayName:   attrs.SourceDisplayName,
		Properties:    map[string]any{"external_ref": attrs.SourceExternalRef},
		Aliases:       []EntityAliasRef{sourceAlias},
	}

	targetAlias := kp.makeEntityRef(pe.Event, attrs.TargetExternalRef)
	targetEntity := ProjectedKnowledgeEntity{
		IsPlaceholder: true,
		Kind:          attrs.TargetKind,
		AssertionKind: assertionSystemComponentExists,
		DisplayName:   attrs.TargetDisplayName,
		Properties:    map[string]any{"external_ref": attrs.TargetExternalRef},
		Aliases:       []EntityAliasRef{targetAlias},
	}

	relationship := ProjectedKnowledgeRelationship{
		Kind:          attrs.Kind,
		AssertionKind: assertionSystemRelationshipExists,
		DisplayName:   attrs.DisplayName,
		Description:   attrs.Description,
		Properties:    attrs.Properties,
		FromAlias:     sourceAlias,
		ToAlias:       targetAlias,
	}

	return &KnowledgeProjection{
		Entities:      []ProjectedKnowledgeEntity{sourceEntity, targetEntity},
		Relationships: []ProjectedKnowledgeRelationship{relationship},
	}
}
