package postgres

import (
	"context"
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

func (s *RetrospectiveService) GetByIncident(ctx context.Context, inc *ent.Incident, createMissing bool) (*ent.Retrospective, error) {
	retro, retroErr := inc.QueryRetrospective().Only(ctx)
	if retroErr == nil && retro != nil {
		return retro, nil
	}
	if ent.IsNotFound(retroErr) && createMissing {
		return s.db.Retrospective.Create().
			SetIncidentID(inc.ID).
			SetDocumentName(inc.Slug + "-retrospective").
			SetState(retrospective.StateDraft).
			Save(ctx)
	}
	return nil, retroErr
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
