package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/integrations"
)

type IntegrationsHandler interface {
	ListSupportedIntegrations(context.Context, *ListSupportedIntegrationsRequest) (*ListSupportedIntegrationsResponse, error)

	ListIntegrations(context.Context, *ListIntegrationsRequest) (*ListIntegrationsResponse, error)
	CreateIntegration(context.Context, *CreateIntegrationRequest) (*CreateIntegrationResponse, error)
	GetIntegration(context.Context, *GetIntegrationRequest) (*GetIntegrationResponse, error)
	UpdateIntegration(context.Context, *UpdateIntegrationRequest) (*UpdateIntegrationResponse, error)
	DeleteIntegration(context.Context, *DeleteIntegrationRequest) (*DeleteIntegrationResponse, error)

	StartIntegrationOAuth(context.Context, *StartIntegrationOAuthRequest) (*StartIntegrationOAuthResponse, error)
	CompleteIntegrationOAuth(context.Context, *CompleteIntegrationOAuthRequest) (*CompleteIntegrationOAuthResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListSupportedIntegrations, o.ListSupportedIntegrations)

	huma.Register(api, ListIntegrations, o.ListIntegrations)
	huma.Register(api, CreateIntegration, o.CreateIntegration)
	huma.Register(api, GetIntegration, o.GetIntegration)
	huma.Register(api, UpdateIntegration, o.UpdateIntegration)
	huma.Register(api, DeleteIntegration, o.DeleteIntegration)

	huma.Register(api, StartIntegrationOAuth, o.StartIntegrationOAuth)
	huma.Register(api, CompleteIntegrationOAuth, o.CompleteIntegrationOAuth)
}

type (
	SupportedIntegration struct {
		Name          string   `json:"name"`
		DataKinds     []string `json:"supportedDataKinds"`
		OAuthRequired bool     `json:"oauthRequired"`
	}

	Integration struct {
		Id         uuid.UUID             `json:"id"`
		Attributes IntegrationAttributes `json:"attributes"`
	}

	IntegrationAttributes struct {
		Name             string          `json:"name"`
		Config           json.RawMessage `json:"config"`
		ConfigValid      bool            `json:"configValid"`
		EnabledDataKinds []string        `json:"enabledDataKinds"`
	}
)

func IntegrationFromEnt(intg *ent.Integration) Integration {
	attrs := IntegrationAttributes{
		Name:             intg.Name,
		Config:           intg.Config,
		ConfigValid:      integrations.CheckConfigValid(intg),
		EnabledDataKinds: integrations.GetEnabledDataKinds(intg),
	}

	return Integration{
		Id:         intg.ID,
		Attributes: attrs,
	}
}

var integrationsTags = []string{"Integrations"}

var ListSupportedIntegrations = huma.Operation{
	OperationID: "list-supported-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations",
	Summary:     "List Supported Integrations",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ListSupportedIntegrationsRequest ListRequest
type ListSupportedIntegrationsResponse PaginatedResponse[SupportedIntegration]

var ListIntegrations = huma.Operation{
	OperationID: "list-configured-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations/configured",
	Summary:     "List Configured Integrations",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ListIntegrationsRequest ListRequest
type ListIntegrationsResponse PaginatedResponse[Integration]

var CreateIntegration = huma.Operation{
	OperationID: "create-integration",
	Method:      http.MethodPost,
	Path:        "/integrations/configured",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type CreateIntegrationRequestAttributes struct {
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}
type CreateIntegrationRequest RequestWithBodyAttributes[CreateIntegrationRequestAttributes]
type CreateIntegrationResponse ItemResponse[Integration]

var GetIntegration = huma.Operation{
	OperationID: "get-configured-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/configured/{id}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type GetIntegrationRequest GetIdRequest
type GetIntegrationResponse ItemResponse[Integration]

var UpdateIntegration = huma.Operation{
	OperationID: "update-configured-integration",
	Method:      http.MethodPatch,
	Path:        "/integrations/configured/{id}",
	Summary:     "Update an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type UpdateIntegrationAttributes struct {
	Enabled *bool            `json:"enabled,omitempty"`
	Config  *json.RawMessage `json:"config,omitempty"`
}
type UpdateIntegrationRequest UpdateIdRequest[UpdateIntegrationAttributes]
type UpdateIntegrationResponse ItemResponse[Integration]

var DeleteIntegration = huma.Operation{
	OperationID: "delete-configured-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/configured/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type DeleteIntegrationRequest DeleteIdRequest
type DeleteIntegrationResponse EmptyResponse

var StartIntegrationOAuth = huma.Operation{
	OperationID: "start-integration-oauth",
	Method:      http.MethodPost,
	Path:        "/integrations/oauth/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type StartIntegrationOAuthRequestAttributes struct {
	Name string `json:"name"`
}
type StartIntegrationOAuthRequest RequestWithBodyAttributes[StartIntegrationOAuthRequestAttributes]

type IntegrationOAuthFlow struct {
	FlowUrl string `json:"flow_url"`
}
type StartIntegrationOAuthResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuth = huma.Operation{
	OperationID: "complete-integration-oauth",
	Method:      http.MethodPost,
	Path:        "/integrations/oauth/complete",
	Summary:     "Complete OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type CompleteIntegrationOAuthRequestAttributes struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Code  string `json:"code"`
}
type CompleteIntegrationOAuthRequest RequestWithBodyAttributes[CompleteIntegrationOAuthRequestAttributes]
type CompleteIntegrationOAuthResponse ItemResponse[Integration]
