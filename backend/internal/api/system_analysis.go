package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemanalysiscomponent"
	"github.com/rezible/rezible/ent/systemanalysisrelationship"
	oapi "github.com/rezible/rezible/openapi"
)

type systemAnalysisHandler struct {
	db *ent.Client
}

func newSystemAnalysisHandler(db *ent.Client) *systemAnalysisHandler {
	return &systemAnalysisHandler{db: db}
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

	sysAn, getErr := s.db.SystemAnalysis.Get(ctx, request.Id)
	if getErr != nil {
		return nil, detailError("failed to get system analysis", getErr)
	}
	resp.Body.Data = oapi.SystemAnalysisFromEnt(sysAn)

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisComponents(ctx context.Context, request *oapi.ListSystemAnalysisComponentsRequest) (*oapi.ListSystemAnalysisComponentsResponse, error) {
	var resp oapi.ListSystemAnalysisComponentsResponse

	query := s.db.SystemAnalysisComponent.Query().
		Where(systemanalysiscomponent.AnalysisID(request.Id)).
		WithComponent()
	cmps, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query system analysis components", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemAnalysisComponent, len(cmps))
	for i, cmp := range cmps {
		resp.Body.Data[i] = oapi.SystemAnalysisComponentFromEnt(cmp)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) AddSystemAnalysisComponent(ctx context.Context, request *oapi.AddSystemAnalysisComponentRequest) (*oapi.AddSystemAnalysisComponentResponse, error) {
	var resp oapi.AddSystemAnalysisComponentResponse

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisComponent(ctx context.Context, request *oapi.GetSystemAnalysisComponentRequest) (*oapi.GetSystemAnalysisComponentResponse, error) {
	var resp oapi.GetSystemAnalysisComponentResponse

	cmp, getErr := s.db.SystemAnalysisComponent.Get(ctx, request.Id)
	if getErr != nil {
		return nil, detailError("failed to get system analysis component", getErr)
	}
	resp.Body.Data = oapi.SystemAnalysisComponentFromEnt(cmp)

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisComponent(ctx context.Context, request *oapi.UpdateSystemAnalysisComponentRequest) (*oapi.UpdateSystemAnalysisComponentResponse, error) {
	var resp oapi.UpdateSystemAnalysisComponentResponse

	attr := request.Body.Attributes
	update := s.db.SystemAnalysisComponent.UpdateOneID(request.Id)

	if attr.Position != nil {
		//update.SetPosX(attr.Position.X)
		//update.SetPosY(attr.Position.Y)
	}

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, detailError("failed to update system analysis component", updateErr)
	}
	resp.Body.Data = oapi.SystemAnalysisComponentFromEnt(updated)

	return &resp, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisComponent(ctx context.Context, request *oapi.DeleteSystemAnalysisComponentRequest) (*oapi.DeleteSystemAnalysisComponentResponse, error) {
	var resp oapi.DeleteSystemAnalysisComponentResponse

	if delErr := s.db.SystemAnalysisComponent.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to delete system analysis component", delErr)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) ListSystemAnalysisRelationships(ctx context.Context, request *oapi.ListSystemAnalysisRelationshipsRequest) (*oapi.ListSystemAnalysisRelationshipsResponse, error) {
	var resp oapi.ListSystemAnalysisRelationshipsResponse

	query := s.db.SystemAnalysisRelationship.Query().
		Where(systemanalysisrelationship.AnalysisID(request.Id))

	rels, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, detailError("failed to query system analysis relationships", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemAnalysisRelationship, len(rels))
	for i, rel := range rels {
		resp.Body.Data[i] = oapi.SystemAnalysisRelationshipFromEnt(rel)
	}

	return &resp, nil
}

func (s *systemAnalysisHandler) CreateSystemAnalysisRelationship(ctx context.Context, request *oapi.CreateSystemAnalysisRelationshipRequest) (*oapi.CreateSystemAnalysisRelationshipResponse, error) {
	var resp oapi.CreateSystemAnalysisRelationshipResponse

	attr := request.Body.Attributes

	var created *ent.SystemAnalysisRelationship

	createRelationshipTx := func(tx *ent.Tx) error {
		create := tx.SystemAnalysisRelationship.Create().
			SetAnalysisID(request.Id).
			SetSourceComponentID(attr.SourceId).
			SetTargetComponentID(attr.TargetId).
			SetDescription(attr.Description)
		rel, createErr := create.Save(ctx)
		if createErr != nil {
			return createErr
		}

		// TODO: controls & signals

		created = rel

		return nil
	}

	if createErr := ent.WithTx(ctx, s.db, createRelationshipTx); createErr != nil {
		return nil, detailError("failed to create system analysis relationship", createErr)
	}
	resp.Body.Data = oapi.SystemAnalysisRelationshipFromEnt(created)

	return &resp, nil
}

func (s *systemAnalysisHandler) GetSystemAnalysisRelationship(ctx context.Context, request *oapi.GetSystemAnalysisRelationshipRequest) (*oapi.GetSystemAnalysisRelationshipResponse, error) {
	var resp oapi.GetSystemAnalysisRelationshipResponse

	rel, getErr := s.db.SystemAnalysisRelationship.Get(ctx, request.Id)
	if getErr != nil {
		return nil, detailError("failed to get system analysis relationship", getErr)
	}
	resp.Body.Data = oapi.SystemAnalysisRelationshipFromEnt(rel)

	return &resp, nil
}

func (s *systemAnalysisHandler) UpdateSystemAnalysisRelationship(ctx context.Context, request *oapi.UpdateSystemAnalysisRelationshipRequest) (*oapi.UpdateSystemAnalysisRelationshipResponse, error) {
	var resp oapi.UpdateSystemAnalysisRelationshipResponse

	attr := request.Body.Attributes

	var updated *ent.SystemAnalysisRelationship

	current, getErr := s.db.SystemAnalysisRelationship.Get(ctx, request.Id)
	if getErr != nil {
		return nil, detailError("failed to update system analysis relationship", getErr)
	}

	updateRelationshipTx := func(tx *ent.Tx) error {
		fmt.Printf("compare traits %v\n", current)

		update := s.db.SystemAnalysisRelationship.UpdateOneID(request.Id).
			SetNillableDescription(attr.Description)

		rel, saveErr := update.Save(ctx)
		if saveErr != nil {
			return saveErr
		}

		updated = rel
		return nil
	}

	if updateErr := ent.WithTx(ctx, s.db, updateRelationshipTx); updateErr != nil {
		return nil, detailError("failed to update system analysis relationship", updateErr)
	}
	resp.Body.Data = oapi.SystemAnalysisRelationshipFromEnt(updated)

	return &resp, nil
}

func (s *systemAnalysisHandler) DeleteSystemAnalysisRelationship(ctx context.Context, request *oapi.DeleteSystemAnalysisRelationshipRequest) (*oapi.DeleteSystemAnalysisRelationshipResponse, error) {
	var resp oapi.DeleteSystemAnalysisRelationshipResponse

	if delErr := s.db.SystemAnalysisRelationship.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, detailError("failed to delete system analysis relationship", delErr)
	}

	return &resp, nil
}
