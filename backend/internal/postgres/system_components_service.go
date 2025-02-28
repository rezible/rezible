package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type SystemComponentsService struct {
	db     *ent.Client
	loader rez.ProviderLoader
}

func NewSystemComponentsService(db *ent.Client, pl rez.ProviderLoader) (*SystemComponentsService, error) {
	s := &SystemComponentsService{
		db:     db,
		loader: pl,
	}

	return s, nil
}

func (s *SystemComponentsService) SyncData(ctx context.Context) error {
	prov, provErr := s.loader.LoadSystemComponentsDataProvider(ctx)
	if provErr != nil {
		return fmt.Errorf("loading system components data provider: %w", provErr)
	}
	syncer := newSystemComponentsDataSyncer(s.db, prov)
	return syncer.syncProviderData(ctx)
}

func (s *SystemComponentsService) Create(ctx context.Context, cmp ent.SystemComponent) (*ent.SystemComponent, error) {
	var created *ent.SystemComponent

	createComponentTx := func(tx *ent.Tx) (*ent.SystemComponent, error) {
		create := tx.SystemComponent.Create().
			SetName(cmp.Name).
			SetDescription(cmp.Description).
			SetKindID(cmp.KindID)
		return create.Save(ctx)
	}

	createConstraintsTx := func(tx *ent.Tx, cmpId uuid.UUID) (ent.SystemComponentConstraints, error) {
		cstrs := cmp.Edges.Constraints
		mapFn := func(c *ent.SystemComponentConstraintCreate, idx int) {
			cstr := cstrs[idx]
			c.SetComponentID(cmpId).SetLabel(cstr.Label).SetDescription(cstr.Description)
		}
		return tx.SystemComponentConstraint.MapCreateBulk(cstrs, mapFn).Save(ctx)
	}

	createSignalsTx := func(tx *ent.Tx, cmpId uuid.UUID) (ent.SystemComponentSignals, error) {
		sigs := cmp.Edges.Signals
		create := tx.SystemComponentSignal.MapCreateBulk(sigs, func(c *ent.SystemComponentSignalCreate, idx int) {
			sig := sigs[idx]
			c.SetComponentID(cmpId).SetLabel(sig.Label).SetDescription(sig.Description)
		})
		return create.Save(ctx)
	}

	createControlsTx := func(tx *ent.Tx, cmpId uuid.UUID) (ent.SystemComponentControls, error) {
		ctrls := cmp.Edges.Controls
		create := tx.SystemComponentControl.MapCreateBulk(ctrls, func(c *ent.SystemComponentControlCreate, idx int) {
			ctrl := ctrls[idx]
			c.SetComponentID(cmpId).SetLabel(ctrl.Label).SetDescription(ctrl.Description)
		})
		return create.Save(ctx)
	}

	createTx := func(tx *ent.Tx) error {
		var createErr error

		created, createErr = createComponentTx(tx)
		if createErr != nil {
			return fmt.Errorf("creating component: %w", createErr)
		}

		created.Edges.Constraints, createErr = createConstraintsTx(tx, created.ID)
		if createErr != nil {
			return fmt.Errorf("creating constraints: %w", createErr)
		}

		created.Edges.Controls, createErr = createControlsTx(tx, created.ID)
		if createErr != nil {
			return fmt.Errorf("creating controls: %w", createErr)
		}

		created.Edges.Signals, createErr = createSignalsTx(tx, created.ID)
		if createErr != nil {
			return fmt.Errorf("creating signals: %w", createErr)
		}

		return nil
	}

	if createErr := ent.WithTx(ctx, s.db, createTx); createErr != nil {
		return nil, fmt.Errorf("create system component tx: %w", createErr)
	}
	return created, nil
}

func (s *SystemComponentsService) GetRelationship(ctx context.Context, id1 uuid.UUID, id2 uuid.UUID) (*ent.SystemComponentRelationship, error) {

	return nil, fmt.Errorf("not implemented")
}

func (s *SystemComponentsService) CreateRelationship(ctx context.Context, rel ent.SystemComponentRelationship) (*ent.SystemComponentRelationship, error) {

	return nil, fmt.Errorf("not implemented")
}

func (s *SystemComponentsService) CreateSystemAnalysisRelationship(ctx context.Context, params rez.CreateSystemAnalysisRelationshipParams) (*ent.SystemAnalysisRelationship, error) {

	/*
		var created *ent.SystemAnalysisRelationship

		createRelationshipTx := func(tx *ent.Tx) error {
			//create := tx.SystemAnalysisRelationship.Create().
			//	SetAnalysisID(request.Id).
			//	SetSourceComponentID(attr.SourceId).
			//	SetTargetComponentID(attr.TargetId).
			//	SetDescription(attr.Description)
			//rel, createErr := create.Save(ctx)
			//if createErr != nil {
			//	return createErr
			//}
			//
			//// TODO: controls & signals
			//
			//created = rel

			return nil
		}

		if createErr := ent.WithTx(ctx, s.db, createRelationshipTx); createErr != nil {
			return nil, detailError("failed to create system analysis relationship", createErr)
		}
	*/
	return nil, fmt.Errorf("not implemented")
}
