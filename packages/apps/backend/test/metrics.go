package test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

// The global meter provider is set once per process: instruments created from the global meter before it is set
// forward to the first provider only.
var globalMetricReader = sync.OnceValue(func() *sdkmetric.ManualReader {
	reader := sdkmetric.NewManualReader()
	otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	return reader
})

// MetricRecorder reads what the global meter provider recorded since the recorder started.
type MetricRecorder struct {
	t        *testing.T
	baseline metricdata.ResourceMetrics
}

// RecordMetrics installs a manual reader as the global meter provider, once per process, and starts recording.
// Measurements made before the first call in a process are lost.
func RecordMetrics(t *testing.T) *MetricRecorder {
	r := &MetricRecorder{t: t}
	r.baseline = r.collect()
	return r
}

func (r *MetricRecorder) collect() metricdata.ResourceMetrics {
	var data metricdata.ResourceMetrics
	require.NoError(r.t, globalMetricReader().Collect(r.t.Context(), &data))
	return data
}

// Count returns how many measurements were recorded since the recorder started: a counter's sum or a histogram's
// count, over the points whose attributes include attrs.
func (r *MetricRecorder) Count(name string, attrs ...attribute.KeyValue) int64 {
	count, _ := metricTotals(r.collect(), name, attrs)
	baseCount, _ := metricTotals(r.baseline, name, attrs)
	return count - baseCount
}

// Sum returns the sum of a histogram's values recorded since the recorder started, over the points whose
// attributes include attrs.
func (r *MetricRecorder) Sum(name string, attrs ...attribute.KeyValue) float64 {
	_, sum := metricTotals(r.collect(), name, attrs)
	_, baseSum := metricTotals(r.baseline, name, attrs)
	return sum - baseSum
}

func metricTotals(data metricdata.ResourceMetrics, name string, attrs []attribute.KeyValue) (count int64, sum float64) {
	matches := func(set attribute.Set) bool {
		for _, attr := range attrs {
			if value, found := set.Value(attr.Key); !found || value != attr.Value {
				return false
			}
		}
		return true
	}
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != name {
				continue
			}
			switch agg := m.Data.(type) {
			case metricdata.Sum[int64]:
				for _, point := range agg.DataPoints {
					if matches(point.Attributes) {
						count += point.Value
					}
				}
			case metricdata.Histogram[float64]:
				for _, point := range agg.DataPoints {
					if matches(point.Attributes) {
						count += int64(point.Count)
						sum += point.Sum
					}
				}
			case metricdata.Histogram[int64]:
				for _, point := range agg.DataPoints {
					if matches(point.Attributes) {
						count += int64(point.Count)
						sum += float64(point.Sum)
					}
				}
			}
		}
	}
	return count, sum
}
