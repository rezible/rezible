package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeService struct {
	db *ent.Client
}

func NewKnowledgeService(db *ent.Client) *KnowledgeService {
	return &KnowledgeService{db: db}
}

func (s *KnowledgeService) GetEntity(ctx context.Context, p predicate.KnowledgeEntity) (*ent.KnowledgeEntity, error) {
	return s.db.KnowledgeEntity.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityMutation)) (*ent.KnowledgeEntity, error) {
	return ent.WithTxReturning(ctx, s.db, func(tx *ent.Tx) (*ent.KnowledgeEntity, error) {
		var mutator ent.EntityMutator[*ent.KnowledgeEntity, *ent.KnowledgeEntityMutation]
		if id == uuid.Nil {
			mutator = tx.KnowledgeEntity.Create().SetID(uuid.New())
		} else {
			mutator = tx.KnowledgeEntity.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		savedEntity, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save: %w", saveErr)
		}
		return savedEntity, nil
	})
}

func (s *KnowledgeService) SetEntityAlias(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeEntityAliasMutation)) (*ent.KnowledgeEntityAlias, error) {
	return ent.WithTxReturning(ctx, s.db, func(tx *ent.Tx) (*ent.KnowledgeEntityAlias, error) {
		var mutator ent.EntityMutator[*ent.KnowledgeEntityAlias, *ent.KnowledgeEntityAliasMutation]
		if id == uuid.Nil {
			create := tx.KnowledgeEntityAlias.Create().SetID(uuid.New())
			mutator = create
			create.OnConflict(sql.ConflictColumns(
				knea.FieldTenantID,
				knea.FieldProvider,
				knea.FieldSource,
				knea.FieldExternalKind,
				knea.FieldExternalID,
			)).Update(func(u *ent.KnowledgeEntityAliasUpsert) {
				u.UpdateEntityID()
				u.UpdateLastSeenAt()
				u.UpdateUpdatedAt()
				u.UpdateNormalizedEventID()
			})
		} else {
			mutator = tx.KnowledgeEntityAlias.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		alias, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save: %w", saveErr)
		}

		return alias, nil
	})
}

func (s *KnowledgeService) GetRelationship(ctx context.Context, p predicate.KnowledgeRelationship) (*ent.KnowledgeRelationship, error) {
	return s.db.KnowledgeRelationship.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetRelationship(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeRelationshipMutation)) (*ent.KnowledgeRelationship, error) {
	return ent.WithTxReturning(ctx, s.db, func(tx *ent.Tx) (*ent.KnowledgeRelationship, error) {
		var mutator ent.EntityMutator[*ent.KnowledgeRelationship, *ent.KnowledgeRelationshipMutation]
		var mutation *ent.KnowledgeRelationshipMutation
		if id == uuid.Nil {
			create := tx.KnowledgeRelationship.Create().SetID(uuid.New())
			mutator = create
			create.OnConflict(sql.ConflictColumns(
				knr.FieldTenantID,
				knr.FieldSourceEntityID,
				knr.FieldTargetEntityID,
				knr.FieldKind,
			)).
				Update(func(u *ent.KnowledgeRelationshipUpsert) {
					u.UpdateLastSeenAt()
					u.UpdateUpdatedAt()
					if _, ok := mutation.DisplayName(); ok {
						u.UpdateDisplayName()
					}
					if _, ok := mutation.Description(); ok {
						u.UpdateDescription()
					}
				})
		} else {
			mutator = tx.KnowledgeRelationship.UpdateOneID(id)
		}
		mutation = mutator.Mutation()

		setFn(mutation)

		rel, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save relationship: %w", saveErr)
		}
		return rel, nil
	})
}
