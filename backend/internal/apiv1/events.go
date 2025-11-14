package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type eventsHandler struct {
	auth   rez.AuthService
	events rez.EventsService
}

func newEventsHandler(auth rez.AuthService, events rez.EventsService) *eventsHandler {
	return &eventsHandler{auth: auth, events: events}
}

func (h *eventsHandler) GetEvent(ctx context.Context, req *oapi.GetEventRequest) (*oapi.GetEventResponse, error) {
	var resp oapi.GetEventResponse

	event, eventErr := h.events.GetEvent(ctx, req.Id)
	if eventErr != nil {
		return nil, apiError("failed to get event", eventErr)
	}
	resp.Body.Data = oapi.EventFromEnt(event)

	return &resp, nil
}

func (h *eventsHandler) ListEvents(ctx context.Context, req *oapi.ListEventsRequest) (*oapi.ListEventsResponse, error) {
	var resp oapi.ListEventsResponse

	params := rez.ListEventsParams{
		ListParams: req.ListParams(),
		From:       req.From,
		To:         req.To,
	}

	listRes, eventsErr := h.events.ListEvents(ctx, params)
	if eventsErr != nil {
		return nil, apiError("failed to query events", eventsErr)
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
