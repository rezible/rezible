package openapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type EventAnnotationsHandler interface {
	ListEventAnnotations(context.Context, *ListEventAnnotationsRequest) (*ListEventAnnotationsResponse, error)
	CreateEventAnnotation(context.Context, *CreateEventAnnotationRequest) (*CreateEventAnnotationResponse, error)
	UpdateEventAnnotation(context.Context, *UpdateEventAnnotationRequest) (*UpdateEventAnnotationResponse, error)
	DeleteEventAnnotation(context.Context, *DeleteEventAnnotationRequest) (*DeleteEventAnnotationResponse, error)
}

func (o operations) RegisterEventAnnotations(api huma.API) {
	huma.Register(api, ListEventAnnotations, o.ListEventAnnotations)
	huma.Register(api, CreateEventAnnotation, o.CreateEventAnnotation)
	huma.Register(api, UpdateEventAnnotation, o.UpdateEventAnnotation)
	huma.Register(api, DeleteEventAnnotation, o.DeleteEventAnnotation)
}

type (
	EventAnnotation struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes EventAnnotationAttributes `json:"attributes"`
	}

	EventAnnotationAttributes struct {
		Event           Expandable[EventAttributes]        `json:"event"`
		Roster          Expandable[OncallRosterAttributes] `json:"roster"`
		Creator         Expandable[UserAttributes]         `json:"creator"`
		Notes           string                             `json:"notes"`
		Tags            []string                           `json:"tags"`
		MinutesOccupied int                                `json:"minutesOccupied"`
		AlertFeedback   *AlertFeedbackInstance             `json:"alertFeedback,omitempty"`
	}

	AlertFeedbackInstance struct {
		Actionable               bool   `json:"actionable"`
		Accurate                 string `json:"accurate" enum:"yes,no,unknown"`
		DocumentationAvailable   bool   `json:"documentationAvailable"`
		DocumentationNeedsUpdate bool   `json:"documentationNeedsUpdate"`
	}

	ExpandAnnotationFields struct {
		Creator       bool `json:"creator"`
		Roster        bool `json:"roster"`
		Event         bool `json:"event"`
		AlertFeedback bool `json:"alertFeedback"`
	}
)

func EventFromEnt(e *ent.Event) Event {
	attr := EventAttributes{
		Kind:        e.Kind.String(),
		Title:       e.Title,
		Description: e.Description,
		Timestamp:   e.Timestamp,
		RosterId:    e.RosterID,
		AlertId:     e.AlertID,
	}

	if e.Edges.Annotations != nil {
		attr.Annotations = make([]EventAnnotation, len(e.Edges.Annotations))
		for i, a := range e.Edges.Annotations {
			attr.Annotations[i] = EventAnnotationFromEnt(a)
		}
	}

	return Event{
		Id:         e.ID,
		Attributes: attr,
	}
}

func EventAnnotationFromEnt(an *ent.EventAnnotation) EventAnnotation {
	attr := EventAnnotationAttributes{
		Notes:           an.Notes,
		Tags:            nil,
		MinutesOccupied: an.MinutesOccupied,
		Roster:          Expandable[OncallRosterAttributes]{Id: an.RosterID},
		Creator:         Expandable[UserAttributes]{Id: an.CreatorID},
		Event:           Expandable[EventAttributes]{Id: an.EventID},
	}

	if an.Edges.Roster != nil {
		roster := OncallRosterFromEnt(an.Edges.Roster)
		attr.Roster.Attributes = &roster.Attributes
	}

	if an.Edges.Creator != nil {
		usr := UserFromEnt(an.Edges.Creator)
		attr.Creator.Attributes = &usr.Attributes
	}

	if fb := an.Edges.AlertFeedback; fb != nil {
		attr.AlertFeedback = &AlertFeedbackInstance{
			Accurate:                 fb.Accurate.String(),
			Actionable:               fb.Actionable,
			DocumentationAvailable:   fb.DocumentationAvailable,
			DocumentationNeedsUpdate: fb.DocumentationNeedsUpdate,
		}
	}

	if an.Edges.Event != nil {
		ev := EventFromEnt(an.Edges.Event)
		attr.Event.Attributes = &ev.Attributes
	}

	return EventAnnotation{
		Id:         an.ID,
		Attributes: attr,
	}
}

var EventsTags = []string{"Oncall Events"}

// ops

var GetEvent = huma.Operation{
	OperationID: "get-oncall-event",
	Method:      http.MethodGet,
	Path:        "/oncall/events/{id}",
	Summary:     "Get Oncall Event",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type GetEventRequest struct {
	GetIdRequest
}
type GetEventResponse ItemResponse[Event]

var ListEvents = huma.Operation{
	OperationID: "list-oncall-events",
	Method:      http.MethodGet,
	Path:        "/oncall/events",
	Summary:     "List Oncall Events",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type ListEventsRequest struct {
	ListRequest
	From               time.Time `query:"from"`
	To                 time.Time `query:"to"`
	ShiftId            uuid.UUID `query:"shiftId"`
	AlertId            uuid.UUID `query:"alertId"`
	RosterId           uuid.UUID `query:"rosterId"`
	AnnotationRosterId uuid.UUID `query:"annotationRosterId"`
	WithAnnotations    bool      `query:"withAnnotations"`
}
type ListEventsResponse PaginatedResponse[Event]

var ListEventAnnotations = huma.Operation{
	OperationID: "list-oncall-annotations",
	Method:      http.MethodGet,
	Path:        "/oncall/annotations",
	Summary:     "List Oncall Annotations",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type ListEventAnnotationsRequest struct {
	ListRequest
	From       time.Time `query:"from"`
	To         time.Time `query:"to"`
	RosterId   uuid.UUID `query:"rosterId"`
	ShiftId    uuid.UUID `query:"shiftId"`
	WithEvents bool      `query:"withEvents"`
}
type ListEventAnnotationsResponse PaginatedResponse[EventAnnotation]

var CreateEventAnnotation = huma.Operation{
	OperationID: "create-oncall-annotation",
	Method:      http.MethodPost,
	Path:        "/oncall/annotations",
	Summary:     "Create an Oncall Annotation",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type CreateEventAnnotationRequestAttributes struct {
	EventId         uuid.UUID              `json:"eventId"`
	RosterId        uuid.UUID              `json:"rosterId"`
	Notes           string                 `json:"notes"`
	MinutesOccupied int                    `json:"minutesOccupied"`
	Tags            []string               `json:"tags"`
	AlertFeedback   *AlertFeedbackInstance `json:"alertFeedback,omitempty"`
}
type CreateEventAnnotationRequest RequestWithBodyAttributes[CreateEventAnnotationRequestAttributes]
type CreateEventAnnotationResponse ItemResponse[EventAnnotation]

var UpdateEventAnnotation = huma.Operation{
	OperationID: "update-oncall-annotation",
	Method:      http.MethodPatch,
	Path:        "/oncall/annotations/{id}",
	Summary:     "Update an Oncall Event Annotation",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type UpdateEventAnnotationRequestAttributes struct {
	Notes           *string                `json:"notes,omitempty"`
	MinutesOccupied *int                   `json:"minutesOccupied,omitempty"`
	Tags            *[]string              `json:"tags,omitempty"`
	AlertFeedback   *AlertFeedbackInstance `json:"alertFeedback,omitempty"`
}
type UpdateEventAnnotationRequest UpdateIdRequest[UpdateEventAnnotationRequestAttributes]
type UpdateEventAnnotationResponse ItemResponse[EventAnnotation]

var DeleteEventAnnotation = huma.Operation{
	OperationID: "delete-oncall-annotation",
	Method:      http.MethodDelete,
	Path:        "/oncall/annotations/{id}",
	Summary:     "Delete an Oncall Event Annotation",
	Tags:        EventsTags,
	Errors:      errorCodes(),
}

type DeleteEventAnnotationRequest DeleteIdRequest
type DeleteEventAnnotationResponse EmptyResponse
