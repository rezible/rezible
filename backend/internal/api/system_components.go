package api

import (
	"context"
	"github.com/google/uuid"
	oapi "github.com/rezible/rezible/openapi"
)

type systemComponentsHandler struct {
}

var fakeComponents = makeFakeComponents()

func newSystemComponentsHandler() *systemComponentsHandler {
	return &systemComponentsHandler{}
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

	makeKind := func(label string, desc string) oapi.SystemComponentKind {
		attr := oapi.SystemComponentKindAttributes{Label: label, Description: desc}
		return oapi.SystemComponentKind{Id: uuid.New(), Attributes: attr}
	}

	paymentUi := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "Payment UI",
			Kind:        makeKind("frontend", "A frontend ui"),
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

	serviceKind := makeKind("service", "A backend service")

	apiGateway := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        "API Gateway",
			Kind:        serviceKind,
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
			Kind:        serviceKind,
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
			Kind:        makeKind("database", "A database"),
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
			Kind:        makeKind("monitor", "A monitor"),
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
			Kind:        makeKind("external", "An external entity"),
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

func (s *systemComponentsHandler) ListSystemComponents(ctx context.Context, request *oapi.ListSystemComponentsRequest) (*oapi.ListSystemComponentsResponse, error) {
	var resp oapi.ListSystemComponentsResponse

	resp.Body.Data = fakeComponents

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	kind := oapi.SystemComponentKind{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentKindAttributes{
			Label:       "",
			Description: "",
		},
	}

	newComponent := oapi.SystemComponent{
		Id: uuid.New(),
		Attributes: oapi.SystemComponentAttributes{
			Name:        request.Body.Attributes.Name,
			Kind:        kind,
			Description: "",
			Properties:  nil,
			Constraints: nil,
			Signals:     nil,
			Controls:    nil,
		},
	}
	fakeComponents = append(fakeComponents, newComponent)

	resp.Body.Data = newComponent

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	idx := -1
	for i, cmp := range fakeComponents {
		if cmp.Id == request.Id {
			idx = i
		}
	}
	if idx == -1 {
		return nil, oapi.ErrorNotFound("not found")
	}
	resp.Body.Data = fakeComponents[idx]

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponent(ctx context.Context, request *oapi.UpdateSystemComponentRequest) (*oapi.UpdateSystemComponentResponse, error) {
	var resp oapi.UpdateSystemComponentResponse

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	return &resp, nil
}

func (s *systemComponentsHandler) ListSystemComponentKinds(ctx context.Context, request *oapi.ListSystemComponentKindsRequest) (*oapi.ListSystemComponentKindsResponse, error) {
	var resp oapi.ListSystemComponentKindsResponse

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentKind(ctx context.Context, request *oapi.CreateSystemComponentKindRequest) (*oapi.CreateSystemComponentKindResponse, error) {
	var resp oapi.CreateSystemComponentKindResponse

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentKind(ctx context.Context, request *oapi.GetSystemComponentKindRequest) (*oapi.GetSystemComponentKindResponse, error) {
	var resp oapi.GetSystemComponentKindResponse

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentKind(ctx context.Context, request *oapi.UpdateSystemComponentKindRequest) (*oapi.UpdateSystemComponentKindResponse, error) {
	var resp oapi.UpdateSystemComponentKindResponse

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentKind(ctx context.Context, request *oapi.ArchiveSystemComponentKindRequest) (*oapi.ArchiveSystemComponentKindResponse, error) {
	var resp oapi.ArchiveSystemComponentKindResponse

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentConstraint(ctx context.Context, request *oapi.CreateSystemComponentConstraintRequest) (*oapi.CreateSystemComponentConstraintResponse, error) {
	var resp oapi.CreateSystemComponentConstraintResponse

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentConstraint(ctx context.Context, request *oapi.GetSystemComponentConstraintRequest) (*oapi.GetSystemComponentConstraintResponse, error) {
	var resp oapi.GetSystemComponentConstraintResponse

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentConstraint(ctx context.Context, request *oapi.UpdateSystemComponentConstraintRequest) (*oapi.UpdateSystemComponentConstraintResponse, error) {
	var resp oapi.UpdateSystemComponentConstraintResponse

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentConstraint(ctx context.Context, request *oapi.ArchiveSystemComponentConstraintRequest) (*oapi.ArchiveSystemComponentConstraintResponse, error) {
	var resp oapi.ArchiveSystemComponentConstraintResponse

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentControl(ctx context.Context, request *oapi.CreateSystemComponentControlRequest) (*oapi.CreateSystemComponentControlResponse, error) {
	var resp oapi.CreateSystemComponentControlResponse

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentControl(ctx context.Context, request *oapi.GetSystemComponentControlRequest) (*oapi.GetSystemComponentControlResponse, error) {
	var resp oapi.GetSystemComponentControlResponse

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentControl(ctx context.Context, request *oapi.UpdateSystemComponentControlRequest) (*oapi.UpdateSystemComponentControlResponse, error) {
	var resp oapi.UpdateSystemComponentControlResponse

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentControl(ctx context.Context, request *oapi.ArchiveSystemComponentControlRequest) (*oapi.ArchiveSystemComponentControlResponse, error) {
	var resp oapi.ArchiveSystemComponentControlResponse

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentSignal(ctx context.Context, request *oapi.CreateSystemComponentSignalRequest) (*oapi.CreateSystemComponentSignalResponse, error) {
	var resp oapi.CreateSystemComponentSignalResponse

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentSignal(ctx context.Context, request *oapi.GetSystemComponentSignalRequest) (*oapi.GetSystemComponentSignalResponse, error) {
	var resp oapi.GetSystemComponentSignalResponse

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentSignal(ctx context.Context, request *oapi.UpdateSystemComponentSignalRequest) (*oapi.UpdateSystemComponentSignalResponse, error) {
	var resp oapi.UpdateSystemComponentSignalResponse

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentSignal(ctx context.Context, request *oapi.ArchiveSystemComponentSignalRequest) (*oapi.ArchiveSystemComponentSignalResponse, error) {
	var resp oapi.ArchiveSystemComponentSignalResponse

	return &resp, nil
}
