package api

import (
	"context"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
	"time"
)

type oncallEventsHandler struct {
	auth      rez.AuthSessionService
	users     rez.UserService
	oncall    rez.OncallService
	incidents rez.IncidentService
	events    rez.OncallEventsService
}

func newOncallEventsHandler(auth rez.AuthSessionService, users rez.UserService, oncall rez.OncallService, inc rez.IncidentService, events rez.OncallEventsService) *oncallEventsHandler {
	return &oncallEventsHandler{auth: auth, users: users, oncall: oncall, incidents: inc, events: events}
}

func (h *oncallEventsHandler) ListOncallEvents(ctx context.Context, request *oapi.ListOncallEventsRequest) (*oapi.ListOncallEventsResponse, error) {
	var resp oapi.ListOncallEventsResponse

	params := rez.ListOncallEventsParams{
		Start: time.Now().Add(time.Hour * -24 * 7), // 7 days ago
		End:   time.Now(),
	}

	if request.ShiftId != uuid.Nil {
		shift, shiftErr := h.oncall.GetShiftByID(ctx, request.ShiftId)
		if shiftErr != nil {
			return nil, detailError("failed to query shift", shiftErr)
		}
		params.Start = shift.StartAt
		params.End = shift.EndAt
	}

	if request.RosterIds != nil {
		// TODO: handle this
	}

	events, eventsErr := h.events.ListEvents(ctx, params)
	if eventsErr != nil {
		return nil, detailError("failed to query events", eventsErr)
	}
	resp.Body.Data = make([]oapi.OncallEvent, len(events))
	for i, event := range events {
		resp.Body.Data[i] = oapi.OncallEventFromEnt(event)
	}

	return &resp, nil
}

func (h *oncallEventsHandler) ListOncallAnnotations(ctx context.Context, request *oapi.ListOncallAnnotationsRequest) (*oapi.ListOncallAnnotationsResponse, error) {
	var resp oapi.ListOncallAnnotationsResponse

	annos, annosErr := h.events.ListAnnotations(ctx, rez.ListOncallAnnotationsParams{
		ListParams: request.ListParams(),
		RosterID:   request.RosterId,
		ShiftID:    request.ShiftId,
	})
	if annosErr != nil {
		return nil, detailError("query shift annotations", annosErr)
	}

	resp.Body.Data = make([]oapi.OncallAnnotation, len(annos))
	for i, anno := range annos {
		resp.Body.Data[i] = oapi.OncallAnnotationFromEnt(anno)
	}

	return &resp, nil
}

func (h *oncallEventsHandler) CreateOncallAnnotation(ctx context.Context, request *oapi.CreateOncallAnnotationRequest) (*oapi.CreateOncallAnnotationResponse, error) {
	var resp oapi.CreateOncallAnnotationResponse

	attr := request.Body.Attributes

	anno := &ent.OncallAnnotation{
		EventID:         attr.EventId,
		RosterID:        attr.RosterId,
		MinutesOccupied: attr.MinutesOccupied,
		Notes:           attr.Notes,
	}

	var createErr error
	anno, createErr = h.events.CreateAnnotation(ctx, anno)
	if createErr != nil {
		return nil, detailError("failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.OncallAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *oncallEventsHandler) UpdateOncallAnnotation(ctx context.Context, request *oapi.UpdateOncallAnnotationRequest) (*oapi.UpdateOncallAnnotationResponse, error) {
	var resp oapi.UpdateOncallAnnotationResponse

	anno, annoErr := h.events.GetAnnotation(ctx, request.Id)
	if annoErr != nil {
		return nil, detailError("failed to get annotation", annoErr)
	}

	attr := request.Body.Attributes
	update := anno.Update().
		SetNillableNotes(attr.Notes).
		SetNillableMinutesOccupied(attr.MinutesOccupied)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update annotation", updateErr)
	}
	resp.Body.Data = oapi.OncallAnnotationFromEnt(updated)

	return &resp, nil
}

func (h *oncallEventsHandler) DeleteOncallAnnotation(ctx context.Context, request *oapi.DeleteOncallAnnotationRequest) (*oapi.DeleteOncallAnnotationResponse, error) {
	var resp oapi.DeleteOncallAnnotationResponse

	if err := h.events.DeleteAnnotation(ctx, request.Id); err != nil {
		return nil, detailError("failed to archive annotation", err)
	}

	return &resp, nil
}
