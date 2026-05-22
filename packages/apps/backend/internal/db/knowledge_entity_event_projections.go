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

func knowledgeEntityEventProjectionHandler(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	decoded, validationErr := projections.DecodeEvent[any](event)
	if validationErr != nil || decoded == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	proj := newKnowledgeEntityEventProjector(event, newKnowledgeService(client))
	var result *KnowledgeProjectionResult
	switch attrs := decoded.Attributes.(type) {
	case projections.RepositoryObservedAttributes:
		result = proj.projectRepositoryObserved(projections.RepositoryObserved{Event: event, Attributes: attrs})
	case projections.ChangeEventObservedAttributes:
		result = proj.projectCodeChangeEventObserved(projections.ChangeEventObserved{Event: event, Attributes: attrs})
	case projections.SystemComponentObservedAttributes:
		result = proj.projectSystemComponentObserved(projections.SystemComponentObserved{Event: event, Attributes: attrs})
	case projections.SystemRelationshipObservedAttributes:
		result = proj.projectSystemRelationshipObserved(projections.SystemRelationshipObserved{Event: event, Attributes: attrs})
	}
	if result != nil {
		return proj.saveProjectionResult(ctx, result)
	}
	return nil
}

type knowledgeEntityEventProjector struct {
	event     *ent.NormalizedEvent
	knowledge *KnowledgeService
}

func newKnowledgeEntityEventProjector(ev *ent.NormalizedEvent, ks *KnowledgeService) *knowledgeEntityEventProjector {
	return &knowledgeEntityEventProjector{event: ev, knowledge: ks}
}

type EntityAliasRef struct {
	Provider           string
	ProviderSource     string
	ProviderSubjectRef string
}

func (kp *knowledgeEntityEventProjector) saveProjectionResult(ctx context.Context, result *KnowledgeProjectionResult) error {
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

func (kp *knowledgeEntityEventProjector) resolveEventObservedAt(ev *ent.NormalizedEvent) time.Time {
	if !ev.OccurredAt.IsZero() {
		return ev.OccurredAt
	}
	if !ev.ReceivedAt.IsZero() {
		return ev.ReceivedAt
	}
	return time.Now().UTC()
}

func (kp *knowledgeEntityEventProjector) saveProjectedEntity(ctx context.Context, proj ProjectedKnowledgeEntity) (*ent.KnowledgeEntity, error) {
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
	}

	observedAt := kp.resolveEventObservedAt(kp.event)
	mergedProperties := proj.Properties
	evidenceKind := knev.EvidenceKindObserved
	if current != nil {
		mergedProperties = proj.mergeProperties(current.Properties)
		if !reflect.DeepEqual(current.Properties, mergedProperties) {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	setEntity := func(m *ent.KnowledgeEntityMutation) {
		if current == nil {
			m.SetKind(proj.Kind)
		}
		if !proj.IsPlaceholder || current == nil || current.DisplayName == "" {
			m.SetDisplayName(proj.DisplayName)
		}
		if !proj.IsPlaceholder || current == nil || current.Description == "" {
			m.SetDescription(proj.Description)
		}
		if current == nil || current.FirstObservedAt == nil {
			m.SetFirstObservedAt(observedAt)
		}
		m.SetProperties(mergedProperties)
		m.SetLastObservedAt(observedAt)
		m.ClearDeletedAt()
	}
	savedEntity, entityErr := kp.knowledge.SetEntity(ctx, existingId, setEntity)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}
	for _, alias := range proj.Aliases {
		existingAlias, aliasLookupErr := kp.knowledge.lookupEntityAliasRef(ctx, alias)
		if aliasLookupErr != nil && !ent.IsNotFound(aliasLookupErr) {
			return nil, fmt.Errorf("lookup existing alias: %w", aliasLookupErr)
		}
		var existingAliasId uuid.UUID
		if existingAlias != nil {
			if existingAlias.EntityID != savedEntity.ID {
				return nil, fmt.Errorf(
					"knowledge alias %s/%s/%s already belongs to entity %s, cannot attach to entity %s",
					alias.Provider,
					alias.ProviderSource,
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
			m.SetProviderSource(alias.ProviderSource)
			m.SetProviderSubjectRef(alias.ProviderSubjectRef)
		}
		savedAlias, aliasErr := kp.knowledge.SetEntityAlias(ctx, existingAliasId, setEntityAlias)
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
		_, evidenceErr := kp.knowledge.AddEvidence(ctx, setAliasEntityEvidence)
		if evidenceErr != nil {
			return nil, fmt.Errorf("record entity evidence: %w", evidenceErr)
		}
	}
	return savedEntity, nil
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

	observedAt := kp.resolveEventObservedAt(kp.event)
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
			m.SetFirstObservedAt(observedAt)
		}
		m.SetLastObservedAt(observedAt)
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
		m.SetObservedAt(observedAt)
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

type KnowledgeProjectionResult struct {
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

func (kp *knowledgeEntityEventProjector) makeEntityRef(ev *ent.NormalizedEvent, subjectRef string) EntityAliasRef {
	if subjectRef == "" {
		subjectRef = ev.SubjectRef
	}
	return EntityAliasRef{
		Provider:           ev.Provider,
		ProviderSource:     ev.ProviderSource,
		ProviderSubjectRef: subjectRef,
	}
}

func (kp *knowledgeEntityEventProjector) projectRepositoryObserved(pe projections.RepositoryObserved) *KnowledgeProjectionResult {
	repoEntity := ProjectedKnowledgeEntity{
		Kind:          knowledgeKindCodeRepository,
		AssertionKind: assertionCodeRepositoryExists,
		DisplayName:   pe.Attributes.DisplayName,
		Properties:    pe.Event.Attributes,
		Aliases:       []EntityAliasRef{kp.makeEntityRef(pe.Event, "")},
	}
	return &KnowledgeProjectionResult{Entities: []ProjectedKnowledgeEntity{repoEntity}}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEventObserved(pe projections.ChangeEventObserved) *KnowledgeProjectionResult {
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

	return &KnowledgeProjectionResult{Entities: entities, Relationships: relationships}
}

func (kp *knowledgeEntityEventProjector) projectSystemComponentObserved(pe projections.SystemComponentObserved) *KnowledgeProjectionResult {
	attrs := pe.Attributes
	componentEntity := ProjectedKnowledgeEntity{
		AssertionKind: assertionSystemComponentExists,
		Kind:          attrs.Kind,
		DisplayName:   attrs.DisplayName,
		Description:   attrs.Description,
		Properties:    attrs.Properties,
		Aliases:       []EntityAliasRef{kp.makeEntityRef(pe.Event, attrs.ExternalRef)},
	}
	return &KnowledgeProjectionResult{
		Entities: []ProjectedKnowledgeEntity{componentEntity},
	}
}

func (kp *knowledgeEntityEventProjector) projectSystemRelationshipObserved(pe projections.SystemRelationshipObserved) *KnowledgeProjectionResult {
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

	return &KnowledgeProjectionResult{
		Entities:      []ProjectedKnowledgeEntity{sourceEntity, targetEntity},
		Relationships: []ProjectedKnowledgeRelationship{relationship},
	}
}
