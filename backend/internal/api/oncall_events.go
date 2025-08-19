package api

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallEventsHandler struct {
	auth      rez.AuthService
	users     rez.UserService
	oncall    rez.OncallService
	incidents rez.IncidentService
	events    rez.OncallEventsService
}

func newOncallEventsHandler(auth rez.AuthService, users rez.UserService, oncall rez.OncallService, inc rez.IncidentService, events rez.OncallEventsService) *oncallEventsHandler {
	return &oncallEventsHandler{auth: auth, users: users, oncall: oncall, incidents: inc, events: events}
}

func (h *oncallEventsHandler) GetOncallEvent(ctx context.Context, req *oapi.GetOncallEventRequest) (*oapi.GetOncallEventResponse, error) {
	var resp oapi.GetOncallEventResponse

	event, eventErr := h.events.GetEvent(ctx, req.Id)
	if eventErr != nil {
		return nil, apiError("failed to get oncall event", eventErr)
	}
	resp.Body.Data = oapi.OncallEventFromEnt(event)

	return &resp, nil
}

func (h *oncallEventsHandler) ListOncallEvents(ctx context.Context, req *oapi.ListOncallEventsRequest) (*oapi.ListOncallEventsResponse, error) {
	var resp oapi.ListOncallEventsResponse

	params := rez.ListOncallEventsParams{
		ListParams:         req.ListParams(),
		From:               req.From,
		To:                 req.To,
		RosterID:           req.RosterId,
		AnnotationRosterID: req.AnnotationRosterId,
		WithAnnotations:    req.WithAnnotations,
		AlertID:            req.AlertId,
	}

	if req.ShiftId != uuid.Nil {
		shift, shiftErr := h.oncall.GetShiftByID(ctx, req.ShiftId)
		if shiftErr != nil {
			return nil, apiError("failed to query shift", shiftErr)
		}
		params.From = shift.StartAt
		params.To = shift.EndAt
		params.RosterID = shift.RosterID
	}

	if rez.DebugMode {
		params.RosterID = uuid.Nil
	}

	events, count, eventsErr := h.events.ListEvents(ctx, params)
	if eventsErr != nil {
		return nil, apiError("failed to query events", eventsErr)
	}
	resp.Body.Data = make([]oapi.OncallEvent, len(events))
	for i, event := range events {
		resp.Body.Data[i] = oapi.OncallEventFromEnt(event)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Next:     nil,
		Previous: nil,
		Total:    count,
	}

	return &resp, nil
}

func (h *oncallEventsHandler) ListOncallAnnotations(ctx context.Context, request *oapi.ListOncallAnnotationsRequest) (*oapi.ListOncallAnnotationsResponse, error) {
	var resp oapi.ListOncallAnnotationsResponse

	params := rez.ListOncallAnnotationsParams{
		ListParams: request.ListParams(),
		RosterID:   request.RosterId,
		Expand: rez.ExpandAnnotationsParams{
			WithEvent: request.WithEvents,
		},
	}
	if request.ShiftId != uuid.Nil {
		shift, shiftErr := h.oncall.GetShiftByID(ctx, request.ShiftId)
		if shiftErr != nil {
			return nil, apiError("failed to query shift", shiftErr)
		}
		params.Shift = shift
	}

	annos, count, annosErr := h.events.ListAnnotations(ctx, params)
	if annosErr != nil {
		return nil, apiError("query shift annotations", annosErr)
	}

	resp.Body.Data = make([]oapi.OncallAnnotation, len(annos))
	for i, anno := range annos {
		resp.Body.Data[i] = oapi.OncallAnnotationFromEnt(anno)
	}

	resp.Body.Pagination = oapi.ResponsePagination{
		Next:     nil,
		Previous: nil,
		Total:    count,
	}

	return &resp, nil
}

func (h *oncallEventsHandler) CreateOncallAnnotation(ctx context.Context, request *oapi.CreateOncallAnnotationRequest) (*oapi.CreateOncallAnnotationResponse, error) {
	var resp oapi.CreateOncallAnnotationResponse

	user := getRequestAuthSession(ctx, h.auth)

	attr := request.Body.Attributes

	anno := &ent.OncallAnnotation{
		EventID:         attr.EventId,
		RosterID:        attr.RosterId,
		CreatorID:       user.UserId,
		MinutesOccupied: attr.MinutesOccupied,
		Notes:           attr.Notes,
		Tags:            attr.Tags,
	}

	var createErr error
	anno, createErr = h.events.UpdateAnnotation(ctx, anno)
	if createErr != nil {
		return nil, apiError("failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.OncallAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *oncallEventsHandler) UpdateOncallAnnotation(ctx context.Context, request *oapi.UpdateOncallAnnotationRequest) (*oapi.UpdateOncallAnnotationResponse, error) {
	var resp oapi.UpdateOncallAnnotationResponse

	attr := request.Body.Attributes
	anno, annoErr := h.events.GetAnnotation(ctx, request.Id)
	if annoErr != nil {
		return nil, apiError("failed to get annotation", annoErr)
	}

	update := anno.Update().
		SetNillableNotes(attr.Notes).
		SetNillableMinutesOccupied(attr.MinutesOccupied)

	if attr.Tags != nil {
		update.SetTags(*attr.Tags)
	}

	if attr.AlertFeedback != nil {

	}

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, apiError("failed to update annotation", updateErr)
	}
	resp.Body.Data = oapi.OncallAnnotationFromEnt(updated)

	return &resp, nil
}

func (h *oncallEventsHandler) DeleteOncallAnnotation(ctx context.Context, request *oapi.DeleteOncallAnnotationRequest) (*oapi.DeleteOncallAnnotationResponse, error) {
	var resp oapi.DeleteOncallAnnotationResponse

	if err := h.events.DeleteAnnotation(ctx, request.Id); err != nil {
		return nil, apiError("failed to archive annotation", err)
	}

	return &resp, nil
}
