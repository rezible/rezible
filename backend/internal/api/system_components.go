package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

type systemComponentsHandler struct {
	db *ent.Client
}

var fakeComponents = makeFakeComponents()

func newSystemComponentsHandler(db *ent.Client) *systemComponentsHandler {
	return &systemComponentsHandler{db: db}
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

	query := s.db.SystemComponent.Query()
	// TODO ListParams

	cmps, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query system components", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemComponent, len(cmps))
	for i, cmp := range cmps {
		resp.Body.Data[i] = oapi.SystemComponentFromEnt(cmp)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	attr := request.Body.Attributes

	var created *ent.SystemComponent

	createComponentTx := func(tx *ent.Tx) error {
		create := tx.SystemComponent.Create().
			SetName(attr.Name).
			SetDescription(attr.Description).
			SetKindID(attr.KindId)

		cmp, createErr := create.Save(ctx)
		if createErr != nil {
			return createErr
		}

		// TODO: controls, constraints & signals

		created = cmp

		return nil
	}

	if createErr := ent.WithTx(ctx, s.db, createComponentTx); createErr != nil {
		return nil, detailError("failed to create system component", createErr)
	}
	resp.Body.Data = oapi.SystemComponentFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	cmp, queryErr := s.db.SystemComponent.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("failed to query system component", queryErr)
	}
	resp.Body.Data = oapi.SystemComponentFromEnt(cmp)

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponent(ctx context.Context, request *oapi.UpdateSystemComponentRequest) (*oapi.UpdateSystemComponentResponse, error) {
	var resp oapi.UpdateSystemComponentResponse

	attr := request.Body.Attributes

	update := s.db.SystemComponent.UpdateOneID(request.Id).
		SetNillableName(attr.Name).
		SetNillableDescription(attr.Description).
		SetNillableKindID(attr.KindId)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update system component", updateErr)
	}

	resp.Body.Data = oapi.SystemComponentFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	if delErr := s.db.SystemComponent.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to delete system component", delErr)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) ListSystemComponentKinds(ctx context.Context, request *oapi.ListSystemComponentKindsRequest) (*oapi.ListSystemComponentKindsResponse, error) {
	var resp oapi.ListSystemComponentKindsResponse

	query := s.db.SystemComponentKind.Query()
	// TODO ListParams

	kinds, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query system component kinds", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemComponentKind, len(kinds))
	for i, kind := range kinds {
		resp.Body.Data[i] = oapi.SystemComponentKindFromEnt(kind)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentKind(ctx context.Context, request *oapi.CreateSystemComponentKindRequest) (*oapi.CreateSystemComponentKindResponse, error) {
	var resp oapi.CreateSystemComponentKindResponse

	attr := request.Body.Attributes
	create := s.db.SystemComponentKind.Create().
		SetLabel(attr.Label).
		SetDescription(attr.Description)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create system component kind", createErr)
	}
	resp.Body.Data = oapi.SystemComponentKindFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentKind(ctx context.Context, request *oapi.GetSystemComponentKindRequest) (*oapi.GetSystemComponentKindResponse, error) {
	var resp oapi.GetSystemComponentKindResponse

	kind, queryErr := s.db.SystemComponentKind.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("failed to query system component kind", queryErr)
	}
	resp.Body.Data = oapi.SystemComponentKindFromEnt(kind)

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentKind(ctx context.Context, request *oapi.UpdateSystemComponentKindRequest) (*oapi.UpdateSystemComponentKindResponse, error) {
	var resp oapi.UpdateSystemComponentKindResponse

	attr := request.Body.Attributes
	update := s.db.SystemComponentKind.UpdateOneID(request.Id).
		SetNillableLabel(attr.Label).
		SetNillableDescription(attr.Description)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update system component kind", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentKindFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentKind(ctx context.Context, request *oapi.ArchiveSystemComponentKindRequest) (*oapi.ArchiveSystemComponentKindResponse, error) {
	var resp oapi.ArchiveSystemComponentKindResponse

	if delErr := s.db.SystemComponentKind.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to archive system component kind", delErr)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentConstraint(ctx context.Context, request *oapi.CreateSystemComponentConstraintRequest) (*oapi.CreateSystemComponentConstraintResponse, error) {
	var resp oapi.CreateSystemComponentConstraintResponse

	attr := request.Body.Attributes
	create := s.db.SystemComponentConstraint.Create().
		SetLabel(attr.Label).
		SetDescription(attr.Description)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create system component constraint", createErr)
	}
	resp.Body.Data = oapi.SystemComponentConstraintFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentConstraint(ctx context.Context, request *oapi.GetSystemComponentConstraintRequest) (*oapi.GetSystemComponentConstraintResponse, error) {
	var resp oapi.GetSystemComponentConstraintResponse

	constraint, queryErr := s.db.SystemComponentConstraint.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("failed to query system component constraint", queryErr)
	}
	resp.Body.Data = oapi.SystemComponentConstraintFromEnt(constraint)

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentConstraint(ctx context.Context, request *oapi.UpdateSystemComponentConstraintRequest) (*oapi.UpdateSystemComponentConstraintResponse, error) {
	var resp oapi.UpdateSystemComponentConstraintResponse

	attr := request.Body.Attributes
	update := s.db.SystemComponentConstraint.UpdateOneID(request.Id).
		SetNillableLabel(attr.Label).
		SetNillableDescription(attr.Description)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update system component constraint", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentConstraintFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentConstraint(ctx context.Context, request *oapi.ArchiveSystemComponentConstraintRequest) (*oapi.ArchiveSystemComponentConstraintResponse, error) {
	var resp oapi.ArchiveSystemComponentConstraintResponse

	if delErr := s.db.SystemComponentConstraint.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to archive system component constraint", delErr)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentControl(ctx context.Context, request *oapi.CreateSystemComponentControlRequest) (*oapi.CreateSystemComponentControlResponse, error) {
	var resp oapi.CreateSystemComponentControlResponse

	attr := request.Body.Attributes
	create := s.db.SystemComponentControl.Create().
		SetLabel(attr.Label).
		SetDescription(attr.Description)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create system component control", createErr)
	}
	resp.Body.Data = oapi.SystemComponentControlFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentControl(ctx context.Context, request *oapi.GetSystemComponentControlRequest) (*oapi.GetSystemComponentControlResponse, error) {
	var resp oapi.GetSystemComponentControlResponse

	control, queryErr := s.db.SystemComponentControl.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("failed to query system component control", queryErr)
	}
	resp.Body.Data = oapi.SystemComponentControlFromEnt(control)

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentControl(ctx context.Context, request *oapi.UpdateSystemComponentControlRequest) (*oapi.UpdateSystemComponentControlResponse, error) {
	var resp oapi.UpdateSystemComponentControlResponse

	attr := request.Body.Attributes
	update := s.db.SystemComponentControl.UpdateOneID(request.Id).
		SetNillableLabel(attr.Label).
		SetNillableDescription(attr.Description)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update system component control", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentControlFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentControl(ctx context.Context, request *oapi.ArchiveSystemComponentControlRequest) (*oapi.ArchiveSystemComponentControlResponse, error) {
	var resp oapi.ArchiveSystemComponentControlResponse

	if delErr := s.db.SystemComponentControl.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to archive system component control", delErr)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentSignal(ctx context.Context, request *oapi.CreateSystemComponentSignalRequest) (*oapi.CreateSystemComponentSignalResponse, error) {
	var resp oapi.CreateSystemComponentSignalResponse

	attr := request.Body.Attributes
	create := s.db.SystemComponentSignal.Create().
		SetLabel(attr.Label).
		SetDescription(attr.Description)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, detailError("failed to create system component signal", createErr)
	}
	resp.Body.Data = oapi.SystemComponentSignalFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentSignal(ctx context.Context, request *oapi.GetSystemComponentSignalRequest) (*oapi.GetSystemComponentSignalResponse, error) {
	var resp oapi.GetSystemComponentSignalResponse

	signal, queryErr := s.db.SystemComponentSignal.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, detailError("failed to query system component signal", queryErr)
	}
	resp.Body.Data = oapi.SystemComponentSignalFromEnt(signal)

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentSignal(ctx context.Context, request *oapi.UpdateSystemComponentSignalRequest) (*oapi.UpdateSystemComponentSignalResponse, error) {
	var resp oapi.UpdateSystemComponentSignalResponse

	attr := request.Body.Attributes
	update := s.db.SystemComponentSignal.UpdateOneID(request.Id).
		SetNillableLabel(attr.Label).
		SetNillableDescription(attr.Description)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update system component signal", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentSignalFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentSignal(ctx context.Context, request *oapi.ArchiveSystemComponentSignalRequest) (*oapi.ArchiveSystemComponentSignalResponse, error) {
	var resp oapi.ArchiveSystemComponentSignalResponse

	if delErr := s.db.SystemComponentSignal.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to archive system component signal", delErr)
	}

	return &resp, nil
}
