package apiv1

import (
	"context"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type investigationsHandler struct {
	investigations rez.InvestigationService
}

func newInvestigationsHandler(investigations rez.InvestigationService) *investigationsHandler {
	return &investigationsHandler{investigations: investigations}
}

func (h *investigationsHandler) GetInvestigation(ctx context.Context, request *oapi.GetInvestigationRequest) (*oapi.GetInvestigationResponse, error) {
	var response oapi.GetInvestigationResponse
	investigation, getErr := h.investigations.GetInvestigation(ctx, request.Id)
	if getErr != nil {
		return nil, oapi.Error(ctx, "get investigation", getErr)
	}
	response.Body.Data = oapi.InvestigationFromEnt(investigation)
	return &response, nil
}
