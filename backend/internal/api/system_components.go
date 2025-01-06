package api

import (
	"context"
	oapi "github.com/rezible/rezible/openapi"
)

type systemComponentsHandler struct {
}

func newSystemComponentsHandler() *systemComponentsHandler {
	return &systemComponentsHandler{}
}

func (h *systemComponentsHandler) ListSystemComponents(ctx context.Context, request *oapi.ListSystemComponentsRequest) (*oapi.ListSystemComponentsResponse, error) {
	var resp oapi.ListSystemComponentsResponse

	return &resp, nil
}

func (h *systemComponentsHandler) ListIncidentSystemComponents(ctx context.Context, request *oapi.ListIncidentSystemComponentsRequest) (*oapi.ListIncidentSystemComponentsResponse, error) {
	var resp oapi.ListIncidentSystemComponentsResponse

	return &resp, nil
}

func (h *systemComponentsHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) UpdateSystemComponent(ctx context.Context, request *oapi.UpdateSystemComponentRequest) (*oapi.UpdateSystemComponentResponse, error) {
	var resp oapi.UpdateSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	return &resp, nil
}
