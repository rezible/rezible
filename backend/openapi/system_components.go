package openapi

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"net/http"
)

type SystemComponentsHandler interface {
	ListSystemComponents(context.Context, *ListSystemComponentsRequest) (*ListSystemComponentsResponse, error)
	CreateSystemComponent(context.Context, *CreateSystemComponentRequest) (*CreateSystemComponentResponse, error)
	GetSystemComponent(context.Context, *GetSystemComponentRequest) (*GetSystemComponentResponse, error)
	UpdateSystemComponent(context.Context, *UpdateSystemComponentRequest) (*UpdateSystemComponentResponse, error)
	ArchiveSystemComponent(context.Context, *ArchiveSystemComponentRequest) (*ArchiveSystemComponentResponse, error)

	CreateSystemComponentConstraint(context.Context, *CreateSystemComponentConstraintRequest) (*CreateSystemComponentConstraintResponse, error)
	GetSystemComponentConstraint(context.Context, *GetSystemComponentConstraintRequest) (*GetSystemComponentConstraintResponse, error)
	UpdateSystemComponentConstraint(context.Context, *UpdateSystemComponentConstraintRequest) (*UpdateSystemComponentConstraintResponse, error)
	ArchiveSystemComponentConstraint(context.Context, *ArchiveSystemComponentConstraintRequest) (*ArchiveSystemComponentConstraintResponse, error)

	CreateSystemComponentControl(context.Context, *CreateSystemComponentControlRequest) (*CreateSystemComponentControlResponse, error)
	GetSystemComponentControl(context.Context, *GetSystemComponentControlRequest) (*GetSystemComponentControlResponse, error)
	UpdateSystemComponentControl(context.Context, *UpdateSystemComponentControlRequest) (*UpdateSystemComponentControlResponse, error)
	ArchiveSystemComponentControl(context.Context, *ArchiveSystemComponentControlRequest) (*ArchiveSystemComponentControlResponse, error)

	CreateSystemComponentSignal(context.Context, *CreateSystemComponentSignalRequest) (*CreateSystemComponentSignalResponse, error)
	GetSystemComponentSignal(context.Context, *GetSystemComponentSignalRequest) (*GetSystemComponentSignalResponse, error)
	UpdateSystemComponentSignal(context.Context, *UpdateSystemComponentSignalRequest) (*UpdateSystemComponentSignalResponse, error)
	ArchiveSystemComponentSignal(context.Context, *ArchiveSystemComponentSignalRequest) (*ArchiveSystemComponentSignalResponse, error)
}

func (o operations) RegisterSystemComponents(api huma.API) {
	huma.Register(api, ListSystemComponents, o.ListSystemComponents)
	huma.Register(api, CreateSystemComponent, o.CreateSystemComponent)
	huma.Register(api, GetSystemComponent, o.GetSystemComponent)
	huma.Register(api, UpdateSystemComponent, o.UpdateSystemComponent)
	huma.Register(api, ArchiveSystemComponent, o.ArchiveSystemComponent)

	huma.Register(api, CreateSystemComponentConstraint, o.CreateSystemComponentConstraint)
	huma.Register(api, GetSystemComponentConstraint, o.GetSystemComponentConstraint)
	huma.Register(api, UpdateSystemComponentConstraint, o.UpdateSystemComponentConstraint)
	huma.Register(api, ArchiveSystemComponentConstraint, o.ArchiveSystemComponentConstraint)

	huma.Register(api, CreateSystemComponentControl, o.CreateSystemComponentControl)
	huma.Register(api, GetSystemComponentControl, o.GetSystemComponentControl)
	huma.Register(api, UpdateSystemComponentControl, o.UpdateSystemComponentControl)
	huma.Register(api, ArchiveSystemComponentControl, o.ArchiveSystemComponentControl)

	huma.Register(api, CreateSystemComponentSignal, o.CreateSystemComponentSignal)
	huma.Register(api, GetSystemComponentSignal, o.GetSystemComponentSignal)
	huma.Register(api, UpdateSystemComponentSignal, o.UpdateSystemComponentSignal)
	huma.Register(api, ArchiveSystemComponentSignal, o.ArchiveSystemComponentSignal)
}

type (
	SystemComponent struct {
		Id         uuid.UUID                 `json:"id"`
		Attributes SystemComponentAttributes `json:"attributes"`
	}
	SystemComponentAttributes struct {
		Name        string                      `json:"name"`
		Kind        string                      `json:"kind"`
		Description string                      `json:"description"`
		Properties  map[string]any              `json:"properties"`
		Constraints []SystemComponentConstraint `json:"constraints"`
		Signals     []SystemComponentSignal     `json:"signals"`
		Controls    []SystemComponentControl    `json:"controls"`
	}

	SystemComponentConstraint struct {
		Id         uuid.UUID                           `json:"id"`
		Attributes SystemComponentConstraintAttributes `json:"attributes"`
	}
	SystemComponentConstraintAttributes struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}

	SystemComponentSignal struct {
		Id         uuid.UUID                       `json:"id"`
		Attributes SystemComponentSignalAttributes `json:"attributes"`
	}
	SystemComponentSignalAttributes struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}

	SystemComponentControl struct {
		Id         uuid.UUID                        `json:"id"`
		Attributes SystemComponentControlAttributes `json:"attributes"`
	}
	SystemComponentControlAttributes struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}
)

func SystemComponentFromEnt(sc *ent.SystemComponent) SystemComponent {
	return SystemComponent{
		Id: sc.ID,
		Attributes: SystemComponentAttributes{
			Name: sc.Name,
		},
	}
}

// System Components
var systemComponentsTags = []string{"System Components"}

var ListSystemComponents = huma.Operation{
	OperationID: "list-system-components",
	Method:      http.MethodGet,
	Path:        "/system_components",
	Summary:     "List System Components",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ListSystemComponentsRequest struct {
	ListRequest
}
type ListSystemComponentsResponse PaginatedResponse[SystemComponent]

var CreateSystemComponent = huma.Operation{
	OperationID: "create-system-component",
	Method:      http.MethodPost,
	Path:        "/system_components",
	Summary:     "Create a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type CreateSystemComponentAttributes struct {
	Name string `json:"name"`
}
type CreateSystemComponentRequest RequestWithBodyAttributes[CreateSystemComponentAttributes]
type CreateSystemComponentResponse ItemResponse[SystemComponent]

var GetSystemComponent = huma.Operation{
	OperationID: "get-system-component",
	Method:      http.MethodGet,
	Path:        "/system_components/{id}",
	Summary:     "Get a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type GetSystemComponentRequest GetIdRequest
type GetSystemComponentResponse ItemResponse[SystemComponent]

var UpdateSystemComponent = huma.Operation{
	OperationID: "update-system-component",
	Method:      http.MethodPatch,
	Path:        "/system_components/{id}",
	Summary:     "Update a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type UpdateSystemComponentAttributes struct {
}
type UpdateSystemComponentRequest UpdateIdRequest[UpdateSystemComponentAttributes]
type UpdateSystemComponentResponse ItemResponse[SystemComponent]

var ArchiveSystemComponent = huma.Operation{
	OperationID: "archive-system-component",
	Method:      http.MethodDelete,
	Path:        "/system_components/{id}",
	Summary:     "Archive a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentRequest ArchiveIdRequest
type ArchiveSystemComponentResponse EmptyResponse

// System Component Constraint Operations

type SystemComponentConstraintRequest struct {
	ComponentId  uuid.UUID `path:"componentId"`
	ConstraintId uuid.UUID `path:"constraintId"`
}

type SystemComponentConstraintBodyRequest[A any] struct {
	ComponentId  uuid.UUID `path:"componentId"`
	ConstraintId uuid.UUID `path:"constraintId"`
	Body         A         `json:"body"`
}

const componentConstraintPath = "/system_components/{componentId}/constraints/{constraintId}"

var CreateSystemComponentConstraint = huma.Operation{
	OperationID: "create-system-component-constraint",
	Method:      http.MethodPost,
	Path:        "/system_components/{id}/constraints",
	Summary:     "Create a System Component Constraint",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type CreateSystemComponentConstraintAttributes struct {
	Label       string `json:"name"`
	Description string `json:"description"`
}
type CreateSystemComponentConstraintRequest RequestWithBodyAttributes[CreateSystemComponentConstraintAttributes]
type CreateSystemComponentConstraintResponse ItemResponse[SystemComponentConstraint]

var GetSystemComponentConstraint = huma.Operation{
	OperationID: "get-system-component-constraint",
	Method:      http.MethodGet,
	Path:        componentConstraintPath,
	Summary:     "Get a System Component",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type GetSystemComponentConstraintRequest GetIdRequest
type GetSystemComponentConstraintResponse ItemResponse[SystemComponentConstraint]

var UpdateSystemComponentConstraint = huma.Operation{
	OperationID: "update-system-component-constraint",
	Method:      http.MethodPatch,
	Path:        componentConstraintPath,
	Summary:     "Update a System Component Constraint",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type UpdateSystemComponentConstraintAttributes struct {
}
type UpdateSystemComponentConstraintRequest SystemComponentConstraintBodyRequest[UpdateSystemComponentConstraintAttributes]
type UpdateSystemComponentConstraintResponse ItemResponse[SystemComponentConstraint]

var ArchiveSystemComponentConstraint = huma.Operation{
	OperationID: "archive-system-component-constraint",
	Method:      http.MethodDelete,
	Path:        componentConstraintPath,
	Summary:     "Archive a System Component Constraint",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentConstraintRequest SystemComponentConstraintRequest
type ArchiveSystemComponentConstraintResponse EmptyResponse

// System Component Control Operations

type SystemComponentControlRequest struct {
	ComponentId uuid.UUID `path:"componentId"`
	ControlId   uuid.UUID `path:"controlId"`
}

type SystemComponentControlBodyRequest[A any] struct {
	ComponentId uuid.UUID `path:"componentId"`
	ControlId   uuid.UUID `path:"controlId"`
	Body        A         `json:"body"`
}

const componentControlPath = "/system_components/{componentId}/controls/{controlId}"

var CreateSystemComponentControl = huma.Operation{
	OperationID: "create-system-component-control",
	Method:      http.MethodPost,
	Path:        "/system_components/{id}/controls",
	Summary:     "Create a System Component Control",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type CreateSystemComponentControlAttributes struct {
	Label       string `json:"name"`
	Description string `json:"description"`
}
type CreateSystemComponentControlRequest RequestWithBodyAttributes[CreateSystemComponentControlAttributes]
type CreateSystemComponentControlResponse ItemResponse[SystemComponentControl]

var GetSystemComponentControl = huma.Operation{
	OperationID: "get-system-component-control",
	Method:      http.MethodGet,
	Path:        componentControlPath,
	Summary:     "Get a System Component Control",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type GetSystemComponentControlRequest GetIdRequest
type GetSystemComponentControlResponse ItemResponse[SystemComponentControl]

var UpdateSystemComponentControl = huma.Operation{
	OperationID: "update-system-component-control",
	Method:      http.MethodPatch,
	Path:        componentControlPath,
	Summary:     "Update a System Component Control",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type UpdateSystemComponentControlAttributes struct {
}
type UpdateSystemComponentControlRequest SystemComponentControlBodyRequest[UpdateSystemComponentControlAttributes]
type UpdateSystemComponentControlResponse ItemResponse[SystemComponentControl]

var ArchiveSystemComponentControl = huma.Operation{
	OperationID: "archive-system-component-control",
	Method:      http.MethodDelete,
	Path:        componentControlPath,
	Summary:     "Archive a System Component Control",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentControlRequest SystemComponentControlRequest
type ArchiveSystemComponentControlResponse EmptyResponse

// System Component Signal Operations

type SystemComponentSignalRequest struct {
	ComponentId uuid.UUID `path:"componentId"`
	SignalId    uuid.UUID `path:"signalId"`
}

type SystemComponentSignalBodyRequest[A any] struct {
	ComponentId uuid.UUID `path:"componentId"`
	SignalId    uuid.UUID `path:"signalId"`
	Body        A         `json:"body"`
}

const componentSignalPath = "/system_components/{componentId}/signals/{signalId}"

var CreateSystemComponentSignal = huma.Operation{
	OperationID: "create-system-component-signal",
	Method:      http.MethodPost,
	Path:        "/system_components/{id}/signals",
	Summary:     "Create a System Component Signal",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type CreateSystemComponentSignalAttributes struct {
	Label       string `json:"name"`
	Description string `json:"description"`
}
type CreateSystemComponentSignalRequest RequestWithBodyAttributes[CreateSystemComponentSignalAttributes]
type CreateSystemComponentSignalResponse ItemResponse[SystemComponentSignal]

var GetSystemComponentSignal = huma.Operation{
	OperationID: "get-system-component-signal",
	Method:      http.MethodGet,
	Path:        componentSignalPath,
	Summary:     "Get a System Component Signal",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type GetSystemComponentSignalRequest GetIdRequest
type GetSystemComponentSignalResponse ItemResponse[SystemComponentSignal]

var UpdateSystemComponentSignal = huma.Operation{
	OperationID: "update-system-component-signal",
	Method:      http.MethodPatch,
	Path:        componentSignalPath,
	Summary:     "Update a System Component Signal",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type UpdateSystemComponentSignalAttributes struct {
}
type UpdateSystemComponentSignalRequest SystemComponentSignalBodyRequest[UpdateSystemComponentSignalAttributes]
type UpdateSystemComponentSignalResponse ItemResponse[SystemComponentSignal]

var ArchiveSystemComponentSignal = huma.Operation{
	OperationID: "archive-system-component-signal",
	Method:      http.MethodDelete,
	Path:        componentSignalPath,
	Summary:     "Archive a System Component Signal",
	Tags:        systemComponentsTags,
	Errors:      errorCodes(),
}

type ArchiveSystemComponentSignalRequest SystemComponentSignalRequest
type ArchiveSystemComponentSignalResponse EmptyResponse
