package river

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/jobs"
	"github.com/riverqueue/river/rivertype"
	"github.com/rs/zerolog/log"

	"log/slog"

	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rezible/rezible/ent"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type (
	pgxClient  = river.Client[pgx.Tx]
	JobService struct {
		client *pgxClient
	}
)

func NewJobService(pool *pgxpool.Pool) (*JobService, error) {
	middleware := []rivertype.Middleware{
		&accessContextMiddleware{},
	}

	queues := map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 20},
	}

	slogOpts := slogzerolog.Option{
		Level:  slog.LevelInfo,
		Logger: zerolog.DefaultContextLogger,
	}

	cfg := &river.Config{
		Middleware:   middleware,
		Workers:      jobs.Workers,
		Queues:       queues,
		Logger:       slog.New(slogOpts.NewZerologHandler()),
		PeriodicJobs: jobs.PeriodicJobs,
	}

	client, clientErr := river.NewClient(riverpgxv5.New(pool), cfg)
	if clientErr != nil {
		return nil, fmt.Errorf("failed to create client: %w", clientErr)
	}

	svc := &JobService{client: client}

	return svc, nil
}

func (s *JobService) Start(ctx context.Context) error {
	pj := s.client.PeriodicJobs()
	_, pjErr := pj.AddManySafely(jobs.PeriodicJobs)
	if pjErr != nil {
		return fmt.Errorf("failed to add periodic jobs: %w", pjErr)
	}
	return s.client.Start(access.SystemContext(ctx))
}

func (s *JobService) Stop(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.Stop(ctx)
}

func (s *JobService) Insert(ctx context.Context, args jobs.JobArgs, opts *jobs.InsertOpts) error {
	res, insertErr := s.client.Insert(ctx, args, opts)
	if insertErr != nil {
		return fmt.Errorf("could not insert job: %w", insertErr)
	}
	log.Debug().
		Str("kind", args.Kind()).
		Bool("skipped_unique", res.UniqueSkippedAsDuplicate).
		Msg("inserted job")
	return nil
}

func (s *JobService) InsertMany(ctx context.Context, params []jobs.InsertManyParams) error {
	_, insertErr := s.client.InsertMany(ctx, params)
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
	_, insertErr := s.client.InsertTx(ctx, pgxTx, args, opts)
	if insertErr != nil {
		return fmt.Errorf("could not insert job in tx: %w", insertErr)
	}
	return nil
}
