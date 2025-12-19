package db

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	pc "github.com/rezible/rezible/ent/providerconfig"
)

type ProviderConfigService struct {
	db *ent.Client
}

func NewProviderConfigService(db *ent.Client) (*ProviderConfigService, error) {
	s := &ProviderConfigService{
		db: db,
	}

	return s, nil
}

func (s *ProviderConfigService) listQuery(p rez.ListProviderConfigsParams) *ent.ProviderConfigQuery {
	query := s.db.ProviderConfig.Query()
	if p.Enabled {
		query.Where(pc.EnabledEQ(true))
	}
	if p.ProviderType != "" {
		query.Where(pc.ProviderTypeEQ(p.ProviderType))
	}
	if p.ProviderId != "" {
		query.Where(pc.ProviderIDEQ(p.ProviderId))
	}
	return query
}

func (s *ProviderConfigService) ListProviderConfigs(ctx context.Context, params rez.ListProviderConfigsParams) ([]*ent.ProviderConfig, error) {
	query := s.listQuery(params)
	return query.All(ctx)
}

func (s *ProviderConfigService) GetProviderConfig(ctx context.Context, id uuid.UUID) (*ent.ProviderConfig, error) {
	return s.db.ProviderConfig.Get(ctx, id)
}

func (s *ProviderConfigService) LookupProviderConfig(ctx context.Context, pt pc.ProviderType, id string) (*ent.ProviderConfig, error) {
	return s.db.ProviderConfig.Query().
		Where(pc.And(pc.ProviderTypeEQ(pt), pc.ProviderID(id))).
		Only(ctx)
}

type updateConfigMutation interface {
	Save(context.Context) (*ent.ProviderConfig, error)
}

func (s *ProviderConfigService) UpdateProviderConfig(ctx context.Context, pc ent.ProviderConfig) (*ent.ProviderConfig, error) {
	var m updateConfigMutation
	if pc.ID != uuid.Nil {
		m = s.db.ProviderConfig.UpdateOneID(pc.ID).
			SetConfig(pc.Config).
			SetEnabled(pc.Enabled)
	} else {
		m = s.db.ProviderConfig.Create().
			SetProviderType(pc.ProviderType).
			SetProviderID(pc.ProviderID).
			SetConfig(pc.Config).
			SetEnabled(pc.Enabled)
	}
	return m.Save(ctx)
}

func (s *ProviderConfigService) DeleteProviderConfig(ctx context.Context, id uuid.UUID) error {
	return s.db.ProviderConfig.DeleteOneID(id).Exec(ctx)
}
