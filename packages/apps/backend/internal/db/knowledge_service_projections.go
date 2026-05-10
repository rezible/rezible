package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	ke "github.com/rezible/rezible/ent/knowledgeentity"
	knfh "github.com/rezible/rezible/ent/knowledgefacthistory"
	"github.com/rezible/rezible/internal/projections"
)

func KnowledgeEntityEventProjectionHandler(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	projectionEvent, validationErr := projections.DecodeEvent(event)
	if validationErr != nil || projectionEvent == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	proj := newKnowledgeEntityEventProjector(event, newKnowledgeService(client))
	var result *ProjectionResult
	switch ev := projectionEvent.(type) {
	case projections.RepositoryObserved:
		result = proj.projectRepositoryObserved(ev)
	case projections.ChangeEventObserved:
		result = proj.projectCodeChangeEventObserved(ev)
	}
	if result != nil {
		return proj.saveProjectionResult(ctx, result)
	}
	return nil
}

type knowledgeEntityEventProjector struct {
	event *ent.NormalizedEvent
	ks    *KnowledgeService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, ks *KnowledgeService) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{event: ev, ks: ks}
}

func (kp *knowledgeEntityEventProjector) saveProjectionResult(ctx context.Context, result *ProjectionResult) error {
	savedAliasRefLookup := make(map[EntityAliasRef]uuid.UUID)
	for i, projEnt := range result.Entities {
		if projEnt.Entity == nil {
			return fmt.Errorf("nil entity in projection result (index %d)", i)
		}
		refs := make([]EntityAliasRef, len(projEnt.Entity.Edges.Aliases))
		for ai, alias := range projEnt.Entity.Edges.Aliases {
			if validErr := kp.validateProjectedAliasFields(alias); validErr != nil {
				return fmt.Errorf("invalid alias for projected %s entity: %w", projEnt.Entity.Kind, validErr)
			}
			refs[ai] = EntityAliasRef{
				Provider:       alias.Provider,
				ProviderSource: alias.ProviderSource,
				SubjectKind:    alias.SubjectKind,
				SubjectRef:     alias.SubjectRef,
			}
		}
		saved, saveErr := kp.saveProjectedEntity(ctx, projEnt, refs)
		if saveErr != nil {
			return fmt.Errorf("saving projected entity: %w", saveErr)
		}
		for _, ref := range refs {
			savedAliasRefLookup[ref] = saved.ID
		}
	}
	for _, projRel := range result.Relationships {
		if projRel.Relationship == nil {
			return fmt.Errorf("nil relationship in projeciton")
		}

		var resolveEntityErr error
		fromId, fromAliasWasSaved := savedAliasRefLookup[projRel.FromAlias]
		if !fromAliasWasSaved || fromId == uuid.Nil {
			fromId, resolveEntityErr = kp.resolveExistingProjectedEntityID(ctx, projRel.FromAlias)
			if resolveEntityErr != nil {
				return fmt.Errorf("FromAlias: %w", resolveEntityErr)
			}
		}

		toId, toAliasWasSaved := savedAliasRefLookup[projRel.ToAlias]
		if !toAliasWasSaved || toId == uuid.Nil {
			toId, resolveEntityErr = kp.resolveExistingProjectedEntityID(ctx, projRel.ToAlias)
			if resolveEntityErr != nil {
				return fmt.Errorf("ToAlias: %w", resolveEntityErr)
			}
		}

		if saveErr := kp.saveProjectedRelationship(ctx, projRel.Relationship, fromId, toId); saveErr != nil {
			return fmt.Errorf("save projected relationship: %w", saveErr)
		}
	}
	return nil
}

func (kp *knowledgeEntityEventProjector) validateProjectedAliasFields(alias *ent.KnowledgeEntityAlias) error {
	if alias == nil {
		return fmt.Errorf("projected entity alias is nil")
	}
	if alias.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if alias.ProviderSource == "" {
		return fmt.Errorf("provider_source is required")
	}
	if alias.SubjectKind == "" {
		return fmt.Errorf("subject_kind is required")
	}
	if alias.SubjectRef == "" {
		return fmt.Errorf("subject_ref is required")
	}
	return nil
}

func (kp *knowledgeEntityEventProjector) resolveExistingProjectedEntityID(ctx context.Context, refs ...EntityAliasRef) (id uuid.UUID, err error) {
	// definitely a better way to do this
	for _, ref := range refs {
		alias, queryErr := kp.ks.lookupEntityAliasRef(ctx, ref)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			err = fmt.Errorf("failed to lookup entity alias: %w", queryErr)
			break
		}
		if alias == nil {
			continue
		} else if id == uuid.Nil {
			id = alias.EntityID
		} else if alias.EntityID != id {
			err = fmt.Errorf(
				"projected entity aliases resolve to different entities: %s resolves to %s, expected %s",
				ref.SubjectRef,
				alias.EntityID,
				id,
			)
			break
		}
	}
	return id, err
}

func (kp *knowledgeEntityEventProjector) saveProjectedEntity(ctx context.Context, projEnt ProjectedEntity, aliasRefs []EntityAliasRef) (*ent.KnowledgeEntity, error) {
	// needed as knowledge entities do not have a stable identifier
	existingId, lookupErr := kp.resolveExistingProjectedEntityID(ctx, aliasRefs...)
	if lookupErr != nil {
		return nil, fmt.Errorf("failed to resolve existing projected entity: %w", lookupErr)
	}

	var existing *ent.KnowledgeEntity
	if existingId != uuid.Nil {
		var existingErr error
		existing, existingErr = kp.ks.GetEntity(ctx, ke.ID(existingId))
		if existingErr != nil {
			return nil, fmt.Errorf("query existing projected entity: %w", existingErr)
		}
	}

	pe := projEnt.Entity
	setEntity := func(m *ent.KnowledgeEntityMutation) {
		if existing != nil {
			m.SetKind(existing.Kind)
			m.SetDisplayName(existing.DisplayName)
			m.SetDescription(existing.Description)
			m.SetProperties(projEnt.mergeProperties(existing.Properties))
		} else {
			m.SetKind(pe.Kind)
			m.SetDisplayName(pe.DisplayName)
			m.SetDescription(pe.Description)
			m.SetProperties(pe.Properties)
		}
	}
	savedEntity, entityErr := kp.ks.SetEntity(ctx, existingId, setEntity)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}
	for _, alias := range pe.Edges.Aliases {
		setEntityAlias := func(m *ent.KnowledgeEntityAliasMutation) {
			m.SetEntityID(savedEntity.ID)
			m.SetProvider(alias.Provider)
			m.SetProviderSource(alias.ProviderSource)
			m.SetSubjectKind(alias.SubjectKind)
			m.SetSubjectRef(alias.SubjectRef)
			m.SetNormalizedEventID(kp.event.ID)
			m.SetFirstSeenAt(kp.event.OccurredAt)
			m.SetLastSeenAt(kp.event.OccurredAt)
		}
		savedAlias, aliasErr := kp.ks.SetEntityAlias(ctx, alias.ID, setEntityAlias)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}

		setAliasFactProvenance := func(m *ent.KnowledgeFactProvenanceMutation) {
			m.SetAliasID(savedAlias.ID)
			m.SetExtractionMethod("normalized_event_projection")
			kp.setEventProjectedFactProvenanceFields(m)
		}
		_, provErr := kp.ks.SetFactProvenance(ctx, uuid.Nil, setAliasFactProvenance)
		if provErr != nil {
			return nil, fmt.Errorf("record alias provenance: %w", provErr)
		}

		factHistoryAttrs := map[string]any{
			"normalized_kind": kp.event.Kind.String(),
			"entity_id":       savedEntity.ID.String(),
			"entity_kind":     savedEntity.Kind,
			"display_name":    savedEntity.DisplayName,
			"subject_kind":    savedAlias.SubjectKind,
			"subject_ref":     savedAlias.SubjectRef,
		}
		setFactHistory := func(m *ent.KnowledgeFactHistoryMutation) {
			m.SetFactKind(knfh.FactKindAlias)
			m.SetAliasID(savedAlias.ID)
			m.SetNormalizedEventID(kp.event.ID)
			m.SetEventKind("alias_observed")
			m.SetHistoryKey(fmt.Sprintf("knowledge-alias:%s:%s", savedAlias.ID, kp.event.ID))
			m.SetOccurredAt(kp.event.OccurredAt)
			m.SetProvider(kp.event.Provider)
			m.SetProviderSource(kp.event.ProviderSource)
			m.SetProviderEventRef(kp.event.ProviderEventRef)
			m.SetExtractionMethod("normalized_event_projection")
			m.SetAttributes(factHistoryAttrs)
		}
		_, histErr := kp.ks.SetFactHistory(ctx, uuid.Nil, setFactHistory)
		if histErr != nil {
			return nil, fmt.Errorf("record alias history: %w", histErr)
		}
	}
	return savedEntity, nil
}

func (kp *knowledgeEntityEventProjector) setEventProjectedFactProvenanceFields(m *ent.KnowledgeFactProvenanceMutation) {
	m.SetExtractionMethod("normalized_event_projection")
	m.SetNormalizedEventID(kp.event.ID)
	m.SetProvider(kp.event.Provider)
	m.SetProviderSource(kp.event.ProviderSource)
	m.SetProviderEventRef(kp.event.ProviderEventRef)
	m.SetFirstSeenAt(kp.event.OccurredAt)
	m.SetLastSeenAt(kp.event.OccurredAt)
}

func (kp *knowledgeEntityEventProjector) saveProjectedRelationship(ctx context.Context, rel *ent.KnowledgeRelationship, fromId, toId uuid.UUID) error {
	setRelationshipFn := func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(fromId)
		m.SetTargetEntityID(toId)
		m.SetKind(rel.Kind)
		m.SetDisplayName(rel.DisplayName)
		m.SetDescription(rel.Description)
		m.SetFirstSeenAt(kp.event.OccurredAt)
		m.SetLastSeenAt(kp.event.OccurredAt)
		if rel.Properties != nil {
			m.SetProperties(rel.Properties)
		}
	}
	savedRel, saveErr := kp.ks.SetRelationship(ctx, uuid.Nil, setRelationshipFn)
	if saveErr != nil {
		return fmt.Errorf("upsert knowledge relationship: %w", saveErr)
	}

	setFactProvenanceFn := func(m *ent.KnowledgeFactProvenanceMutation) {
		m.SetRelationshipID(savedRel.ID)
		kp.setEventProjectedFactProvenanceFields(m)
	}
	_, provErr := kp.ks.SetFactProvenance(ctx, uuid.Nil, setFactProvenanceFn)
	if provErr != nil {
		return fmt.Errorf("record relationship provenance: %w", provErr)
	}

	setFactHistoryFn := func(m *ent.KnowledgeFactHistoryMutation) {
		m.SetFactKind(knfh.FactKindRelationship)
		m.SetRelationshipID(savedRel.ID)
		m.SetNormalizedEventID(kp.event.ID)
		m.SetEventKind("relationship_observed")
		m.SetHistoryKey(fmt.Sprintf("knowledge-relationship:%s:%s", savedRel.ID, kp.event.ID))
		m.SetOccurredAt(kp.event.OccurredAt)
		m.SetProvider(kp.event.Provider)
		m.SetProviderSource(kp.event.ProviderSource)
		m.SetProviderEventRef(kp.event.ProviderEventRef)
		m.SetExtractionMethod("normalized_event_projection")
		m.SetAttributes(map[string]any{
			"kind":           savedRel.Kind,
			"from_entity_id": fromId.String(),
			"to_entity_id":   toId.String(),
		})
	}
	_, histErr := kp.ks.SetFactHistory(ctx, uuid.Nil, setFactHistoryFn)
	if histErr != nil {
		return fmt.Errorf("record relationship history: %w", histErr)
	}
	return nil
}

// Event projections

type ProjectionResult struct {
	Entities      []ProjectedEntity
	Relationships []ProjectedRelationship
}

type ProjectedEntity struct {
	Entity        *ent.KnowledgeEntity
	IsPlaceholder bool
}

func (pe ProjectedEntity) mergeProperties(existing map[string]any) map[string]any {
	pp := pe.Entity.Properties
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

type ProjectedRelationship struct {
	Relationship *ent.KnowledgeRelationship
	FromAlias    EntityAliasRef
	ToAlias      EntityAliasRef
}

type EntityAliasRef struct {
	Provider       string
	ProviderSource string
	SubjectKind    string
	SubjectRef     string
}

func (r EntityAliasRef) makeAlias() *ent.KnowledgeEntityAlias {
	return &ent.KnowledgeEntityAlias{
		Provider:       r.Provider,
		ProviderSource: r.ProviderSource,
		SubjectKind:    r.SubjectKind,
		SubjectRef:     r.SubjectRef,
	}
}

func (kp *knowledgeEntityEventProjector) projectRepositoryObserved(pe projections.RepositoryObserved) *ProjectionResult {
	repoAlias := EntityAliasRef{
		Provider:       pe.Event.Provider,
		ProviderSource: pe.Event.ProviderSource,
		SubjectKind:    "repository",
		SubjectRef:     pe.Event.SubjectRef,
	}
	repoEntity := &ent.KnowledgeEntity{
		Kind:        "repository",
		DisplayName: pe.Attributes.DisplayName,
		Properties:  pe.Event.Attributes,
		Edges: ent.KnowledgeEntityEdges{
			Aliases: []*ent.KnowledgeEntityAlias{repoAlias.makeAlias()},
		},
	}
	return &ProjectionResult{Entities: []ProjectedEntity{{Entity: repoEntity}}}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEventObserved(event projections.ChangeEventObserved) *ProjectionResult {
	ev := event.Event
	attrs := event.Attributes
	changeAlias := EntityAliasRef{
		Provider:       ev.Provider,
		ProviderSource: ev.ProviderSource,
		SubjectKind:    "change_event",
		SubjectRef:     ev.SubjectRef,
	}
	changeEventEntity := &ent.KnowledgeEntity{
		Kind:        "change_event",
		DisplayName: attrs.DisplayName,
		Properties:  ev.Attributes,
		Edges: ent.KnowledgeEntityEdges{
			Aliases: []*ent.KnowledgeEntityAlias{changeAlias.makeAlias()},
		},
	}

	entities := []ProjectedEntity{{Entity: changeEventEntity}}
	relationships := make([]ProjectedRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := EntityAliasRef{
			Provider:       ev.Provider,
			ProviderSource: ev.ProviderSource,
			SubjectKind:    "repository",
			SubjectRef:     attrs.RepositoryExternalRef,
		}
		repoEntity := &ent.KnowledgeEntity{
			Kind:        "repository",
			DisplayName: attrs.RepositoryExternalRef,
			Properties:  map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Edges: ent.KnowledgeEntityEdges{
				Aliases: []*ent.KnowledgeEntityAlias{repoAlias.makeAlias()},
			},
		}
		entities = append(entities, ProjectedEntity{
			Entity:        repoEntity,
			IsPlaceholder: true,
		})

		repoChangeRelationship := ProjectedRelationship{
			Relationship: &ent.KnowledgeRelationship{
				Kind:        "changes_repository",
				DisplayName: "changes repository",
				Properties: map[string]any{
					"repository_external_ref": attrs.RepositoryExternalRef,
				},
			},
			FromAlias: changeAlias,
			ToAlias:   repoAlias,
		}
		relationships = append(relationships, repoChangeRelationship)
	}

	return &ProjectionResult{Entities: entities, Relationships: relationships}
}
