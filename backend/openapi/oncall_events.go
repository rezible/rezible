package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"net/http"
	"time"

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
		Id         uuid.UUID             `json:"id"`
		Attributes OncallEventAttributes `json:"attributes"`
	}

	OncallEventAttributes struct {
		Kind        string    `json:"kind"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Timestamp   time.Time `json:"timestamp"`
		// TODO: don't include annotations with the event
		Annotations []OncallAnnotation `json:"annotations"`
	}

	OncallAnnotation struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes OncallAnnotationAttributes `json:"attributes"`
	}

	OncallAnnotationAttributes struct {
		Event           OncallEvent                    `json:"event"`
		Roster          OncallRoster                   `json:"roster"`
		Creator         User                           `json:"creator"`
		Notes           string                         `json:"notes"`
		Tags            []string                       `json:"tags"`
		MinutesOccupied int                            `json:"minutesOccupied"`
		AlertFeedback   *OncallAnnotationAlertFeedback `json:"alertFeedback,omitempty"`
	}

	OncallAnnotationAlertFeedback struct {
		Accurate               string `json:"accurate" enum:"yes,no,unknown"`
		Actionable             bool   `json:"actionable"`
		DocumentationAvailable string `json:"documentationAvailable" enum:"yes,needs_update,no"`
	}
)

func OncallEventFromEnt(e *ent.OncallEvent) OncallEvent {
	attr := OncallEventAttributes{
		Kind:        e.Kind,
		Title:       e.Title,
		Description: e.Description,
		Timestamp:   e.Timestamp,
	}

	attr.Annotations = make([]OncallAnnotation, len(e.Edges.Annotations))
	for i, a := range e.Edges.Annotations {
		attr.Annotations[i] = OncallAnnotationFromEnt(a)
	}

	return OncallEvent{
		Id:         e.ID,
		Attributes: attr,
	}
}

func OncallAnnotationFromEnt(e *ent.OncallAnnotation) OncallAnnotation {
	attr := OncallAnnotationAttributes{
		Notes:           e.Notes,
		Tags:            nil,
		MinutesOccupied: e.MinutesOccupied,
	}

	if e.Edges.Roster != nil {
		attr.Roster = OncallRosterFromEnt(e.Edges.Roster)
	}

	if e.Edges.Creator != nil {
		attr.Creator = UserFromEnt(e.Edges.Creator)
	}

	if e.Edges.AlertFeedback != nil {
		attr.AlertFeedback = &OncallAnnotationAlertFeedback{
			Accurate:               e.Edges.AlertFeedback.Accurate.String(),
			Actionable:             e.Edges.AlertFeedback.Actionable,
			DocumentationAvailable: e.Edges.AlertFeedback.DocumentationAvailable.String(),
		}
	}

	if e.Edges.Event != nil {
		attr.Event = OncallEventFromEnt(e.Edges.Event)
	}

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
	Annotated bool      `query:"annotated"`
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
	EventId         uuid.UUID                      `json:"eventId"`
	RosterId        uuid.UUID                      `json:"rosterId"`
	Notes           string                         `json:"notes"`
	MinutesOccupied int                            `json:"minutesOccupied"`
	Tags            []string                       `json:"tags"`
	AlertFeedback   *OncallAnnotationAlertFeedback `json:"alertFeedback,omitempty"`
}
type CreateOncallAnnotationRequest RequestWithBodyAttributes[CreateOncallAnnotationRequestAttributes]
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
	Notes           *string                        `json:"notes,omitempty"`
	MinutesOccupied *int                           `json:"minutesOccupied,omitempty"`
	Tags            *[]string                      `json:"tags,omitempty"`
	AlertFeedback   *OncallAnnotationAlertFeedback `json:"alertFeedback,omitempty"`
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
