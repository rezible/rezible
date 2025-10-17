package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	afb "github.com/rezible/rezible/ent/alertfeedback"
	ae "github.com/rezible/rezible/ent/alertinstance"
	"github.com/rezible/rezible/ent/event"
)

type AlertService struct {
	db *ent.Client
}

func NewAlertService(db *ent.Client) (*AlertService, error) {
	s := &AlertService{
		db: db,
	}

	return s, nil
}
func (s *AlertService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) ([]*ent.Alert, int, error) {
	query := s.db.Alert.Query().
		Where()

	qCtx := params.GetQueryContext(ctx)
	count, queryErr := query.Count(qCtx)
	if queryErr != nil {
		return nil, 0, fmt.Errorf("count: %w", queryErr)
	}
	alerts := make([]*ent.Alert, 0)
	if count > 0 {
		alerts, queryErr = query.All(qCtx)
	}
	if queryErr != nil {
		return nil, 0, fmt.Errorf("query: %w", queryErr)
	}
	return alerts, count, nil
}

func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*ent.Alert, error) {
	return s.db.Alert.Query().Where(alert.ID(id)).WithRoster().Only(ctx)
}

func (s *AlertService) GetAlertMetrics(ctx context.Context, params rez.GetAlertMetricsParams) (*ent.AlertMetrics, error) {
	query := s.db.AlertInstance.Query().
		Where(ae.AlertID(params.AlertId)).
		WithFeedback().
		WithEvent(func(q *ent.EventQuery) {
			q.Where(event.Or(event.TimestampGTE(params.From), event.TimestampGTE(params.To)))
		})

	instances, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("get alert instances: %w", queryErr)
	}

	metrics := &ent.AlertMetrics{
		EventCount: len(instances),
	}

	for _, ins := range instances {
		if !ins.AcknowledgedAt.IsZero() {
			hour := ins.AcknowledgedAt.Hour()
			if hour > 18 || hour < 9 {
				metrics.NightInterruptCount++
			}
		}
		if fb := ins.Edges.Feedback; fb != nil {
			metrics.FeedbackCount++
			if fb.Actionable {
				metrics.FeedbackActionable++
			}
			if fb.DocumentationAvailable {
				metrics.FeedbackDocsAvailable++
			}
			if fb.DocumentationAvailable {
				metrics.FeedbackDocsNeedUpdate++
			}
			if fb.Accurate == afb.AccurateYes {
				metrics.FeedbackAccurate++
			}
			if fb.Accurate == afb.AccurateUnknown {
				metrics.FeedbackAccurateUnknown++
			}
		}
	}

	return metrics, nil
}

func (s *AlertService) GetActiveAlertsForComponents(ctx context.Context, componentIDs []uuid.UUID) ([]*ent.Alert, error) {
	// TODO: Implement actual alert correlation logic for components
	return []*ent.Alert{}, nil
}
