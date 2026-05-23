package db

import (
	"context"
	stdsql "database/sql"
	"errors"
	"fmt"

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

func (s *KnowledgeService) GetRelationship(ctx context.Context, p predicate.KnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	return s.dbc.KnowledgeRelationship.Query().Where(p).Only(ctx)
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
		knev.FieldAssertionKind,
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
