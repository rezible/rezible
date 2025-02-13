package api

import (
	"context"
	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/openapi"
)

var fakeAnalysis = makeFakeSystemAnalysis(fakeComponents)

type systemAnalysisHandler struct {
}

func newSystemAnalysisHandler() *systemAnalysisHandler {
	return &systemAnalysisHandler{}
}

func makeFakeSystemAnalysis(cmps []oapi.SystemComponent) oapi.SystemAnalysis {
	position := func(x, y float64) oapi.SystemAnalysisDiagramPosition {
		return oapi.SystemAnalysisDiagramPosition{X: x, Y: y}
	}

	relSignal := func(id uuid.UUID, desc string) *oapi.SystemAnalysisRelationshipFeedbackSignal {
		attr := oapi.SystemAnalysisRelationshipFeedbackSignalAttributes{SignalId: id, Description: desc}
		return &oapi.SystemAnalysisRelationshipFeedbackSignal{Id: uuid.New(), Attributes: attr}
	}

	relControl := func(id uuid.UUID, desc string) *oapi.SystemAnalysisRelationshipControlAction {
		attr := oapi.SystemAnalysisRelationshipControlActionAttributes{ControlId: id, Description: desc}
		return &oapi.SystemAnalysisRelationshipControlAction{Id: uuid.New(), Attributes: attr}
	}

	makeAnalysisComponent := func(cmp oapi.SystemComponent, pos oapi.SystemAnalysisDiagramPosition) oapi.SystemAnalysisComponent {
		attr := oapi.SystemAnalysisComponentAttributes{Component: cmp, Position: pos}
		return oapi.SystemAnalysisComponent{Id: uuid.New(), Attributes: attr}
	}

	makeRelationship := func(sId, tId uuid.UUID, feedback *oapi.SystemAnalysisRelationshipFeedbackSignal, control *oapi.SystemAnalysisRelationshipControlAction) oapi.SystemAnalysisRelationship {
		attr := oapi.SystemAnalysisRelationshipAttributes{
			SourceId:        sId,
			TargetId:        tId,
			Description:     "description",
			FeedbackSignals: make([]oapi.SystemAnalysisRelationshipFeedbackSignal, 0, 1),
			ControlActions:  make([]oapi.SystemAnalysisRelationshipControlAction, 0, 1),
		}
		if feedback != nil {
			attr.FeedbackSignals = append(attr.FeedbackSignals, *feedback)
		}
		if control != nil {
			attr.ControlActions = append(attr.ControlActions, *control)
		}
		return oapi.SystemAnalysisRelationship{Id: uuid.New(), Attributes: attr}
	}

	paymentUi := cmps[0]
	apiGateway := cmps[1]
	paymentSvc := cmps[2]
	paymentsMonitor := cmps[3]
	db := cmps[4]
	extPaymentsProvider := cmps[5]

	components := []oapi.SystemAnalysisComponent{
		makeAnalysisComponent(paymentUi, position(0, 0)),
		makeAnalysisComponent(apiGateway, position(200, 100)),
		makeAnalysisComponent(paymentSvc, position(400, 200)),
		makeAnalysisComponent(paymentsMonitor, position(600, 300)),
		makeAnalysisComponent(db, position(600, 100)),
		makeAnalysisComponent(extPaymentsProvider, position(700, 200)),
	}

	feApiErrsSignal := relSignal(apiGateway.Attributes.Signals[0].Id, "api errors are returned")
	feThrottleControl := relControl(apiGateway.Attributes.Controls[0].Id, "frontend can be throttled")
	relationships := []oapi.SystemAnalysisRelationship{
		makeRelationship(paymentUi.Id, apiGateway.Id, feApiErrsSignal, feThrottleControl),
		makeRelationship(apiGateway.Id, paymentSvc.Id, nil, nil),
		makeRelationship(paymentSvc.Id, db.Id, nil, nil),
		makeRelationship(paymentSvc.Id, paymentsMonitor.Id, nil, nil),
		makeRelationship(paymentSvc.Id, extPaymentsProvider.Id, nil, nil),
	}

	return oapi.SystemAnalysis{
		Id: uuid.New(),
		Attributes: oapi.SystemAnalysisAttributes{
			Components:    components,
			Relationships: relationships,
		},
	}
}

func (s *systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	var resp oapi.GetSystemAnalysisResponse

	resp.Body.Data = fakeAnalysis

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisComponents(ctx context.Context, request *oapi.ListSystemAnalysisComponentsRequest) (*oapi.ListSystemAnalysisComponentsResponse, error) {
	var resp oapi.ListSystemAnalysisComponentsResponse

	resp.Body.Data = fakeAnalysis.Attributes.Components

	return &resp, nil
}

func (s *systemAnalysisHandler) AddSystemAnalysisComponent(ctx context.Context, request *oapi.AddSystemAnalysisComponentRequest) (*oapi.AddSystemAnalysisComponentResponse, error) {
	var resp oapi.AddSystemAnalysisComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisComponent(ctx context.Context, request *oapi.GetSystemAnalysisComponentRequest) (*oapi.GetSystemAnalysisComponentResponse, error) {
	var resp oapi.GetSystemAnalysisComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisComponent(ctx context.Context, request *oapi.UpdateSystemAnalysisComponentRequest) (*oapi.UpdateSystemAnalysisComponentResponse, error) {
	var resp oapi.UpdateSystemAnalysisComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisComponent(ctx context.Context, request *oapi.DeleteSystemAnalysisComponentRequest) (*oapi.DeleteSystemAnalysisComponentResponse, error) {
	var resp oapi.DeleteSystemAnalysisComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisRelationships(ctx context.Context, request *oapi.ListSystemAnalysisRelationshipsRequest) (*oapi.ListSystemAnalysisRelationshipsResponse, error) {
	var resp oapi.ListSystemAnalysisRelationshipsResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) CreateSystemAnalysisRelationship(ctx context.Context, request *oapi.CreateSystemAnalysisRelationshipRequest) (*oapi.CreateSystemAnalysisRelationshipResponse, error) {
	var resp oapi.CreateSystemAnalysisRelationshipResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisRelationship(ctx context.Context, request *oapi.GetSystemAnalysisRelationshipRequest) (*oapi.GetSystemAnalysisRelationshipResponse, error) {
	var resp oapi.GetSystemAnalysisRelationshipResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisRelationship(ctx context.Context, request *oapi.UpdateSystemAnalysisRelationshipRequest) (*oapi.UpdateSystemAnalysisRelationshipResponse, error) {
	var resp oapi.UpdateSystemAnalysisRelationshipResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisRelationship(ctx context.Context, request *oapi.DeleteSystemAnalysisRelationshipRequest) (*oapi.DeleteSystemAnalysisRelationshipResponse, error) {
	var resp oapi.DeleteSystemAnalysisRelationshipResponse

	return &resp, nil
}
