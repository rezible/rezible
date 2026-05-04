package river

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/jobs"
)

const SchemaName = "river"

type pgxClient = river.Client[pgx.Tx]

type JobService struct {
	logger *slog.Logger
	client *pgxClient
}

func NewJobService(pool *pgxpool.Pool) (*JobService, error) {
	queues := map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 20},
	}

	s := &JobService{
		logger: slog.New(logger{
			base:  slog.Default().With("package", "river").Handler(),
			level: slog.LevelWarn,
		}),
	}

	cfg := &river.Config{
		Schema: SchemaName,
		Middleware: []rivertype.Middleware{
			&accessContextMiddleware{},
		},
		Workers: jobs.Workers,
		Queues:  queues,
		Logger:  s.logger,
	}
	var clientErr error
	s.client, clientErr = river.NewClient(riverpgxv5.New(pool), cfg)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to create client: %w", clientErr)
	}

	return s, nil
}

func (s *JobService) Start(ctx context.Context) error {
	pj := s.client.PeriodicJobs()
	_, pjErr := pj.AddManySafely(jobs.PeriodicJobs)
	if pjErr != nil {
		return fmt.Errorf("failed to add periodic jobs: %w", pjErr)
	}
	jobsCtx := execution.NewContext(ctx, execution.KindSystem, execution.SourceJob)
	return s.client.Start(jobsCtx)
}

func (s *JobService) Stop(ctx context.Context) error {
	if s.client != nil {
		return s.client.Stop(ctx)
	}
	return nil
}

func (s *JobService) Insert(ctx context.Context, args jobs.JobArgs, opts *jobs.InsertOpts) error {
	res, insertErr := s.client.Insert(ctx, args, opts)
	if insertErr != nil {
		return fmt.Errorf("could not insert job: %w", insertErr)
	}
	s.logger.Debug("inserted job",
		"kind", args.Kind(),
		"skipped_unique", res.UniqueSkippedAsDuplicate,
	)
	return nil
}

func (s *JobService) InsertMany(ctx context.Context, params []jobs.InsertManyParams) error {
	if _, insertErr := s.client.InsertMany(ctx, params); insertErr != nil {
		return fmt.Errorf("could not insert jobs: %w", insertErr)
	}
	return nil
}

func (s *JobService) InsertTx(ctx context.Context, tx *ent.Tx, args jobs.JobArgs, opts *jobs.InsertOpts) error {
	pgxTx, pgErr := ent.ExtractPgxTx(tx)
	if pgErr != nil {
		return fmt.Errorf("not using pgx driver: %w", pgErr)
	}
	if _, insertErr := s.client.InsertTx(ctx, pgxTx, args, opts); insertErr != nil {
		return fmt.Errorf("could not insert job in tx: %w", insertErr)
	}
	return nil
}
