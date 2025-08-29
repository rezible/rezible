package api

import (
	"context"
	"encoding/json"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	oapi "github.com/rezible/rezible/openapi"
)

type integrationsHandler struct {
	configs rez.ProviderConfigService
}

func newIntegrationsHandler(configs rez.ProviderConfigService) *integrationsHandler {
	return &integrationsHandler{configs: configs}
}

func toValidProviderType(kind string) (providerconfig.ProviderType, error) {
	pt := providerconfig.ProviderType(kind)
	return pt, providerconfig.ProviderTypeValidator(pt)
}

func (h *integrationsHandler) ListIntegrations(ctx context.Context, req *oapi.ListIntegrationsRequest) (*oapi.ListIntegrationsResponse, error) {
	var resp oapi.ListIntegrationsResponse

	params := rez.ListProviderConfigsParams{
		ProviderId: req.ProviderId,
	}
	if req.Kind != "" {
		pt, validErr := toValidProviderType(req.Kind)
		if validErr != nil {
			return nil, apiError("invalid provider type", validErr)
		}
		params.ProviderType = pt
	}
	pcs, pcsErr := h.configs.ListProviderConfigs(ctx, params)
	if pcsErr != nil {
		return nil, apiError("failed to list provider configs", pcsErr)
	}

	resp.Body.Data = make([]oapi.Integration, len(pcs))
	for i, pc := range pcs {
		resp.Body.Data[i] = oapi.IntegrationFromEnt(pc)
	}
	resp.Body.Pagination = oapi.ResponsePagination{
		Total: len(pcs),
	}

	return &resp, nil
}

func (h *integrationsHandler) CreateIntegration(ctx context.Context, req *oapi.CreateIntegrationRequest) (*oapi.CreateIntegrationResponse, error) {
	var resp oapi.CreateIntegrationResponse

	attr := req.Body.Attributes

	pt, validErr := toValidProviderType(attr.Kind)
	if validErr != nil {
		return nil, apiError("invalid provider type", validErr)
	}

	cfg, cfgErr := json.Marshal(attr.Config)
	if cfgErr != nil {
		return nil, apiError("failed to marshal provider config", cfgErr)
	}

	pc := ent.ProviderConfig{
		ProviderType: pt,
		ProviderID:   attr.ProviderId,
		Enabled:      attr.Enabled,
		Config:       cfg,
	}

	created, createErr := h.configs.UpdateProviderConfig(ctx, pc)
	if createErr != nil {
		return nil, apiError("failed to update provider config", createErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(created)

	return &resp, nil
}

func (h *integrationsHandler) GetIntegration(ctx context.Context, req *oapi.GetIntegrationRequest) (*oapi.GetIntegrationResponse, error) {
	var resp oapi.GetIntegrationResponse

	pc, getErr := h.configs.GetProviderConfig(ctx, req.Id)
	if getErr != nil {
		return nil, apiError("failed to get provider config", getErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(pc)

	return &resp, nil
}

func (h *integrationsHandler) UpdateIntegration(ctx context.Context, req *oapi.UpdateIntegrationRequest) (*oapi.UpdateIntegrationResponse, error) {
	var resp oapi.UpdateIntegrationResponse

	pc, currErr := h.configs.GetProviderConfig(ctx, req.Id)
	if currErr != nil {
		return nil, apiError("failed to get provider config", currErr)
	}

	attr := req.Body.Attributes
	if attr.Config != nil {
		cfg, cfgErr := json.Marshal(attr.Config)
		if cfgErr != nil {
			return nil, apiError("invalid config", cfgErr)
		}
		pc.Config = cfg
	}
	if attr.Enabled != nil {
		pc.Enabled = *attr.Enabled
	}

	created, createErr := h.configs.UpdateProviderConfig(ctx, *pc)
	if createErr != nil {
		return nil, apiError("failed to update provider config", createErr)
	}
	resp.Body.Data = oapi.IntegrationFromEnt(created)

	return &resp, nil
}

func (h *integrationsHandler) DeleteIntegration(ctx context.Context, req *oapi.DeleteIntegrationRequest) (*oapi.DeleteIntegrationResponse, error) {
	var resp oapi.DeleteIntegrationResponse

	if delErr := h.configs.DeleteProviderConfig(ctx, req.Id); delErr != nil {
		return nil, apiError("failed to delete provider config", delErr)
	}

	return &resp, nil
}
