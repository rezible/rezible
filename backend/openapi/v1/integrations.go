package v1

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
)

type IntegrationsHandler interface {
	ListIntegrations(context.Context, *ListIntegrationsRequest) (*ListIntegrationsResponse, error)
	CreateIntegration(context.Context, *CreateIntegrationRequest) (*CreateIntegrationResponse, error)
	GetIntegration(context.Context, *GetIntegrationRequest) (*GetIntegrationResponse, error)
	UpdateIntegration(context.Context, *UpdateIntegrationRequest) (*UpdateIntegrationResponse, error)
	DeleteIntegration(context.Context, *DeleteIntegrationRequest) (*DeleteIntegrationResponse, error)

	StartIntegrationOAuth(context.Context, *StartIntegrationOAuthRequest) (*StartIntegrationOAuthResponse, error)
	CompleteIntegrationOAuth(context.Context, *CompleteIntegrationOAuthRequest) (*CompleteIntegrationOAuthResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListIntegrations, o.ListIntegrations)
	huma.Register(api, CreateIntegration, o.CreateIntegration)
	huma.Register(api, GetIntegration, o.GetIntegration)
	huma.Register(api, UpdateIntegration, o.UpdateIntegration)
	huma.Register(api, DeleteIntegration, o.DeleteIntegration)

	huma.Register(api, StartIntegrationOAuth, o.StartIntegrationOAuth)
	huma.Register(api, CompleteIntegrationOAuth, o.CompleteIntegrationOAuth)
}

type (
	Integration struct {
		Id         uuid.UUID             `json:"id"`
		Attributes IntegrationAttributes `json:"attributes"`
	}

	IntegrationAttributes struct {
		Name    string            `json:"name"`
		Type    string            `json:"type"`
		Enabled bool              `json:"enabled"`
		Config  map[string]string `json:"config"`
	}
)

func IntegrationFromEnt(intg *ent.Integration) Integration {
	config := make(map[string]string)
	if jsonErr := json.Unmarshal(intg.Config, &config); jsonErr != nil {
		log.Warn().Err(jsonErr).Msg("Failed to unmarshal integration config")
	}
	attr := IntegrationAttributes{
		Name:    intg.Name,
		Type:    intg.IntegrationType.String(),
		Enabled: intg.Enabled,
		Config:  config,
	}

	return Integration{
		Id:         intg.ID,
		Attributes: attr,
	}
}

var integrationsTags = []string{"Integrations"}

var ListIntegrations = huma.Operation{
	OperationID: "list-integrations",
	Method:      http.MethodGet,
	Path:        "/integrations",
	Summary:     "List Integrations",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ListIntegrationsRequest struct {
	ListRequest
	Name string `query:"name" required:"false"`
	Type string `query:"type" required:"false"`
}
type ListIntegrationsResponse PaginatedResponse[Integration]

var CreateIntegration = huma.Operation{
	OperationID: "create-integration",
	Method:      http.MethodPost,
	Path:        "/integrations",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type CreateIntegrationRequestAttributes struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config"`
}
type CreateIntegrationRequest RequestWithBodyAttributes[CreateIntegrationRequestAttributes]
type CreateIntegrationResponse ItemResponse[Integration]

var GetIntegration = huma.Operation{
	OperationID: "get-integration",
	Method:      http.MethodGet,
	Path:        "/integrations/{id}",
	Summary:     "Get an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type GetIntegrationRequest GetIdRequest
type GetIntegrationResponse ItemResponse[Integration]

var UpdateIntegration = huma.Operation{
	OperationID: "update-integration",
	Method:      http.MethodPatch,
	Path:        "/integrations/{id}",
	Summary:     "Update an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type UpdateIntegrationAttributes struct {
	Enabled *bool              `json:"enabled,omitempty"`
	Config  *map[string]string `json:"config,omitempty"`
}
type UpdateIntegrationRequest UpdateIdRequest[UpdateIntegrationAttributes]
type UpdateIntegrationResponse ItemResponse[Integration]

var DeleteIntegration = huma.Operation{
	OperationID: "delete-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/{id}",
	Summary:     "Delete an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type DeleteIntegrationRequest DeleteIdRequest
type DeleteIntegrationResponse EmptyResponse

var StartIntegrationOAuth = huma.Operation{
	OperationID: "start-integration-oauth",
	Method:      http.MethodPost,
	Path:        "/integrations_oauth/start",
	Summary:     "Start OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type StartIntegrationOAuthRequestAttributes struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type StartIntegrationOAuthRequest RequestWithBodyAttributes[StartIntegrationOAuthRequestAttributes]

type IntegrationOAuthFlow struct {
	FlowUrl string `json:"flow_url"`
}
type StartIntegrationOAuthResponse ItemResponse[IntegrationOAuthFlow]

var CompleteIntegrationOAuth = huma.Operation{
	OperationID: "complete-integration-oauth",
	Method:      http.MethodPost,
	Path:        "/integrations_oauth/complete",
	Summary:     "Complete OAuth flow for an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type CompleteIntegrationOAuthRequestAttributes struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	State string `json:"state"`
	Code  string `json:"code"`
}
type CompleteIntegrationOAuthRequest RequestWithBodyAttributes[CompleteIntegrationOAuthRequestAttributes]
type CompleteIntegrationOAuthResponse ItemResponse[Integration]
