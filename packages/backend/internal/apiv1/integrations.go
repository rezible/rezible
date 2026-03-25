package apiv1

import (
	"context"
	"net/url"

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

func (h *integrationsHandler) ListAvailableIntegrations(ctx context.Context, req *oapi.ListAvailableIntegrationsRequest) (*oapi.ListAvailableIntegrationsResponse, error) {
	var resp oapi.ListAvailableIntegrationsResponse

	available := integrations.GetAvailable()
	resp.Body.Data = make([]oapi.AvailableIntegration, len(available))
	for i, intg := range available {
		resp.Body.Data[i] = oapi.AvailableIntegrationFromPackage(intg)
	}

	return &resp, nil
}

func (h *integrationsHandler) ListConfiguredIntegrations(ctx context.Context, req *oapi.ListConfiguredIntegrationsRequest) (*oapi.ListConfiguredIntegrationsResponse, error) {
	var resp oapi.ListConfiguredIntegrationsResponse

	params := rez.ListIntegrationsParams{}
	results, listErr := h.integrations.ListConfigured(ctx, params)
	if listErr != nil {
		return nil, oapi.Error("failed to list integrations", listErr)
	}

	resp.Body.Data = make([]oapi.ConfiguredIntegration, len(results))
	for i, ci := range results {
		resp.Body.Data[i] = oapi.ConfiguredIntegrationFromConfig(ci)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(results),
	}

	return &resp, nil
}

func (h *integrationsHandler) GetConfiguredIntegration(ctx context.Context, req *oapi.GetConfiguredIntegrationRequest) (*oapi.GetConfiguredIntegrationResponse, error) {
	var resp oapi.GetConfiguredIntegrationResponse

	ci, getErr := h.integrations.GetConfigured(ctx, req.Name)
	if getErr != nil {
		return nil, oapi.Error("failed to get integration", getErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}

func (h *integrationsHandler) ConfigureIntegration(ctx context.Context, req *oapi.ConfigureIntegrationRequest) (*oapi.ConfigureIntegrationResponse, error) {
	var resp oapi.ConfigureIntegrationResponse

	ci, setErr := h.integrations.Configure(ctx, req.Name, req.Body.Attributes.Config)
	if setErr != nil {
		return nil, oapi.Error("failed to configure integration", setErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}

func (h *integrationsHandler) UpdateConfiguredIntegrationPreferences(ctx context.Context, req *oapi.UpdateConfiguredIntegrationPreferencesRequest) (*oapi.UpdateConfiguredIntegrationPreferencesResponse, error) {
	var resp oapi.UpdateConfiguredIntegrationPreferencesResponse

	ci, setErr := h.integrations.UpdateConfiguredPreferences(ctx, req.Name, req.Body.Attributes.Preferences)
	if setErr != nil {
		return nil, oapi.Error("failed to configure integration", setErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}

func (h *integrationsHandler) DeleteConfiguredIntegration(ctx context.Context, req *oapi.DeleteConfiguredIntegrationRequest) (*oapi.DeleteConfiguredIntegrationResponse, error) {
	var resp oapi.DeleteConfiguredIntegrationResponse

	if delErr := h.integrations.DeleteConfigured(ctx, req.Name); delErr != nil {
		return nil, oapi.Error("failed to delete integration", delErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) StartIntegrationOAuthFlow(ctx context.Context, req *oapi.StartIntegrationOAuthFlowRequest) (*oapi.StartIntegrationOAuthFlowResponse, error) {
	var resp oapi.StartIntegrationOAuthFlowResponse

	callbackUrl, pathErr := url.JoinPath(rez.Config.AppUrl(), req.Body.Attributes.CallbackPath)
	if pathErr != nil {
		return nil, oapi.Error("invalid callback path", pathErr)
	}
	redirectUrl, urlErr := url.Parse(callbackUrl)
	if urlErr != nil {
		return nil, oapi.Error("invalid callback url", urlErr)
	}

	startFlowUrl, flowErr := h.integrations.StartOAuth2Flow(ctx, req.Name, redirectUrl)
	if flowErr != nil {
		return nil, oapi.Error("failed to start flow", flowErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlow{FlowUrl: startFlowUrl}

	return &resp, nil
}

func (h *integrationsHandler) CompleteIntegrationOAuthFlow(ctx context.Context, req *oapi.CompleteIntegrationOAuthFlowRequest) (*oapi.CompleteIntegrationOAuthFlowResponse, error) {
	var resp oapi.CompleteIntegrationOAuthFlowResponse

	attr := req.Body.Attributes

	ci, completeErr := h.integrations.CompleteOAuth2Flow(ctx, req.Name, attr.State, attr.Code)
	if completeErr != nil {
		return nil, oapi.Error("failed to complete integration", completeErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}
