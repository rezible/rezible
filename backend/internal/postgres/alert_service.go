package postgres

import (
	"context"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type AlertService struct {
	db   *ent.Client
	jobs rez.JobsService
	prov rez.AlertDataProvider
}

func NewAlertService(db *ent.Client, jobs rez.JobsService, prov rez.AlertDataProvider) (*AlertService, error) {
	s := &AlertService{
		db:   db,
		jobs: jobs,
		prov: prov,
	}

	return s, nil
}

// TODO: don't fake
var fakeAlert = &ent.Alert{
	ID:         uuid.New(),
	Title:      "Example Alert",
	ProviderID: "example",
	Edges:      ent.AlertEdges{},
}

func (s *AlertService) ListAlerts(ctx context.Context, params *rez.ListAlertsParams) ([]*ent.Alert, int, error) {
	return []*ent.Alert{fakeAlert}, 1, nil
}

func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*ent.Alert, error) {
	return fakeAlert, nil
}
