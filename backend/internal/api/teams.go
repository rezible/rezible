package api

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/schema"
	entteam "github.com/twohundreds/rezible/ent/team"
	oapi "github.com/twohundreds/rezible/openapi"
)

type teamsHandler struct {
	teams *ent.TeamClient
}

func newTeamsHandler(teams *ent.TeamClient) *teamsHandler {
	return &teamsHandler{teams}
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
		return nil, detailError("Failed to query teams", queryErr)
	}

	resp.Body.Data = make([]oapi.Team, len(res))
	for i, team := range res {
		resp.Body.Data[i] = oapi.TeamFromEnt(team)
	}

	count, countErr := query.Count(ctx)
	if countErr != nil {
		return nil, detailError("Failed to query teams count", countErr)
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
		return nil, detailError("failed to create team", err)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) GetTeam(ctx context.Context, request *oapi.GetTeamRequest) (*oapi.GetTeamResponse, error) {
	var resp oapi.GetTeamResponse

	var pred predicate.Team
	if request.Id.IsUUID {
		pred = entteam.ID(request.Id.UUID)
	} else {
		pred = entteam.Slug(request.Id.Slug)
	}
	team, queryErr := h.teams.Query().Where(pred).Only(ctx)
	if queryErr != nil {
		return nil, detailError("failed to get team", queryErr)
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
		return nil, detailError("failed to update team", err)
	}
	resp.Body.Data = oapi.TeamFromEnt(team)

	return &resp, nil
}

func (h *teamsHandler) ArchiveTeam(ctx context.Context, request *oapi.ArchiveTeamRequest) (*oapi.ArchiveTeamResponse, error) {
	var resp oapi.ArchiveTeamResponse

	err := h.teams.DeleteOneID(request.Id).Exec(ctx)
	if err != nil {
		return nil, detailError("failed to archive team", err)
	}

	return &resp, nil
}
