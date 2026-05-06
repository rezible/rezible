package datasync

import (
	"context"

	"github.com/rezible/rezible/telemetry"
)

type metrics struct {
	runs         telemetry.Int64Counter
	setupSeconds telemetry.Float64Histogram
	syncSeconds  telemetry.Float64Histogram
	records      telemetry.Int64Counter
	mutations    telemetry.Int64Counter
}

func newMetrics() *metrics {
	meter := telemetry.DefaultMeter()
	return &metrics{
		runs:         telemetry.Int64CounterInstrument(meter, "rezible.backend.datasync.runs", "Integration data sync runs"),
		setupSeconds: telemetry.Float64HistogramInstrument(meter, "rezible.backend.datasync.setup_duration", "Integration data sync setup duration", "s"),
		syncSeconds:  telemetry.Float64HistogramInstrument(meter, "rezible.backend.datasync.sync_duration", "Integration data sync sync duration", "s"),
		records:      telemetry.Int64CounterInstrument(meter, "rezible.backend.datasync.records", "Integration data sync records pulled"),
		mutations:    telemetry.Int64CounterInstrument(meter, "rezible.backend.datasync.mutations", "Integration data sync mutations applied"),
	}
}

func (m *metrics) recordRun(ctx context.Context, dataType, providerType string, res *syncResult, err error) {
	if m == nil {
		return
	}
	skipped := res != nil && res.skipped
	attrs := []telemetry.KeyValue{
		telemetry.StringAttr("data_type", telemetry.NormalizeLabel(dataType)),
		telemetry.StringAttr("provider_type", telemetry.NormalizeLabel(providerType)),
		telemetry.ResultAttr(err),
		telemetry.BoolAttr("skipped", skipped),
	}
	m.runs.Add(ctx, 1, telemetry.WithMetricAttributes(attrs...))
	if res != nil {
		m.setupSeconds.Record(ctx, res.setupTime.Seconds(), telemetry.WithMetricAttributes(attrs...))
		if !skipped {
			m.syncSeconds.Record(ctx, res.syncTime.Seconds(), telemetry.WithMetricAttributes(attrs...))
			m.records.Add(ctx, res.recordsCount, telemetry.WithMetricAttributes(attrs...))
			m.mutations.Add(ctx, res.mutationCount, telemetry.WithMetricAttributes(attrs...))
		}
	}
}
