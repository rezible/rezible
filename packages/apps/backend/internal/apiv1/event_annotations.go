package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type eventAnnotationsHandler struct {
	auth  rez.AuthSessionService
	annos rez.EventAnnotationsService
}

func newEventAnnotationsHandler(auth rez.AuthSessionService, annos rez.EventAnnotationsService) *eventAnnotationsHandler {
	return &eventAnnotationsHandler{auth: auth, annos: annos}
}

func (h *eventAnnotationsHandler) ListEventAnnotations(ctx context.Context, req *oapi.ListEventAnnotationsRequest) (*oapi.ListEventAnnotationsResponse, error) {
	var resp oapi.ListEventAnnotationsResponse

	var uids []uuid.UUID
	var eventIds []uuid.UUID

	params := rez.ListAnnotationsParams{
		ListParams: req.ListParams(),
		From:       req.From,
		To:         req.To,
		UserIds:    uids,
		EventIds:   eventIds,
		Expand: rez.ExpandAnnotationsParams{
			WithEvent: req.WithEvents,
		},
	}

	listRes, annosErr := h.annos.ListAnnotations(ctx, params)
	if annosErr != nil {
		return nil, oapi.Error("query shift annotations", annosErr)
	}

	resp.Body.Data = make([]oapi.EventAnnotation, len(listRes.Data))
	for i, anno := range listRes.Data {
		resp.Body.Data[i] = oapi.EventAnnotationFromEnt(anno)
	}

	resp.Body.Pagination = oapi.ResponsePagination{
		Total: listRes.Count,
	}

	return &resp, nil
}

func (h *eventAnnotationsHandler) CreateEventAnnotation(ctx context.Context, request *oapi.CreateEventAnnotationRequest) (*oapi.CreateEventAnnotationResponse, error) {
	var resp oapi.CreateEventAnnotationResponse

	userId := h.auth.GetAuthSession(ctx).UserId

	attr := request.Body.Attributes

	anno := &ent.EventAnnotation{
		CreatorID:       userId,
		EventID:         attr.EventId,
		MinutesOccupied: attr.MinutesOccupied,
		Notes:           attr.Notes,
		Tags:            attr.Tags,
	}

	var createErr error
	anno, createErr = h.annos.SetAnnotation(ctx, anno)
	if createErr != nil {
		return nil, oapi.Error("failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.EventAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *eventAnnotationsHandler) UpdateEventAnnotation(ctx context.Context, request *oapi.UpdateEventAnnotationRequest) (*oapi.UpdateEventAnnotationResponse, error) {
	var resp oapi.UpdateEventAnnotationResponse

	attr := request.Body.Attributes
	anno, annoErr := h.annos.GetAnnotation(ctx, request.Id)
	if annoErr != nil {
		return nil, oapi.Error("failed to get annotation", annoErr)
	}

	update := anno.Update().
		SetNillableNotes(attr.Notes).
		SetNillableMinutesOccupied(attr.MinutesOccupied)

	if attr.Tags != nil {
		update.SetTags(*attr.Tags)
	}

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error("failed to update annotation", updateErr)
	}
	resp.Body.Data = oapi.EventAnnotationFromEnt(updated)

	return &resp, nil
}

func (h *eventAnnotationsHandler) DeleteEventAnnotation(ctx context.Context, request *oapi.DeleteEventAnnotationRequest) (*oapi.DeleteEventAnnotationResponse, error) {
	var resp oapi.DeleteEventAnnotationResponse

	if err := h.annos.DeleteAnnotation(ctx, request.Id); err != nil {
		return nil, oapi.Error("failed to archive annotation", err)
	}

	return &resp, nil
}
