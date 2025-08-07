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
	GetOncallEvent(context.Context, *GetOncallEventRequest) (*GetOncallEventResponse, error)

	ListOncallAnnotations(context.Context, *ListOncallAnnotationsRequest) (*ListOncallAnnotationsResponse, error)
	CreateOncallAnnotation(context.Context, *CreateOncallAnnotationRequest) (*CreateOncallAnnotationResponse, error)
	UpdateOncallAnnotation(context.Context, *UpdateOncallAnnotationRequest) (*UpdateOncallAnnotationResponse, error)
	DeleteOncallAnnotation(context.Context, *DeleteOncallAnnotationRequest) (*DeleteOncallAnnotationResponse, error)
}

func (o operations) RegisterOncallEvents(api huma.API) {
	huma.Register(api, ListOncallEvents, o.ListOncallEvents)
	huma.Register(api, GetOncallEvent, o.GetOncallEvent)

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
		Kind        string             `json:"kind"`
		Title       string             `json:"title"`
		Description string             `json:"description"`
		Timestamp   time.Time          `json:"timestamp"`
		RosterId    uuid.UUID          `json:"roster_id,omitempty"`
		AlertId     uuid.UUID          `json:"alert_id,omitempty"`
		Annotations []OncallAnnotation `json:"annotations,omitempty"`
	}

	OncallAnnotation struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes OncallAnnotationAttributes `json:"attributes"`
	}

	OncallAnnotationAttributes struct {
		Event           Expandable[OncallEventAttributes]  `json:"event"`
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

func OncallEventFromEnt(e *ent.OncallEvent) OncallEvent {
	attr := OncallEventAttributes{
		Kind:        e.Kind.String(),
		Title:       e.Title,
		Description: e.Description,
		Timestamp:   e.Timestamp,
		RosterId:    e.RosterID,
		AlertId:     e.AlertID,
	}

	if e.Edges.Annotations != nil {
		attr.Annotations = make([]OncallAnnotation, len(e.Edges.Annotations))
		for i, a := range e.Edges.Annotations {
			attr.Annotations[i] = OncallAnnotationFromEnt(a)
		}
	}

	return OncallEvent{
		Id:         e.ID,
		Attributes: attr,
	}
}

func OncallAnnotationFromEnt(an *ent.OncallAnnotation) OncallAnnotation {
	attr := OncallAnnotationAttributes{
		Notes:           an.Notes,
		Tags:            nil,
		MinutesOccupied: an.MinutesOccupied,
		Roster:          Expandable[OncallRosterAttributes]{Id: an.RosterID},
		Creator:         Expandable[UserAttributes]{Id: an.CreatorID},
		Event:           Expandable[OncallEventAttributes]{Id: an.EventID},
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
		ev := OncallEventFromEnt(an.Edges.Event)
		attr.Event.Attributes = &ev.Attributes
	}

	return OncallAnnotation{
		Id:         an.ID,
		Attributes: attr,
	}
}

var oncallEventsTags = []string{"Oncall Events"}

// ops

var GetOncallEvent = huma.Operation{
	OperationID: "get-oncall-event",
	Method:      http.MethodGet,
	Path:        "/oncall/events/{id}",
	Summary:     "Get Oncall Event",
	Tags:        oncallEventsTags,
	Errors:      errorCodes(),
}

type GetOncallEventRequest struct {
	GetIdRequest
}
type GetOncallEventResponse ItemResponse[OncallEvent]

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
	From               time.Time `query:"from"`
	To                 time.Time `query:"to"`
	ShiftId            uuid.UUID `query:"shiftId"`
	AlertId            uuid.UUID `query:"alertId"`
	RosterId           uuid.UUID `query:"rosterId"`
	AnnotationRosterId uuid.UUID `query:"annotationRosterId"`
	WithAnnotations    bool      `query:"withAnnotations"`
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
	From       time.Time `query:"from"`
	To         time.Time `query:"to"`
	RosterId   uuid.UUID `query:"rosterId"`
	ShiftId    uuid.UUID `query:"shiftId"`
	WithEvents bool      `query:"withEvents"`
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
	EventId         uuid.UUID              `json:"eventId"`
	RosterId        uuid.UUID              `json:"rosterId"`
	Notes           string                 `json:"notes"`
	MinutesOccupied int                    `json:"minutesOccupied"`
	Tags            []string               `json:"tags"`
	AlertFeedback   *AlertFeedbackInstance `json:"alertFeedback,omitempty"`
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
	Notes           *string                `json:"notes,omitempty"`
	MinutesOccupied *int                   `json:"minutesOccupied,omitempty"`
	Tags            *[]string              `json:"tags,omitempty"`
	AlertFeedback   *AlertFeedbackInstance `json:"alertFeedback,omitempty"`
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
