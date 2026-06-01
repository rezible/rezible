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
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeService struct {
	db rez.Database
}

func NewKnowledgeService(db rez.Database) *KnowledgeService {
	s := &KnowledgeService{db: db}
	return s
}

func (s *KnowledgeService) GetEntity(ctx context.Context, p predicate.KnowledgeEntity) (*ent.KnowledgeEntity, error) {
	return s.db.Client(ctx).KnowledgeEntity.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityMutation)) (*ent.KnowledgeEntity, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeEntity, *ent.KnowledgeEntityMutation]
	if id == uuid.Nil {
		mutator = s.db.Client(ctx).KnowledgeEntity.Create().SetID(uuid.New())
	} else {
		mutator = s.db.Client(ctx).KnowledgeEntity.UpdateOneID(id)
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
		mutator = s.db.Client(ctx).KnowledgeEntityAlias.UpdateOneID(id)
	} else {
		create := s.db.Client(ctx).KnowledgeEntityAlias.Create().SetID(uuid.New())
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

func (s *KnowledgeService) ResolveEntityAliases(ctx context.Context, aliases ...*ent.KnowledgeEntityAlias) ([]*ent.KnowledgeEntityAlias, error) {
	resolved := make([]*ent.KnowledgeEntityAlias, 0, len(aliases))
	var entityId uuid.UUID

	for _, ref := range aliases {
		query := s.db.Client(ctx).KnowledgeEntityAlias.Query().
			Where(knea.Provider(ref.Provider), knea.ProviderSubjectRef(ref.ProviderSubjectRef)).
			WithEntity()
		alias, queryErr := query.Only(ctx)
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("lookup entity alias: %w", queryErr)
		}
		if alias == nil {
			continue
		}
		resolved = append(resolved, alias)
		if entityId == uuid.Nil {
			entityId = alias.EntityID
			continue
		}
		if entityId != alias.EntityID {
			return nil, fmt.Errorf("projected aliases resolve to different entities: %s -> %s (expected %s)",
				ref.ProviderSubjectRef, alias.EntityID, entityId)
		}
	}

	return resolved, nil
}

func (s *KnowledgeService) ResolveEntity(ctx context.Context, params rez.ResolveKnowledgeEntityParams) (*ent.KnowledgeEntity, error) {
	var current *ent.KnowledgeEntity
	resolvedAliases, resolveAliasesErr := s.ResolveEntityAliases(ctx, params.Aliases...)
	if resolveAliasesErr != nil {
		return nil, fmt.Errorf("resolve existing entity aliases: %w", resolveAliasesErr)
	}
	for _, alias := range resolvedAliases {
		if alias.Edges.Entity != nil {
			current = alias.Edges.Entity
			break
		}
	}
	isCreate := current == nil

	mergedProperties := params.Event.Attributes
	if mergedProperties == nil {
		mergedProperties = map[string]any{}
	}
	evidenceKind := knev.EvidenceKindObserved
	var existingId uuid.UUID
	if current != nil {
		existingId = current.ID
		if current.Kind != params.Entity.Kind {
			return nil, fmt.Errorf("knowledge alias resolved to incompatible entity kind %q, expected %q",
				current.Kind, params.Entity.Kind)
		}
		mergedProperties = s.mergeProperties(current.Properties, params.Event.Attributes, params.IsPlaceholder)
		propsChanged := !reflect.DeepEqual(current.Properties, mergedProperties)
		undeleted := current.DeletedAt != nil
		displayChanged := !params.IsPlaceholder &&
			(current.DisplayName != params.Entity.DisplayName || current.Description != params.Entity.Description)
		if propsChanged || displayChanged || undeleted {
			evidenceKind = knev.EvidenceKindChanged
		}
	}

	observedAt := observedAtForEvent(params.Event)
	setEntityFn := func(m *ent.KnowledgeEntityMutation) {
		shouldRefreshIdentity := isCreate || !params.IsPlaceholder
		if shouldRefreshIdentity {
			m.SetKind(params.Entity.Kind)
		}
		if shouldRefreshIdentity || current.DisplayName == "" {
			m.SetDisplayName(params.Entity.DisplayName)
		}
		if shouldRefreshIdentity || current.Description == "" {
			m.SetDescription(params.Entity.Description)
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

	aliasRefs := make(map[string]*ent.KnowledgeEntityAlias)
	for _, alias := range resolvedAliases {
		aliasRefs[s.makeAliasRef(alias)] = alias
	}

	savedAliases := make([]*ent.KnowledgeEntityAlias, 0, len(params.Aliases))
	for _, alias := range params.Aliases {
		existingAlias, resolved := aliasRefs[s.makeAliasRef(alias)]
		if resolved && existingAlias != nil {
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
	savedEntity.Edges.Aliases = savedAliases

	// add evidence for event
	if params.EvidenceAssertion != "" {
		builders := make([]*ent.KnowledgeEvidenceCreate, len(savedAliases))
		for i, alias := range savedAliases {
			builders[i] = s.db.Client(ctx).KnowledgeEvidence.Create().
				SetEventID(params.Event.ID).
				SetAssertion(params.EvidenceAssertion).
				SetEntityID(savedEntity.ID).
				SetAliasID(alias.ID).
				SetEvidenceKind(evidenceKind)
		}
		savedEvidence, evidenceErr := s.AddEvidence(ctx, builders...)
		if evidenceErr != nil {
			return nil, fmt.Errorf("record entity evidence: %w", evidenceErr)
		}
		savedEntity.Edges.Evidence = savedEvidence
	}

	return savedEntity, nil
}

func (s *KnowledgeService) makeAliasRef(a *ent.KnowledgeEntityAlias) string {
	return fmt.Sprintf("%s_%s", a.Provider, a.ProviderSubjectRef)
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

func (s *KnowledgeService) ResolveRelationship(ctx context.Context, rel *ent.KnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	lookupExistingPred := knr.And(
		knr.Kind(rel.Kind),
		knr.SourceEntityID(rel.SourceEntityID),
		knr.TargetEntityID(rel.TargetEntityID),
	)
	lookupQuery := s.db.Client(ctx).KnowledgeRelationship.Query().Where(lookupExistingPred)
	existing, existingErr := lookupQuery.Only(ctx)
	if existingErr != nil && !ent.IsNotFound(existingErr) {
		return nil, fmt.Errorf("query existing relationship: %w", existingErr)
	}

	evidenceKind := knev.EvidenceKindObserved
	var existingID uuid.UUID
	if existing != nil {
		existingID = existing.ID
		propsChanged := !reflect.DeepEqual(existing.Properties, rel.Properties)
		displayChanged := existing.DisplayName != rel.DisplayName || existing.Description != rel.Description
		undeleted := existing.DeletedAt != nil
		if propsChanged || displayChanged || undeleted {
			evidenceKind = knev.EvidenceKindChanged
		}
	}
	for _, ev := range rel.Edges.Evidence {
		ev.EvidenceKind = evidenceKind
	}

	firstObservedAt := time.Now().UTC()
	if rel.FirstObservedAt != nil {
		firstObservedAt = *rel.FirstObservedAt
	}
	lastObservedAt := time.Now().UTC()
	if rel.LastObservedAt != nil {
		lastObservedAt = *rel.LastObservedAt
	}
	setRelationshipFn := func(m *ent.KnowledgeRelationshipMutation) {
		m.SetSourceEntityID(rel.SourceEntityID)
		m.SetTargetEntityID(rel.TargetEntityID)
		m.SetKind(rel.Kind)
		m.SetDisplayName(rel.DisplayName)
		m.SetDescription(rel.Description)
		if existing == nil || existing.FirstObservedAt == nil {
			m.SetFirstObservedAt(firstObservedAt)
		}
		m.SetLastObservedAt(lastObservedAt)
		m.ClearDeletedAt()
		if rel.Properties != nil {
			m.SetProperties(rel.Properties)
		}
	}

	resolved := existing
	txFn := func(txCtx context.Context, tx *ent.Client) error {
		savedRel, saveErr := s.SetRelationship(ctx, existingID, setRelationshipFn)
		if saveErr != nil {
			return fmt.Errorf("upsert relationship: %w", saveErr)
		}
		if len(rel.Edges.Evidence) > 0 {
			builders := make([]*ent.KnowledgeEvidenceCreate, len(rel.Edges.Evidence))
			for i, e := range rel.Edges.Evidence {
				createEv := tx.KnowledgeEvidence.Create().
					SetSubjectType(knev.SubjectTypeRelationship).
					SetRelationshipID(savedRel.ID).
					SetEventID(e.ID).
					SetAssertion(e.Assertion).
					SetEvidenceKind(evidenceKind).
					SetObservedAt(lastObservedAt)
				builders[i] = createEv
			}
			savedEvidence, evidenceErr := s.AddEvidence(ctx, builders...)
			if evidenceErr != nil {
				return fmt.Errorf("record relationship evidence: %w", evidenceErr)
			}
			savedRel.Edges.Evidence = savedEvidence
		}
		resolved = savedRel
		return nil
	}
	if txErr := s.db.WithTx(ctx, txFn); txErr != nil {
		return nil, fmt.Errorf("failed to resolve relationship: %w", txErr)
	}
	return resolved, nil
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

	rel, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save relationship: %w", saveErr)
	}
	return rel, nil
}

func (s *KnowledgeService) AddEvidence(ctx context.Context, builders ...*ent.KnowledgeEvidenceCreate) ([]*ent.KnowledgeEvidence, error) {
	create := s.db.Client(ctx).KnowledgeEvidence.CreateBulk(builders...)

	ids := make([]uuid.UUID, len(builders))
	for i, b := range builders {
		ids[i] = uuid.New()
		b.SetID(ids[i])

		mut := b.Mutation()
		entityID, hasEntity := mut.EntityID()
		relationshipID, hasRelationship := mut.RelationshipID()
		if hasEntity == hasRelationship {
			return nil, fmt.Errorf("requires exactly one of entity_id or relationship_id")
		}
		if subjectType, ok := mut.SubjectType(); ok {
			if hasEntity && subjectType != knev.SubjectTypeEntity {
				return nil, fmt.Errorf("subject_type must be entity for entity_id %s", entityID)
			}
			if hasRelationship && subjectType != knev.SubjectTypeRelationship {
				return nil, fmt.Errorf("subject_type must be relationship for relationship_id %s", relationshipID)
			}
		}
	}

	conflictCols := []string{
		knev.FieldTenantID,
		knev.FieldEventID,
		knev.FieldSubjectType,
		knev.FieldEntityID,
		knev.FieldRelationshipID,
	}
	upsert := create.OnConflictColumns(conflictCols...).DoNothing()

	if saveErr := upsert.Exec(ctx); saveErr != nil && !errors.Is(saveErr, stdsql.ErrNoRows) {
		return nil, fmt.Errorf("save evidence: %w", saveErr)
	}

	queryEvidence := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.IDIn(ids...))
	results, queryErr := queryEvidence.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query evidence: %w", queryErr)
	}
	return results, nil
}
