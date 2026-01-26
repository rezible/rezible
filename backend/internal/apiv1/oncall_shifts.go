package apiv1

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type oncallShiftsHandler struct {
	auth      rez.AuthService
	users     rez.UserService
	incidents rez.IncidentService
	shifts    rez.OncallShiftsService
}

func newOncallShiftsHandler(auth rez.AuthService, users rez.UserService, inc rez.IncidentService, shifts rez.OncallShiftsService) *oncallShiftsHandler {
	return &oncallShiftsHandler{auth: auth, users: users, incidents: inc, shifts: shifts}
}

func (h *oncallShiftsHandler) ListOncallShifts(ctx context.Context, request *oapi.ListOncallShiftsRequest) (*oapi.ListOncallShiftsResponse, error) {
	var resp oapi.ListOncallShiftsResponse

	listParams := rez.ListOncallShiftsParams{
		ListParams: request.ListParams(),
		UserID:     request.UserId,
	}

	if request.Active {
		listParams.Anchor = time.Now()
		listParams.Window = time.Minute
	}

	listRes, shiftsErr := h.shifts.ListShifts(ctx, listParams)
	if shiftsErr != nil {
		return nil, apiError("failed to list oncall shifts", shiftsErr)
	}

	resp.Body.Data = make([]oapi.OncallShift, len(listRes.Data))
	for i, s := range listRes.Data {
		resp.Body.Data[i] = oapi.OncallShiftFromEnt(s)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: listRes.Count,
	}

	return &resp, nil
}

func (h *oncallShiftsHandler) GetOncallShift(ctx context.Context, request *oapi.GetOncallShiftRequest) (*oapi.GetOncallShiftResponse, error) {
	var resp oapi.GetOncallShiftResponse

	shift, shiftErr := h.shifts.GetShiftByID(ctx, request.Id)
	if shiftErr != nil {
		return nil, apiError("failed to query shift", shiftErr)
	}
	resp.Body.Data = oapi.OncallShiftFromEnt(shift)

	return &resp, nil
}

func (h *oncallShiftsHandler) GetAdjacentOncallShifts(ctx context.Context, request *oapi.GetAdjacentOncallShiftsRequest) (*oapi.GetAdjacentOncallShiftsResponse, error) {
	var resp oapi.GetAdjacentOncallShiftsResponse

	prev, next, shiftErr := h.shifts.GetAdjacentShifts(ctx, request.Id)
	if shiftErr != nil {
		log.Debug().Err(shiftErr).Msg("GetAdjacentOncallShifts")
		return nil, apiError("failed to query adjacent shifts", shiftErr)
	}
	var adj oapi.OncallShiftsAdjacent
	if prev != nil {
		s := oapi.OncallShiftFromEnt(prev)
		adj.Previous = &s
	}
	if next != nil {
		s := oapi.OncallShiftFromEnt(next)
		adj.Next = &s
	}
	resp.Body.Data = adj

	return &resp, nil
}

func (h *oncallShiftsHandler) CreateOncallShiftHandoverTemplate(ctx context.Context, request *oapi.CreateOncallShiftHandoverTemplateRequest) (*oapi.CreateOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.CreateOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallShiftsHandler) GetOncallShiftHandoverTemplate(ctx context.Context, request *oapi.GetOncallShiftHandoverTemplateRequest) (*oapi.GetOncallShiftHandoverTemplateResponse, error) {
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

func (h *oncallShiftsHandler) UpdateOncallShiftHandoverTemplate(ctx context.Context, request *oapi.UpdateOncallShiftHandoverTemplateRequest) (*oapi.UpdateOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.UpdateOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallShiftsHandler) ArchiveOncallShiftHandoverTemplate(ctx context.Context, request *oapi.ArchiveOncallShiftHandoverTemplateRequest) (*oapi.ArchiveOncallShiftHandoverTemplateResponse, error) {
	var resp oapi.ArchiveOncallShiftHandoverTemplateResponse

	return &resp, nil
}

func (h *oncallShiftsHandler) GetOncallShiftHandover(ctx context.Context, request *oapi.GetOncallShiftHandoverRequest) (*oapi.GetOncallShiftHandoverResponse, error) {
	var resp oapi.GetOncallShiftHandoverResponse

	handover, handoverErr := h.shifts.GetHandoverForShift(ctx, request.Id)
	if handoverErr != nil && !ent.IsNotFound(handoverErr) {
		return nil, apiError("failed to get handover", handoverErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(handover)

	return &resp, nil
}

func (h *oncallShiftsHandler) UpdateOncallShiftHandover(ctx context.Context, request *oapi.UpdateOncallShiftHandoverRequest) (*oapi.UpdateOncallShiftHandoverResponse, error) {
	var resp oapi.UpdateOncallShiftHandoverResponse

	ho := &ent.OncallShiftHandover{
		ID: request.Id,
	}
	attr := request.Body.Attributes
	if attr.Content != nil {
		contentJson, jsonErr := json.Marshal(attr.Content)
		if jsonErr != nil {
			return nil, apiError("failed to marshal content", jsonErr)
		}
		ho.Contents = contentJson
	}
	if attr.PinnedAnnotationIds != nil {
		ho.Edges.PinnedAnnotations = make([]*ent.EventAnnotation, len(*attr.PinnedAnnotationIds))
		for i, id := range *attr.PinnedAnnotationIds {
			ho.Edges.PinnedAnnotations[i] = &ent.EventAnnotation{ID: id}
		}
	}

	updated, updateErr := h.shifts.UpdateShiftHandover(ctx, ho)
	if updateErr != nil {
		return nil, apiError("failed to update handover", updateErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(updated)

	return &resp, nil
}

func (h *oncallShiftsHandler) SendOncallShiftHandover(ctx context.Context, request *oapi.SendOncallShiftHandoverRequest) (*oapi.SendOncallShiftHandoverResponse, error) {
	var resp oapi.SendOncallShiftHandoverResponse

	handover, sendErr := h.shifts.SendShiftHandover(ctx, request.Id)
	if sendErr != nil {
		return nil, apiError("failed to send handover", sendErr)
	}
	resp.Body.Data = oapi.OncallShiftHandoverFromEnt(handover)

	return &resp, nil
}
