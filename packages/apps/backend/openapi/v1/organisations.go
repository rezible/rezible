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
	UpdateOrganizationPreferences(context.Context, *UpdateOrganizationPreferencesRequest) (*UpdateOrganizationPreferencesResponse, error)
}

func (o operations) RegisterOrganizations(api huma.API) {
	huma.Register(api, GetOrganization, o.GetOrganization)
	huma.Register(api, UpdateOrganizationPreferences, o.UpdateOrganizationPreferences)
}

type (
	Organization struct {
		Id         uuid.UUID              `json:"id"`
		Attributes OrganizationAttributes `json:"attributes"`
	}

	OrganizationAttributes struct {
		Name          string                  `json:"name"`
		SetupRequired bool                    `json:"setupRequired"`
		Preferences   OrganizationPreferences `json:"preferences"`
	}

	OrganizationPreferences struct {
		EnableIncidentManagement bool `json:"enableIncidentManagement"`
	}
)

func OrganizationFromEnt(org *ent.Organization) Organization {
	attr := OrganizationAttributes{
		Name:          org.Name,
		SetupRequired: org.Edges.Preferences == nil || org.Edges.Preferences.InitialSetupAt.IsZero(),
		Preferences:   OrganizationPreferencesFromEnt(org.Edges.Preferences),
	}

	return Organization{Id: org.ID, Attributes: attr}
}

func OrganizationPreferencesFromEnt(prefs *ent.OrganizationPreferences) OrganizationPreferences {
	if prefs == nil {
		return OrganizationPreferences{}
	}
	return OrganizationPreferences{
		EnableIncidentManagement: prefs.EnableIncidentManagement,
	}
}

var organizationsTags = []string{"Organizations"}

var GetOrganization = huma.Operation{
	OperationID: "get-organization",
	Method:      http.MethodGet,
	Path:        "/organizations/{id}",
	Summary:     "Get Organization",
	Tags:        organizationsTags,
	Errors:      ErrorCodes(),
}

type GetOrganizationRequest EmptyIdRequest
type GetOrganizationResponse ItemResponse[Organization]

var UpdateOrganizationPreferences = huma.Operation{
	OperationID: "update-organization-preferences",
	Method:      http.MethodPatch,
	Path:        "/organizations/{id}/preferences",
	Summary:     "Update Organization Preferences",
	Tags:        organizationsTags,
	Errors:      ErrorCodes(),
}

type UpdateOrganizationPreferencesRequestAttributes struct {
	InitialSetupComplete     *bool `json:"initialSetupComplete,omitempty"`
	EnableIncidentManagement *bool `json:"enableIncidentManagement,omitempty"`
}
type UpdateOrganizationPreferencesRequest IdRequest[UpdateOrganizationPreferencesRequestAttributes]
type UpdateOrganizationPreferencesResponse ItemResponse[OrganizationPreferences]
