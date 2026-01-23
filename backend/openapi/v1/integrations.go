package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
)

type IntegrationsHandler interface {
	ListSupported(context.Context, *ListSupportedIntegrationsRequest) (*ListSupportedIntegrationsResponse, error)

	ListConfigured(context.Context, *ListConfiguredIntegrationsRequest) (*ListConfiguredIntegrationsResponse, error)
	ConfigureIntegration(context.Context, *ConfigureIntegrationRequest) (*ConfigureIntegrationResponse, error)
	GetIntegration(context.Context, *GetIntegrationRequest) (*GetIntegrationResponse, error)
	DeleteIntegration(context.Context, *DeleteIntegrationRequest) (*DeleteIntegrationResponse, error)

	StartIntegrationOAuthFlow(context.Context, *StartIntegrationOAuthFlowRequest) (*StartIntegrationOAuthFlowResponse, error)
	CompleteIntegrationOAuthFlow(context.Context, *CompleteIntegrationOAuthFlowRequest) (*CompleteIntegrationOAuthFlowResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListSupportedIntegrations, o.ListSupported)

	huma.Register(api, ListConfiguredIntegrations, o.ListConfigured)
	huma.Register(api, ConfigureIntegration, o.ConfigureIntegration)
	huma.Register(api, GetIntegration, o.GetIntegration)
	huma.Register(api, DeleteIntegration, o.DeleteIntegration)

	huma.Register(api, StartIntegrationOAuthFlow, o.StartIntegrationOAuthFlow)
	huma.Register(api, CompleteIntegrationOAuthFlow, o.CompleteIntegrationOAuthFlow)
}

type (
	SupportedIntegration struct {
		Name          string   `json:"name"`
		DataKinds     []string `json:"supportedDataKinds"`
		OAuthRequired bool     `json:"oauthRequired"`
	}

	ConfiguredIntegration struct {
		Name       string                          `json:"name"`
		Attributes ConfiguredIntegrationAttributes `json:"attributes"`
	}

	ConfiguredIntegrationAttributes struct {
		Config      json.RawMessage `json:"config"`
		ConfigValid bool            `json:"configValid"`
		DataKinds   map[string]bool `json:"dataKinds"`
	}

	IntegrationOAuthFlow struct {
		FlowUrl string `json:"flow_url"`
	}
)

func IntegrationFromEnt(intg *ent.Integration) ConfiguredIntegration {
	configValid, _ := integrations.ValidateConfig(intg.Name, intg.Config)
	attrs := ConfiguredIntegrationAttributes{
		Config:      intg.Config,
		ConfigValid: configValid,
	}

	return ConfiguredIntegration{
		Name:       intg.Name,
		Attributes: attrs,
	}
}

var integrationsTags = []string{"Integrations"}

type (
	NamedIntegrationRequest struct {
		Name string `path:"name"`
	}

	NamedIntegrationRequestWithAttributes[A any] struct {
		NamedIntegrationRequest
		RequestWithBodyAttributes[A]
	}

	RawIntegrationConfigRequestAttributes struct {
		Config json.RawMessage `json:"config"`
	}
	NamedIntegrationRawConfigRequest NamedIntegrationRequestWithAttributes[RawIntegrationConfigRequestAttributes]
)

var ListSupportedIntegrations = huma.Operation{
	OperationID: "list-supported-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/supported",
	Summary:     "List Supported Integrations",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ListSupportedIntegrationsRequest ListRequest
type ListSupportedIntegrationsResponse PaginatedResponse[SupportedIntegration]

var ListConfiguredIntegrations = huma.Operation{
	OperationID: "list-configured-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/configured",
	Summary:     "List Integrations",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ListConfiguredIntegrationsRequest ListRequest
type ListConfiguredIntegrationsResponse PaginatedResponse[ConfiguredIntegration]

var GetIntegration = huma.Operation{
	OperationID: "get-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/configured/{name}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type GetIntegrationRequest NamedIntegrationRequest
type GetIntegrationResponse ItemResponse[ConfiguredIntegration]

var ConfigureIntegration = huma.Operation{
	OperationID: "configure-integration",
	Method:      http.MethodPost,
	Path:        "/integrations/configured/{name}",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ConfigureIntegrationRequest NamedIntegrationRawConfigRequest
type ConfigureIntegrationResponse ItemResponse[ConfiguredIntegration]

var DeleteIntegration = huma.Operation{
	OperationID: "delete-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/configured/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type DeleteIntegrationRequest NamedIntegrationRequest
type DeleteIntegrationResponse EmptyResponse

var StartIntegrationOAuthFlow = huma.Operation{
	OperationID: "start-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/oauth/{name}/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type StartIntegrationOAuthFlowRequest NamedIntegrationRequest
type StartIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuthFlow = huma.Operation{
	OperationID: "complete-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/oauth/{name}/complete",
	Summary:     "Complete OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type CompleteIntegrationOAuthFlowRequestAttributes struct {
	State string `json:"state"`
	Code  string `json:"code"`
}
type CompleteIntegrationOAuthFlowRequest NamedIntegrationRequestWithAttributes[CompleteIntegrationOAuthFlowRequestAttributes]
type CompleteIntegrationOAuthFlowResponse ItemResponse[ConfiguredIntegration]
