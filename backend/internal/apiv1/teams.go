package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema"
	entteam "github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type teamsHandler struct {
	auth        rez.AuthService
	users       *ent.UserClient
	teams       *ent.TeamClient
	memberships *ent.TeamMembershipClient
}

func newTeamsHandler(auth rez.AuthService, users *ent.UserClient, teams *ent.TeamClient, memberships *ent.TeamMembershipClient) *teamsHandler {
	return &teamsHandler{auth: auth, users: users, teams: teams, memberships: memberships}
}

func (h *teamsHandler) ListTeams(ctx context.Context, request *oapi.ListTeamsRequest) (*oapi.ListTeamsResponse, error) {
	var resp oapi.ListTeamsResponse

	query := h.teams.Query()

	if request.IncludeArchived {
		ctx = schema.IncludeArchived(ctx)
	}

	if len(request.Search) > 0 {
		query = query.Where(entteam.NameContainsFold(request.Search))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(entteam.ByID())

	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, apiError("Failed to query teams", queryErr)
	}

	resp.Body.Data = make([]oapi.Team, len(res))
	for i, team := range res {
		resp.Body.Data[i] = oapi.TeamFromEnt(team)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, apiError("Failed to query teams count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *teamsHandler) CreateTeam(ctx context.Context, request *oapi.CreateTeamRequest) (*oapi.CreateTeamResponse, error) {
	var resp oapi.CreateTeamResponse

	attr := request.Body.Attributes
	query := h.teams.Create().
		SetName(attr.Name)

	team, err := query.Save(ctx)
	if err != nil {
		return nil, apiError("failed to create team", err)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) GetTeam(ctx context.Context, request *oapi.GetTeamRequest) (*oapi.GetTeamResponse, error) {
	var resp oapi.GetTeamResponse

	pred := entteam.ID(request.Id.UUID)
	if request.Id.IsSlug {
		pred = entteam.Slug(request.Id.Slug)
	}
	team, queryErr := h.teams.Query().Where(pred).Only(ctx)
	if queryErr != nil {
		return nil, apiError("failed to get team", queryErr)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) UpdateTeam(ctx context.Context, request *oapi.UpdateTeamRequest) (*oapi.UpdateTeamResponse, error) {
	var resp oapi.UpdateTeamResponse

	attr := request.Body.Attributes
	query := h.teams.UpdateOneID(request.Id).
		SetNillableName(attr.Name.NillableValue())

	team, err := query.Save(ctx)
	if err != nil {
		return nil, apiError("failed to update team", err)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) ArchiveTeam(ctx context.Context, request *oapi.ArchiveTeamRequest) (*oapi.ArchiveTeamResponse, error) {
	var resp oapi.ArchiveTeamResponse

	err := h.teams.DeleteOneID(request.Id).Exec(ctx)
	if err != nil {
		return nil, apiError("failed to archive team", err)
	}

	return &resp, nil
}

func (h *teamsHandler) ListTeamMemberships(ctx context.Context, request *oapi.ListTeamMembershipsRequest) (*oapi.ListTeamMembershipsResponse, error) {
	var resp oapi.ListTeamMembershipsResponse

	query := h.memberships.Query().
		WithTeam().
		WithUser()
	if request.TeamId != uuid.Nil {
		query = query.Where(teammembership.TeamID(request.TeamId))
	}
	if request.UserId != uuid.Nil {
		query = query.Where(teammembership.UserID(request.UserId))
	}

	limitedQuery := query.Clone().
		Limit(request.Limit).
		Offset(request.Offset).
		Order(teammembership.ByID())
	res, queryErr := limitedQuery.All(ctx)
	if queryErr != nil {
		return nil, apiError("failed to query team memberships", queryErr)
	}

	resp.Body.Data = make([]oapi.TeamMembership, len(res))
	for i, m := range res {
		resp.Body.Data[i] = oapi.TeamMembershipFromEnt(m)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, apiError("failed to query team memberships count", countErr)
	}
	resp.Body.Pagination.Total = count

	return &resp, nil
}

func (h *teamsHandler) CreateTeamMembership(ctx context.Context, request *oapi.CreateTeamMembershipRequest) (*oapi.CreateTeamMembershipResponse, error) {
	var resp oapi.CreateTeamMembershipResponse

	attr := request.Body.Attributes
	created, createErr := h.memberships.Create().
		SetTeamID(attr.TeamId).
		SetUserID(attr.UserId).
		SetRole(teammembership.Role(attr.Role)).
		Save(ctx)
	if createErr != nil {
		return nil, apiError("failed to create team membership", createErr)
	}

	membership, queryErr := h.memberships.Query().
		Where(teammembership.ID(created.ID)).
		WithTeam().
		WithUser().
		Only(ctx)
	if queryErr != nil {
		return nil, apiError("failed to query created team membership", queryErr)
	}
	resp.Body.Data = oapi.TeamMembershipFromEnt(membership)
	return &resp, nil
}

func (h *teamsHandler) UpdateTeamMembership(ctx context.Context, request *oapi.UpdateTeamMembershipRequest) (*oapi.UpdateTeamMembershipResponse, error) {
	var resp oapi.UpdateTeamMembershipResponse

	attr := request.Body.Attributes
	query := h.memberships.UpdateOneID(request.Id)
	if attr.Role != nil {
		query = query.SetRole(teammembership.Role(*attr.Role))
	}
	if _, saveErr := query.Save(ctx); saveErr != nil {
		return nil, apiError("failed to update team membership", saveErr)
	}

	membership, queryErr := h.memberships.Query().
		Where(teammembership.ID(request.Id)).
		WithTeam().
		WithUser().
		Only(ctx)
	if queryErr != nil {
		return nil, apiError("failed to query updated team membership", queryErr)
	}
	resp.Body.Data = oapi.TeamMembershipFromEnt(membership)
	return &resp, nil
}

func (h *teamsHandler) ArchiveTeamMembership(ctx context.Context, request *oapi.ArchiveTeamMembershipRequest) (*oapi.ArchiveTeamMembershipResponse, error) {
	var resp oapi.ArchiveTeamMembershipResponse
	
	if delErr := h.memberships.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to archive team membership", delErr)
	}
	return &resp, nil
}
