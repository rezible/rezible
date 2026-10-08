package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	dc "github.com/rezible/rezible/ent/discussioncomment"
	dt "github.com/rezible/rezible/ent/discussionthread"
	"github.com/rezible/rezible/ent/review"
	"github.com/rezible/rezible/pkg/errs"
)

type DiscussionService struct {
	db rez.Database
}

func NewDiscussionService(db rez.Database) *DiscussionService {
	return &DiscussionService{db: db}
}

func (s *DiscussionService) ListThreads(ctx context.Context, params rez.ListDiscussionThreadsParams) (*ent.ListResult[ent.DiscussionThread], error) {
	if params.AnalysisID != uuid.Nil {
		if _, queryErr := s.db.Client(ctx).SystemAnalysis.Get(ctx, params.AnalysisID); queryErr != nil {
			return nil, queryErr
		}
	}
	if params.RetrospectiveID != uuid.Nil {
		if _, queryErr := s.db.Client(ctx).Retrospective.Get(ctx, params.RetrospectiveID); queryErr != nil {
			return nil, queryErr
		}
	}
	query := s.db.Client(ctx).DiscussionThread.Query()
	if params.AnalysisID != uuid.Nil {
		query = query.Where(dt.AnalysisID(params.AnalysisID))
	}
	if params.RetrospectiveID != uuid.Nil {
		query = query.Where(dt.RetrospectiveID(params.RetrospectiveID))
	}
	if params.TargetKind != "" {
		query = query.Where(dt.TargetKindEQ(params.TargetKind))
	}
	if params.TargetID != uuid.Nil {
		query = query.Where(dt.TargetID(params.TargetID))
	}
	if params.Resolved != nil {
		pred := dt.ResolvedAtIsNil()
		if *params.Resolved {
			pred = dt.ResolvedAtNotNil()
		}
		query = query.Where(pred)
	}
	order := params.ListParams.GetOrder()
	query = query.Order(dt.ByCreatedAt(order), dt.ByID(order))
	return ent.DoListQuery[ent.DiscussionThread, *ent.DiscussionThreadQuery](ctx, query, params.ListParams)
}

func (s *DiscussionService) CreateThread(ctx context.Context, params rez.CreateDiscussionThreadParams) (*ent.DiscussionThread, error) {
	//TODO implement me
	panic("implement me")
}

func (s *DiscussionService) GetThread(ctx context.Context, id uuid.UUID) (*ent.DiscussionThread, error) {
	thread, queryErr := s.db.Client(ctx).DiscussionThread.Get(ctx, id)
	if queryErr != nil {
		return nil, queryErr
	}
	return thread, nil
}

func (s *DiscussionService) ListComments(ctx context.Context, params rez.ListDiscussionCommentsParams) (*ent.ListResult[ent.DiscussionComment], error) {
	if params.ParentID != uuid.Nil {
		return nil, fmt.Errorf("%w: discussion replies use a flat stream", errs.ErrInvalidInput)
	}
	if params.PageSize < 1 || params.PageSize > 50 {
		params.PageSize = 50
	}
	query := s.db.Client(ctx).DiscussionComment.Query().
		Where(dc.ThreadID(params.ThreadID)).
		Where(dc.ParentIDIsNil()).
		Order(dc.ByCreatedAt(), dc.ByID())
	return ent.DoListQuery[ent.DiscussionComment, *ent.DiscussionCommentQuery](ctx, query, params.ListParams)
}

func (s *DiscussionService) GetComment(ctx context.Context, id uuid.UUID) (*ent.DiscussionComment, error) {
	return s.db.Client(ctx).DiscussionComment.Get(ctx, id)
}

func (s *DiscussionService) CreateComment(ctx context.Context, params rez.CreateDiscussionCommentParams) (*ent.DiscussionComment, error) {
	//TODO implement me
	panic("implement me")
}

func (s *DiscussionService) UpdateComment(ctx context.Context, u uuid.UUID, params rez.UpdateDiscussionCommentParams) (*ent.DiscussionComment, error) {
	//TODO implement me
	panic("implement me")
}

func (s *DiscussionService) ListReviews(ctx context.Context, params rez.ListReviewsParams) (*ent.ListResult[ent.Review], error) {
	query := s.db.Client(ctx).Review.Query()
	if params.RetrospectiveID != uuid.Nil {
		query = query.Where(review.RetrospectiveID(params.RetrospectiveID))
	}
	if params.AnalysisEntryID != uuid.Nil {
		query = query.Where(review.AnalysisEntryID(params.AnalysisEntryID))
	}
	query = query.Order(review.ByCreatedAt(), review.ByID())
	return ent.DoListQuery[ent.Review, *ent.ReviewQuery](ctx, query, params.ListParams)
}

func (s *DiscussionService) GetReview(ctx context.Context, id uuid.UUID) (*ent.Review, error) {
	return s.db.Client(ctx).Review.Get(ctx, id)
}

func (s *DiscussionService) CreateReviewRequest(ctx context.Context, params rez.CreateReviewRequestParams) (*ent.Review, error) {
	//TODO implement me
	panic("implement me")
}
