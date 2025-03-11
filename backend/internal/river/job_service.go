package river

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/jobs"
)

var (
	workers *river.Workers
)

type (
	pgxClient  = river.Client[pgx.Tx]
	JobService struct {
		driver    riverdriver.Driver[pgx.Tx]
		clientCfg *river.Config
		client    *pgxClient
	}

	// Job[T jobs.JobArgs]    = river.Job[T]
	// Worker[T jobs.JobArgs] = river.Worker[T]
)

func NewJobService(pool *pgxpool.Pool) (*JobService, error) {
	queues := map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 20},
	}

	slogOpts := slogzerolog.Option{
		Level:  slog.LevelInfo,
		Logger: zerolog.DefaultContextLogger,
	}

	workers = river.NewWorkers()

	if regErr := jobs.SetWorkerRegistry(workers); regErr != nil {
		return nil, fmt.Errorf("failed to set river.Worker registry: %w", regErr)
	}

	cfg := &river.Config{
		Workers:      workers,
		Queues:       queues,
		Logger:       slog.New(slogOpts.NewZerologHandler()),
		PeriodicJobs: []*river.PeriodicJob{},
	}

	svc := &JobService{
		driver:    riverpgxv5.New(pool),
		clientCfg: cfg,
	}

	return svc, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	cfg := &rivermigrate.Config{}
	migrator, migErr := rivermigrate.New(riverpgxv5.New(pool), cfg)
	if migErr != nil {
		return fmt.Errorf("failed to create migrator: %w", migErr)
	}

	opts := &rivermigrate.MigrateOpts{}
	res, migrationErr := migrator.Migrate(ctx, rivermigrate.DirectionUp, opts)
	if migrationErr != nil {
		return fmt.Errorf("failed to migrate: %w", migrationErr)
	}

	if len(res.Versions) > 0 {
		log.Info().Int("versions", len(res.Versions)).Msg("ran river migrations")
	}

	return nil
}

func (s *JobService) Start(ctx context.Context) error {
	client, clientErr := river.NewClient(s.driver, s.clientCfg)
	if clientErr != nil {
		return fmt.Errorf("could not create river client: %w", clientErr)
	}
	s.client = client

	return s.client.Start(ctx)
}

func (s *JobService) Stop(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.Stop(ctx)
}

func convertPeriodicOpts(o *jobs.PeriodicJobOpts) *river.PeriodicJobOpts {
	if o == nil {
		return nil
	}
	return &river.PeriodicJobOpts{
		RunOnStart: o.RunOnStart,
	}
}

func convertInsertOpts(opts *jobs.InsertOpts) *river.InsertOpts {
	if opts == nil {
		return nil
	}
	riverOpts := &river.InsertOpts{}
	if opts.Uniqueness != nil {
		uOpts := river.UniqueOpts{
			ByArgs:      opts.Uniqueness.Args,
			ByPeriod:    opts.Uniqueness.ByPeriod,
			ByQueue:     opts.Uniqueness.ByQueue,
			ExcludeKind: opts.Uniqueness.ExcludeKind,
		}

		if opts.Uniqueness.ByState != nil {
			uOpts.ByState = make([]rivertype.JobState, len(opts.Uniqueness.ByState))
			for i, s := range opts.Uniqueness.ByState {
				uOpts.ByState[i] = rivertype.JobState(s)
			}
		}

		riverOpts.UniqueOpts = uOpts
	}
	return riverOpts
}

func (s *JobService) Insert(ctx context.Context, args jobs.JobArgs, opts *jobs.InsertOpts) error {
	_, insertErr := s.client.Insert(ctx, args, convertInsertOpts(opts))
	if insertErr != nil {
		return fmt.Errorf("could not insert job: %w", insertErr)
	}
	return nil
}

func (s *JobService) InsertMany(ctx context.Context, params []jobs.InsertManyParams) error {
	insertParams := make([]river.InsertManyParams, len(params))
	for i, p := range params {
		insertParams[i] = river.InsertManyParams{
			Args:       p.Args,
			InsertOpts: convertInsertOpts(p.Opts),
		}
	}
	_, insertErr := s.client.InsertMany(ctx, insertParams)
	if insertErr != nil {
		return fmt.Errorf("could not insert jobs: %w", insertErr)
	}
	return nil
}

func (s *JobService) InsertTx(ctx context.Context, tx *ent.Tx, args jobs.JobArgs, opts *jobs.InsertOpts) error {
	pgxTx, pgErr := ent.ExtractPgxTx(tx)
	if pgErr != nil {
		return fmt.Errorf("not using pgx driver: %w", pgErr)
	}
	_, insertErr := s.client.InsertTx(ctx, pgxTx, args, convertInsertOpts(opts))
	if insertErr != nil {
		return fmt.Errorf("could not insert job in tx: %w", insertErr)
	}
	return nil
}

func (s *JobService) RegisterPeriodicJob(j *jobs.PeriodicJob) {
	constructor := func() (river.JobArgs, *river.InsertOpts) {
		args, opts := j.ConstructorFunc()
		return args, convertInsertOpts(opts)
	}
	riverJob := river.NewPeriodicJob(j.Schedule, constructor, convertPeriodicOpts(j.Opts))
	s.clientCfg.PeriodicJobs = append(s.clientCfg.PeriodicJobs, riverJob)
}
