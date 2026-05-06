package db

import (
	"context"

	"github.com/rezible/rezible/telemetry"
)

type providerEventMetrics struct {
	ingested          telemetry.Int64Counter
	processed         telemetry.Int64Counter
	processSeconds    telemetry.Float64Histogram
	projectionSeconds telemetry.Float64Histogram
	normalizedEvents  telemetry.Int64Counter
}

func newProviderEventMetrics() *providerEventMetrics {
	meter := telemetry.DefaultMeter()
	return &providerEventMetrics{
		ingested:          telemetry.Int64CounterInstrument(meter, "rezible.backend.provider_events.ingested", "Provider events ingested"),
		processed:         telemetry.Int64CounterInstrument(meter, "rezible.backend.provider_events.processed", "Provider events processed"),
		processSeconds:    telemetry.Float64HistogramInstrument(meter, "rezible.backend.provider_events.normalize_duration", "Provider event normalization processing duration", "s"),
		normalizedEvents:  telemetry.Int64CounterInstrument(meter, "rezible.backend.provider_events.normalized_events", "Normalized provider events saved"),
		projectionSeconds: telemetry.Float64HistogramInstrument(meter, "rezible.backend.provider_events.projection_duration", "Normalized event projection duration", "s"),
	}
}

func (m *providerEventMetrics) recordIngested(ctx context.Context, provider, source string, res *ingestProviderEventResult, err error) {
	if m != nil {
		m.ingested.Add(ctx, 1, telemetry.WithAttributes(
			telemetry.StringAttr("provider", telemetry.NormalizeLabel(provider)),
			telemetry.StringAttr("source", telemetry.NormalizeLabel(source)),
			telemetry.ResultAttr(err),
			telemetry.BoolAttr("duplicate", res != nil && res.duplicate),
		))
	}
}

func (m *providerEventMetrics) recordProcessed(ctx context.Context, provider, source string, res *processProviderEventResult, err error) {
	if m != nil {
		processSuccess := res != nil && res.processSuccess
		projectionSuccess := res != nil && res.projectionSuccess
		attrs := []telemetry.KeyValue{
			telemetry.StringAttr("provider", telemetry.NormalizeLabel(provider)),
			telemetry.StringAttr("source", telemetry.NormalizeLabel(source)),
			telemetry.ResultAttr(err),
			telemetry.BoolAttr("process_success", processSuccess),
			telemetry.BoolAttr("projection_success", projectionSuccess),
		}
		m.processed.Add(ctx, 1, telemetry.WithAttributes(attrs...))
		if res != nil {
			m.processSeconds.Record(ctx, res.processTime.Seconds(), telemetry.WithAttributes(attrs...))
			if res.normalizeCount > 0 {
				m.normalizedEvents.Add(ctx, int64(res.normalizeCount), telemetry.WithAttributes(attrs...))
				m.projectionSeconds.Record(ctx, res.projectionTime.Seconds(), telemetry.WithAttributes(attrs...))
			}
		}
	}
}
