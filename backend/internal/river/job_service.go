package river

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
)

var (
	NonCompletedJobStates = []rivertype.JobState{rivertype.JobStateAvailable, rivertype.JobStatePending, rivertype.JobStateRunning, rivertype.JobStateRetryable, rivertype.JobStateScheduled}
)

type JobService struct {
	driver    riverdriver.Driver[pgx.Tx]
	clientCfg *river.Config
	client    *river.Client[pgx.Tx]
}

func NewJobService(pool *pgxpool.Pool) *JobService {
	queues := map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: 20},
	}

	slogOpts := slogzerolog.Option{
		Level:  slog.LevelInfo,
		Logger: zerolog.DefaultContextLogger,
	}

	cfg := &river.Config{
		Workers:      river.NewWorkers(),
		Queues:       queues,
		Logger:       slog.New(slogOpts.NewZerologHandler()),
		PeriodicJobs: []*river.PeriodicJob{},
	}

	return &JobService{
		driver:    riverpgxv5.New(pool),
		clientCfg: cfg,
	}
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

func (s *JobService) addPeriodicJob(j *river.PeriodicJob) {
	s.clientCfg.PeriodicJobs = append(s.clientCfg.PeriodicJobs, j)
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

func (s *JobService) insertTx(ctx context.Context, tx *ent.Tx, args river.JobArgs, opts *river.InsertOpts) error {
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

func (s *JobService) RequestGenerateIncidentDebriefResponse(ctx context.Context, tx *ent.Tx, id uuid.UUID) error {
	return s.insertTx(ctx, tx, generateIncidentDebriefResponseJobArgs{DebriefId: id}, nil)
}

func (s *JobService) RequestSendUserDebriefRequests(ctx context.Context, tx *ent.Tx, incidentId uuid.UUID) error {
	return s.insertTx(ctx, tx, sendIncidentDebriefRequestsJobArgs{IncidentId: incidentId}, nil)
}

func (s *JobService) RequestGenerateIncidentDebriefSuggestions(ctx context.Context, tx *ent.Tx, debriefId uuid.UUID) error {
	return s.insertTx(ctx, tx, generateIncidentDebriefSuggestionsJobArgs{DebriefId: debriefId}, nil)
}
