package db

import (
	"context"
	"fmt"
	"sync"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	kfh "github.com/rezible/rezible/ent/knowledgefacthistory"
	kfp "github.com/rezible/rezible/ent/knowledgefactprovenance"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/jobs"
)

type KnowledgeService struct {
	dbc *ent.Client
}

var registerKnowledgeServiceJobs sync.Once

func NewKnowledgeService(dbc *ent.Client) *KnowledgeService {
	svc := &KnowledgeService{dbc: dbc}
	registerKnowledgeServiceJobs.Do(svc.registerJobs)
	return svc
}

func (s *KnowledgeService) registerJobs() {
	jobs.RegisterWorkerFunc(s.HandleEventProjection)
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
	if id == uuid.Nil {
		create := s.dbc.KnowledgeEntityAlias.Create().SetID(uuid.New())
		mutator = create
		conflictCols := sql.ConflictColumns(
			knea.FieldTenantID,
			knea.FieldProvider,
			knea.FieldProviderSource,
			knea.FieldSubjectKind,
			knea.FieldSubjectRef,
		)
		create.OnConflict(conflictCols).Update(func(u *ent.KnowledgeEntityAliasUpsert) {
			u.UpdateEntityID()
			u.UpdateLastSeenAt()
			u.UpdateUpdatedAt()
			u.UpdateNormalizedEventID()
		})
	} else {
		mutator = s.dbc.KnowledgeEntityAlias.UpdateOneID(id)
	}

	setFn(mutator.Mutation())

	alias, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save: %w", saveErr)
	}

	return alias, nil
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
		conflictCols := sql.ConflictColumns(
			knr.FieldTenantID,
			knr.FieldSourceEntityID,
			knr.FieldTargetEntityID,
			knr.FieldKind,
		)
		create.OnConflict(conflictCols).Update(func(u *ent.KnowledgeRelationshipUpsert) {
			u.UpdateLastSeenAt()
			u.UpdateUpdatedAt()
			if _, ok := mutation.DisplayName(); ok {
				u.UpdateDisplayName()
			}
			if _, ok := mutation.Description(); ok {
				u.UpdateDescription()
			}
		})
	}

	rel, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save relationship: %w", saveErr)
	}
	return rel, nil
}

func (s *KnowledgeService) SetFactProvenance(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeFactProvenanceMutation)) (*ent.KnowledgeFactProvenance, error) {
	creating := id == uuid.Nil
	var mutator ent.EntityMutator[*ent.KnowledgeFactProvenance, *ent.KnowledgeFactProvenanceMutation]
	var create *ent.KnowledgeFactProvenanceCreate
	if creating {
		create = s.dbc.KnowledgeFactProvenance.Create().SetID(uuid.New())
		mutator = create
	} else {
		mutator = s.dbc.KnowledgeFactProvenance.UpdateOneID(id)
	}

	mutation := mutator.Mutation()
	setFn(mutation)

	if creating {
		_, aliasSet := mutation.AliasID()
		_, relationshipSet := mutation.RelationshipID()
		if aliasSet == relationshipSet {
			return nil, fmt.Errorf("fact provenance must reference exactly one alias or relationship")
		}
		conflictCols := []string{
			kfp.FieldTenantID,
			kfp.FieldProvider,
			kfp.FieldProviderSource,
			kfp.FieldProviderEventRef,
			kfp.FieldExtractionMethod,
		}
		if aliasSet {
			conflictCols = append(conflictCols, kfp.FieldAliasID)
		} else {
			conflictCols = append(conflictCols, kfp.FieldRelationshipID)
		}
		create.
			OnConflict(sql.ConflictColumns(conflictCols...)).
			Update(func(u *ent.KnowledgeFactProvenanceUpsert) {
				u.UpdateLastSeenAt()
				u.UpdateUpdatedAt()
				if _, ok := mutation.NormalizedEventID(); ok {
					u.UpdateNormalizedEventID()
				}
			})
	}

	prov, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save fact provenance: %w", saveErr)
	}
	return prov, nil
}

func (s *KnowledgeService) SetFactHistory(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeFactHistoryMutation)) (*ent.KnowledgeFactHistory, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeFactHistory, *ent.KnowledgeFactHistoryMutation]
	if id == uuid.Nil {
		create := s.dbc.KnowledgeFactHistory.Create().SetID(uuid.New())
		conflictCols := sql.ConflictColumns(
			kfh.FieldTenantID,
			kfh.FieldHistoryKey,
		)
		create.OnConflict(conflictCols).Ignore()
		mutator = create
	} else {
		mutator = s.dbc.KnowledgeFactHistory.UpdateOneID(id)
	}

	setFn(mutator.Mutation())

	history, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save fact history: %w", saveErr)
	}
	return history, nil
}
