package db

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	dc "github.com/rezible/rezible/ent/discussioncomment"
	dt "github.com/rezible/rezible/ent/discussionthread"
	"github.com/rezible/rezible/ent/predicate"
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
		targetPredicate, targetErr := s.threadTargetPredicate(params.TargetKind, params.TargetID)
		if targetErr != nil {
			return nil, targetErr
		}
		query = query.Where(targetPredicate)
	} else if params.TargetID != uuid.Nil {
		return nil, fmt.Errorf("%w: a discussion target ID needs its kind", errs.ErrInvalidInput)
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

// threadTargetPredicate matches threads with a target of the kind, or, given an ID, that target.
func (s *DiscussionService) threadTargetPredicate(kind rez.DiscussionThreadTargetKind, id uuid.UUID) (predicate.DiscussionThread, error) {
	var column string
	switch kind {
	case rez.DiscussionThreadTargetFinding:
		column = dt.FieldTargetFindingID
	case rez.DiscussionThreadTargetEntry:
		column = dt.FieldTargetEntryID
	case rez.DiscussionThreadTargetKnowledgeEntity:
		column = dt.FieldTargetKnowledgeEntityID
	case rez.DiscussionThreadTargetKnowledgeRelationship:
		column = dt.FieldTargetKnowledgeRelationshipID
	case rez.DiscussionThreadTargetNormalizedEvent:
		column = dt.FieldTargetNormalizedEventID
	default:
		return nil, fmt.Errorf("%w: unknown discussion target kind %q", errs.ErrInvalidInput, kind)
	}
	if id == uuid.Nil {
		return predicate.DiscussionThread(sql.FieldNotNull(column)), nil
	}
	return predicate.DiscussionThread(sql.FieldEQ(column, id)), nil
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
