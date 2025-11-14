package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type EventsHandler interface {
	ListEvents(context.Context, *ListEventsRequest) (*ListEventsResponse, error)
	GetEvent(context.Context, *GetEventRequest) (*GetEventResponse, error)
}

func (o operations) RegisterEvents(api huma.API) {
	huma.Register(api, ListEvents, o.ListEvents)
	huma.Register(api, GetEvent, o.GetEvent)
}

type (
	Event struct {
		Id         uuid.UUID       `json:"id"`
		Attributes EventAttributes `json:"attributes"`
	}

	EventAttributes struct {
		Kind        string    `json:"kind"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Timestamp   time.Time `json:"timestamp"`
		RosterId    uuid.UUID `json:"roster_id,omitempty"`
		AlertId     uuid.UUID `json:"alert_id,omitempty"`
	}
)

func EventFromEnt(e *ent.Event) Event {
	attr := EventAttributes{
		Kind:        e.Kind.String(),
		Title:       e.Title,
		Description: e.Description,
		Timestamp:   e.Timestamp,
	}

	//if e.Edges.Annotations != nil {
	//	attr.Annotations = make([]EventAnnotation, len(e.Edges.Annotations))
	//	for i, a := range e.Edges.Annotations {
	//		attr.Annotations[i] = OncallAnnotationFromEnt(a)
	//	}
	//}

	return Event{
		Id:         e.ID,
		Attributes: attr,
	}
}

var EventsTags = []string{"Events"}

// ops

var ListEvents = huma.Operation{
	OperationID: "list-events",
	Method:      http.MethodGet,
	Path:        "/events",
	Summary:     "List Events",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type ListEventsRequest struct {
	ListRequest
	From time.Time `query:"from"`
	To   time.Time `query:"to"`
}
type ListEventsResponse PaginatedResponse[Event]

var GetEvent = huma.Operation{
	OperationID: "get-event",
	Method:      http.MethodGet,
	Path:        "/events/{id}",
	Summary:     "Get Event",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type GetEventRequest struct {
	GetIdRequest
}
type GetEventResponse ItemResponse[Event]
