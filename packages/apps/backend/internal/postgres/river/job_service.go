package river

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/rivercontrib/otelriver"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
)

const SchemaName = "river"

type riverClient = river.Client[pgx.Tx]

type JobService struct {
	logger *slog.Logger
	client *riverClient
}

func NewJobService(pool *pgxpool.Pool, tel rez.TelemetryService) (*JobService, error) {
	s := &JobService{
		logger: tel.NewLogger(rez.NewLoggerOptions{
			PackageName: "river",
			Level:       slog.LevelInfo,
		}),
	}

	telemetryMiddleware := otelriver.NewMiddleware(&otelriver.MiddlewareConfig{
		DurationUnit:                "s",
		EnableSemanticMetrics:       true,
		EnableWorkSpanJobKindSuffix: true,
		MeterProvider:               tel.MeterProvider(),
		TracerProvider:              tel.TracerProvider(),
	})

	cfg := &river.Config{
		Schema:      SchemaName,
		Logger:      s.logger,
		MaxAttempts: 3,
		Middleware: []rivertype.Middleware{
			telemetryMiddleware,
			&accessContextMiddleware{},
		},
		Workers: jobs.GetWorkers(),
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 20},
		},
	}
	var clientErr error
	s.client, clientErr = river.NewClient(riverpgxv5.New(pool), cfg)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to create client: %w", clientErr)
	}

	return s, nil
}

func (s *JobService) Start(ctx context.Context) error {
	_, pjErr := s.client.PeriodicJobs().AddManySafely(jobs.GetPeriodicJobs())
	if pjErr != nil {
		return fmt.Errorf("failed to add periodic jobs: %w", pjErr)
	}
	ctx = execution.NewRootContext(ctx, execution.KindSystem, execution.SourceJob)
	return s.client.Start(ctx)
}

func (s *JobService) Shutdown(ctx context.Context) error {
	if s.client != nil {
		return s.client.Stop(ctx)
	}
	return nil
}

func (s *JobService) extractContextPgxTx(ctx context.Context) (bool, pgx.Tx, error) {
	if tx := ent.TxFromContext(ctx); tx != nil {
		pgxTx, pgErr := ent.ExtractPgxTx(tx)
		return true, pgxTx, pgErr
	}
	return false, nil, nil
}

func (s *JobService) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if isTx, pgxTx, txErr := s.extractContextPgxTx(ctx); isTx {
		if txErr != nil {
			return nil, fmt.Errorf("extract pgx tx: %w", txErr)
		}
		return s.client.InsertTx(ctx, pgxTx, args, opts)
	}
	return s.client.Insert(ctx, args, opts)
}

func (s *JobService) InsertMany(ctx context.Context, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error) {
	if isTx, pgxTx, txErr := s.extractContextPgxTx(ctx); isTx {
		if txErr != nil {
			return nil, fmt.Errorf("extract pgx tx: %w", txErr)
		}
		return s.client.InsertManyTx(ctx, pgxTx, params)
	}
	return s.client.InsertMany(ctx, params)
}
