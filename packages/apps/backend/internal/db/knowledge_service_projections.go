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
	"github.com/rezible/rezible/integrations/eventprojections"
)

const (
	evidenceSourceNormalizedEventProjection = "normalized_event_projection"

	assertionCodeRepositoryExists        = "code_repository_exists"
	assertionCodeChangeObserved          = "code_change_observed"
	assertionCodeChangeTouchedRepository = "code_change_touched_repository"

	knowledgeKindCodeRepository = "code_repository"
	knowledgeKindCodeChange     = "code_change"
	relationshipKindTouched     = "touched_repository"
)

func knowledgeEntityEventProjectionHandler(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	projectionEvent, validationErr := eventprojections.DecodeEvent(event)
	if validationErr != nil || projectionEvent == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	proj := newKnowledgeEntityEventProjector(event, newKnowledgeService(client))
	var result *ProjectionResult
	switch ev := projectionEvent.(type) {
	case eventprojections.RepositoryObserved:
		result = proj.projectRepositoryObserved(ev)
	case eventprojections.ChangeEventObserved:
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

type EntityAliasRef struct {
	Provider           string
	ProviderSource     string
	ProviderSubjectRef string
}

func (kp *knowledgeEntityEventProjector) observedAt() time.Time {
	if !kp.event.OccurredAt.IsZero() {
		return kp.event.OccurredAt
	}
	if !kp.event.ReceivedAt.IsZero() {
		return kp.event.ReceivedAt
	}
	return time.Now().UTC()
}

func (kp *knowledgeEntityEventProjector) saveProjectionResult(ctx context.Context, result *ProjectionResult) error {
	savedAliasRefLookup := make(map[EntityAliasRef]uuid.UUID)
	for _, projEntity := range result.Entities {
		saved, saveErr := kp.saveProjectedEntity(ctx, projEntity)
		if saveErr != nil {
			return fmt.Errorf("saving projected entity: %w", saveErr)
		}
		for _, aliasRef := range projEntity.Aliases {
			savedAliasRefLookup[aliasRef] = saved.ID
		}
	}
	for _, projRel := range result.Relationships {
		if saveErr := kp.saveProjectedRelationship(ctx, projRel, savedAliasRefLookup); saveErr != nil {
			return fmt.Errorf("save projected relationship: %w", saveErr)
		}
	}
	return nil
}

func (kp *knowledgeEntityEventProjector) convertProjectedAlias(alias *ent.KnowledgeEntityAlias) (*EntityAliasRef, error) {
	if alias == nil {
		return nil, fmt.Errorf("projected fact alias is nil")
	}
	if alias.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if alias.ProviderSource == "" {
		return nil, fmt.Errorf("provider_source is required")
	}
	if alias.ProviderSubjectRef == "" {
		return nil, fmt.Errorf("provider_subject_ref is required")
	}
	return &EntityAliasRef{
		Provider:           alias.Provider,
		ProviderSource:     alias.ProviderSource,
		ProviderSubjectRef: alias.ProviderSubjectRef,
	}, nil
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
				"projected aliases resolve to different entities: %s resolves to %s, expected %s",
				ref.ProviderSubjectRef,
				alias.EntityID,
				id,
			)
			break
		}
	}
	return id, err
}

func (kp *knowledgeEntityEventProjector) saveProjectedEntity(ctx context.Context, proj ProjectedEntity) (*ent.KnowledgeEntity, error) {
	// needed as knowledge entities do not have a stable identifier
	existingId, lookupErr := kp.resolveExistingProjectedEntityID(ctx, proj.Aliases...)
	if lookupErr != nil {
		return nil, fmt.Errorf("failed to resolve existing projected entity: %w", lookupErr)
	}

	var existing *ent.KnowledgeEntity
	if existingId != uuid.Nil {
		var existingErr error
		existing, existingErr = kp.ks.GetEntity(ctx, kne.ID(existingId))
		if existingErr != nil {
			return nil, fmt.Errorf("query existing projected entity: %w", existingErr)
		}
	}

	observedAt := kp.observedAt()
	mergedProperties := proj.Properties
	evidenceKind := knev.EvidenceKindObserved
	if existing != nil {
		mergedProperties = proj.mergeProperties(existing.Properties)
		if !reflect.DeepEqual(existing.Properties, mergedProperties) {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	setEntity := func(m *ent.KnowledgeEntityMutation) {
		if existing != nil {
			m.SetKind(existing.Kind)
			if proj.DisplayName != "" && proj.DisplayName != existing.DisplayName && !proj.IsPlaceholder {
				m.SetDisplayName(proj.DisplayName)
			} else {
				m.SetDisplayName(existing.DisplayName)
			}
			m.SetDescription(existing.Description)
			m.SetProperties(mergedProperties)
			if existing.FirstObservedAt == nil {
				m.SetFirstObservedAt(observedAt)
			}
		} else {
			m.SetKind(proj.Kind)
			m.SetDisplayName(proj.DisplayName)
			m.SetDescription(proj.Description)
			m.SetProperties(mergedProperties)
			m.SetFirstObservedAt(observedAt)
		}
		m.SetLastObservedAt(observedAt)
		m.ClearDeletedAt()
	}
	savedEntity, entityErr := kp.ks.SetEntity(ctx, existingId, setEntity)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}
	for _, alias := range proj.Aliases {
		existingAlias, aliasLookupErr := kp.ks.lookupEntityAliasRef(ctx, alias)
		if aliasLookupErr != nil && !ent.IsNotFound(aliasLookupErr) {
			return nil, fmt.Errorf("lookup existing alias: %w", aliasLookupErr)
		} else if existingAlias != nil && existingAlias.EntityID != savedEntity.ID {
			return nil, fmt.Errorf(
				"knowledge alias %s/%s/%s already belongs to entity %s, cannot attach to entity %s",
				alias.Provider,
				alias.ProviderSource,
				alias.ProviderSubjectRef,
				existingAlias.EntityID,
				savedEntity.ID,
			)
		}

		setEntityAlias := func(m *ent.KnowledgeEntityAliasMutation) {
			m.SetEntityID(savedEntity.ID)
			m.SetProvider(alias.Provider)
			m.SetProviderSource(alias.ProviderSource)
			m.SetProviderSubjectRef(alias.ProviderSubjectRef)
		}
		savedAlias, aliasErr := kp.ks.SetEntityAlias(ctx, uuid.Nil, setEntityAlias)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}

		setAliasEntityEvidence := func(m *ent.KnowledgeEvidenceMutation) {
			m.SetSubjectType(knev.SubjectTypeEntity)
			m.SetEntityID(savedEntity.ID)
			m.SetAliasID(savedAlias.ID)
			m.SetNormalizedEventID(kp.event.ID)
			m.SetAssertionKind(proj.AssertionKind)
			m.SetEvidenceKind(evidenceKind)
			m.SetObservedAt(observedAt)
			m.SetSource(evidenceSourceNormalizedEventProjection)
			m.SetProperties(proj.Properties)
		}
		_, evidenceErr := kp.ks.AddEvidence(ctx, setAliasEntityEvidence)
		if evidenceErr != nil {
			return nil, fmt.Errorf("record entity evidence: %w", evidenceErr)
		}
	}
	return savedEntity, nil
}

func (kp *knowledgeEntityEventProjector) saveProjectedRelationship(ctx context.Context, rel ProjectedRelationship, refLookup map[EntityAliasRef]uuid.UUID) error {
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

	existing, existingErr := kp.ks.GetRelationship(ctx, knr.And(
		knr.SourceEntityID(fromId),
		knr.TargetEntityID(toId),
		knr.Kind(rel.Kind),
	))
	if existingErr != nil && !ent.IsNotFound(existingErr) {
		return fmt.Errorf("query existing relationship: %w", existingErr)
	}

	observedAt := kp.observedAt()
	evidenceKind := knev.EvidenceKindObserved
	if existing != nil && !reflect.DeepEqual(existing.Properties, rel.Properties) {
		evidenceKind = knev.EvidenceKindChanged
	}

	setRelationshipFn := func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(fromId)
		m.SetTargetEntityID(toId)
		m.SetKind(rel.Kind)
		m.SetDisplayName(rel.DisplayName)
		m.SetDescription(rel.Description)
		if existing == nil || existing.FirstObservedAt == nil {
			m.SetFirstObservedAt(observedAt)
		}
		m.SetLastObservedAt(observedAt)
		m.ClearDeletedAt()
		if rel.Properties != nil {
			m.SetProperties(rel.Properties)
		}
	}
	savedRel, saveErr := kp.ks.SetRelationship(ctx, uuid.Nil, setRelationshipFn)
	if saveErr != nil {
		return fmt.Errorf("upsert knowledge relationship: %w", saveErr)
	}

	addRelationshipEvidenceFn := func(m *ent.KnowledgeEvidenceMutation) {
		m.SetSubjectType(knev.SubjectTypeRelationship)
		m.SetRelationshipID(savedRel.ID)
		m.SetNormalizedEventID(kp.event.ID)
		m.SetAssertionKind(rel.AssertionKind)
		m.SetEvidenceKind(evidenceKind)
		m.SetObservedAt(observedAt)
		m.SetSource(evidenceSourceNormalizedEventProjection)
		m.SetProperties(rel.Properties)
	}
	_, evidenceErr := kp.ks.AddEvidence(ctx, addRelationshipEvidenceFn)
	if evidenceErr != nil {
		return fmt.Errorf("record relationship evidence: %w", evidenceErr)
	}
	return nil
}

// Event projections

type ProjectionResult struct {
	Entities      []ProjectedEntity
	Relationships []ProjectedRelationship
}

type ProjectedEntity struct {
	Kind          string
	AssertionKind string
	DisplayName   string
	Description   string
	Properties    map[string]any
	Aliases       []EntityAliasRef
	IsPlaceholder bool
}

func (pe ProjectedEntity) mergeProperties(existing map[string]any) map[string]any {
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

type ProjectedRelationship struct {
	Kind          string
	AssertionKind string
	DisplayName   string
	Description   string
	Properties    map[string]any
	FromAlias     EntityAliasRef
	ToAlias       EntityAliasRef
}

func (kp *knowledgeEntityEventProjector) projectRepositoryObserved(pe eventprojections.RepositoryObserved) *ProjectionResult {
	repoEntity := ProjectedEntity{
		Kind:          knowledgeKindCodeRepository,
		AssertionKind: assertionCodeRepositoryExists,
		DisplayName:   pe.Attributes.DisplayName,
		Properties:    pe.Event.Attributes,
		Aliases: []EntityAliasRef{
			{
				Provider:           pe.Event.Provider,
				ProviderSource:     pe.Event.ProviderSource,
				ProviderSubjectRef: pe.Event.SubjectRef,
			},
		},
	}
	return &ProjectionResult{Entities: []ProjectedEntity{repoEntity}}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEventObserved(pe eventprojections.ChangeEventObserved) *ProjectionResult {
	ev := pe.Event
	attrs := pe.Attributes
	changeEventAlias := EntityAliasRef{
		Provider:           pe.Event.Provider,
		ProviderSource:     pe.Event.ProviderSource,
		ProviderSubjectRef: pe.Event.SubjectRef,
	}
	codeChangedEntity := ProjectedEntity{
		Kind:          knowledgeKindCodeChange,
		AssertionKind: assertionCodeChangeObserved,
		DisplayName:   attrs.DisplayName,
		Properties:    ev.Attributes,
		Aliases:       []EntityAliasRef{changeEventAlias},
	}

	entities := []ProjectedEntity{codeChangedEntity}
	relationships := make([]ProjectedRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := EntityAliasRef{
			Provider:           ev.Provider,
			ProviderSource:     ev.ProviderSource,
			ProviderSubjectRef: attrs.RepositoryExternalRef,
		}
		entities = append(entities, ProjectedEntity{
			Kind:          knowledgeKindCodeRepository,
			AssertionKind: assertionCodeRepositoryExists,
			DisplayName:   attrs.RepositoryExternalRef,
			Properties:    map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Aliases:       []EntityAliasRef{repoAlias},
			IsPlaceholder: true,
		})

		repoChangeRelationship := ProjectedRelationship{
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

	return &ProjectionResult{Entities: entities, Relationships: relationships}
}
