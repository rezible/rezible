package api

import (
	"context"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type incidentEventsHandler struct {
	db *ent.Client
}

func newIncidentEventsHandler(db *ent.Client) *incidentEventsHandler {
	return &incidentEventsHandler{db}
}

func (i *incidentEventsHandler) ListIncidentEvents(ctx context.Context, request *oapi.ListIncidentEventsRequest) (*oapi.ListIncidentEventsResponse, error) {
	var resp oapi.ListIncidentEventsResponse
	
	return &resp, nil
}

func (i *incidentEventsHandler) CreateIncidentEvent(ctx context.Context, request *oapi.CreateIncidentEventRequest) (*oapi.CreateIncidentEventResponse, error) {
	var resp oapi.CreateIncidentEventResponse

	return &resp, nil
}

func (i *incidentEventsHandler) UpdateIncidentEvent(ctx context.Context, request *oapi.UpdateIncidentEventRequest) (*oapi.UpdateIncidentEventResponse, error) {
	var resp oapi.UpdateIncidentEventResponse

	return &resp, nil
}

func (i *incidentEventsHandler) DeleteIncidentEvent(ctx context.Context, request *oapi.DeleteIncidentEventRequest) (*oapi.DeleteIncidentEventResponse, error) {
	var resp oapi.DeleteIncidentEventResponse

	return &resp, nil
}
