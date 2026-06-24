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
	GetInstallableIntegrations(context.Context, *GetInstallableIntegrationsRequest) (*GetInstallableIntegrationsResponse, error)

	InstallIntegration(context.Context, *InstallIntegrationRequest) (*InstallIntegrationResponse, error)

	StartIntegrationOAuthFlow(context.Context, *StartIntegrationOAuthFlowRequest) (*StartIntegrationOAuthFlowResponse, error)
	CompleteIntegrationOAuthFlow(context.Context, *CompleteIntegrationOAuthFlowRequest) (*CompleteIntegrationOAuthFlowResponse, error)

	ListIntegrationInstallTargets(context.Context, *ListIntegrationInstallTargetsRequest) (*ListIntegrationInstallTargetsResponse, error)
	InstallIntegrationFromTargets(context.Context, *InstallIntegrationFromTargetsRequest) (*InstallIntegrationFromTargetsResponse, error)

	ListIntegrationInstallations(context.Context, *ListIntegrationInstallationsRequest) (*ListIntegrationInstallationsResponse, error)
	GetIntegrationInstallation(context.Context, *GetIntegrationInstallationRequest) (*GetIntegrationInstallationResponse, error)
	UpdateIntegrationInstallation(context.Context, *UpdateIntegrationInstallationRequest) (*UpdateIntegrationInstallationResponse, error)
	DeleteIntegrationInstallation(context.Context, *DeleteIntegrationInstallationRequest) (*DeleteIntegrationInstallationResponse, error)

	RequestIntegrationEventSync(context.Context, *RequestIntegrationEventSyncRequest) (*RequestIntegrationEventSyncResponse, error)
	ListIntegrationEventSyncRun(context.Context, *ListIntegrationEventSyncRunRequest) (*ListIntegrationEventSyncRunResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, GetInstallableIntegrations, o.GetInstallableIntegrations)

	huma.Register(api, InstallIntegration, o.InstallIntegration)
	huma.Register(api, InstallIntegrationFromTargets, o.InstallIntegrationFromTargets)

	huma.Register(api, StartIntegrationOAuthFlow, o.StartIntegrationOAuthFlow)
	huma.Register(api, ListIntegrationInstallTargets, o.ListIntegrationInstallTargets)
	huma.Register(api, CompleteIntegrationOAuthFlow, o.CompleteIntegrationOAuthFlow)

	huma.Register(api, GetIntegrationInstallation, o.GetIntegrationInstallation)
	huma.Register(api, ListIntegrationInstallations, o.ListIntegrationInstallations)
	huma.Register(api, UpdateIntegrationInstallation, o.UpdateIntegrationInstallation)
	huma.Register(api, DeleteIntegrationInstallation, o.DeleteIntegrationInstallation)

	huma.Register(api, RequestIntegrationEventSync, o.RequestIntegrationEventSync)
	huma.Register(api, ListIntegrationEventSyncRuns, o.ListIntegrationEventSyncRun)
}

type (
	InstallableIntegration struct {
		Name         string `json:"name"`
		DisplayName  string `json:"displayName"`
		Description  string `json:"description"`
		Provider     string `json:"provider"`
		MaxInstalls  *int   `json:"maxInstalls,omitempty"`
		OAuthInstall bool   `json:"oauthInstall"`
	}

	IntegrationInstallation struct {
		Id         uuid.UUID                         `json:"id"`
		Attributes IntegrationInstallationAttributes `json:"attributes"`
	}

	IntegrationInstallationAttributes struct {
		IntegrationName string         `json:"integrationName"`
		ProviderName    string         `json:"providerName"`
		DisplayName     string         `json:"displayName"`
		ExternalRef     string         `json:"externalRef"`
		Config          map[string]any `json:"config"`
		Settings        map[string]any `json:"settings"`
	}

	IntegrationOAuthInstallResult struct {
		TargetSelectionRequired bool                       `json:"targetSelectionRequired"`
		Installed               []IntegrationInstallation  `json:"installed,omitempty"`
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

func InstallableIntegrationFromPackage(p rez.IntegrationPackage) InstallableIntegration {
	return InstallableIntegration{
		Name:         p.Name(),
		DisplayName:  p.DisplayName(),
		Description:  p.Description(),
		Provider:     p.Provider(),
		OAuthInstall: p.OAuthInstallRequired(),
		MaxInstalls:  p.MaxInstalls(),
	}
}

func IntegrationInstallationFromRez(ii rez.InstalledIntegration) IntegrationInstallation {
	intg := ii.Integration()
	attrs := IntegrationInstallationAttributes{
		IntegrationName: intg.IntegrationName,
		ProviderName:    ii.ProviderName(),
		DisplayName:     ii.DisplayName(),
		ExternalRef:     intg.ExternalProviderRef,
		Settings:        intg.UserSettings,
		Config:          ii.GetSanitizedConfig(),
	}
	return IntegrationInstallation{Id: intg.ID, Attributes: attrs}
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
		res.Installed = make([]IntegrationInstallation, len(result.Installed))
		for i, ci := range result.Installed {
			res.Installed[i] = IntegrationInstallationFromRez(ci)
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

var GetInstallableIntegrations = huma.Operation{
	OperationID: "get-installable-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations",
	Summary:     "Get Installable Integrations",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetInstallableIntegrationsRequest ListRequest
type GetInstallableIntegrationsResponse ItemResponse[[]InstallableIntegration]

var InstallIntegration = huma.Operation{
	OperationID: "install-integration",
	Method:      http.MethodPost,
	Path:        "/integrations/install/{name}",
	Summary:     "Install an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type InstallIntegrationRequestAttributes struct {
	Config       map[string]any `json:"config"`
	UserSettings map[string]any `json:"userSettings"`
}
type InstallIntegrationRequest NameRequest[InstallIntegrationRequestAttributes]
type InstallIntegrationResponse ItemResponse[IntegrationInstallation]

var ListIntegrationInstallTargets = huma.Operation{
	OperationID: "list-integration-install-targets",
	Method:      http.MethodGet,
	Path:        "/integrations/install_targets",
	Summary:     "List integration installation targets requiring selection",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListIntegrationInstallTargetsRequest EmptyRequest
type ListIntegrationInstallTargetsResponse ListResponse[IntegrationInstallTarget]

var InstallIntegrationFromTargets = huma.Operation{
	OperationID: "install-integration-from-targets",
	Method:      http.MethodPost,
	Path:        "/integrations/install/{name}/targets",
	Summary:     "Select installation targets for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type InstallIntegrationFromTargetsRequestAttributes struct {
	ExternalRefs []string `json:"externalRefs"`
}
type InstallIntegrationFromTargetsRequest NameRequest[InstallIntegrationFromTargetsRequestAttributes]
type InstallIntegrationFromTargetsResponse ListResponse[IntegrationInstallation]

var StartIntegrationOAuthFlow = huma.Operation{
	OperationID: "start-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/install/{name}/oauth/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type StartIntegrationOAuthFlowRequest EmptyNameRequest
type StartIntegrationOAuthFlowResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuthFlow = huma.Operation{
	OperationID: "complete-integration-oauth-flow",
	Method:      http.MethodPost,
	Path:        "/integrations/install/{name}/oauth/complete",
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

var ListIntegrationInstallations = huma.Operation{
	OperationID: "list-integration-installations",
	Method:      http.MethodGet,
	Path:        "/integrations/installations",
	Summary:     "List Installed Integrations",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListIntegrationInstallationsRequest ListRequest
type ListIntegrationInstallationsResponse ListResponse[IntegrationInstallation]

var GetIntegrationInstallation = huma.Operation{
	OperationID: "get-integration-installation",
	Method:      http.MethodGet,
	Path:        "/integrations/installations/{id}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type GetIntegrationInstallationRequest EmptyIdRequest
type GetIntegrationInstallationResponse ItemResponse[IntegrationInstallation]

var UpdateIntegrationInstallation = huma.Operation{
	OperationID: "update-integration-installation",
	Method:      http.MethodPatch,
	Path:        "/integrations/installations/{id}",
	Summary:     "Update an installed Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type UpdateIntegrationInstallationRequestAttributes struct {
	UserSettings map[string]any `json:"userSettings"`
}
type UpdateIntegrationInstallationRequest IdRequest[UpdateIntegrationInstallationRequestAttributes]
type UpdateIntegrationInstallationResponse ItemResponse[IntegrationInstallation]

var DeleteIntegrationInstallation = huma.Operation{
	OperationID: "delete-integration-installation",
	Method:      http.MethodDelete,
	Path:        "/integrations/installations/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type DeleteIntegrationInstallationRequest EmptyIdRequest
type DeleteIntegrationInstallationResponse EmptyResponse

var RequestIntegrationEventSync = huma.Operation{
	OperationID: "request-integration-event-sync",
	Method:      http.MethodPost,
	Path:        "/integrations/installations/{id}/sync/start",
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
	Path:        "/integrations/installations/{id}/sync",
	Summary:     "Get event sync runs for an integration",
	Tags:        integrationsTags,
	Errors:      ErrorCodes(),
}

type ListIntegrationEventSyncRunRequest EmptyIdRequest
type ListIntegrationEventSyncRunResponse ListResponse[IntegrationEventSyncRun]
