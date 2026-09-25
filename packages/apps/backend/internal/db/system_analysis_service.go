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
	q.Order(ent.Asc(saes.FieldCreatedAt), ent.Asc(saes.FieldID))
}

func (s *SystemAnalysisService) systemAnalysisEntitiesQuery(q *ent.SystemAnalysisEntityQuery) {
	q.Order(ent.Asc(saent.FieldCreatedAt), ent.Asc(saent.FieldID)).
		WithKnowledgeEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases(subjectAliasWithEvidence())
		})
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
	var result *ent.SystemAnalysis
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
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
			return validateErr
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return s.checkSaveErr(saveErr, "system analysis")
		}
		result = saved
		return nil
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
		Where(params.Predicates...)
	switch params.Order {
	case rez.SystemAnalysisSubjectOrderCreatedAt:
		s.systemAnalysisEntitiesQuery(query)
	case rez.SystemAnalysisSubjectOrderCanonicalID:
		query.Order(saent.ByKnowledgeEntityID(sql.OrderAsc()), saent.ByID(sql.OrderAsc())).
			WithKnowledgeEntity(func(eq *ent.KnowledgeEntityQuery) {
				eq.WithAliases(subjectAliasWithEvidence())
			})
	default:
		return nil, fmt.Errorf("%w: invalid system analysis subject order %d", rez.ErrInvalidInput, params.Order)
	}
	return ent.DoListQuery[ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityQuery](ctx, query, params.ListParams)
}

func (s *SystemAnalysisService) SetSystemAnalysisEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntityMutation)) (*ent.SystemAnalysisEntity, error) {
	var result *ent.SystemAnalysisEntity
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysisEntity.Create()
		} else {
			mutator = tx.SystemAnalysisEntity.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)

		if mutErr := s.validateMutationFields(mut, saent.FieldAnalysisID, saent.FieldKnowledgeEntityID); mutErr != nil {
			return mutErr
		}
		if id == uuid.Nil {
			analysisID, _ := mut.AnalysisID()
			if validateErr := s.validateSystemAnalysisTarget(ctx, tx, analysisID); validateErr != nil {
				return validateErr
			}
			entityID, _ := mut.KnowledgeEntityID()
			if validateErr := s.validateKnowledgeEntityTargets(ctx, tx, mapset.NewSet(entityID)); validateErr != nil {
				return validateErr
			}
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return s.checkSaveErr(saveErr, "entity")
		}
		result = saved
		return nil
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
	switch params.Order {
	case rez.SystemAnalysisSubjectOrderCreatedAt:
		query.Order(ent.Asc(sarel.FieldCreatedAt), ent.Asc(sarel.FieldID))
	case rez.SystemAnalysisSubjectOrderCanonicalID:
		query.Order(sarel.ByKnowledgeRelationshipID(sql.OrderAsc()), sarel.ByID(sql.OrderAsc()))
	default:
		return nil, fmt.Errorf("%w: invalid system analysis subject order %d", rez.ErrInvalidInput, params.Order)
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
	var result *ent.SystemAnalysisRelationship
	isCreate := id == uuid.Nil
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisRelationship, *ent.SystemAnalysisRelationshipMutation]
		if isCreate {
			mutator = tx.SystemAnalysisRelationship.Create()
		} else {
			mutator = tx.SystemAnalysisRelationship.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)

		if mutErr := s.validateMutationFields(mut, sarel.FieldAnalysisID, sarel.FieldKnowledgeRelationshipID); mutErr != nil {
			return mutErr
		}
		var relationshipEntityIDs []uuid.UUID
		if isCreate {
			analysisID, _ := mut.AnalysisID()
			if validateErr := s.validateSystemAnalysisTarget(ctx, tx, analysisID); validateErr != nil {
				return validateErr
			}
			relationshipID, _ := mut.KnowledgeRelationshipID()
			var resolveErr error
			relationshipEntityIDs, resolveErr = s.resolveRelationshipEntityIDs(ctx, tx, mapset.NewSet(relationshipID))
			if resolveErr != nil {
				return fmt.Errorf("resolve relationship entity IDs: %w", resolveErr)
			}
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return s.checkSaveErr(saveErr, "relationship")
		}

		if isCreate {
			includeRelEntsErr := s.includeAnalysisEntities(ctx, tx, saved.AnalysisID, mapset.NewSet(relationshipEntityIDs...))
			if includeRelEntsErr != nil {
				return fmt.Errorf("include relationship entities: %w", includeRelEntsErr)
			}
		}

		result = saved.Unwrap()
		return nil
	})
}

func (s *SystemAnalysisService) DeleteSystemAnalysisRelationship(ctx context.Context, id uuid.UUID) error {
	deleteRelationship := s.db.Client(ctx).SystemAnalysisRelationship.DeleteOneID(id)
	if deleteErr := deleteRelationship.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete system analysis relationship: %w", deleteErr)
	}
	return nil
}

func (s *SystemAnalysisService) ListSystemAnalysisEntries(ctx context.Context, params rez.ListSystemAnalysisEntriesParams) (*ent.ListResult[ent.SystemAnalysisEntry], error) {
	query := s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(params.Predicates...)
	switch params.Order {
	case rez.SystemAnalysisEntryOrderOccurrence:
		query.Order(ent.Asc(sae.FieldOccurredAt), ent.Asc(sae.FieldSequence), ent.Asc(sae.FieldID))
	case rez.SystemAnalysisEntryOrderSequence:
		query.Order(sae.BySequence(sql.OrderAsc()), sae.ByID(sql.OrderAsc()))
	default:
		return nil, fmt.Errorf("%w: invalid system analysis entry order %d", rez.ErrInvalidInput, params.Order)
	}
	if !params.SummaryOnly {
		query.WithReviews().WithSubjects(s.systemAnalysisEntrySubjectsQuery)
	}
	return ent.DoListQuery[ent.SystemAnalysisEntry, *ent.SystemAnalysisEntryQuery](ctx, query, params.ListParams)
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

const systemAnalysisWriteLock = "system_analysis_write"

var (
	systemAnalysisEntrySubjectImmutableFields = []string{
		saes.FieldKnowledgeEntityID,
		saes.FieldKnowledgeRelationshipID,
		saes.FieldKnowledgeEvidenceID,
	}
)

func (s *SystemAnalysisService) SetSystemAnalysisEntry(
	ctx context.Context,
	id uuid.UUID,
	setEntry func(*ent.SystemAnalysisEntryMutation),
	setSubjects ...func(*ent.SystemAnalysisEntrySubjectMutation),
) (*ent.SystemAnalysisEntry, error) {
	isCreate := id == uuid.Nil
	var result *ent.SystemAnalysisEntry
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntry, *ent.SystemAnalysisEntryMutation]
		if isCreate {
			mutator = tx.SystemAnalysisEntry.Create()
		} else {
			mutator = tx.SystemAnalysisEntry.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setEntry(mut)

		if mutErr := s.validateMutationFields(mut, sae.FieldAnalysisID); mutErr != nil {
			return mutErr
		}

		occurredAt, hasOccurredAt := mut.OccurredAt()
		sequence := 1
		if isCreate && hasOccurredAt {
			analysisID, _ := mut.AnalysisID()
			if lockErr := s.db.AcquireTxLocks(ctx, systemAnalysisWriteLock, analysisID.String()); lockErr != nil {
				return fmt.Errorf("lock system analysis: %w", lockErr)
			}

			querySameTime := tx.SystemAnalysisEntry.Query().
				Where(sae.AnalysisID(analysisID), sae.OccurredAt(occurredAt))
			count, countErr := querySameTime.Count(ctx)
			if countErr != nil && !ent.IsNotFound(countErr) {
				return fmt.Errorf("query same entry times: %w", countErr)
			}
			sequence = count + 1
		}
		if isCreate {
			mut.SetSequence(sequence)
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return s.checkSaveErr(saveErr, "analysis entry")
		}

		if len(setSubjects) > 0 {
			var subjectMutErr error
			upsertSubjects := tx.SystemAnalysisEntrySubject.
				MapCreateBulk(setSubjects, func(c *ent.SystemAnalysisEntrySubjectCreate, i int) {
					c.SetEntryID(saved.ID)
					cm := c.Mutation()
					setSubjects[i](cm)
					subjectMutErr = errors.Join(subjectMutErr,
						s.validateMutationFields(cm, systemAnalysisEntrySubjectImmutableFields...))
				}).
				OnConflict().
				DoNothing()
			if saveSubjectsErr := upsertSubjects.Exec(ctx); saveSubjectsErr != nil {
				return s.checkSaveErr(saveSubjectsErr, "subjects")
			}
		}

		entry, queryErr := s.LookupSystemAnalysisEntry(ctx, sae.ID(saved.ID))
		if queryErr != nil {
			return queryErr
		}
		result = entry.Unwrap()
		return nil
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
		WithSubjects().
		Only(ctx)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntry(ctx context.Context, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		deleteSubjects := tx.SystemAnalysisEntrySubject.Delete().
			Where(saes.EntryID(id))
		if _, subjectsErr := deleteSubjects.Exec(ctx); subjectsErr != nil {
			return subjectsErr
		}
		return tx.SystemAnalysisEntry.DeleteOneID(id).Exec(ctx)
	})
}

func (s *SystemAnalysisService) SetSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntrySubjectMutation)) (*ent.SystemAnalysisEntrySubject, error) {
	isCreate := id == uuid.Nil
	var result *ent.SystemAnalysisEntrySubject
	return result, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntrySubject, *ent.SystemAnalysisEntrySubjectMutation]
		if isCreate {
			mutator = tx.SystemAnalysisEntrySubject.Create()
		} else {
			mutator = tx.SystemAnalysisEntrySubject.UpdateOneID(id)
		}

		mut := mutator.Mutation()
		setFn(mut)

		mutFields := []string{saes.FieldEntryID}
		if !isCreate {
			mutFields = append(mutFields, systemAnalysisEntrySubjectImmutableFields...)
		} else {
		}
		if mutErr := s.validateMutationFields(mut, mutFields...); mutErr != nil {
			return mutErr
		}

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return s.checkSaveErr(saveErr, "entry subject")
		}
		result = saved.Unwrap()
		return nil
	})
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID) error {
	deleteSubject := s.db.Client(ctx).SystemAnalysisEntrySubject.DeleteOneID(id)
	if deleteErr := deleteSubject.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete entry subject: %w", deleteErr)
	}
	return nil
}
