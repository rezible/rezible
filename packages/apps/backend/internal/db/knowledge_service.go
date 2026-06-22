package db

import (
	"context"
	"fmt"
	"reflect"
	"sort"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeService struct {
	db rez.Database
}

func NewKnowledgeService(db rez.Database) *KnowledgeService {
	return &KnowledgeService{db: db}
}

func (s *KnowledgeService) GetEntity(ctx context.Context, p predicate.KnowledgeEntity) (*ent.KnowledgeEntity, error) {
	return s.db.Client(ctx).KnowledgeEntity.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityMutation)) (*ent.KnowledgeEntity, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeEntity, *ent.KnowledgeEntityMutation]
	if id == uuid.Nil {
		mutator = s.db.Client(ctx).KnowledgeEntity.Create()
	} else {
		mutator = s.db.Client(ctx).KnowledgeEntity.UpdateOneID(id)
	}
	setFn(mutator.Mutation())
	return mutator.Save(ctx)
}

func (s *KnowledgeService) GetEntityAlias(ctx context.Context, p predicate.KnowledgeEntityAlias) (*ent.KnowledgeEntityAlias, error) {
	return s.db.Client(ctx).KnowledgeEntityAlias.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetEntityAlias(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityAliasMutation)) (*ent.KnowledgeEntityAlias, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeEntityAlias, *ent.KnowledgeEntityAliasMutation]
	if id != uuid.Nil {
		mutator = s.db.Client(ctx).KnowledgeEntityAlias.UpdateOneID(id)
	} else {
		create := s.db.Client(ctx).KnowledgeEntityAlias.Create()
		create.OnConflictColumns(knea.FieldTenantID, knea.FieldProvider, knea.FieldProviderSubjectRef).
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

func (s *KnowledgeService) lookupAliasesFromRefs(ctx context.Context, refs ...ent.KnowledgeEntityAliasRef) (ent.KnowledgeEntityAliasSlice, error) {
	preds := make([]predicate.KnowledgeEntityAlias, len(refs))
	for i, ref := range refs {
		preds[i] = ref.Predicate()
	}
	queryAliases := s.db.Client(ctx).KnowledgeEntityAlias.Query().
		Where(knea.Or(preds...))
	return queryAliases.All(ctx)
}

func (s *KnowledgeService) LookupEntityIDFromAliasRefs(ctx context.Context, refs ...ent.KnowledgeEntityAliasRef) (uuid.UUID, error) {
	var entityId uuid.UUID

	aliases, aliasesErr := s.lookupAliasesFromRefs(ctx, refs...)
	if aliasesErr != nil {
		return entityId, fmt.Errorf("query aliases: %w", aliasesErr)
	}
	for _, a := range aliases {
		if entityId == uuid.Nil {
			entityId = a.EntityID
			continue
		}
		if entityId != a.EntityID {
			return entityId, fmt.Errorf("projected aliases resolve to different entities: %s -> %s (expected %s)",
				a.ProviderSubjectRef, a.EntityID, entityId)
		}
	}
	return entityId, nil
}

func sortKnowledgeEntityAliasRefs(refs []ent.KnowledgeEntityAliasRef) []ent.KnowledgeEntityAliasRef {
	sortedRefs := append([]ent.KnowledgeEntityAliasRef(nil), refs...)
	sort.SliceStable(sortedRefs, func(i, j int) bool {
		return sortedRefs[i].SortKey() < sortedRefs[j].SortKey()
	})
	return sortedRefs
}

func knowledgeEntitySortKey(pe rez.ProjectedKnowledgeEntity) string {
	if len(pe.AliasRefs) > 0 {
		firstRef := sortKnowledgeEntityAliasRefs(pe.AliasRefs)[0]
		return firstRef.SortKey()
	}
	return pe.Kind + "\x1f" + pe.DisplayName
}

func (s *KnowledgeService) acquireAliasRefsTxLock(ctx context.Context, refs []ent.KnowledgeEntityAliasRef) ([]ent.KnowledgeEntityAliasRef, error) {
	sortedRefs := sortKnowledgeEntityAliasRefs(refs)
	lockKeys := make([]string, 0, len(sortedRefs))
	seen := make(map[string]struct{}, len(sortedRefs))
	for _, ref := range sortedRefs {
		key := ref.SortKey()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		lockKeys = append(lockKeys, key)
	}
	if lockErr := s.db.AcquireTxLocks(ctx, "knowledge_entity_alias", lockKeys...); lockErr != nil {
		return nil, lockErr
	}
	return sortedRefs, nil
}

func (s *KnowledgeService) ResolveProjectedEntity(ctx context.Context, ev *ent.NormalizedEvent, pe rez.ProjectedKnowledgeEntity) (uuid.UUID, error) {
	var resolvedId uuid.UUID
	return resolvedId, s.db.WithTx(ctx, func(ctx context.Context, _ *ent.Client) error {
		aliasRefs, lockErr := s.acquireAliasRefsTxLock(ctx, pe.AliasRefs)
		if lockErr != nil {
			return fmt.Errorf("lock entity aliases: %w", lockErr)
		}
		resolved, resolveErr := s.resolveProjectedEntity(ctx, ev, pe, aliasRefs)
		if resolved != nil {
			resolvedId = resolved.ID
		}
		return resolveErr
	})
}

func (s *KnowledgeService) resolveProjectedEntity(ctx context.Context, ev *ent.NormalizedEvent, pe rez.ProjectedKnowledgeEntity, aliasRefs []ent.KnowledgeEntityAliasRef) (*ent.KnowledgeEntity, error) {
	currId, lookupCurrErr := s.LookupEntityIDFromAliasRefs(ctx, aliasRefs...)
	if lookupCurrErr != nil {
		return nil, fmt.Errorf("lookup existing entity from alias refs: %w", lookupCurrErr)
	}

	var current *ent.KnowledgeEntity
	if currId != uuid.Nil {
		current, lookupCurrErr = s.GetEntity(ctx, kne.ID(currId))
		if lookupCurrErr != nil {
			return nil, fmt.Errorf("lookup existing entity: %w", lookupCurrErr)
		}
	}

	properties := pe.Properties
	if pe.Properties == nil {
		properties = map[string]any{}
	}
	evidenceKind := knev.EvidenceKindObserved

	var existingId uuid.UUID
	if current != nil {
		existingId = current.ID
		if current.Kind != pe.Kind {
			return nil, fmt.Errorf("knowledge alias resolved to incompatible entity kind %q, expected %q", current.Kind, pe.Kind)
		}
		mergedProperties := make(map[string]any, len(current.Properties)+len(pe.Properties))
		for k, v := range current.Properties {
			mergedProperties[k] = v
		}
		var propertiesChanged bool
		for k, v := range pe.Properties {
			_, exists := mergedProperties[k]
			if !exists || !pe.IsPlaceholder {
				mergedProperties[k] = v
				propertiesChanged = true
			}
		}
		properties = mergedProperties

		undeleted := current.DeletedAt != nil
		displayChanged := current.DisplayName != pe.DisplayName || current.Description != pe.Description
		if propertiesChanged || undeleted || (displayChanged && !pe.IsPlaceholder) {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	observedAt := ev.DeriveObservedAt()
	setEntityFn := func(m *ent.KnowledgeEntityMutation) {
		shouldUpdate := current == nil || !pe.IsPlaceholder
		if shouldUpdate {
			m.SetKind(pe.Kind)
			m.SetDisplayName(pe.DisplayName)
			m.SetDescription(pe.Description)
		}
		if current == nil {
			m.SetFirstObservedAt(observedAt)
		}
		m.SetProperties(properties)
		m.SetLastObservedAt(observedAt)
		m.ClearDeletedAt()
	}
	savedEntity, entityErr := s.SetEntity(ctx, existingId, setEntityFn)
	if entityErr != nil {
		return nil, fmt.Errorf("set entity: %w", entityErr)
	}

	aliasIds := make(map[ent.KnowledgeEntityAliasRef]uuid.UUID)
	for _, ref := range aliasRefs {
		if _, seen := aliasIds[ref]; seen {
			continue
		}
		setAliasFn := func(m *ent.KnowledgeEntityAliasMutation) {
			m.SetEntityID(savedEntity.ID)
			m.SetProvider(ref.Provider)
			m.SetProviderSubjectRef(ref.ProviderSubjectRef)
		}
		alias, aliasErr := s.SetEntityAlias(ctx, uuid.Nil, setAliasFn)
		if aliasErr != nil {
			return nil, fmt.Errorf("upsert knowledge alias: %w", aliasErr)
		}
		aliasIds[ref] = alias.ID
	}

	if pe.EvidenceAssertion != "" {
		builders := make([]*ent.KnowledgeEvidenceCreate, 0, len(aliasIds))
		for _, aliasId := range aliasIds {
			createEvidence := s.db.Client(ctx).KnowledgeEvidence.Create().
				SetSubjectType(knev.SubjectTypeEntity).
				SetEventID(ev.ID).
				SetAssertion(pe.EvidenceAssertion).
				SetEntity(savedEntity).
				SetAliasID(aliasId).
				SetEvidenceKind(evidenceKind).
				SetObservedAt(observedAt)
			builders = append(builders, createEvidence)
		}
		_, evidenceErr := s.AddEvidence(ctx, builders...)
		if evidenceErr != nil {
			return nil, fmt.Errorf("record entity evidence: %w", evidenceErr)
		}
	}

	return savedEntity, nil
}

func (s *KnowledgeService) ResolveProjectedRelationship(ctx context.Context, ev *ent.NormalizedEvent, pr rez.ProjectedKnowledgeRelationship) (uuid.UUID, error) {
	var resolvedId uuid.UUID
	return resolvedId, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		_, lockErr := s.acquireAliasRefsTxLock(ctx, []ent.KnowledgeEntityAliasRef{pr.FromAliasRef, pr.ToAliasRef})
		if lockErr != nil {
			return fmt.Errorf("lock relationship aliases: %w", lockErr)
		}
		resolved, resolveErr := s.resolveProjectedRelationship(ctx, ev, pr)
		if resolved != nil {
			resolvedId = resolved.ID
		}
		return resolveErr
	})
}

func (s *KnowledgeService) resolveProjectedRelationship(ctx context.Context, ev *ent.NormalizedEvent, pr rez.ProjectedKnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	fromId, lookupFromRefErr := s.LookupEntityIDFromAliasRefs(ctx, pr.FromAliasRef)
	if lookupFromRefErr != nil {
		return nil, fmt.Errorf("lookup from ref: %w", lookupFromRefErr)
	}
	toId, lookupToRefErr := s.LookupEntityIDFromAliasRefs(ctx, pr.ToAliasRef)
	if lookupToRefErr != nil {
		return nil, fmt.Errorf("lookup from ref: %w", lookupToRefErr)
	}
	if fromId == uuid.Nil || toId == uuid.Nil {
		return nil, fmt.Errorf("could not resolve entity aliases")
	}

	lookupExisting := knr.And(
		knr.Kind(pr.Kind),
		knr.SourceEntityID(fromId),
		knr.TargetEntityID(toId),
	)
	existing, existingErr := s.GetRelationship(ctx, lookupExisting)
	if existingErr != nil && !ent.IsNotFound(existingErr) {
		return nil, fmt.Errorf("query existing relationship: %w", existingErr)
	}

	evidenceKind := knev.EvidenceKindObserved
	var existingID uuid.UUID
	if existing != nil {
		existingID = existing.ID
		propsChanged := !reflect.DeepEqual(existing.Properties, pr.Properties)
		displayChanged := existing.DisplayName != pr.DisplayName || existing.Description != pr.Description
		undeleted := existing.DeletedAt != nil
		if propsChanged || displayChanged || undeleted {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	observedAt := ev.DeriveObservedAt()
	setRelationshipFn := func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(fromId)
		m.SetTargetEntityID(toId)
		m.SetKind(pr.Kind)
		m.SetDisplayName(pr.DisplayName)
		m.SetDescription(pr.Description)
		if existing == nil || existing.FirstObservedAt == nil {
			m.SetFirstObservedAt(observedAt)
		}
		m.SetLastObservedAt(observedAt)
		m.ClearDeletedAt()
		if pr.Properties != nil {
			m.SetProperties(pr.Properties)
		}
	}

	savedRel, saveErr := s.SetRelationship(ctx, existingID, setRelationshipFn)
	if saveErr != nil {
		return nil, fmt.Errorf("upsert relationship: %w", saveErr)
	}
	if len(pr.EvidenceAssertion) > 0 {
		createEvidence := s.db.Client(ctx).KnowledgeEvidence.Create().
			SetSubjectType(knev.SubjectTypeRelationship).
			SetRelationshipID(savedRel.ID).
			SetEventID(ev.ID).
			SetAssertion(pr.EvidenceAssertion).
			SetEvidenceKind(evidenceKind).
			SetObservedAt(observedAt)
		if _, evidenceErr := s.AddEvidence(ctx, createEvidence); evidenceErr != nil {
			return nil, fmt.Errorf("record relationship evidence: %w", evidenceErr)
		}
	}
	return savedRel, nil
}

func (s *KnowledgeService) GetRelationship(ctx context.Context, pred predicate.KnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	return s.db.Client(ctx).KnowledgeRelationship.Query().Where(pred).Only(ctx)
}

var uniqueKnowledgeRelationshipCols = sql.ConflictColumns(
	knr.FieldTenantID,
	knr.FieldSourceEntityID,
	knr.FieldTargetEntityID,
	knr.FieldKind,
)

func (s *KnowledgeService) SetRelationship(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeRelationshipMutation)) (*ent.KnowledgeRelationship, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeRelationship, *ent.KnowledgeRelationshipMutation]
	creating := id == uuid.Nil
	if creating {
		create := s.db.Client(ctx).KnowledgeRelationship.Create().
			SetID(uuid.New())
		create.OnConflict(uniqueKnowledgeRelationshipCols).
			UpdateNewValues()
		mutator = create
	} else {
		mutator = s.db.Client(ctx).KnowledgeRelationship.UpdateOneID(id)
	}

	mutation := mutator.Mutation()
	setFn(mutation)

	return mutator.Save(ctx)
}

func (s *KnowledgeService) AddEvidence(ctx context.Context, builders ...*ent.KnowledgeEvidenceCreate) ([]*ent.KnowledgeEvidence, error) {
	results := make([]*ent.KnowledgeEvidence, 0, len(builders))
	for _, b := range builders {
		mut := b.Mutation()
		entityID, hasEntity := mut.EntityID()
		relationshipID, hasRelationship := mut.RelationshipID()
		if hasEntity == hasRelationship {
			return nil, fmt.Errorf("requires exactly one of entity_id or relationship_id")
		}

		var subjectType knev.SubjectType
		if hasEntity {
			subjectType = knev.SubjectTypeEntity
		} else if hasRelationship {
			subjectType = knev.SubjectTypeRelationship
		} else {
			return nil, fmt.Errorf("builder requires exactly one of entity_id or relationship_id")
		}
		b.SetSubjectType(subjectType)

		eventID, hasEventID := mut.EventID()
		if !hasEventID {
			return nil, fmt.Errorf("event_id is required")
		}

		assertion, hasAssertion := mut.Assertion()
		if !hasAssertion || assertion == "" {
			return nil, fmt.Errorf("assertion is required")
		}

		query := s.db.Client(ctx).KnowledgeEvidence.Query().
			Where(
				knev.EventID(eventID),
				knev.SubjectTypeEQ(subjectType),
				knev.Assertion(assertion),
			)
		if hasEntity {
			query.Where(knev.EntityID(entityID))
		}
		if hasRelationship {
			query.Where(knev.RelationshipID(relationshipID))
		}
		existing, queryErr := query.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("query existing evidence: %w", queryErr)
		}
		if existing != nil {
			results = append(results, existing)
			continue
		}
		saved, saveErr := b.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save evidence: %w", saveErr)
		}
		results = append(results, saved)
	}
	return results, nil
}
