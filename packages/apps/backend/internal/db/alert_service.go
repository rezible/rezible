package db

import (
	"context"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alertdefinition"
	"github.com/rezible/rezible/ent/alertinstance"
)

type AlertService struct {
	db rez.Database
}

func NewAlertService(db rez.Database) (*AlertService, error) {
	s := &AlertService{db: db}

	return s, nil
}

func (s *AlertService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) (*ent.ListResult[ent.AlertDefinition], error) {
	query := s.db.Client(ctx).AlertDefinition.Query().
		Order(alertdefinition.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(alertdefinition.TitleContainsFold(search))
	}
	return ent.DoListQuery[ent.AlertDefinition, *ent.AlertDefinitionQuery](ctx, query, params.ListParams)
}

func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*ent.AlertDefinition, error) {
	query := s.db.Client(ctx).AlertDefinition.Query().
		Where(alertdefinition.ID(id))
	return query.Only(ctx)
}

func (s *AlertService) GetAlertInstance(ctx context.Context, id uuid.UUID) (*ent.AlertInstance, error) {
	query := s.db.Client(ctx).AlertInstance.Query().
		Where(alertinstance.ID(id)).
		WithEpisode(func(query *ent.AlertEpisodeQuery) {
			query.WithAlertDefinition()
		})
	return query.Only(ctx)
}

func (s *AlertService) GetAlertMetrics(ctx context.Context, params rez.GetAlertMetricsParams) (*ent.AlertMetrics, error) {
	return &ent.AlertMetrics{}, nil
}
