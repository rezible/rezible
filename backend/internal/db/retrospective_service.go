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
			rez.NewCommandHandler("retrospectives.create_for_incident", s.createForIncident),
		),
	)
}

func (s *RetrospectiveService) onIncidentUpdated(ctx context.Context, evt *rez.EventOnIncidentUpdated) error {
	retro, retroErr := s.Get(ctx, retrospective.IncidentID(evt.IncidentId))
	if retroErr != nil && !ent.IsNotFound(retroErr) {
		return fmt.Errorf("query retrospective by incident id %q: %w", evt.IncidentId, retroErr)
	}
	if retro != nil {
		// TODO: check if retro needs updating
		return nil
	}
	if cmdErr := s.msgs.SendCommand(ctx, cmdCreateRetrospectiveForIncident{IncidentId: evt.IncidentId}); cmdErr != nil {
		return fmt.Errorf("send cmdCreateRetrospectiveForIncident: %w", cmdErr)
	}
	return nil
}

type cmdCreateRetrospectiveForIncident struct {
	IncidentId uuid.UUID
}

func (s *RetrospectiveService) createForIncident(ctx context.Context, cmd *cmdCreateRetrospectiveForIncident) error {
	inc, incErr := s.incidents.Get(ctx, incident.ID(cmd.IncidentId))
	if incErr != nil {
		return fmt.Errorf("get incident: %w", incErr)
	}
	if _, setErr := s.Create(ctx, inc.ID, retrospective.TypeFull); setErr != nil {
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

func (s *RetrospectiveService) getIncidentRetrospectiveType(ctx context.Context, inc *ent.Incident) (retrospective.Type, error) {
	// TODO: base on severity?
	return retrospective.TypeFull, nil
}

func (s *RetrospectiveService) Create(ctx context.Context, incidentId uuid.UUID, kind retrospective.Type) (*ent.Retrospective, error) {
	var created *ent.Retrospective
	createTxFn := func(tx *ent.Tx) error {
		createdDoc, createDocErr := tx.Document.Create().
			SetContent([]byte("")).
			Save(ctx)
		if createDocErr != nil {
			return fmt.Errorf("create doc: %w", createDocErr)
		}

		create := tx.Retrospective.Create().
			SetIncidentID(incidentId).
			SetType(kind).
			SetDocumentID(createdDoc.ID).
			SetState(retrospective.StateDraft)

		var createRetroErr error
		created, createRetroErr = create.Save(ctx)
		if createRetroErr != nil {
			return fmt.Errorf("create retrospective: %w", createRetroErr)
		}
		created.Edges.Document = createdDoc

		if kind == retrospective.TypeFull {
			createdAnalysis, createAnalysisErr := tx.SystemAnalysis.Create().
				SetRetrospectiveID(created.ID).
				Save(ctx)
			if createAnalysisErr != nil {
				return fmt.Errorf("create analysis: %w", createAnalysisErr)
			}
			created.SystemAnalysisID = createdAnalysis.ID
			created.Edges.SystemAnalysis = createdAnalysis
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
