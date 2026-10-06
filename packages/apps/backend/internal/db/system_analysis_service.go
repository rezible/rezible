package db

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	sa "github.com/rezible/rezible/ent/systemanalysis"
	saent "github.com/rezible/rezible/ent/systemanalysisentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sarel "github.com/rezible/rezible/ent/systemanalysisrelationship"
)

type SystemAnalysisService struct {
	db        rez.Database
	knowledge rez.KnowledgeGraphQueryService
}

func NewSystemAnalysisService(db rez.Database, knowledge rez.KnowledgeGraphQueryService) (*SystemAnalysisService, error) {
	return &SystemAnalysisService{db: db, knowledge: knowledge}, nil
}

func (s *SystemAnalysisService) systemAnalysisEntrySubjectsQuery(q *ent.SystemAnalysisEntrySubjectQuery) {
	q.Order(ent.Asc(saes.FieldCreatedAt), ent.Asc(saes.FieldID)).
		WithKnowledgeEntity().
		WithKnowledgeRelationship().
		WithKnowledgeEvidence()
}

func (s *SystemAnalysisService) checkSaveErr(saveErr error, kind string) error {
	if ent.IsValidationError(saveErr) || ent.IsConstraintError(saveErr) {
		return fmt.Errorf("%w: save %s: %w", rez.ErrInvalidInput, kind, saveErr)
	}
	return fmt.Errorf("save %s: %w", kind, saveErr)
}

func (s *SystemAnalysisService) GetSystemAnalysis(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysis, error) {
	return s.db.Client(ctx).SystemAnalysis.Get(ctx, id)
}

func (s *SystemAnalysisService) HasSystemAnalysisEntity(ctx context.Context, analysisID, knowledgeEntityID uuid.UUID) (bool, error) {
	if analysisID == uuid.Nil || knowledgeEntityID == uuid.Nil {
		return false, fmt.Errorf("%w: analysis and knowledge entity IDs are required", rez.ErrInvalidInput)
	}
	query := s.db.Client(ctx).SystemAnalysisEntity.Query().
		Where(saent.AnalysisID(analysisID), saent.KnowledgeEntityID(knowledgeEntityID))
	exists, queryErr := query.Exist(ctx)
	if queryErr != nil {
		return false, fmt.Errorf("query system analysis entity: %w", queryErr)
	}
	return exists, nil
}

func (s *SystemAnalysisService) SetSystemAnalysis(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisMutation)) (*ent.SystemAnalysis, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.SystemAnalysis, error) {
		var mutator ent.EntityMutator[*ent.SystemAnalysis, *ent.SystemAnalysisMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysis.Create()
		} else {
			mutator = tx.SystemAnalysis.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)
		entityIDs := mapset.NewSet[uuid.UUID]()
		if entityID, set := mut.ScopeEntityID(); set {
			entityIDs.Add(entityID)
		}
		if entityID, set := mut.SubjectEntityID(); set {
			entityIDs.Add(entityID)
		}
		if validateErr := s.validateKnowledgeEntityTargets(ctx, tx, entityIDs); validateErr != nil {
			return nil, validateErr
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, s.checkSaveErr(saveErr, "system analysis")
		}
		return saved, nil
	})
}

func (s *SystemAnalysisService) validateSystemAnalysisTarget(ctx context.Context, tx *ent.Client, analysisID uuid.UUID) error {
	if analysisID == uuid.Nil {
		return fmt.Errorf("%w: system analysis ID is required", rez.ErrInvalidInput)
	}
	accessible, queryErr := tx.SystemAnalysis.Query().Where(sa.ID(analysisID)).Exist(ctx)
	if queryErr != nil {
		return fmt.Errorf("check system analysis access: %w", queryErr)
	}
	if !accessible {
		return fmt.Errorf("%w: system analysis is unavailable", rez.ErrInvalidInput)
	}
	return nil
}

func (s *SystemAnalysisService) validateKnowledgeEntityTargets(ctx context.Context, tx *ent.Client, entityIDs mapset.Set[uuid.UUID]) error {
	if entityIDs.IsEmpty() {
		return nil
	}
	queryEntities := tx.KnowledgeEntity.Query().Where(kne.IDIn(entityIDs.ToSlice()...))
	count, queryErr := queryEntities.Count(ctx)
	if queryErr != nil {
		return fmt.Errorf("check knowledge entity access: %w", queryErr)
	}
	if count != entityIDs.Cardinality() {
		return fmt.Errorf("%w: one or more knowledge entities are unavailable", rez.ErrInvalidInput)
	}
	return nil
}

func (s *SystemAnalysisService) IncludeSystemAnalysisSubjects(ctx context.Context, params rez.IncludeSystemAnalysisSubjectsParams) error {
	if params.AnalysisId == uuid.Nil || (len(params.EntityIds)+len(params.RelationshipIds) == 0) {
		return fmt.Errorf("%w: system analysis ID is required", rez.ErrInvalidInput)
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if validateErr := s.validateSystemAnalysisTarget(ctx, tx, params.AnalysisId); validateErr != nil {
			return validateErr
		}
		entitySubjIds := mapset.NewSet(params.EntityIds...)
		if len(params.RelationshipIds) > 0 {
			relEntIds, relEntsErr := s.resolveRelationshipEntityIDs(ctx, tx, mapset.NewSet(params.RelationshipIds...))
			if relEntsErr != nil {
				return relEntsErr
			}
			entitySubjIds.Append(relEntIds...)
		}
		if validateErr := s.validateKnowledgeEntityTargets(ctx, tx, entitySubjIds); validateErr != nil {
			return validateErr
		}

		if len(params.RelationshipIds) > 0 {
			upsert := tx.SystemAnalysisRelationship.
				MapCreateBulk(params.RelationshipIds, func(c *ent.SystemAnalysisRelationshipCreate, i int) {
					c.SetAnalysisID(params.AnalysisId)
					c.SetKnowledgeRelationshipID(params.RelationshipIds[i])
				}).
				OnConflict().
				DoNothing()
			if upsertErr := upsert.Exec(ctx); upsertErr != nil {
				return s.checkSaveErr(upsertErr, "relationship")
			}
		}

		if !entitySubjIds.IsEmpty() {
			if includeErr := s.includeAnalysisEntities(ctx, tx, params.AnalysisId, entitySubjIds); includeErr != nil {
				return fmt.Errorf("include entities: %w", includeErr)
			}
		}

		return nil
	})
}

func (s *SystemAnalysisService) resolveRelationshipEntityIDs(ctx context.Context, tx *ent.Client, relIds mapset.Set[uuid.UUID]) ([]uuid.UUID, error) {
	if relIds.IsEmpty() {
		return nil, nil
	}

	query := tx.KnowledgeRelationship.Query().
		Where(knr.IDIn(relIds.ToSlice()...)).
		Select(knr.FieldSourceEntityID, knr.FieldTargetEntityID)
	relationships, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query relationships: %w", queryErr)
	}

	numIds := relIds.Cardinality()
	if len(relationships) != numIds {
		return nil, fmt.Errorf("%w: one or more knowledge relationships do not exist", rez.ErrInvalidInput)
	}

	entityIDs := mapset.NewSetWithSize[uuid.UUID](numIds * 2)
	for _, rel := range relationships {
		entityIDs.Append(rel.SourceEntityID, rel.TargetEntityID)
	}
	if validateErr := s.validateKnowledgeEntityTargets(ctx, tx, entityIDs); validateErr != nil {
		return nil, fmt.Errorf("validate relationship endpoints: %w", validateErr)
	}
	return entityIDs.ToSlice(), nil
}

func (s *SystemAnalysisService) includeAnalysisEntities(ctx context.Context, tx *ent.Client, analysisID uuid.UUID, entityIds mapset.Set[uuid.UUID]) error {
	if entityIds.IsEmpty() {
		return nil
	}

	if validateErr := s.validateKnowledgeEntityTargets(ctx, tx, entityIds); validateErr != nil {
		return validateErr
	}
	ids := entityIds.ToSlice()
	upsert := tx.SystemAnalysisEntity.
		MapCreateBulk(ids, func(c *ent.SystemAnalysisEntityCreate, i int) {
			c.SetAnalysisID(analysisID)
			c.SetKnowledgeEntityID(ids[i])
		}).
		OnConflict().
		DoNothing()
	if upsertErr := upsert.Exec(ctx); upsertErr != nil {
		return s.checkSaveErr(upsertErr, "entity")
	}
	return nil
}

func (s *SystemAnalysisService) ListSystemAnalysisEntities(ctx context.Context, params rez.ListSystemAnalysisEntitiesParams) (*ent.ListResult[ent.SystemAnalysisEntity], error) {
	query := s.db.Client(ctx).SystemAnalysisEntity.Query().
		Where(params.Predicates...).
		WithKnowledgeEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases(subjectAliasWithEvidence())
		})
	if len(params.OrderBy) > 0 {
		query.Order(params.OrderBy...)
	} else {
		query.Order(saent.ByCreatedAt(), saent.ByID())
	}
	return ent.DoListQuery[ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityQuery](ctx, query, params.ListParams)
}

func (s *SystemAnalysisService) SetSystemAnalysisEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntityMutation)) (*ent.SystemAnalysisEntity, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.SystemAnalysisEntity, error) {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysisEntity.Create()
		} else {
			mutator = tx.SystemAnalysisEntity.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)

		if mutErr := s.validateMutationFields(mut, saent.FieldAnalysisID, saent.FieldKnowledgeEntityID); mutErr != nil {
			return nil, mutErr
		}
		if id == uuid.Nil {
			analysisID, _ := mut.AnalysisID()
			if validateErr := s.validateSystemAnalysisTarget(ctx, tx, analysisID); validateErr != nil {
				return nil, validateErr
			}
			entityID, _ := mut.KnowledgeEntityID()
			if validateErr := s.validateKnowledgeEntityTargets(ctx, tx, mapset.NewSet(entityID)); validateErr != nil {
				return nil, validateErr
			}
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, s.checkSaveErr(saveErr, "entity")
		}
		return saved, nil
	})
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntity(ctx context.Context, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		entity, queryErr := tx.SystemAnalysisEntity.Get(ctx, id)
		if queryErr != nil {
			return fmt.Errorf("get system analysis entity: %w", queryErr)
		}
		deleteRelationships := tx.SystemAnalysisRelationship.Delete().Where(
			sarel.AnalysisID(entity.AnalysisID),
			sarel.HasKnowledgeRelationshipWith(knr.Or(
				knr.SourceEntityID(entity.KnowledgeEntityID),
				knr.TargetEntityID(entity.KnowledgeEntityID),
			)),
		)
		if _, deleteErr := deleteRelationships.Exec(ctx); deleteErr != nil {
			return fmt.Errorf("delete connected system analysis relationships: %w", deleteErr)
		}
		if deleteEntityErr := tx.SystemAnalysisEntity.DeleteOneID(id).Exec(ctx); deleteEntityErr != nil {
			return fmt.Errorf("delete system analysis entity: %w", deleteEntityErr)
		}
		return nil
	})
}

func (s *SystemAnalysisService) ListSystemAnalysisRelationships(ctx context.Context, params rez.ListSystemAnalysisRelationshipsParams) (*ent.ListResult[ent.SystemAnalysisRelationship], error) {
	query := s.db.Client(ctx).SystemAnalysisRelationship.Query().
		Where(params.Predicates...)
	if len(params.OrderBy) > 0 {
		query.Order(params.OrderBy...)
	} else {
		query.Order(sarel.ByCreatedAt(), sarel.ByID())
	}
	query.WithKnowledgeRelationship(func(rq *ent.KnowledgeRelationshipQuery) {
		rq.WithAliases(subjectAliasWithEvidence())
		if params.WithEndpoints {
			rq.WithSourceEntity(func(eq *ent.KnowledgeEntityQuery) {
				eq.WithAliases(subjectAliasWithEvidence())
			}).WithTargetEntity(func(eq *ent.KnowledgeEntityQuery) {
				eq.WithAliases(subjectAliasWithEvidence())
			})
		}
	})
	return ent.DoListQuery[ent.SystemAnalysisRelationship, *ent.SystemAnalysisRelationshipQuery](ctx, query, params.ListParams)
}

func (s *SystemAnalysisService) SetSystemAnalysisRelationship(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisRelationshipMutation)) (*ent.SystemAnalysisRelationship, error) {
	isCreate := id == uuid.Nil
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.SystemAnalysisRelationship, error) {
		var mutator ent.EntityMutator[*ent.SystemAnalysisRelationship, *ent.SystemAnalysisRelationshipMutation]
		if isCreate {
			mutator = tx.SystemAnalysisRelationship.Create()
		} else {
			mutator = tx.SystemAnalysisRelationship.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)

		if mutErr := s.validateMutationFields(mut, sarel.FieldAnalysisID, sarel.FieldKnowledgeRelationshipID); mutErr != nil {
			return nil, mutErr
		}
		var relationshipEntityIDs []uuid.UUID
		if isCreate {
			analysisID, _ := mut.AnalysisID()
			if validateErr := s.validateSystemAnalysisTarget(ctx, tx, analysisID); validateErr != nil {
				return nil, validateErr
			}
			relationshipID, _ := mut.KnowledgeRelationshipID()
			var resolveErr error
			relationshipEntityIDs, resolveErr = s.resolveRelationshipEntityIDs(ctx, tx, mapset.NewSet(relationshipID))
			if resolveErr != nil {
				return nil, fmt.Errorf("resolve relationship entity IDs: %w", resolveErr)
			}
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, s.checkSaveErr(saveErr, "relationship")
		}

		if isCreate {
			includeRelEntsErr := s.includeAnalysisEntities(ctx, tx, saved.AnalysisID, mapset.NewSet(relationshipEntityIDs...))
			if includeRelEntsErr != nil {
				return nil, fmt.Errorf("include relationship entities: %w", includeRelEntsErr)
			}
		}

		return saved, nil
	})
}

func (s *SystemAnalysisService) DeleteSystemAnalysisRelationship(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).SystemAnalysisRelationship.DeleteOneID(id).Exec(ctx)
}

func (s *SystemAnalysisService) ListSystemAnalysisEntries(ctx context.Context, params rez.ListSystemAnalysisEntriesParams) (*ent.ListResult[ent.SystemAnalysisEntry], error) {
	query := s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(params.Predicates...)
	if len(params.OrderBy) > 0 {
		query.Order(params.OrderBy...)
	} else {
		query.Order(sae.ByOccurredAt(), sae.BySequence(), sae.ByID())
	}
	if !params.SummaryOnly {
		query.WithReviews().WithSubjects(s.systemAnalysisEntrySubjectsQuery)
	}
	return ent.DoListQuery[ent.SystemAnalysisEntry, *ent.SystemAnalysisEntryQuery](ctx, query, params.ListParams)
}

const systemAnalysisWriteLock = "system_analysis_write"

var (
	systemAnalysisEntrySubjectImmutableFields = []string{
		saes.FieldKnowledgeEntityID,
		saes.FieldKnowledgeRelationshipID,
		saes.FieldKnowledgeEvidenceID,
	}
)

func (s *SystemAnalysisService) SetSystemAnalysisEntry(ctx context.Context, id uuid.UUID, params rez.SetSystemAnalysisEntryParams) (*ent.SystemAnalysisEntry, error) {
	isCreate := id == uuid.Nil
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.SystemAnalysisEntry, error) {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntry, *ent.SystemAnalysisEntryMutation]
		if isCreate {
			mutator = tx.SystemAnalysisEntry.Create()
		} else {
			current, queryErr := tx.SystemAnalysisEntry.Get(ctx, id)
			if queryErr != nil {
				return nil, queryErr
			}
			if params.AnalysisID != uuid.Nil && current.AnalysisID != params.AnalysisID {
				return nil, fmt.Errorf("%w: analysis_id cannot be changed", rez.ErrInvalidInput)
			}
			if params.Reference != nil && (current.Reference == nil || *current.Reference != *params.Reference) {
				return nil, fmt.Errorf("%w: reference cannot be changed", rez.ErrInvalidInput)
			}
			mutator = tx.SystemAnalysisEntry.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		if isCreate {
			mut.SetAnalysisID(params.AnalysisID)
			if params.Reference != nil {
				mut.SetReference(*params.Reference)
			}
		}
		mut.SetKind(params.Kind)
		mut.SetTitle(params.Title)
		mut.SetBody(params.Body)
		if params.OccurredAt != nil {
			mut.SetOccurredAt(*params.OccurredAt)
		}

		if mutErr := s.validateMutationFields(mut, sae.FieldAnalysisID); mutErr != nil {
			return nil, mutErr
		}

		occurredAt, hasOccurredAt := mut.OccurredAt()
		sequence := 1
		if isCreate && hasOccurredAt {
			analysisID, _ := mut.AnalysisID()
			if lockErr := s.db.AcquireTxLocks(ctx, systemAnalysisWriteLock, analysisID.String()); lockErr != nil {
				return nil, fmt.Errorf("lock system analysis: %w", lockErr)
			}

			querySameTime := tx.SystemAnalysisEntry.Query().
				Where(sae.AnalysisID(analysisID), sae.OccurredAt(occurredAt))
			count, countErr := querySameTime.Count(ctx)
			if countErr != nil && !ent.IsNotFound(countErr) {
				return nil, fmt.Errorf("query same entry times: %w", countErr)
			}
			sequence = count + 1
		}
		if isCreate {
			mut.SetSequence(sequence)
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, s.checkSaveErr(saveErr, "analysis entry")
		}

		if len(params.SetSubjects) > 0 {
			setSubjects := make([]func(*ent.SystemAnalysisEntrySubjectMutation) error, len(params.SetSubjects))
			for i, ss := range params.SetSubjects {
				setSubjects[i] = func(m *ent.SystemAnalysisEntrySubjectMutation) error {
					return s.mutateEntrySubjectWithParams(m, ss)
				}
			}
			var subjectMutErr error
			upsertSubjects := tx.SystemAnalysisEntrySubject.
				MapCreateBulk(setSubjects, func(c *ent.SystemAnalysisEntrySubjectCreate, i int) {
					c.SetEntryID(saved.ID)
					subjectMutErr = errors.Join(subjectMutErr, setSubjects[i](c.Mutation()))
				}).
				OnConflict().
				DoNothing()
			if subjectMutErr != nil {
				return nil, subjectMutErr
			}
			if saveSubjectsErr := upsertSubjects.Exec(ctx); saveSubjectsErr != nil {
				return nil, s.checkSaveErr(saveSubjectsErr, "subjects")
			}
		}

		entry, queryErr := s.LookupSystemAnalysisEntry(ctx, sae.ID(saved.ID))
		if queryErr != nil {
			return nil, queryErr
		}
		return entry, nil
	})
}

func (s *SystemAnalysisService) validateMutationFields(m ent.Mutation, fields ...string) error {
	isCreate := m.Op() == ent.OpCreate
	for _, name := range fields {
		v, isSet := m.Field(name)
		if isCreate {
			if !isSet {
				return fmt.Errorf("%w: %s is required", rez.ErrInvalidInput, name)
			}
			if id, ok := v.(uuid.UUID); !ok || id == uuid.Nil {
				return fmt.Errorf("%w: %s is invalid uuid", rez.ErrInvalidInput, name)
			}
		} else {
			if isSet || m.FieldCleared(name) {
				return fmt.Errorf("%w: %s cannot be changed", rez.ErrInvalidInput, name)
			}
		}
	}
	return nil
}

func (s *SystemAnalysisService) LookupSystemAnalysisEntry(ctx context.Context, pred predicate.SystemAnalysisEntry) (*ent.SystemAnalysisEntry, error) {
	return s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(pred).
		WithReviews().
		WithSubjects(s.systemAnalysisEntrySubjectsQuery).
		Only(ctx)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntry(ctx context.Context, id uuid.UUID) error {
	// cascade delete
	return s.db.Client(ctx).SystemAnalysisEntry.DeleteOneID(id).Exec(ctx)
}

func (s *SystemAnalysisService) ListSystemAnalysisEntrySubjects(ctx context.Context, params rez.ListSystemAnalysisEntrySubjectsParams) (*ent.ListResult[ent.SystemAnalysisEntrySubject], error) {
	if params.AnalysisID == uuid.Nil || params.EntryID == uuid.Nil {
		return nil, fmt.Errorf("%w: analysis and entry IDs are required", rez.ErrInvalidInput)
	}

	entryExists, queryEntryErr := s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.ID(params.EntryID), sae.AnalysisID(params.AnalysisID)).
		Exist(ctx)
	if queryEntryErr != nil {
		return nil, fmt.Errorf("check analysis entry access: %w", queryEntryErr)
	}
	if !entryExists {
		return nil, rez.ErrNotFound
	}

	query := s.db.Client(ctx).SystemAnalysisEntrySubject.Query().
		Where(saes.EntryID(params.EntryID)).
		Order(saes.ByID(sql.OrderAsc()))
	return ent.DoListQuery[ent.SystemAnalysisEntrySubject, *ent.SystemAnalysisEntrySubjectQuery](ctx, query, params.ListParams)
}

func (s *SystemAnalysisService) SetSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntrySubjectMutation)) (*ent.SystemAnalysisEntrySubject, error) {
	isCreate := id == uuid.Nil
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.SystemAnalysisEntrySubject, error) {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntrySubject, *ent.SystemAnalysisEntrySubjectMutation]
		if isCreate {
			mutator = tx.SystemAnalysisEntrySubject.Create()
		} else {
			mutator = tx.SystemAnalysisEntrySubject.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)

		if validErr := s.validateEntrySubjectMutation(mut); validErr != nil {
			return nil, validErr
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, s.checkSaveErr(saveErr, "entry subject")
		}
		return saved, nil
	})
}

func (s *SystemAnalysisService) validateEntrySubjectMutation(m *ent.SystemAnalysisEntrySubjectMutation) error {
	if entryErr := s.validateMutationFields(m, saes.FieldEntryID); entryErr != nil {
		return entryErr
	}
	if m.Op() != ent.OpCreate {
		return s.validateMutationFields(m, systemAnalysisEntrySubjectImmutableFields...)
	}
	count := 0
	for _, name := range systemAnalysisEntrySubjectImmutableFields {
		if value, isSet := m.Field(name); isSet {
			id, validUUID := value.(uuid.UUID)
			if !validUUID || id == uuid.Nil {
				return fmt.Errorf("%w: %s is invalid uuid", rez.ErrInvalidInput, name)
			}
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("%w: exactly one graph reference is required", rez.ErrUnprocessableInput)
	}
	return nil
}

func (s *SystemAnalysisService) mutateEntrySubjectWithParams(m *ent.SystemAnalysisEntrySubjectMutation, p rez.SetSystemAnalysisEntrySubjectParams) error {
	m.SetRole(p.Role)
	if p.KnowledgeRelationshipID != nil {
		m.SetKnowledgeRelationshipID(*p.KnowledgeRelationshipID)
	}
	if p.KnowledgeEntityID != nil {
		m.SetKnowledgeEntityID(*p.KnowledgeEntityID)
	}
	if p.KnowledgeEvidenceID != nil {
		m.SetKnowledgeEvidenceID(*p.KnowledgeEvidenceID)
	}
	return s.validateEntrySubjectMutation(m)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).SystemAnalysisEntrySubject.DeleteOneID(id).Exec(ctx)
}
