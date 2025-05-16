package api

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"time"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallHandler struct {
	auth      rez.AuthSessionService
	users     rez.UserService
	incidents rez.IncidentService
	oncall    rez.OncallService
}

func newOncallHandler(auth rez.AuthSessionService, users rez.UserService, inc rez.IncidentService, oncall rez.OncallService) *oncallHandler {
	return &oncallHandler{auth: auth, users: users, incidents: inc, oncall: oncall}
}

func (h *oncallHandler) ListOncallRosters(ctx context.Context, request *oapi.ListOncallRostersRequest) (*oapi.ListOncallRostersResponse, error) {
	var resp oapi.ListOncallRostersResponse

	rosters, rostersErr := h.oncall.ListRosters(ctx, rez.ListOncallRostersParams{
		ListParams: request.ListParams(),
		UserID:     request.UserId,
	})
	if rostersErr != nil {
		return nil, detailError("failed to list rosters", rostersErr)
	}

	resp.Body.Data = make([]oapi.OncallRoster, len(rosters))
	for i, r := range rosters {
		resp.Body.Data[i] = oapi.OncallRosterFromEnt(r)
	}

	return &resp, nil
}

func (h *oncallHandler) GetOncallRoster(ctx context.Context, request *oapi.GetOncallRosterRequest) (*oapi.GetOncallRosterResponse, error) {
	var resp oapi.GetOncallRosterResponse

	var roster *ent.OncallRoster
	var rosterErr error
	if request.Id.IsSlug {
		roster, rosterErr = h.oncall.GetRosterBySlug(ctx, request.Id.Slug)
	} else {
		roster, rosterErr = h.oncall.GetRosterByID(ctx, request.Id.UUID)
	}
	if rosterErr != nil {
		return nil, detailError("failed to get oncall roster", rosterErr)
	}

	schedules, schedulesErr := roster.QuerySchedules().All(ctx)
	if schedulesErr != nil {
		return nil, detailError("failed to query schedules", schedulesErr)
	}
	roster.Edges.Schedules = schedules

	resp.Body.Data = oapi.OncallRosterFromEnt(roster)

	return &resp, nil
}

func (h *oncallHandler) getUserWatchedOncallRosters(ctx context.Context, user *ent.User) ([]oapi.OncallRoster, error) {
	rosters, queryErr := user.QueryWatchedOncallRosters().All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query watched oncall rosters", queryErr)
	}
	watched := make([]oapi.OncallRoster, len(rosters))
	for i, r := range rosters {
		watched[i] = oapi.OncallRosterFromEnt(r)
	}
	return watched, nil
}

func (h *oncallHandler) AddWatchedOncallRoster(ctx context.Context, request *oapi.AddWatchedOncallRosterRequest) (*oapi.AddWatchedOncallRosterResponse, error) {
	var resp oapi.AddWatchedOncallRosterResponse

	user, userErr := h.users.GetById(ctx, mustGetAuthSession(ctx, h.auth).UserId)
	if userErr != nil {
		return nil, detailError("failed to get user", userErr)
	}

	if addErr := user.Update().AddWatchedOncallRosterIDs(request.Id).Exec(ctx); addErr != nil {
		return nil, detailError("failed to add watched oncall roster", addErr)
	}

	watched, queryErr := h.getUserWatchedOncallRosters(ctx, user)
	if queryErr != nil {
		return nil, detailError("failed to query watched oncall rosters", queryErr)
	}
	resp.Body.Data = watched

	return &resp, nil
}

func (h *oncallHandler) ListWatchedOncallRosters(ctx context.Context, request *oapi.ListWatchedOncallRostersRequest) (*oapi.ListWatchedOncallRostersResponse, error) {
	var resp oapi.ListWatchedOncallRostersResponse

	user, userErr := h.users.GetById(ctx, mustGetAuthSession(ctx, h.auth).UserId)
	if userErr != nil {
		return nil, detailError("failed to get user", userErr)
	}
	watched, queryErr := h.getUserWatchedOncallRosters(ctx, user)
	if queryErr != nil {
		return nil, detailError("failed to query watched oncall rosters", queryErr)
	}
	resp.Body.Data = watched

	return &resp, nil
}

func (h *oncallHandler) RemoveWatchedOncallRoster(ctx context.Context, request *oapi.RemoveWatchedOncallRosterRequest) (*oapi.RemoveWatchedOncallRosterResponse, error) {
	var resp oapi.RemoveWatchedOncallRosterResponse

	user, userErr := h.users.GetById(ctx, mustGetAuthSession(ctx, h.auth).UserId)
	if userErr != nil {
		return nil, detailError("failed to get user", userErr)
	}

	if addErr := user.Update().RemoveWatchedOncallRosterIDs(request.Id).Exec(ctx); addErr != nil {
		return nil, detailError("failed to add watched oncall roster", addErr)
	}

	watched, queryErr := h.getUserWatchedOncallRosters(ctx, user)
	if queryErr != nil {
		return nil, detailError("failed to query watched oncall rosters", queryErr)
	}
	resp.Body.Data = watched

	return &resp, nil
}

func (h *oncallHandler) GetUserOncallInformation(ctx context.Context, request *oapi.GetUserOncallInformationRequest) (*oapi.GetUserOncallInformationResponse, error) {
	var resp oapi.GetUserOncallInformationResponse

	sess, sessErr := h.auth.GetSession(ctx)
	if sessErr != nil {
		return nil, detailError("failed to get session", sessErr)
	}

	userId := sess.UserId
	if request.UserId != uuid.Nil {
		userId = request.UserId
	}

	user, userErr := h.users.GetById(ctx, userId)
	if userErr != nil {
		return nil, detailError("failed to get user", userErr)
	}

	memberRosters, rostersErr := h.oncall.ListRosters(ctx, rez.ListOncallRostersParams{
		UserID: userId,
	})
	if rostersErr != nil {
		return nil, detailError("failed to list rosters", rostersErr)
	}

	watchedRosters, watchedErr := user.QueryWatchedOncallRosters().All(ctx)
	if watchedErr != nil {
		return nil, detailError("failed to query watched oncall rosters", watchedErr)
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
		return nil, detailError("failed to query user oncall shifts", shiftsErr)
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

func (h *oncallHandler) ListOncallShifts(ctx context.Context, request *oapi.ListOncallShiftsRequest) (*oapi.ListOncallShiftsResponse, error) {
	var resp oapi.ListOncallShiftsResponse

	listParams := rez.ListOncallShiftsParams{
		ListParams: request.ListParams(),
		UserID:     request.UserId,
	}

	if request.Active {
		listParams.Anchor = time.Now()
		listParams.Window = time.Minute
	}

	shifts, shiftsErr := h.oncall.ListShifts(ctx, listParams)
	if shiftsErr != nil {
		return nil, detailError("failed to list oncall shifts", shiftsErr)
	}

	numShifts := len(shifts)
	resp.Body.Data = make([]oapi.OncallShift, numShifts)
	for i, s := range shifts {
		resp.Body.Data[numShifts-i-1] = oapi.OncallShiftFromEnt(s)
	}

	return &resp, nil
}

func (h *oncallHandler) GetOncallShift(ctx context.Context, request *oapi.GetOncallShiftRequest) (*oapi.GetOncallShiftResponse, error) {
	var resp oapi.GetOncallShiftResponse

	shift, shiftErr := h.oncall.GetShiftByID(ctx, request.Id)
	if shiftErr != nil {
		return nil, detailError("failed to query shift", shiftErr)
	}
	resp.Body.Data = oapi.OncallShiftFromEnt(shift)

	return &resp, nil
}

func (h *oncallHandler) GetNextOncallShift(ctx context.Context, request *oapi.GetNextOncallShiftRequest) (*oapi.GetNextOncallShiftResponse, error) {
	var resp oapi.GetNextOncallShiftResponse

	shift, shiftErr := h.oncall.GetNextShift(ctx, request.Id)
	if shiftErr != nil {
		return nil, detailError("failed to query next shift", shiftErr)
	}
	resp.Body.Data = oapi.OncallShiftFromEnt(shift)

	return &resp, nil
}

func (h *oncallHandler) GetPreviousOncallShift(ctx context.Context, request *oapi.GetPreviousOncallShiftRequest) (*oapi.GetPreviousOncallShiftResponse, error) {
	var resp oapi.GetPreviousOncallShiftResponse

	shift, shiftErr := h.oncall.GetPreviousShift(ctx, request.Id)
	if shiftErr != nil {
		return nil, detailError("failed to query previous shift", shiftErr)
	}
	resp.Body.Data = oapi.OncallShiftFromEnt(shift)

	return &resp, nil
}

func (h *oncallHandler) CreateOncallShiftHandoverTemplate(ctx context.Context, request *oapi.CreateOncallShiftHandoverTemplateRequest) (*oapi.CreateOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.CreateOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallHandler) GetOncallShiftHandoverTemplate(ctx context.Context, request *oapi.GetOncallShiftHandoverTemplateRequest) (*oapi.GetOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.GetOncallShiftHandoverTemplateResponse

	sections := []oapi.OncallShiftHandoverSection{
		{
			Kind:   "regular",
			Header: "Overview",
		},
		{
			Kind:   "regular",
			Header: "Handoff Tasks",
		},
		{
			Kind:   "regular",
			Header: "Things to Monitor",
		},
		{
			Kind:   "annotations",
			Header: "Event Annotations",
		},
	}

	resp.Body.Data = oapi.OncallShiftHandoverTemplate{
		Id: uuid.New(),
		Attributes: oapi.OncallShiftHandoverTemplateAttributes{
			Sections: sections,
		},
	}

	return &resp, nil
}

func (h *oncallHandler) UpdateOncallShiftHandoverTemplate(ctx context.Context, request *oapi.UpdateOncallShiftHandoverTemplateRequest) (*oapi.UpdateOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.UpdateOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallHandler) ArchiveOncallShiftHandoverTemplate(ctx context.Context, request *oapi.ArchiveOncallShiftHandoverTemplateRequest) (*oapi.ArchiveOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.ArchiveOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallHandler) GetOncallShiftHandover(ctx context.Context, request *oapi.GetOncallShiftHandoverRequest) (*oapi.GetOncallShiftHandoverResponse, error) {
	var resp oapi.GetOncallShiftHandoverResponse

	// TODO: check if should create?
	const shouldCreate = true

	handover, handoverErr := h.oncall.GetHandoverForShift(ctx, request.Id, shouldCreate)
	if handoverErr != nil && !ent.IsNotFound(handoverErr) {
		return nil, detailError("failed to get handover", handoverErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(handover)

	return &resp, nil
}

func (h *oncallHandler) UpdateOncallShiftHandover(ctx context.Context, request *oapi.UpdateOncallShiftHandoverRequest) (*oapi.UpdateOncallShiftHandoverResponse, error) {
	var resp oapi.UpdateOncallShiftHandoverResponse

	ho := &ent.OncallUserShiftHandover{
		ID: request.Id,
	}
	attr := request.Body.Attributes
	if attr.Content != nil {
		contentJson, jsonErr := json.Marshal(attr.Content)
		if jsonErr != nil {
			return nil, detailError("failed to marshal content", jsonErr)
		}
		ho.Contents = contentJson
	}
	if attr.PinnedAnnotationIds != nil {
		ho.Edges.PinnedAnnotations = make([]*ent.OncallAnnotation, len(*attr.PinnedAnnotationIds))
		for i, id := range *attr.PinnedAnnotationIds {
			ho.Edges.PinnedAnnotations[i] = &ent.OncallAnnotation{ID: id}
		}
	}

	updated, updateErr := h.oncall.UpdateShiftHandover(ctx, ho)
	if updateErr != nil {
		return nil, detailError("failed to update handover", updateErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(updated)

	return &resp, nil
}

func (h *oncallHandler) SendOncallShiftHandover(ctx context.Context, request *oapi.SendOncallShiftHandoverRequest) (*oapi.SendOncallShiftHandoverResponse, error) {
	var resp oapi.SendOncallShiftHandoverResponse

	//reqContent := request.Body.Attributes.Content
	//sections := make([]rez.OncallShiftHandoverSection, len(reqContent))
	//for i, sec := range reqContent {
	//	hoSec := rez.OncallShiftHandoverSection{
	//		Header:  sec.Header,
	//		Kind:    sec.Kind,
	//		Content: nil,
	//	}
	//
	//	if sec.Kind == "regular" {
	//		if sec.JsonContent == nil {
	//			return nil, oapi.ErrorBadRequest("no content provided")
	//		}
	//		var content prosemirror.Node
	//		if jsonErr := json.Unmarshal([]byte(*sec.JsonContent), &content); jsonErr != nil {
	//			return nil, oapi.ErrorBadRequest("invalid section content", jsonErr)
	//		}
	//		hoSec.Content = &content
	//	}
	//
	//	sections[i] = hoSec
	//}

	handover, sendErr := h.oncall.SendShiftHandover(ctx, request.Id)
	if sendErr != nil {
		return nil, detailError("failed to send handover", sendErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(handover)

	return &resp, nil
}
