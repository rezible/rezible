package db

import (
	"context"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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
	knowledge rez.KnowledgeGraphService
}

func NewSystemAnalysisService(db rez.Database, knowledge rez.KnowledgeGraphService) (*SystemAnalysisService, error) {
	return &SystemAnalysisService{db: db, knowledge: knowledge}, nil
}

func (s *SystemAnalysisService) systemAnalysisEntrySubjectsQuery(q *ent.SystemAnalysisEntrySubjectQuery) {
	q.Order(ent.Asc(saes.FieldCreatedAt), ent.Asc(saes.FieldID))
}

func (s *SystemAnalysisService) systemAnalysisEntitiesQuery(q *ent.SystemAnalysisEntityQuery) {
	q.Order(ent.Asc(saent.FieldCreatedAt), ent.Asc(saent.FieldID)).
		WithKnowledgeEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases(knowledgeAliasWithEvidenceQuery())
		})
}

func (s *SystemAnalysisService) systemAnalysisRelationshipsQuery(q *ent.SystemAnalysisRelationshipQuery) {
	q.Order(ent.Asc(sarel.FieldCreatedAt), ent.Asc(sarel.FieldID)).
		WithKnowledgeRelationship(func(rq *ent.KnowledgeRelationshipQuery) {
			rq.WithAliases(knowledgeAliasWithEvidenceQuery())
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
	var mutator ent.EntityMutator[*ent.SystemAnalysis, *ent.SystemAnalysisMutation]
	if id == uuid.Nil {
		mutator = s.db.Client(ctx).SystemAnalysis.Create()
	} else {
		mutator = s.db.Client(ctx).SystemAnalysis.UpdateOneID(id)
	}

	setFn(mutator.Mutation())

	saved, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, s.checkSaveErr(saveErr, "system analysis")
	}
	return saved, nil
}

func (s *SystemAnalysisService) GetSystemAnalysisGraph(ctx context.Context, id uuid.UUID, params rez.GetKnowledgeGraphViewParams) (*rez.KnowledgeGraphView, error) {
	query := s.db.Client(ctx).SystemAnalysis.Query().
		Where(sa.ID(id))
	analysis, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get system analysis: %w", queryErr)
	}

	rootID := uuid.Nil
	if analysis.SubjectEntityID != nil {
		rootID = *analysis.SubjectEntityID
	}
	params.EntityID = rootID

	view, viewErr := s.knowledge.GetView(ctx, params)
	if viewErr != nil {
		return nil, fmt.Errorf("get system analysis graph: %w", viewErr)
	}
	return view, nil
}

func (s *SystemAnalysisService) IncludeSystemAnalysisSubjects(ctx context.Context, params rez.IncludeSystemAnalysisSubjectsParams) error {
	if params.AnalysisId == uuid.Nil || (len(params.EntityIds)+len(params.RelationshipIds) == 0) {
		return fmt.Errorf("%w: system analysis ID is required", rez.ErrInvalidInput)
	}

	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		entitySubjIds := mapset.NewSet(params.EntityIds...)
		if len(params.RelationshipIds) > 0 {
			relEntIds, relEntsErr := s.resolveRelationshipEntityIDs(ctx, mapset.NewSet(params.RelationshipIds...))
			if relEntsErr != nil {
				return relEntsErr
			}
			entitySubjIds.Append(relEntIds...)

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
			if includeErr := s.includeAnalysisEntities(ctx, params.AnalysisId, entitySubjIds); includeErr != nil {
				return fmt.Errorf("include entities: %w", includeErr)
			}
		}

		return nil
	})
}

func (s *SystemAnalysisService) resolveRelationshipEntityIDs(ctx context.Context, relIds mapset.Set[uuid.UUID]) ([]uuid.UUID, error) {
	if relIds.IsEmpty() {
		return nil, nil
	}

	query := s.db.Client(ctx).KnowledgeRelationship.Query().
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
	return entityIDs.ToSlice(), nil
}

func (s *SystemAnalysisService) includeAnalysisEntities(ctx context.Context, analysisID uuid.UUID, entityIds mapset.Set[uuid.UUID]) error {
	if entityIds.IsEmpty() {
		return nil
	}

	ids := entityIds.ToSlice()
	upsert := s.db.Client(ctx).SystemAnalysisEntity.
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
	s.systemAnalysisEntitiesQuery(query)
	return ent.DoListQuery[ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityQuery](ctx, query, params.ListParams)
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

func (s *SystemAnalysisService) SetSystemAnalysisEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntityMutation)) (*ent.SystemAnalysisEntity, error) {
	var mutator ent.EntityMutator[*ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityMutation]
	if id == uuid.Nil {
		mutator = s.db.Client(ctx).SystemAnalysisEntity.Create()
	} else {
		mutator = s.db.Client(ctx).SystemAnalysisEntity.UpdateOneID(id)
	}

	mut := mutator.Mutation()
	setFn(mut)

	if mutErr := s.validateMutationFields(mut, saent.FieldAnalysisID, saent.FieldKnowledgeEntityID); mutErr != nil {
		return nil, mutErr
	}

	saved, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, s.checkSaveErr(saveErr, "entity")
	}
	return saved, nil
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
	s.systemAnalysisRelationshipsQuery(query)
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

		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return s.checkSaveErr(saveErr, "relationship")
		}

		if isCreate {
			relEntityIDs, relEntsErr := s.resolveRelationshipEntityIDs(ctx, mapset.NewSet(saved.KnowledgeRelationshipID))
			if relEntsErr != nil {
				return fmt.Errorf("resolve relationship entity IDs: %w", relEntsErr)
			}

			includeRelEntsErr := s.includeAnalysisEntities(ctx, saved.AnalysisID, mapset.NewSet(relEntityIDs...))
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
		Where(params.Predicates...).
		Order(ent.Asc(sae.FieldOccurredAt), ent.Asc(sae.FieldSequence), ent.Asc(sae.FieldID)).
		WithReviews().
		WithSubjects(s.systemAnalysisEntrySubjectsQuery)
	return ent.DoListQuery[ent.SystemAnalysisEntry, *ent.SystemAnalysisEntryQuery](ctx, query, params.ListParams)
}

const systemAnalysisWriteLock = "system_analysis_write"

var systemAnalysisEntrySubjectImmutableFields = []string{
	saes.FieldKnowledgeEntityID,
	saes.FieldKnowledgeRelationshipID,
	saes.FieldKnowledgeEvidenceID,
}

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
		mut.SetSequence(sequence)

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
				})
			if saveSubjectsErr := upsertSubjects.Exec(ctx); saveSubjectsErr != nil {
				return s.checkSaveErr(saveSubjectsErr, "subjects")
			}
		}

		entry, queryErr := s.GetSystemAnalysisEntry(ctx, saved.ID)
		if queryErr != nil {
			return queryErr
		}
		result = entry.Unwrap()
		return nil
	})
}

func (s *SystemAnalysisService) GetSystemAnalysisEntry(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysisEntry, error) {
	return s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.ID(id)).
		WithReviews().
		WithSubjects().
		Only(ctx)
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
	var mutator ent.EntityMutator[*ent.SystemAnalysisEntrySubject, *ent.SystemAnalysisEntrySubjectMutation]

	allowedTargets := 1
	if id == uuid.Nil {
		mutator = s.db.Client(ctx).SystemAnalysisEntrySubject.Create()
	} else {
		allowedTargets = 0
		mutator = s.db.Client(ctx).SystemAnalysisEntrySubject.UpdateOneID(id)
	}

	mut := mutator.Mutation()
	setFn(mut)

	if mutErr := s.validateMutationFields(mut, saes.FieldEntryID); mutErr != nil {
		return nil, mutErr
	}
	var numTargetsChanged int
	if _, entIdSet := mut.KnowledgeEntityID(); entIdSet || mut.KnowledgeEntityCleared() {
		numTargetsChanged++
	}
	if _, relIdSet := mut.KnowledgeRelationshipID(); relIdSet || mut.KnowledgeRelationshipCleared() {
		numTargetsChanged++
	}
	if _, evIdSet := mut.KnowledgeEvidenceID(); evIdSet || mut.KnowledgeEvidenceCleared() {
		numTargetsChanged++
	}
	if _, normalizedEventIDSet := mut.NormalizedEventID(); normalizedEventIDSet || mut.NormalizedEventCleared() {
		numTargetsChanged++
	}
	if numTargetsChanged != allowedTargets {
		return nil, fmt.Errorf("%w: only one subject id may be set", rez.ErrInvalidInput)
	}

	saved, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, s.checkSaveErr(saveErr, "entry subject")
	}
	return saved, nil
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID) error {
	deleteSubject := s.db.Client(ctx).SystemAnalysisEntrySubject.DeleteOneID(id)
	if deleteErr := deleteSubject.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete entry subject: %w", deleteErr)
	}
	return nil
}
