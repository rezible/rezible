package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/errs"
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
		return nil, oapi.Error(ctx, "exactly one discussion owner is required", errs.ErrInvalidInput)
	}
	params := rez.ListDiscussionThreadsParams{
		ListParams:      request.ListParams(),
		AnalysisID:      request.AnalysisId,
		RetrospectiveID: request.RetrospectiveId,
		TargetKind:      request.TargetKind,
		TargetID:        request.TargetId,
	}
	if request.State.Sent {
		params.Resolved = new(request.State.Value == "resolved")
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

func (h *discussionHandler) CreateDiscussionThread(ctx context.Context, request *oapi.CreateDiscussionThreadRequest) (*oapi.CreateDiscussionThreadResponse, error) {
	attrs := request.Body.Attributes
	params := rez.CreateDiscussionThreadParams{
		AnalysisID:      attrs.AnalysisId,
		RetrospectiveID: attrs.RetrospectiveId,
		TargetKind:      attrs.TargetKind,
		TargetID:        attrs.TargetId,
		InitialMessage:  attrs.InitialMessage,
	}
	thread, createErr := h.discussions.CreateThread(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create discussion thread", createErr)
	}
	var resp oapi.CreateDiscussionThreadResponse
	resp.Body.Data = oapi.DiscussionThreadFromEnt(thread)
	return &resp, nil
}

func (*discussionHandler) UpdateDiscussionThread(ctx context.Context, _ *oapi.UpdateDiscussionThreadRequest) (*oapi.UpdateDiscussionThreadResponse, error) {
	return nil, oapi.Error(ctx, "discussion thread updates are not implemented", errs.ErrNotImplemented)
}

func (h *discussionHandler) CreateDiscussionComment(ctx context.Context, request *oapi.CreateDiscussionCommentRequest) (*oapi.CreateDiscussionCommentResponse, error) {
	attrs := request.Body.Attributes
	if attrs.ParentId != nil {
		return nil, oapi.Error(ctx, "discussion replies use a flat stream", errs.ErrInvalidInput)
	}
	params := rez.CreateDiscussionCommentParams{
		ThreadID: request.Id,
		Content:  attrs.Content,
	}
	comment, createErr := h.discussions.CreateComment(ctx, params)
	if createErr != nil {
		return nil, oapi.Error(ctx, "create discussion comment", createErr)
	}

	var resp oapi.CreateDiscussionCommentResponse
	resp.Body.Data = oapi.DiscussionCommentFromEnt(comment)
	return &resp, nil
}

func (h *discussionHandler) UpdateDiscussionComment(ctx context.Context, request *oapi.UpdateDiscussionCommentRequest) (*oapi.UpdateDiscussionCommentResponse, error) {
	attrs := request.Body.Attributes
	params := rez.UpdateDiscussionCommentParams{
		Content: attrs.Content,
	}
	comment, updateErr := h.discussions.UpdateComment(ctx, request.Id, params)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "update discussion comment", updateErr)
	}
	var resp oapi.UpdateDiscussionCommentResponse
	resp.Body.Data = oapi.DiscussionCommentFromEnt(comment)
	return &resp, nil
}
