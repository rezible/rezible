package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	sa "github.com/rezible/rezible/ent/systemanalysis"
	saentity "github.com/rezible/rezible/ent/systemanalysisentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sarel "github.com/rezible/rezible/ent/systemanalysisrelationship"
)

type SystemAnalysisService struct {
	db        rez.Database
	knowledge rez.KnowledgeGraphService
}

func NewSystemAnalysisService(db rez.Database, knowledge rez.KnowledgeGraphService) (*SystemAnalysisService, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	if knowledge == nil {
		return nil, fmt.Errorf("knowledge graph service is required")
	}
	return &SystemAnalysisService{db: db, knowledge: knowledge}, nil
}

func (s *SystemAnalysisService) systemAnalysisEntriesQuery(q *ent.SystemAnalysisEntryQuery) {
	q.Order(ent.Asc(sae.FieldSequence), ent.Asc(sae.FieldOccurredAt), ent.Asc(sae.FieldID)).
		WithSubjects(s.systemAnalysisEntrySubjectsQuery)
}

func (s *SystemAnalysisService) systemAnalysisEntrySubjectsQuery(q *ent.SystemAnalysisEntrySubjectQuery) {
	q.Order(ent.Asc(saes.FieldCreatedAt), ent.Asc(saes.FieldID)).
		WithKnowledgeEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases(knowledgeAliasWithEvidenceQuery())
		}).
		WithKnowledgeRelationship(func(rq *ent.KnowledgeRelationshipQuery) {
			rq.WithAliases(knowledgeAliasWithEvidenceQuery())
		}).
		WithKnowledgeEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.WithEvent()
			eq.WithSubjectAlias()
		})
}

func (s *SystemAnalysisService) systemAnalysisEntitiesQuery(q *ent.SystemAnalysisEntityQuery) {
	q.Order(ent.Asc(saentity.FieldCreatedAt), ent.Asc(saentity.FieldID)).
		WithKnowledgeEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases(knowledgeAliasWithEvidenceQuery())
		})
}

func (s *SystemAnalysisService) systemAnalysisRelationshipsQuery(q *ent.SystemAnalysisRelationshipQuery) {
	q.Order(ent.Asc(sarel.FieldCreatedAt), ent.Asc(sarel.FieldID)).
		WithKnowledgeRelationship(func(rq *ent.KnowledgeRelationshipQuery) {
			rq.WithAliases(knowledgeAliasWithEvidenceQuery())
		}).
		WithSourceEntity(s.systemAnalysisEntitiesQuery).
		WithTargetEntity(s.systemAnalysisEntitiesQuery)
}

func (s *SystemAnalysisService) GetSystemAnalysis(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysis, error) {
	query := s.db.Client(ctx).SystemAnalysis.Query().
		Where(sa.ID(id)).
		WithAnalysisEntities(s.systemAnalysisEntitiesQuery).
		WithAnalysisRelationships(s.systemAnalysisRelationshipsQuery).
		WithEntries(s.systemAnalysisEntriesQuery)
	analysis, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get system analysis: %w", queryErr)
	}
	return analysis, nil
}

func (s *SystemAnalysisService) SetSystemAnalysis(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisMutation)) (*ent.SystemAnalysis, error) {
	var savedID uuid.UUID
	saveErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysis, *ent.SystemAnalysisMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysis.Create()
		} else {
			mutator = tx.SystemAnalysis.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		saved, err := mutator.Save(ctx)
		if err != nil {
			if ent.IsValidationError(err) {
				return fmt.Errorf("%w: save system analysis: %w", rez.ErrInvalidInput, err)
			}
			return fmt.Errorf("save system analysis: %w", err)
		}
		savedID = saved.Unwrap().ID
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	return s.GetSystemAnalysis(ctx, savedID)
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
	} else if analysis.ScopeEntityID != nil {
		rootID = *analysis.ScopeEntityID
	}
	params.EntityID = rootID

	view, viewErr := s.knowledge.GetView(ctx, params)
	if viewErr != nil {
		return nil, fmt.Errorf("get system analysis graph: %w", viewErr)
	}
	return view, nil
}

func (s *SystemAnalysisService) ListSystemAnalysisEntities(ctx context.Context, analysisID uuid.UUID) (ent.SystemAnalysisEntities, error) {
	query := s.db.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysisID))
	s.systemAnalysisEntitiesQuery(query)
	entities, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list system analysis entities: %w", queryErr)
	}
	return entities, nil
}

func (s *SystemAnalysisService) getSystemAnalysisEntity(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysisEntity, error) {
	query := s.db.Client(ctx).SystemAnalysisEntity.Query().
		Where(saentity.ID(id))
	s.systemAnalysisEntitiesQuery(query)
	entity, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get system analysis entity: %w", queryErr)
	}
	return entity, nil
}

func (s *SystemAnalysisService) validateSystemAnalysisEntityMutation(id uuid.UUID, m *ent.SystemAnalysisEntityMutation) error {
	if id != uuid.Nil {
		_, analysisSet := m.AnalysisID()
		_, knowledgeEntitySet := m.KnowledgeEntityID()
		if analysisSet ||
			knowledgeEntitySet ||
			m.AnalysisCleared() ||
			m.KnowledgeEntityCleared() {
			return fmt.Errorf("%w: system analysis entity target cannot be changed", rez.ErrInvalidInput)
		}
		return nil
	}

	analysisID, analysisSet := m.AnalysisID()
	if !analysisSet || analysisID == uuid.Nil {
		return fmt.Errorf("%w: system analysis entity analysis is required", rez.ErrInvalidInput)
	}
	knowledgeEntityID, knowledgeEntitySet := m.KnowledgeEntityID()
	if !knowledgeEntitySet || knowledgeEntityID == uuid.Nil {
		return fmt.Errorf("%w: system analysis entity knowledge entity is required", rez.ErrInvalidInput)
	}
	return nil
}

func (s *SystemAnalysisService) SetSystemAnalysisEntity(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntityMutation)) (*ent.SystemAnalysisEntity, error) {
	var savedID uuid.UUID
	saveErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntity, *ent.SystemAnalysisEntityMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysisEntity.Create()
		} else {
			mutator = tx.SystemAnalysisEntity.UpdateOneID(id)
		}

		mutation := mutator.Mutation()
		setFn(mutation)
		if validationErr := s.validateSystemAnalysisEntityMutation(id, mutation); validationErr != nil {
			return validationErr
		}

		saved, err := mutator.Save(ctx)
		if err != nil {
			if ent.IsValidationError(err) || ent.IsConstraintError(err) {
				return fmt.Errorf("%w: save system analysis entity: %w", rez.ErrInvalidInput, err)
			}
			return fmt.Errorf("save system analysis entity: %w", err)
		}
		savedID = saved.Unwrap().ID
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	return s.getSystemAnalysisEntity(ctx, savedID)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntity(ctx context.Context, id uuid.UUID) error {
	deleteErr := s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		deleteRelationships := tx.SystemAnalysisRelationship.Delete().
			Where(sarel.Or(sarel.SourceAnalysisEntityID(id), sarel.TargetAnalysisEntityID(id)))
		if _, relationshipErr := deleteRelationships.Exec(txCtx); relationshipErr != nil {
			return relationshipErr
		}
		deleteEntity := tx.SystemAnalysisEntity.DeleteOneID(id)
		return deleteEntity.Exec(txCtx)
	})
	if deleteErr != nil {
		return fmt.Errorf("delete system analysis entity: %w", deleteErr)
	}
	return nil
}

func (s *SystemAnalysisService) ListSystemAnalysisRelationships(ctx context.Context, analysisID uuid.UUID) (ent.SystemAnalysisRelationships, error) {
	query := s.db.Client(ctx).SystemAnalysisRelationship.Query().
		Where(sarel.AnalysisID(analysisID))
	s.systemAnalysisRelationshipsQuery(query)
	relationships, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list system analysis relationships: %w", queryErr)
	}
	return relationships, nil
}

func (s *SystemAnalysisService) getSystemAnalysisRelationship(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysisRelationship, error) {
	query := s.db.Client(ctx).SystemAnalysisRelationship.Query().
		Where(sarel.ID(id))
	s.systemAnalysisRelationshipsQuery(query)
	relationship, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get system analysis relationship: %w", queryErr)
	}
	return relationship, nil
}

func (s *SystemAnalysisService) getOrCreateSystemAnalysisEntity(ctx context.Context, tx *ent.Client, analysisID, knowledgeEntityID uuid.UUID) (*ent.SystemAnalysisEntity, error) {
	query := tx.SystemAnalysisEntity.Query().
		Where(saentity.AnalysisID(analysisID), saentity.KnowledgeEntityID(knowledgeEntityID))
	entity, queryErr := query.Only(ctx)
	if queryErr == nil {
		return entity, nil
	}
	if !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query system analysis entity: %w", queryErr)
	}

	create := tx.SystemAnalysisEntity.Create().
		SetAnalysisID(analysisID).
		SetKnowledgeEntityID(knowledgeEntityID)
	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, fmt.Errorf("create system analysis entity: %w", createErr)
	}
	return created, nil
}

func (s *SystemAnalysisService) prepareSystemAnalysisRelationshipMutation(ctx context.Context, tx *ent.Client, id uuid.UUID, m *ent.SystemAnalysisRelationshipMutation) error {
	if id != uuid.Nil {
		_, analysisSet := m.AnalysisID()
		_, knowledgeRelationshipSet := m.KnowledgeRelationshipID()
		_, sourceEntitySet := m.SourceAnalysisEntityID()
		_, targetEntitySet := m.TargetAnalysisEntityID()
		if analysisSet ||
			knowledgeRelationshipSet ||
			sourceEntitySet ||
			targetEntitySet ||
			m.AnalysisCleared() ||
			m.KnowledgeRelationshipCleared() ||
			m.SourceEntityCleared() ||
			m.TargetEntityCleared() {
			return fmt.Errorf("%w: system analysis relationship target cannot be changed", rez.ErrInvalidInput)
		}
		return nil
	}

	analysisID, analysisSet := m.AnalysisID()
	if !analysisSet || analysisID == uuid.Nil {
		return fmt.Errorf("%w: system analysis relationship analysis is required", rez.ErrInvalidInput)
	}
	knowledgeRelationshipID, knowledgeRelationshipSet := m.KnowledgeRelationshipID()
	if !knowledgeRelationshipSet || knowledgeRelationshipID == uuid.Nil {
		return fmt.Errorf("%w: system analysis relationship knowledge relationship is required", rez.ErrInvalidInput)
	}
	if _, sourceEntitySet := m.SourceAnalysisEntityID(); sourceEntitySet || m.SourceEntityCleared() {
		return fmt.Errorf("%w: system analysis relationship source is derived from the knowledge relationship", rez.ErrInvalidInput)
	}
	if _, targetEntitySet := m.TargetAnalysisEntityID(); targetEntitySet || m.TargetEntityCleared() {
		return fmt.Errorf("%w: system analysis relationship target is derived from the knowledge relationship", rez.ErrInvalidInput)
	}

	relationshipQuery := tx.KnowledgeRelationship.Query().
		Where(knr.ID(knowledgeRelationshipID))
	knowledgeRelationship, relationshipErr := relationshipQuery.Only(ctx)
	if relationshipErr != nil {
		return fmt.Errorf("query knowledge relationship: %w", relationshipErr)
	}
	sourceEntity, sourceErr := s.getOrCreateSystemAnalysisEntity(ctx, tx, analysisID, knowledgeRelationship.SourceEntityID)
	if sourceErr != nil {
		return sourceErr
	}
	targetEntity, targetErr := s.getOrCreateSystemAnalysisEntity(ctx, tx, analysisID, knowledgeRelationship.TargetEntityID)
	if targetErr != nil {
		return targetErr
	}
	m.SetSourceAnalysisEntityID(sourceEntity.ID)
	m.SetTargetAnalysisEntityID(targetEntity.ID)
	return nil
}

func (s *SystemAnalysisService) SetSystemAnalysisRelationship(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisRelationshipMutation)) (*ent.SystemAnalysisRelationship, error) {
	var savedID uuid.UUID
	saveErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisRelationship, *ent.SystemAnalysisRelationshipMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysisRelationship.Create()
		} else {
			mutator = tx.SystemAnalysisRelationship.UpdateOneID(id)
		}

		mutation := mutator.Mutation()
		setFn(mutation)
		if prepareErr := s.prepareSystemAnalysisRelationshipMutation(ctx, tx, id, mutation); prepareErr != nil {
			return prepareErr
		}

		saved, err := mutator.Save(ctx)
		if err != nil {
			if ent.IsValidationError(err) || ent.IsConstraintError(err) {
				return fmt.Errorf("%w: save system analysis relationship: %w", rez.ErrInvalidInput, err)
			}
			return fmt.Errorf("save system analysis relationship: %w", err)
		}
		savedID = saved.Unwrap().ID
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	return s.getSystemAnalysisRelationship(ctx, savedID)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisRelationship(ctx context.Context, id uuid.UUID) error {
	deleteRelationship := s.db.Client(ctx).SystemAnalysisRelationship.DeleteOneID(id)
	if deleteErr := deleteRelationship.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete system analysis relationship: %w", deleteErr)
	}
	return nil
}

func (s *SystemAnalysisService) ListSystemAnalysisEntries(ctx context.Context, analysisID uuid.UUID) (ent.SystemAnalysisEntries, error) {
	query := s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(analysisID)).
		Order(ent.Asc(sae.FieldSequence), ent.Asc(sae.FieldOccurredAt), ent.Asc(sae.FieldID)).
		WithSubjects(s.systemAnalysisEntrySubjectsQuery)
	entries, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("list system analysis entries: %w", queryErr)
	}
	return entries, nil
}

func (s *SystemAnalysisService) getSystemAnalysisEntry(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysisEntry, error) {
	query := s.db.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.ID(id)).
		WithSubjects(s.systemAnalysisEntrySubjectsQuery)
	entry, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get system analysis entry: %w", queryErr)
	}
	return entry, nil
}

func (s *SystemAnalysisService) SetSystemAnalysisEntry(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntryMutation)) (*ent.SystemAnalysisEntry, error) {
	var savedID uuid.UUID
	saveErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntry, *ent.SystemAnalysisEntryMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysisEntry.Create()
		} else {
			mutator = tx.SystemAnalysisEntry.UpdateOneID(id)
		}

		setFn(mutator.Mutation())

		saved, err := mutator.Save(ctx)
		if err != nil {
			if ent.IsValidationError(err) {
				return fmt.Errorf("%w: save system analysis entry: %w", rez.ErrInvalidInput, err)
			}
			return fmt.Errorf("save system analysis entry: %w", err)
		}
		savedID = saved.Unwrap().ID
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	return s.getSystemAnalysisEntry(ctx, savedID)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntry(ctx context.Context, id uuid.UUID) error {
	deleteErr := s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		deleteSubjects := tx.SystemAnalysisEntrySubject.Delete().
			Where(saes.EntryID(id))
		if _, subjectErr := deleteSubjects.Exec(txCtx); subjectErr != nil {
			return subjectErr
		}
		deleteEntry := tx.SystemAnalysisEntry.DeleteOneID(id)
		return deleteEntry.Exec(txCtx)
	})
	if deleteErr != nil {
		return fmt.Errorf("delete system analysis entry: %w", deleteErr)
	}
	return nil
}

func (s *SystemAnalysisService) getSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysisEntrySubject, error) {
	query := s.db.Client(ctx).SystemAnalysisEntrySubject.Query().
		Where(saes.ID(id))
	s.systemAnalysisEntrySubjectsQuery(query)
	subject, queryErr := query.Only(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get system analysis entry subject: %w", queryErr)
	}
	return subject, nil
}

func (s *SystemAnalysisService) validateEntrySubjectMutation(id uuid.UUID, m *ent.SystemAnalysisEntrySubjectMutation) error {
	if id != uuid.Nil {
		_, entitySet := m.KnowledgeEntityID()
		_, relationshipSet := m.KnowledgeRelationshipID()
		_, evidenceSet := m.KnowledgeEvidenceID()
		if entitySet ||
			relationshipSet ||
			evidenceSet ||
			m.KnowledgeEntityIDCleared() ||
			m.KnowledgeRelationshipIDCleared() ||
			m.KnowledgeEvidenceIDCleared() {
			return fmt.Errorf("%w: analysis entry subject target cannot be changed", rez.ErrInvalidInput)
		}
		return nil
	}

	targetCount := 0
	for _, getTarget := range []func() (uuid.UUID, bool){
		m.KnowledgeEntityID,
		m.KnowledgeRelationshipID,
		m.KnowledgeEvidenceID,
	} {
		id, exists := getTarget()
		if !exists {
			continue
		}
		if id == uuid.Nil {
			return fmt.Errorf("%w: analysis entry subject reference is required", rez.ErrInvalidInput)
		}
		targetCount++
	}
	if targetCount != 1 {
		return fmt.Errorf("%w: analysis entry subject must reference exactly one graph object", rez.ErrInvalidInput)
	}
	return nil
}

func (s *SystemAnalysisService) SetSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID, setFn func(*ent.SystemAnalysisEntrySubjectMutation)) (*ent.SystemAnalysisEntrySubject, error) {
	var savedID uuid.UUID
	saveErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.SystemAnalysisEntrySubject, *ent.SystemAnalysisEntrySubjectMutation]
		if id == uuid.Nil {
			mutator = tx.SystemAnalysisEntrySubject.Create()
		} else {
			mutator = tx.SystemAnalysisEntrySubject.UpdateOneID(id)
		}

		mutation := mutator.Mutation()
		setFn(mutation)
		if validationErr := s.validateEntrySubjectMutation(id, mutation); validationErr != nil {
			return validationErr
		}

		saved, err := mutator.Save(ctx)
		if err != nil {
			if ent.IsValidationError(err) {
				return fmt.Errorf("%w: save system analysis entry subject: %w", rez.ErrInvalidInput, err)
			}
			return fmt.Errorf("save system analysis entry subject: %w", err)
		}
		savedID = saved.Unwrap().ID
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	return s.getSystemAnalysisEntrySubject(ctx, savedID)
}

func (s *SystemAnalysisService) DeleteSystemAnalysisEntrySubject(ctx context.Context, id uuid.UUID) error {
	deleteSubject := s.db.Client(ctx).SystemAnalysisEntrySubject.DeleteOneID(id)
	if deleteErr := deleteSubject.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete system analysis entry subject: %w", deleteErr)
	}
	return nil
}
