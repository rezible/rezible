package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"github.com/twohundreds/rezible/ent"
)

// river type aliases
type (
	JobArgs               = river.JobArgs
	JobArgsWithInsertOpts = river.JobArgsWithInsertOpts
	InsertOpts            = river.InsertOpts
	JobInsertResult       = rivertype.JobInsertResult
	UniqueOpts            = river.UniqueOpts

	PeriodicJobOpts = river.PeriodicJobOpts
	PeriodicJob     = struct {
		Schedule    river.PeriodicSchedule
		Constructor river.PeriodicJobConstructor
		Options     *PeriodicJobOpts
	}

	WorkersRegistry = river.Workers

	// TODO: go 1.24 will allow generic type aliases
	//Job[T JobArgs] = river.Job[T]
	//Worker[T JobArgs] = river.Worker[T]
)

var (
	NonCompletedJobStates = []rivertype.JobState{rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning, rivertype.JobStateRetryable, rivertype.JobStateScheduled}
)

func RegisterWorker[T JobArgs](c *BackgroundJobClient, worker river.Worker[T]) error {
	return river.AddWorkerSafely(c.workers, worker)
}

func PeriodicInterval(interval time.Duration) river.PeriodicSchedule {
	return river.PeriodicInterval(interval)
}

type BackgroundJobClient struct {
	pool         *pgxpool.Pool
	workers      *river.Workers
	queues       map[string]river.QueueConfig
	logger       *slog.Logger
	periodicJobs []PeriodicJob
	client       *river.Client[pgx.Tx]
}

func NewBackgroundJobClient(logger *zerolog.Logger, pool *pgxpool.Pool) *BackgroundJobClient {
	queues := map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 100},
	}

	slogOpts := slogzerolog.Option{
		Level:  slog.LevelInfo,
		Logger: logger,
	}

	c := &BackgroundJobClient{
		pool:         pool,
		workers:      river.NewWorkers(),
		logger:       slog.New(slogOpts.NewZerologHandler()),
		queues:       queues,
		periodicJobs: []PeriodicJob{},
	}

	return c
}

func (c *BackgroundJobClient) AddPeriodicJob(j PeriodicJob) {
	c.periodicJobs = append(c.periodicJobs, j)
}

func (c *BackgroundJobClient) Start(ctx context.Context) error {
	periodicJobs := make([]*river.PeriodicJob, len(c.periodicJobs))
	for i, j := range c.periodicJobs {
		periodicJobs[i] = river.NewPeriodicJob(j.Schedule, j.Constructor, j.Options)
	}

	riverCfg := &river.Config{
		Workers:      c.workers,
		Queues:       c.queues,
		Logger:       c.logger,
		PeriodicJobs: periodicJobs,
	}

	client, clientErr := river.NewClient(riverpgxv5.New(c.pool), riverCfg)
	if clientErr != nil {
		return fmt.Errorf("could not create river client: %w", clientErr)
	}
	c.client = client

	return c.client.Start(ctx)
}

func (c *BackgroundJobClient) Stop(ctx context.Context) error {
	return c.client.Stop(ctx)
}

func (c *BackgroundJobClient) Insert(ctx context.Context, args JobArgs, opts *InsertOpts) (*JobInsertResult, error) {
	return c.client.Insert(ctx, args, opts)
}

func (c *BackgroundJobClient) InsertTx(ctx context.Context, tx *ent.Tx, args JobArgs, opts *InsertOpts) (*JobInsertResult, error) {
	pgxTx, pgErr := ent.ExtractPgxTx(tx)
	if pgErr != nil {
		return nil, fmt.Errorf("not using pgx driver: %w", pgErr)
	}
	return c.client.InsertTx(ctx, pgxTx, args, opts)
}
