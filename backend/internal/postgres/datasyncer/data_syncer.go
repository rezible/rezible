package datasyncer

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/providersynchistory"
	"github.com/rezible/rezible/jobs"
	"iter"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
)

type Controller struct {
	db *ent.Client
	pl rez.ProviderLoader
}

func NewController(db *ent.Client, loader rez.ProviderLoader) *Controller {
	return &Controller{db: db, pl: loader}
}

func (s *Controller) RegisterPeriodicSyncJob(j rez.JobsService, interval time.Duration) error {
	args := &jobs.SyncProviderData{
		Users:            true,
		Teams:            true,
		Incidents:        true,
		Oncall:           true,
		OncallEvents:     true,
		SystemComponents: true,
	}

	opts := &jobs.InsertOpts{
		Uniqueness: &jobs.UniquenessOpts{
			ByState: jobs.NonCompletedJobStates,
		},
	}

	job := jobs.NewPeriodicJob(
		jobs.PeriodicInterval(interval),
		func() (jobs.JobArgs, *jobs.InsertOpts) {
			return args, opts
		},
		&jobs.PeriodicJobOpts{
			RunOnStart: true,
		},
	)

	j.RegisterPeriodicJob(job)
	return jobs.RegisterWorkerFunc(s.SyncData)
}

func (s *Controller) SyncData(ctx context.Context, args jobs.SyncProviderData) error {
	if args.Hard {
		// TODO: maybe just pass a flag?
		s.db.ProviderSyncHistory.Delete().ExecX(ctx)
	}

	if args.Teams {
		teams, teamsErr := s.pl.LoadTeamDataProvider(ctx)
		if teamsErr != nil {
			return fmt.Errorf("failed to load teams data provider: %w", teamsErr)
		}
		if syncErr := syncTeams(ctx, s.db, teams); syncErr != nil {
			return fmt.Errorf("teams: %w", syncErr)
		}
	}

	if args.Users {
		users, provErr := s.pl.LoadUserDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load users data provider: %w", provErr)
		}
		if syncErr := syncUsers(ctx, s.db, users); syncErr != nil {
			return fmt.Errorf("users: %w", syncErr)
		}
	}

	if args.Oncall {
		oncall, provErr := s.pl.LoadOncallDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		if syncErr := syncOncallRosters(ctx, s.db, oncall); syncErr != nil {
			return fmt.Errorf("oncall rosters: %w", syncErr)
		}
		if syncErr := syncOncallShifts(ctx, s.db, oncall); syncErr != nil {
			return fmt.Errorf("oncall shifts: %w", syncErr)
		}
	}

	if args.OncallEvents {
		alerts, provErr := s.pl.LoadAlertDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall events data provider: %w", provErr)
		}
		if syncErr := syncOncallEvents(ctx, s.db, alerts); syncErr != nil {
			return fmt.Errorf("oncall events: %w", syncErr)
		}
	}

	if args.Incidents {
		prov, provErr := s.pl.LoadIncidentDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		if syncErr := syncIncidentRoles(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("incident roles: %w", syncErr)
		}
		if syncErr := syncIncidents(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("incidents: %w", syncErr)
		}
	}

	if args.SystemComponents {
		prov, provErr := s.pl.LoadSystemComponentsDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		if syncErr := syncSystemComponents(ctx, s.db, prov); syncErr != nil {
			return fmt.Errorf("system components: %w", syncErr)
		}
	}

	return nil
}

type dataBatcher[T any] interface {
	setup(context.Context) error
	pullData(context.Context) iter.Seq2[T, error]
	createBatchMutations(context.Context, []T) ([]ent.Mutation, error)
	getDeletionMutations() []ent.Mutation
}

type batchedDataSyncer[T any] struct {
	db       *ent.Client
	dataType string
	batcher  dataBatcher[T]
}

func newBatchedDataSyncer[T any](db *ent.Client, dataType string, batcher dataBatcher[T]) *batchedDataSyncer[T] {
	return &batchedDataSyncer[T]{
		db:       db,
		dataType: dataType,
		batcher:  batcher,
	}
}

func (ds *batchedDataSyncer[T]) Sync(ctx context.Context) error {
	start := time.Now()

	lastSync := ds.getLastSyncTime(ctx)
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	var batch []T
	var numMutations int

	batchSize := 10
	for prov, pullErr := range ds.batcher.pullData(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull: %w", pullErr)
		}
		batch = append(batch, prov)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncBatch(ctx, batch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			batch = make([]T, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	if saveErr := ds.saveSyncHistory(ctx, start, numMutations); saveErr != nil {
		log.Error().Err(saveErr).Str("dataType", ds.dataType).Msg("failed to save data sync history")
	}

	return nil
}

func (ds *batchedDataSyncer[T]) getLastSyncTime(ctx context.Context) time.Time {
	lastSync, queryErr := ds.db.ProviderSyncHistory.Query().
		Where(providersynchistory.DataType(ds.dataType)).
		Order(providersynchistory.ByFinishedAt(sql.OrderDesc())).
		First(ctx)
	if queryErr != nil {
		return time.Time{}
	}
	return lastSync.FinishedAt
}

func (ds *batchedDataSyncer[T]) saveSyncHistory(ctx context.Context, start time.Time, num int) error {
	return ds.db.ProviderSyncHistory.Create().
		SetStartedAt(start).
		SetFinishedAt(time.Now()).
		SetNumMutations(num).
		SetDataType(ds.dataType).
		Exec(ctx)
}

func (ds *batchedDataSyncer[T]) syncBatch(ctx context.Context, batch []T) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	batchMutations, syncErr := ds.batcher.createBatchMutations(ctx, batch)
	if syncErr != nil {
		return 0, fmt.Errorf("building mutations: %w", syncErr)
	}

	if applyErr := ds.applySyncMutations(ctx, batchMutations); applyErr != nil {
		return 0, fmt.Errorf("applying mutations: %w", applyErr)
	}

	return len(batchMutations), nil
}

func (ds *batchedDataSyncer[T]) applySyncMutations(ctx context.Context, mutations []ent.Mutation) error {
	if len(mutations) == 0 {
		return nil
	}

	return ent.WithTx(ctx, ds.db, func(tx *ent.Tx) error {
		for _, m := range mutations {
			// fmt.Printf("applying %s %s mutation\n", m.Type(), m.Op().String())
			if _, mutErr := tx.Client().Mutate(ctx, m); mutErr != nil {
				return mutErr
			}
		}
		return nil
	})
}
