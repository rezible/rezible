package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"

	rez "github.com/rezible/rezible"
)

type IntegrationsHandler interface {
	ListAvailableIntegrations(context.Context, *ListAvailableIntegrationsRequest) (*ListAvailableIntegrationsResponse, error)

	CreateInstalledIntegration(context.Context, *CreateInstalledIntegrationRequest) (*CreateInstalledIntegrationResponse, error)
	ListIntegrationInstallTargets(context.Context, *ListIntegrationInstallTargetsRequest) (*ListIntegrationInstallTargetsResponse, error)
	InstallIntegrationTargets(context.Context, *InstallIntegrationTargetsRequest) (*InstallIntegrationTargetsResponse, error)

	ListInstalledIntegrations(context.Context, *ListInstalledIntegrationsRequest) (*ListInstalledIntegrationsResponse, error)
	UpdateInstalledIntegration(context.Context, *UpdateInstalledIntegrationRequest) (*UpdateInstalledIntegrationResponse, error)
	GetInstalledIntegration(context.Context, *GetInstalledIntegrationRequest) (*GetInstalledIntegrationResponse, error)
	DeleteInstalledIntegration(context.Context, *DeleteInstalledIntegrationRequest) (*DeleteInstalledIntegrationResponse, error)

	StartIntegrationOAuthFlow(context.Context, *StartIntegrationOAuthFlowRequest) (*StartIntegrationOAuthFlowResponse, error)
	CompleteIntegrationOAuthFlow(context.Context, *CompleteIntegrationOAuthFlowRequest) (*CompleteIntegrationOAuthFlowResponse, error)

	RequestIntegrationEventSync(context.Context, *RequestIntegrationEventSyncRequest) (*RequestIntegrationEventSyncResponse, error)
	ListIntegrationEventSyncRun(context.Context, *ListIntegrationEventSyncRunRequest) (*ListIntegrationEventSyncRunResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListAvailableIntegrations, o.ListAvailableIntegrations)

	huma.Register(api, CreateInstalledIntegration, o.CreateInstalledIntegration)
	huma.Register(api, InstallIntegrationTargets, o.InstallIntegrationTargets)

	huma.Register(api, StartIntegrationOAuthFlow, o.StartIntegrationOAuthFlow)
	huma.Register(api, ListIntegrationInstallTargets, o.ListIntegrationInstallTargets)
	huma.Register(api, CompleteIntegrationOAuthFlow, o.CompleteIntegrationOAuthFlow)

	huma.Register(api, GetInstalledIntegration, o.GetInstalledIntegration)
	huma.Register(api, ListInstalledIntegrations, o.ListInstalledIntegrations)
	huma.Register(api, UpdateInstalledIntegration, o.UpdateInstalledIntegration)
	huma.Register(api, DeleteInstalledIntegration, o.DeleteInstalledIntegration)

	huma.Register(api, RequestIntegrationEventSync, o.RequestIntegrationEventSync)
	huma.Register(api, ListIntegrationEventSyncRuns, o.ListIntegrationEventSyncRun)
}

type (
	AvailableIntegration struct {
		Name                  string   `json:"name"`
		DisplayName           string   `json:"displayName"`
		Description           string   `json:"description"`
		Provider              string   `json:"provider"`
		SupportedCapabilities []string `json:"supportedCapabilities"`
		MaxInstalls           *int     `json:"maxInstalls,omitempty"`
		OAuthInstall          bool     `json:"oauthInstall"`
	}

	InstalledIntegration struct {
		Id         uuid.UUID                      `json:"id"`
		Attributes InstalledIntegrationAttributes `json:"attributes"`
	}

	InstalledIntegrationAttributes struct {
		IntegrationName string          `json:"integrationName"`
		ProviderName    string          `json:"providerName"`
		DisplayName     string          `json:"displayName"`
		ExternalRef     string          `json:"externalRef"`
		Config          map[string]any  `json:"config"`
		Settings        map[string]any  `json:"settings"`
		Capabilities    map[string]bool `json:"capabilities"`
	}

	IntegrationOAuthInstallResult struct {
		TargetSelectionRequired bool                       `json:"targetSelectionRequired"`
		Installed               []InstalledIntegration     `json:"installed,omitempty"`
		InstallTargetOptions    []IntegrationInstallTarget `json:"installTargetOptions,omitempty"`
	}

	IntegrationInstallTarget struct {
		IntegrationName string `json:"integrationName"`
		ExternalRef     string `json:"externalRef"`
		DisplayName     string `json:"displayName"`
		//Config      map[string]any `json:"config"`
	}

	IntegrationOAuthFlow struct {
		FlowUrl string `json:"flow_url"`
	}

	IntegrationEventSyncRun struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes IntegrationEventSyncRunAttributes `json:"attributes"`
	}

	IntegrationEventSyncRunAttributes struct {
		Status     string     `json:"status" enum:"queued,started,complete,error"`
		StartedAt  time.Time  `json:"startedAt"`
		FinishedAt *time.Time `json:"finishedAt"`
	}
)

func AvailableIntegrationFromPackage(p rez.IntegrationPackage) AvailableIntegration {
	return AvailableIntegration{
		Name:                  p.Name(),
		DisplayName:           p.DisplayName(),
		Description:           p.Description(),
		Provider:              p.Provider(),
		SupportedCapabilities: p.SupportedCapabilities(),
		OAuthInstall:          p.OAuthInstallRequired(),
		MaxInstalls:           p.MaxInstalls(),
	}
}

func InstalledIntegrationFromRez(ii rez.InstalledIntegration) InstalledIntegration {
	intg := ii.Integration()
	attrs := InstalledIntegrationAttributes{
		IntegrationName: intg.IntegrationName,
		ProviderName:    ii.ProviderName(),
		DisplayName:     intg.DisplayName,
		ExternalRef:     intg.ExternalProviderRef,
		Settings:        intg.UserSettings,
		Config:          ii.SanitizedInstallationConfig(),
		Capabilities:    ii.GetCapabilities(),
	}
	return InstalledIntegration{Id: intg.ID, Attributes: attrs}
}

func IntegrationInstallTargetOptionsFromRez(intgName string, targets []rez.IntegrationInstallationTarget) []IntegrationInstallTarget {
	res := make([]IntegrationInstallTarget, len(targets))
	for i, t := range targets {
		res[i] = IntegrationInstallTarget{
			IntegrationName: intgName,
			ExternalRef:     t.ExternalRef,
			DisplayName:     t.DisplayName,
		}
	}
	return res
}

func IntegrationOAuthFlowResultFromRez(intgName string, result *rez.CompleteIntegrationOAuth2FlowResult) IntegrationOAuthInstallResult {
	res := IntegrationOAuthInstallResult{
		TargetSelectionRequired: result.InstallationTargetSelectionRequired,
	}
	if res.TargetSelectionRequired {
		res.InstallTargetOptions = IntegrationInstallTargetOptionsFromRez(intgName, result.InstallationTargetOptions)
	} else {
		res.Installed = make([]InstalledIntegration, len(result.Installed))
		for i, ci := range result.Installed {
			res.Installed[i] = InstalledIntegrationFromRez(ci)
		}
	}
	return res
}

func IntegrationEventSyncRunFromEnt(r *ent.IntegrationEventSyncRun) IntegrationEventSyncRun {
	attrs := IntegrationEventSyncRunAttributes{
		Status:     r.Status.String(),
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
	}
	return IntegrationEventSyncRun{Id: r.ID, Attributes: attrs}
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

var CreateInstalledIntegration = huma.Operation{
	OperationID: "create-installed-integration",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/install",
	Summary:     "Install an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type CreateInstalledIntegrationRequestAttributes struct {
	DisplayName string         `json:"displayName,omitempty"`
	Config      map[string]any `json:"config"`
	Preferences map[string]any `json:"preferences"`
}
type CreateInstalledIntegrationRequest NameRequest[CreateInstalledIntegrationRequestAttributes]
type CreateInstalledIntegrationResponse ItemResponse[InstalledIntegration]

var ListIntegrationInstallTargets = huma.Operation{
	OperationID: "list-integration-install-targets",
	Method:      http.MethodGet,
	Path:        "/integrations/install_targets",
	Summary:     "List current unselected integration installation targets",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListIntegrationInstallTargetsRequest EmptyRequest
type ListIntegrationInstallTargetsResponse ListResponse[IntegrationInstallTarget]

var InstallIntegrationTargets = huma.Operation{
	OperationID: "install-integration-targets",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/install/targets",
	Summary:     "Select installation targets for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type InstallIntegrationTargetsRequestAttributes struct {
	ExternalRefs []string `json:"externalRefs"`
}
type InstallIntegrationTargetsRequest NameRequest[InstallIntegrationTargetsRequestAttributes]
type InstallIntegrationTargetsResponse ListResponse[InstalledIntegration]

var StartIntegrationOAuthFlow = huma.Operation{
	OperationID: "start-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/providers/{name}/oauth/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type StartIntegrationOAuthFlowRequest EmptyNameRequest
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
type CompleteIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthInstallResult]

var ListInstalledIntegrations = huma.Operation{
	OperationID: "list-installed-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/installed",
	Summary:     "List Installed Integrations",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListInstalledIntegrationsRequest ListRequest
type ListInstalledIntegrationsResponse ListResponse[InstalledIntegration]

var GetInstalledIntegration = huma.Operation{
	OperationID: "get-installed-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/installed/{id}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetInstalledIntegrationRequest EmptyIdRequest
type GetInstalledIntegrationResponse ItemResponse[InstalledIntegration]

var UpdateInstalledIntegration = huma.Operation{
	OperationID: "update-installed-integration",
	Method:      http.MethodPatch,
	Path:        "/integrations/installed/{id}",
	Summary:     "Update an installed Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type UpdateInstalledIntegrationRequestAttributes struct {
	Preferences map[string]any `json:"preferences"`
}
type UpdateInstalledIntegrationRequest IdRequest[UpdateInstalledIntegrationRequestAttributes]
type UpdateInstalledIntegrationResponse ItemResponse[InstalledIntegration]

var DeleteInstalledIntegration = huma.Operation{
	OperationID: "delete-installed-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/installed/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type DeleteInstalledIntegrationRequest EmptyIdRequest
type DeleteInstalledIntegrationResponse EmptyResponse

var RequestIntegrationEventSync = huma.Operation{
	OperationID: "request-integration-event-sync",
	Method:      http.MethodPost,
	Path:        "/integrations/installed/{id}/sync",
	Summary:     "Request a manual event sync for an integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type RequestIntegrationEventSyncRequestAttributes struct {
	Sources []string `json:"sources,omitempty"`
}
type RequestIntegrationEventSyncRequest IdRequest[RequestIntegrationEventSyncRequestAttributes]
type RequestIntegrationEventSyncResponse EmptyResponse

var ListIntegrationEventSyncRuns = huma.Operation{
	OperationID: "list-integration-event-sync-runs",
	Method:      http.MethodGet,
	Path:        "/integrations/installed/{id}/sync",
	Summary:     "Get event sync runs for an integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListIntegrationEventSyncRunRequest EmptyIdRequest
type ListIntegrationEventSyncRunResponse ListResponse[IntegrationEventSyncRun]
