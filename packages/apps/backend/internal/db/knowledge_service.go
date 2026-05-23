package db

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeService struct {
	dbc *ent.Client
}

func newKnowledgeService(dbc *ent.Client) *KnowledgeService {
	return &KnowledgeService{dbc: dbc}
}

func (s *KnowledgeService) GetEntity(ctx context.Context, p predicate.KnowledgeEntity) (*ent.KnowledgeEntity, error) {
	return s.dbc.KnowledgeEntity.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityMutation)) (*ent.KnowledgeEntity, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeEntity, *ent.KnowledgeEntityMutation]
	if id == uuid.Nil {
		mutator = s.dbc.KnowledgeEntity.Create().SetID(uuid.New())
	} else {
		mutator = s.dbc.KnowledgeEntity.UpdateOneID(id)
	}

	setFn(mutator.Mutation())

	savedEntity, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save: %w", saveErr)
	}
	return savedEntity, nil
}

func (s *KnowledgeService) SetEntityAlias(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityAliasMutation)) (*ent.KnowledgeEntityAlias, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeEntityAlias, *ent.KnowledgeEntityAliasMutation]
	if id != uuid.Nil {
		mutator = s.dbc.KnowledgeEntityAlias.UpdateOneID(id)
	} else {
		create := s.dbc.KnowledgeEntityAlias.Create().SetID(uuid.New())
		conflictCols := sql.ConflictColumns(knea.FieldTenantID, knea.FieldProvider, knea.FieldProviderSubjectRef)
		create.OnConflict(conflictCols).
			UpdateUpdatedAt().
			UpdateDisplayName()
		mutator = create
	}

	setFn(mutator.Mutation())

	alias, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save: %w", saveErr)
	}

	return alias, nil
}

func (s *KnowledgeService) lookupEntityAliasRef(ctx context.Context, ref EntityAliasRef) (*ent.KnowledgeEntityAlias, error) {
	queryExisting := s.dbc.KnowledgeEntityAlias.Query().
		Where(knea.Provider(ref.Provider), knea.ProviderSubjectRef(ref.ProviderSubjectRef))
	return queryExisting.Only(ctx)
}

func (s *KnowledgeService) resolveExistingEntityIDByAliases(ctx context.Context, refs ...EntityAliasRef) (uuid.UUID, error) {
	result, err := s.resolveAliases(ctx, refs...)
	if err != nil || result.Entity == nil {
		return uuid.Nil, err
	}
	return result.Entity.ID, nil
}

type resolvedKnowledgeAlias struct {
	Entity  *ent.KnowledgeEntity
	Aliases map[EntityAliasRef]*ent.KnowledgeEntityAlias
}

func (s *KnowledgeService) resolveAliases(ctx context.Context, refs ...EntityAliasRef) (*resolvedKnowledgeAlias, error) {
	result := &resolvedKnowledgeAlias{
		Aliases: make(map[EntityAliasRef]*ent.KnowledgeEntityAlias, len(refs)),
	}

	var resolvedId uuid.UUID
	for _, ref := range refs {
		query := s.dbc.KnowledgeEntityAlias.Query().
			Where(knea.Provider(ref.Provider), knea.ProviderSubjectRef(ref.ProviderSubjectRef)).
			WithEntity()
		alias, queryErr := query.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("lookup entity alias: %w", queryErr)
		}
		if alias == nil {
			continue
		}
		result.Aliases[ref] = alias
		if resolvedId == uuid.Nil {
			resolvedId = alias.EntityID
			result.Entity = alias.Edges.Entity
			continue
		}
		if resolvedId != alias.EntityID {
			return nil, fmt.Errorf("projected aliases resolve to different entities: %s -> %s (expected %s)",
				ref.ProviderSubjectRef, alias.EntityID, resolvedId)
		}
	}

	return result, nil
}

type ResolveKnowledgeEntityParams struct {
	IsPlaceholder bool
	Kind          string
	DisplayName   string
	Description   string
	Aliases       []EntityAliasRef
	Event         *ent.NormalizedEvent
}

type ResolvedKnowledgeEntity struct {
	Entity       *ent.KnowledgeEntity
	Aliases      []*ent.KnowledgeEntityAlias
	EvidenceKind knev.EvidenceKind
}

func (s *KnowledgeService) ResolveEntityByAliases(ctx context.Context, params ResolveKnowledgeEntityParams) (*ResolvedKnowledgeEntity, error) {
	aliasResolution, lookupErr := s.resolveAliases(ctx, params.Aliases...)
	if lookupErr != nil {
		return nil, fmt.Errorf("resolve existing entity: %w", lookupErr)
	}

	current := aliasResolution.Entity
	isCreate := current == nil

	mergedProperties := params.Event.Attributes
	if mergedProperties == nil {
		mergedProperties = map[string]any{}
	}
	evidenceKind := knev.EvidenceKindObserved
	var existingId uuid.UUID
	if current != nil {
		existingId = current.ID
		if current.Kind != params.Kind {
			return nil, fmt.Errorf("knowledge alias resolved to incompatible entity kind %q, expected %q", current.Kind, params.Kind)
		}

		mergedProperties = s.mergeProperties(current.Properties, params.Event.Attributes, params.IsPlaceholder)
		propsChanged := !reflect.DeepEqual(current.Properties, mergedProperties)
		undeleted := current.DeletedAt != nil
		displayChanged := !params.IsPlaceholder &&
			(current.DisplayName != params.DisplayName || current.Description != params.Description)
		if propsChanged || displayChanged || undeleted {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	observedAt := observedAtForEvent(params.Event)
	setEntityFn := func(m *ent.KnowledgeEntityMutation) {
		shouldRefreshIdentity := isCreate || !params.IsPlaceholder
		if shouldRefreshIdentity {
			m.SetKind(params.Kind)
		}
		if shouldRefreshIdentity || current.DisplayName == "" {
			m.SetDisplayName(params.DisplayName)
		}
		if shouldRefreshIdentity || current.Description == "" {
			m.SetDescription(params.Description)
		}
		if isCreate || current.FirstObservedAt == nil {
			m.SetFirstObservedAt(observedAt)
		}
		m.SetProperties(mergedProperties)
		m.SetLastObservedAt(observedAt)
		m.ClearDeletedAt()
	}
	savedEntity, entityErr := s.SetEntity(ctx, existingId, setEntityFn)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}

	savedAliases := make([]*ent.KnowledgeEntityAlias, 0, len(params.Aliases))
	for _, alias := range params.Aliases {
		if existingAlias, resolved := aliasResolution.Aliases[alias]; resolved && existingAlias != nil {
			savedAliases = append(savedAliases, existingAlias)
			continue
		}

		setAliasFn := func(m *ent.KnowledgeEntityAliasMutation) {
			m.SetEntityID(savedEntity.ID)
			m.SetProvider(alias.Provider)
			m.SetProviderSubjectRef(alias.ProviderSubjectRef)
		}
		savedAlias, aliasErr := s.SetEntityAlias(ctx, uuid.Nil, setAliasFn)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}

		savedAliases = append(savedAliases, savedAlias)
	}

	return &ResolvedKnowledgeEntity{
		Entity:       savedEntity,
		Aliases:      savedAliases,
		EvidenceKind: evidenceKind,
	}, nil
}

func (s *KnowledgeService) ResolveEntityWithAssertion(ctx context.Context, entityParams ResolveKnowledgeEntityParams, assertion string) (*ResolvedKnowledgeEntity, error) {
	resolved, resolveErr := s.ResolveEntityByAliases(ctx, entityParams)
	if resolveErr != nil {
		return nil, fmt.Errorf("resolve entity: %w", resolveErr)
	}
	if len(resolved.Aliases) == 0 {
		return nil, fmt.Errorf("resolved knowledge entity has no aliases")
	}
	evidenceParams := RecordKnowledgeEntityEvidenceParams{
		Event:        entityParams.Event,
		Assertion:    assertion,
		EntityID:     resolved.Entity.ID,
		Aliases:      resolved.Aliases,
		EvidenceKind: resolved.EvidenceKind,
	}
	if evidenceErr := s.RecordEntityEvidence(ctx, evidenceParams); evidenceErr != nil {
		return nil, fmt.Errorf("record entity evidence: %w", evidenceErr)
	}
	return resolved, nil
}

func (s *KnowledgeService) mergeProperties(existing, projected map[string]any, preserveExisting bool) map[string]any {
	if projected == nil {
		projected = map[string]any{}
	}
	merged := make(map[string]any, len(existing)+len(projected))
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range projected {
		if _, exists := merged[k]; exists && preserveExisting {
			continue
		}
		merged[k] = v
	}
	return merged
}

type RecordKnowledgeEntityEvidenceParams struct {
	EntityID     uuid.UUID
	Aliases      []*ent.KnowledgeEntityAlias
	Event        *ent.NormalizedEvent
	Assertion    string
	EvidenceKind knev.EvidenceKind
}

func (s *KnowledgeService) RecordEntityEvidence(ctx context.Context, params RecordKnowledgeEntityEvidenceParams) error {
	for _, alias := range params.Aliases {
		_, evidenceErr := s.AddEvidence(ctx, func(m *ent.KnowledgeEvidenceMutation) {
			m.SetSubjectType(knev.SubjectTypeEntity)
			m.SetEntityID(params.EntityID)
			m.SetAliasID(alias.ID)
			m.SetNormalizedEventID(params.Event.ID)
			m.SetAssertion(params.Assertion)
			m.SetEvidenceKind(params.EvidenceKind)
			m.SetObservedAt(observedAtForEvent(params.Event))
			if params.Event.Attributes != nil {
				m.SetProperties(params.Event.Attributes)
			}
		})
		if evidenceErr != nil {
			return fmt.Errorf("entity alias evidence: %w", evidenceErr)
		}
	}
	return nil
}

func (s *KnowledgeService) GetRelationship(ctx context.Context, p predicate.KnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	return s.dbc.KnowledgeRelationship.Query().Where(p).Only(ctx)
}

type ResolveKnowledgeRelationshipParams struct {
	Kind           string
	DisplayName    string
	Description    string
	SourceEntityID uuid.UUID
	TargetEntityID uuid.UUID
	ObservedAt     time.Time
	Attributes     map[string]any
}

type ResolvedKnowledgeRelationship struct {
	Relationship *ent.KnowledgeRelationship
	EvidenceKind knev.EvidenceKind
}

func (s *KnowledgeService) ResolveRelationship(ctx context.Context, params ResolveKnowledgeRelationshipParams) (*ResolvedKnowledgeRelationship, error) {
	lookupExistingPred := knr.And(
		knr.Kind(params.Kind),
		knr.SourceEntityID(params.SourceEntityID),
		knr.TargetEntityID(params.TargetEntityID),
	)
	existing, existingErr := s.GetRelationship(ctx, lookupExistingPred)
	if existingErr != nil && !ent.IsNotFound(existingErr) {
		return nil, fmt.Errorf("query existing relationship: %w", existingErr)
	}

	evidenceKind := knev.EvidenceKindObserved
	var existingID uuid.UUID
	if existing != nil {
		existingID = existing.ID
		propsChanged := !reflect.DeepEqual(existing.Properties, params.Attributes)
		displayChanged := existing.DisplayName != params.DisplayName || existing.Description != params.Description
		undeleted := existing.DeletedAt != nil
		if propsChanged || displayChanged || undeleted {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	savedRel, saveErr := s.SetRelationship(ctx, existingID, func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(params.SourceEntityID)
		m.SetTargetEntityID(params.TargetEntityID)
		m.SetKind(params.Kind)
		m.SetDisplayName(params.DisplayName)
		m.SetDescription(params.Description)
		if existing == nil || existing.FirstObservedAt == nil {
			m.SetFirstObservedAt(params.ObservedAt)
		}
		m.SetLastObservedAt(params.ObservedAt)
		m.ClearDeletedAt()
		if params.Attributes != nil {
			m.SetProperties(params.Attributes)
		}
	})
	if saveErr != nil {
		return nil, fmt.Errorf("upsert relationship: %w", saveErr)
	}

	return &ResolvedKnowledgeRelationship{
		Relationship: savedRel,
		EvidenceKind: evidenceKind,
	}, nil
}

type RecordKnowledgeRelationshipEvidenceParams struct {
	RelationshipID    uuid.UUID
	NormalizedEventID uuid.UUID
	EvidenceKind      knev.EvidenceKind
	Assertion         string
	ObservedAt        time.Time
	Attributes        map[string]any
}

func (s *KnowledgeService) RecordRelationshipEvidence(ctx context.Context, params RecordKnowledgeRelationshipEvidenceParams) error {
	_, evidenceErr := s.AddEvidence(ctx, func(m *ent.KnowledgeEvidenceMutation) {
		m.SetSubjectType(knev.SubjectTypeRelationship)
		m.SetRelationshipID(params.RelationshipID)
		m.SetNormalizedEventID(params.NormalizedEventID)
		m.SetAssertion(params.Assertion)
		m.SetEvidenceKind(params.EvidenceKind)
		m.SetObservedAt(params.ObservedAt)
		if params.Attributes != nil {
			m.SetProperties(params.Attributes)
		}
	})
	if evidenceErr != nil {
		return fmt.Errorf("record relationship evidence: %w", evidenceErr)
	}
	return nil
}

func (s *KnowledgeService) SetRelationship(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeRelationshipMutation)) (*ent.KnowledgeRelationship, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeRelationship, *ent.KnowledgeRelationshipMutation]
	var create *ent.KnowledgeRelationshipCreate
	creating := id == uuid.Nil
	if creating {
		create = s.dbc.KnowledgeRelationship.Create().SetID(uuid.New())
		mutator = create
	} else {
		mutator = s.dbc.KnowledgeRelationship.UpdateOneID(id)
	}

	mutation := mutator.Mutation()
	setFn(mutation)

	if creating {
		relationshipConflictCols := sql.ConflictColumns(
			knr.FieldTenantID,
			knr.FieldSourceEntityID,
			knr.FieldTargetEntityID,
			knr.FieldKind,
		)
		upsertRelationshipFn := func(u *ent.KnowledgeRelationshipUpsert) {
			u.UpdateUpdatedAt()
			if _, ok := mutation.DisplayName(); ok {
				u.UpdateDisplayName()
			}
			if _, ok := mutation.Description(); ok {
				u.UpdateDescription()
			}
			if _, ok := mutation.Properties(); ok {
				u.UpdateProperties()
			}
			if _, ok := mutation.FirstObservedAt(); ok {
				u.UpdateFirstObservedAt()
			}
			if _, ok := mutation.LastObservedAt(); ok {
				u.UpdateLastObservedAt()
			}
			if mutation.DeletedAtCleared() {
				u.ClearDeletedAt()
			} else if _, ok := mutation.DeletedAt(); ok {
				u.UpdateDeletedAt()
			}
		}
		create.OnConflict(relationshipConflictCols).Update(upsertRelationshipFn)
	}

	rel, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save relationship: %w", saveErr)
	}
	return rel, nil
}

func (s *KnowledgeService) AddEvidence(ctx context.Context, setFn func(*ent.KnowledgeEvidenceMutation)) (uuid.UUID, error) {
	create := s.dbc.KnowledgeEvidence.Create().SetID(uuid.New())
	mutation := create.Mutation()

	setFn(mutation)

	entityID, hasEntity := mutation.EntityID()
	relationshipID, hasRelationship := mutation.RelationshipID()
	if hasEntity == hasRelationship {
		return uuid.Nil, fmt.Errorf("requires exactly one of entity_id or relationship_id")
	}
	if subjectType, ok := mutation.SubjectType(); ok {
		if hasEntity && subjectType != knev.SubjectTypeEntity {
			return uuid.Nil, fmt.Errorf("subject_type must be entity for entity_id %s", entityID)
		}
		if hasRelationship && subjectType != knev.SubjectTypeRelationship {
			return uuid.Nil, fmt.Errorf("subject_type must be relationship for relationship_id %s", relationshipID)
		}
	}

	conflictCols := []string{
		knev.FieldTenantID,
		knev.FieldNormalizedEventID,
		knev.FieldSubjectType,
	}
	if hasEntity {
		conflictCols = append(conflictCols, knev.FieldEntityID)
	} else {
		conflictCols = append(conflictCols, knev.FieldRelationshipID)
	}
	upsert := create.OnConflictColumns(conflictCols...).DoNothing()

	id, saveErr := upsert.ID(ctx)
	if saveErr != nil && !errors.Is(saveErr, stdsql.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("save evidence: %w", saveErr)
	}
	return id, nil
}
