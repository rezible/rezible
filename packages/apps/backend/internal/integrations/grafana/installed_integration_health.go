package grafana

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/rezible/rezible/pkg/integrations"
)

const healthCheckWindow = 15 * time.Minute

var _ integrations.HealthChecker = (*InstalledIntegration)(nil)

// CheckHealth lists the values of each configured service label over the last 15 minutes, in Loki and in
// Prometheus. Failures for both data sources are reported together.
func (ii *InstalledIntegration) CheckHealth(ctx context.Context) error {
	end := ii.clock.Now()
	start := end.Add(-healthCheckWindow)

	var failures []error
	logsLabel := ii.settings.LogServiceLabel
	logs, logsErr := ii.logsDataSource()
	if logsErr == nil {
		logsErr = ii.checkLabelHasValues(ctx, logs, "/loki/api/v1/label/"+logsLabel+"/values", logsLabel, start, end)
	}
	if logsErr != nil {
		failures = append(failures, fmt.Errorf("logs: %w", logsErr))
	}
	metricsLabel := ii.settings.MetricServiceLabel
	metrics, metricsErr := ii.metricsDataSource()
	if metricsErr == nil {
		metricsErr = ii.checkLabelHasValues(ctx, metrics, "/api/v1/label/"+metricsLabel+"/values", metricsLabel, start, end)
	}
	if metricsErr != nil {
		failures = append(failures, fmt.Errorf("metrics: %w", metricsErr))
	}
	return errors.Join(failures...)
}

func (ii *InstalledIntegration) checkLabelHasValues(ctx context.Context, ds dataSource, path, label string, start, end time.Time) error {
	values := url.Values{
		"start": {formatTime(start)},
		"end":   {formatTime(end)},
	}
	var labelValues []string
	if getErr := ii.client.get(ctx, ds, path, values, &labelValues); getErr != nil {
		return getErr
	}
	if len(labelValues) == 0 {
		return fmt.Errorf("no values for label `%s` in the last 15 minutes", label)
	}
	return nil
}
