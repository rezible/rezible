package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	knf "github.com/rezible/rezible/ent/knowledgefact"
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

type FactAliasRef struct {
	Provider           string
	ProviderSource     string
	ProviderSubjectRef string
}

func (kp *knowledgeEntityEventProjector) saveProjectionResult(ctx context.Context, result *ProjectionResult) error {
	savedAliasRefLookup := make(map[FactAliasRef]uuid.UUID)
	for _, projFact := range result.Facts {
		saved, saveErr := kp.saveProjectedFact(ctx, projFact)
		if saveErr != nil {
			return fmt.Errorf("saving projected entity: %w", saveErr)
		}
		for _, aliasRef := range projFact.Aliases {
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

func (kp *knowledgeEntityEventProjector) convertProjectedAlias(alias *ent.KnowledgeFactAlias) (*FactAliasRef, error) {
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
	return &FactAliasRef{
		Provider:           alias.Provider,
		ProviderSource:     alias.ProviderSource,
		ProviderSubjectRef: alias.ProviderSubjectRef,
	}, nil
}

func (kp *knowledgeEntityEventProjector) resolveExistingProjectedEntityID(ctx context.Context, refs ...FactAliasRef) (id uuid.UUID, err error) {
	// definitely a better way to do this
	for _, ref := range refs {
		alias, queryErr := kp.ks.lookupFactAliasRef(ctx, ref)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			err = fmt.Errorf("failed to lookup fact alias: %w", queryErr)
			break
		}
		if alias == nil {
			continue
		} else if id == uuid.Nil {
			id = alias.FactID
		} else if alias.FactID != id {
			err = fmt.Errorf(
				"projected aliases resolve to different facts: %s resolves to %s, expected %s",
				ref.ProviderSubjectRef,
				alias.FactID,
				id,
			)
			break
		}
	}
	return id, err
}

func (kp *knowledgeEntityEventProjector) saveProjectedFact(ctx context.Context, proj ProjectedFact) (*ent.KnowledgeFact, error) {
	// needed as knowledge entities do not have a stable identifier
	existingId, lookupErr := kp.resolveExistingProjectedEntityID(ctx, proj.Aliases...)
	if lookupErr != nil {
		return nil, fmt.Errorf("failed to resolve existing projected entity: %w", lookupErr)
	}

	var existing *ent.KnowledgeFact
	if existingId != uuid.Nil {
		var existingErr error
		existing, existingErr = kp.ks.GetFact(ctx, knf.ID(existingId))
		if existingErr != nil {
			return nil, fmt.Errorf("query existing projected entity: %w", existingErr)
		}
	}

	setEntity := func(m *ent.KnowledgeFactMutation) {
		if existing != nil {
			m.SetKind(existing.Kind)
			m.SetDisplayName(existing.DisplayName)
			m.SetDescription(existing.Description)
			m.SetProperties(proj.mergeProperties(existing.Properties))
		} else {
			m.SetKind(proj.Kind)
			m.SetDisplayName(proj.DisplayName)
			m.SetDescription(proj.Description)
			m.SetProperties(proj.Properties)
		}
	}
	savedEntity, entityErr := kp.ks.SetFact(ctx, existingId, setEntity)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}
	for _, alias := range proj.Aliases {
		setFactAlias := func(m *ent.KnowledgeFactAliasMutation) {
			m.SetFactID(savedEntity.ID)
			m.SetProvider(alias.Provider)
			m.SetProviderSource(alias.ProviderSource)
			m.SetProviderSubjectRef(alias.ProviderSubjectRef)
		}
		savedAlias, aliasErr := kp.ks.SetFactAlias(ctx, uuid.Nil, setFactAlias)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}

		setAliasFactProvenance := func(m *ent.KnowledgeFactProvenanceMutation) {
			m.SetAliasID(savedAlias.ID)
			m.SetSource("normalized_event_projection")
			m.SetNormalizedEventID(kp.event.ID)
		}
		_, provErr := kp.ks.AddFactProvenance(ctx, setAliasFactProvenance)
		if provErr != nil {
			return nil, fmt.Errorf("record alias provenance: %w", provErr)
		}
	}
	return savedEntity, nil
}

func (kp *knowledgeEntityEventProjector) saveProjectedRelationship(ctx context.Context, rel ProjectedFactRelationship, refLookup map[FactAliasRef]uuid.UUID) error {
	var resolveEntityErr error
	fromId, fromAliasWasSaved := refLookup[rel.FromAlias]
	if !fromAliasWasSaved || fromId == uuid.Nil {
		fromId, resolveEntityErr = kp.resolveExistingProjectedEntityID(ctx, rel.FromAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("FromAlias: %w", resolveEntityErr)
		}
	}

	toId, toAliasWasSaved := refLookup[rel.FromAlias]
	if !toAliasWasSaved || toId == uuid.Nil {
		toId, resolveEntityErr = kp.resolveExistingProjectedEntityID(ctx, rel.ToAlias)
		if resolveEntityErr != nil {
			return fmt.Errorf("ToAlias: %w", resolveEntityErr)
		}
	}

	setRelationshipFn := func(m *ent.KnowledgeFactRelationshipMutation) {
		m.SetSourceFactID(fromId)
		m.SetTargetFactID(toId)
		m.SetKind(rel.Kind)
		m.SetDisplayName(rel.DisplayName)
		m.SetDescription(rel.Description)
		//m.SetFirstSeenAt(kp.event.OccurredAt)
		//m.SetLastSeenAt(kp.event.OccurredAt)
		if rel.Properties != nil {
			m.SetProperties(rel.Properties)
		}
	}
	savedRel, saveErr := kp.ks.SetRelationship(ctx, uuid.Nil, setRelationshipFn)
	if saveErr != nil {
		return fmt.Errorf("upsert knowledge relationship: %w", saveErr)
	}

	addFactProvenanceFn := func(m *ent.KnowledgeFactProvenanceMutation) {
		m.SetRelationshipID(savedRel.ID)
		m.SetSource("normalized_event_projection")
		m.SetNormalizedEventID(kp.event.ID)
	}
	_, provErr := kp.ks.AddFactProvenance(ctx, addFactProvenanceFn)
	if provErr != nil {
		return fmt.Errorf("record relationship provenance: %w", provErr)
	}
	return nil
}

// Event projections

type ProjectionResult struct {
	Facts         []ProjectedFact
	Relationships []ProjectedFactRelationship
}

type ProjectedFact struct {
	Kind          string
	DisplayName   string
	Description   string
	Properties    map[string]any
	Aliases       []FactAliasRef
	IsPlaceholder bool
}

func (pe ProjectedFact) mergeProperties(existing map[string]any) map[string]any {
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

type ProjectedFactRelationship struct {
	Kind        string
	DisplayName string
	Description string
	Properties  map[string]any
	FromAlias   FactAliasRef
	ToAlias     FactAliasRef
}

func (kp *knowledgeEntityEventProjector) projectRepositoryObserved(pe projections.RepositoryObserved) *ProjectionResult {
	repoFact := ProjectedFact{
		Kind:        "repository_exists",
		DisplayName: pe.Attributes.DisplayName,
		Properties:  pe.Event.Attributes,
		Aliases: []FactAliasRef{
			{
				Provider:           pe.Event.Provider,
				ProviderSource:     pe.Event.ProviderSource,
				ProviderSubjectRef: pe.Event.SubjectRef,
			},
		},
	}
	return &ProjectionResult{Facts: []ProjectedFact{repoFact}}
}

func (kp *knowledgeEntityEventProjector) projectCodeChangeEventObserved(pe projections.ChangeEventObserved) *ProjectionResult {
	ev := pe.Event
	attrs := pe.Attributes
	changeEventAlias := FactAliasRef{
		Provider:           pe.Event.Provider,
		ProviderSource:     pe.Event.ProviderSource,
		ProviderSubjectRef: pe.Event.SubjectRef,
	}
	codeChangedFact := ProjectedFact{
		Kind:        "code_changed",
		DisplayName: attrs.DisplayName,
		Properties:  ev.Attributes,
		Aliases:     []FactAliasRef{changeEventAlias},
	}

	facts := []ProjectedFact{codeChangedFact}
	relationships := make([]ProjectedFactRelationship, 0, 1)

	if attrs.RepositoryExternalRef != "" {
		repoAlias := FactAliasRef{
			Provider:           ev.Provider,
			ProviderSource:     ev.ProviderSource,
			ProviderSubjectRef: attrs.RepositoryExternalRef,
		}
		facts = append(facts, ProjectedFact{
			Kind:          "repository_exists",
			Properties:    map[string]any{"external_ref": attrs.RepositoryExternalRef},
			Aliases:       []FactAliasRef{repoAlias},
			IsPlaceholder: true,
		})

		repoChangeRelationship := ProjectedFactRelationship{
			Kind:        "changed_repository",
			DisplayName: "code changed repository",
			Properties: map[string]any{
				"repository_external_ref": attrs.RepositoryExternalRef,
			},
			FromAlias: changeEventAlias,
			ToAlias:   repoAlias,
		}
		relationships = append(relationships, repoChangeRelationship)
	}

	return &ProjectionResult{Facts: facts, Relationships: relationships}
}
