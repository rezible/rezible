package postgres

import (
	"context"
	"github.com/google/uuid"
	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/retrospective"
	"github.com/twohundreds/rezible/ent/retrospectivediscussion"
)

type RetrospectiveService struct {
	db *ent.Client
}

func NewRetrospectiveService(db *ent.Client) (*RetrospectiveService, error) {
	return &RetrospectiveService{db: db}, nil
}

func (s *RetrospectiveService) GetByIncidentID(ctx context.Context, incidentId uuid.UUID) (*ent.Retrospective, error) {
	return s.db.Retrospective.Query().
		Where(retrospective.HasIncidentWith(incident.IDEQ(incidentId))).
		Only(ctx)
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
