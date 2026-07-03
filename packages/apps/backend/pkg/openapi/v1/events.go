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
		Kind               string            `json:"kind"`
		OccurredAt         time.Time         `json:"occurredAt"`
		ReceivedAt         time.Time         `json:"receivedAt"`
		Provider           string            `json:"provider"`
		ProviderSource     string            `json:"providerSource"`
		ProviderSubjectRef string            `json:"providerSubjectRef"`
		SubjectKind        string            `json:"subjectKind"`
		Attributes         map[string]any    `json:"attributes"`
		Projections        []EventProjection `json:"projections"`
	}

	EventProjection struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes EventProjectionAttributes `json:"attributes"`
	}

	EventProjectionAttributes struct {
		Projector string                  `json:"projector"`
		Status    string                  `json:"status" enum:"pending,succeeded,failed"`
		StartedAt time.Time               `json:"startedAt"`
		Error     *string                 `json:"error,omitempty"`
		Entities  []EventProjectionEntity `json:"entities"`
	}

	EventProjectionEntity struct {
		EntityId   uuid.UUID `json:"entityId"`
		EntityKind string    `json:"entityKind"`
	}
)

func EventFromEnt(e *ent.NormalizedEvent) Event {
	attr := EventAttributes{
		Kind:               e.Kind.String(),
		OccurredAt:         e.OccurredAt,
		ReceivedAt:         e.ReceivedAt,
		Provider:           e.Provider,
		ProviderSource:     e.ProviderSource,
		ProviderSubjectRef: e.ProviderSubjectRef,
		SubjectKind:        e.SubjectKind,
		Attributes:         e.Attributes,
		Projections:        make([]EventProjection, len(e.Edges.Projections)),
	}

	for i, proj := range e.Edges.Projections {
		attr.Projections[i] = EventProjectionFromEnt(proj)
	}

	return Event{Id: e.ID, Attributes: attr}
}

func EventProjectionFromEnt(p *ent.NormalizedEventProjection) EventProjection {
	attr := EventProjectionAttributes{
		Projector: p.Projector,
		Status:    p.Status.String(),
		StartedAt: p.StartedAt,
		Entities:  make([]EventProjectionEntity, len(p.Edges.ProjectionEntities)),
	}
	if p.Error != "" {
		attr.Error = &p.Error
	}

	for i, e := range p.Edges.ProjectionEntities {
		attr.Entities[i] = EventProjectionEntityFromEnt(e)
	}

	return EventProjection{Id: p.ID, Attributes: attr}
}

func EventProjectionEntityFromEnt(e *ent.NormalizedEventProjectionEntity) EventProjectionEntity {
	return EventProjectionEntity{
		EntityId:   e.DomainEntityID,
		EntityKind: e.DomainEntityKind,
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
	Errors:      ErrorCodes(),
}

type ListEventsRequest struct {
	ListRequest
	From            time.Time `query:"from"`
	To              time.Time `query:"to"`
	WithProjections bool      `query:"withProjections"`
}
type ListEventsResponse ListResponse[Event]

var GetEvent = huma.Operation{
	OperationID: "get-event",
	Method:      http.MethodGet,
	Path:        "/events/{id}",
	Summary:     "Get Event",
	Tags:        EventsTags,
	Errors:      ErrorCodes(),
}

type GetEventRequest struct {
	IdRequest
	WithProjections bool `query:"withProjections"`
}
type GetEventResponse ItemResponse[Event]
