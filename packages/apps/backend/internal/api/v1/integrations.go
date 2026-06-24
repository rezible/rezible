package apiv1

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type integrationsHandler struct {
	integrations rez.IntegrationService
}

func newIntegrationsHandler(integrations rez.IntegrationService) *integrationsHandler {
	return &integrationsHandler{integrations: integrations}
}

func (h *integrationsHandler) GetInstallableIntegrations(ctx context.Context, req *oapi.GetInstallableIntegrationsRequest) (*oapi.GetInstallableIntegrationsResponse, error) {
	var resp oapi.GetInstallableIntegrationsResponse

	available := h.integrations.GetAvailable()
	resp.Body.Data = make([]oapi.InstallableIntegration, len(available))
	for i, intg := range available {
		resp.Body.Data[i] = oapi.InstallableIntegrationFromPackage(intg)
	}

	return &resp, nil
}

func (h *integrationsHandler) InstallIntegration(ctx context.Context, req *oapi.InstallIntegrationRequest) (*oapi.InstallIntegrationResponse, error) {
	var resp oapi.InstallIntegrationResponse

	attr := req.Body.Attributes
	intg, installErr := h.integrations.InstallNew(ctx, req.Name, attr.Config, attr.UserSettings)
	if installErr != nil {
		return nil, oapi.Error(ctx, "failed to install integration", installErr)
	}
	resp.Body.Data = oapi.IntegrationInstallationFromRez(intg)

	return &resp, nil
}

func (h *integrationsHandler) ListIntegrationInstallations(ctx context.Context, req *oapi.ListIntegrationInstallationsRequest) (*oapi.ListIntegrationInstallationsResponse, error) {
	var resp oapi.ListIntegrationInstallationsResponse

	params := rez.ListIntegrationsParams{}
	results, listErr := h.integrations.ListInstalled(ctx, params)
	if listErr != nil {
		return nil, oapi.Error(ctx, "failed to list integrations", listErr)
	}

	resp.Body.Data = make([]oapi.IntegrationInstallation, len(results))
	for i, intg := range results {
		resp.Body.Data[i] = oapi.IntegrationInstallationFromRez(intg)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(results),
	}

	return &resp, nil
}

func (h *integrationsHandler) GetIntegrationInstallation(ctx context.Context, req *oapi.GetIntegrationInstallationRequest) (*oapi.GetIntegrationInstallationResponse, error) {
	var resp oapi.GetIntegrationInstallationResponse

	intg, lookupErr := h.integrations.LookupInstallation(ctx, integration.ID(req.Id))
	if lookupErr != nil {
		return nil, oapi.Error(ctx, "failed to get integration", lookupErr)
	}
	inst, instErr := h.integrations.AsInstalledIntegration(intg)
	if instErr != nil {
		return nil, oapi.Error(ctx, "failed to get integration", instErr)
	}
	resp.Body.Data = oapi.IntegrationInstallationFromRez(inst)

	return &resp, nil
}

func (h *integrationsHandler) UpdateIntegrationInstallation(ctx context.Context, req *oapi.UpdateIntegrationInstallationRequest) (*oapi.UpdateIntegrationInstallationResponse, error) {
	var resp oapi.UpdateIntegrationInstallationResponse

	setFn := func(m *ent.IntegrationMutation) {
		m.SetUserSettings(req.Body.Attributes.UserSettings)
	}
	intg, setErr := h.integrations.UpdateInstallation(ctx, req.Id, setFn)
	if setErr != nil {
		return nil, oapi.Error(ctx, "failed to configure integration", setErr)
	}
	resp.Body.Data = oapi.IntegrationInstallationFromRez(intg)

	return &resp, nil
}

func (h *integrationsHandler) DeleteIntegrationInstallation(ctx context.Context, req *oapi.DeleteIntegrationInstallationRequest) (*oapi.DeleteIntegrationInstallationResponse, error) {
	var resp oapi.DeleteIntegrationInstallationResponse

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

func (h *integrationsHandler) InstallIntegrationFromTargets(ctx context.Context, req *oapi.InstallIntegrationFromTargetsRequest) (*oapi.InstallIntegrationFromTargetsResponse, error) {
	var resp oapi.InstallIntegrationFromTargetsResponse

	attr := req.Body.Attributes
	results, installErr := h.integrations.InstallFromUserInstallationTargets(ctx, req.Name, attr.ExternalRefs)
	if installErr != nil {
		return nil, oapi.Error(ctx, "failed to install selected integrations", installErr)
	}

	resp.Body.Data = make([]oapi.IntegrationInstallation, len(results))
	for i, intg := range results {
		resp.Body.Data[i] = oapi.IntegrationInstallationFromRez(intg)
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
