package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	afb "github.com/rezible/rezible/ent/alertfeedback"
	ne "github.com/rezible/rezible/ent/normalizedevent"
)

type AlertService struct {
	client    *ent.Client
	knowledge rez.KnowledgeService
}

func NewAlertService(client *ent.Client, knowledge rez.KnowledgeService) (*AlertService, error) {
	s := &AlertService{
		client:    client,
		knowledge: knowledge,
	}

	return s, nil
}

func (s *AlertService) ListAlerts(ctx context.Context, params rez.ListAlertsParams) ([]*ent.Alert, int, error) {
	query := s.client.Alert.Query().
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
	return s.client.Alert.Query().Where(alert.ID(id)).WithRoster().Only(ctx)
}

func (s *AlertService) GetAlertMetrics(ctx context.Context, params rez.GetAlertMetricsParams) (*ent.AlertMetrics, error) {
	query := s.client.Alert.Query().Where(alert.ID(params.AlertId)).
		WithFeedback(func(afbq *ent.AlertFeedbackQuery) {
			afbq.WithAlertInstance(func(neq *ent.NormalizedEventQuery) {
				neq.Where(ne.Or(ne.OccurredAtGTE(params.From), ne.OccurredAtLTE(params.To)))
			})
		})
	a, alertErr := query.Only(ctx)
	if alertErr != nil {
		return nil, fmt.Errorf("get alert: %w", alertErr)
	}

	metrics := &ent.AlertMetrics{}

	for _, fb := range a.Edges.Feedback {
		if ev := fb.Edges.AlertInstance; ev != nil {
			// metrics.EventCount++
			//if _, insErr := projections.DecodeAlertObserved(ev); insErr == nil {
			//if !ins.AcknowledgedAt.IsZero() {
			//	hour := ins.AcknowledgedAt.Hour()
			//	if hour > 18 || hour < 9 {
			//		metrics.NightInterruptCount++
			//	}
			//}
			//}
		}
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

	return metrics, nil
}

func (s *AlertService) GetActiveAlertsForComponents(ctx context.Context, componentIDs []uuid.UUID) ([]*ent.Alert, error) {
	// TODO: Implement actual alert correlation logic for components
	return []*ent.Alert{}, nil
}
