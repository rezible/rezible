package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
)

type RetrospectiveService struct {
	db rez.Database
}

func NewRetrospectiveService(db rez.Database) (*RetrospectiveService, error) {
	return &RetrospectiveService{db: db}, nil
}

func (s *RetrospectiveService) Get(ctx context.Context, p predicate.Retrospective) (*ent.Retrospective, error) {
	return s.db.Client(ctx).Retrospective.Query().
		Where(p).
		WithReviews().
		Only(ctx)
}

func (s *RetrospectiveService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.RetrospectiveMutation)) (*ent.Retrospective, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.Retrospective, error) {
		if lockErr := s.db.AcquireTxLocks(ctx, "retrospective_state", id.String()); lockErr != nil {
			return nil, fmt.Errorf("lock retrospective lifecycle state: %w", lockErr)
		}
		current, queryErr := tx.Retrospective.Get(ctx, id)
		if queryErr != nil {
			return nil, queryErr
		}
		update := tx.Retrospective.UpdateOneID(id)
		setFn(update.Mutation())
		updatedValue, updateErr := update.Save(ctx)
		if updateErr != nil {
			return nil, updateErr
		}
		updated := updatedValue
		if current.State == retrospective.StateInReview && updated.State == retrospective.StateClosed {
			if compErr := s.onRetrospectiveCompleted(ctx, updated); compErr != nil {
				return nil, fmt.Errorf("on completed: %w", compErr)
			}
		}
		return updated, nil
	})
}

func (s *RetrospectiveService) onRetrospectiveCompleted(ctx context.Context, retro *ent.Retrospective) error {
	slog.InfoContext(ctx, "retrospective.completed",
		slog.String("id", retro.ID.String()),
		slog.String("incidentId", retro.IncidentID.String()),
		slog.String("fromState", retrospective.StateInReview.String()),
		slog.String("toState", retrospective.StateClosed.String()),
	)
	// TODO: invoke future retrospective completion processing here.
	return nil
}

func (s *RetrospectiveService) CreateForIncident(ctx context.Context, incidentID uuid.UUID) (*ent.Retrospective, error) {
	return ent.WithTxReturning(ctx, s.db, func(txCtx context.Context, tx *ent.Client) (*ent.Retrospective, error) {
		// Coordinate automatic and manual creation, including concurrent status changes.
		queryIncident := tx.Incident.Query().Where(incident.ID(incidentID)).
			ForUpdate()
		inc, incidentErr := queryIncident.Only(txCtx)
		if incidentErr != nil {
			return nil, fmt.Errorf("get incident: %w", incidentErr)
		}
		queryExisting := tx.Retrospective.Query().Where(retrospective.IncidentID(incidentID))
		existing, queryErr := queryExisting.Only(txCtx)
		if queryErr == nil {
			return existing, nil
		}
		if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("get incident retrospective: %w", queryErr)
		}
		if inc.ResponseState != incident.ResponseStateResolved {
			return nil, fmt.Errorf("%w: incident must be resolved before starting a retrospective", rez.ErrConflict)
		}

		createDoc := tx.Document.Create().
			SetContent([]byte("")).
			SetAccessRestricted(false)
		createdDoc, createDocErr := createDoc.Save(txCtx)
		if createDocErr != nil {
			return nil, fmt.Errorf("create doc: %w", createDocErr)
		}

		analysis, createAnalysisErr := tx.SystemAnalysis.Create().Save(txCtx)
		if createAnalysisErr != nil {
			return nil, fmt.Errorf("create system analysis: %w", createAnalysisErr)
		}

		createRetro := tx.Retrospective.Create().
			SetIncidentID(inc.ID).
			SetDocument(createdDoc).
			SetState(retrospective.StateDraft).
			SetSystemAnalysisID(analysis.ID)
		created, createRetroErr := createRetro.Save(txCtx)
		if createRetroErr != nil {
			return nil, fmt.Errorf("create retrospective: %w", createRetroErr)
		}
		return created, nil
	})
}

func (s *RetrospectiveService) GetReportComposition(ctx context.Context, id uuid.UUID) (*rez.RetrospectiveReportComposition, error) {
	return &rez.RetrospectiveReportComposition{}, nil
}

func (s *RetrospectiveService) SetReport(ctx context.Context, id uuid.UUID, params rez.SetRetrospectiveReportParams) (*rez.RetrospectiveReport, error) {
	return &rez.RetrospectiveReport{}, nil
}
