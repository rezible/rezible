package api

import (
	"context"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type playbooksHandler struct {
	db *ent.Client
}

func newPlaybooksHandler(db *ent.Client) *playbooksHandler {
	return &playbooksHandler{db: db}
}

func (h *playbooksHandler) ListPlaybooks(ctx context.Context, request *oapi.ListPlaybooksRequest) (*oapi.ListPlaybooksResponse, error) {
	var resp oapi.ListPlaybooksResponse

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
