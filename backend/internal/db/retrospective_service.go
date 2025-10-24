package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivecomment"
)

type RetrospectiveService struct {
	db *ent.Client
}

func NewRetrospectiveService(db *ent.Client) (*RetrospectiveService, error) {
	svc := &RetrospectiveService{
		db: db,
	}

	return svc, nil
}

func (s *RetrospectiveService) GetById(ctx context.Context, id uuid.UUID) (*ent.Retrospective, error) {
	return s.db.Retrospective.Get(ctx, id)
}

func (s *RetrospectiveService) getIncidentRetrospectiveType(ctx context.Context, inc *ent.Incident) (retrospective.Type, error) {
	// TODO: base on severity?
	return retrospective.TypeFull, nil
}

func (s *RetrospectiveService) Create(ctx context.Context, params ent.Retrospective) (*ent.Retrospective, error) {
	var createdRetro *ent.Retrospective
	var createdAnalysis *ent.SystemAnalysis

	createTxFn := func(tx *ent.Tx) error {
		var createErr error
		createdDoc, createErr := s.db.Document.Create().SetContent([]byte("")).Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create doc: %w", createErr)
		}

		createdRetro, createErr = tx.Retrospective.Create().
			SetIncidentID(params.IncidentID).
			SetDocumentID(createdDoc.ID).
			SetType(params.Type).
			SetState(retrospective.StateDraft).
			Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create retrospective: %w", createErr)
		}
		if params.Type == retrospective.TypeFull {
			createdAnalysis, createErr = tx.SystemAnalysis.Create().
				SetRetrospectiveID(createdRetro.ID).
				Save(ctx)
			if createErr != nil {
				return fmt.Errorf("create analysis: %w", createErr)
			}
		}
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, createTxFn); txErr != nil {
		return nil, fmt.Errorf("create tx failed: %w", txErr)
	}

	if createdAnalysis != nil {
		createdRetro.SystemAnalysisID = createdAnalysis.ID
		createdRetro.Edges.SystemAnalysis = createdAnalysis
	}
	return createdRetro, nil
}

func (s *RetrospectiveService) GetForIncident(ctx context.Context, inc *ent.Incident) (*ent.Retrospective, error) {
	return s.db.Retrospective.Query().Where(retrospective.IncidentID(inc.ID)).Only(ctx)
}

func (s *RetrospectiveService) GetComment(ctx context.Context, id uuid.UUID) (*ent.RetrospectiveComment, error) {
	return s.db.RetrospectiveComment.Get(ctx, id)
}

func (s *RetrospectiveService) SetComment(ctx context.Context, cmt *ent.RetrospectiveComment) (*ent.RetrospectiveComment, error) {
	var m *ent.RetrospectiveCommentMutation
	if cmt.ID != uuid.Nil {
		m = s.db.RetrospectiveComment.UpdateOneID(cmt.ID).Mutation()
	} else {
		m = s.db.RetrospectiveComment.Create().Mutation()
	}
	v, setErr := s.db.Mutate(ctx, m)
	if setErr != nil {
		return nil, fmt.Errorf("failed to %s comment: %w", m.Op(), setErr)
	}
	updated, ok := v.(*ent.RetrospectiveComment)
	if !ok {
		return nil, fmt.Errorf("invalid ")
	}
	return updated, nil
}

func (s *RetrospectiveService) ListComments(ctx context.Context, params rez.ListRetrospectiveCommentsParams) ([]*ent.RetrospectiveComment, error) {
	query := s.db.RetrospectiveComment.Query().
		Where(retrospectivecomment.RetrospectiveID(params.RetrospectiveID))

	if params.WithReplies {
		query = query.WithReplies()
	}

	return query.All(ctx)
}
