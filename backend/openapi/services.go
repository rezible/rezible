package openapi

import (
	"context"
	"github.com/twohundreds/rezible/ent"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type ServicesHandler interface {
	ListServices(context.Context, *ListServicesRequest) (*ListServicesResponse, error)
	CreateService(context.Context, *CreateServiceRequest) (*CreateServiceResponse, error)
	GetService(context.Context, *GetServiceRequest) (*GetServiceResponse, error)
	UpdateService(context.Context, *UpdateServiceRequest) (*UpdateServiceResponse, error)
	ArchiveService(context.Context, *ArchiveServiceRequest) (*ArchiveServiceResponse, error)
}

func (o operations) RegisterServices(api huma.API) {
	huma.Register(api, ListServices, o.ListServices)
	huma.Register(api, CreateService, o.CreateService)
	huma.Register(api, GetService, o.GetService)
	huma.Register(api, UpdateService, o.UpdateService)
	huma.Register(api, ArchiveService, o.ArchiveService)
}

type Service struct {
	Attributes ServiceAttributes `json:"attributes"`
	Id         uuid.UUID         `json:"id"`
}

type ServiceAttributes struct {
	Name string `json:"name"`
}

func ServiceFromEnt(serv *ent.Service) Service {
	return Service{
		Id: serv.ID,
		Attributes: ServiceAttributes{
			Name: serv.Name,
		},
	}
}

var servicesTags = []string{"Services"}

// Operations

var ListServices = huma.Operation{
	OperationID: "list-services",
	Method:      http.MethodGet,
	Path:        "/services",
	Summary:     "List Services",
	Tags:        servicesTags,
	Errors:      errorCodes(),
}

type ListServicesRequest struct {
	ListRequest
	TeamId uuid.UUID `query:"team_id" required:"false"`
}
type ListServicesResponse PaginatedResponse[Service]

var CreateService = huma.Operation{
	OperationID: "create-service",
	Method:      http.MethodPost,
	Path:        "/services",
	Summary:     "Create a Service",
	Tags:        servicesTags,
	Errors:      errorCodes(),
}

type CreateServiceAttributes struct {
	Name string `json:"name"`
}
type CreateServiceRequest RequestWithBodyAttributes[CreateServiceAttributes]
type CreateServiceResponse ItemResponse[Service]

var GetService = huma.Operation{
	OperationID: "get-service",
	Method:      http.MethodGet,
	Path:        "/services/{id}",
	Summary:     "Get a Service",
	Tags:        servicesTags,
	Errors:      errorCodes(),
}

type GetServiceRequest GetFlexibleIdRequest
type GetServiceResponse ItemResponse[Service]

var UpdateService = huma.Operation{
	OperationID: "update-service",
	Method:      http.MethodPatch,
	Path:        "/services/{id}",
	Summary:     "Update a Service",
	Tags:        servicesTags,
	Errors:      errorCodes(),
}

type UpdateServiceAttributes struct {
	Name OmittableNullable[string] `json:"name"`
}
type UpdateServiceRequest UpdateIdRequest[UpdateServiceAttributes]
type UpdateServiceResponse ItemResponse[Service]

var ArchiveService = huma.Operation{
	OperationID: "archive-service",
	Method:      http.MethodDelete,
	Path:        "/services/{id}",
	Summary:     "Archive a Service",
	Tags:        servicesTags,
	Errors:      errorCodes(),
}

type ArchiveServiceRequest ArchiveIdRequest
type ArchiveServiceResponse EmptyResponse
