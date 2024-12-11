package api

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/incidentevent"
	oapi "github.com/twohundreds/rezible/openapi"
)

type incidentEventsHandler struct {
	events *ent.IncidentEventClient
}

func newIncidentEventsHandler(events *ent.IncidentEventClient) *incidentEventsHandler {
	return &incidentEventsHandler{events}
}

func (h *incidentEventsHandler) ListIncidentEvents(ctx context.Context, input *oapi.ListIncidentEventsRequest) (*oapi.ListIncidentEventsResponse, error) {
	var resp oapi.ListIncidentEventsResponse

	query := h.events.Query()

	query.Limit(10)
	query.Offset(0)

	results, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query incident events", queryErr)
	}

	resp.Body.Data = make([]oapi.IncidentEvent, len(results))
	for i, ev := range results {
		resp.Body.Data[i] = oapi.IncidentEventFromEnt(ev)
	}

	return &resp, nil
}

func (h *incidentEventsHandler) CreateIncidentEvent(ctx context.Context, input *oapi.CreateIncidentEventRequest) (*oapi.CreateIncidentEventResponse, error) {
	var resp oapi.CreateIncidentEventResponse

	attrs := input.Body.Attributes
	query := h.events.Create().
		SetType(incidentevent.Type(attrs.Type)).
		SetTime(attrs.StartTime)

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentEventFromEnt(ev)

	return &resp, nil
}

func (h *incidentEventsHandler) UpdateIncidentEvent(ctx context.Context, input *oapi.UpdateIncidentEventRequest) (*oapi.UpdateIncidentEventResponse, error) {
	var resp oapi.UpdateIncidentEventResponse

	attrs := input.Body.Attributes

	query := h.events.UpdateOneID(input.Id).
		SetNillableTime(attrs.StartTime)

	if attrs.Type != nil {
		query.SetType(incidentevent.Type(*attrs.Type))
	}

	ev, createErr := query.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to update incident event", createErr)
	}
	resp.Body.Data = oapi.IncidentEventFromEnt(ev)

	return &resp, nil
}

func (h *incidentEventsHandler) ArchiveIncidentEvent(ctx context.Context, input *oapi.ArchiveIncidentEventRequest) (*oapi.ArchiveIncidentEventResponse, error) {
	var resp oapi.ArchiveIncidentEventResponse

	deleteErr := h.events.DeleteOneID(input.Id).Exec(ctx)
	if deleteErr != nil {
		return nil, detailError("failed to archive incident event", deleteErr)
	}

	return &resp, nil
}
