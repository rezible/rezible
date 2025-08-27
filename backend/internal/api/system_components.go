package api

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemcomponentrelationship"
	oapi "github.com/rezible/rezible/openapi"
)

type systemComponentsHandler struct {
	db         *ent.Client
	components rez.SystemComponentsService
}

func newSystemComponentsHandler(db *ent.Client, components rez.SystemComponentsService) *systemComponentsHandler {
	return &systemComponentsHandler{db: db, components: components}
}

func (s *systemComponentsHandler) ListSystemComponents(ctx context.Context, request *oapi.ListSystemComponentsRequest) (*oapi.ListSystemComponentsResponse, error) {
	var resp oapi.ListSystemComponentsResponse

	params := rez.ListSystemComponentsParams{
		ListParams: request.ListParams(),
	}

	listRes, queryErr := s.components.ListSystemComponents(ctx, params)
	if queryErr != nil {
		return nil, apiError("failed to query system components", queryErr)
	}
	resp.Body.Data = make([]oapi.SystemComponent, len(listRes.Data))
	for i, cmp := range listRes.Data {
		resp.Body.Data[i] = oapi.SystemComponentFromEnt(cmp)
	}
	resp.Body.Pagination = oapi.ResponsePagination{Total: listRes.Count}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponent(ctx context.Context, request *oapi.CreateSystemComponentRequest) (*oapi.CreateSystemComponentResponse, error) {
	var resp oapi.CreateSystemComponentResponse

	attr := request.Body.Attributes

	cmp := ent.SystemComponent{
		Name:        attr.Name,
		KindID:      attr.KindId,
		Description: attr.Description,
		Properties:  attr.Properties,
		Edges:       ent.SystemComponentEdges{},
	}

	cmp.Edges.Constraints = make([]*ent.SystemComponentConstraint, len(attr.Constraints))
	for i, cstr := range attr.Constraints {
		cmp.Edges.Constraints[i] = &ent.SystemComponentConstraint{Label: cstr.Label, Description: cstr.Description}
	}

	cmp.Edges.Controls = make([]*ent.SystemComponentControl, len(attr.Controls))
	for i, ctrl := range attr.Controls {
		cmp.Edges.Controls[i] = &ent.SystemComponentControl{Label: ctrl.Label, Description: ctrl.Description}
	}

	cmp.Edges.Signals = make([]*ent.SystemComponentSignal, len(attr.Signals))
	for i, sig := range attr.Signals {
		cmp.Edges.Signals[i] = &ent.SystemComponentSignal{Label: sig.Label, Description: sig.Description}
	}

	created, createErr := s.components.Create(ctx, cmp)
	if createErr != nil {
		return nil, apiError("failed to create system component", createErr)
	}
	resp.Body.Data = oapi.SystemComponentFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponent(ctx context.Context, request *oapi.GetSystemComponentRequest) (*oapi.GetSystemComponentResponse, error) {
	var resp oapi.GetSystemComponentResponse

	cmp, queryErr := s.db.SystemComponent.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("failed to query system component", queryErr)
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
		return nil, apiError("failed to update system component", updateErr)
	}

	resp.Body.Data = oapi.SystemComponentFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponent(ctx context.Context, request *oapi.ArchiveSystemComponentRequest) (*oapi.ArchiveSystemComponentResponse, error) {
	var resp oapi.ArchiveSystemComponentResponse

	if delErr := s.db.SystemComponent.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to delete system component", delErr)
	}

	return &resp, nil
}

// Relationships

func (s *systemComponentsHandler) ListSystemComponentRelationships(ctx context.Context, request *oapi.ListSystemComponentRelationshipsRequest) (*oapi.ListSystemComponentRelationshipsResponse, error) {
	var resp oapi.ListSystemComponentRelationshipsResponse

	query := s.db.SystemComponentRelationship.Query().
		Limit(request.Limit).
		Offset(request.Offset)

	srcPred := systemcomponentrelationship.SourceID(request.SourceId)
	if request.SourceId != uuid.Nil {
		query.Where(srcPred)
	}
	targetPred := systemcomponentrelationship.TargetID(request.TargetId)
	if request.TargetId != uuid.Nil {
		query.Where(targetPred)
	}
	if request.ComponentId != uuid.Nil {
		query.Where(systemcomponentrelationship.Or(srcPred, targetPred))
	}

	rels, relsErr := query.All(ctx)
	if relsErr != nil {
		return nil, apiError("failed to query system component relationships", relsErr)
	}
	resp.Body.Data = make([]oapi.SystemComponentRelationship, len(rels))
	for i, r := range rels {
		resp.Body.Data[i] = oapi.SystemComponentRelationshipFromEnt(r)
	}

	return &resp, nil
}

func (s *systemComponentsHandler) CreateSystemComponentRelationship(ctx context.Context, request *oapi.CreateSystemComponentRelationshipRequest) (*oapi.CreateSystemComponentRelationshipResponse, error) {
	var resp oapi.CreateSystemComponentRelationshipResponse

	attr := request.Body.Attributes
	create := s.db.SystemComponentRelationship.Create().
		SetSourceID(attr.SourceComponentId).
		SetTargetID(attr.TargetComponentId).
		SetDescription(attr.Description)

	created, createErr := create.Save(ctx)
	if createErr != nil {
		return nil, apiError("failed to create system component relationship", createErr)
	}
	resp.Body.Data = oapi.SystemComponentRelationshipFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentRelationship(ctx context.Context, request *oapi.GetSystemComponentRelationshipRequest) (*oapi.GetSystemComponentRelationshipResponse, error) {
	var resp oapi.GetSystemComponentRelationshipResponse

	rel, relErr := s.db.SystemComponentRelationship.Get(ctx, request.Id)
	if relErr != nil {
		return nil, apiError("failed to query system component relationship", relErr)
	}
	resp.Body.Data = oapi.SystemComponentRelationshipFromEnt(rel)

	return &resp, nil
}

func (s *systemComponentsHandler) UpdateSystemComponentRelationship(ctx context.Context, request *oapi.UpdateSystemComponentRelationshipRequest) (*oapi.UpdateSystemComponentRelationshipResponse, error) {
	var resp oapi.UpdateSystemComponentRelationshipResponse

	attr := request.Body.Attributes
	update := s.db.SystemComponentRelationship.UpdateOneID(request.Id)
	update.SetNillableDescription(attr.Description)

	updated, updateErr := update.Save(ctx)
	if updateErr != nil {
		return nil, apiError("failed to update system component relationship", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentRelationshipFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentRelationship(ctx context.Context, request *oapi.ArchiveSystemComponentRelationshipRequest) (*oapi.ArchiveSystemComponentRelationshipResponse, error) {
	var resp oapi.ArchiveSystemComponentRelationshipResponse

	if delErr := s.db.SystemComponent.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to delete system component", delErr)
	}

	return &resp, nil
}

// Kinds

func (s *systemComponentsHandler) ListSystemComponentKinds(ctx context.Context, request *oapi.ListSystemComponentKindsRequest) (*oapi.ListSystemComponentKindsResponse, error) {
	var resp oapi.ListSystemComponentKindsResponse

	query := s.db.SystemComponentKind.Query()
	// TODO ListParams

	kinds, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, apiError("failed to query system component kinds", queryErr)
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
		return nil, apiError("failed to create system component kind", createErr)
	}
	resp.Body.Data = oapi.SystemComponentKindFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentKind(ctx context.Context, request *oapi.GetSystemComponentKindRequest) (*oapi.GetSystemComponentKindResponse, error) {
	var resp oapi.GetSystemComponentKindResponse

	kind, queryErr := s.db.SystemComponentKind.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("failed to query system component kind", queryErr)
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
		return nil, apiError("failed to update system component kind", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentKindFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentKind(ctx context.Context, request *oapi.ArchiveSystemComponentKindRequest) (*oapi.ArchiveSystemComponentKindResponse, error) {
	var resp oapi.ArchiveSystemComponentKindResponse

	if delErr := s.db.SystemComponentKind.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to archive system component kind", delErr)
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
		return nil, apiError("failed to create system component constraint", createErr)
	}
	resp.Body.Data = oapi.SystemComponentConstraintFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentConstraint(ctx context.Context, request *oapi.GetSystemComponentConstraintRequest) (*oapi.GetSystemComponentConstraintResponse, error) {
	var resp oapi.GetSystemComponentConstraintResponse

	constraint, queryErr := s.db.SystemComponentConstraint.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("failed to query system component constraint", queryErr)
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
		return nil, apiError("failed to update system component constraint", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentConstraintFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentConstraint(ctx context.Context, request *oapi.ArchiveSystemComponentConstraintRequest) (*oapi.ArchiveSystemComponentConstraintResponse, error) {
	var resp oapi.ArchiveSystemComponentConstraintResponse

	if delErr := s.db.SystemComponentConstraint.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to archive system component constraint", delErr)
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
		return nil, apiError("failed to create system component control", createErr)
	}
	resp.Body.Data = oapi.SystemComponentControlFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentControl(ctx context.Context, request *oapi.GetSystemComponentControlRequest) (*oapi.GetSystemComponentControlResponse, error) {
	var resp oapi.GetSystemComponentControlResponse

	control, queryErr := s.db.SystemComponentControl.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("failed to query system component control", queryErr)
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
		return nil, apiError("failed to update system component control", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentControlFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentControl(ctx context.Context, request *oapi.ArchiveSystemComponentControlRequest) (*oapi.ArchiveSystemComponentControlResponse, error) {
	var resp oapi.ArchiveSystemComponentControlResponse

	if delErr := s.db.SystemComponentControl.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to archive system component control", delErr)
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
		return nil, apiError("failed to create system component signal", createErr)
	}
	resp.Body.Data = oapi.SystemComponentSignalFromEnt(created)

	return &resp, nil
}

func (s *systemComponentsHandler) GetSystemComponentSignal(ctx context.Context, request *oapi.GetSystemComponentSignalRequest) (*oapi.GetSystemComponentSignalResponse, error) {
	var resp oapi.GetSystemComponentSignalResponse

	signal, queryErr := s.db.SystemComponentSignal.Get(ctx, request.Id)
	if queryErr != nil {
		return nil, apiError("failed to query system component signal", queryErr)
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
		return nil, apiError("failed to update system component signal", updateErr)
	}
	resp.Body.Data = oapi.SystemComponentSignalFromEnt(updated)

	return &resp, nil
}

func (s *systemComponentsHandler) ArchiveSystemComponentSignal(ctx context.Context, request *oapi.ArchiveSystemComponentSignalRequest) (*oapi.ArchiveSystemComponentSignalResponse, error) {
	var resp oapi.ArchiveSystemComponentSignalResponse

	if delErr := s.db.SystemComponentSignal.DeleteOneID(request.Id).Exec(ctx); delErr != nil {
		return nil, apiError("failed to archive system component signal", delErr)
	}

	return &resp, nil
}
