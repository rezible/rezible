package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/projections"
	"github.com/rezible/rezible/jobs"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knfh "github.com/rezible/rezible/ent/knowledgefacthistory"
)

type ProjectionResult struct {
	Entities      ent.KnowledgeEntities
	Relationships []ProjectedRelationship
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

func makeEntityAliasRef(alias *ent.KnowledgeEntityAlias) (*EntityAliasRef, error) {
	if alias == nil {
		return nil, fmt.Errorf("projected entity alias is nil")
	}
	ref := &EntityAliasRef{
		Provider:       alias.Provider,
		ProviderSource: alias.ProviderSource,
		SubjectKind:    alias.SubjectKind,
		SubjectRef:     alias.SubjectRef,
	}
	if ref.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if ref.ProviderSource == "" {
		return nil, fmt.Errorf("provider_source is required")
	}
	if ref.SubjectKind == "" {
		return nil, fmt.Errorf("subject_kind is required")
	}
	if ref.SubjectRef == "" {
		return nil, fmt.Errorf("subject_ref is required")
	}
	return ref, nil
}

func (s *KnowledgeService) HandleEventProjection(ctx context.Context, args jobs.ProjectNormalizedEvent) error {
	ev, queryErr := s.dbc.NormalizedEvent.Get(ctx, args.EventId)
	if queryErr != nil {
		return fmt.Errorf("query event: %w", queryErr)
	}

	ps := newEventProjectionProcessor(ev)
	if projErr := ps.projectAndSave(ctx, s.dbc); projErr != nil {
		return fmt.Errorf("save projection: %w", projErr)
	}
	return nil
}

type eventProjectionProcessor struct {
	event                *ent.NormalizedEvent
	aliasRefEntityLookup map[EntityAliasRef]uuid.UUID
	observedAt           time.Time
	providerEventRef     string
}

func newEventProjectionProcessor(event *ent.NormalizedEvent) *eventProjectionProcessor {
	observedAt := event.OccurredAt
	if observedAt.IsZero() {
		observedAt = event.ReceivedAt
	}

	providerEventRef := event.ProviderEventRef
	if providerEventRef == "" {
		// TODO: this should probably not be empty??
		providerEventRef = event.SubjectRef
	}

	return &eventProjectionProcessor{
		event:                event,
		aliasRefEntityLookup: make(map[EntityAliasRef]uuid.UUID),
		observedAt:           observedAt,
		providerEventRef:     providerEventRef,
	}
}

func (ps *eventProjectionProcessor) projectAndSave(ctx context.Context, db *ent.Client) error {
	projectionEvent, validationErr := projections.ValidateEvent(ps.event)
	if validationErr != nil || projectionEvent == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	var result ProjectionResult
	var projectionErr error
	switch event := projectionEvent.(type) {
	case projections.ChatMessage:
		result, projectionErr = ps.projectChatMessage(ctx, event)
	case projections.RepositoryObserved:
		result, projectionErr = ps.projectRepositoryObserved(ctx, event)
	case projections.ChangeEventObserved:
		result, projectionErr = ps.projectCodeChangeEventObserved(ctx, event)
	default:
		return fmt.Errorf("unsupported projection event %T", event)
	}

	if projectionErr != nil {
		return fmt.Errorf("projection: %w", projectionErr)
	}

	if saveErr := ps.saveProjectionResult(ctx, db, result); saveErr != nil {
		return fmt.Errorf("save projection result: %w", saveErr)
	}
	return nil
}

func (ps *eventProjectionProcessor) saveProjectionResult(ctx context.Context, dbc *ent.Client, result ProjectionResult) error {
	projEntityAliasRefs := make([][]EntityAliasRef, len(result.Entities))

	for i, projEntity := range result.Entities {
		if projEntity == nil {
			return fmt.Errorf("nil entity in projection result (index %d)", i)
		}
		if len(projEntity.Edges.Aliases) == 0 {
			return fmt.Errorf("projected %s entity missing alias", projEntity.Kind)
		}
		refs := make([]EntityAliasRef, len(projEntity.Edges.Aliases))
		for ai, alias := range projEntity.Edges.Aliases {
			ref, refErr := makeEntityAliasRef(alias)
			if refErr != nil {
				return fmt.Errorf("invalid alias for projected %s entity: %w", projEntity.Kind, refErr)
			}
			refs[ai] = *ref
		}
		projEntityAliasRefs[i] = refs
	}

	return ent.WithTx(ctx, dbc, func(tx *ent.Tx) error {
		txDb := tx.Client()
		for i, projEnt := range result.Entities {
			saved, saveErr := ps.saveProjectedEntity(ctx, txDb, projEnt, projEntityAliasRefs[i])
			if saveErr != nil {
				return fmt.Errorf("saving projected entity: %w", saveErr)
			}
			for _, ref := range projEntityAliasRefs[i] {
				ps.aliasRefEntityLookup[ref] = saved.ID
			}
		}

		for _, projRel := range result.Relationships {
			if saveErr := ps.saveProjectedRelationship(ctx, txDb, projRel); saveErr != nil {
				return fmt.Errorf("save projected relationship: %w", saveErr)
			}
		}

		return nil
	})
}

func (ps *eventProjectionProcessor) resolveExistingRelationshipEntityIDs(ctx context.Context, dbc *ent.Client, pr ProjectedRelationship) (uuid.UUID, uuid.UUID, error) {
	var resolveEntityErr error
	fromId, fromOk := ps.aliasRefEntityLookup[pr.FromAlias]
	if fromId == uuid.Nil || !fromOk {
		fromId, resolveEntityErr = ps.resolveExistingProjectedEntityID(ctx, dbc, pr.FromAlias)
		if resolveEntityErr != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("FromAlias: %w", resolveEntityErr)
		}
	}
	toId, toOk := ps.aliasRefEntityLookup[pr.ToAlias]
	if toId == uuid.Nil || !toOk {
		toId, resolveEntityErr = ps.resolveExistingProjectedEntityID(ctx, dbc, pr.ToAlias)
		if resolveEntityErr != nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("ToAlias: %w", resolveEntityErr)
		}
	}
	return fromId, toId, nil
}

func (ps *eventProjectionProcessor) resolveExistingProjectedEntityID(ctx context.Context, dbc *ent.Client, refs ...EntityAliasRef) (uuid.UUID, error) {
	// definitely a better way to do this
	entityId := uuid.Nil
	for _, ref := range refs {
		queryExisting := dbc.KnowledgeEntityAlias.Query().Where(
			knea.Provider(ref.Provider), knea.ProviderSource(ref.ProviderSource),
			knea.SubjectKind(ref.SubjectKind), knea.SubjectRef(ref.SubjectRef))
		alias, queryErr := queryExisting.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return uuid.Nil, queryErr
		}
		if alias == nil {
			continue
		}
		if entityId == uuid.Nil {
			entityId = alias.EntityID
			continue
		}
		if alias.EntityID != entityId {
			return uuid.Nil, fmt.Errorf(
				"projected entity aliases resolve to different entities: %s resolves to %s, expected %s",
				ref.SubjectRef,
				alias.EntityID,
				entityId,
			)
		}
	}
	return entityId, nil
}

func (ps *eventProjectionProcessor) saveProjectedEntity(ctx context.Context, db *ent.Client, projEnt *ent.KnowledgeEntity, aliasRefs []EntityAliasRef) (*ent.KnowledgeEntity, error) {
	// needed as knowledge entities do not have a stable identifier
	existingId, lookupErr := ps.resolveExistingProjectedEntityID(ctx, db, aliasRefs...)
	if lookupErr != nil {
		return nil, fmt.Errorf("failed to resolve existing projected entity: %w", lookupErr)
	}

	ks := NewKnowledgeService(db)

	setEntity := func(m *ent.KnowledgeEntityMutation) {
		m.SetKind(projEnt.Kind)
		m.SetDisplayName(projEnt.DisplayName)
		m.SetProperties(projEnt.Properties)
		m.SetDescription(projEnt.Description)
	}
	savedEntity, entityErr := ks.SetEntity(ctx, existingId, setEntity)
	if entityErr != nil {
		return nil, fmt.Errorf("set knowledge entity: %w", entityErr)
	}

	for _, alias := range projEnt.Edges.Aliases {
		setEntityAlias := func(m *ent.KnowledgeEntityAliasMutation) {
			m.SetEntityID(savedEntity.ID)
			m.SetProvider(alias.Provider)
			m.SetProviderSource(alias.ProviderSource)
			m.SetSubjectKind(alias.SubjectKind)
			m.SetSubjectRef(alias.SubjectRef)
			m.SetNormalizedEventID(ps.event.ID)
			m.SetFirstSeenAt(ps.observedAt)
			m.SetLastSeenAt(ps.observedAt)
		}
		savedAlias, aliasErr := ks.SetEntityAlias(ctx, alias.ID, setEntityAlias)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}

		setAliasFactProvenance := func(m *ent.KnowledgeFactProvenanceMutation) {
			ps.setEventFactProvenanceFields(m)
			m.SetAliasID(savedAlias.ID)
			m.SetExtractionMethod("normalized_event_projection")
		}
		_, provErr := ks.SetFactProvenance(ctx, uuid.Nil, setAliasFactProvenance)
		if provErr != nil {
			return nil, fmt.Errorf("record alias provenance: %w", provErr)
		}

		setFactHistory := func(m *ent.KnowledgeFactHistoryMutation) {
			m.SetFactKind(knfh.FactKindAlias)
			m.SetAliasID(savedAlias.ID)
			m.SetNormalizedEventID(ps.event.ID)
			m.SetEventKind("alias_observed")
			m.SetHistoryKey(fmt.Sprintf("knowledge-alias:%s:%s", savedAlias.ID, ps.event.ID))
			m.SetOccurredAt(ps.observedAt)
			m.SetProvider(ps.event.Provider)
			m.SetProviderSource(ps.event.ProviderSource)
			m.SetProviderEventRef(ps.providerEventRef)
			m.SetExtractionMethod("normalized_event_projection")
			m.SetAttributes(map[string]any{
				"entity_id":       savedEntity.ID.String(),
				"entity_kind":     projEnt.Kind.String(),
				"subject_kind":    alias.SubjectKind,
				"subject_ref":     alias.SubjectRef,
				"display_name":    projEnt.DisplayName,
				"normalized_kind": ps.event.Kind.String(),
			})
		}
		_, histErr := ks.SetFactHistory(ctx, uuid.Nil, setFactHistory)
		if histErr != nil {
			return nil, fmt.Errorf("record alias history: %w", histErr)
		}
	}
	return savedEntity, nil
}

func (ps *eventProjectionProcessor) saveProjectedRelationship(ctx context.Context, db *ent.Client, projRel ProjectedRelationship) error {
	rel := projRel.Relationship
	if rel == nil {
		return fmt.Errorf("projected relationship cannot be nil")
	}

	fromId, toId, lookupErr := ps.resolveExistingRelationshipEntityIDs(ctx, db, projRel)
	if lookupErr != nil {
		return fmt.Errorf("resolve existing relationship entities: %w", lookupErr)
	}

	ks := NewKnowledgeService(db)

	setRelationshipFn := func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(fromId)
		m.SetTargetEntityID(toId)
		m.SetKind(rel.Kind)
		m.SetDisplayName(rel.DisplayName)
		m.SetDescription(rel.Description)
		m.SetFirstSeenAt(ps.observedAt)
		m.SetLastSeenAt(ps.observedAt)
	}
	savedRel, saveErr := ks.SetRelationship(ctx, uuid.Nil, setRelationshipFn)
	if saveErr != nil {
		return fmt.Errorf("upsert knowledge relationship: %w", saveErr)
	}

	setRelationshipFactProvenanceFn := func(m *ent.KnowledgeFactProvenanceMutation) {
		ps.setEventFactProvenanceFields(m)
		m.SetRelationshipID(savedRel.ID)
		m.SetExtractionMethod("normalized_event_projection")
	}
	_, provErr := ks.SetFactProvenance(ctx, uuid.Nil, setRelationshipFactProvenanceFn)
	if provErr != nil {
		return fmt.Errorf("record relationship provenance: %w", provErr)
	}

	updateFactHistoryFn := func(m *ent.KnowledgeFactHistoryMutation) {
		m.SetFactKind(knfh.FactKindRelationship)
		m.SetRelationshipID(savedRel.ID)
		m.SetNormalizedEventID(ps.event.ID)
		m.SetEventKind("relationship_observed")
		m.SetHistoryKey(fmt.Sprintf("knowledge-relationship:%s:%s", savedRel.ID, ps.event.ID))
		m.SetOccurredAt(ps.observedAt)
		m.SetProvider(ps.event.Provider)
		m.SetProviderSource(ps.event.ProviderSource)
		m.SetProviderEventRef(ps.providerEventRef)
		m.SetExtractionMethod("normalized_event_projection")
		m.SetAttributes(map[string]any{
			"kind":           savedRel.Kind,
			"from_entity_id": fromId.String(),
			"to_entity_id":   toId.String(),
		})
	}
	_, histErr := ks.SetFactHistory(ctx, uuid.Nil, updateFactHistoryFn)
	if histErr != nil {
		return fmt.Errorf("record relationship history: %w", histErr)
	}
	return nil
}

func (ps *eventProjectionProcessor) setEventFactProvenanceFields(m *ent.KnowledgeFactProvenanceMutation) {
	m.SetNormalizedEventID(ps.event.ID)
	m.SetProvider(ps.event.Provider)
	m.SetProviderSource(ps.event.ProviderSource)
	m.SetProviderEventRef(ps.providerEventRef)
	m.SetFirstSeenAt(ps.observedAt)
	m.SetLastSeenAt(ps.observedAt)
}

// Event projections

func (ps *eventProjectionProcessor) projectChatMessage(ctx context.Context, pe projections.ChatMessage) (ProjectionResult, error) {
	// TODO: incident channel message
	slog.DebugContext(ctx, "projected incident chat message",
		"conversationExternalRef", pe.Attributes.ConversationExternalRef,
		"body", pe.Attributes.Body,
	)
	return ProjectionResult{}, nil
}

func (ps *eventProjectionProcessor) projectRepositoryObserved(ctx context.Context, pe projections.RepositoryObserved) (ProjectionResult, error) {
	repoEntity := &ent.KnowledgeEntity{
		Kind:        kne.KindRepository,
		DisplayName: pe.Attributes.DisplayName,
		Properties:  pe.Event.Attributes,
		Edges: ent.KnowledgeEntityEdges{
			Aliases: []*ent.KnowledgeEntityAlias{
				{
					Provider:       pe.Event.Provider,
					ProviderSource: pe.Event.ProviderSource,
					SubjectKind:    "repository",
					SubjectRef:     pe.Event.SubjectRef,
				},
			},
		},
	}
	return ProjectionResult{Entities: ent.KnowledgeEntities{repoEntity}}, nil
}

func (ps *eventProjectionProcessor) projectCodeChangeEventObserved(_ context.Context, event projections.ChangeEventObserved) (ProjectionResult, error) {
	ev := event.Event
	attrs := event.Attributes
	changeAlias := EntityAliasRef{
		Provider:       ev.Provider,
		ProviderSource: ev.ProviderSource,
		SubjectKind:    "change_event",
		SubjectRef:     ev.SubjectRef,
	}
	changeEventEntity := &ent.KnowledgeEntity{
		Kind:        kne.KindChangeEvent,
		DisplayName: attrs.DisplayName,
		Properties:  ev.Attributes,
		Edges: ent.KnowledgeEntityEdges{
			Aliases: []*ent.KnowledgeEntityAlias{
				{
					Provider:       changeAlias.Provider,
					ProviderSource: changeAlias.ProviderSource,
					SubjectKind:    changeAlias.SubjectKind,
					SubjectRef:     changeAlias.SubjectRef,
				},
			},
		},
	}

	entities := ent.KnowledgeEntities{changeEventEntity}
	relationships := make([]ProjectedRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := EntityAliasRef{
			Provider:       ev.Provider,
			ProviderSource: ev.ProviderSource,
			SubjectKind:    "repository",
			SubjectRef:     attrs.RepositoryExternalRef,
		}
		repoEntity := &ent.KnowledgeEntity{
			Kind:        kne.KindRepository,
			DisplayName: attrs.RepositoryExternalRef,
			Properties:  map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Edges: ent.KnowledgeEntityEdges{
				Aliases: []*ent.KnowledgeEntityAlias{
					{
						Provider:       repoAlias.Provider,
						ProviderSource: repoAlias.ProviderSource,
						SubjectKind:    repoAlias.SubjectKind,
						SubjectRef:     repoAlias.SubjectRef,
					},
				},
			},
		}
		entities = append(entities, repoEntity)

		repoChangeRelationship := ProjectedRelationship{
			Relationship: &ent.KnowledgeRelationship{
				Kind:        "changes_repository",
				DisplayName: "changes repository",
			},
			FromAlias: changeAlias,
			ToAlias:   repoAlias,
		}
		relationships = append(relationships, repoChangeRelationship)
	}

	return ProjectionResult{Entities: entities, Relationships: relationships}, nil
}
