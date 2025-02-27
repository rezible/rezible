package fakeprovider

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
	"iter"
)

type SystemComponentsDataProvider struct {
	components []*ent.SystemComponent
}

type SystemComponentsDataProviderConfig struct{}

func NewSystemComponentsDataProvider(cfg SystemComponentsDataProviderConfig) (*SystemComponentsDataProvider, error) {
	p := &SystemComponentsDataProvider{
		components: makeFakeSystemComponents(),
	}

	return p, nil
}

func (p *SystemComponentsDataProvider) SystemComponentDataMapping() *ent.SystemComponent {
	return &systemComponentDataMapping
}

func (p *SystemComponentsDataProvider) PullSystemComponents(ctx context.Context) iter.Seq2[*ent.SystemComponent, error] {
	return func(yield func(i *ent.SystemComponent, err error) bool) {
		for _, inc := range p.components {
			yield(inc, nil)
		}
	}
}

// TODO convert this
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

func makeFakeSystemComponents() []*ent.SystemComponent {
	makeConstraint := func(label, desc string) *ent.SystemComponentConstraint {
		return &ent.SystemComponentConstraint{ID: uuid.New(), Label: label, Description: desc}
	}

	makeSignal := func(label string, desc string) *ent.SystemComponentSignal {
		return &ent.SystemComponentSignal{ID: uuid.New(), Label: label, Description: desc}
	}

	makeControl := func(label string, desc string) *ent.SystemComponentControl {
		return &ent.SystemComponentControl{ID: uuid.New(), Label: label, Description: desc}
	}

	makeKind := func(label string, desc string) *ent.SystemComponentKind {
		return &ent.SystemComponentKind{ID: uuid.New(), ProviderID: label, Label: label, Description: desc}
	}

	frontendKind := makeKind("frontend", "A frontend ui")
	serviceKind := makeKind("service", "A backend service")
	dbKind := makeKind("database", "A database")
	monitorKind := makeKind("monitor", "A monitor")
	kindExternal := makeKind("external", "An external entity")

	paymentUi := &ent.SystemComponent{
		ID:          uuid.New(),
		Name:        "Payment UI",
		ProviderID:  "payment-ui",
		Description: "The UI for handling payments",
		Edges: ent.SystemComponentEdges{
			Kind: frontendKind,
			Constraints: []*ent.SystemComponentConstraint{
				makeConstraint("Input Validated", "Must validate the users form input"),
				makeConstraint("Shows Error Feedback", "Must show feedback in the case of an error"),
			},
			Controls: []*ent.SystemComponentControl{
				makeControl("Input Validation", "validates input with Zod"),
			},
			Signals: []*ent.SystemComponentSignal{},
		},
	}

	apiGateway := &ent.SystemComponent{
		ID:          uuid.New(),
		Name:        "API Gateway",
		ProviderID:  "api-gateway",
		Description: "Handles incoming API requests",
		Edges: ent.SystemComponentEdges{
			Kind: serviceKind,
			Constraints: []*ent.SystemComponentConstraint{
				makeConstraint("Rate Limiting", "Rate limits requests to 1000 req/sec"),
				makeConstraint("Request Timeouts", "Enforces 30s timeout"),
			},
			Controls: []*ent.SystemComponentControl{
				makeControl("Request Throttling", "Configurable request throttling"),
				makeControl("Circuit Breaker", "Can trigger circuit breaker"),
			},
			Signals: []*ent.SystemComponentSignal{
				makeSignal("Validated Requests", "Requests allowed through gateway"),
				makeSignal("Request Errors", "Failed Requests"),
			},
		},
	}

	paymentSvc := &ent.SystemComponent{
		ID:          uuid.New(),
		Name:        "Payments Service",
		ProviderID:  "payment-service",
		Description: "Handles incoming API requests",
		Edges: ent.SystemComponentEdges{
			Kind: serviceKind,
			Constraints: []*ent.SystemComponentConstraint{
				makeConstraint("Time Limit", "Must process requests within 5s"),
				makeConstraint("Transactions Verified", "Must verify all transactions"),
			},
			Controls: []*ent.SystemComponentControl{
				makeControl("Transaction Verification", "Can verify transaction success"),
				makeControl("Retry Mechanism", "Can retry requests"),
			},
			Signals: []*ent.SystemComponentSignal{
				makeSignal("Transaction Records", "Completed transaction data"),
				makeSignal("Failed Payments", "Failed payment requests"),
			},
		},
	}

	db := &ent.SystemComponent{
		ID:          uuid.New(),
		Name:        "Payments Database",
		ProviderID:  "payment-database",
		Description: "PostgreSQL database",
		Edges: ent.SystemComponentEdges{
			Kind: dbKind,
			Constraints: []*ent.SystemComponentConstraint{
				makeConstraint("Connection Limit", "Max 100 connections"),
				makeConstraint("ACID properties", "Must maintain ACID compliance"),
			},
			Controls: []*ent.SystemComponentControl{
				makeControl("Connection Pooling", "configurable pool of connections"),
				makeControl("Transaction Management", "group operations in transaction"),
			},
			Signals: []*ent.SystemComponentSignal{
				makeSignal("Transaction Status", "result of transactions"),
				makeSignal("Connection Status", "state of connection"),
			},
		},
	}

	paymentsMonitor := &ent.SystemComponent{
		ID:          uuid.New(),
		Name:        "Payments Monitor",
		ProviderID:  "payment-monitor",
		Description: "A monitor using payments metrics",
		Edges: ent.SystemComponentEdges{
			Kind: monitorKind,
			Constraints: []*ent.SystemComponentConstraint{
				makeConstraint("Alerts Within 30s", "Must alert"),
				makeConstraint("Tracks all transaction", "Must track all transactions"),
			},
			Controls: []*ent.SystemComponentControl{
				makeControl("Alerting Configuration", "configurable alert rules"),
				makeControl("Metric Collection", "Collects metrics"),
			},
			Signals: []*ent.SystemComponentSignal{
				makeSignal("Alerts", "Alerts when rules met"),
			},
		},
	}

	extPaymentsProvider := &ent.SystemComponent{
		ID:          uuid.New(),
		Name:        "Payment Provider",
		ProviderID:  "payment-provider",
		Description: "Stripe",
		Edges: ent.SystemComponentEdges{
			Kind: kindExternal,
			Constraints: []*ent.SystemComponentConstraint{
				makeConstraint("Uptime SLA", "99.99%"),
				makeConstraint("Latency SLA", "2s response time"),
			},
			Controls: []*ent.SystemComponentControl{
				makeControl("Failover", "able to change provider"),
				makeControl("Health Checks", "scrape health status"),
			},
			Signals: []*ent.SystemComponentSignal{
				makeSignal("Transaction Results", "results of transaction"),
				makeSignal("Provider Status", "health status"),
			},
		},
	}

	return []*ent.SystemComponent{
		paymentUi,
		apiGateway,
		paymentSvc,
		paymentsMonitor,
		db,
		extPaymentsProvider,
	}
}
