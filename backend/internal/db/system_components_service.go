package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemanalysis"
	scr "github.com/rezible/rezible/ent/systemcomponentrelationship"
)

type SystemComponentsService struct {
	db *ent.Client
}

func NewSystemComponentsService(db *ent.Client) (*SystemComponentsService, error) {
	s := &SystemComponentsService{
		db: db,
	}

	return s, nil
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

func (s *SystemComponentsService) ListSystemComponents(ctx context.Context, params rez.ListSystemComponentsParams) (*ent.ListResult[*ent.SystemComponent], error) {
	query := s.db.SystemComponent.Query()
	return ent.DoListQuery[*ent.SystemComponent, *ent.SystemComponentQuery](ctx, query, params.ListParams)
}

func (s *SystemComponentsService) GetRelationship(ctx context.Context, id1 uuid.UUID, id2 uuid.UUID) (*ent.SystemComponentRelationship, error) {
	pred1 := scr.And(scr.SourceID(id1), scr.TargetID(id2))
	pred2 := scr.And(scr.SourceID(id2), scr.TargetID(id1))

	query := s.db.SystemComponentRelationship.Query().
		Where(scr.Or(pred1, pred2))

	return query.Only(ctx)
}

func (s *SystemComponentsService) CreateRelationship(ctx context.Context, rel ent.SystemComponentRelationship) (*ent.SystemComponentRelationship, error) {
	create := s.db.SystemComponentRelationship.Create().
		SetSourceID(rel.SourceID).
		SetTargetID(rel.TargetID).
		SetDescription(rel.Description)
	return create.Save(ctx)
}

func (s *SystemComponentsService) GetSystemAnalysis(ctx context.Context, id uuid.UUID) (*ent.SystemAnalysis, error) {
	query := s.db.SystemAnalysis.Query().
		Where(systemanalysis.ID(id))

	// TODO: optimize this

	query.WithAnalysisComponents(func(q *ent.SystemAnalysisComponentQuery) {
		q.WithComponent(func(cq *ent.SystemComponentQuery) {
			cq.WithControls()
			cq.WithConstraints()
			cq.WithSignals()
		})
	})

	query.WithRelationships(func(q *ent.SystemAnalysisRelationshipQuery) {
		q.WithComponentRelationship()
		q.WithControlActions()
		q.WithFeedbackSignals()
	})

	return query.Only(ctx)
}

func (s *SystemComponentsService) CreateSystemAnalysisRelationship(ctx context.Context, params rez.CreateSystemAnalysisRelationshipParams) (*ent.SystemAnalysisRelationship, error) {
	var created *ent.SystemAnalysisRelationship

	cmpRel, relErr := s.GetRelationship(ctx, params.SourceId, params.TargetId)
	if relErr != nil && !ent.IsNotFound(relErr) {
		return nil, fmt.Errorf("failed to look up existing relationship: %w", relErr)
	}

	signals := params.FeedbackSignals
	mapCreateSignals := func(c *ent.SystemRelationshipFeedbackSignalCreate, i int) {
		sig := signals[i]
		c.SetDescription(sig.Description).SetSignalID(sig.Id)
	}

	controls := params.ControlActions
	mapCreateControls := func(c *ent.SystemRelationshipControlActionCreate, i int) {
		ctrl := controls[i]
		c.SetDescription(ctrl.Description).SetControlID(ctrl.Id)
	}

	createRelationshipTx := func(tx *ent.Tx) error {
		var createErr error

		if cmpRel == nil {
			createRel := tx.SystemComponentRelationship.Create().
				SetSourceID(params.SourceId).
				SetTargetID(params.TargetId)
			cmpRel, createErr = createRel.Save(ctx)
			if createErr != nil {
				return fmt.Errorf("creating base relationship: %w", createErr)
			}
		}

		createAnRel := tx.SystemAnalysisRelationship.Create().
			SetAnalysisID(params.AnalysisId).
			SetDescription(params.Description).
			SetComponentRelationshipID(cmpRel.ID)
		created, createErr = createAnRel.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("creating analysis relationship: %w", createErr)
		}

		created.Edges.ComponentRelationship = cmpRel

		createSignals := tx.SystemRelationshipFeedbackSignal.MapCreateBulk(signals, mapCreateSignals)
		created.Edges.FeedbackSignals, createErr = createSignals.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("creating feedback signals: %w", createErr)
		}

		createControls := tx.SystemRelationshipControlAction.MapCreateBulk(controls, mapCreateControls)
		created.Edges.ControlActions, createErr = createControls.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("creating control actions: %w", createErr)
		}

		return nil
	}

	if createErr := ent.WithTx(ctx, s.db, createRelationshipTx); createErr != nil {
		return nil, fmt.Errorf("failed to create system analysis relationship: %w", createErr)
	}

	return created, nil
}
