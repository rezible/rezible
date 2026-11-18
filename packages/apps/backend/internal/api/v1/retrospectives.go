package apiv1

import (
	"context"
	"errors"
	"log/slog"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/integrations"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type retrospectivesHandler struct {
	retros       rez.RetrospectiveService
	integrations rez.IntegrationService
}

func newRetrospectivesHandler(retros rez.RetrospectiveService, integrationService rez.IntegrationService) *retrospectivesHandler {
	return &retrospectivesHandler{retros: retros, integrations: integrationService}
}

// retrospective converts the retrospective, with how its reviewers are told of a review request. Failing to
// work that out leaves it unknown rather than failing the request, which may already have changed the
// retrospective.
func (h *retrospectivesHandler) retrospective(ctx context.Context, retro *ent.Retrospective) oapi.Retrospective {
	dto := oapi.RetrospectiveFromEnt(retro)
	dto.Attributes.ReviewerMessages = oapi.RetrospectiveReviewerMessages{Delivery: "unknown"}
	senders, listErr := integrations.ListDirectMessageSenders(ctx, h.integrations)
	if listErr != nil {
		slog.WarnContext(ctx, "could not list direct message senders", "error", listErr)
		return dto
	}
	description, known := senders.Description()
	if !known {
		slog.WarnContext(ctx, "could not resolve an installation that might send direct messages",
			"error", errors.Join(senders.Unresolved...))
		return dto
	}
	if description == "" {
		dto.Attributes.ReviewerMessages.Delivery = "none"
		return dto
	}
	dto.Attributes.ReviewerMessages = oapi.RetrospectiveReviewerMessages{Delivery: "direct_message", Description: description}
	return dto
}

func (h *retrospectivesHandler) GetRetrospective(ctx context.Context, input *oapi.GetRetrospectiveRequest) (*oapi.GetRetrospectiveResponse, error) {
	retro, retroErr := h.retros.GetRetrospective(ctx, retrospective.ID(input.Id))
	if retroErr != nil {
		return nil, oapi.Error(ctx, "failed to get retrospective", retroErr)
	}
	var resp oapi.GetRetrospectiveResponse
	resp.Body.Data = h.retrospective(ctx, retro)
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
	resp.Body.Data = h.retrospective(ctx, retro)
	return &resp, nil
}

func (h *retrospectivesHandler) RequestRetrospectiveReview(ctx context.Context, input *oapi.RequestRetrospectiveReviewRequest) (*oapi.RequestRetrospectiveReviewResponse, error) {
	retro, updateErr := h.retros.RequestReview(ctx, input.Id, input.Body.Attributes.ReviewerIds)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to request retrospective review", updateErr)
	}
	var resp oapi.RequestRetrospectiveReviewResponse
	resp.Body.Data = h.retrospective(ctx, retro)
	return &resp, nil
}

func (h *retrospectivesHandler) AddRetrospectiveReviewer(ctx context.Context, input *oapi.AddRetrospectiveReviewerRequest) (*oapi.AddRetrospectiveReviewerResponse, error) {
	retro, updateErr := h.retros.AddReviewer(ctx, input.Id, input.Body.Attributes.UserId)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to add retrospective reviewer", updateErr)
	}
	var resp oapi.AddRetrospectiveReviewerResponse
	resp.Body.Data = h.retrospective(ctx, retro)
	return &resp, nil
}

func (h *retrospectivesHandler) RemoveRetrospectiveReviewer(ctx context.Context, input *oapi.RemoveRetrospectiveReviewerRequest) (*oapi.RemoveRetrospectiveReviewerResponse, error) {
	retro, updateErr := h.retros.RemoveReviewer(ctx, input.Id, input.UserId)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to remove retrospective reviewer", updateErr)
	}
	var resp oapi.RemoveRetrospectiveReviewerResponse
	resp.Body.Data = h.retrospective(ctx, retro)
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
	resp.Body.Data = h.retrospective(ctx, retro)
	return &resp, nil
}

func (h *retrospectivesHandler) CloseRetrospective(ctx context.Context, input *oapi.CloseRetrospectiveRequest) (*oapi.CloseRetrospectiveResponse, error) {
	retro, updateErr := h.retros.CloseReview(ctx, input.Id)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to close retrospective", updateErr)
	}
	var resp oapi.CloseRetrospectiveResponse
	resp.Body.Data = h.retrospective(ctx, retro)
	return &resp, nil
}

func (h *retrospectivesHandler) SkipRetrospective(ctx context.Context, input *oapi.SkipRetrospectiveRequest) (*oapi.SkipRetrospectiveResponse, error) {
	retro, updateErr := h.retros.SkipReview(ctx, input.Id, input.Body.Attributes.Reason)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to skip retrospective", updateErr)
	}
	var resp oapi.SkipRetrospectiveResponse
	resp.Body.Data = h.retrospective(ctx, retro)
	return &resp, nil
}
