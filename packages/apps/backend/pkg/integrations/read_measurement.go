package integrations

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

type readMetrics struct {
	duration   metric.Float64Histogram
	resultSize metric.Int64Histogram
}

var loadReadMetrics = sync.OnceValues(func() (*readMetrics, error) {
	meter := otel.Meter("github.com/rezible/rezible")
	duration, durationErr := meter.Float64Histogram("rezible.backend.provider.read.duration",
		metric.WithDescription("Provider read duration"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10))
	resultSize, resultSizeErr := meter.Int64Histogram("rezible.backend.provider.read.result_size",
		metric.WithDescription("Items returned by successful provider reads"),
		metric.WithUnit("{item}"))
	if instrumentsErr := errors.Join(durationErr, resultSizeErr); instrumentsErr != nil {
		return nil, instrumentsErr
	}
	return &readMetrics{duration: duration, resultSize: resultSize}, nil
})

// MeasureRead runs one provider read as the unit of work "provider.read" and records its measurements.
// read returns how many items it returned.
func MeasureRead(ctx context.Context, provider, operation string, read func(context.Context) (int, error)) error {
	metrics, metricsErr := loadReadMetrics()
	if metricsErr != nil {
		return fmt.Errorf("provider read metrics: %w", metricsErr)
	}

	providerAttr := attribute.String("provider", provider)
	operationAttr := attribute.String("operation", operation)
	var items int
	started := time.Now()
	readErr := execution.Do(ctx, "provider.read", func(ctx context.Context) error {
		var itemsErr error
		items, itemsErr = read(ctx)
		return itemsErr
	}, providerAttr, operationAttr)
	duration := time.Since(started)

	// The provider may have been called whatever the outcome.
	rezai.AddRead(ctx)
	outcome := "succeeded"
	if readErr != nil {
		outcome = "failed"
	}
	metrics.duration.Record(ctx, duration.Seconds(), metric.WithAttributes(providerAttr, operationAttr, attribute.String("outcome", outcome)))
	if readErr == nil {
		metrics.resultSize.Record(ctx, int64(items), metric.WithAttributes(providerAttr, operationAttr))
	}
	return readErr
}
