package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/pkg/messages"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivecomment"
)

type RetrospectiveService struct {
	db        rez.Database
	msgs      rez.MessageService
	incidents rez.IncidentService
}

func NewRetrospectiveService(
	db rez.Database,
	msgs rez.MessageService,
	incidents rez.IncidentService,
) (*RetrospectiveService, error) {
	svc := &RetrospectiveService{
		db:        db,
		msgs:      msgs,
		incidents: incidents,
	}

	if msgsErr := svc.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}

	return svc, nil
}

func (s *RetrospectiveService) registerMessageHandlers() error {
	return s.msgs.AddHandlers(
		messages.NewEventHandler("retrospectives.on_incident_updated", s.onIncidentUpdated),
	)
}

func (s *RetrospectiveService) onIncidentUpdated(ctx context.Context, evt *rez.EventOnIncidentUpdated) error {
	// TODO: update retrospective
	return nil
}

func (s *RetrospectiveService) Get(ctx context.Context, p predicate.Retrospective) (*ent.Retrospective, error) {
	return s.db.Client(ctx).Retrospective.Query().Where(p).Only(ctx)
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

func (s *RetrospectiveService) getRetrospectiveKind(ctx context.Context, inc *ent.Incident) (retrospective.Kind, error) {
	// TODO: base on severity?
	return retrospective.KindFull, nil
}

func (s *RetrospectiveService) createForIncident(ctx context.Context, inc *ent.Incident) (*ent.Retrospective, error) {
	exists, queryErr := s.db.Client(ctx).Retrospective.Query().Where(retrospective.IncidentID(inc.ID)).Exist(ctx)
	if exists || queryErr != nil {
		return nil, queryErr
	}
	kind, kindErr := s.getRetrospectiveKind(ctx, inc)
	if kindErr != nil {
		return nil, fmt.Errorf("get retrospective kind: %w", kindErr)
	}

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
	return s.db.Client(ctx).Retrospective.Query().Where(retrospective.IncidentID(inc.ID)).Only(ctx)
}

func (s *RetrospectiveService) GetComment(ctx context.Context, id uuid.UUID) (*ent.RetrospectiveComment, error) {
	return s.db.Client(ctx).RetrospectiveComment.Get(ctx, id)
}

func (s *RetrospectiveService) SetComment(ctx context.Context, cmt *ent.RetrospectiveComment) (*ent.RetrospectiveComment, error) {
	var m *ent.RetrospectiveCommentMutation
	if cmt.ID != uuid.Nil {
		m = s.db.Client(ctx).RetrospectiveComment.UpdateOneID(cmt.ID).Mutation()
	} else {
		m = s.db.Client(ctx).RetrospectiveComment.Create().Mutation()
	}
	v, setErr := s.db.Client(ctx).Mutate(ctx, m)
	if setErr != nil {
		return nil, fmt.Errorf("failed to %s comment: %w", m.Op(), setErr)
	}
	updated, ok := v.(*ent.RetrospectiveComment)
	if !ok {
		return nil, fmt.Errorf("invalid ")
	}
	return updated, nil
}

func (s *RetrospectiveService) ListComments(ctx context.Context, params rez.ListRetrospectiveCommentsParams) (*ent.ListResult[ent.RetrospectiveComment], error) {
	query := s.db.Client(ctx).RetrospectiveComment.Query().
		Where(retrospectivecomment.RetrospectiveID(params.RetrospectiveID)).
		Order(retrospectivecomment.ByID(params.GetOrder()))

	if params.WithReplies {
		query = query.WithReplies()
	}

	return ent.DoListQuery[ent.RetrospectiveComment, *ent.RetrospectiveCommentQuery](ctx, query, params.ListParams)
}
