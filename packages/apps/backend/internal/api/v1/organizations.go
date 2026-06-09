package apiv1

import (
	"context"

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

func (h *organizationsHandler) UpdateOrganization(ctx context.Context, req *oapi.UpdateOrganizationRequest) (*oapi.UpdateOrganizationResponse, error) {
	var resp oapi.UpdateOrganizationResponse

	attrs := req.Body.Attributes

	var prefs *ent.OrganizationPreferences
	if attrs.Preferences != nil {
		prefs = &ent.OrganizationPreferences{
			EnableIncidentManagement: attrs.Preferences.EnableIncidentManagement,
		}
	}

	org, orgErr := h.orgs.Get(ctx, organization.ID(req.Id))
	if orgErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch organization", orgErr)
	}
	update := org.Update().
		SetNillableName(attrs.Name).
		SetNillablePreferences(prefs)
	org, orgErr = update.Save(ctx)
	if orgErr != nil {
		return nil, oapi.Error(ctx, "failed to update organization", orgErr)
	}
	resp.Body.Data = oapi.OrganizationFromEnt(org)

	return &resp, nil
}

func (h *organizationsHandler) FinishOrganizationSetup(ctx context.Context, req *oapi.FinishOrganizationSetupRequest) (*oapi.FinishOrganizationSetupResponse, error) {
	var resp oapi.FinishOrganizationSetupResponse

	org, orgErr := h.orgs.Get(ctx, organization.ID(req.Id))
	if orgErr != nil {
		return nil, oapi.Error(ctx, "failed to fetch organization", orgErr)
	}

	completeErr := h.orgs.CompleteSetup(ctx, org)
	if completeErr != nil {
		return nil, oapi.Error(ctx, "failed to finish setup", completeErr)
	}

	return &resp, nil
}
