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

func makeFakeShiftEvent(date time.Time) oapi.OncallEvent {
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

	return oapi.OncallEvent{
		Id: uuid.New().String(),
		Attributes: oapi.OncallEventAttributes{
			Title:       "title",
			Timestamp:   timestamp,
			Kind:        eventKind,
			Annotations: make([]oapi.OncallAnnotation, 0),
		},
	}
}

func makeFakeOncallEvents(start time.Time) []oapi.OncallEvent {
	const NumDays = 7
	events := make([]oapi.OncallEvent, 0, NumDays*10)

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

func (h *oncallEventsHandler) ListOncallAnnotations(ctx context.Context, request *oapi.ListOncallAnnotationsRequest) (*oapi.ListOncallAnnotationsResponse, error) {
	var resp oapi.ListOncallAnnotationsResponse

	annos, annosErr := h.oncall.ListAnnotations(ctx, rez.ListOncallAnnotationsParams{
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
	anno, createErr = h.oncall.CreateAnnotation(ctx, anno)
	if createErr != nil {
		return nil, detailError("failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.OncallAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *oncallEventsHandler) UpdateOncallAnnotation(ctx context.Context, request *oapi.UpdateOncallAnnotationRequest) (*oapi.UpdateOncallAnnotationResponse, error) {
	var resp oapi.UpdateOncallAnnotationResponse

	anno, annoErr := h.oncall.GetAnnotation(ctx, request.Id)
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

	if err := h.oncall.DeleteAnnotation(ctx, request.Id); err != nil {
		return nil, detailError("failed to archive annotation", err)
	}

	return &resp, nil
}
