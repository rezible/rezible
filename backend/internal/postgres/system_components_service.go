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
