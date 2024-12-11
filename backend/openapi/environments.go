package openapi

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type EnvironmentsHandler interface {
	ListEnvironments(context.Context, *ListEnvironmentsRequest) (*ListEnvironmentsResponse, error)
	CreateEnvironment(context.Context, *CreateEnvironmentRequest) (*CreateEnvironmentResponse, error)
	GetEnvironment(context.Context, *GetEnvironmentRequest) (*GetEnvironmentResponse, error)
	UpdateEnvironment(context.Context, *UpdateEnvironmentRequest) (*UpdateEnvironmentResponse, error)
	ArchiveEnvironment(context.Context, *ArchiveEnvironmentRequest) (*ArchiveEnvironmentResponse, error)
}

func (o operations) RegisterEnvironments(api huma.API) {
	huma.Register(api, ListEnvironments, o.ListEnvironments)
	huma.Register(api, CreateEnvironment, o.CreateEnvironment)
	huma.Register(api, GetEnvironment, o.GetEnvironment)
	huma.Register(api, UpdateEnvironment, o.UpdateEnvironment)
	huma.Register(api, ArchiveEnvironment, o.ArchiveEnvironment)
}

type Environment struct {
	Id         uuid.UUID             `json:"id"`
	Attributes EnvironmentAttributes `json:"attributes"`
}

type EnvironmentAttributes struct {
	Name     string `json:"name"`
	Archived bool   `json:"archived"`
}

func EnvironmentFromEnt(env *ent.Environment) Environment {
	return Environment{
		Id: env.ID,
		Attributes: EnvironmentAttributes{
			Name:     env.Name,
			Archived: !env.ArchiveTime.IsZero(),
		},
	}
}

var environmentsTags = []string{"Environments"}

// Operations

var ListEnvironments = huma.Operation{
	OperationID: "list-environments",
	Method:      http.MethodGet,
	Path:        "/environments",
	Summary:     "List Environments",
	Tags:        environmentsTags,
	Errors:      errorCodes(),
}

type ListEnvironmentsRequest ListRequest
type ListEnvironmentsResponse PaginatedResponse[Environment]

var CreateEnvironment = huma.Operation{
	OperationID: "create-environment",
	Method:      http.MethodPost,
	Path:        "/environments",
	Summary:     "Create an Environment",
	Tags:        environmentsTags,
	Errors:      errorCodes(),
}

type CreateEnvironmentAttributes struct {
	Name string `json:"name"`
}
type CreateEnvironmentRequest RequestWithBodyAttributes[CreateEnvironmentAttributes]
type CreateEnvironmentResponse ItemResponse[Environment]

var GetEnvironment = huma.Operation{
	OperationID: "get-environment",
	Method:      http.MethodGet,
	Path:        "/environments/{id}",
	Summary:     "Get an Environment",
	Tags:        environmentsTags,
	Errors:      errorCodes(),
}

type GetEnvironmentRequest GetIdRequest
type GetEnvironmentResponse ItemResponse[Environment]

var UpdateEnvironment = huma.Operation{
	OperationID: "update-environment",
	Method:      http.MethodPatch,
	Path:        "/environments/{id}",
	Summary:     "Update an Environment",
	Tags:        environmentsTags,
	Errors:      errorCodes(),
}

type UpdateEnvironmentAttributes struct {
	Name     *string `json:"name,omitempty"`
	Archived *bool   `json:"archived,omitempty"`
}
type UpdateEnvironmentRequest UpdateIdRequest[UpdateEnvironmentAttributes]
type UpdateEnvironmentResponse ItemResponse[Environment]

var ArchiveEnvironment = huma.Operation{
	OperationID: "archive-environment",
	Method:      http.MethodDelete,
	Path:        "/environments/{id}",
	Summary:     "Archive an Environment",
	Tags:        environmentsTags,
	Errors:      errorCodes(),
}

type ArchiveEnvironmentRequest ArchiveIdRequest
type ArchiveEnvironmentResponse EmptyResponse
