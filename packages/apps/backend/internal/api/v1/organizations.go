package apiv1

import (
	"context"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type organizationsHandler struct {
	orgs rez.OrganizationService
}

func newOrganizationsHandler(orgs rez.OrganizationService) *organizationsHandler {
	return &organizationsHandler{orgs: orgs}
}

func (h *organizationsHandler) GetOrganization(ctx context.Context, req *oapi.GetOrganizationRequest) (*oapi.GetOrganizationResponse, error) {
	var resp oapi.GetOrganizationResponse

	org, orgErr := h.orgs.Get(ctx, organization.ID(req.Id))
	if orgErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch organization", orgErr)
	}
	resp.Body.Data = oapi.OrganizationFromEnt(org)

	return &resp, nil
}

func (h *organizationsHandler) CompleteOrgSetup(ctx context.Context, req *oapi.CompleteOrgSetupRequest) (*oapi.CompleteOrgSetupResponse, error) {
	// TODO: require organization admin authorization.
	var resp oapi.CompleteOrgSetupResponse

	attrs := req.Body.Attributes
	params := rez.CompleteOrgSetupParams{
		Name:     strings.TrimSpace(attrs.Name),
		Timezone: strings.TrimSpace(attrs.Timezone),
	}
	setupOrg, completeSetupErr := h.orgs.CompleteOrgSetup(ctx, req.Id, params)
	if completeSetupErr != nil {
		return nil, oapi.Error(ctx, "failed to complete organization setup", completeSetupErr)
	}
	resp.Body.Data = oapi.OrganizationFromEnt(setupOrg)
	return &resp, nil
}

func (h *organizationsHandler) UpdateOrganizationPreferences(ctx context.Context, req *oapi.UpdateOrganizationPreferencesRequest) (*oapi.UpdateOrganizationPreferencesResponse, error) {
	var resp oapi.UpdateOrganizationPreferencesResponse

	attrs := req.Body.Attributes
	setFn := func(m *ent.OrganizationPreferencesMutation) {
		if attrs.EnableIncidentManagement != nil {
			m.SetEnableIncidentManagement(*attrs.EnableIncidentManagement)
		}
		if attrs.Timezone != nil {
			timezone := strings.TrimSpace(*attrs.Timezone)
			if timezone == "" {
				m.ClearTimezone()
			} else {
				m.SetTimezone(timezone)
			}
		}
	}

	prefs, prefsErr := h.orgs.SetPreferences(ctx, req.Id, setFn)
	if prefsErr != nil {
		return nil, oapi.Error(ctx, "failed to update", prefsErr)
	}
	resp.Body.Data = oapi.OrganizationPreferencesFromEnt(prefs)

	return &resp, nil
}
