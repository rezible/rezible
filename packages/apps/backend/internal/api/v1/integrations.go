package apiv1

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type integrationsHandler struct {
	integrations rez.IntegrationService
}

func newIntegrationsHandler(integrations rez.IntegrationService) *integrationsHandler {
	return &integrationsHandler{integrations: integrations}
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

func (h *integrationsHandler) ListInstalledIntegrations(ctx context.Context, req *oapi.ListInstalledIntegrationsRequest) (*oapi.ListInstalledIntegrationsResponse, error) {
	var resp oapi.ListInstalledIntegrationsResponse

	params := rez.ListIntegrationsParams{}
	results, listErr := h.integrations.ListInstalled(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list integrations", listErr)
	}

	resp.Body.Data = make([]oapi.InstalledIntegration, len(results))
	for i, intg := range results {
		resp.Body.Data[i] = oapi.InstalledIntegrationFromRez(intg)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(results),
	}

	return &resp, nil
}

func (h *integrationsHandler) GetInstalledIntegration(ctx context.Context, req *oapi.GetInstalledIntegrationRequest) (*oapi.GetInstalledIntegrationResponse, error) {
	var resp oapi.GetInstalledIntegrationResponse

	intg, getErr := h.integrations.GetInstalled(ctx, req.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "failed to get integration", getErr)
	}
	resp.Body.Data = oapi.InstalledIntegrationFromRez(intg)

	return &resp, nil
}

func (h *integrationsHandler) CreateInstalledIntegration(ctx context.Context, req *oapi.CreateInstalledIntegrationRequest) (*oapi.CreateInstalledIntegrationResponse, error) {
	var resp oapi.CreateInstalledIntegrationResponse

	attr := req.Body.Attributes
	params := rez.InstallIntegrationParams{
		DisplayName:        attr.DisplayName,
		InstallationConfig: attr.Config,
		UserSettings:       attr.Preferences,
	}
	intg, installErr := h.integrations.InstallNew(ctx, req.Name, params)
	if installErr != nil {
		return nil, oapi.Error(ctx, "failed to install integration", installErr)
	}
	resp.Body.Data = oapi.InstalledIntegrationFromRez(intg)

	return &resp, nil
}

func (h *integrationsHandler) UpdateInstalledIntegration(ctx context.Context, req *oapi.UpdateInstalledIntegrationRequest) (*oapi.UpdateInstalledIntegrationResponse, error) {
	var resp oapi.UpdateInstalledIntegrationResponse

	attr := req.Body.Attributes
	params := rez.UpdateIntegrationParams{
		DisplayName:  attr.DisplayName,
		UserSettings: attr.Preferences,
	}
	intg, setErr := h.integrations.UpdateInstalled(ctx, req.Id, params)
	if setErr != nil {
		return nil, oapi.Error(ctx, "failed to configure integration", setErr)
	}
	resp.Body.Data = oapi.InstalledIntegrationFromRez(intg)

	return &resp, nil
}

func (h *integrationsHandler) DeleteInstalledIntegration(ctx context.Context, req *oapi.DeleteInstalledIntegrationRequest) (*oapi.DeleteInstalledIntegrationResponse, error) {
	var resp oapi.DeleteInstalledIntegrationResponse

	if delErr := h.integrations.DeleteInstalled(ctx, req.Id); delErr != nil {
		return nil, oapi.Error(ctx, "failed to delete integration", delErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) StartIntegrationOAuthFlow(ctx context.Context, req *oapi.StartIntegrationOAuthFlowRequest) (*oapi.StartIntegrationOAuthFlowResponse, error) {
	var resp oapi.StartIntegrationOAuthFlowResponse

	startFlowUrl, flowErr := h.integrations.StartOAuth2Flow(ctx, req.Name)
	if flowErr != nil {
		return nil, oapi.Error(ctx, "failed to start oauth2 flow", flowErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlow{FlowUrl: startFlowUrl}

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
	resp.Body.Data = oapi.IntegrationOAuthFlowResultFromRez(req.Name, result)

	return &resp, nil
}

func (h *integrationsHandler) ListIntegrationInstallTargets(ctx context.Context, req *oapi.ListIntegrationInstallTargetsRequest) (*oapi.ListIntegrationInstallTargetsResponse, error) {
	var resp oapi.ListIntegrationInstallTargetsResponse

	targets, listErr := h.integrations.ListUserInstallationTargets(ctx)
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list integration install targets", listErr)
	}
	resp.Body.Data = make([]oapi.IntegrationInstallTarget, 0)
	for name, intgTargets := range targets {
		resp.Body.Data = append(resp.Body.Data,
			oapi.IntegrationInstallTargetOptionsFromRez(name, intgTargets)...)
	}

	return &resp, nil
}

func (h *integrationsHandler) InstallIntegrationTargets(ctx context.Context, req *oapi.InstallIntegrationTargetsRequest) (*oapi.InstallIntegrationTargetsResponse, error) {
	var resp oapi.InstallIntegrationTargetsResponse

	attr := req.Body.Attributes
	results, installErr := h.integrations.InstallFromUserInstallationTargets(ctx, req.Name, attr.ExternalRefs)
	if installErr != nil {
		return nil, oapi.Error(ctx, "failed to install selected integrations", installErr)
	}

	resp.Body.Data = make([]oapi.InstalledIntegration, len(results))
	for i, intg := range results {
		resp.Body.Data[i] = oapi.InstalledIntegrationFromRez(intg)
	}

	return &resp, nil
}

func (h *integrationsHandler) RequestIntegrationEventSync(ctx context.Context, req *oapi.RequestIntegrationEventSyncRequest) (*oapi.RequestIntegrationEventSyncResponse, error) {
	var resp oapi.RequestIntegrationEventSyncResponse

	if requestErr := h.integrations.RequestIntegrationEventSync(ctx, req.Id, req.Body.Attributes.Sources); requestErr != nil {
		return nil, oapi.Error(ctx, "failed to request integration data sync", requestErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) ListIntegrationEventSyncRun(ctx context.Context, req *oapi.ListIntegrationEventSyncRunRequest) (*oapi.ListIntegrationEventSyncRunResponse, error) {
	var resp oapi.ListIntegrationEventSyncRunResponse

	result, completeErr := h.integrations.ListIntegrationEventSyncRuns(ctx, req.Id)
	if completeErr != nil {
		return nil, oapi.Error(ctx, "failed to complete integration", completeErr)
	}
	resp.Body.Data = make([]oapi.IntegrationEventSyncRun, len(result.Data))
	for i, r := range result.Data {
		resp.Body.Data[i] = oapi.IntegrationEventSyncRunFromEnt(r)
	}

	return &resp, nil
}
