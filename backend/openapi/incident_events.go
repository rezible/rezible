package openapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type IncidentEventsHandler interface {
	ListIncidentEvents(context.Context, *ListIncidentEventsRequest) (*ListIncidentEventsResponse, error)
	CreateIncidentEvent(context.Context, *CreateIncidentEventRequest) (*CreateIncidentEventResponse, error)
	UpdateIncidentEvent(context.Context, *UpdateIncidentEventRequest) (*UpdateIncidentEventResponse, error)
	ArchiveIncidentEvent(context.Context, *ArchiveIncidentEventRequest) (*ArchiveIncidentEventResponse, error)
}

func (o operations) RegisterIncidentEvents(api huma.API) {
	huma.Register(api, ListIncidentEvents, o.ListIncidentEvents)
	huma.Register(api, CreateIncidentEvent, o.CreateIncidentEvent)
	huma.Register(api, UpdateIncidentEvent, o.UpdateIncidentEvent)
	huma.Register(api, ArchiveIncidentEvent, o.ArchiveIncidentEvent)
}

type IncidentEvent struct {
	Id         uuid.UUID               `json:"id"`
	Attributes IncidentEventAttributes `json:"attributes"`
}

type IncidentEventAttributes struct {
	Type       string     `json:"type" enum:"default,incident"`
	Title      string     `json:"title"`
	StartTime  time.Time  `json:"start_time"`
	IncidentId *uuid.UUID `json:"incident_id,omitempty"`
}

func IncidentEventFromEnt(ev *ent.IncidentEvent) IncidentEvent {
	var incidentId *uuid.UUID
	if ev.IncidentID != uuid.Nil {
		incidentId = &ev.IncidentID
	}
	return IncidentEvent{
		Id: ev.ID,
		Attributes: IncidentEventAttributes{
			Type:       ev.Type.String(),
			StartTime:  ev.Time,
			IncidentId: incidentId,
		},
	}
}

var incidentEventsTags = []string{"Incident Events"}

// ops

var ListIncidentEvents = huma.Operation{
	OperationID: "list-incident-events",
	Method:      http.MethodGet,
	Path:        "/incidents/{id}/events",
	Summary:     "List Events for Incident",
	Tags:        append(incidentsTags, incidentEventsTags...),
	Errors:      errorCodes(),
}

type ListIncidentEventsRequest GetIdRequest
type ListIncidentEventsResponse PaginatedResponse[IncidentEvent]

var CreateIncidentEvent = huma.Operation{
	OperationID: "create-incident-event",
	Method:      http.MethodPost,
	Path:        "/incidents/{id}/events",
	Summary:     "Create an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type CreateIncidentEventAttributes struct {
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	StartTime time.Time `json:"start_time"`
}
type CreateIncidentEventRequest CreateIdRequest[CreateIncidentEventAttributes]
type CreateIncidentEventResponse ItemResponse[IncidentEvent]

var UpdateIncidentEvent = huma.Operation{
	OperationID: "update-incident-event",
	Method:      http.MethodPatch,
	Path:        "/incident_events/{id}",
	Summary:     "Update an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type UpdateIncidentEventAttributes struct {
	Title     *string    `json:"title"`
	Type      *string    `json:"type"`
	StartTime *time.Time `json:"start_time"`
}
type UpdateIncidentEventRequest UpdateIdRequest[UpdateIncidentEventAttributes]
type UpdateIncidentEventResponse ItemResponse[IncidentEvent]

var ArchiveIncidentEvent = huma.Operation{
	OperationID: "archive-incident-event",
	Method:      http.MethodDelete,
	Path:        "/incident_events/{id}",
	Summary:     "Archive an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type ArchiveIncidentEventRequest ArchiveIdRequest
type ArchiveIncidentEventResponse EmptyResponse
