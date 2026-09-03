package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"

	entteam "github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/teammembership"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type teamsHandler struct {
	db rez.Database
}

func newTeamsHandler(db rez.Database) *teamsHandler {
	return &teamsHandler{db: db}
}

func (h *teamsHandler) ListTeams(ctx context.Context, request *oapi.ListTeamsRequest) (*oapi.ListTeamsResponse, error) {
	var resp oapi.ListTeamsResponse

	query := h.db.Client(ctx).Team.Query()

	if len(request.Search) > 0 {
		query = query.Where(entteam.NameContainsFold(request.Search))
	}
	query.Order(entteam.ByID())
	params := request.ListParams()
	params.Search = request.Search
	params.IncludeArchived = request.IncludeArchived
	res, queryErr := ent.DoListQuery[ent.Team, *ent.TeamQuery](ctx, query, params)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "Failed to query teams", queryErr)
	}

	resp.Body = oapi.ConvertPaginatedResultBody(res, oapi.TeamFromEnt)

	return &resp, nil
}

func (h *teamsHandler) CreateTeam(ctx context.Context, request *oapi.CreateTeamRequest) (*oapi.CreateTeamResponse, error) {
	var resp oapi.CreateTeamResponse

	attr := request.Body.Attributes
	query := h.db.Client(ctx).Team.Create().
		SetName(attr.Name)

	team, err := query.Save(ctx)
	if err != nil {
		return nil, oapi.Error(ctx, "failed to create team", err)
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
	team, queryErr := h.db.Client(ctx).Team.Query().Where(pred).Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to get team", queryErr)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) UpdateTeam(ctx context.Context, request *oapi.UpdateTeamRequest) (*oapi.UpdateTeamResponse, error) {
	var resp oapi.UpdateTeamResponse

	attr := request.Body.Attributes
	query := h.db.Client(ctx).Team.UpdateOneID(request.Id).
		SetNillableName(attr.Name.NillableValue())

	team, err := query.Save(ctx)
	if err != nil {
		return nil, oapi.Error(ctx, "failed to update team", err)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) ArchiveTeam(ctx context.Context, request *oapi.ArchiveTeamRequest) (*oapi.ArchiveTeamResponse, error) {
	var resp oapi.ArchiveTeamResponse

	err := h.db.Client(ctx).Team.DeleteOneID(request.Id).Exec(ctx)
	if err != nil {
		return nil, oapi.Error(ctx, "failed to archive team", err)
	}

	return &resp, nil
}

func (h *teamsHandler) ListTeamMemberships(ctx context.Context, request *oapi.ListTeamMembershipsRequest) (*oapi.ListTeamMembershipsResponse, error) {
	var resp oapi.ListTeamMembershipsResponse

	query := h.db.Client(ctx).TeamMembership.Query().
		WithTeam().
		WithUser()
	if request.TeamId != uuid.Nil {
		query = query.Where(teammembership.TeamID(request.TeamId))
	}
	if request.UserId != uuid.Nil {
		query = query.Where(teammembership.UserID(request.UserId))
	}

	query.Order(teammembership.ByID())
	res, queryErr := ent.DoListQuery[ent.TeamMembership, *ent.TeamMembershipQuery](ctx, query, request.ListParams())
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to query team memberships", queryErr)
	}

	resp.Body = oapi.ConvertPaginatedResultBody(res, oapi.TeamMembershipFromEnt)

	return &resp, nil
}

func (h *teamsHandler) CreateTeamMembership(ctx context.Context, request *oapi.CreateTeamMembershipRequest) (*oapi.CreateTeamMembershipResponse, error) {
	var resp oapi.CreateTeamMembershipResponse

	attr := request.Body.Attributes
	created, createErr := h.db.Client(ctx).TeamMembership.Create().
		SetTeamID(attr.TeamId).
		SetUserID(attr.UserId).
		SetRole(teammembership.Role(attr.Role)).
		Save(ctx)
	if createErr != nil {
		return nil, oapi.Error(ctx, "failed to create team membership", createErr)
	}

	membership, queryErr := h.db.Client(ctx).TeamMembership.Query().
		Where(teammembership.ID(created.ID)).
		WithTeam().
		WithUser().
		Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to query created team membership", queryErr)
	}
	resp.Body.Data = oapi.TeamMembershipFromEnt(membership)
	return &resp, nil
}

func (h *teamsHandler) UpdateTeamMembership(ctx context.Context, request *oapi.UpdateTeamMembershipRequest) (*oapi.UpdateTeamMembershipResponse, error) {
	var resp oapi.UpdateTeamMembershipResponse

	attr := request.Body.Attributes
	query := h.db.Client(ctx).TeamMembership.UpdateOneID(request.Id)
	if attr.Role != nil {
		query = query.SetRole(teammembership.Role(*attr.Role))
	}
	if _, saveErr := query.Save(ctx); saveErr != nil {
		return nil, oapi.Error(ctx, "failed to update team membership", saveErr)
	}

	membership, queryErr := h.db.Client(ctx).TeamMembership.Query().
		Where(teammembership.ID(request.Id)).
		WithTeam().
		WithUser().
		Only(ctx)
	if queryErr != nil {
		return nil, oapi.Error(ctx, "failed to query updated team membership", queryErr)
	}
	resp.Body.Data = oapi.TeamMembershipFromEnt(membership)
	return &resp, nil
}

func (h *teamsHandler) DeleteTeamMembership(ctx context.Context, request *oapi.DeleteTeamMembershipRequest) (*oapi.DeleteTeamMembershipResponse, error) {
	var resp oapi.DeleteTeamMembershipResponse

	if delErr := h.db.Client(ctx).TeamMembership.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, oapi.Error(ctx, "failed to archive team membership", delErr)
	}
	return &resp, nil
}
