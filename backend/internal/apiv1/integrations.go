package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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

	supportedIntegrations := integrations.GetAvailable()
	resp.Body.Data = make([]oapi.SupportedIntegration, len(supportedIntegrations))
	for i, p := range supportedIntegrations {
		resp.Body.Data[i] = oapi.SupportedIntegrationFromPackage(p)
	}

	return &resp, nil
}

func (h *integrationsHandler) ListConfigured(ctx context.Context, req *oapi.ListConfiguredIntegrationsRequest) (*oapi.ListConfiguredIntegrationsResponse, error) {
	var resp oapi.ListConfiguredIntegrationsResponse

	params := rez.ListIntegrationsParams{}
	results, listErr := h.integrations.ListConfigured(ctx, params)
	if listErr != nil {
		return nil, apiError("failed to list integrations", listErr)
	}

	resp.Body.Data = make([]oapi.ConfiguredIntegration, len(results))
	for i, intg := range results {
		oapiIntg, intgErr := oapi.ConfiguredIntegrationFromConfig(intg)
		if intgErr != nil {
			return nil, apiError("failed to convert integration", intgErr)
		}
		resp.Body.Data[i] = *oapiIntg
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(results),
	}

	return &resp, nil
}

func (h *integrationsHandler) GetIntegration(ctx context.Context, req *oapi.GetIntegrationRequest) (*oapi.GetIntegrationResponse, error) {
	var resp oapi.GetIntegrationResponse

	ci, getErr := h.integrations.GetConfigured(ctx, req.Name)
	if getErr != nil {
		return nil, apiError("failed to get integration", getErr)
	}
	oapiIntg, intgErr := oapi.ConfiguredIntegrationFromConfig(ci)
	if intgErr != nil {
		return nil, apiError("failed to convert integration", intgErr)
	}
	resp.Body.Data = *oapiIntg

	return &resp, nil
}

func (h *integrationsHandler) ConfigureIntegration(ctx context.Context, req *oapi.ConfigureIntegrationRequest) (*oapi.ConfigureIntegrationResponse, error) {
	var resp oapi.ConfigureIntegrationResponse

	attr := req.Body.Attributes
	setFn := func(m *ent.IntegrationMutation) {
		if attr.Config != nil {
			m.SetConfig(attr.Config)
		}
		if attr.UserPreferences != nil {
			m.SetUserPreferences(attr.UserPreferences)
		}
	}
	ci, setErr := h.integrations.SetIntegration(ctx, req.Name, setFn)
	if setErr != nil {
		return nil, apiError("failed to configure integration", setErr)
	}
	oapiIntg, intgErr := oapi.ConfiguredIntegrationFromConfig(ci)
	if intgErr != nil {
		return nil, apiError("failed to convert integration", intgErr)
	}
	resp.Body.Data = *oapiIntg

	return &resp, nil
}

func (h *integrationsHandler) DeleteIntegration(ctx context.Context, req *oapi.DeleteIntegrationRequest) (*oapi.DeleteIntegrationResponse, error) {
	var resp oapi.DeleteIntegrationResponse

	if delErr := h.integrations.DeleteConfigured(ctx, req.Name); delErr != nil {
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

	ci, completeErr := h.integrations.CompleteOAuth2Flow(ctx, req.Name, attr.State, attr.Code)
	if completeErr != nil {
		return nil, apiError("failed to complete integration", completeErr)
	}
	oapiIntg, intgErr := oapi.ConfiguredIntegrationFromConfig(ci)
	if intgErr != nil {
		return nil, apiError("failed to convert integration", intgErr)
	}
	resp.Body.Data = *oapiIntg

	return &resp, nil
}
