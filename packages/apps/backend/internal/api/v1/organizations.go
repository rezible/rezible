package apiv1

import (
	"context"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
	oapi "github.com/rezible/rezible/openapi/v1"
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

func (h *organizationsHandler) UpdateOrganizationPreferences(ctx context.Context, req *oapi.UpdateOrganizationPreferencesRequest) (*oapi.UpdateOrganizationPreferencesResponse, error) {
	var resp oapi.UpdateOrganizationPreferencesResponse

	attrs := req.Body.Attributes
	prefs, prefsErr := h.orgs.SetPreferences(ctx, req.Id, func(m *ent.OrganizationPreferencesMutation) {
		if attrs.InitialSetupComplete != nil {
			m.SetInitialSetupAt(time.Now().UTC())
		}
		if attrs.EnableIncidentManagement != nil {
			m.SetEnableIncidentManagement(*attrs.EnableIncidentManagement)
		}
	})
	if prefsErr != nil {
		return nil, oapi.Error(ctx, "failed to update", prefsErr)
	}
	resp.Body.Data = oapi.OrganizationPreferencesFromEnt(prefs)

	return &resp, nil
}
