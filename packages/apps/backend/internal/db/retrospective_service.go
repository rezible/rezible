package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
)

type RetrospectiveService struct {
	db        rez.Database
	incidents rez.IncidentService
}

func NewRetrospectiveService(
	db rez.Database,
	incidents rez.IncidentService,
) (*RetrospectiveService, error) {
	svc := &RetrospectiveService{
		db:        db,
		incidents: incidents,
	}
	return svc, nil
}

func (s *RetrospectiveService) Get(ctx context.Context, p predicate.Retrospective) (*ent.Retrospective, error) {
	return s.db.Client(ctx).Retrospective.Query().
		Where(p).
		WithReviews().
		Only(ctx)
}

func (s *RetrospectiveService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.RetrospectiveMutation)) (*ent.Retrospective, error) {
	update := s.db.Client(ctx).Retrospective.UpdateOneID(id)

	setFn(update.Mutation())

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, updateErr
	}

	return updated, nil
}

func (s *RetrospectiveService) createForIncident(ctx context.Context, inc *ent.Incident) (*ent.Retrospective, error) {
	exists, queryErr := s.db.Client(ctx).Retrospective.Query().Where(retrospective.IncidentID(inc.ID)).Exist(ctx)
	if exists || queryErr != nil {
		return nil, queryErr
	}

	// TODO: base on severity?
	kind := retrospective.KindFull

	var retro *ent.Retrospective
	createTxFn := func(txCtx context.Context, tx *ent.Client) error {
		createdDoc, createDocErr := tx.Document.Create().
			SetContent([]byte("")).
			SetAccessRestricted(false).
			Save(txCtx)
		if createDocErr != nil {
			return fmt.Errorf("create doc: %w", createDocErr)
		}

		createRetro := tx.Retrospective.Create().
			SetIncident(inc).
			SetDocument(createdDoc).
			SetKind(kind).
			SetState(retrospective.StateDraft)
		analysis, createAnalysisErr := tx.SystemAnalysis.Create().Save(txCtx)
		if createAnalysisErr != nil {
			return fmt.Errorf("create system analysis: %w", createAnalysisErr)
		}
		createRetro.SetSystemAnalysis(analysis)

		created, createRetroErr := createRetro.Save(txCtx)
		if createRetroErr != nil {
			return fmt.Errorf("create retrospective: %w", createRetroErr)
		}
		retro = created.Unwrap()
		return nil
	}
	return retro, s.db.WithTx(ctx, createTxFn)
}

func (s *RetrospectiveService) GetForIncident(ctx context.Context, inc *ent.Incident) (*ent.Retrospective, error) {
	return s.db.Client(ctx).Retrospective.Query().
		Where(retrospective.IncidentID(inc.ID)).
		WithReviews().
		Only(ctx)
}
