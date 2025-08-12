package api

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type playbooksHandler struct {
	playbooks rez.PlaybookService
}

func newPlaybooksHandler(pb rez.PlaybookService) *playbooksHandler {
	return &playbooksHandler{playbooks: pb}
}

func (h *playbooksHandler) ListPlaybooks(ctx context.Context, request *oapi.ListPlaybooksRequest) (*oapi.ListPlaybooksResponse, error) {
	var resp oapi.ListPlaybooksResponse

	playbooks, count, playbooksErr := h.playbooks.ListPlaybooks(ctx, rez.ListPlaybooksParams{})
	if playbooksErr != nil {
		return nil, apiError("failed to list playbooks", playbooksErr)
	}

	resp.Body.Data = make([]oapi.Playbook, len(playbooks))
	for i, pb := range playbooks {
		resp.Body.Data[i] = oapi.PlaybookFromEnt(pb)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: count,
	}

	return &resp, nil
}

func (h *playbooksHandler) CreatePlaybook(ctx context.Context, request *oapi.CreatePlaybookRequest) (*oapi.CreatePlaybookResponse, error) {
	var resp oapi.CreatePlaybookResponse

	attr := request.Body.Attributes
	reqPb := &ent.Playbook{
		Title:   attr.Title,
		Content: []byte(attr.Content),
	}
	pb, createErr := h.playbooks.UpdatePlaybook(ctx, reqPb)
	if createErr != nil {
		return nil, apiError("failed to create", createErr)
	}
	resp.Body.Data = oapi.PlaybookFromEnt(pb)

	return &resp, nil
}

func (h *playbooksHandler) GetPlaybook(ctx context.Context, request *oapi.GetPlaybookRequest) (*oapi.GetPlaybookResponse, error) {
	var resp oapi.GetPlaybookResponse

	pb, getErr := h.playbooks.GetPlaybook(ctx, request.Id)
	if getErr != nil {
		return nil, apiError("get playbook", getErr)
	}

	resp.Body.Data = oapi.PlaybookFromEnt(pb)

	return &resp, nil
}

func (h *playbooksHandler) UpdatePlaybook(ctx context.Context, request *oapi.UpdatePlaybookRequest) (*oapi.UpdatePlaybookResponse, error) {
	var resp oapi.UpdatePlaybookResponse

	pb, pbErr := h.playbooks.GetPlaybook(ctx, request.Id)
	if pbErr != nil {
		return nil, apiError("failed to get playbook", pbErr)
	}

	attr := request.Body.Attributes
	if attr.Content != nil {
		pb.Content = []byte(*attr.Content)
	}
	if attr.Title != nil {
		pb.Title = *attr.Title
	}

	updated, updateErr := h.playbooks.UpdatePlaybook(ctx, pb)
	if updateErr != nil {
		return nil, apiError("failed to update", updateErr)
	}
	resp.Body.Data = oapi.PlaybookFromEnt(updated)

	return &resp, nil
}

func (h *playbooksHandler) ArchivePlaybook(ctx context.Context, request *oapi.ArchivePlaybookRequest) (*oapi.ArchivePlaybookResponse, error) {
	var resp oapi.ArchivePlaybookResponse

	return &resp, nil
}
