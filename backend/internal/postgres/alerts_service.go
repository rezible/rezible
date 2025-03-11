package postgres

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type AlertsService struct {
	db    *ent.Client
	users rez.UserService
}

func NewAlertsService(ctx context.Context, db *ent.Client, users rez.UserService) (*AlertsService, error) {
	s := &AlertsService{
		db:    db,
		users: users,
	}

	return s, nil
}

//func (s *AlertsService) LoadDataProvider(ctx context.Context) error {
//	provider, provErr := s.loader.LoadAlertsDataProvider(ctx)
//	if provErr != nil {
//		return fmt.Errorf("failed to load alerts data provider: %w", provErr)
//	}
//	s.provider = provider
//	provider.SetOnAlertInstanceUpdatedCallback(func(providerId string, updatedAt time.Time) {})
//	return nil
//}

func (s *AlertsService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) ([]*ent.OncallAlert, error) {
	alerts := make([]*ent.OncallAlert, 0)

	return alerts, nil
}
