package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/organization"
	oapi "github.com/rezible/rezible/openapi/v1"
)

type organizationsHandler struct {
	orgs rez.OrganizationService
}

func newOrganizationsHandler(orgs rez.OrganizationService) *organizationsHandler {
	return &organizationsHandler{orgs: orgs}
}

func (h *organizationsHandler) GetOrganization(ctx context.Context, request *oapi.GetOrganizationRequest) (*oapi.GetOrganizationResponse, error) {
	var resp oapi.GetOrganizationResponse

	org, orgErr := h.orgs.Get(ctx, organization.ID(request.Id))
	if orgErr != nil {
		return nil, oapi.Error("failed to fetch organization", orgErr)
	}
	resp.Body.Data = oapi.OrganizationFromEnt(org)

	return &resp, nil
}

func (h *organizationsHandler) FinishOrganizationSetup(ctx context.Context, request *oapi.FinishOrganizationSetupRequest) (*oapi.FinishOrganizationSetupResponse, error) {
	var resp oapi.FinishOrganizationSetupResponse

	org, orgErr := h.orgs.Get(ctx, organization.ID(request.Id))
	if orgErr != nil {
		return nil, oapi.Error("failed to fetch organization", orgErr)
	}

	completeErr := h.orgs.CompleteSetup(ctx, org)
	if completeErr != nil {
		return nil, oapi.Error("failed to finish setup", completeErr)
	}

	return &resp, nil
}
