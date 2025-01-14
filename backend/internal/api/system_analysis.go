package api

import (
	"context"
	oapi "github.com/rezible/rezible/openapi"
)

type systemAnalysisHandler struct {
}

func newSystemAnalysisHandler() *systemAnalysisHandler {
	return &systemAnalysisHandler{}
}

func (s systemAnalysisHandler) ListSystemComponents(ctx context.Context, request *oapi.ListSystemComponentsRequest) (*oapi.ListSystemComponentsResponse, error) {
	var resp oapi.ListSystemComponentsResponse

	return &resp, nil
}

func (s systemAnalysisHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) UpdateSystemComponent(ctx context.Context, request *oapi.UpdateSystemComponentRequest) (*oapi.UpdateSystemComponentResponse, error) {
	var resp oapi.UpdateSystemComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	var resp oapi.GetSystemAnalysisResponse

	return &resp, nil
}

func (s systemAnalysisHandler) ListSystemAnalysisComponents(ctx context.Context, request *oapi.ListSystemAnalysisComponentsRequest) (*oapi.ListSystemAnalysisComponentsResponse, error) {
	var resp oapi.ListSystemAnalysisComponentsResponse

	return &resp, nil
}

func (s systemAnalysisHandler) AddSystemAnalysisComponent(ctx context.Context, request *oapi.AddSystemAnalysisComponentRequest) (*oapi.AddSystemAnalysisComponentResponse, error) {
	var resp oapi.AddSystemAnalysisComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) GetSystemAnalysisComponent(ctx context.Context, request *oapi.GetSystemAnalysisComponentRequest) (*oapi.GetSystemAnalysisComponentResponse, error) {
	var resp oapi.GetSystemAnalysisComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) UpdateSystemAnalysisComponent(ctx context.Context, request *oapi.UpdateSystemAnalysisComponentRequest) (*oapi.UpdateSystemAnalysisComponentResponse, error) {
	var resp oapi.UpdateSystemAnalysisComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) DeleteSystemAnalysisComponent(ctx context.Context, request *oapi.DeleteSystemAnalysisComponentRequest) (*oapi.DeleteSystemAnalysisComponentResponse, error) {
	var resp oapi.DeleteSystemAnalysisComponentResponse

	return &resp, nil
}

func (s systemAnalysisHandler) ListSystemAnalysisRelationships(ctx context.Context, request *oapi.ListSystemAnalysisRelationshipsRequest) (*oapi.ListSystemAnalysisRelationshipsResponse, error) {
	var resp oapi.ListSystemAnalysisRelationshipsResponse

	return &resp, nil
}

func (s systemAnalysisHandler) CreateSystemAnalysisRelationship(ctx context.Context, request *oapi.CreateSystemAnalysisRelationshipRequest) (*oapi.CreateSystemAnalysisRelationshipResponse, error) {
	var resp oapi.CreateSystemAnalysisRelationshipResponse

	return &resp, nil
}

func (s systemAnalysisHandler) GetSystemAnalysisRelationship(ctx context.Context, request *oapi.GetSystemAnalysisRelationshipRequest) (*oapi.GetSystemAnalysisRelationshipResponse, error) {
	var resp oapi.GetSystemAnalysisRelationshipResponse

	return &resp, nil
}

func (s systemAnalysisHandler) UpdateSystemAnalysisRelationship(ctx context.Context, request *oapi.UpdateSystemAnalysisRelationshipRequest) (*oapi.UpdateSystemAnalysisRelationshipResponse, error) {
	var resp oapi.UpdateSystemAnalysisRelationshipResponse

	return &resp, nil
}

func (s systemAnalysisHandler) DeleteSystemAnalysisRelationship(ctx context.Context, request *oapi.DeleteSystemAnalysisRelationshipRequest) (*oapi.DeleteSystemAnalysisRelationshipResponse, error) {
	var resp oapi.DeleteSystemAnalysisRelationshipResponse

	return &resp, nil
}
