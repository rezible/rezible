package river

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/rivercontrib/otelriver"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
)

const SchemaName = "river"

type riverClient = river.Client[pgx.Tx]

type JobService struct {
	logger   *slog.Logger
	pool     *pgxpool.Pool
	registry *jobs.Registry
	config   *river.Config
	client   *riverClient
}

func NewJobService(cfg rez.Config, pool *pgxpool.Pool, tel rez.TelemetryService, reg *jobs.Registry) (*JobService, error) {
	s := &JobService{
		pool:     pool,
		registry: reg,
		logger: tel.NewLogger(rez.NewLoggerOptions{
			Name:  "river",
			Level: slog.LevelInfo,
		}),
	}

	telemetryMiddleware := otelriver.NewMiddleware(&otelriver.MiddlewareConfig{
		DurationUnit:                "s",
		EnableSemanticMetrics:       true,
		EnableWorkSpanJobKindSuffix: true,
		MeterProvider:               tel.MeterProvider(),
		TracerProvider:              tel.TracerProvider(),
	})

	s.config = &river.Config{
		Schema:          SchemaName,
		Logger:          s.logger,
		MaxAttempts:     3,
		SoftStopTimeout: 5 * time.Second,
		Middleware: []rivertype.Middleware{
			telemetryMiddleware,
			&accessContextMiddleware{},
		},
		Queues: map[string]river.QueueConfig{
			river.QueueDefault:   {MaxWorkers: 20},
			jobs.AgentTurnsQueue: {MaxWorkers: cfg.AI.Agents.MaxWorkers},
		},
	}
	return s, nil
}

func (s *JobService) Finalize() error {
	defs, sealErr := s.registry.Seal()
	if sealErr != nil {
		return fmt.Errorf("seal registry: %w", sealErr)
	}
	if s.client != nil {
		return fmt.Errorf("job service is already finalized")
	}

	s.config.Workers = defs.Workers
	s.config.PeriodicJobs = defs.PeriodicJobs

	client, clientErr := river.NewClient(riverpgxv5.New(s.pool), s.config)
	if clientErr != nil {
		return fmt.Errorf("create river client: %w", clientErr)
	}
	s.client = client
	s.logger.Info("finalized job service", "kinds", defs.Kinds)
	return nil
}

var ErrNotFinalized = fmt.Errorf("job service is not finalized")

func (s *JobService) getClient() (*riverClient, error) {
	if s.client == nil {
		return nil, ErrNotFinalized
	}
	return s.client, nil
}

func (s *JobService) Lifecycle() *rez.ServiceLifecycle {
	runFn := func(ctx context.Context) error {
		client, clientErr := s.getClient()
		if clientErr != nil {
			return clientErr
		}
		clientCtx := execution.NewRootContext(ctx, execution.KindSystem, execution.SourceJob)
		if startErr := client.Start(clientCtx); startErr != nil {
			return fmt.Errorf("start river client: %w", startErr)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-client.Stopped():
			if ctx.Err() == nil {
				return fmt.Errorf("river client stopped unexpectedly")
			}
			return nil
		}
	}
	return &rez.ServiceLifecycle{
		StartFns: []rez.LifecycleFunc{runFn},
		StopFn:   s.stop,
	}
}

func (s *JobService) stop(ctx context.Context) error {
	if client, _ := s.getClient(); client != nil {
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
	client, clientErr := s.getClient()
	if clientErr != nil {
		return nil, clientErr
	}
	if isTx, pgxTx, txErr := s.extractContextPgxTx(ctx); isTx {
		if txErr != nil {
			return nil, fmt.Errorf("extract pgx tx: %w", txErr)
		}
		return client.InsertTx(ctx, pgxTx, args, opts)
	}
	return client.Insert(ctx, args, opts)
}

func (s *JobService) InsertMany(ctx context.Context, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error) {
	client, clientErr := s.getClient()
	if clientErr != nil {
		return nil, clientErr
	}
	if isTx, pgxTx, txErr := s.extractContextPgxTx(ctx); isTx {
		if txErr != nil {
			return nil, fmt.Errorf("extract pgx tx: %w", txErr)
		}
		return client.InsertManyTx(ctx, pgxTx, params)
	}
	return client.InsertMany(ctx, params)
}

func (s *JobService) Cancel(ctx context.Context, jobID int64) error {
	client, clientErr := s.getClient()
	if clientErr != nil {
		return clientErr
	}
	if isTx, pgxTx, txErr := s.extractContextPgxTx(ctx); isTx {
		if txErr != nil {
			return fmt.Errorf("extract pgx tx: %w", txErr)
		}
		if _, cancelErr := client.JobCancelTx(ctx, pgxTx, jobID); cancelErr != nil {
			return fmt.Errorf("cancel job: %w", cancelErr)
		}
		return nil
	}
	if _, cancelErr := client.JobCancel(ctx, jobID); cancelErr != nil {
		return fmt.Errorf("cancel job: %w", cancelErr)
	}
	return nil
}
