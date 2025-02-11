package api

import (
	"context"
	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/openapi"
)

type systemAnalysisHandler struct {
	fakeComponents []oapi.SystemComponent
	fakeAnalysis   oapi.SystemAnalysis
}

func newSystemAnalysisHandler() *systemAnalysisHandler {
	fakeComponents := makeFakeComponents()
	return &systemAnalysisHandler{
		fakeComponents: fakeComponents,
		fakeAnalysis:   makeFakeSystemAnalysis(fakeComponents),
	}
}

func makeFakeComponents() []oapi.SystemComponent {
	makeConstraint := func(label, desc string) oapi.SystemComponentConstraint {
		attr := oapi.SystemComponentConstraintAttributes{Label: label, Description: desc}
		return oapi.SystemComponentConstraint{Id: uuid.New(), Attributes: attr}
	}

	makeSignal := func(label string, desc string) oapi.SystemComponentSignal {
		attr := oapi.SystemComponentSignalAttributes{Label: label, Description: desc}
		return oapi.SystemComponentSignal{Id: uuid.New(), Attributes: attr}
	}

	makeControl := func(label string, desc string) oapi.SystemComponentControl {
		attr := oapi.SystemComponentControlAttributes{Label: label, Description: desc}
		return oapi.SystemComponentControl{Id: uuid.New(), Attributes: attr}
	}

	paymentUi := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "Payment UI",
			Kind:        "frontend",
			Description: "The UI for handling payments",
			Constraints: []oapi.SystemComponentConstraint{
				makeConstraint("Input Validated", "Must validate the users form input"),
				makeConstraint("Shows Error Feedback", "Must show feedback in the case of an error"),
			},
			Controls: []oapi.SystemComponentControl{
				makeControl("Input Validation", "validates input with Zod"),
			},
			Signals: []oapi.SystemComponentSignal{},
		},
	}

	apiGateway := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "API Gateway",
			Kind:        "service",
			Description: "Handles incoming API requests",
			Constraints: []oapi.SystemComponentConstraint{
				makeConstraint("Rate Limiting", "Rate limits requests to 1000 req/sec"),
				makeConstraint("Request Timeouts", "Enforces 30s timeout"),
			},
			Controls: []oapi.SystemComponentControl{
				makeControl("Request Throttling", "Configurable request throttling"),
				makeControl("Circuit Breaker", "Can trigger circuit breaker"),
			},
			Signals: []oapi.SystemComponentSignal{
				makeSignal("Validated Requests", "Requests allowed through gateway"),
				makeSignal("Request Errors", "Failed Requests"),
			},
		},
	}

	paymentSvc := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "Payments Service",
			Kind:        "service",
			Description: "Handles incoming API requests",
			Constraints: []oapi.SystemComponentConstraint{
				makeConstraint("Time Limit", "Must process requests within 5s"),
				makeConstraint("Transactions Verified", "Must verify all transactions"),
			},
			Controls: []oapi.SystemComponentControl{
				makeControl("Transaction Verification", "Can verify transaction success"),
				makeControl("Retry Mechanism", "Can retry requests"),
			},
			Signals: []oapi.SystemComponentSignal{
				makeSignal("Transaction Records", "Completed transaction data"),
				makeSignal("Failed Payments", "Failed payment requests"),
			},
		},
	}

	db := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "Payments Database",
			Kind:        "database",
			Description: "RDS PostgreSQL database",
			Constraints: []oapi.SystemComponentConstraint{
				makeConstraint("Connection Limit", "Max 100 connections"),
				makeConstraint("ACID properties", "Must maintain ACID compliance"),
			},
			Controls: []oapi.SystemComponentControl{
				makeControl("Connection Pooling", "configurable pool of connections"),
				makeControl("Transaction Management", "group operations in transaction"),
			},
			Signals: []oapi.SystemComponentSignal{
				makeSignal("Transaction Status", "result of transactions"),
				makeSignal("Connection Status", "state of connection"),
			},
		},
	}

	paymentsMonitor := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "Payments Monitor",
			Kind:        "monitor",
			Description: "A monitor using payments metrics",
			Constraints: []oapi.SystemComponentConstraint{
				makeConstraint("Alerts Within 30s", "Must alert"),
				makeConstraint("Tracks all transaction", "Must track all transactions"),
			},
			Controls: []oapi.SystemComponentControl{
				makeControl("Alerting Configuration", "configurable alert rules"),
				makeControl("Metric Collection", "Collects metrics"),
			},
			Signals: []oapi.SystemComponentSignal{
				makeSignal("Alerts", "Alerts when rules met"),
			},
		},
	}

	extPaymentsProvider := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "External Payments Provider",
			Kind:        "external",
			Description: "Stripe",
			Constraints: []oapi.SystemComponentConstraint{
				makeConstraint("Uptime SLA", "99.99%"),
				makeConstraint("Latency SLA", "2s response time"),
			},
			Controls: []oapi.SystemComponentControl{
				makeControl("Failover", "able to change provider"),
				makeControl("Health Checks", "scrape health status"),
			},
			Signals: []oapi.SystemComponentSignal{
				makeSignal("Transaction Results", "results of transaction"),
				makeSignal("Provider Status", "health status"),
			},
		},
	}

	return []oapi.SystemComponent{
		paymentUi,
		apiGateway,
		paymentSvc,
		paymentsMonitor,
		db,
		extPaymentsProvider,
	}
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

func (s *systemAnalysisHandler) ListSystemComponents(ctx context.Context, request *oapi.ListSystemComponentsRequest) (*oapi.ListSystemComponentsResponse, error) {
	var resp oapi.ListSystemComponentsResponse

	resp.Body.Data = s.fakeComponents

	return &resp, nil
}

func (s *systemAnalysisHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	newComponent := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        request.Body.Attributes.Name,
			Kind:        "",
			Description: "",
			Properties:  nil,
			Constraints: nil,
			Signals:     nil,
			Controls:    nil,
		},
	}
	s.fakeComponents = append(s.fakeComponents, newComponent)

	resp.Body.Data = newComponent

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	idx := -1
	for i, cmp := range s.fakeComponents {
		if cmp.Id == request.Id {
			idx = i
		}
	}
	if idx == -1 {
		return nil, oapi.ErrorNotFound("not found")
	}
	resp.Body.Data = s.fakeComponents[idx]

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemComponent(ctx context.Context, request *oapi.UpdateSystemComponentRequest) (*oapi.UpdateSystemComponentResponse, error) {
	var resp oapi.UpdateSystemComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysis(ctx context.Context, request *oapi.GetSystemAnalysisRequest) (*oapi.GetSystemAnalysisResponse, error) {
	var resp oapi.GetSystemAnalysisResponse

	resp.Body.Data = s.fakeAnalysis

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisComponents(ctx context.Context, request *oapi.ListSystemAnalysisComponentsRequest) (*oapi.ListSystemAnalysisComponentsResponse, error) {
	var resp oapi.ListSystemAnalysisComponentsResponse

	resp.Body.Data = s.fakeAnalysis.Attributes.Components

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
