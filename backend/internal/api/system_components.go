package api

import (
	"context"
	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/openapi"
)

type systemComponentsHandler struct {
}

func newSystemComponentsHandler() *systemComponentsHandler {
	return &systemComponentsHandler{}
}

func (h *systemComponentsHandler) ListSystemComponents(ctx context.Context, request *oapi.ListSystemComponentsRequest) (*oapi.ListSystemComponentsResponse, error) {
	var resp oapi.ListSystemComponentsResponse

	return &resp, nil
}

func (h *systemComponentsHandler) ListIncidentSystemComponents(ctx context.Context, request *oapi.ListIncidentSystemComponentsRequest) (*oapi.ListIncidentSystemComponentsResponse, error) {
	var resp oapi.ListIncidentSystemComponentsResponse

	cmp1Id := uuid.New()
	cmp2Id := uuid.New()

	controlRel := oapi.SystemComponentRelationship{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentRelationshipAttributes{
			Kind: "control",
			Details: oapi.SystemComponentControlRelationshipDetails{
				ControllerId: cmp1Id,
				ControlledId: cmp2Id,
				Control:      "rate_limits",
				Description:  "Rate limits service",
			},
		},
	}

	feedbackRel := oapi.SystemComponentRelationship{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentRelationshipAttributes{
			Kind: "feedback",
			Details: oapi.SystemComponentFeedbackRelationshipDetails{
				SourceId:    cmp2Id,
				TargetId:    cmp1Id,
				Feedback:    "metrics",
				Description: "Provides usage metrics",
			},
		},
	}

	rels := []oapi.SystemComponentRelationship{controlRel, feedbackRel}

	cmp1 := oapi.SystemComponent{
		Id: cmp1Id,
		Attributes: oapi.SystemComponentAttributes{
			Name:          "API Service",
			Kind:          "service",
			Description:   "an api service",
			Properties:    nil,
			Relationships: rels,
		},
	}

	cmp2 := oapi.SystemComponent{
		Id: cmp2Id,
		Attributes: oapi.SystemComponentAttributes{
			Name:          "Other Service",
			Kind:          "service",
			Description:   "another api service",
			Properties:    nil,
			Relationships: rels,
		},
	}

	components := []oapi.IncidentSystemComponent{
		{
			Id: uuid.New(),
			Attributes: oapi.IncidentSystemComponentAttributes{
				Role:      "primary",
				Component: cmp1,
			},
		},
		{
			Id: uuid.New(),
			Attributes: oapi.IncidentSystemComponentAttributes{
				Role:      "primary",
				Component: cmp2,
			},
		},
	}
	resp.Body.Data = components

	return &resp, nil
}

func (h *systemComponentsHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) UpdateSystemComponent(ctx context.Context, request *oapi.UpdateSystemComponentRequest) (*oapi.UpdateSystemComponentResponse, error) {
	var resp oapi.UpdateSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	return &resp, nil
}

func (h *systemComponentsHandler) ListSystemComponentRelationships(ctx context.Context, request *oapi.ListSystemComponentRelationshipsRequest) (*oapi.ListSystemComponentRelationshipsResponse, error) {
	var resp oapi.ListSystemComponentRelationshipsResponse

	return &resp, nil
}

func (h *systemComponentsHandler) CreateSystemComponentRelationship(ctx context.Context, request *oapi.CreateSystemComponentRelationshipRequest) (*oapi.CreateSystemComponentRelationshipResponse, error) {
	var resp oapi.CreateSystemComponentRelationshipResponse

	return &resp, nil
}

func (h *systemComponentsHandler) UpdateSystemComponentRelationship(ctx context.Context, request *oapi.UpdateSystemComponentRelationshipRequest) (*oapi.UpdateSystemComponentRelationshipResponse, error) {
	var resp oapi.UpdateSystemComponentRelationshipResponse

	return &resp, nil
}

func (h *systemComponentsHandler) ArchiveSystemComponentRelationship(ctx context.Context, request *oapi.ArchiveSystemComponentRelationshipRequest) (*oapi.ArchiveSystemComponentRelationshipResponse, error) {
	var resp oapi.ArchiveSystemComponentRelationshipResponse

	return &resp, nil
}
