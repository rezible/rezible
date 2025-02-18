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
	DeleteIncidentEvent(context.Context, *DeleteIncidentEventRequest) (*DeleteIncidentEventResponse, error)
}

func (o operations) RegisterIncidentEvents(api huma.API) {
	huma.Register(api, ListIncidentEvents, o.ListIncidentEvents)
	huma.Register(api, CreateIncidentEvent, o.CreateIncidentEvent)
	huma.Register(api, UpdateIncidentEvent, o.UpdateIncidentEvent)
	huma.Register(api, DeleteIncidentEvent, o.DeleteIncidentEvent)
}

/*

	id: string;
	incidentId: string;
	timestamp: Date | null; // null for "unknown time"
	kind: "observation" | "action" | "decision" | "context";
	title: string;
	description: string;
	createdAt: Date;
	updatedAt: Date;
	createdBy: string;
	sequence: number; // for ordering events with same timestamp
	isDraft: boolean;
*/

type (
	IncidentEvent struct {
		Id         uuid.UUID               `json:"id"`
		Attributes IncidentEventAttributes `json:"attributes"`
	}
	IncidentEventAttributes struct {
		IncidentId uuid.UUID `json:"incidentId"`
		Type       string    `json:"type"`
		Title      string    `json:"title"`
		Timestamp  time.Time `json:"timestamp"`
	}
)

func IncidentEventFromEnt(m *ent.IncidentEvent) IncidentEvent {
	return IncidentEvent{
		Id: m.ID,
		Attributes: IncidentEventAttributes{
			Type: m.Type.String(),
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
	Path:        "/incident_events",
	Summary:     "Create an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type CreateIncidentEventAttributes struct {
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
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
	Timestamp *time.Time `json:"timestamp"`
}
type UpdateIncidentEventRequest UpdateIdRequest[UpdateIncidentEventAttributes]
type UpdateIncidentEventResponse ItemResponse[IncidentEvent]

var DeleteIncidentEvent = huma.Operation{
	OperationID: "delete-incident-event",
	Method:      http.MethodDelete,
	Path:        "/incident_events/{id}",
	Summary:     "Delete an Incident Event",
	Tags:        incidentEventsTags,
	Errors:      errorCodes(),
}

type DeleteIncidentEventRequest DeleteIdRequest
type DeleteIncidentEventResponse EmptyResponse
