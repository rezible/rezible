package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	knfa "github.com/rezible/rezible/ent/knowledgefactalias"
	knfr "github.com/rezible/rezible/ent/knowledgefactrelationship"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeService struct {
	dbc *ent.Client
}

func newKnowledgeService(dbc *ent.Client) *KnowledgeService {
	return &KnowledgeService{dbc: dbc}
}

func (s *KnowledgeService) GetFact(ctx context.Context, p predicate.KnowledgeFact) (*ent.KnowledgeFact, error) {
	return s.dbc.KnowledgeFact.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetFact(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeFactMutation)) (*ent.KnowledgeFact, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeFact, *ent.KnowledgeFactMutation]
	if id == uuid.Nil {
		mutator = s.dbc.KnowledgeFact.Create().SetID(uuid.New())
	} else {
		mutator = s.dbc.KnowledgeFact.UpdateOneID(id)
	}

	setFn(mutator.Mutation())

	savedFact, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save: %w", saveErr)
	}
	return savedFact, nil
}

func (s *KnowledgeService) SetFactAlias(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeFactAliasMutation)) (*ent.KnowledgeFactAlias, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeFactAlias, *ent.KnowledgeFactAliasMutation]
	if id == uuid.Nil {
		create := s.dbc.KnowledgeFactAlias.Create().SetID(uuid.New())
		mutator = create
		conflictCols := sql.ConflictColumns(
			knfa.FieldTenantID,
			knfa.FieldFactID,
		)
		create.OnConflict(conflictCols).Update(func(u *ent.KnowledgeFactAliasUpsert) {
			u.UpdateUpdatedAt()
			u.UpdateDisplayName()
		})
	} else {
		mutator = s.dbc.KnowledgeFactAlias.UpdateOneID(id)
	}

	setFn(mutator.Mutation())

	alias, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save: %w", saveErr)
	}

	return alias, nil
}

func (s *KnowledgeService) lookupFactAliasRef(ctx context.Context, ref FactAliasRef) (*ent.KnowledgeFactAlias, error) {
	queryExisting := s.dbc.KnowledgeFactAlias.Query().Where(
		knfa.Provider(ref.Provider), knfa.ProviderSource(ref.ProviderSource),
		knfa.ProviderSubjectRef(ref.ProviderSubjectRef))
	return queryExisting.Only(ctx)
}

func (s *KnowledgeService) GetRelationship(ctx context.Context, p predicate.KnowledgeFactRelationship) (*ent.KnowledgeFactRelationship, error) {
	return s.dbc.KnowledgeFactRelationship.Query().Where(p).Only(ctx)
}

func (s *KnowledgeService) SetRelationship(ctx context.Context, id uuid.UUID, setFn func(*ent.KnowledgeFactRelationshipMutation)) (*ent.KnowledgeFactRelationship, error) {
	var mutator ent.EntityMutator[*ent.KnowledgeFactRelationship, *ent.KnowledgeFactRelationshipMutation]
	var create *ent.KnowledgeFactRelationshipCreate
	creating := id == uuid.Nil
	if creating {
		create = s.dbc.KnowledgeFactRelationship.Create().SetID(uuid.New())
		mutator = create
	} else {
		mutator = s.dbc.KnowledgeFactRelationship.UpdateOneID(id)
	}

	mutation := mutator.Mutation()
	setFn(mutation)

	if creating {
		conflictCols := sql.ConflictColumns(
			knfr.FieldTenantID,
			knfr.FieldSourceFactID,
			knfr.FieldTargetFactID,
			knfr.FieldKind,
		)
		create.OnConflict(conflictCols).Update(func(u *ent.KnowledgeFactRelationshipUpsert) {
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
		})
	}

	rel, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save relationship: %w", saveErr)
	}
	return rel, nil
}

func (s *KnowledgeService) AddFactProvenance(ctx context.Context, setFn func(*ent.KnowledgeFactProvenanceMutation)) (*ent.KnowledgeFactProvenance, error) {
	create := s.dbc.KnowledgeFactProvenance.Create().SetID(uuid.New())
	mutation := create.Mutation()

	setFn(mutation)

	prov, saveErr := create.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save fact provenance: %w", saveErr)
	}
	return prov, nil
}
