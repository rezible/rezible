package v1

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
		Event           Expandable[EventAttributes] `json:"event"`
		Creator         Expandable[UserAttributes]  `json:"creator"`
		Notes           string                      `json:"notes"`
		Tags            []string                    `json:"tags"`
		MinutesOccupied int                         `json:"minutesOccupied"`
	}

	ExpandAnnotationFields struct {
		Creator bool `json:"creator"`
		Event   bool `json:"event"`
	}
)

func EventAnnotationFromEnt(an *ent.EventAnnotation) EventAnnotation {
	attr := EventAnnotationAttributes{
		Notes:           an.Notes,
		Tags:            nil,
		MinutesOccupied: an.MinutesOccupied,
		Creator:         Expandable[UserAttributes]{Id: an.CreatorID},
		Event:           Expandable[EventAttributes]{Id: an.EventID},
	}

	if an.Edges.Creator != nil {
		usr := UserFromEnt(an.Edges.Creator)
		attr.Creator.Attributes = &usr.Attributes
	}

	//if fb := an.Edges.AlertFeedback; fb != nil {
	//	attr.AlertFeedback = &AlertFeedbackInstance{
	//		Accurate:                 fb.Accurate.String(),
	//		Actionable:               fb.Actionable,
	//		DocumentationAvailable:   fb.DocumentationAvailable,
	//		DocumentationNeedsUpdate: fb.DocumentationNeedsUpdate,
	//	}
	//}

	if an.Edges.Event != nil {
		ev := EventFromEnt(an.Edges.Event)
		attr.Event.Attributes = &ev.Attributes
	}

	return EventAnnotation{
		Id:         an.ID,
		Attributes: attr,
	}
}

var EventAnnotationsTags = []string{"Event Annotations"}

// ops

var ListEventAnnotations = huma.Operation{
	OperationID: "list-event-annotations",
	Method:      http.MethodGet,
	Path:        "/event_annotations",
	Summary:     "List Event Annotations",
	Tags:        EventAnnotationsTags,
	Errors:      errorCodes(),
}

type ListEventAnnotationsRequest struct {
	ListRequest
	From       time.Time `query:"from"`
	To         time.Time `query:"to"`
	UserIds    uuid.UUID `query:"userIds"`
	ShiftIds   uuid.UUID `query:"shiftIds"`
	WithEvents bool      `query:"withEvents"`
}
type ListEventAnnotationsResponse PaginatedResponse[EventAnnotation]

var CreateEventAnnotation = huma.Operation{
	OperationID: "create-event-annotation",
	Method:      http.MethodPost,
	Path:        "/event_annotations",
	Summary:     "Create an Event Annotation",
	Tags:        EventAnnotationsTags,
	Errors:      errorCodes(),
}

type CreateEventAnnotationRequestAttributes struct {
	EventId         uuid.UUID `json:"eventId"`
	Notes           string    `json:"notes"`
	MinutesOccupied int       `json:"minutesOccupied"`
	Tags            []string  `json:"tags"`
}
type CreateEventAnnotationRequest RequestWithBodyAttributes[CreateEventAnnotationRequestAttributes]
type CreateEventAnnotationResponse ItemResponse[EventAnnotation]

var UpdateEventAnnotation = huma.Operation{
	OperationID: "update-event-annotation",
	Method:      http.MethodPatch,
	Path:        "/event_annotations/{id}",
	Summary:     "Update an Event Annotation",
	Tags:        EventAnnotationsTags,
	Errors:      errorCodes(),
}

type UpdateEventAnnotationRequestAttributes struct {
	Notes           *string   `json:"notes,omitempty"`
	MinutesOccupied *int      `json:"minutesOccupied,omitempty"`
	Tags            *[]string `json:"tags,omitempty"`
}
type UpdateEventAnnotationRequest UpdateIdRequest[UpdateEventAnnotationRequestAttributes]
type UpdateEventAnnotationResponse ItemResponse[EventAnnotation]

var DeleteEventAnnotation = huma.Operation{
	OperationID: "delete-event-annotation",
	Method:      http.MethodDelete,
	Path:        "/event_annotations/{id}",
	Summary:     "Delete an Event Annotation",
	Tags:        EventAnnotationsTags,
	Errors:      errorCodes(),
}

type DeleteEventAnnotationRequest DeleteIdRequest
type DeleteEventAnnotationResponse EmptyResponse
