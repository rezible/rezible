package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type eventsHandler struct {
	events rez.EventsService
}

func newEventsHandler(events rez.EventsService) *eventsHandler {
	return &eventsHandler{events: events}
}

func (h *eventsHandler) GetEvent(ctx context.Context, req *oapi.GetEventRequest) (*oapi.GetEventResponse, error) {
	var resp oapi.GetEventResponse

	event, eventErr := h.events.GetEvent(ctx, req.Id)
	if eventErr != nil {
		return nil, oapi.Error(ctx, "failed to get event", eventErr)
	}
	resp.Body.Data = oapi.EventFromEnt(event)

	return &resp, nil
}

func (h *eventsHandler) ListEvents(ctx context.Context, req *oapi.ListEventsRequest) (*oapi.ListEventsResponse, error) {
	var resp oapi.ListEventsResponse

	params := rez.ListEventsParams{
		ListParams: req.ListParams(),
	}
	if !req.From.IsZero() {
		params.Predicates = append(params.Predicates, ne.OccurredAtGTE(req.From))
	}
	if !req.To.IsZero() {
		params.Predicates = append(params.Predicates, ne.OccurredAtLTE(req.To))
	}

	listRes, eventsErr := h.events.ListEvents(ctx, params)
	if eventsErr != nil {
		return nil, oapi.Error(ctx, "failed to query events", eventsErr)
	}

	resp.Body.Data = make([]oapi.Event, len(listRes.Data))
	for i, event := range listRes.Data {
		resp.Body.Data[i] = oapi.EventFromEnt(event)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: listRes.Count,
	}

	return &resp, nil
}

func (h *eventsHandler) ListEventAnnotations(ctx context.Context, req *oapi.ListEventAnnotationsRequest) (*oapi.ListEventAnnotationsResponse, error) {
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

	listRes, annosErr := h.events.ListAnnotations(ctx, params)
	if annosErr != nil {
		return nil, oapi.Error(ctx, "query shift annotations", annosErr)
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

func (h *eventsHandler) CreateEventAnnotation(ctx context.Context, request *oapi.CreateEventAnnotationRequest) (*oapi.CreateEventAnnotationResponse, error) {
	var resp oapi.CreateEventAnnotationResponse

	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, oapi.Error(ctx, "failed to get auth session", rez.ErrAuthSessionMissing)
	}

	attr := request.Body.Attributes

	anno := &ent.EventAnnotation{
		CreatorID:       userId,
		EventID:         attr.EventId,
		MinutesOccupied: attr.MinutesOccupied,
		Notes:           attr.Notes,
		Tags:            attr.Tags,
	}

	var createErr error
	anno, createErr = h.events.SetAnnotation(ctx, anno)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.EventAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *eventsHandler) UpdateEventAnnotation(ctx context.Context, request *oapi.UpdateEventAnnotationRequest) (*oapi.UpdateEventAnnotationResponse, error) {
	var resp oapi.UpdateEventAnnotationResponse

	attr := request.Body.Attributes
	anno, annoErr := h.events.GetAnnotation(ctx, request.Id)
	if annoErr != nil {
		return nil, oapi.Error(ctx, "failed to get annotation", annoErr)
	}

	update := anno.Update().
		SetNillableNotes(attr.Notes).
		SetNillableMinutesOccupied(attr.MinutesOccupied)

	if attr.Tags != nil {
		update.SetTags(*attr.Tags)
	}

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, oapi.Error(ctx, "failed to update annotation", updateErr)
	}
	resp.Body.Data = oapi.EventAnnotationFromEnt(updated)

	return &resp, nil
}

func (h *eventsHandler) DeleteEventAnnotation(ctx context.Context, request *oapi.DeleteEventAnnotationRequest) (*oapi.DeleteEventAnnotationResponse, error) {
	var resp oapi.DeleteEventAnnotationResponse

	if err := h.events.DeleteAnnotation(ctx, request.Id); err != nil {
		return nil, oapi.Error(ctx, "failed to archive annotation", err)
	}

	return &resp, nil
}
