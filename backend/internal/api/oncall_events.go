package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"math/rand"
	"time"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallEventsHandler struct {
	auth      rez.AuthSessionService
	users     rez.UserService
	incidents rez.IncidentService
	oncall    rez.OncallService
	alerts    rez.AlertsService
}

func newOncallEventsHandler(auth rez.AuthSessionService, users rez.UserService, inc rez.IncidentService, oncall rez.OncallService, alerts rez.AlertsService) *oncallEventsHandler {
	return &oncallEventsHandler{auth: auth, users: users, incidents: inc, oncall: oncall, alerts: alerts}
}

func makeFakeShiftEvent(date time.Time) rez.OncallEvent {
	isAlert := rand.Float64() > 0.25
	eventKind := "incident"
	if isAlert {
		eventKind = "alert"
	}

	hour := rand.Intn(24)
	minute := rand.Intn(60)

	timestamp := time.Date(
		date.Year(), date.Month(), date.Day(),
		hour, minute, 0, 0, date.Location(),
	)

	description := "description"

	return rez.OncallEvent{
		ID:          uuid.New().String(),
		Title:       "title",
		Timestamp:   timestamp,
		Kind:        eventKind,
		Description: &description,
	}
}

func makeFakeOncallEvents(start time.Time) []rez.OncallEvent {
	const NumDays = 7
	events := make([]rez.OncallEvent, 0, NumDays*10)

	for day := 0; day < NumDays; day++ {
		dayDate := start.AddDate(0, 0, day)
		numDayEvents := rand.Intn(10)

		for i := 0; i < numDayEvents; i++ {
			events = append(events, makeFakeShiftEvent(dayDate))
		}
	}

	return events
}

func (h *oncallEventsHandler) ListOncallEvents(ctx context.Context, request *oapi.ListOncallEventsRequest) (*oapi.ListOncallEventsResponse, error) {
	var resp oapi.ListOncallEventsResponse

	eventsStart := time.Now().Add(time.Hour * -24 * 7)
	//if request.RosterId != uuid.Nil {
	//	shift, shiftErr := h.oncall.GetShiftByID(ctx, request.ShiftId)
	//	if shiftErr != nil {
	//		return nil, detailError("failed to query shift", shiftErr)
	//	}
	//	eventsStart = shift.StartAt
	//}

	resp.Body.Data = makeFakeOncallEvents(eventsStart)

	return &resp, nil
}

func (h *oncallEventsHandler) ListOncallEventAnnotations(ctx context.Context, request *oapi.ListOncallEventAnnotationsRequest) (*oapi.ListOncallEventAnnotationsResponse, error) {
	var resp oapi.ListOncallEventAnnotationsResponse

	annos, annosErr := h.oncall.ListEventAnnotations(ctx, rez.ListOncallEventAnnotationsParams{
		ListParams: request.ListParams(),
		RosterID:   request.RosterId,
		ShiftID:    request.ShiftId,
	})
	if annosErr != nil {
		return nil, detailError("query shift annotations", annosErr)
	}

	resp.Body.Data = make([]oapi.OncallEventAnnotation, len(annos))
	for i, anno := range annos {
		resp.Body.Data[i] = oapi.OncallEventAnnotationFromEnt(anno)
	}

	return &resp, nil
}

func (h *oncallEventsHandler) CreateOncallEventAnnotation(ctx context.Context, request *oapi.CreateOncallEventAnnotationRequest) (*oapi.CreateOncallEventAnnotationResponse, error) {
	var resp oapi.CreateOncallEventAnnotationResponse

	attr := request.Body.Attributes

	anno := &ent.OncallEventAnnotation{
		EventID:         attr.EventId,
		RosterID:        attr.RosterId,
		MinutesOccupied: attr.MinutesOccupied,
		Notes:           attr.Notes,
	}

	var createErr error
	anno, createErr = h.oncall.CreateEventAnnotation(ctx, anno)
	if createErr != nil {
		return nil, detailError("failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.OncallEventAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *oncallEventsHandler) UpdateOncallEventAnnotation(ctx context.Context, request *oapi.UpdateOncallEventAnnotationRequest) (*oapi.UpdateOncallEventAnnotationResponse, error) {
	var resp oapi.UpdateOncallEventAnnotationResponse

	anno, annoErr := h.oncall.GetEventAnnotation(ctx, request.Id)
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
	resp.Body.Data = oapi.OncallEventAnnotationFromEnt(updated)

	return &resp, nil
}

func (h *oncallEventsHandler) DeleteOncallEventAnnotation(ctx context.Context, request *oapi.DeleteOncallEventAnnotationRequest) (*oapi.DeleteOncallEventAnnotationResponse, error) {
	var resp oapi.DeleteOncallEventAnnotationResponse

	if err := h.oncall.DeleteEventAnnotation(ctx, request.Id); err != nil {
		return nil, detailError("failed to archive annotation", err)
	}

	return &resp, nil
}
