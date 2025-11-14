package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
)

type OrganizationsHandler interface {
	GetOrganization(context.Context, *GetOrganizationRequest) (*GetOrganizationResponse, error)
	FinishOrganizationSetup(context.Context, *FinishOrganizationSetupRequest) (*FinishOrganizationSetupResponse, error)
}

func (o operations) RegisterOrganizations(api huma.API) {
	huma.Register(api, GetOrganization, o.GetOrganization)
	huma.Register(api, FinishOrganizationSetup, o.FinishOrganizationSetup)
}

type (
	Organization struct {
		Id         uuid.UUID              `json:"id"`
		Attributes OrganizationAttributes `json:"attributes"`
	}

	OrganizationAttributes struct {
		Name          string `json:"name"`
		SetupRequired bool   `json:"setupRequired"`
	}
)

func OrganizationFromEnt(org *ent.Organization) Organization {
	attr := OrganizationAttributes{
		Name:          org.Name,
		SetupRequired: org.InitialSetupAt.IsZero(),
	}

	return Organization{
		Id:         org.ID,
		Attributes: attr,
	}
}

var organizationsTags = []string{"Organizations"}

var GetOrganization = huma.Operation{
	OperationID: "get-organization",
	Method:      http.MethodGet,
	Path:        "/organizations/{id}",
	Summary:     "Get Organization",
	Tags:        organizationsTags,
	Errors:      errorCodes(),
}

type GetOrganizationRequest GetIdRequest
type GetOrganizationResponse ItemResponse[Organization]

var FinishOrganizationSetup = huma.Operation{
	OperationID: "finish-organization-setup",
	Method:      http.MethodPost,
	Path:        "/organizations/{id}/setup",
	Summary:     "Finish initial org setup",
	Tags:        organizationsTags,
	Errors:      errorCodes(),
}

type FinishOrganizationSetupRequest PostIdRequest
type FinishOrganizationSetupResponse EmptyResponse
