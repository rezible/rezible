package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
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

	org, orgErr := h.orgs.GetById(ctx, request.Id)
	if orgErr != nil {
		return nil, apiError("failed to fetch organization", orgErr)
	}
	resp.Body.Data = oapi.OrganizationFromEnt(org)

	return &resp, nil
}

func (h *organizationsHandler) FinishOrganizationSetup(ctx context.Context, request *oapi.FinishOrganizationSetupRequest) (*oapi.FinishOrganizationSetupResponse, error) {
	var resp oapi.FinishOrganizationSetupResponse

	completeErr := h.orgs.CompleteSetup(ctx, request.Id)
	if completeErr != nil {
		return nil, apiError("failed to finish setup", completeErr)
	}

	return &resp, nil
}
