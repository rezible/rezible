package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"

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

	RequestIntegrationDataSync(context.Context, *RequestIntegrationDataSyncRequest) (*RequestIntegrationDataSyncResponse, error)
	GetIntegrationDataSyncStatus(context.Context, *GetIntegrationDataSyncStatusRequest) (*GetIntegrationDataSyncStatusResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListAvailableIntegrations, o.ListAvailableIntegrations)

	huma.Register(api, ConfigureIntegration, o.ConfigureIntegration)
	huma.Register(api, StartIntegrationOAuthFlow, o.StartIntegrationOAuthFlow)
	huma.Register(api, CompleteIntegrationOAuthFlow, o.CompleteIntegrationOAuthFlow)
	huma.Register(api, SelectIntegrationOAuthFlow, o.SelectIntegrationOAuthFlow)

	huma.Register(api, ListConfiguredIntegrations, o.ListConfiguredIntegrations)
	huma.Register(api, UpdateConfiguredIntegrationPreferences, o.UpdateConfiguredIntegrationPreferences)
	huma.Register(api, GetConfiguredIntegration, o.GetConfiguredIntegration)
	huma.Register(api, DeleteConfiguredIntegration, o.DeleteConfiguredIntegration)

	huma.Register(api, RequestIntegrationDataSync, o.RequestIntegrationDataSync)
	huma.Register(api, GetIntegrationDataSyncStatus, o.GetIntegrationDataSyncStatus)
}

type (
	AvailableIntegration struct {
		Name          string   `json:"name"`
		DataKinds     []string `json:"dataKinds"`
		OAuthRequired bool     `json:"oauthRequired"`
	}

	ConfiguredIntegration struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes ConfiguredIntegrationAttributes `json:"attributes"`
	}

	ConfiguredIntegrationAttributes struct {
		Provider        string          `json:"provider"`
		DisplayName     string          `json:"displayName"`
		ExternalRef     string          `json:"externalRef"`
		Config          map[string]any  `json:"config"`
		UserPreferences map[string]any  `json:"preferences"`
		DataKinds       map[string]bool `json:"dataKinds"`
	}

	IntegrationOAuthFlowResult struct {
		Status         string                      `json:"status"`
		Configured     []ConfiguredIntegration     `json:"configured"`
		SelectionToken string                      `json:"selectionToken,omitempty"`
		Options        []ExternalIntegrationOption `json:"options"`
	}

	ExternalIntegrationOption struct {
		ExternalRef string         `json:"externalRef"`
		DisplayName string         `json:"displayName"`
		Config      map[string]any `json:"config"`
	}

	IntegrationOAuthFlow struct {
		FlowUrl string `json:"flow_url"`
	}

	IntegrationProviderDataSyncStatus struct {
		Id         uuid.UUID                                   `json:"id"`
		Attributes IntegrationProviderDataSyncStatusAttributes `json:"attributes"`
	}

	IntegrationProviderDataSyncStatusAttributes struct {
		Status string `json:"status" enum:"queued,started,complete,error"`
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
		Provider:        cfg.Provider(),
		DisplayName:     cfg.DisplayName(),
		ExternalRef:     cfg.ExternalRef(),
		Config:          cfg.GetSanitizedConfig(),
		UserPreferences: cfg.GetUserPreferences(),
		DataKinds:       cfg.GetDataKinds(),
	}

	return ConfiguredIntegration{Id: cfg.ID(), Attributes: attrs}
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

func IntegrationProviderDataSyncStatusFromEnt(res *ent.ProviderEventSyncRun) IntegrationProviderDataSyncStatus {
	attrs := IntegrationProviderDataSyncStatusAttributes{
		Status: res.Status.String(),
	}
	return IntegrationProviderDataSyncStatus{Id: res.ID, Attributes: attrs}
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
type ListAvailableIntegrationsResponse PaginatedResponse[AvailableIntegration]

var ConfigureIntegration = huma.Operation{
	OperationID: "configure-integration",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/configured",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ConfigureIntegrationRequestAttributes struct {
	DisplayName string         `json:"displayName"`
	ExternalRef string         `json:"externalRef"`
	Config      map[string]any `json:"config"`
}
type ConfigureIntegrationRequest NameRequest[ConfigureIntegrationRequestAttributes]
type ConfigureIntegrationResponse ItemResponse[ConfiguredIntegration]

var StartIntegrationOAuthFlow = huma.Operation{
	OperationID: "start-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/oauth/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type StartOAuthFlowRequestAttributes struct {
	CallbackPath string `json:"callbackPath"`
}
type StartIntegrationOAuthFlowRequest NameRequest[StartOAuthFlowRequestAttributes]
type StartIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuthFlow = huma.Operation{
	OperationID: "complete-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/oauth/complete",
	Summary:     "Complete OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type CompleteIntegrationOAuthFlowRequestAttributes struct {
	Code           string  `json:"code"`
	State          *string `json:"state,omitempty"`
	ClientVerifier *string `json:"client_verifier,omitempty"`
}
type CompleteIntegrationOAuthFlowRequest NameRequest[CompleteIntegrationOAuthFlowRequestAttributes]
type CompleteIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlowResult]

var SelectIntegrationOAuthFlow = huma.Operation{
	OperationID: "select-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/oauth/select",
	Summary:     "Select OAuth installations for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type SelectIntegrationOAuthFlowRequestAttributes struct {
	SelectionToken string   `json:"selectionToken"`
	ExternalRefs   []string `json:"externalRefs"`
}
type SelectIntegrationOAuthFlowRequest NameRequest[SelectIntegrationOAuthFlowRequestAttributes]
type SelectIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlowResult]

var ListConfiguredIntegrations = huma.Operation{
	OperationID: "list-configured-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/configured",
	Summary:     "List Integrations",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListConfiguredIntegrationsRequest ListRequest
type ListConfiguredIntegrationsResponse PaginatedResponse[ConfiguredIntegration]

var GetConfiguredIntegration = huma.Operation{
	OperationID: "get-configured-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/configured/{id}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetConfiguredIntegrationRequest EmptyIdRequest
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
type UpdateConfiguredIntegrationPreferencesRequest IdRequest[UpdateConfiguredIntegrationPreferencesRequestAttributes]
type UpdateConfiguredIntegrationPreferencesResponse ItemResponse[ConfiguredIntegration]

var DeleteConfiguredIntegration = huma.Operation{
	OperationID: "delete-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/configured/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type DeleteConfiguredIntegrationRequest EmptyIdRequest
type DeleteConfiguredIntegrationResponse EmptyResponse

var RequestIntegrationDataSync = huma.Operation{
	OperationID: "request-integration-data-sync",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/sync",
	Summary:     "Request a manual data sync for an integration provider",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type RequestIntegrationDataSyncRequestAttributes struct {
	Sources []string `json:"sources,omitempty"`
}
type RequestIntegrationDataSyncRequest NameRequest[RequestIntegrationDataSyncRequestAttributes]
type RequestIntegrationDataSyncResponse EmptyResponse

var GetIntegrationDataSyncStatus = huma.Operation{
	OperationID: "get-integration-data-sync-status",
	Method:      http.MethodGet,
	Path:        "/integrations/providers/{name}/sync",
	Summary:     "Get data sync status for an integration provider",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetIntegrationDataSyncStatusRequest EmptyNameRequest
type GetIntegrationDataSyncStatusResponse PaginatedResponse[IntegrationProviderDataSyncStatus]
