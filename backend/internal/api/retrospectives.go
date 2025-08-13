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

func (h *retrospectivesHandler) ListRetrospectiveComments(ctx context.Context, request *oapi.ListRetrospectiveCommentsRequest) (*oapi.ListRetrospectiveCommentsResponse, error) {
	var resp oapi.ListRetrospectiveCommentsResponse

	comments, listErr := h.retros.ListComments(ctx, rez.ListRetrospectiveCommentsParams{
		ListParams:      request.ListParams(),
		RetrospectiveID: request.Id,
		WithReplies:     true,
	})
	if listErr != nil {
		return nil, apiError("failed to list comments", listErr)
	}

	resp.Body.Data = make([]oapi.RetrospectiveComment, len(comments))
	for i, disc := range comments {
		resp.Body.Data[i] = oapi.RetrospectiveCommentFromEnt(disc)
	}

	return &resp, nil
}

func (h *retrospectivesHandler) CreateRetrospectiveComment(ctx context.Context, request *oapi.CreateRetrospectiveCommentRequest) (*oapi.CreateRetrospectiveCommentResponse, error) {
	var resp oapi.CreateRetrospectiveCommentResponse

	userId := requestUserId(ctx, h.auth)

	comment, createErr := h.retros.SetComment(ctx, &ent.RetrospectiveComment{
		RetrospectiveID: request.Id,
		UserID:          userId,
		Content:         request.Body.Attributes.Content,
	})
	if createErr != nil {
		return nil, apiError("failed to create retrospective comment", createErr)
	}
	resp.Body.Data = oapi.RetrospectiveCommentFromEnt(comment)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveComment(ctx context.Context, request *oapi.GetRetrospectiveCommentRequest) (*oapi.GetRetrospectiveCommentResponse, error) {
	var resp oapi.GetRetrospectiveCommentResponse

	comment, getErr := h.retros.GetComment(ctx, request.Id)
	if getErr != nil {
		return nil, apiError("get comment", getErr)
	}
	resp.Body.Data = oapi.RetrospectiveCommentFromEnt(comment)

	return &resp, nil
}

func (h *retrospectivesHandler) UpdateRetrospectiveComment(ctx context.Context, request *oapi.UpdateRetrospectiveCommentRequest) (*oapi.UpdateRetrospectiveCommentResponse, error) {
	var resp oapi.UpdateRetrospectiveCommentResponse

	comment, getErr := h.retros.GetComment(ctx, request.Id)
	if getErr != nil {
		return nil, apiError("get comment", getErr)
	}

	attr := request.Body.Attributes
	// TODO: set fields to update
	if attr.Content != nil {
		comment.Content = []byte(*attr.Content)
	}

	updated, saveErr := h.retros.SetComment(ctx, comment)
	if saveErr != nil {
		return nil, apiError("update comment", saveErr)
	}
	resp.Body.Data = oapi.RetrospectiveCommentFromEnt(updated)

	return &resp, nil
}
