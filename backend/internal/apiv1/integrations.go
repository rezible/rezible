package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/integrations"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type integrationsHandler struct {
	integrations rez.IntegrationsService
}

func newIntegrationsHandler(integrations rez.IntegrationsService) *integrationsHandler {
	return &integrationsHandler{integrations: integrations}
}

func (h *integrationsHandler) ListSupported(ctx context.Context, req *oapi.ListSupportedIntegrationsRequest) (*oapi.ListSupportedIntegrationsResponse, error) {
	var resp oapi.ListSupportedIntegrationsResponse

	supportedIntegrations := integrations.GetSupported()
	resp.Body.Data = make([]oapi.SupportedIntegration, len(supportedIntegrations))
	for i, detail := range supportedIntegrations {
		resp.Body.Data[i] = oapi.SupportedIntegration{
			Name:          detail.Name(),
			DataKinds:     detail.SupportedDataKinds(),
			OAuthRequired: detail.OAuthConfigRequired(),
		}
	}

	return &resp, nil
}

func (h *integrationsHandler) ListConfigured(ctx context.Context, req *oapi.ListConfiguredIntegrationsRequest) (*oapi.ListConfiguredIntegrationsResponse, error) {
	var resp oapi.ListConfiguredIntegrationsResponse

	params := rez.ListIntegrationsParams{}
	results, listErr := h.integrations.ListIntegrations(ctx, params)
	if listErr != nil {
		return nil, apiError("failed to list integrations", listErr)
	}

	resp.Body.Data = make([]oapi.ConfiguredIntegration, len(results))
	for i, intg := range results {
		resp.Body.Data[i] = oapi.IntegrationFromEnt(intg)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(results),
	}

	return &resp, nil
}

func (h *integrationsHandler) GetIntegration(ctx context.Context, req *oapi.GetIntegrationRequest) (*oapi.GetIntegrationResponse, error) {
	var resp oapi.GetIntegrationResponse

	intg, getErr := h.integrations.GetIntegration(ctx, req.Name)
	if getErr != nil {
		return nil, apiError("failed to get integration", getErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(intg)

	return &resp, nil
}

func (h *integrationsHandler) ConfigureIntegration(ctx context.Context, req *oapi.ConfigureIntegrationRequest) (*oapi.ConfigureIntegrationResponse, error) {
	var resp oapi.ConfigureIntegrationResponse

	attr := req.Body.Attributes

	created, createErr := h.integrations.ConfigureIntegration(ctx, req.Name, attr.Config)
	if createErr != nil {
		return nil, apiError("failed to update integration", createErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(created)

	return &resp, nil
}

func (h *integrationsHandler) DeleteIntegration(ctx context.Context, req *oapi.DeleteIntegrationRequest) (*oapi.DeleteIntegrationResponse, error) {
	var resp oapi.DeleteIntegrationResponse

	if delErr := h.integrations.DeleteIntegration(ctx, req.Name); delErr != nil {
		return nil, apiError("failed to delete integration", delErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) StartIntegrationOAuthFlow(ctx context.Context, req *oapi.StartIntegrationOAuthFlowRequest) (*oapi.StartIntegrationOAuthFlowResponse, error) {
	var resp oapi.StartIntegrationOAuthFlowResponse

	startFlowUrl, flowErr := h.integrations.StartOAuth2Flow(ctx, req.Name)
	if flowErr != nil {
		return nil, apiError("failed to start flow", flowErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlow{FlowUrl: startFlowUrl}

	return &resp, nil
}

func (h *integrationsHandler) CompleteIntegrationOAuthFlow(ctx context.Context, req *oapi.CompleteIntegrationOAuthFlowRequest) (*oapi.CompleteIntegrationOAuthFlowResponse, error) {
	var resp oapi.CompleteIntegrationOAuthFlowResponse

	attr := req.Body.Attributes

	intg, completeErr := h.integrations.CompleteOAuth2Flow(ctx, req.Name, attr.State, attr.Code)
	if completeErr != nil {
		return nil, apiError("failed to complete integration", completeErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(intg)

	return &resp, nil
}
