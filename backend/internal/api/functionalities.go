package api

import (
	"context"
	oapi "github.com/twohundreds/rezible/openapi"
)

type functionalitiesHandler struct {
}

func newFunctionalitiesHandler() *functionalitiesHandler {
	return &functionalitiesHandler{}
}

func (h *functionalitiesHandler) ListFunctionalities(ctx context.Context, request *oapi.ListFunctionalitiesRequest) (*oapi.ListFunctionalitiesResponse, error) {
	var resp oapi.ListFunctionalitiesResponse

	return &resp, nil
}

func (h *functionalitiesHandler) CreateFunctionality(ctx context.Context, request *oapi.CreateFunctionalityRequest) (*oapi.CreateFunctionalityResponse, error) {
	var resp oapi.CreateFunctionalityResponse

	return &resp, nil
}

func (h *functionalitiesHandler) GetFunctionality(ctx context.Context, request *oapi.GetFunctionalityRequest) (*oapi.GetFunctionalityResponse, error) {
	var resp oapi.GetFunctionalityResponse

	return &resp, nil
}

func (h *functionalitiesHandler) UpdateFunctionality(ctx context.Context, request *oapi.UpdateFunctionalityRequest) (*oapi.UpdateFunctionalityResponse, error) {
	var resp oapi.UpdateFunctionalityResponse

	return &resp, nil
}

func (h *functionalitiesHandler) ArchiveFunctionality(ctx context.Context, request *oapi.ArchiveFunctionalityRequest) (*oapi.ArchiveFunctionalityResponse, error) {
	var resp oapi.ArchiveFunctionalityResponse

	return &resp, nil
}
