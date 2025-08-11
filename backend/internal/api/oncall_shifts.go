package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi"
)

type oncallShiftsHandler struct {
	auth      rez.AuthSessionService
	users     rez.UserService
	incidents rez.IncidentService
	oncall    rez.OncallService
}

func newOncallShiftsHandler(auth rez.AuthSessionService, users rez.UserService, inc rez.IncidentService, oncall rez.OncallService) *oncallShiftsHandler {
	return &oncallShiftsHandler{auth: auth, users: users, incidents: inc, oncall: oncall}
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

func (h *oncallShiftsHandler) GetOncallShift(ctx context.Context, request *oapi.GetOncallShiftRequest) (*oapi.GetOncallShiftResponse, error) {
	var resp oapi.GetOncallShiftResponse

	shift, shiftErr := h.oncall.GetShiftByID(ctx, request.Id)
	if shiftErr != nil {
		return nil, detailError("failed to query shift", shiftErr)
	}
	resp.Body.Data = oapi.OncallShiftFromEnt(shift)

	return &resp, nil
}

func (h *oncallShiftsHandler) GetAdjacentOncallShifts(ctx context.Context, request *oapi.GetAdjacentOncallShiftsRequest) (*oapi.GetAdjacentOncallShiftsResponse, error) {
	var resp oapi.GetAdjacentOncallShiftsResponse

	prev, next, shiftErr := h.oncall.GetAdjacentShifts(ctx, request.Id)
	if shiftErr != nil {
		log.Debug().Err(shiftErr).Msg("GetAdjacentOncallShifts")
		return nil, detailError("failed to query adjacent shifts", shiftErr)
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

	handover, handoverErr := h.oncall.GetHandoverForShift(ctx, request.Id)
	if handoverErr != nil && !ent.IsNotFound(handoverErr) {
		return nil, detailError("failed to get handover", handoverErr)
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

func (h *oncallShiftsHandler) SendOncallShiftHandover(ctx context.Context, request *oapi.SendOncallShiftHandoverRequest) (*oapi.SendOncallShiftHandoverResponse, error) {
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
