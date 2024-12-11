package api

import (
	"context"

	oapi "github.com/rezible/rezible/openapi"
)

type integrationsHandler struct {
}

func newIntegrationsHandler() *integrationsHandler {
	return &integrationsHandler{}
}

func (h *integrationsHandler) ListIntegrations(ctx context.Context, input *oapi.ListIntegrationsRequest) (*oapi.ListIntegrationsResponse, error) {
	var resp oapi.ListIntegrationsResponse

	return &resp, nil
}

func (h *integrationsHandler) CreateIntegration(ctx context.Context, input *oapi.CreateIntegrationRequest) (*oapi.CreateIntegrationResponse, error) {
	var resp oapi.CreateIntegrationResponse

	return &resp, nil
}

func (h *integrationsHandler) GetIntegration(ctx context.Context, input *oapi.GetIntegrationRequest) (*oapi.GetIntegrationResponse, error) {
	var resp oapi.GetIntegrationResponse

	return &resp, nil
}

func (h *integrationsHandler) UpdateIntegration(ctx context.Context, input *oapi.UpdateIntegrationRequest) (*oapi.UpdateIntegrationResponse, error) {
	var resp oapi.UpdateIntegrationResponse

	return &resp, nil
}

func (h *integrationsHandler) ArchiveIntegration(ctx context.Context, input *oapi.ArchiveIntegrationRequest) (*oapi.ArchiveIntegrationResponse, error) {
	var resp oapi.ArchiveIntegrationResponse

	return &resp, nil
}
