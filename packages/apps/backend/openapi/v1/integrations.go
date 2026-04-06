package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	rez "github.com/rezible/rezible"
)

type IntegrationsHandler interface {
	ListAvailableIntegrations(context.Context, *ListAvailableIntegrationsRequest) (*ListAvailableIntegrationsResponse, error)

	ListConfiguredIntegrations(context.Context, *ListConfiguredIntegrationsRequest) (*ListConfiguredIntegrationsResponse, error)
	ConfigureIntegration(context.Context, *ConfigureIntegrationRequest) (*ConfigureIntegrationResponse, error)
	UpdateConfiguredIntegrationPreferences(context.Context, *UpdateConfiguredIntegrationPreferencesRequest) (*UpdateConfiguredIntegrationPreferencesResponse, error)
	GetConfiguredIntegration(context.Context, *GetConfiguredIntegrationRequest) (*GetConfiguredIntegrationResponse, error)
	DeleteConfiguredIntegration(context.Context, *DeleteConfiguredIntegrationRequest) (*DeleteConfiguredIntegrationResponse, error)

	StartIntegrationOAuthFlow(context.Context, *StartIntegrationOAuthFlowRequest) (*StartIntegrationOAuthFlowResponse, error)
	CompleteIntegrationOAuthFlow(context.Context, *CompleteIntegrationOAuthFlowRequest) (*CompleteIntegrationOAuthFlowResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListAvailableIntegrations, o.ListAvailableIntegrations)

	huma.Register(api, ListConfiguredIntegrations, o.ListConfiguredIntegrations)
	huma.Register(api, ConfigureIntegration, o.ConfigureIntegration)
	huma.Register(api, UpdateConfiguredIntegrationPreferences, o.UpdateConfiguredIntegrationPreferences)
	huma.Register(api, GetConfiguredIntegration, o.GetConfiguredIntegration)
	huma.Register(api, DeleteConfiguredIntegration, o.DeleteConfiguredIntegration)

	huma.Register(api, StartIntegrationOAuthFlow, o.StartIntegrationOAuthFlow)
	huma.Register(api, CompleteIntegrationOAuthFlow, o.CompleteIntegrationOAuthFlow)
}

type (
	AvailableIntegration struct {
		Name          string   `json:"name"`
		DataKinds     []string `json:"dataKinds" nullable:"false"`
		OAuthRequired bool     `json:"oauthRequired"`
	}

	ConfiguredIntegration struct {
		Name       string                          `json:"name"`
		Attributes ConfiguredIntegrationAttributes `json:"attributes"`
	}

	ConfiguredIntegrationAttributes struct {
		Config          map[string]any  `json:"config"`
		UserPreferences map[string]any  `json:"preferences"`
		DataKinds       map[string]bool `json:"dataKinds"`
	}

	IntegrationOAuthFlow struct {
		FlowUrl string `json:"flow_url"`
	}
)

func AvailableIntegrationFromPackage(p rez.IntegrationPackage) AvailableIntegration {
	return AvailableIntegration{
		Name:          p.Name(),
		DataKinds:     p.SupportedDataKinds(),
		OAuthRequired: p.OAuthConfigRequired(),
	}
}

func ConfiguredIntegrationFromConfig(cfg rez.ConfiguredIntegration) ConfiguredIntegration {
	attrs := ConfiguredIntegrationAttributes{
		Config:          cfg.GetSanitizedConfig(),
		UserPreferences: cfg.GetUserPreferences(),
		DataKinds:       cfg.GetDataKinds(),
	}

	return ConfiguredIntegration{Name: cfg.Name(), Attributes: attrs}
}

var integrationsTags = []string{"Integrations"}

var ListAvailableIntegrations = huma.Operation{
	OperationID: "list-available-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations",
	Summary:     "List Available Integrations",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListAvailableIntegrationsRequest ListRequest
type ListAvailableIntegrationsResponse ListResponse[AvailableIntegration]

var ListConfiguredIntegrations = huma.Operation{
	OperationID: "list-configured-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/configured",
	Summary:     "List Integrations",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListConfiguredIntegrationsRequest ListRequest
type ListConfiguredIntegrationsResponse ListResponse[ConfiguredIntegration]

var ConfigureIntegration = huma.Operation{
	OperationID: "configure-integration",
	Method:      http.MethodPost,
	Path:        "/integrations/{name}",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ConfigureIntegrationRequestAttributes struct {
	Config map[string]any `json:"config"`
}
type ConfigureIntegrationRequest NameRequestWithAttributes[ConfigureIntegrationRequestAttributes]
type ConfigureIntegrationResponse ItemResponse[ConfiguredIntegration]

var GetConfiguredIntegration = huma.Operation{
	OperationID: "get-configured-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/configured/{name}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetConfiguredIntegrationRequest NameRequest
type GetConfiguredIntegrationResponse ItemResponse[ConfiguredIntegration]

var UpdateConfiguredIntegrationPreferences = huma.Operation{
	OperationID: "update-configured-integration-preferences",
	Method:      http.MethodPost,
	Path:        "/integrations/configured/{name}/preferences",
	Summary:     "Update Preferences for a Configured Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type UpdateConfiguredIntegrationPreferencesRequestAttributes struct {
	Preferences map[string]any `json:"preferences"`
}
type UpdateConfiguredIntegrationPreferencesRequest NameRequestWithAttributes[UpdateConfiguredIntegrationPreferencesRequestAttributes]
type UpdateConfiguredIntegrationPreferencesResponse ItemResponse[ConfiguredIntegration]

var DeleteConfiguredIntegration = huma.Operation{
	OperationID: "delete-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/configured/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type DeleteConfiguredIntegrationRequest NameRequest
type DeleteConfiguredIntegrationResponse EmptyResponse

var StartIntegrationOAuthFlow = huma.Operation{
	OperationID: "start-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/configured/{name}/start_oauth_flow",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type StartOAuthFlowRequestAttributes struct {
	CallbackPath string `json:"callbackPath"`
}
type StartIntegrationOAuthFlowRequest NameRequestWithAttributes[StartOAuthFlowRequestAttributes]
type StartIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuthFlow = huma.Operation{
	OperationID: "complete-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/configured/{name}/complete_oauth_flow",
	Summary:     "Complete OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type CompleteIntegrationOAuthFlowRequestAttributes struct {
	Code           string  `json:"code"`
	State          *string `json:"state,omitempty"`
	ClientVerifier *string `json:"client_verifier,omitempty"`
}
type CompleteIntegrationOAuthFlowRequest NameRequestWithAttributes[CompleteIntegrationOAuthFlowRequestAttributes]
type CompleteIntegrationOAuthFlowResponse ItemResponse[ConfiguredIntegration]
