package river

import (
	"context"
	"fmt"
	"time"

	"github.com/rezible/rezible/jobs"

	"log/slog"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"

	"github.com/rezible/rezible/ent"
)

var (
	workers      = river.NewWorkers()
	periodicJobs []*river.PeriodicJob
)

type (
	pgxClient  = river.Client[pgx.Tx]
	JobService struct {
		client *pgxClient
	}

	Job[T jobs.JobArgs]    = river.Job[T]
	Worker[T jobs.JobArgs] = river.Worker[T]
)

func NewJobService(pool *pgxpool.Pool) (*JobService, error) {
	queues := map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 20},
	}

	slogOpts := slogzerolog.Option{
		Level:  slog.LevelInfo,
		Logger: zerolog.DefaultContextLogger,
	}

	cfg := &river.Config{
		Workers:      workers,
		Queues:       queues,
		Logger:       slog.New(slogOpts.NewZerologHandler()),
		PeriodicJobs: periodicJobs,
	}

	client, clientErr := river.NewClient(riverpgxv5.New(pool), cfg)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to create client: %w", clientErr)
	}

	svc := &JobService{client: client}

	return svc, nil
}

func RegisterWorkerFunc[A jobs.JobArgs](work jobs.WorkFn[A]) {
	river.AddWorker[A](workers, river.WorkFunc(func(ctx context.Context, j *river.Job[A]) error {
		return work(ctx, j.Args)
	}))
}

func RegisterPeriodicJob[A jobs.JobArgs](job jobs.PeriodicJob, workFn jobs.WorkFn[A]) {
	constructor := func() (river.JobArgs, *river.InsertOpts) {
		params := job.ConstructorFunc()
		return params.Args, convertInsertOpts(params)
	}
	riverJob := river.NewPeriodicJob(periodicInterval(job.Interval), constructor, convertPeriodicOpts(job.Opts))
	periodicJobs = append(periodicJobs, riverJob)
	RegisterWorkerFunc[A](workFn)
}

type periodicIntervalSchedule struct {
	interval time.Duration
}

func periodicInterval(interval time.Duration) river.PeriodicSchedule {
	return &periodicIntervalSchedule{interval}
}
func (s *periodicIntervalSchedule) Next(t time.Time) time.Time {
	return t.Add(s.interval)
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
	pj := s.client.PeriodicJobs()
	_, pjErr := pj.AddManySafely(periodicJobs)
	if pjErr != nil {
		return fmt.Errorf("failed to add periodic jobs: %w", pjErr)
	}
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

func convertInsertOpts(p jobs.InsertJobParams) *river.InsertOpts {
	riverOpts := &river.InsertOpts{
		ScheduledAt: p.ScheduledAt,
	}
	if p.Uniqueness != nil {
		uOpts := river.UniqueOpts{
			ByArgs:      p.Uniqueness.Args,
			ByPeriod:    p.Uniqueness.ByPeriod,
			ByQueue:     p.Uniqueness.ByQueue,
			ExcludeKind: p.Uniqueness.ExcludeKind,
		}

		if p.Uniqueness.ByState != nil {
			uOpts.ByState = make([]rivertype.JobState, len(p.Uniqueness.ByState))
			for i, s := range p.Uniqueness.ByState {
				uOpts.ByState[i] = rivertype.JobState(s)
			}
		}

		riverOpts.UniqueOpts = uOpts
	}
	return riverOpts
}

func (s *JobService) Insert(ctx context.Context, params jobs.InsertJobParams) error {
	_, insertErr := s.client.Insert(ctx, params.Args, convertInsertOpts(params))
	if insertErr != nil {
		return fmt.Errorf("could not insert job: %w", insertErr)
	}
	return nil
}

func (s *JobService) InsertMany(ctx context.Context, params []jobs.InsertJobParams) error {
	insertParams := make([]river.InsertManyParams, len(params))
	for i, p := range params {
		insertParams[i] = river.InsertManyParams{
			Args:       p.Args,
			InsertOpts: convertInsertOpts(p),
		}
	}
	_, insertErr := s.client.InsertMany(ctx, insertParams)
	if insertErr != nil {
		return fmt.Errorf("could not insert jobs: %w", insertErr)
	}
	return nil
}

func (s *JobService) InsertTx(ctx context.Context, tx *ent.Tx, params jobs.InsertJobParams) error {
	pgxTx, pgErr := ent.ExtractPgxTx(tx)
	if pgErr != nil {
		return fmt.Errorf("not using pgx driver: %w", pgErr)
	}
	_, insertErr := s.client.InsertTx(ctx, pgxTx, params.Args, convertInsertOpts(params))
	if insertErr != nil {
		return fmt.Errorf("could not insert job in tx: %w", insertErr)
	}
	return nil
}
