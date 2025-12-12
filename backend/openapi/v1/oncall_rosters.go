package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type OncallRostersHandler interface {
	GetUserOncallInformation(context.Context, *GetUserOncallInformationRequest) (*GetUserOncallInformationResponse, error)

	ListOncallRosters(context.Context, *ListOncallRostersRequest) (*ListOncallRostersResponse, error)
	GetOncallRoster(context.Context, *GetOncallRosterRequest) (*GetOncallRosterResponse, error)

	AddWatchedOncallRoster(context.Context, *AddWatchedOncallRosterRequest) (*AddWatchedOncallRosterResponse, error)
	ListWatchedOncallRosters(context.Context, *ListWatchedOncallRostersRequest) (*ListWatchedOncallRostersResponse, error)
	RemoveWatchedOncallRoster(context.Context, *RemoveWatchedOncallRosterRequest) (*RemoveWatchedOncallRosterResponse, error)
}

func (o operations) RegisterOncallRosters(api huma.API) {
	huma.Register(api, GetUserOncallInformation, o.GetUserOncallInformation)

	huma.Register(api, ListOncallRosters, o.ListOncallRosters)
	huma.Register(api, GetOncallRoster, o.GetOncallRoster)

	huma.Register(api, AddWatchedOncallRoster, o.AddWatchedOncallRoster)
	huma.Register(api, ListWatchedOncallRosters, o.ListWatchedOncallRosters)
	huma.Register(api, RemoveWatchedOncallRoster, o.RemoveWatchedOncallRoster)
}

type (
	UserOncallInformation struct {
		MemberRosters   []OncallRoster `json:"rosters"`
		WatchingRosters []OncallRoster `json:"watchingRosters"`
		ActiveShifts    []OncallShift  `json:"activeShifts"`
		UpcomingShifts  []OncallShift  `json:"upcomingShifts"`
		PastShifts      []OncallShift  `json:"pastShifts"`
	}

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
		Timezone:    schedule.Timezone,
		Description: "",
	}

	attr.Participants = make([]OncallScheduleParticipant, len(schedule.Edges.Participants))
	for i, p := range schedule.Edges.Participants {
		attr.Participants[i] = OncallScheduleParticipant{
			User:  UserFromEnt(p.Edges.User),
			Index: p.Index,
		}
	}

	return OncallSchedule{
		Id:         schedule.ID,
		Attributes: attr,
	}
}

var oncallRostersTags = []string{"Oncall Rosters"}

// ops

// TODO: remove this

var GetUserOncallInformation = huma.Operation{
	OperationID: "get-user-oncall-information",
	Method:      http.MethodGet,
	Path:        "/oncall/user",
	Summary:     "Get oncall information for a user",
	Tags:        oncallRostersTags,
	Errors:      errorCodes(),
}

type GetUserOncallInformationRequest struct {
	UserId         uuid.UUID `query:"userId" required:"true"`
	ActiveShifts   bool      `query:"activeShifts" required:"false" default:"true"`
	UpcomingShifts int       `query:"upcomingShifts" required:"false" default:"0"`
	PastShifts     int       `query:"pastShifts" required:"false" default:"0"`
}
type GetUserOncallInformationResponse ItemResponse[UserOncallInformation]

var ListOncallRosters = huma.Operation{
	OperationID: "list-oncall-rosters",
	Method:      http.MethodGet,
	Path:        "/oncall/rosters",
	Summary:     "List Oncall Rosters",
	Tags:        oncallRostersTags,
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
	Tags:        oncallRostersTags,
	Errors:      errorCodes(),
}

type GetOncallRosterRequest = GetFlexibleIdRequest
type GetOncallRosterResponse ItemResponse[OncallRoster]

var AddWatchedOncallRoster = huma.Operation{
	OperationID: "add-watched-oncall-roster",
	Method:      http.MethodPost,
	Path:        "/oncall/watched_rosters/{id}",
	Summary:     "Add a watched oncall roster",
	Tags:        oncallRostersTags,
	Errors:      errorCodes(),
}

type AddWatchedOncallRosterRequest PostIdEmptyRequest
type AddWatchedOncallRosterResponse ListResponse[OncallRoster]

var ListWatchedOncallRosters = huma.Operation{
	OperationID: "list-watched-oncall-rosters",
	Method:      http.MethodGet,
	Path:        "/oncall/watched_rosters",
	Summary:     "List watched oncall rosters",
	Tags:        oncallRostersTags,
	Errors:      errorCodes(),
}

type ListWatchedOncallRostersRequest EmptyRequest
type ListWatchedOncallRostersResponse ListResponse[OncallRoster]

var RemoveWatchedOncallRoster = huma.Operation{
	OperationID: "remove-watched-oncall-roster",
	Method:      http.MethodDelete,
	Path:        "/oncall/watched_rosters/{id}",
	Summary:     "Remove a watched oncall roster",
	Tags:        oncallRostersTags,
	Errors:      errorCodes(),
}

type RemoveWatchedOncallRosterRequest DeleteIdRequest
type RemoveWatchedOncallRosterResponse ListResponse[OncallRoster]
