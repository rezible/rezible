package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivecomment"
)

type RetrospectiveService struct {
	db        *ent.Client
	msgs      rez.MessageService
	incidents rez.IncidentService
}

func NewRetrospectiveService(db *ent.Client, msgs rez.MessageService, incidents rez.IncidentService) (*RetrospectiveService, error) {
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
	return errors.Join(
		s.msgs.AddEventHandlers(
			rez.NewEventHandler("retrospectives.on_incident_updated", s.onIncidentUpdated),
		),
		s.msgs.AddCommandHandlers(
			rez.NewCommandHandler("retrospectives.update_for_incident", s.handleUpdateForIncident),
		),
	)
}

func (s *RetrospectiveService) onIncidentUpdated(ctx context.Context, evt *rez.EventOnIncidentUpdated) error {
	if cmdErr := s.msgs.SendCommand(ctx, cmdUpdateIncidentRetrospective{IncidentId: evt.IncidentId}); cmdErr != nil {
		return fmt.Errorf("send cmdCreateRetrospectiveForIncident: %w", cmdErr)
	}
	return nil
}

type cmdUpdateIncidentRetrospective struct {
	IncidentId uuid.UUID
}

func (s *RetrospectiveService) handleUpdateForIncident(ctx context.Context, cmd *cmdUpdateIncidentRetrospective) error {
	_, retroErr := s.Get(ctx, retrospective.IncidentID(cmd.IncidentId))
	if retroErr != nil && !ent.IsNotFound(retroErr) {
		return fmt.Errorf("query retrospective by incident id %q: %w", cmd.IncidentId, retroErr)
	}
	inc, incErr := s.incidents.Get(ctx, incident.ID(cmd.IncidentId))
	if incErr != nil {
		return fmt.Errorf("get incident: %w", incErr)
	}
	if _, setErr := s.createForIncident(ctx, inc); setErr != nil {
		return fmt.Errorf("create retrospective: %w", setErr)
	}
	return nil
}

func (s *RetrospectiveService) Get(ctx context.Context, p predicate.Retrospective) (*ent.Retrospective, error) {
	return s.db.Retrospective.Query().Where(p).Only(ctx)
}

func (s *RetrospectiveService) GetById(ctx context.Context, id uuid.UUID) (*ent.Retrospective, error) {
	return s.db.Retrospective.Get(ctx, id)
}

func (s *RetrospectiveService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.RetrospectiveMutation)) (*ent.Retrospective, error) {
	update := s.db.Retrospective.UpdateOneID(id)

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
	exists, queryErr := s.db.Retrospective.Query().Where(retrospective.IncidentID(inc.ID)).Exist(ctx)
	if exists || queryErr != nil {
		return nil, queryErr
	}
	kind, kindErr := s.getRetrospectiveKind(ctx, inc)
	if kindErr != nil {
		return nil, fmt.Errorf("get retrospective kind: %w", kindErr)
	}

	var created *ent.Retrospective
	createTxFn := func(tx *ent.Tx) error {
		createdDoc, createDocErr := tx.Document.Create().
			SetContent([]byte("")).
			SetAccessRestricted(false).
			Save(ctx)
		if createDocErr != nil {
			return fmt.Errorf("create doc: %w", createDocErr)
		}

		create := tx.Retrospective.Create().
			SetIncident(inc).
			SetDocument(createdDoc).
			SetKind(kind).
			SetState(retrospective.StateDraft)

		var createRetroErr error
		created, createRetroErr = create.Save(ctx)
		if createRetroErr != nil {
			return fmt.Errorf("create retrospective: %w", createRetroErr)
		}

		if kind == retrospective.KindFull {
			createdAnalysis, createAnalysisErr := tx.SystemAnalysis.Create().
				SetRetrospective(created).
				Save(ctx)
			if createAnalysisErr != nil {
				return fmt.Errorf("create analysis: %w", createAnalysisErr)
			}
			created.SystemAnalysisID = createdAnalysis.ID
		}
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, createTxFn); txErr != nil {
		return nil, fmt.Errorf("create tx failed: %w", txErr)
	}
	return created, nil
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
