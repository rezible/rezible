package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	sa "github.com/rezible/rezible/ent/systemanalysis"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
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

func (s *SystemAnalysisService) GetSystemAnalysis(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysis, error) {
	query := s.db.Client(ctx).SystemAnalysis.Query().
		Where(sa.ID(id)).
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
