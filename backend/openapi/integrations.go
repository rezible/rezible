package openapi

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
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListIntegrations, o.ListIntegrations)
	huma.Register(api, CreateIntegration, o.CreateIntegration)
	huma.Register(api, GetIntegration, o.GetIntegration)
	huma.Register(api, UpdateIntegration, o.UpdateIntegration)
	huma.Register(api, DeleteIntegration, o.DeleteIntegration)
}

type (
	Integration struct {
		Id         uuid.UUID             `json:"id"`
		Attributes IntegrationAttributes `json:"attributes"`
	}

	IntegrationAttributes struct {
		ProviderId string            `json:"provider_id"`
		Kind       string            `json:"kind"`
		Enabled    bool              `json:"enabled"`
		Config     map[string]string `json:"config"`
	}
)

func IntegrationFromEnt(pc *ent.ProviderConfig) Integration {
	config := make(map[string]string)
	if jsonErr := json.Unmarshal(pc.Config, &config); jsonErr != nil {
		log.Warn().Err(jsonErr).Msg("Failed to unmarshal provider config")
	}
	attr := IntegrationAttributes{
		ProviderId: pc.ProviderID,
		Kind:       pc.ProviderType.String(),
		Enabled:    pc.Enabled,
		Config:     config,
	}

	return Integration{
		Id:         pc.ID,
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
	ProviderId string `query:"provider_id" required:"false"`
	Kind       string `query:"kind" required:"false"`
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
	ProviderId string            `json:"provider_id"`
	Kind       string            `json:"kind"`
	Enabled    bool              `json:"enabled"`
	Config     map[string]string `json:"config"`
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
