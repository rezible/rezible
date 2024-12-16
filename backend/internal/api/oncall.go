package api

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallusershiftannotation"
	"github.com/texm/prosemirror-go"
	"time"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallHandler struct {
	auth      rez.AuthService
	users     rez.UserService
	incidents rez.IncidentService
	oncall    rez.OncallService
	alerts    rez.AlertsService
}

func newOncallHandler(auth rez.AuthService, users rez.UserService, inc rez.IncidentService, oncall rez.OncallService, alerts rez.AlertsService) *oncallHandler {
	return &oncallHandler{auth: auth, users: users, incidents: inc, oncall: oncall, alerts: alerts}
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

func (h *oncallHandler) GetUserOncallDetails(ctx context.Context, request *oapi.GetUserOncallDetailsRequest) (*oapi.GetUserOncallDetailsResponse, error) {
	var resp oapi.GetUserOncallDetailsResponse

	sess, sessErr := h.auth.GetSession(ctx)
	if sessErr != nil {
		return nil, detailError("failed to get session", sessErr)
	}

	userId := sess.UserId
	if request.UserId != uuid.Nil {
		userId = request.UserId
	}

	rosters, rostersErr := h.oncall.ListRosters(ctx, rez.ListOncallRostersParams{
		UserID: userId,
	})
	if rostersErr != nil {
		return nil, detailError("failed to list rosters", rostersErr)
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

	details := oapi.UserOncallDetails{
		ActiveShifts:   make([]oapi.OncallShift, 0),
		UpcomingShifts: make([]oapi.OncallShift, 0),
		PastShifts:     make([]oapi.OncallShift, 0),
	}

	details.Rosters = make([]oapi.OncallRoster, len(rosters))
	for i, r := range rosters {
		details.Rosters[i] = oapi.OncallRosterFromEnt(r)
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
		return nil, detailError("failed to query shift", shiftErr)
	}
	resp.Body.Data = oapi.OncallShiftFromEnt(shift)

	return &resp, nil
}

func (h *oncallHandler) ListOncallShiftIncidents(ctx context.Context, request *oapi.ListOncallShiftIncidentsRequest) (*oapi.ListOncallShiftIncidentsResponse, error) {
	var resp oapi.ListOncallShiftIncidentsResponse

	shift, shiftErr := h.oncall.GetShiftByID(ctx, request.Id)
	if shiftErr != nil {
		return nil, detailError("failed to get shift", shiftErr)
	}

	incidents, incidentsErr := h.incidents.ListIncidents(ctx, rez.ListIncidentsParams{
		ListParams: request.ListParams(),
		//OpenedAfter:  shift.StartAt,
		OpenedBefore: shift.EndAt,
	})
	if incidentsErr != nil {
		return nil, detailError("failed to query incidents", incidentsErr)
	}

	resp.Body.Data = make([]oapi.Incident, len(incidents))
	for i, inc := range incidents {
		resp.Body.Data[i] = oapi.IncidentFromEnt(inc)
	}

	return &resp, nil
}

func (h *oncallHandler) ListOncallShiftAlerts(ctx context.Context, request *oapi.ListOncallShiftAlertsRequest) (*oapi.ListOncallShiftAlertsResponse, error) {
	var resp oapi.ListOncallShiftAlertsResponse

	shift, shiftErr := h.oncall.GetShiftByID(ctx, request.Id)
	if shiftErr != nil {
		return nil, detailError("failed to get shift", shiftErr)
	}

	alerts, alertsErr := h.alerts.ListAlerts(ctx, rez.ListAlertsParams{
		ListParams: request.ListParams(),
		Start:      shift.StartAt,
		End:        shift.EndAt,
	})
	if alertsErr != nil {
		return nil, detailError("failed to query alerts", alertsErr)
	}

	resp.Body.Data = make([]oapi.OncallAlert, len(alerts))
	for i, alert := range alerts {
		resp.Body.Data[i] = oapi.OncallAlertFromEnt(alert)
	}

	return &resp, nil
}

func (h *oncallHandler) CreateOncallShiftHandoverTemplate(ctx context.Context, request *oapi.CreateOncallShiftHandoverTemplateRequest) (*oapi.CreateOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.CreateOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallHandler) GetOncallShiftHandoverTemplate(ctx context.Context, request *oapi.GetOncallShiftHandoverTemplateRequest) (*oapi.GetOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.GetOncallShiftHandoverTemplateResponse

	sections := []oapi.OncallShiftHandoverTemplateSection{
		{
			Type:   "regular",
			Header: "Overview",
		},
		{
			Type:   "regular",
			Header: "Handoff Tasks",
			List:   true,
		},
		{
			Type:   "regular",
			Header: "Things to Monitor",
			List:   true,
		},
		{
			Type:   "annotations",
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

	handover, handoverErr := h.oncall.GetShiftHandover(ctx, request.Id)
	if handoverErr != nil {
		return nil, detailError("failed to get handover", handoverErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(handover)

	return &resp, nil
}

func (h *oncallHandler) UpdateOncallShiftHandover(context.Context, *oapi.UpdateOncallShiftHandoverRequest) (*oapi.UpdateOncallShiftHandoverResponse, error) {
	var resp oapi.UpdateOncallShiftHandoverResponse

	return &resp, nil
}

func (h *oncallHandler) SendOncallShiftHandover(ctx context.Context, request *oapi.SendOncallShiftHandoverRequest) (*oapi.SendOncallShiftHandoverResponse, error) {
	var resp oapi.SendOncallShiftHandoverResponse

	reqContent := request.Body.Attributes.Content
	sections := make([]rez.OncallShiftHandoverSection, len(reqContent))
	for i, sec := range reqContent {
		hoSec := rez.OncallShiftHandoverSection{
			Header:  sec.Header,
			Kind:    sec.Kind,
			Content: nil,
		}

		if sec.Kind == "regular" {
			if sec.JsonContent == nil {
				return nil, oapi.ErrorBadRequest("no content provided")
			}
			var content prosemirror.Node
			if jsonErr := json.Unmarshal([]byte(*sec.JsonContent), &content); jsonErr != nil {
				return nil, oapi.ErrorBadRequest("invalid section content", jsonErr)
			}
			hoSec.Content = &content
		}

		sections[i] = hoSec
	}

	handover, sendErr := h.oncall.SendShiftHandover(ctx, request.Id, sections)
	if sendErr != nil {
		return nil, detailError("failed to send handover", sendErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(handover)

	return &resp, nil
}

func (h *oncallHandler) ListOncallShiftAnnotations(ctx context.Context, request *oapi.ListOncallShiftAnnotationsRequest) (*oapi.ListOncallShiftAnnotationsResponse, error) {
	var resp oapi.ListOncallShiftAnnotationsResponse

	annos, annosErr := h.oncall.ListShiftAnnotations(ctx, rez.ListOncallShiftAnnotationsParams{
		ListParams: request.ListParams(),
		ShiftID:    request.Id,
	})
	if annosErr != nil {
		return nil, detailError("query shift annotations", annosErr)
	}

	resp.Body.Data = make([]oapi.OncallShiftAnnotation, len(annos))
	for i, anno := range annos {
		resp.Body.Data[i] = oapi.OncallShiftAnnotationFromEnt(anno)
	}

	return &resp, nil
}

func (h *oncallHandler) CreateOncallShiftAnnotation(ctx context.Context, request *oapi.CreateOncallShiftAnnotationRequest) (*oapi.CreateOncallShiftAnnotationResponse, error) {
	var resp oapi.CreateOncallShiftAnnotationResponse

	attr := request.Body.Attributes
	eventKind := oncallusershiftannotation.EventKind(attr.EventKind)
	if kindErr := oncallusershiftannotation.EventKindValidator(eventKind); kindErr != nil {
		return nil, detailError("invalid event kind", kindErr)
	}

	occurredAt, occErr := time.Parse(time.RFC3339, attr.OccurredAt)
	if occErr != nil {
		return nil, detailError("failed to parse occurred at datetime", occErr)
	}

	anno := &ent.OncallUserShiftAnnotation{
		ShiftID:         request.Id,
		EventID:         attr.EventId,
		EventKind:       eventKind,
		Title:           attr.Title,
		OccurredAt:      occurredAt,
		MinutesOccupied: attr.MinutesOccupied,
		Pinned:          attr.Pinned,
		Notes:           attr.Notes,
	}

	var createErr error
	anno, createErr = h.oncall.CreateShiftAnnotation(ctx, anno)
	if createErr != nil {
		return nil, detailError("failed to create annotation", createErr)
	}
	resp.Body.Data = oapi.OncallShiftAnnotationFromEnt(anno)

	return &resp, nil
}

func (h *oncallHandler) UpdateOncallShiftAnnotation(ctx context.Context, request *oapi.UpdateOncallShiftAnnotationRequest) (*oapi.UpdateOncallShiftAnnotationResponse, error) {
	var resp oapi.UpdateOncallShiftAnnotationResponse

	anno, annoErr := h.oncall.GetShiftAnnotation(ctx, request.Id)
	if annoErr != nil {
		return nil, detailError("failed to get annotation", annoErr)
	}

	attr := request.Body.Attributes
	update := anno.Update().
		SetNillableNotes(attr.Notes).
		SetNillableMinutesOccupied(attr.MinutesOccupied).
		SetNillablePinned(attr.Pinned)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update annotation", updateErr)
	}
	resp.Body.Data = oapi.OncallShiftAnnotationFromEnt(updated)

	return &resp, nil
}

func (h *oncallHandler) ArchiveOncallShiftAnnotation(ctx context.Context, request *oapi.ArchiveOncallShiftAnnotationRequest) (*oapi.ArchiveOncallShiftAnnotationResponse, error) {
	var resp oapi.ArchiveOncallShiftAnnotationResponse

	if err := h.oncall.ArchiveShiftAnnotation(ctx, request.Id); err != nil {
		return nil, detailError("failed to archive annotation", err)
	}

	return &resp, nil
}
