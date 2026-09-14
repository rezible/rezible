package apiv1

import (
	"context"

	oapi "github.com/rezible/rezible/pkg/openapi/v1"
)

type activityHandler struct {
}

func newActivityHandler() *activityHandler {
	return &activityHandler{}
}

func (h *activityHandler) ListActivity(ctx context.Context, request *oapi.ListActivityRequest) (*oapi.ListActivityResponse, error) {
	var response oapi.ListActivityResponse
	response.Body.Data = make([]oapi.ActivityRecord, 0)
	response.Body.Pagination = oapi.Pagination{Page: 1, PageSize: request.PageSize, Total: 0}
	return &response, nil
}
