package postgres

import (
	"context"
	"fmt"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type AlertsService struct {
	db       *ent.Client
	jobs     rez.BackgroundJobService
	loader   rez.ProviderLoader
	provider rez.AlertsDataProvider
	users    rez.UserService
}

func NewAlertsService(ctx context.Context, db *ent.Client, jobs rez.BackgroundJobService, pl rez.ProviderLoader, users rez.UserService) (*AlertsService, error) {
	s := &AlertsService{
		db:     db,
		jobs:   jobs,
		loader: pl,
		users:  users,
	}

	if dataErr := s.LoadDataProvider(ctx); dataErr != nil {
		return nil, fmt.Errorf("failed to register data provider: %w", dataErr)
	}

	return s, nil
}

func (s *AlertsService) LoadDataProvider(ctx context.Context) error {
	provider, provErr := s.loader.LoadAlertsDataProvider(ctx)
	if provErr != nil {
		return fmt.Errorf("failed to load alerts data provider: %w", provErr)
	}
	s.provider = provider

	provider.SetOnAlertInstanceUpdatedCallback(func(providerId string, updatedAt time.Time) {

	})

	return nil
}

func (s *AlertsService) SyncData(ctx context.Context) error {
	return nil //s.dataSyncer.syncProviderData(ctx)
}

func (s *AlertsService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) ([]*ent.OncallAlert, error) {
	alerts := make([]*ent.OncallAlert, 0)

	return alerts, nil
}
