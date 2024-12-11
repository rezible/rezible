package openapi

import (
	"context"
	"github.com/rezible/rezible/ent"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type FunctionalitiesHandler interface {
	ListFunctionalities(context.Context, *ListFunctionalitiesRequest) (*ListFunctionalitiesResponse, error)
	CreateFunctionality(context.Context, *CreateFunctionalityRequest) (*CreateFunctionalityResponse, error)
	GetFunctionality(context.Context, *GetFunctionalityRequest) (*GetFunctionalityResponse, error)
	UpdateFunctionality(context.Context, *UpdateFunctionalityRequest) (*UpdateFunctionalityResponse, error)
	ArchiveFunctionality(context.Context, *ArchiveFunctionalityRequest) (*ArchiveFunctionalityResponse, error)
}

func (o operations) RegisterFunctionalities(api huma.API) {
	huma.Register(api, ListFunctionalities, o.ListFunctionalities)
	huma.Register(api, CreateFunctionality, o.CreateFunctionality)
	huma.Register(api, GetFunctionality, o.GetFunctionality)
	huma.Register(api, UpdateFunctionality, o.UpdateFunctionality)
	huma.Register(api, ArchiveFunctionality, o.ArchiveFunctionality)
}

type Functionality struct {
	Id         uuid.UUID               `json:"id"`
	Attributes FunctionalityAttributes `json:"attributes"`
}

type FunctionalityAttributes struct {
	Name string `json:"name"`
}

func FunctionalityFromEnt(fun *ent.Functionality) Functionality {
	return Functionality{
		Id: fun.ID,
		Attributes: FunctionalityAttributes{
			Name: fun.Name,
		},
	}
}

var functionalitiesTags = []string{"Functionalities"}

// Operations

var ListFunctionalities = huma.Operation{
	OperationID: "list-functionalities",
	Method:      http.MethodGet,
	Path:        "/functionalities",
	Summary:     "List Functionalities",
	Tags:        functionalitiesTags,
	Errors:      errorCodes(),
}

type ListFunctionalitiesRequest ListRequest
type ListFunctionalitiesResponse PaginatedResponse[Functionality]

var CreateFunctionality = huma.Operation{
	OperationID: "create-functionality",
	Method:      http.MethodPost,
	Path:        "/functionalities",
	Summary:     "Create a Functionality",
	Tags:        functionalitiesTags,
	Errors:      errorCodes(),
}

type CreateFunctionalityAttributes struct {
	Name string `json:"name"`
}
type CreateFunctionalityRequest RequestWithBodyAttributes[CreateFunctionalityAttributes]
type CreateFunctionalityResponse ItemResponse[Functionality]

var GetFunctionality = huma.Operation{
	OperationID: "get-functionality",
	Method:      http.MethodGet,
	Path:        "/functionalities/{id}",
	Summary:     "Get a Functionality",
	Tags:        functionalitiesTags,
	Errors:      errorCodes(),
}

type GetFunctionalityRequest GetIdRequest
type GetFunctionalityResponse ItemResponse[Functionality]

var UpdateFunctionality = huma.Operation{
	OperationID: "update-functionality",
	Method:      http.MethodPatch,
	Path:        "/functionalities/{id}",
	Summary:     "Update a Functionality",
	Tags:        functionalitiesTags,
	Errors:      errorCodes(),
}

type UpdateFunctionalityAttributes struct {
	Name OmittableNullable[string] `json:"name"`
}
type UpdateFunctionalityRequest UpdateIdRequest[UpdateFunctionalityAttributes]
type UpdateFunctionalityResponse ItemResponse[Functionality]

var ArchiveFunctionality = huma.Operation{
	OperationID: "archive-functionality",
	Method:      http.MethodDelete,
	Path:        "/functionalities/{id}",
	Summary:     "Archive a Functionality",
	Tags:        functionalitiesTags,
	Errors:      errorCodes(),
}

type ArchiveFunctionalityRequest ArchiveIdRequest
type ArchiveFunctionalityResponse EmptyResponse
