package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type OncallEventsHandler interface {
	ListOncallEvents(context.Context, *ListOncallEventsRequest) (*ListOncallEventsResponse, error)

	ListOncallEventAnnotations(context.Context, *ListOncallEventAnnotationsRequest) (*ListOncallEventAnnotationsResponse, error)
	CreateOncallEventAnnotation(context.Context, *CreateOncallEventAnnotationRequest) (*CreateOncallEventAnnotationResponse, error)
	UpdateOncallEventAnnotation(context.Context, *UpdateOncallEventAnnotationRequest) (*UpdateOncallEventAnnotationResponse, error)
	DeleteOncallEventAnnotation(context.Context, *DeleteOncallEventAnnotationRequest) (*DeleteOncallEventAnnotationResponse, error)
}

func (o operations) RegisterOncallEvents(api huma.API) {
	huma.Register(api, ListOncallEvents, o.ListOncallEvents)

	huma.Register(api, ListOncallEventAnnotations, o.ListOncallEventAnnotations)
	huma.Register(api, CreateOncallEventAnnotation, o.CreateOncallEventAnnotation)
	huma.Register(api, UpdateOncallEventAnnotation, o.UpdateOncallEventAnnotation)
	huma.Register(api, DeleteOncallEventAnnotation, o.DeleteOncallEventAnnotation)
}

type (
	OncallEventAnnotation struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes OncallEventAnnotationAttributes `json:"attributes"`
	}

	OncallEventAnnotationAttributes struct {
		ShiftId         uuid.UUID        `json:"shiftId"`
		Pinned          bool             `json:"pinned"`
		Notes           string           `json:"notes"`
		Event           *rez.OncallEvent `json:"event"`
		MinutesOccupied int              `json:"minutesOccupied"`
	}
)

func OncallEventAnnotationFromEnt(e *ent.OncallEventAnnotation) OncallEventAnnotation {
	attr := OncallEventAnnotationAttributes{
		Pinned:          e.Pinned,
		Notes:           e.Notes,
		MinutesOccupied: e.MinutesOccupied,
	}

	return OncallEventAnnotation{
		Id:         e.ID,
		Attributes: attr,
	}
}

var oncallEventsTags = []string{"Oncall Events"}

// ops

var ListOncallEvents = huma.Operation{
	OperationID: "list-oncall-events",
	Method:      http.MethodGet,
	Path:        "/oncall/events",
	Summary:     "List Oncall Events",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type ListOncallEventsRequest struct {
	ListRequest
	ShiftId   uuid.UUID `query:"shiftId"`
	RosterIds []string  `query:"rosterIds"`
}
type ListOncallEventsResponse PaginatedResponse[rez.OncallEvent]

var ListOncallEventAnnotations = huma.Operation{
	OperationID: "list-oncall-event-annotations",
	Method:      http.MethodGet,
	Path:        "/oncall/event_annotations",
	Summary:     "List Event Annotations",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type ListOncallEventAnnotationsRequest ListIdRequest
type ListOncallEventAnnotationsResponse PaginatedResponse[OncallEventAnnotation]

var CreateOncallEventAnnotation = huma.Operation{
	OperationID: "create-oncall-event-annotation",
	Method:      http.MethodPost,
	Path:        "/oncall/event_annotations",
	Summary:     "Create an Oncall Event Annotation",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type CreateOncallEventAnnotationRequestAttributes struct {
	EventID         string `json:"eventId"`
	MinutesOccupied int    `json:"minutesOccupied"`
	Notes           string `json:"notes"`
	Pinned          bool   `json:"pinned"`
}
type CreateOncallEventAnnotationRequest CreateIdRequest[CreateOncallEventAnnotationRequestAttributes]
type CreateOncallEventAnnotationResponse ItemResponse[OncallEventAnnotation]

var UpdateOncallEventAnnotation = huma.Operation{
	OperationID: "update-oncall-event-annotation",
	Method:      http.MethodPatch,
	Path:        "/oncall/event_annotations/{id}",
	Summary:     "Update an Oncall Event Annotation",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type UpdateOncallEventAnnotationRequestAttributes struct {
	Pinned          *bool   `json:"pinned,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	MinutesOccupied *int    `json:"minutesOccupied,omitempty"`
}
type UpdateOncallEventAnnotationRequest UpdateIdRequest[UpdateOncallEventAnnotationRequestAttributes]
type UpdateOncallEventAnnotationResponse ItemResponse[OncallEventAnnotation]

var DeleteOncallEventAnnotation = huma.Operation{
	OperationID: "delete-oncall-event-annotation",
	Method:      http.MethodDelete,
	Path:        "/oncall/event_annotations/{id}",
	Summary:     "Delete an Oncall Event Annotation",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type DeleteOncallEventAnnotationRequest DeleteIdRequest
type DeleteOncallEventAnnotationResponse EmptyResponse
