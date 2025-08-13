package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	afb "github.com/rezible/rezible/ent/alertfeedback"
	oe "github.com/rezible/rezible/ent/oncallevent"
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
	eventsQuery := s.db.OncallEvent.Query().
		Where(oe.AlertID(params.AlertId)).
		Where(oe.Or(oe.TimestampGTE(params.From), oe.TimestampGTE(params.To))).
		WithAnnotations(func(q *ent.OncallAnnotationQuery) {
			q.WithAlertFeedback()
		})

	//if params.RosterId != uuid.Nil {
	//	eventsQuery = eventsQuery.Where(oe.RosterID(params.RosterId))
	//}

	events, eventsQueryErr := eventsQuery.All(ctx)
	if eventsQueryErr != nil {
		return nil, fmt.Errorf("events: %w", eventsQueryErr)
	}

	metrics := &ent.AlertMetrics{
		EventCount: len(events),
	}

	for _, ev := range events {
		if ev.RosterID != uuid.Nil {
			hour := ev.Timestamp.Hour()
			if hour > 18 || hour < 9 {
				metrics.NightInterruptCount++
			}
		}
		for _, anno := range ev.Edges.Annotations {
			fb := anno.Edges.AlertFeedback
			if fb == nil {
				continue
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
	}

	return metrics, nil
}
