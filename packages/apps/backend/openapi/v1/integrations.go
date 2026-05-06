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
	SelectIntegrationOAuthFlow(context.Context, *SelectIntegrationOAuthFlowRequest) (*SelectIntegrationOAuthFlowResponse, error)
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
	huma.Register(api, SelectIntegrationOAuthFlow, o.SelectIntegrationOAuthFlow)
}

type (
	AvailableIntegration struct {
		Name          string   `json:"name"`
		DataKinds     []string `json:"dataKinds" nullable:"false"`
		OAuthRequired bool     `json:"oauthRequired"`
	}

	ConfiguredIntegration struct {
		Id         string                          `json:"id"`
		Provider   string                          `json:"provider"`
		Attributes ConfiguredIntegrationAttributes `json:"attributes"`
	}

	ExternalIntegrationOption struct {
		ExternalRef string         `json:"externalRef"`
		DisplayName string         `json:"displayName"`
		Config      map[string]any `json:"config"`
	}

	IntegrationOAuthFlowResult struct {
		Status         string                      `json:"status"`
		Configured     []ConfiguredIntegration     `json:"configured" nullable:"false"`
		SelectionToken string                      `json:"selectionToken,omitempty"`
		Options        []ExternalIntegrationOption `json:"options" nullable:"false"`
	}

	ConfiguredIntegrationAttributes struct {
		DisplayName     string          `json:"displayName"`
		ExternalRef     string          `json:"externalRef"`
		Config          map[string]any  `json:"config"`
		UserPreferences map[string]any  `json:"preferences"`
		DataKinds       map[string]bool `json:"dataKinds"`
	}

	IntegrationOAuthFlow struct {
		FlowUrl string `json:"flow_url"`
	}
)

type (
	IntegrationProviderRequest struct {
		Provider string `path:"provider"`
	}
	IntegrationProviderRequestWithAttributes[A any] struct {
		IntegrationProviderRequest
		RequestWithBodyAttributes[A]
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
		DisplayName:     cfg.DisplayName(),
		ExternalRef:     cfg.ExternalRef(),
		Config:          cfg.GetSanitizedConfig(),
		UserPreferences: cfg.GetUserPreferences(),
		DataKinds:       cfg.GetDataKinds(),
	}

	return ConfiguredIntegration{Id: cfg.ID().String(), Provider: cfg.Provider(), Attributes: attrs}
}

func ExternalIntegrationOptionFromCore(opt rez.ExternalIntegrationOption) ExternalIntegrationOption {
	return ExternalIntegrationOption{
		ExternalRef: opt.ExternalRef,
		DisplayName: opt.DisplayName,
		Config:      opt.Config,
	}
}

func IntegrationOAuthFlowResultFromCore(result *rez.CompleteIntegrationOAuth2Result) IntegrationOAuthFlowResult {
	configured := make([]ConfiguredIntegration, len(result.Configured))
	for i, ci := range result.Configured {
		configured[i] = ConfiguredIntegrationFromConfig(ci)
	}
	options := make([]ExternalIntegrationOption, len(result.Options))
	for i, opt := range result.Options {
		options[i] = ExternalIntegrationOptionFromCore(opt)
	}
	return IntegrationOAuthFlowResult{
		Status:         result.Status,
		Configured:     configured,
		SelectionToken: result.SelectionToken,
		Options:        options,
	}
}

var integrationsTags = []string{"Integrations"}

var ListAvailableIntegrations = huma.Operation{
	OperationID: "list-available-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/providers",
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
	Path:        "/integrations/providers/{provider}/configured",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ConfigureIntegrationRequestAttributes struct {
	DisplayName string         `json:"displayName"`
	ExternalRef string         `json:"externalRef"`
	Config      map[string]any `json:"config"`
}
type ConfigureIntegrationRequest IntegrationProviderRequestWithAttributes[ConfigureIntegrationRequestAttributes]
type ConfigureIntegrationResponse ItemResponse[ConfiguredIntegration]

var GetConfiguredIntegration = huma.Operation{
	OperationID: "get-configured-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/configured/{id}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetConfiguredIntegrationRequest GetIdRequest
type GetConfiguredIntegrationResponse ItemResponse[ConfiguredIntegration]

var UpdateConfiguredIntegrationPreferences = huma.Operation{
	OperationID: "update-configured-integration-preferences",
	Method:      http.MethodPost,
	Path:        "/integrations/configured/{id}/preferences",
	Summary:     "Update Preferences for a Configured Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type UpdateConfiguredIntegrationPreferencesRequestAttributes struct {
	Preferences map[string]any `json:"preferences"`
}
type UpdateConfiguredIntegrationPreferencesRequest UpdateIdRequest[UpdateConfiguredIntegrationPreferencesRequestAttributes]
type UpdateConfiguredIntegrationPreferencesResponse ItemResponse[ConfiguredIntegration]

var DeleteConfiguredIntegration = huma.Operation{
	OperationID: "delete-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/configured/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type DeleteConfiguredIntegrationRequest DeleteIdRequest
type DeleteConfiguredIntegrationResponse EmptyResponse

var StartIntegrationOAuthFlow = huma.Operation{
	OperationID: "start-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{provider}/oauth/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type StartOAuthFlowRequestAttributes struct {
	CallbackPath string `json:"callbackPath"`
}
type StartIntegrationOAuthFlowRequest IntegrationProviderRequestWithAttributes[StartOAuthFlowRequestAttributes]
type StartIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuthFlow = huma.Operation{
	OperationID: "complete-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{provider}/oauth/complete",
	Summary:     "Complete OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type CompleteIntegrationOAuthFlowRequestAttributes struct {
	Code           string  `json:"code"`
	State          *string `json:"state,omitempty"`
	ClientVerifier *string `json:"client_verifier,omitempty"`
}
type CompleteIntegrationOAuthFlowRequest IntegrationProviderRequestWithAttributes[CompleteIntegrationOAuthFlowRequestAttributes]
type CompleteIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlowResult]

var SelectIntegrationOAuthFlow = huma.Operation{
	OperationID: "select-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{provider}/oauth/select",
	Summary:     "Select OAuth installations for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type SelectIntegrationOAuthFlowRequestAttributes struct {
	SelectionToken string   `json:"selectionToken"`
	ExternalRefs   []string `json:"externalRefs" nullable:"false"`
}
type SelectIntegrationOAuthFlowRequest IntegrationProviderRequestWithAttributes[SelectIntegrationOAuthFlowRequestAttributes]
type SelectIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlowResult]
