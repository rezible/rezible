package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type discussionHandler struct {
	discussions rez.DiscussionService
}

func newDiscussionHandler(discussions rez.DiscussionService) *discussionHandler {
	return &discussionHandler{discussions: discussions}
}

func (h *discussionHandler) ListDiscussionThreads(ctx context.Context, request *oapi.ListDiscussionThreadsRequest) (*oapi.ListDiscussionThreadsResponse, error) {
	var resp oapi.ListDiscussionThreadsResponse

	owner := request.DiscussionOwnerQuery
	if (owner.AnalysisId == uuid.Nil) == (owner.RetrospectiveId == uuid.Nil) {
		return nil, oapi.Error(ctx, "exactly one discussion owner is required", rez.ErrInvalidInput)
	}
	params := rez.ListDiscussionThreadsParams{
		ListParams:      request.ListParams(),
		AnalysisID:      request.AnalysisId,
		RetrospectiveID: request.RetrospectiveId,
		Kind:            request.Kind,
		TargetKind:      request.TargetKind,
		TargetID:        request.TargetId,
		ResolutionState: request.ResolutionState,
	}
	result, listErr := h.discussions.ListThreads(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list discussion threads", listErr)
	}
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.DiscussionThreadFromEnt)
	return &resp, nil
}

func (h *discussionHandler) GetDiscussionThread(ctx context.Context, request *oapi.GetDiscussionThreadRequest) (*oapi.GetDiscussionThreadResponse, error) {
	var resp oapi.GetDiscussionThreadResponse

	thread, getErr := h.discussions.GetThread(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get discussion thread", getErr)
	}
	resp.Body.Data = oapi.DiscussionThreadFromEnt(thread)
	return &resp, nil
}

func (h *discussionHandler) ListDiscussionComments(ctx context.Context, request *oapi.ListDiscussionCommentsRequest) (*oapi.ListDiscussionCommentsResponse, error) {
	var resp oapi.ListDiscussionCommentsResponse

	if _, queryErr := h.discussions.GetThread(ctx, request.Id); queryErr != nil {
		return nil, oapi.Error(ctx, "get discussion comment thread", queryErr)
	}
	params := rez.ListDiscussionCommentsParams{
		ListParams: request.ListParams(),
		ThreadID:   request.Id,
		ParentID:   request.ParentId,
	}
	result, listErr := h.discussions.ListComments(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "list discussion comments", listErr)
	}
	resp.Body = oapi.ConvertPaginatedResultBody(result, oapi.DiscussionCommentFromEnt)
	return &resp, nil
}

func (h *discussionHandler) GetDiscussionComment(ctx context.Context, request *oapi.GetDiscussionCommentRequest) (*oapi.GetDiscussionCommentResponse, error) {
	var resp oapi.GetDiscussionCommentResponse

	comment, getErr := h.discussions.GetComment(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get discussion comment", getErr)
	}
	resp.Body.Data = oapi.DiscussionCommentFromEnt(comment)
	return &resp, nil
}

func (*discussionHandler) CreateDiscussionThread(ctx context.Context, _ *oapi.CreateDiscussionThreadRequest) (*oapi.CreateDiscussionThreadResponse, error) {
	return nil, oapi.Error(ctx, "discussion thread creation is not implemented", rez.ErrNotImplemented)
}

func (*discussionHandler) UpdateDiscussionThread(ctx context.Context, _ *oapi.UpdateDiscussionThreadRequest) (*oapi.UpdateDiscussionThreadResponse, error) {
	return nil, oapi.Error(ctx, "discussion thread updates are not implemented", rez.ErrNotImplemented)
}

func (*discussionHandler) CreateDiscussionComment(ctx context.Context, _ *oapi.CreateDiscussionCommentRequest) (*oapi.CreateDiscussionCommentResponse, error) {
	return nil, oapi.Error(ctx, "discussion comment creation is not implemented", rez.ErrNotImplemented)
}

func (*discussionHandler) UpdateDiscussionComment(ctx context.Context, _ *oapi.UpdateDiscussionCommentRequest) (*oapi.UpdateDiscussionCommentResponse, error) {
	return nil, oapi.Error(ctx, "discussion comment updates are not implemented", rez.ErrNotImplemented)
}

func (h *discussionHandler) ListReviews(ctx context.Context, request *oapi.ListReviewsRequest) (*oapi.ListReviewsResponse, error) {
	var resp oapi.ListReviewsResponse
	if request.RetrospectiveId != uuid.Nil && request.AnalysisEntryId != uuid.Nil {
		return nil, oapi.Error(ctx, "specify either retrospectiveId or analysisEntryId", rez.ErrInvalidInput)
	}
	params := rez.ListReviewsParams{
		ListParams:      request.ListParams(),
		RetrospectiveID: request.RetrospectiveId,
		AnalysisEntryID: request.AnalysisEntryId,
	}
	results, queryErr := h.discussions.ListReviews(ctx, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to list reviews", queryErr)
	}
	resp.Body = oapi.ConvertPaginatedResultBody(results, oapi.ReviewFromEnt)
	return &resp, nil
}

func (h *discussionHandler) GetReview(ctx context.Context, request *oapi.GetReviewRequest) (*oapi.GetReviewResponse, error) {
	var resp oapi.GetReviewResponse
	value, queryErr := h.discussions.GetReview(ctx, request.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get review", queryErr)
	}
	resp.Body.Data = oapi.ReviewFromEnt(value)
	return &resp, nil
}

func (*discussionHandler) CreateReview(ctx context.Context, _ *oapi.CreateReviewRequest) (*oapi.CreateReviewResponse, error) {
	return nil, oapi.Error(ctx, "create review is not implemented", rez.ErrNotImplemented)
}

func (*discussionHandler) UpdateReview(ctx context.Context, _ *oapi.UpdateReviewRequest) (*oapi.UpdateReviewResponse, error) {
	return nil, oapi.Error(ctx, "update review is not implemented", rez.ErrNotImplemented)
}

func (*discussionHandler) ArchiveReview(ctx context.Context, _ *oapi.ArchiveReviewRequest) (*oapi.ArchiveReviewResponse, error) {
	return nil, oapi.Error(ctx, "archive review is not implemented", rez.ErrNotImplemented)
}
