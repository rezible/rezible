package apiv1

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type integrationsHandler struct {
	integrations rez.IntegrationsService
}

func newIntegrationsHandler(integrations rez.IntegrationsService) *integrationsHandler {
	return &integrationsHandler{integrations: integrations}
}

func (h *integrationsHandler) ListSupportedIntegrations(ctx context.Context, req *oapi.ListSupportedIntegrationsRequest) (*oapi.ListSupportedIntegrationsResponse, error) {
	var resp oapi.ListSupportedIntegrationsResponse

	supportedIntegrations := []oapi.SupportedIntegration{
		{
			Name:          "slack",
			DataKinds:     []string{"chat", "users"},
			OAuthRequired: true,
		},
	}

	if rez.Config.DebugMode() {
		fakeIntegration := oapi.SupportedIntegration{
			Name:          "fake",
			DataKinds:     []string{},
			OAuthRequired: false,
		}
		supportedIntegrations = append(supportedIntegrations, fakeIntegration)
	}

	resp.Body.Data = supportedIntegrations

	return &resp, nil
}

func (h *integrationsHandler) ListIntegrations(ctx context.Context, req *oapi.ListIntegrationsRequest) (*oapi.ListIntegrationsResponse, error) {
	var resp oapi.ListIntegrationsResponse

	params := rez.ListIntegrationsParams{}
	results, listErr := h.integrations.ListIntegrations(ctx, params)
	if listErr != nil {
		return nil, apiError("failed to list integrations", listErr)
	}

	resp.Body.Data = make([]oapi.Integration, len(results))
	for i, intg := range results {
		resp.Body.Data[i] = oapi.IntegrationFromEnt(intg)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(results),
	}

	return &resp, nil
}

func (h *integrationsHandler) CreateIntegration(ctx context.Context, req *oapi.CreateIntegrationRequest) (*oapi.CreateIntegrationResponse, error) {
	var resp oapi.CreateIntegrationResponse

	attr := req.Body.Attributes

	setFn := func(m *ent.IntegrationMutation) {
		m.SetName(attr.Name)
		m.SetConfig(attr.Config)
	}

	created, createErr := h.integrations.SetIntegration(ctx, uuid.Nil, setFn)
	if createErr != nil {
		return nil, apiError("failed to update integration", createErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(created)

	return &resp, nil
}

func (h *integrationsHandler) GetIntegration(ctx context.Context, req *oapi.GetIntegrationRequest) (*oapi.GetIntegrationResponse, error) {
	var resp oapi.GetIntegrationResponse

	intg, getErr := h.integrations.GetIntegration(ctx, req.Id)
	if getErr != nil {
		return nil, apiError("failed to get integration", getErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(intg)

	return &resp, nil
}

func (h *integrationsHandler) UpdateIntegration(ctx context.Context, req *oapi.UpdateIntegrationRequest) (*oapi.UpdateIntegrationResponse, error) {
	var resp oapi.UpdateIntegrationResponse

	attr := req.Body.Attributes

	setFn := func(m *ent.IntegrationMutation) {
		if attr.Config != nil {
			m.SetConfig(*attr.Config)
		}
	}

	created, createErr := h.integrations.SetIntegration(ctx, req.Id, setFn)
	if createErr != nil {
		return nil, apiError("failed to update integration", createErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(created)

	return &resp, nil
}

func (h *integrationsHandler) DeleteIntegration(ctx context.Context, req *oapi.DeleteIntegrationRequest) (*oapi.DeleteIntegrationResponse, error) {
	var resp oapi.DeleteIntegrationResponse

	if delErr := h.integrations.DeleteIntegration(ctx, req.Id); delErr != nil {
		return nil, apiError("failed to delete integration", delErr)
	}

	return &resp, nil
}

func (h *integrationsHandler) StartIntegrationOAuth(ctx context.Context, req *oapi.StartIntegrationOAuthRequest) (*oapi.StartIntegrationOAuthResponse, error) {
	var resp oapi.StartIntegrationOAuthResponse

	attr := req.Body.Attributes

	startFlowUrl, flowErr := h.integrations.StartOAuth2Flow(ctx, attr.Name)
	if flowErr != nil {
		return nil, apiError("failed to start flow", flowErr)
	}
	resp.Body.Data = oapi.IntegrationOAuthFlow{FlowUrl: startFlowUrl}

	return &resp, nil
}

func (h *integrationsHandler) CompleteIntegrationOAuth(ctx context.Context, req *oapi.CompleteIntegrationOAuthRequest) (*oapi.CompleteIntegrationOAuthResponse, error) {
	var resp oapi.CompleteIntegrationOAuthResponse

	attr := req.Body.Attributes

	intg, completeErr := h.integrations.CompleteOAuth2Flow(ctx, attr.Name, attr.State, attr.Code)
	if completeErr != nil {
		return nil, apiError("failed to complete integration", completeErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(intg)

	return &resp, nil
}
