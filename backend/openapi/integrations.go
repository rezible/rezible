package openapi

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type IntegrationsHandler interface {
	ListIntegrations(context.Context, *ListIntegrationsRequest) (*ListIntegrationsResponse, error)
	CreateIntegration(context.Context, *CreateIntegrationRequest) (*CreateIntegrationResponse, error)
	GetIntegration(context.Context, *GetIntegrationRequest) (*GetIntegrationResponse, error)
	UpdateIntegration(context.Context, *UpdateIntegrationRequest) (*UpdateIntegrationResponse, error)
	ArchiveIntegration(context.Context, *ArchiveIntegrationRequest) (*ArchiveIntegrationResponse, error)
}

func (o operations) RegisterIntegrations(api huma.API) {
	huma.Register(api, ListIntegrations, o.ListIntegrations)
	huma.Register(api, CreateIntegration, o.CreateIntegration)
	huma.Register(api, GetIntegration, o.GetIntegration)
	huma.Register(api, UpdateIntegration, o.UpdateIntegration)
	huma.Register(api, ArchiveIntegration, o.ArchiveIntegration)
}

type Integration struct {
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

type ListIntegrationsRequest ListRequest
type ListIntegrationsResponse PaginatedResponse[Integration]

var CreateIntegration = huma.Operation{
	OperationID: "create-integration",
	Method:      http.MethodPost,
	Path:        "/integrations",
	Summary:     "Create an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type CreateIntegrationRequest struct {
	Body struct {
		// Attributes IntegrationAttributes
	}
}
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
}
type UpdateIntegrationRequest UpdateIdRequest[UpdateIntegrationAttributes]
type UpdateIntegrationResponse ItemResponse[Integration]

var ArchiveIntegration = huma.Operation{
	OperationID: "archive-integration",
	Method:      http.MethodDelete,
	Path:        "/integrations/{id}",
	Summary:     "Archive an Integration",
	Tags:        integrationsTags,
	Errors:      errorCodes(),
}

type ArchiveIntegrationRequest ArchiveIdRequest
type ArchiveIntegrationResponse EmptyResponse
