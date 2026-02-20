package v1

import (
	"context"
	"net/http"

	"github.com/rezible/rezible/ent"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type TeamsHandler interface {
	ListTeams(context.Context, *ListTeamsRequest) (*ListTeamsResponse, error)
	CreateTeam(context.Context, *CreateTeamRequest) (*CreateTeamResponse, error)
	GetTeam(context.Context, *GetTeamRequest) (*GetTeamResponse, error)
	UpdateTeam(context.Context, *UpdateTeamRequest) (*UpdateTeamResponse, error)
	ArchiveTeam(context.Context, *ArchiveTeamRequest) (*ArchiveTeamResponse, error)

	ListTeamMemberships(context.Context, *ListTeamMembershipsRequest) (*ListTeamMembershipsResponse, error)
	CreateTeamMembership(context.Context, *CreateTeamMembershipRequest) (*CreateTeamMembershipResponse, error)
	UpdateTeamMembership(context.Context, *UpdateTeamMembershipRequest) (*UpdateTeamMembershipResponse, error)
	ArchiveTeamMembership(context.Context, *ArchiveTeamMembershipRequest) (*ArchiveTeamMembershipResponse, error)
}

func (o operations) RegisterTeams(api huma.API) {
	huma.Register(api, ListTeams, o.ListTeams)
	huma.Register(api, CreateTeam, o.CreateTeam)
	huma.Register(api, GetTeam, o.GetTeam)
	huma.Register(api, UpdateTeam, o.UpdateTeam)
	huma.Register(api, ArchiveTeam, o.ArchiveTeam)

	huma.Register(api, ListTeamMemberships, o.ListTeamMemberships)
	huma.Register(api, CreateTeamMembership, o.CreateTeamMembership)
	huma.Register(api, UpdateTeamMembership, o.UpdateTeamMembership)
	huma.Register(api, ArchiveTeamMembership, o.ArchiveTeamMembership)
}

type (
	Team struct {
		Id         uuid.UUID      `json:"id"`
		Attributes TeamAttributes `json:"attributes"`
	}

	TeamAttributes struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}

	TeamMembership struct {
		Id         uuid.UUID                `json:"id"`
		Attributes TeamMembershipAttributes `json:"attributes"`
	}

	TeamMembershipAttributes struct {
		Role   string    `json:"role" enum:"admin,member"`
		TeamId uuid.UUID `json:"teamId"`
		UserId uuid.UUID `json:"userId"`
		Team   *Team     `json:"team,omitempty"`
		User   *User     `json:"user,omitempty"`
	}
)

func TeamFromEnt(team *ent.Team) Team {
	return Team{
		Id: team.ID,
		Attributes: TeamAttributes{
			Slug: team.Slug,
			Name: team.Name,
		},
	}
}

func TeamMembershipFromEnt(m *ent.TeamMembership) TeamMembership {
	attr := TeamMembershipAttributes{
		Role:   m.Role.String(),
		TeamId: m.TeamID,
		UserId: m.UserID,
	}
	if team, teamErr := m.Edges.TeamOrErr(); teamErr == nil {
		teamModel := TeamFromEnt(team)
		attr.Team = &teamModel
	}
	if user, userErr := m.Edges.UserOrErr(); userErr == nil {
		userModel := UserFromEnt(user)
		attr.User = &userModel
	}
	return TeamMembership{
		Id:         m.ID,
		Attributes: attr,
	}
}

var teamsTags = []string{"Teams"}
var teamMembershipsTags = []string{"Team Memberships"}

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

var ListTeamMemberships = huma.Operation{
	OperationID: "list-team-memberships",
	Method:      http.MethodGet,
	Path:        "/team_memberships",
	Summary:     "List Team Memberships",
	Tags:        teamMembershipsTags,
	Errors:      errorCodes(),
}

type ListTeamMembershipsRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"teamId" required:"false"`
	UserId uuid.UUID `query:"userId" required:"false"`
}
type ListTeamMembershipsResponse PaginatedResponse[TeamMembership]

var CreateTeamMembership = huma.Operation{
	OperationID: "create-team-membership",
	Method:      http.MethodPost,
	Path:        "/team_memberships",
	Summary:     "Create Team Membership",
	Tags:        teamMembershipsTags,
	Errors:      errorCodes(),
}

type CreateTeamMembershipAttributes struct {
	Role   string    `json:"role" enum:"admin,member"`
	TeamId uuid.UUID `json:"teamId"`
	UserId uuid.UUID `json:"userId"`
}
type CreateTeamMembershipRequest RequestWithBodyAttributes[CreateTeamMembershipAttributes]
type CreateTeamMembershipResponse ItemResponse[TeamMembership]

var UpdateTeamMembership = huma.Operation{
	OperationID: "update-team-membership",
	Method:      http.MethodPatch,
	Path:        "/team_memberships/{id}",
	Summary:     "Update Team Membership",
	Tags:        teamMembershipsTags,
	Errors:      errorCodes(),
}

type UpdateTeamMembershipAttributes struct {
	Role *string `json:"role,omitempty" enum:"admin,member"`
}
type UpdateTeamMembershipRequest UpdateIdRequest[UpdateTeamMembershipAttributes]
type UpdateTeamMembershipResponse ItemResponse[TeamMembership]

var ArchiveTeamMembership = huma.Operation{
	OperationID: "archive-team-membership",
	Method:      http.MethodDelete,
	Path:        "/team_memberships/{id}",
	Summary:     "Archive Team Membership",
	Tags:        teamMembershipsTags,
	Errors:      errorCodes(),
}

type ArchiveTeamMembershipRequest ArchiveIdRequest
type ArchiveTeamMembershipResponse EmptyResponse
