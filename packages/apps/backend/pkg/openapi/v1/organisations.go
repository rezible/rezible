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
	CompleteOrgSetup(context.Context, *CompleteOrgSetupRequest) (*CompleteOrgSetupResponse, error)
	UpdateOrganizationPreferences(context.Context, *UpdateOrganizationPreferencesRequest) (*UpdateOrganizationPreferencesResponse, error)
}

func (o operations) RegisterOrganizations(api huma.API) {
	huma.Register(api, GetOrganization, o.GetOrganization)
	huma.Register(api, CompleteOrgSetup, o.CompleteOrgSetup)
	huma.Register(api, UpdateOrganizationPreferences, o.UpdateOrganizationPreferences)
}

type (
	Organization struct {
		Id         uuid.UUID              `json:"id"`
		Attributes OrganizationAttributes `json:"attributes"`
	}

	OrganizationAttributes struct {
		Name          string                   `json:"name"`
		SetupRequired bool                     `json:"setupRequired"`
		Preferences   *OrganizationPreferences `json:"preferences"`
	}

	OrganizationPreferences struct {
		EnableIncidentManagement bool   `json:"enableIncidentManagement"`
		Timezone                 string `json:"timezone,omitempty"`
	}
)

func OrganizationFromEnt(org *ent.Organization) Organization {
	attr := OrganizationAttributes{
		Name:          org.Name,
		SetupRequired: org.Edges.Preferences == nil || org.Edges.Preferences.InitialSetupAt.IsZero(),
	}
	if org.Edges.Preferences != nil {
		attr.Preferences = new(OrganizationPreferencesFromEnt(org.Edges.Preferences))
	}

	return Organization{Id: org.ID, Attributes: attr}
}

func OrganizationPreferencesFromEnt(prefs *ent.OrganizationPreferences) OrganizationPreferences {
	if prefs == nil {
		return OrganizationPreferences{}
	}
	return OrganizationPreferences{
		EnableIncidentManagement: prefs.EnableIncidentManagement,
		Timezone:                 prefs.Timezone,
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

type GetOrganizationRequest IdRequest
type GetOrganizationResponse ItemResponse[Organization]

// TODO: require organization admin authorization.
var CompleteOrgSetup = huma.Operation{
	OperationID: "complete-org-setup",
	Method:      http.MethodPost,
	Path:        "/organizations/{id}/setup",
	Summary:     "Complete Organization Setup",
	Tags:        organizationsTags,
	Errors:      ErrorCodes(http.StatusConflict),
}

type CompleteOrgSetupRequestAttributes struct {
	Name     string `json:"name"`
	Timezone string `json:"timezone,omitempty"`
}
type CompleteOrgSetupRequest IdRequestWithBody[CompleteOrgSetupRequestAttributes]
type CompleteOrgSetupResponse ItemResponse[Organization]

var UpdateOrganizationPreferences = huma.Operation{
	OperationID: "update-organization-preferences",
	Method:      http.MethodPatch,
	Path:        "/organizations/{id}/preferences",
	Summary:     "Update Organization Preferences",
	Tags:        organizationsTags,
	Errors:      ErrorCodes(),
}

type UpdateOrganizationPreferencesRequestAttributes struct {
	EnableIncidentManagement *bool   `json:"enableIncidentManagement,omitempty"`
	Timezone                 *string `json:"timezone,omitempty"`
}
type UpdateOrganizationPreferencesRequest IdRequestWithBody[UpdateOrganizationPreferencesRequestAttributes]
type UpdateOrganizationPreferencesResponse ItemResponse[OrganizationPreferences]
