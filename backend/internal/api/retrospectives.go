package api

import (
	"context"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type retrospectivesHandler struct {
	auth      rez.AuthService
	users     rez.UserService
	retros    rez.RetrospectiveService
	documents rez.DocumentsService
}

func newRetrospectivesHandler(auth rez.AuthService, users rez.UserService, retros rez.RetrospectiveService, documents rez.DocumentsService) *retrospectivesHandler {
	return &retrospectivesHandler{auth, users, retros, documents}
}

func (h *retrospectivesHandler) ListRetrospectives(ctx context.Context, input *oapi.ListRetrospectivesRequest) (*oapi.ListRetrospectivesResponse, error) {
	var resp oapi.ListRetrospectivesResponse

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	var resp oapi.GetRetrospectiveResponse

	resp.Body.Data = oapi.Retrospective{
		Id: uuid.New(),
		Attributes: oapi.RetrospectiveAttributes{
			Title:        "Foolish Florbing",
			DocumentName: "document-name",
			Status:       oapi.RetrospectiveStatusOpen,
			Summary:      "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		},
	}

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveForIncident(ctx context.Context, input *oapi.GetRetrospectiveForIncidentRequest) (*oapi.GetRetrospectiveForIncidentResponse, error) {
	var resp oapi.GetRetrospectiveForIncidentResponse

	retro, retroErr := h.retros.GetByIncidentID(ctx, input.Id)
	if retroErr != nil {
		return nil, detailError("failed to get retrospective", retroErr)
	}

	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)

	return &resp, nil
}

func (h *retrospectivesHandler) ListRetrospectiveReviews(ctx context.Context, request *oapi.ListRetrospectiveReviewsRequest) (*oapi.ListRetrospectiveReviewsResponse, error) {
	var resp oapi.ListRetrospectiveReviewsResponse

	return &resp, nil
}

func (h *retrospectivesHandler) CreateRetrospectiveReview(ctx context.Context, request *oapi.CreateRetrospectiveReviewRequest) (*oapi.CreateRetrospectiveReviewResponse, error) {
	var resp oapi.CreateRetrospectiveReviewResponse

	return &resp, nil
}

func (h *retrospectivesHandler) UpdateRetrospectiveReview(ctx context.Context, request *oapi.UpdateRetrospectiveReviewRequest) (*oapi.UpdateRetrospectiveReviewResponse, error) {
	var resp oapi.UpdateRetrospectiveReviewResponse

	return &resp, nil
}

func (h *retrospectivesHandler) ArchiveRetrospectiveReview(ctx context.Context, request *oapi.ArchiveRetrospectiveReviewRequest) (*oapi.ArchiveRetrospectiveReviewResponse, error) {
	var resp oapi.ArchiveRetrospectiveReviewResponse

	return &resp, nil
}

func (h *retrospectivesHandler) ListRetrospectiveTemplates(ctx context.Context, request *oapi.ListRetrospectiveTemplatesRequest) (*oapi.ListRetrospectiveTemplatesResponse, error) {
	var resp oapi.ListRetrospectiveTemplatesResponse
	// TODO
	return &resp, nil
}
func (h *retrospectivesHandler) CreateRetrospectiveTemplate(ctx context.Context, request *oapi.CreateRetrospectiveTemplateRequest) (*oapi.CreateRetrospectiveTemplateResponse, error) {
	var resp oapi.CreateRetrospectiveTemplateResponse
	// TODO
	return &resp, nil
}
func (h *retrospectivesHandler) UpdateRetrospectiveTemplate(ctx context.Context, request *oapi.UpdateRetrospectiveTemplateRequest) (*oapi.UpdateRetrospectiveTemplateResponse, error) {
	var resp oapi.UpdateRetrospectiveTemplateResponse
	// TODO
	return &resp, nil
}
func (h *retrospectivesHandler) ArchiveRetrospectiveTemplate(ctx context.Context, request *oapi.ArchiveRetrospectiveTemplateRequest) (*oapi.ArchiveRetrospectiveTemplateResponse, error) {
	var resp oapi.ArchiveRetrospectiveTemplateResponse
	// TODO
	return &resp, nil
}

func (h *retrospectivesHandler) ListRetrospectiveDiscussions(ctx context.Context, request *oapi.ListRetrospectiveDiscussionsRequest) (*oapi.ListRetrospectiveDiscussionsResponse, error) {
	var resp oapi.ListRetrospectiveDiscussionsResponse

	discussions, discErr := h.retros.ListDiscussions(ctx, rez.ListRetrospectiveDiscussionsParams{
		RetrospectiveID: request.Id,
		WithReplies:     true,
		ListParams:      request.ListParams(),
	})
	if discErr != nil {
		return nil, detailError("failed to list discussions", discErr)
	}

	resp.Body.Data = make([]oapi.RetrospectiveDiscussion, len(discussions))
	for i, disc := range discussions {
		resp.Body.Data[i] = oapi.RetrospectiveDiscussionFromEnt(disc)
	}

	return &resp, nil
}

func (h *retrospectivesHandler) CreateRetrospectiveDiscussion(ctx context.Context, request *oapi.CreateRetrospectiveDiscussionRequest) (*oapi.CreateRetrospectiveDiscussionResponse, error) {
	var resp oapi.CreateRetrospectiveDiscussionResponse

	sess, sessErr := h.auth.GetSession(ctx)
	if sessErr != nil {
		return nil, detailError("failed to get session", sessErr)
	}

	discussion, createErr := h.retros.CreateDiscussion(ctx, rez.CreateRetrospectiveDiscussionParams{
		RetrospectiveID: request.Id,
		UserID:          sess.User.ID,
		Content:         request.Body.Attributes.Content,
	})
	if createErr != nil {
		return nil, detailError("failed to create retrospective discussion", createErr)
	}
	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(discussion)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveDiscussion(ctx context.Context, request *oapi.GetRetrospectiveDiscussionRequest) (*oapi.GetRetrospectiveDiscussionResponse, error) {
	var resp oapi.GetRetrospectiveDiscussionResponse

	discussion, discErr := h.retros.GetDiscussionByID(ctx, request.DiscussionId)
	if discErr != nil {
		return nil, detailError("failed to get retrospective discussion", discErr)
	}
	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(discussion)

	return &resp, nil
}

func (h *retrospectivesHandler) UpdateRetrospectiveDiscussion(ctx context.Context, request *oapi.UpdateRetrospectiveDiscussionRequest) (*oapi.UpdateRetrospectiveDiscussionResponse, error) {
	var resp oapi.UpdateRetrospectiveDiscussionResponse

	discussion, discErr := h.retros.GetDiscussionByID(ctx, request.DiscussionId)
	if discErr != nil {
		return nil, detailError("failed to get retrospective discussion", discErr)
	}

	update := discussion.Update()
	// TODO: update stuff
	//.SetNillableResolved(request.Body.Attributes.Resolved)
	updated, saveErr := update.Save(ctx)
	if saveErr != nil {
		return nil, detailError("failed to update retrospective discussion", saveErr)
	}

	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(updated)

	return &resp, nil
}

func (h *retrospectivesHandler) AddRetrospectiveDiscussionReply(ctx context.Context, request *oapi.AddRetrospectiveDiscussionReplyRequest) (*oapi.AddRetrospectiveDiscussionReplyResponse, error) {
	var resp oapi.AddRetrospectiveDiscussionReplyResponse

	sess, sessErr := h.auth.GetSession(ctx)
	if sessErr != nil {
		return nil, detailError("failed to get session", sessErr)
	}

	discussion, discussionErr := h.retros.GetDiscussionByID(ctx, request.DiscussionId)
	if discussionErr != nil {
		return nil, detailError("failed to get existing discussion", discussionErr)
	}

	attr := request.Body.Attributes
	reply, replyErr := h.retros.AddDiscussionReply(ctx, rez.AddRetrospectiveDiscussionReplyParams{
		DiscussionId: discussion.ID,
		UserID:       sess.User.ID,
		ParentID:     attr.ParentReplyId,
		Content:      attr.Content,
	})
	if replyErr != nil {
		return nil, detailError("failed to add retrospective discussion reply", replyErr)
	}
	discussion.Edges.Replies = append(discussion.Edges.Replies, reply)

	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(discussion)

	return &resp, nil
}
