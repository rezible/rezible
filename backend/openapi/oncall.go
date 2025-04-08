package openapi

import (
	"context"
	"encoding/json"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type OncallHandler interface {
	ListOncallRosters(context.Context, *ListOncallRostersRequest) (*ListOncallRostersResponse, error)
	GetOncallRoster(context.Context, *GetOncallRosterRequest) (*GetOncallRosterResponse, error)

	GetUserOncallDetails(context.Context, *GetUserOncallDetailsRequest) (*GetUserOncallDetailsResponse, error)
	ListOncallShifts(context.Context, *ListOncallShiftsRequest) (*ListOncallShiftsResponse, error)

	ListOncallShiftEvents(context.Context, *ListOncallShiftEventsRequest) (*ListOncallShiftEventsResponse, error)

	GetOncallShift(context.Context, *GetOncallShiftRequest) (*GetOncallShiftResponse, error)
	GetPreviousOncallShift(context.Context, *GetPreviousOncallShiftRequest) (*GetPreviousOncallShiftResponse, error)
	GetNextOncallShift(context.Context, *GetNextOncallShiftRequest) (*GetNextOncallShiftResponse, error)

	CreateOncallShiftHandoverTemplate(context.Context, *CreateOncallShiftHandoverTemplateRequest) (*CreateOncallShiftHandoverTemplateResponse, error)
	GetOncallShiftHandoverTemplate(context.Context, *GetOncallShiftHandoverTemplateRequest) (*GetOncallShiftHandoverTemplateResponse, error)
	UpdateOncallShiftHandoverTemplate(context.Context, *UpdateOncallShiftHandoverTemplateRequest) (*UpdateOncallShiftHandoverTemplateResponse, error)
	ArchiveOncallShiftHandoverTemplate(context.Context, *ArchiveOncallShiftHandoverTemplateRequest) (*ArchiveOncallShiftHandoverTemplateResponse, error)

	GetOncallShiftHandover(context.Context, *GetOncallShiftHandoverRequest) (*GetOncallShiftHandoverResponse, error)
	UpdateOncallShiftHandover(context.Context, *UpdateOncallShiftHandoverRequest) (*UpdateOncallShiftHandoverResponse, error)
	SendOncallShiftHandover(context.Context, *SendOncallShiftHandoverRequest) (*SendOncallShiftHandoverResponse, error)

	ListOncallShiftAnnotations(context.Context, *ListOncallShiftAnnotationsRequest) (*ListOncallShiftAnnotationsResponse, error)
	CreateOncallShiftAnnotation(context.Context, *CreateOncallShiftAnnotationRequest) (*CreateOncallShiftAnnotationResponse, error)
	UpdateOncallShiftAnnotation(context.Context, *UpdateOncallShiftAnnotationRequest) (*UpdateOncallShiftAnnotationResponse, error)
	ArchiveOncallShiftAnnotation(context.Context, *ArchiveOncallShiftAnnotationRequest) (*ArchiveOncallShiftAnnotationResponse, error)
}

func (o operations) RegisterOncall(api huma.API) {
	huma.Register(api, ListOncallRosters, o.ListOncallRosters)
	huma.Register(api, GetOncallRoster, o.GetOncallRoster)

	huma.Register(api, GetUserOncallDetails, o.GetUserOncallDetails)

	huma.Register(api, ListOncallShifts, o.ListOncallShifts)

	huma.Register(api, GetOncallShift, o.GetOncallShift)
	huma.Register(api, GetPreviousOncallShift, o.GetPreviousOncallShift)
	huma.Register(api, GetNextOncallShift, o.GetNextOncallShift)

	huma.Register(api, CreateOncallShiftHandoverTemplate, o.CreateOncallShiftHandoverTemplate)
	huma.Register(api, GetOncallShiftHandoverTemplate, o.GetOncallShiftHandoverTemplate)
	huma.Register(api, UpdateOncallShiftHandoverTemplate, o.UpdateOncallShiftHandoverTemplate)
	huma.Register(api, ArchiveOncallShiftHandoverTemplate, o.ArchiveOncallShiftHandoverTemplate)

	huma.Register(api, GetOncallShiftHandover, o.GetOncallShiftHandover)
	huma.Register(api, SendOncallShiftHandover, o.SendOncallShiftHandover)

	huma.Register(api, ListOncallShiftEvents, o.ListOncallShiftEvents)

	huma.Register(api, ListOncallShiftAnnotations, o.ListOncallShiftAnnotations)
	huma.Register(api, CreateOncallShiftAnnotation, o.CreateOncallShiftAnnotation)
	huma.Register(api, UpdateOncallShiftAnnotation, o.UpdateOncallShiftAnnotation)
	huma.Register(api, ArchiveOncallShiftAnnotation, o.ArchiveOncallShiftAnnotation)
}

type (
	OncallRoster struct {
		Id         uuid.UUID              `json:"id"`
		Attributes OncallRosterAttributes `json:"attributes"`
	}

	OncallRosterAttributes struct {
		Name               string           `json:"name"`
		Slug               string           `json:"slug"`
		Schedules          []OncallSchedule `json:"schedules"`
		HandoverTemplateId uuid.UUID        `json:"handoverTemplateId"`
	}

	OncallSchedule struct {
		Id         uuid.UUID                `json:"id"`
		Attributes OncallScheduleAttributes `json:"attributes"`
	}

	OncallScheduleAttributes struct {
		Roster       OncallRoster                `json:"roster"`
		Description  string                      `json:"description"`
		Timezone     string                      `json:"timezone"`
		Participants []OncallScheduleParticipant `json:"participants"`
	}

	OncallScheduleParticipant struct {
		User  User `json:"user"`
		Index int  `json:"order"`
	}

	OncallShift struct {
		Id         uuid.UUID             `json:"id"`
		Attributes OncallShiftAttributes `json:"attributes"`
	}

	OncallShiftAttributes struct {
		User    User               `json:"user"`
		Roster  OncallRoster       `json:"roster"`
		Role    string             `json:"role"`
		StartAt time.Time          `json:"startAt"`
		EndAt   time.Time          `json:"endAt"`
		Covers  []OncallShiftCover `json:"covers"`
	}

	OncallShiftCover struct {
		Id         uuid.UUID                  `json:"id"`
		Attributes OncallShiftCoverAttributes `json:"attributes"`
	}

	OncallShiftCoverAttributes struct {
		User    User      `json:"user"`
		StartAt time.Time `json:"startAt"`
		EndAt   time.Time `json:"endAt"`
	}

	OncallShiftHandoverTemplate struct {
		Id         uuid.UUID                             `json:"id"`
		Attributes OncallShiftHandoverTemplateAttributes `json:"attributes"`
	}

	OncallShiftHandoverTemplateAttributes struct {
		Sections []OncallShiftHandoverTemplateSection `json:"sections"`
	}

	OncallShiftHandoverTemplateSection struct {
		Type   string `json:"type" enum:"regular,annotations,incidents"`
		Header string `json:"header"`
		List   bool   `json:"list"`
	}

	OncallShiftHandover struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes OncallShiftHandoverAttributes `json:"attributes"`
	}

	OncallShiftHandoverAttributes struct {
		ShiftId   uuid.UUID                    `json:"shiftId"`
		CreatedAt time.Time                    `json:"createdAt"`
		UpdatedAt time.Time                    `json:"updatedAt"`
		Content   []OncallShiftHandoverSection `json:"content"`
		SentAt    time.Time                    `json:"sentAt"`
	}

	OncallShiftHandoverSection struct {
		Header      string  `json:"header"`
		Kind        string  `json:"kind" enum:"regular,annotations,incidents"`
		JsonContent *string `json:"jsonContent,omitempty"`
	}

	OncallShiftAnnotation struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes OncallShiftAnnotationAttributes `json:"attributes"`
	}

	OncallShiftAnnotationAttributes struct {
		ShiftId         uuid.UUID        `json:"shiftId"`
		Pinned          bool             `json:"pinned"`
		Notes           string           `json:"notes"`
		Event           *ent.OncallEvent `json:"event"`
		MinutesOccupied int              `json:"minutesOccupied"`
	}
)

func OncallRosterFromEnt(roster *ent.OncallRoster) OncallRoster {
	attr := OncallRosterAttributes{
		Name: roster.Name,
		Slug: roster.Slug,
	}

	attr.Schedules = make([]OncallSchedule, len(roster.Edges.Schedules))
	for i, schedule := range roster.Edges.Schedules {
		attr.Schedules[i] = OncallScheduleFromEnt(schedule)
	}

	// attr.Services = make([]Service, len(roster.Edges.Services))

	return OncallRoster{
		Id:         roster.ID,
		Attributes: attr,
	}
}

func OncallScheduleFromEnt(schedule *ent.OncallSchedule) OncallSchedule {
	attr := OncallScheduleAttributes{
		Description: "",
		Timezone:    schedule.Timezone,
	}

	attr.Participants = make([]OncallScheduleParticipant, len(schedule.Edges.Participants))
	for i, p := range schedule.Edges.Participants {
		attr.Participants[i] = OncallScheduleParticipantFromEnt(p)
	}

	return OncallSchedule{
		Id:         schedule.ID,
		Attributes: attr,
	}
}

func OncallScheduleParticipantFromEnt(p *ent.OncallScheduleParticipant) OncallScheduleParticipant {
	return OncallScheduleParticipant{
		User:  UserFromEnt(p.Edges.User),
		Index: p.Index,
	}
}

func OncallShiftFromEnt(shift *ent.OncallUserShift) OncallShift {
	attr := OncallShiftAttributes{
		Role:    "primary",
		StartAt: shift.StartAt,
		EndAt:   shift.EndAt,
	}

	if shift.Edges.Roster != nil {
		attr.Roster = OncallRosterFromEnt(shift.Edges.Roster)
	}
	if shift.Edges.User != nil {
		attr.User = UserFromEnt(shift.Edges.User)
	}
	attr.Covers = make([]OncallShiftCover, len(shift.Edges.Covers))
	for i, o := range shift.Edges.Covers {
		attr.Covers[i] = OncallShiftCoverFromEnt(o)
	}

	return OncallShift{
		Id:         shift.ID,
		Attributes: attr,
	}
}

func OncallShiftCoverFromEnt(shift *ent.OncallUserShiftCover) OncallShiftCover {
	attr := OncallShiftCoverAttributes{
		StartAt: shift.StartAt,
		EndAt:   shift.EndAt,
	}
	if shift.Edges.User != nil {
		attr.User = UserFromEnt(shift.Edges.User)
	}
	return OncallShiftCover{
		Id:         shift.ID,
		Attributes: attr,
	}
}

type unmarshalOncallShiftContentSection struct {
	Header      string          `json:"header"`
	Kind        string          `json:"kind" enum:"regular,annotations,incidents"`
	JsonContent json.RawMessage `json:"jsonContent,omitempty"`
}

func OncallShiftHandoverFromEnt(p *ent.OncallUserShiftHandover) OncallShiftHandover {
	var rawContents []unmarshalOncallShiftContentSection
	if jsonErr := json.Unmarshal(p.Contents, &rawContents); jsonErr != nil {
		log.Error().Err(jsonErr).Msg("Error unmarshalling OncallShiftHandover contents")
	}
	content := make([]OncallShiftHandoverSection, len(rawContents))
	for i, rawContent := range rawContents {
		content[i] = OncallShiftHandoverSection{
			Header: rawContent.Header,
			Kind:   rawContent.Kind,
		}
		if rawContent.Kind == "regular" && rawContent.JsonContent != nil {
			str := string(rawContent.JsonContent)
			content[i].JsonContent = &str
		}
	}
	attr := OncallShiftHandoverAttributes{
		ShiftId:   p.ShiftID,
		Content:   content,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		SentAt:    p.SentAt,
	}

	return OncallShiftHandover{
		Id:         p.ID,
		Attributes: attr,
	}
}

func OncallShiftAnnotationFromEnt(e *ent.OncallUserShiftAnnotation) OncallShiftAnnotation {
	attr := OncallShiftAnnotationAttributes{
		ShiftId:         e.ShiftID,
		Pinned:          e.Pinned,
		Notes:           e.Notes,
		MinutesOccupied: e.MinutesOccupied,
	}

	return OncallShiftAnnotation{
		Id:         e.ID,
		Attributes: attr,
	}
}

var oncallTags = []string{"Oncall"}

// ops

var ListOncallRosters = huma.Operation{
	OperationID: "list-oncall-rosters",
	Method:      http.MethodGet,
	Path:        "/oncall/rosters",
	Summary:     "List Oncall Rosters",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ListOncallRostersRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"teamId" required:"false" nullable:"false"`
	UserId uuid.UUID `query:"userId" required:"false" nullable:"false"`
	Pinned bool      `query:"pinned" required:"false" nullable:"false"`
}
type ListOncallRostersResponse PaginatedResponse[OncallRoster]

var GetOncallRoster = huma.Operation{
	OperationID: "get-oncall-roster",
	Method:      http.MethodGet,
	Path:        "/oncall/rosters/{id}",
	Summary:     "Get oncall roster",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetOncallRosterRequest = GetFlexibleIdRequest
type GetOncallRosterResponse ItemResponse[OncallRoster]

var GetUserOncallDetails = huma.Operation{
	OperationID: "get-user-oncall-details",
	Method:      http.MethodGet,
	Path:        "/oncall/user",
	Summary:     "Get user oncall details",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetUserOncallDetailsRequest struct {
	UserId uuid.UUID `query:"userId" required:"false" nullable:"false"`
}
type UserOncallDetails struct {
	Rosters        []OncallRoster `json:"rosters"`
	ActiveShifts   []OncallShift  `json:"activeShifts"`
	UpcomingShifts []OncallShift  `json:"upcomingShifts"`
	PastShifts     []OncallShift  `json:"pastShifts"`
}
type GetUserOncallDetailsResponse ItemResponse[UserOncallDetails]

var ListOncallShifts = huma.Operation{
	OperationID: "list-oncall-shifts",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts",
	Summary:     "List Oncall Shifts",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ListOncallShiftsRequest struct {
	ListRequest
	UserId uuid.UUID `query:"userId" required:"false" nullable:"false"`
	Active bool      `query:"active" required:"false" nullable:"false"`
}
type ListOncallShiftsResponse PaginatedResponse[OncallShift]

var GetOncallShift = huma.Operation{
	OperationID: "get-oncall-shift",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}",
	Summary:     "Get an Oncall Shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetOncallShiftRequest GetIdRequest
type GetOncallShiftResponse ItemResponse[OncallShift]

var GetNextOncallShift = huma.Operation{
	OperationID: "get-next-oncall-shift",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/next",
	Summary:     "Get the next Oncall Shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetNextOncallShiftRequest GetIdRequest
type GetNextOncallShiftResponse ItemResponse[OncallShift]

var GetPreviousOncallShift = huma.Operation{
	OperationID: "get-previous-oncall-shift",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/previous",
	Summary:     "Get the previous Oncall Shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetPreviousOncallShiftRequest GetIdRequest
type GetPreviousOncallShiftResponse ItemResponse[OncallShift]

var CreateOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "create-oncall-handover-template",
	Method:      http.MethodPost,
	Path:        "/oncall/handover_templates",
	Summary:     "Create an Oncall Handover Template",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type CreateOncallShiftHandoverTemplateRequestAttributes struct {
	Sections []OncallShiftHandoverTemplateSection `json:"sections"`
}
type CreateOncallShiftHandoverTemplateRequest RequestWithBodyAttributes[CreateOncallShiftHandoverTemplateRequestAttributes]
type CreateOncallShiftHandoverTemplateResponse ItemResponse[OncallShiftHandoverTemplate]

var GetOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "get-oncall-shift-handover-template",
	Method:      http.MethodGet,
	Path:        "/oncall/handover_templates/{id}",
	Summary:     "Get handover for a shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetOncallShiftHandoverTemplateRequest GetIdRequest
type GetOncallShiftHandoverTemplateResponse ItemResponse[OncallShiftHandoverTemplate]

var UpdateOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "update-oncall-handover-template",
	Method:      http.MethodPatch,
	Path:        "/oncall/handover_templates/{id}",
	Summary:     "Update an Oncall Handover Template",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type UpdateOncallShiftHandoverTemplateRequestAttributes struct {
	Sections []OncallShiftHandoverTemplateSection `json:"sections"`
}
type UpdateOncallShiftHandoverTemplateRequest UpdateIdRequest[UpdateOncallShiftHandoverTemplateRequestAttributes]
type UpdateOncallShiftHandoverTemplateResponse ItemResponse[OncallShiftHandoverTemplate]

var ArchiveOncallShiftHandoverTemplate = huma.Operation{
	OperationID: "archive-oncall-handover-template",
	Method:      http.MethodDelete,
	Path:        "/oncall/handover_templates/{id}",
	Summary:     "Archive an Oncall Handover Template",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ArchiveOncallShiftHandoverTemplateRequest ArchiveIdRequest
type ArchiveOncallShiftHandoverTemplateResponse EmptyResponse

var GetOncallShiftHandover = huma.Operation{
	OperationID: "get-oncall-shift-handover",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/handover",
	Summary:     "Get handover for a shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetOncallShiftHandoverRequest GetIdRequest
type GetOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]

var UpdateOncallShiftHandover = huma.Operation{
	OperationID: "update-oncall-shift-handover",
	Method:      http.MethodPatch,
	Path:        "/oncall/handovers/{id}",
	Summary:     "Update a Shift Handover",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type UpdateOncallShiftHandoverAttributes struct {
	Content []OncallShiftHandoverSection `json:"content"`
}
type UpdateOncallShiftHandoverRequest UpdateIdRequest[UpdateOncallShiftHandoverAttributes]
type UpdateOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]

var SendOncallShiftHandover = huma.Operation{
	OperationID: "send-oncall-shift-handover",
	Method:      http.MethodPost,
	Path:        "/oncall/shifts/{id}/handover",
	Summary:     "Send a Shift Handover",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type SendOncallShiftHandoverAttributes struct {
	Content []OncallShiftHandoverSection `json:"content"`
}
type SendOncallShiftHandoverRequest CreateIdRequest[SendOncallShiftHandoverAttributes]
type SendOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]

var ListOncallShiftEvents = huma.Operation{
	OperationID: "list-oncall-shift-events",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/events",
	Summary:     "List Events For an Oncall Shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ListOncallShiftEventsRequest ListIdRequest
type ListOncallShiftEventsResponse PaginatedResponse[ent.OncallEvent]

var ListOncallShiftAnnotations = huma.Operation{
	OperationID: "list-oncall-shift-annotations",
	Method:      http.MethodGet,
	Path:        "/oncall/shifts/{id}/annotations",
	Summary:     "List Annotations For an Oncall Shift",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ListOncallShiftAnnotationsRequest ListIdRequest
type ListOncallShiftAnnotationsResponse PaginatedResponse[OncallShiftAnnotation]

var CreateOncallShiftAnnotation = huma.Operation{
	OperationID: "create-oncall-shift-annotation",
	Method:      http.MethodPost,
	Path:        "/oncall/shifts/{id}/annotations",
	Summary:     "Create an Oncall Shift Annotation",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type CreateOncallShiftAnnotationRequestAttributes struct {
	Event           *ent.OncallEvent `json:"event"`
	MinutesOccupied int              `json:"minutesOccupied"`
	Notes           string           `json:"notes"`
	Pinned          bool             `json:"pinned"`
}
type CreateOncallShiftAnnotationRequest CreateIdRequest[CreateOncallShiftAnnotationRequestAttributes]
type CreateOncallShiftAnnotationResponse ItemResponse[OncallShiftAnnotation]

var UpdateOncallShiftAnnotation = huma.Operation{
	OperationID: "update-oncall-shift-annotation",
	Method:      http.MethodPatch,
	Path:        "/oncall/annotations/{id}",
	Summary:     "Update an Oncall Shift Annotation",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type UpdateOncallShiftAnnotationRequestAttributes struct {
	Pinned          *bool   `json:"pinned,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	MinutesOccupied *int    `json:"minutesOccupied,omitempty"`
}
type UpdateOncallShiftAnnotationRequest UpdateIdRequest[UpdateOncallShiftAnnotationRequestAttributes]
type UpdateOncallShiftAnnotationResponse ItemResponse[OncallShiftAnnotation]

var ArchiveOncallShiftAnnotation = huma.Operation{
	OperationID: "archive-oncall-shift-annotation",
	Method:      http.MethodDelete,
	Path:        "/oncall/annotations/{id}",
	Summary:     "Archive an Oncall Shift Annotation",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ArchiveOncallShiftAnnotationRequest ArchiveIdRequest
type ArchiveOncallShiftAnnotationResponse EmptyResponse
