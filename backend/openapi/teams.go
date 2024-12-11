package openapi

import (
	"context"
	"github.com/rezible/rezible/ent"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type TeamsHandler interface {
	ListTeams(context.Context, *ListTeamsRequest) (*ListTeamsResponse, error)
	CreateTeam(context.Context, *CreateTeamRequest) (*CreateTeamResponse, error)
	GetTeam(context.Context, *GetTeamRequest) (*GetTeamResponse, error)
	UpdateTeam(context.Context, *UpdateTeamRequest) (*UpdateTeamResponse, error)
	ArchiveTeam(context.Context, *ArchiveTeamRequest) (*ArchiveTeamResponse, error)
}

func (o operations) RegisterTeams(api huma.API) {
	huma.Register(api, ListTeams, o.ListTeams)
	huma.Register(api, CreateTeam, o.CreateTeam)
	huma.Register(api, GetTeam, o.GetTeam)
	huma.Register(api, UpdateTeam, o.UpdateTeam)
	huma.Register(api, ArchiveTeam, o.ArchiveTeam)
}

type Team struct {
	Id         uuid.UUID      `json:"id"`
	Attributes TeamAttributes `json:"attributes"`
}

type TeamAttributes struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func TeamFromEnt(team *ent.Team) Team {
	return Team{
		Id: team.ID,
		Attributes: TeamAttributes{
			Slug: team.Slug,
			Name: team.Name,
		},
	}
}

var teamsTags = []string{"Teams"}

// Operations

var ListTeams = huma.Operation{
	OperationID: "list-teams",
	Method:      http.MethodGet,
	Path:        "/teams",
	Summary:     "List Teams",
	Tags:        teamsTags,
	Errors:      errorCodes(),
}

type ListTeamsRequest ListRequest
type ListTeamsResponse PaginatedResponse[Team]

var CreateTeam = huma.Operation{
	OperationID: "create-team",
	Method:      http.MethodPost,
	Path:        "/teams",
	Summary:     "Create a Team",
	Tags:        teamsTags,
	Errors:      errorCodes(),
}

type CreateTeamAttributes struct {
	Name string `json:"name"`
}
type CreateTeamRequest RequestWithBodyAttributes[CreateTeamAttributes]
type CreateTeamResponse ItemResponse[Team]

var GetTeam = huma.Operation{
	OperationID: "get-team",
	Method:      http.MethodGet,
	Path:        "/teams/{id}",
	Summary:     "Get a Team",
	Tags:        teamsTags,
	Errors:      errorCodes(),
}

type GetTeamRequest = GetFlexibleIdRequest
type GetTeamResponse ItemResponse[Team]

var UpdateTeam = huma.Operation{
	OperationID: "update-teams",
	Method:      http.MethodPatch,
	Path:        "/teams/{id}",
	Summary:     "Update a Team",
	Tags:        teamsTags,
	Errors:      errorCodes(),
}

type UpdateTeamAttributes struct {
	Name OmittableNullable[string] `json:"name"`
}
type UpdateTeamRequest UpdateIdRequest[UpdateTeamAttributes]
type UpdateTeamResponse ItemResponse[Team]

var ArchiveTeam = huma.Operation{
	OperationID: "archive-team",
	Method:      http.MethodDelete,
	Path:        "/teams/{id}",
	Summary:     "Archive a Team",
	Tags:        teamsTags,
	Errors:      errorCodes(),
}

type ArchiveTeamRequest ArchiveIdRequest
type ArchiveTeamResponse EmptyResponse
