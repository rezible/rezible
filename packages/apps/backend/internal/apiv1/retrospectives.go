package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	oapi "github.com/rezible/rezible/openapi/v1"
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

func (h *retrospectivesHandler) UpdateRetrospective(ctx context.Context, req *oapi.UpdateRetrospectiveRequest) (*oapi.UpdateRetrospectiveResponse, error) {
	var resp oapi.UpdateRetrospectiveResponse

	//attrs := req.Body.Attributes
	setFn := func(m *ent.RetrospectiveMutation) {
		//kind := retrospective.KindSimple
		//if attrs.SystemAnalysis {
		//	kind = retrospective.TypeFull
		//}
	}
	retro, updateErr := h.retros.Set(ctx, req.Id, setFn)
	if updateErr != nil {
		return nil, oapi.Error("update retrospective", updateErr)
	}
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	var resp oapi.GetRetrospectiveResponse

	retro, retroErr := h.retros.Get(ctx, retrospective.ID(input.Id))
	if retroErr != nil {
		return nil, oapi.Error("failed to get retrospective", retroErr)
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
		return nil, oapi.Error("list retrospective comments", listErr)
	}

	resp.Body.Data = make([]oapi.RetrospectiveComment, len(comments))
	for i, disc := range comments {
		resp.Body.Data[i] = oapi.RetrospectiveCommentFromEnt(disc)
	}

	return &resp, nil
}

func (h *retrospectivesHandler) CreateRetrospectiveComment(ctx context.Context, request *oapi.CreateRetrospectiveCommentRequest) (*oapi.CreateRetrospectiveCommentResponse, error) {
	var resp oapi.CreateRetrospectiveCommentResponse

	comment, createErr := h.retros.SetComment(ctx, &ent.RetrospectiveComment{
		RetrospectiveID: request.Id,
		UserID:          h.auth.GetAuthSession(ctx).UserId,
		Content:         request.Body.Attributes.Content,
	})
	if createErr != nil {
		return nil, oapi.Error("create retrospective comment", createErr)
	}
	resp.Body.Data = oapi.RetrospectiveCommentFromEnt(comment)

	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveComment(ctx context.Context, request *oapi.GetRetrospectiveCommentRequest) (*oapi.GetRetrospectiveCommentResponse, error) {
	var resp oapi.GetRetrospectiveCommentResponse

	comment, getErr := h.retros.GetComment(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error("get retrospective comment", getErr)
	}
	resp.Body.Data = oapi.RetrospectiveCommentFromEnt(comment)

	return &resp, nil
}

func (h *retrospectivesHandler) UpdateRetrospectiveComment(ctx context.Context, request *oapi.UpdateRetrospectiveCommentRequest) (*oapi.UpdateRetrospectiveCommentResponse, error) {
	var resp oapi.UpdateRetrospectiveCommentResponse

	comment, getErr := h.retros.GetComment(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error("get retrospective comment", getErr)
	}

	attr := request.Body.Attributes
	// TODO: set fields to update
	if attr.Content != nil {
		comment.Content = []byte(*attr.Content)
	}

	updated, saveErr := h.retros.SetComment(ctx, comment)
	if saveErr != nil {
		return nil, oapi.Error("update retrospective comment", saveErr)
	}
	resp.Body.Data = oapi.RetrospectiveCommentFromEnt(updated)

	return &resp, nil
}
