package openapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/ent"
)

type OncallHandler interface {
	ListOncallRosters(context.Context, *ListOncallRostersRequest) (*ListOncallRostersResponse, error)
	GetOncallRoster(context.Context, *GetOncallRosterRequest) (*GetOncallRosterResponse, error)

	AddWatchedOncallRoster(context.Context, *AddWatchedOncallRosterRequest) (*AddWatchedOncallRosterResponse, error)
	ListWatchedOncallRosters(context.Context, *ListWatchedOncallRostersRequest) (*ListWatchedOncallRostersResponse, error)
	RemoveWatchedOncallRoster(context.Context, *RemoveWatchedOncallRosterRequest) (*RemoveWatchedOncallRosterResponse, error)

	GetUserOncallInformation(context.Context, *GetUserOncallInformationRequest) (*GetUserOncallInformationResponse, error)

	ListOncallShifts(context.Context, *ListOncallShiftsRequest) (*ListOncallShiftsResponse, error)

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
}

func (o operations) RegisterOncall(api huma.API) {
	huma.Register(api, ListOncallRosters, o.ListOncallRosters)
	huma.Register(api, GetOncallRoster, o.GetOncallRoster)

	huma.Register(api, AddWatchedOncallRoster, o.AddWatchedOncallRoster)
	huma.Register(api, ListWatchedOncallRosters, o.ListWatchedOncallRosters)
	huma.Register(api, RemoveWatchedOncallRoster, o.RemoveWatchedOncallRoster)

	huma.Register(api, GetUserOncallInformation, o.GetUserOncallInformation)

	huma.Register(api, ListOncallShifts, o.ListOncallShifts)

	huma.Register(api, GetOncallShift, o.GetOncallShift)
	huma.Register(api, GetPreviousOncallShift, o.GetPreviousOncallShift)
	huma.Register(api, GetNextOncallShift, o.GetNextOncallShift)

	huma.Register(api, CreateOncallShiftHandoverTemplate, o.CreateOncallShiftHandoverTemplate)
	huma.Register(api, GetOncallShiftHandoverTemplate, o.GetOncallShiftHandoverTemplate)
	huma.Register(api, UpdateOncallShiftHandoverTemplate, o.UpdateOncallShiftHandoverTemplate)
	huma.Register(api, ArchiveOncallShiftHandoverTemplate, o.ArchiveOncallShiftHandoverTemplate)

	huma.Register(api, GetOncallShiftHandover, o.GetOncallShiftHandover)
	huma.Register(api, UpdateOncallShiftHandover, o.UpdateOncallShiftHandover)
	huma.Register(api, SendOncallShiftHandover, o.SendOncallShiftHandover)
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
		Type   string `json:"type" enum:"regular,annotations"`
		Header string `json:"header"`
		List   bool   `json:"list"`
	}

	OncallShiftHandover struct {
		Id         uuid.UUID                     `json:"id"`
		Attributes OncallShiftHandoverAttributes `json:"attributes"`
	}

	OncallShiftHandoverAttributes struct {
		ShiftId      uuid.UUID                    `json:"shiftId"`
		SentAt       time.Time                    `json:"sentAt"`
		Content      []OncallShiftHandoverSection `json:"content"`
		PinnedEvents []OncallEventAnnotation      `json:"pinnedEvents"`
	}

	OncallShiftHandoverSection struct {
		Header      string  `json:"header"`
		Kind        string  `json:"kind" enum:"regular,annotations"`
		JsonContent *string `json:"jsonContent,omitempty"`
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
		// TODO: just return an error
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
		ShiftId: p.ShiftID,
		Content: content,
		SentAt:  p.SentAt,
	}

	return OncallShiftHandover{
		Id:         p.ID,
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

var AddWatchedOncallRoster = huma.Operation{
	OperationID: "add-watched-oncall-roster",
	Method:      http.MethodPost,
	Path:        "/oncall/watched_rosters/{id}",
	Summary:     "Add a watched oncall roster",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type AddWatchedOncallRosterRequest PostIdRequest
type AddWatchedOncallRosterResponse ListResponse[OncallRoster]

var ListWatchedOncallRosters = huma.Operation{
	OperationID: "list-watched-oncall-rosters",
	Method:      http.MethodGet,
	Path:        "/oncall/watched_rosters",
	Summary:     "List watched oncall rosters",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type ListWatchedOncallRostersRequest EmptyRequest
type ListWatchedOncallRostersResponse ListResponse[OncallRoster]

var RemoveWatchedOncallRoster = huma.Operation{
	OperationID: "remove-watched-oncall-roster",
	Method:      http.MethodDelete,
	Path:        "/oncall/watched_rosters/{id}",
	Summary:     "Remove a watched oncall roster",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type RemoveWatchedOncallRosterRequest DeleteIdRequest
type RemoveWatchedOncallRosterResponse ListResponse[OncallRoster]

var GetUserOncallInformation = huma.Operation{
	OperationID: "get-user-oncall-information",
	Method:      http.MethodGet,
	Path:        "/oncall/user",
	Summary:     "Get current user oncall information",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type GetUserOncallInformationRequest struct {
	UserId         uuid.UUID `query:"userId" required:"false" nullable:"false"`
	ActiveShifts   bool      `query:"activeShifts" required:"false" default:"true"`
	UpcomingShifts int       `query:"upcomingShifts" required:"false" default:"1"`
	PastShifts     int       `query:"pastShifts" required:"false" default:"1"`
}
type UserOncallInformation struct {
	MemberRosters   []OncallRoster `json:"rosters"`
	WatchingRosters []OncallRoster `json:"watchingRosters"`
	ActiveShifts    []OncallShift  `json:"activeShifts"`
	UpcomingShifts  []OncallShift  `json:"upcomingShifts"`
	PastShifts      []OncallShift  `json:"pastShifts"`
}
type GetUserOncallInformationResponse ItemResponse[UserOncallInformation]

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
	Summary:     "Update an Oncall Shift Handover",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type UpdateOncallShiftHandoverAttributes struct {
	Content             *[]OncallShiftHandoverSection `json:"content,omitempty"`
	PinnedAnnotationIds *[]uuid.UUID                  `json:"pinnedAnnotationIds,omitempty"`
}
type UpdateOncallShiftHandoverRequest UpdateIdRequest[UpdateOncallShiftHandoverAttributes]
type UpdateOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]

var SendOncallShiftHandover = huma.Operation{
	OperationID: "send-oncall-shift-handover",
	Method:      http.MethodPost,
	Path:        "/oncall/handovers/{id}/send",
	Summary:     "Send a Shift Handover",
	Tags:        oncallTags,
	Errors:      errorCodes(),
}

type SendOncallShiftHandoverAttributes struct {
}
type SendOncallShiftHandoverRequest CreateIdRequest[SendOncallShiftHandoverAttributes]
type SendOncallShiftHandoverResponse ItemResponse[OncallShiftHandover]
