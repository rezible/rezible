package apiv1

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type integrationsHandler struct {
	integrations rez.IntegrationsService
}

func newIntegrationsHandler(integrations rez.IntegrationsService) *integrationsHandler {
	return &integrationsHandler{
		integrations: integrations,
	}
}

func (h *integrationsHandler) ListAvailableIntegrations(ctx context.Context, req *oapi.ListAvailableIntegrationsRequest) (*oapi.ListAvailableIntegrationsResponse, error) {
	var resp oapi.ListAvailableIntegrationsResponse

	available := h.integrations.GetAvailable()
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
		return nil, oapi.Error(ctx, "failed to list integrations", listErr)
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

	ci, getErr := h.integrations.GetConfigured(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "failed to get integration", getErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}

func (h *integrationsHandler) ConfigureIntegration(ctx context.Context, req *oapi.ConfigureIntegrationRequest) (*oapi.ConfigureIntegrationResponse, error) {
	var resp oapi.ConfigureIntegrationResponse

	attr := req.Body.Attributes
	ci, setErr := h.integrations.Configure(ctx, rez.ConfigureIntegrationParams{
		Provider:    req.Name,
		DisplayName: attr.DisplayName,
		ExternalRef: attr.ExternalRef,
		Config:      attr.Config,
	})
	if setErr != nil {
		return nil, oapi.Error(ctx, "failed to configure integration", setErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}

func (h *integrationsHandler) UpdateConfiguredIntegrationPreferences(ctx context.Context, req *oapi.UpdateConfiguredIntegrationPreferencesRequest) (*oapi.UpdateConfiguredIntegrationPreferencesResponse, error) {
	var resp oapi.UpdateConfiguredIntegrationPreferencesResponse

	ci, setErr := h.integrations.UpdateConfiguredPreferences(ctx, req.Id, req.Body.Attributes.Preferences)
	if setErr != nil {
		return nil, oapi.Error(ctx, "failed to configure integration", setErr)
	}
	resp.Body.Data = oapi.ConfiguredIntegrationFromConfig(ci)

	return &resp, nil
}

func (h *integrationsHandler) DeleteConfiguredIntegration(ctx context.Context, req *oapi.DeleteConfiguredIntegrationRequest) (*oapi.DeleteConfiguredIntegrationResponse, error) {
	var resp oapi.DeleteConfiguredIntegrationResponse

	if delErr := h.integrations.DeleteConfigured(ctx, req.Id); delErr != nil {
		return nil, oapi.Error(ctx, "failed to delete integration", delErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) StartIntegrationOAuthFlow(ctx context.Context, req *oapi.StartIntegrationOAuthFlowRequest) (*oapi.StartIntegrationOAuthFlowResponse, error) {
	var resp oapi.StartIntegrationOAuthFlowResponse

	startFlowUrl, flowErr := h.integrations.StartOAuth2Flow(ctx, req.Name, req.Body.Attributes.CallbackPath)
	if flowErr != nil {
		return nil, oapi.Error(ctx, "failed to start flow", flowErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlow{FlowUrl: startFlowUrl}

	return &resp, nil
}

func (h *integrationsHandler) SelectIntegrationOAuthFlow(ctx context.Context, req *oapi.SelectIntegrationOAuthFlowRequest) (*oapi.SelectIntegrationOAuthFlowResponse, error) {
	var resp oapi.SelectIntegrationOAuthFlowResponse

	attr := req.Body.Attributes
	result, selectErr := h.integrations.SelectOAuth2Flow(ctx, req.Name, rez.SelectIntegrationOAuth2Params{
		SelectionToken: attr.SelectionToken,
		ExternalRefs:   attr.ExternalRefs,
	})
	if selectErr != nil {
		return nil, oapi.Error(ctx, "failed to select integration", selectErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlowResultFromCore(result)

	return &resp, nil
}

func (h *integrationsHandler) CompleteIntegrationOAuthFlow(ctx context.Context, req *oapi.CompleteIntegrationOAuthFlowRequest) (*oapi.CompleteIntegrationOAuthFlowResponse, error) {
	var resp oapi.CompleteIntegrationOAuthFlowResponse

	attr := req.Body.Attributes

	if attr.State == nil && attr.ClientVerifier == nil {
		return nil, oapi.Error(ctx, "invalid params", fmt.Errorf("missing state or client_verifier"))
	}
	params := rez.CompleteIntegrationOAuth2Params{
		Code:           attr.Code,
		State:          attr.State,
		ClientVerifier: attr.ClientVerifier,
	}
	result, completeErr := h.integrations.CompleteOAuth2Flow(ctx, req.Name, params)
	if completeErr != nil {
		return nil, oapi.Error(ctx, "failed to complete integration", completeErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlowResultFromCore(result)

	return &resp, nil
}

func (h *integrationsHandler) RequestIntegrationDataSync(ctx context.Context, req *oapi.RequestIntegrationDataSyncRequest) (*oapi.RequestIntegrationDataSyncResponse, error) {
	var resp oapi.RequestIntegrationDataSyncResponse

	if requestErr := h.integrations.RequestDataSync(ctx, req.Name, req.Body.Attributes.Sources); requestErr != nil {
		return nil, oapi.Error(ctx, "failed to request integration data sync", requestErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) GetIntegrationDataSyncStatus(ctx context.Context, req *oapi.GetIntegrationDataSyncStatusRequest) (*oapi.GetIntegrationDataSyncStatusResponse, error) {
	var resp oapi.GetIntegrationDataSyncStatusResponse

	result, completeErr := h.integrations.GetDataSyncStatus(ctx, req.Name)
	if completeErr != nil {
		return nil, oapi.Error(ctx, "failed to complete integration", completeErr)
	}
	resp.Body.Data = make([]oapi.IntegrationProviderDataSyncStatus, len(result.Data))
	for i, r := range result.Data {
		resp.Body.Data[i] = oapi.IntegrationProviderDataSyncStatusFromEnt(r)
	}

	return &resp, nil
}
