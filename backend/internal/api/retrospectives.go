package api

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	oapi "github.com/rezible/rezible/openapi"
)

type retrospectivesHandler struct {
	auth      rez.AuthSessionService
	users     rez.UserService
	incidents rez.IncidentService
	retros    rez.RetrospectiveService
	documents rez.DocumentsService
}

func newRetrospectivesHandler(auth rez.AuthSessionService, users rez.UserService, incidents rez.IncidentService, retros rez.RetrospectiveService, documents rez.DocumentsService) *retrospectivesHandler {
	return &retrospectivesHandler{auth, users, incidents, retros, documents}
}

func (h *retrospectivesHandler) ListRetrospectives(ctx context.Context, input *oapi.ListRetrospectivesRequest) (*oapi.ListRetrospectivesResponse, error) {
	var resp oapi.ListRetrospectivesResponse

	return &resp, nil
}

func (h *retrospectivesHandler) CreateRetrospective(ctx context.Context, request *oapi.CreateRetrospectiveRequest) (*oapi.CreateRetrospectiveResponse, error) {
	var resp oapi.CreateRetrospectiveResponse

	attrs := request.Body.Attributes
	params := ent.Retrospective{
		IncidentID: attrs.IncidentId,
		Type:       retrospective.TypeSimple,
	}
	if attrs.SystemAnalysis {
		params.Type = retrospective.TypeFull
	}

	retro, createErr := h.retros.Create(ctx, params)
	if createErr != nil {
		return nil, apiError("failed to create retro", createErr)
	}
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	var resp oapi.GetRetrospectiveResponse

	retro, retroErr := h.retros.GetById(ctx, input.Id)
	if retroErr != nil {
		return nil, apiError("failed to get retrospective", retroErr)
	}
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveForIncident(ctx context.Context, input *oapi.GetRetrospectiveForIncidentRequest) (*oapi.GetRetrospectiveForIncidentResponse, error) {
	var resp oapi.GetRetrospectiveForIncidentResponse

	var inc *ent.Incident
	var incErr error
	if input.Id.IsSlug {
		inc, incErr = h.incidents.GetBySlug(ctx, input.Id.Slug)
	} else {
		inc, incErr = h.incidents.GetByID(ctx, input.Id.UUID)
	}
	if incErr != nil {
		return nil, apiError("failed to get incident", incErr)
	}

	retro, retroErr := h.retros.GetForIncident(ctx, inc)
	if retroErr != nil {
		return nil, apiError("failed to get retrospective", retroErr)
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

//func (h *retrospectivesHandler) ListRetrospectiveTemplates(ctx context.Context, request *oapi.ListRetrospectiveTemplatesRequest) (*oapi.ListRetrospectiveTemplatesResponse, error) {
//	var resp oapi.ListRetrospectiveTemplatesResponse
//	// TODO
//	return &resp, nil
//}
//func (h *retrospectivesHandler) CreateRetrospectiveTemplate(ctx context.Context, request *oapi.CreateRetrospectiveTemplateRequest) (*oapi.CreateRetrospectiveTemplateResponse, error) {
//	var resp oapi.CreateRetrospectiveTemplateResponse
//	// TODO
//	return &resp, nil
//}
//func (h *retrospectivesHandler) UpdateRetrospectiveTemplate(ctx context.Context, request *oapi.UpdateRetrospectiveTemplateRequest) (*oapi.UpdateRetrospectiveTemplateResponse, error) {
//	var resp oapi.UpdateRetrospectiveTemplateResponse
//	// TODO
//	return &resp, nil
//}
//func (h *retrospectivesHandler) ArchiveRetrospectiveTemplate(ctx context.Context, request *oapi.ArchiveRetrospectiveTemplateRequest) (*oapi.ArchiveRetrospectiveTemplateResponse, error) {
//	var resp oapi.ArchiveRetrospectiveTemplateResponse
//	// TODO
//	return &resp, nil
//}

func (h *retrospectivesHandler) ListRetrospectiveDiscussions(ctx context.Context, request *oapi.ListRetrospectiveDiscussionsRequest) (*oapi.ListRetrospectiveDiscussionsResponse, error) {
	var resp oapi.ListRetrospectiveDiscussionsResponse

	discussions, discErr := h.retros.ListDiscussions(ctx, rez.ListRetrospectiveDiscussionsParams{
		RetrospectiveID: request.Id,
		WithReplies:     true,
		ListParams:      request.ListParams(),
	})
	if discErr != nil {
		return nil, apiError("failed to list discussions", discErr)
	}

	resp.Body.Data = make([]oapi.RetrospectiveDiscussion, len(discussions))
	for i, disc := range discussions {
		resp.Body.Data[i] = oapi.RetrospectiveDiscussionFromEnt(disc)
	}

	return &resp, nil
}

func (h *retrospectivesHandler) CreateRetrospectiveDiscussion(ctx context.Context, request *oapi.CreateRetrospectiveDiscussionRequest) (*oapi.CreateRetrospectiveDiscussionResponse, error) {
	var resp oapi.CreateRetrospectiveDiscussionResponse

	userId := requestUserId(ctx, h.auth)

	discussion, createErr := h.retros.CreateDiscussion(ctx, rez.CreateRetrospectiveDiscussionParams{
		RetrospectiveID: request.Id,
		UserID:          userId,
		Content:         request.Body.Attributes.Content,
	})
	if createErr != nil {
		return nil, apiError("failed to create retrospective discussion", createErr)
	}
	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(discussion)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveDiscussion(ctx context.Context, request *oapi.GetRetrospectiveDiscussionRequest) (*oapi.GetRetrospectiveDiscussionResponse, error) {
	var resp oapi.GetRetrospectiveDiscussionResponse

	discussion, discErr := h.retros.GetDiscussionByID(ctx, request.DiscussionId)
	if discErr != nil {
		return nil, apiError("failed to get retrospective discussion", discErr)
	}
	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(discussion)

	return &resp, nil
}

func (h *retrospectivesHandler) UpdateRetrospectiveDiscussion(ctx context.Context, request *oapi.UpdateRetrospectiveDiscussionRequest) (*oapi.UpdateRetrospectiveDiscussionResponse, error) {
	var resp oapi.UpdateRetrospectiveDiscussionResponse

	discussion, discErr := h.retros.GetDiscussionByID(ctx, request.DiscussionId)
	if discErr != nil {
		return nil, apiError("failed to get retrospective discussion", discErr)
	}

	update := discussion.Update()
	// TODO: update stuff
	//.SetNillableResolved(request.Body.Attributes.Resolved)
	updated, saveErr := update.Save(ctx)
	if saveErr != nil {
		return nil, apiError("failed to update retrospective discussion", saveErr)
	}

	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(updated)

	return &resp, nil
}

func (h *retrospectivesHandler) AddRetrospectiveDiscussionReply(ctx context.Context, request *oapi.AddRetrospectiveDiscussionReplyRequest) (*oapi.AddRetrospectiveDiscussionReplyResponse, error) {
	var resp oapi.AddRetrospectiveDiscussionReplyResponse

	userId := requestUserId(ctx, h.auth)

	discussion, discussionErr := h.retros.GetDiscussionByID(ctx, request.DiscussionId)
	if discussionErr != nil {
		return nil, apiError("failed to get existing discussion", discussionErr)
	}

	attr := request.Body.Attributes
	reply, replyErr := h.retros.AddDiscussionReply(ctx, rez.AddRetrospectiveDiscussionReplyParams{
		DiscussionId: discussion.ID,
		UserID:       userId,
		ParentID:     attr.ParentReplyId,
		Content:      attr.Content,
	})
	if replyErr != nil {
		return nil, apiError("failed to add retrospective discussion reply", replyErr)
	}
	discussion.Edges.Replies = append(discussion.Edges.Replies, reply)

	resp.Body.Data = oapi.RetrospectiveDiscussionFromEnt(discussion)

	return &resp, nil
}
