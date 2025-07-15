package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivediscussion"
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
	inc, incErr := params.Edges.IncidentOrErr()
	if incErr != nil {
		inc, incErr = s.db.Incident.Get(ctx, params.IncidentID)
		if incErr != nil {
			return nil, fmt.Errorf("get incident: %w", incErr)
		}
	}

	var createdRetro *ent.Retrospective
	var createdAnalysis *ent.SystemAnalysis

	createTxFn := func(tx *ent.Tx) error {
		var createErr error
		createdRetro, createErr = tx.Retrospective.Create().
			SetIncidentID(inc.ID).
			SetDocumentName(inc.Slug + "-retrospective").
			SetType(params.Type).
			SetState(retrospective.StateDraft).
			Save(ctx)
		if createErr != nil {
			return createErr
		}
		if params.Type == retrospective.TypeSimple {
			return nil
		}
		createdAnalysis, createErr = tx.SystemAnalysis.Create().
			SetRetrospectiveID(createdRetro.ID).
			Save(ctx)
		if createErr != nil {
			return createErr
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
	retro, retroErr := s.db.Retrospective.Query().Where(retrospective.IncidentID(inc.ID)).Only(ctx)
	if retroErr == nil {
		return retro, nil
	} else if !ent.IsNotFound(retroErr) {
		return nil, fmt.Errorf("querying retrospective: %w", retroErr)
	}
	// TODO: query based on incident severity?
	retroType := retrospective.TypeFull
	retro, retroErr = s.Create(ctx, ent.Retrospective{
		Type:  retroType,
		Edges: ent.RetrospectiveEdges{Incident: inc},
	})
	if retroErr != nil {
		return nil, fmt.Errorf("creating retro: %w", retroErr)
	}
	return retro, nil
}

func (s *RetrospectiveService) CreateDiscussion(ctx context.Context, params rez.CreateRetrospectiveDiscussionParams) (*ent.RetrospectiveDiscussion, error) {
	return s.db.RetrospectiveDiscussion.Create().
		SetRetrospectiveID(params.RetrospectiveID).
		SetContent(params.Content).
		//SetUserID(params.UserID).
		Save(ctx)
}

func (s *RetrospectiveService) GetDiscussionByID(ctx context.Context, id uuid.UUID) (*ent.RetrospectiveDiscussion, error) {
	return s.db.RetrospectiveDiscussion.Get(ctx, id)
}

func (s *RetrospectiveService) AddDiscussionReply(ctx context.Context, params rez.AddRetrospectiveDiscussionReplyParams) (*ent.RetrospectiveDiscussionReply, error) {
	return s.db.RetrospectiveDiscussionReply.Create().
		SetDiscussionID(params.DiscussionId).
		SetContent(params.Content).
		SetNillableParentReplyID(params.ParentID).
		//SetUserID(params.UserID).
		Save(ctx)
}

func (s *RetrospectiveService) ListDiscussions(ctx context.Context, params rez.ListRetrospectiveDiscussionsParams) ([]*ent.RetrospectiveDiscussion, error) {
	query := s.db.RetrospectiveDiscussion.Query().
		Where(retrospectivediscussion.RetrospectiveID(params.RetrospectiveID))

	if params.WithReplies {
		query = query.WithReplies()
	}

	return query.All(ctx)
}
