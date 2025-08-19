package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"time"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallRostersHandler struct {
	auth      rez.AuthService
	users     rez.UserService
	incidents rez.IncidentService
	oncall    rez.OncallService
}

func newOncallRostersHandler(auth rez.AuthService, users rez.UserService, inc rez.IncidentService, oncall rez.OncallService) *oncallRostersHandler {
	return &oncallRostersHandler{auth: auth, users: users, incidents: inc, oncall: oncall}
}

func (h *oncallRostersHandler) ListOncallRosters(ctx context.Context, request *oapi.ListOncallRostersRequest) (*oapi.ListOncallRostersResponse, error) {
	var resp oapi.ListOncallRostersResponse

	rosters, rostersErr := h.oncall.ListRosters(ctx, rez.ListOncallRostersParams{
		ListParams: request.ListParams(),
		UserID:     request.UserId,
	})
	if rostersErr != nil {
		return nil, apiError("failed to list rosters", rostersErr)
	}

	resp.Body.Data = make([]oapi.OncallRoster, len(rosters))
	for i, r := range rosters {
		resp.Body.Data[i] = oapi.OncallRosterFromEnt(r)
	}

	return &resp, nil
}

func (h *oncallRostersHandler) GetOncallRoster(ctx context.Context, request *oapi.GetOncallRosterRequest) (*oapi.GetOncallRosterResponse, error) {
	var resp oapi.GetOncallRosterResponse

	var roster *ent.OncallRoster
	var rosterErr error
	if request.Id.IsSlug {
		roster, rosterErr = h.oncall.GetRosterBySlug(ctx, request.Id.Slug)
	} else {
		roster, rosterErr = h.oncall.GetRosterByID(ctx, request.Id.UUID)
	}
	if rosterErr != nil {
		return nil, apiError("failed to get oncall roster", rosterErr)
	}

	schedules, schedulesErr := roster.QuerySchedules().All(ctx)
	if schedulesErr != nil {
		return nil, apiError("failed to query schedules", schedulesErr)
	}
	roster.Edges.Schedules = schedules

	resp.Body.Data = oapi.OncallRosterFromEnt(roster)

	return &resp, nil
}

func (h *oncallRostersHandler) getUserWatchedOncallRosters(ctx context.Context, user *ent.User) ([]oapi.OncallRoster, error) {
	rosters, queryErr := user.QueryWatchedOncallRosters().All(ctx)
	if queryErr != nil {
		return nil, apiError("failed to query watched oncall rosters", queryErr)
	}
	watched := make([]oapi.OncallRoster, len(rosters))
	for i, r := range rosters {
		watched[i] = oapi.OncallRosterFromEnt(r)
	}
	return watched, nil
}

func (h *oncallRostersHandler) AddWatchedOncallRoster(ctx context.Context, request *oapi.AddWatchedOncallRosterRequest) (*oapi.AddWatchedOncallRosterResponse, error) {
	var resp oapi.AddWatchedOncallRosterResponse

	user, userErr := h.users.GetById(ctx, requestUserId(ctx, h.auth))
	if userErr != nil {
		return nil, apiError("failed to get user", userErr)
	}

	if addErr := user.Update().AddWatchedOncallRosterIDs(request.Id).Exec(ctx); addErr != nil {
		return nil, apiError("failed to add watched oncall roster", addErr)
	}

	watched, queryErr := h.getUserWatchedOncallRosters(ctx, user)
	if queryErr != nil {
		return nil, apiError("failed to query watched oncall rosters", queryErr)
	}
	resp.Body.Data = watched

	return &resp, nil
}

func (h *oncallRostersHandler) ListWatchedOncallRosters(ctx context.Context, request *oapi.ListWatchedOncallRostersRequest) (*oapi.ListWatchedOncallRostersResponse, error) {
	var resp oapi.ListWatchedOncallRostersResponse

	user, userErr := h.users.GetById(ctx, requestUserId(ctx, h.auth))
	if userErr != nil {
		return nil, apiError("failed to get user", userErr)
	}
	watched, queryErr := h.getUserWatchedOncallRosters(ctx, user)
	if queryErr != nil {
		return nil, apiError("failed to query watched oncall rosters", queryErr)
	}
	resp.Body.Data = watched

	return &resp, nil
}

func (h *oncallRostersHandler) RemoveWatchedOncallRoster(ctx context.Context, request *oapi.RemoveWatchedOncallRosterRequest) (*oapi.RemoveWatchedOncallRosterResponse, error) {
	var resp oapi.RemoveWatchedOncallRosterResponse

	user, userErr := h.users.GetById(ctx, requestUserId(ctx, h.auth))
	if userErr != nil {
		return nil, apiError("failed to get user", userErr)
	}

	if addErr := user.Update().RemoveWatchedOncallRosterIDs(request.Id).Exec(ctx); addErr != nil {
		return nil, apiError("failed to add watched oncall roster", addErr)
	}

	watched, queryErr := h.getUserWatchedOncallRosters(ctx, user)
	if queryErr != nil {
		return nil, apiError("failed to query watched oncall rosters", queryErr)
	}
	resp.Body.Data = watched

	return &resp, nil
}

func (h *oncallRostersHandler) GetUserOncallInformation(ctx context.Context, request *oapi.GetUserOncallInformationRequest) (*oapi.GetUserOncallInformationResponse, error) {
	var resp oapi.GetUserOncallInformationResponse

	userId := requestUserId(ctx, h.auth)
	if request.UserId != uuid.Nil {
		userId = request.UserId
	}

	user, userErr := h.users.GetById(ctx, userId)
	if userErr != nil {
		return nil, apiError("failed to get user", userErr)
	}

	memberRosters, rostersErr := h.oncall.ListRosters(ctx, rez.ListOncallRostersParams{
		UserID: userId,
	})
	if rostersErr != nil {
		return nil, apiError("failed to list rosters", rostersErr)
	}

	watchedRosters, watchedErr := user.QueryWatchedOncallRosters().All(ctx)
	if watchedErr != nil {
		return nil, apiError("failed to query watched oncall rosters", watchedErr)
	}

	oneWeek := time.Hour * 24 * 7
	userShifts, shiftsErr := h.oncall.ListShifts(ctx, rez.ListOncallShiftsParams{
		UserID: userId,
		Anchor: time.Now(),
		Window: oneWeek,
		ListParams: rez.ListParams{
			Limit: 20,
		},
	})
	if shiftsErr != nil {
		return nil, apiError("failed to query user oncall shifts", shiftsErr)
	}

	details := oapi.UserOncallInformation{
		ActiveShifts:    make([]oapi.OncallShift, 0),
		UpcomingShifts:  make([]oapi.OncallShift, 0),
		PastShifts:      make([]oapi.OncallShift, 0),
		MemberRosters:   make([]oapi.OncallRoster, len(memberRosters)),
		WatchingRosters: make([]oapi.OncallRoster, len(watchedRosters)),
	}

	for i, r := range memberRosters {
		details.MemberRosters[i] = oapi.OncallRosterFromEnt(r)
	}

	for i, r := range watchedRosters {
		details.WatchingRosters[i] = oapi.OncallRosterFromEnt(r)
	}

	for _, s := range userShifts {
		shift := oapi.OncallShiftFromEnt(s)
		if s.EndAt.Before(time.Now()) {
			details.PastShifts = append(details.PastShifts, shift)
		} else if s.StartAt.Before(time.Now()) {
			details.ActiveShifts = append(details.ActiveShifts, shift)
		} else {
			details.UpcomingShifts = append(details.UpcomingShifts, shift)
		}
	}

	resp.Body.Data = details

	return &resp, nil
}
