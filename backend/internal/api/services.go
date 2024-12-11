package api

import (
	"context"
	oapi "github.com/rezible/rezible/openapi"
)

type servicesHandler struct {
}

func newServicesHandler() *servicesHandler {
	return &servicesHandler{}
}

func (h *servicesHandler) ListServices(ctx context.Context, request *oapi.ListServicesRequest) (*oapi.ListServicesResponse, error) {
	var resp oapi.ListServicesResponse

	return &resp, nil
}

func (h *servicesHandler) CreateService(ctx context.Context, request *oapi.CreateServiceRequest) (*oapi.CreateServiceResponse, error) {
	var resp oapi.CreateServiceResponse

	return &resp, nil
}

func (h *servicesHandler) GetService(ctx context.Context, request *oapi.GetServiceRequest) (*oapi.GetServiceResponse, error) {
	var resp oapi.GetServiceResponse

	return &resp, nil
}

func (h *servicesHandler) UpdateService(ctx context.Context, request *oapi.UpdateServiceRequest) (*oapi.UpdateServiceResponse, error) {
	var resp oapi.UpdateServiceResponse

	return &resp, nil
}

func (h *servicesHandler) ArchiveService(ctx context.Context, request *oapi.ArchiveServiceRequest) (*oapi.ArchiveServiceResponse, error) {
	var resp oapi.ArchiveServiceResponse

	return &resp, nil
}
