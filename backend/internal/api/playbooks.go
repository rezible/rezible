package api

import (
	"context"
	rez "github.com/rezible/rezible"
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

	playbooks, count, playbooksErr := h.playbooks.ListPlaybooks(ctx, nil)
	if playbooksErr != nil {
		return nil, detailError("failed to list playbooks", playbooksErr)
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

	return &resp, nil
}

func (h *playbooksHandler) GetPlaybook(ctx context.Context, request *oapi.GetPlaybookRequest) (*oapi.GetPlaybookResponse, error) {
	var resp oapi.GetPlaybookResponse

	return &resp, nil
}

func (h *playbooksHandler) UpdatePlaybook(ctx context.Context, request *oapi.UpdatePlaybookRequest) (*oapi.UpdatePlaybookResponse, error) {
	var resp oapi.UpdatePlaybookResponse

	return &resp, nil
}

func (h *playbooksHandler) ArchivePlaybook(ctx context.Context, request *oapi.ArchivePlaybookRequest) (*oapi.ArchivePlaybookResponse, error) {
	var resp oapi.ArchivePlaybookResponse

	return &resp, nil
}
