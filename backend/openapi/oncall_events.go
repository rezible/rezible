package openapi

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type OncallEventsHandler interface {
	ListOncallEvents(context.Context, *ListOncallEventsRequest) (*ListOncallEventsResponse, error)

	ListOncallAnnotations(context.Context, *ListOncallAnnotationsRequest) (*ListOncallAnnotationsResponse, error)
	CreateOncallAnnotation(context.Context, *CreateOncallAnnotationRequest) (*CreateOncallAnnotationResponse, error)
	UpdateOncallAnnotation(context.Context, *UpdateOncallAnnotationRequest) (*UpdateOncallAnnotationResponse, error)
	DeleteOncallAnnotation(context.Context, *DeleteOncallAnnotationRequest) (*DeleteOncallAnnotationResponse, error)
}

func (o operations) RegisterOncallEvents(api huma.API) {
	huma.Register(api, ListOncallEvents, o.ListOncallEvents)

	huma.Register(api, ListOncallAnnotations, o.ListOncallAnnotations)
	huma.Register(api, CreateOncallAnnotation, o.CreateOncallAnnotation)
	huma.Register(api, UpdateOncallAnnotation, o.UpdateOncallAnnotation)
	huma.Register(api, DeleteOncallAnnotation, o.DeleteOncallAnnotation)
}

type (
	OncallEvent struct {
		Id         string                `json:"id"`
		Attributes OncallEventAttributes `json:"attributes"`
	}

	OncallEventAttributes struct {
		Timestamp  time.Time         `json:"timestamp"`
		Kind       string            `json:"kind"`
		Title      string            `json:"title"`
		Annotation *OncallAnnotation `json:"annotation"`
	}

	OncallAnnotation struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes OncallAnnotationAttributes `json:"attributes"`
	}

	OncallAnnotationAttributes struct {
		RosterId        uuid.UUID `json:"rosterId"`
		Creator         *User     `json:"creator"`
		Notes           string    `json:"notes"`
		MinutesOccupied int       `json:"minutesOccupied"`
	}
)

func OncallAnnotationFromEnt(e *ent.OncallAnnotation) OncallAnnotation {
	// TODO
	attr := OncallAnnotationAttributes{}

	return OncallAnnotation{
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
type ListOncallEventsResponse PaginatedResponse[OncallEvent]

var ListOncallAnnotations = huma.Operation{
	OperationID: "list-oncall-annotations",
	Method:      http.MethodGet,
	Path:        "/oncall/annotations",
	Summary:     "List Oncall Annotations",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type ListOncallAnnotationsRequest struct {
	ListRequest
	RosterId uuid.UUID `query:"rosterId"`
	ShiftId  uuid.UUID `query:"shiftId"`
}
type ListOncallAnnotationsResponse PaginatedResponse[OncallAnnotation]

var CreateOncallAnnotation = huma.Operation{
	OperationID: "create-oncall-annotation",
	Method:      http.MethodPost,
	Path:        "/oncall/annotations",
	Summary:     "Create an Oncall Annotation",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type CreateOncallAnnotationRequestAttributes struct {
	EventId         string    `json:"eventId"`
	RosterId        uuid.UUID `json:"rosterId"`
	MinutesOccupied int       `json:"minutesOccupied"`
	Notes           string    `json:"notes"`
}
type CreateOncallAnnotationRequest CreateIdRequest[CreateOncallAnnotationRequestAttributes]
type CreateOncallAnnotationResponse ItemResponse[OncallAnnotation]

var UpdateOncallAnnotation = huma.Operation{
	OperationID: "update-oncall-annotation",
	Method:      http.MethodPatch,
	Path:        "/oncall/annotations/{id}",
	Summary:     "Update an Oncall Event Annotation",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type UpdateOncallAnnotationRequestAttributes struct {
	Notes           *string `json:"notes,omitempty"`
	MinutesOccupied *int    `json:"minutesOccupied,omitempty"`
}
type UpdateOncallAnnotationRequest UpdateIdRequest[UpdateOncallAnnotationRequestAttributes]
type UpdateOncallAnnotationResponse ItemResponse[OncallAnnotation]

var DeleteOncallAnnotation = huma.Operation{
	OperationID: "delete-oncall-annotation",
	Method:      http.MethodDelete,
	Path:        "/oncall/annotations/{id}",
	Summary:     "Delete an Oncall Event Annotation",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type DeleteOncallAnnotationRequest DeleteIdRequest
type DeleteOncallAnnotationResponse EmptyResponse
