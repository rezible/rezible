package apiv1

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providerconfig"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type integrationsHandler struct {
	configs rez.ProviderConfigService
	chat    rez.ChatService
}

func newIntegrationsHandler(configs rez.ProviderConfigService, chat rez.ChatService) *integrationsHandler {
	return &integrationsHandler{configs: configs, chat: chat}
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

func (h *integrationsHandler) makeOAuthState(ctx context.Context, providerId string) (string, error) {
	// TODO
	return "TODO", nil
}

func (h *integrationsHandler) verifyOAuthState(ctx context.Context, providerId string, state string) error {
	// TODO
	return nil
}

func (h *integrationsHandler) StartIntegrationOAuth(ctx context.Context, req *oapi.StartIntegrationOAuthRequest) (*oapi.StartIntegrationOAuthResponse, error) {
	var resp oapi.StartIntegrationOAuthResponse

	attr := req.Body.Attributes
	pt, ptErr := toValidProviderType(attr.Kind)
	if ptErr != nil {
		return nil, apiError("invalid provider kind", ptErr)
	}

	if pt == providerconfig.ProviderTypeChat {
		state, stateErr := h.makeOAuthState(ctx, attr.ProviderId)
		if stateErr != nil {
			return nil, fmt.Errorf("failed to make oauth state: %w", stateErr)
		}

		flowUrl, urlErr := h.chat.GetOAuth2URL(ctx, state)
		if urlErr != nil {
			return nil, fmt.Errorf("failed to make oauth flow url: %w", urlErr)
		}

		resp.Body.Data = oapi.IntegrationOAuthFlow{FlowUrl: flowUrl}
	} else {
		return nil, oapi.ErrorBadRequest("invalid provider type")
	}

	return &resp, nil
}

func (h *integrationsHandler) CompleteIntegrationOAuth(ctx context.Context, req *oapi.CompleteIntegrationOAuthRequest) (*oapi.CompleteIntegrationOAuthResponse, error) {
	var resp oapi.CompleteIntegrationOAuthResponse

	attr := req.Body.Attributes
	pt, ptErr := toValidProviderType(attr.Kind)
	if ptErr != nil {
		return nil, apiError("invalid provider kind", ptErr)
	}

	if pt == providerconfig.ProviderTypeChat {
		if stateErr := h.verifyOAuthState(ctx, attr.ProviderId, attr.State); stateErr != nil {
			return nil, oapi.ErrorForbidden("invalid state", stateErr)
		}

		cfg, cfgErr := h.chat.CompleteOAuth2Flow(ctx, attr.Code)
		if cfgErr != nil {
			return nil, oapi.ErrorBadRequest("invalid code", cfgErr)
		}

		pc, pcErr := h.configs.UpdateProviderConfig(ctx, *cfg)
		if pcErr != nil {
			return nil, apiError("failed to update provider config", pcErr)
		}
		resp.Body.Data = oapi.IntegrationFromEnt(pc)
	} else {
		return nil, oapi.ErrorBadRequest("invalid provider type")
	}

	return &resp, nil
}
