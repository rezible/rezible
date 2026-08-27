package river_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/rezible/rezible/internal/postgres"
	postgrespgtestdb "github.com/rezible/rezible/internal/postgres/pgtestdb"
	rezriver "github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	rez "github.com/rezible/rezible"
)

type jobServiceTestArgs struct{}

func (jobServiceTestArgs) Kind() string { return "job-service-test" }

type jobServiceSuite struct {
	test.Suite
}

func TestJobServiceSuite(t *testing.T) {
	suite.Run(t, &jobServiceSuite{Suite: test.NewSuite()})
}

func (s *jobServiceSuite) newJobService() (*rezriver.JobService, *jobs.Registry) {
	testDatabase, databaseErr := postgrespgtestdb.New(s.Config().Postgres)
	s.Require().NoError(databaseErr)
	s.T().Cleanup(func() { s.Require().NoError(testDatabase.Shutdown()) })

	pool, poolErr := postgres.MakePgxPool(s.T().Context(), testDatabase.Config(), false)
	s.Require().NoError(poolErr)
	s.T().Cleanup(pool.Close)

	reg := jobs.NewRegistry()
	service, serviceErr := rezriver.NewJobService(s.Config(), pool, newTestTelemetry(), reg)
	s.Require().NoError(serviceErr)
	return service, reg
}

func (s *jobServiceSuite) TestRegistrationAfterConstructionIsUsedByInserts() {
	service, registry := s.newJobService()
	s.Require().NoError(registry.AddWorkerFunc(func(context.Context, jobServiceTestArgs) error { return nil }))
	s.Require().NoError(service.Finalize())

	result, insertErr := service.Insert(s.T().Context(), jobServiceTestArgs{}, nil)
	s.Require().NoError(insertErr)
	s.Require().NotNil(result.Job)
	s.Equal("job-service-test", result.Job.Kind)
}

type testTelemetry struct {
	logger         *slog.Logger
	meterProvider  metric.MeterProvider
	tracerProvider trace.TracerProvider
}

func newTestTelemetry() *testTelemetry {
	return &testTelemetry{
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		meterProvider:  metricnoop.NewMeterProvider(),
		tracerProvider: tracenoop.NewTracerProvider(),
	}
}

func (t *testTelemetry) NewLogger(rez.NewLoggerOptions) *slog.Logger { return t.logger }
func (t *testTelemetry) Logger() *slog.Logger                        { return t.logger }
func (t *testTelemetry) TracerProvider() trace.TracerProvider        { return t.tracerProvider }
func (t *testTelemetry) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return t.tracerProvider.Tracer(name, opts...)
}
func (t *testTelemetry) DefaultTracer() trace.Tracer {
	return t.tracerProvider.Tracer("test")
}
func (t *testTelemetry) MeterProvider() metric.MeterProvider { return t.meterProvider }
func (t *testTelemetry) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	return t.meterProvider.Meter(name, opts...)
}
func (t *testTelemetry) DefaultMeter() metric.Meter {
	return t.meterProvider.Meter("test")
}
