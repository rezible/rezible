package postgres

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
	"time"
)

type AlertsService struct {
	db        *ent.Client
	jobClient *jobs.BackgroundJobClient
	loader    rez.ProviderLoader
	provider  rez.AlertsDataProvider
	users     rez.UserService
}

func NewAlertsService(ctx context.Context, db *ent.Client, jobClient *jobs.BackgroundJobClient, pl rez.ProviderLoader, users rez.UserService) (*AlertsService, error) {
	s := &AlertsService{
		db:        db,
		jobClient: jobClient,
		loader:    pl,
		users:     users,
	}

	if dataErr := s.LoadDataProvider(ctx); dataErr != nil {
		return nil, fmt.Errorf("failed to register data provider: %w", dataErr)
	}

	if jobsErr := s.RegisterJobs(); jobsErr != nil {
		return nil, fmt.Errorf("failed to register background job: %w", jobsErr)
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
