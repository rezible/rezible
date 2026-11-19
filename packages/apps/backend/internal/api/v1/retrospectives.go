package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type retrospectivesHandler struct{ retros rez.RetrospectiveService }

func newRetrospectivesHandler(retros rez.RetrospectiveService) *retrospectivesHandler {
	return &retrospectivesHandler{retros: retros}
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	retro, retroErr := h.retros.GetRetrospective(ctx, retrospective.ID(input.Id))
	if retroErr != nil {
		return nil, oapi.Error(ctx, "failed to get retrospective", retroErr)
	}
	var resp oapi.GetRetrospectiveResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) GetRetrospectiveReportComposition(ctx context.Context, input *oapi.GetRetrospectiveReportCompositionRequest) (*oapi.GetRetrospectiveReportCompositionResponse, error) {
	composition, queryErr := h.retros.GetReportComposition(ctx, input.Id)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get report composition", queryErr)
	}
	var resp oapi.GetRetrospectiveReportCompositionResponse
	resp.Body.Data = oapi.RetrospectiveReportCompositionFromRez(composition)
	return &resp, nil
}

func (h *retrospectivesHandler) SetRetrospectiveReportFindingSelection(ctx context.Context, input *oapi.SetRetrospectiveReportFindingSelectionRequest) (*oapi.SetRetrospectiveReportFindingSelectionResponse, error) {
	return nil, oapi.Error(ctx, "not implemented", errs.ErrNotImplemented)
}

func (h *retrospectivesHandler) SetRetrospectiveOwner(ctx context.Context, input *oapi.SetRetrospectiveOwnerRequest) (*oapi.SetRetrospectiveOwnerResponse, error) {
	retro, updateErr := h.retros.SetOwner(ctx, input.Id, input.Body.Attributes.UserId)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to set retrospective owner", updateErr)
	}
	var resp oapi.SetRetrospectiveOwnerResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) RequestRetrospectiveReview(ctx context.Context, input *oapi.RequestRetrospectiveReviewRequest) (*oapi.RequestRetrospectiveReviewResponse, error) {
	retro, updateErr := h.retros.RequestReview(ctx, input.Id, input.Body.Attributes.ReviewerIds)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to request retrospective review", updateErr)
	}
	var resp oapi.RequestRetrospectiveReviewResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) AddRetrospectiveReviewer(ctx context.Context, input *oapi.AddRetrospectiveReviewerRequest) (*oapi.AddRetrospectiveReviewerResponse, error) {
	retro, updateErr := h.retros.AddReviewer(ctx, input.Id, input.Body.Attributes.UserId)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to add retrospective reviewer", updateErr)
	}
	var resp oapi.AddRetrospectiveReviewerResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) RemoveRetrospectiveReviewer(ctx context.Context, input *oapi.RemoveRetrospectiveReviewerRequest) (*oapi.RemoveRetrospectiveReviewerResponse, error) {
	retro, updateErr := h.retros.RemoveReviewer(ctx, input.Id, input.UserId)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to remove retrospective reviewer", updateErr)
	}
	var resp oapi.RemoveRetrospectiveReviewerResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) SetRetrospectiveReviewDecision(ctx context.Context, input *oapi.SetRetrospectiveReviewDecisionRequest) (*oapi.SetRetrospectiveReviewDecisionResponse, error) {
	userID, hasUser := execution.GetContext(ctx).UserID()
	if !hasUser {
		return nil, oapi.Error(ctx, "set review decision", errs.ErrAuthSessionMissing)
	}
	params := rez.SetReviewDecisionParams{
		ReviewerID: userID,
		Decision:   input.Body.Attributes.Decision,
		Feedback:   input.Body.Attributes.Feedback,
	}
	retro, updateErr := h.retros.SetReviewDecision(ctx, input.Id, params)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to set review decision", updateErr)
	}
	var resp oapi.SetRetrospectiveReviewDecisionResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) CloseRetrospective(ctx context.Context, input *oapi.CloseRetrospectiveRequest) (*oapi.CloseRetrospectiveResponse, error) {
	retro, updateErr := h.retros.CloseReview(ctx, input.Id)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to close retrospective", updateErr)
	}
	var resp oapi.CloseRetrospectiveResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}

func (h *retrospectivesHandler) SkipRetrospective(ctx context.Context, input *oapi.SkipRetrospectiveRequest) (*oapi.SkipRetrospectiveResponse, error) {
	retro, updateErr := h.retros.SkipReview(ctx, input.Id, input.Body.Attributes.Reason)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to skip retrospective", updateErr)
	}
	var resp oapi.SkipRetrospectiveResponse
	resp.Body.Data = oapi.RetrospectiveFromEnt(retro)
	return &resp, nil
}
