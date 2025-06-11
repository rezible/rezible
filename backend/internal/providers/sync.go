package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/gosimple/slug"

	"entgo.io/ent/dialect/sql"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providersynchistory"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/jobs"
)

func NewDataSyncer(db *ent.Client, loader *Loader) *DataSyncer {
	return &DataSyncer{db: db, l: loader}
}

type DataSyncer struct {
	db *ent.Client
	l  *Loader
}

func (s *DataSyncer) RegisterPeriodicSyncJob(j rez.JobsService, interval time.Duration) error {
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

func (s *DataSyncer) SyncData(ctx context.Context, args jobs.SyncProviderData) error {
	if args.Hard {
		// TODO: maybe just pass a flag?
		s.db.ProviderSyncHistory.Delete().ExecX(ctx)
	}

	if args.Teams {
		teamsProv, provErr := s.l.LoadTeamDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load teams data provider: %w", provErr)
		}
		syncer := newTeamDataSyncer(s.db, teamsProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("teams sync failed: %w", syncErr)
		}
	}

	if args.Users {
		usersProv, provErr := s.l.LoadUserDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load users data provider: %w", provErr)
		}
		syncer := newUserDataSyncer(s.db, usersProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("users sync failed: %w", syncErr)
		}
	}

	if args.Oncall {
		oncallProv, provErr := s.l.LoadOncallDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		syncer := newOncallDataSyncer(s.db, oncallProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("oncall sync failed: %w", syncErr)
		}
	}

	if args.OncallEvents {
		oncallEventsProv, provErr := s.l.LoadOncallEventsDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall events data provider: %w", provErr)
		}
		syncer := newOncallEventsDataSyncer(s.db, oncallEventsProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("oncall events sync failed: %w", syncErr)
		}
	}

	if args.Incidents {
		incProv, provErr := s.l.LoadIncidentDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		syncer := newIncidentDataSyncer(s.db, incProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("incidents sync failed: %w", syncErr)
		}
	}

	if args.SystemComponents {
		cmpProv, provErr := s.l.LoadSystemComponentsDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		syncer := newSystemComponentsDataSyncer(s.db, cmpProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("system components sync failed: %w", syncErr)
		}
	}
	return nil
}

func getLastSyncTime(ctx context.Context, db *ent.Client, dataType string) time.Time {
	lastSync, queryErr := db.ProviderSyncHistory.Query().
		Where(providersynchistory.DataType(dataType)).
		Order(providersynchistory.ByFinishedAt(sql.OrderDesc())).
		First(ctx)
	if queryErr != nil {
		return time.Time{}
	}
	return lastSync.FinishedAt
}

func saveSyncHistory(ctx context.Context, db *ent.Client, start time.Time, num int, dataType string) error {
	return db.ProviderSyncHistory.Create().
		SetStartedAt(start).
		SetFinishedAt(time.Now()).
		SetNumMutations(num).
		SetDataType(dataType).
		Exec(ctx)
}

func applySyncMutations(ctx context.Context, client *ent.Client, mutations []ent.Mutation) error {
	if len(mutations) == 0 {
		return nil
	}

	return ent.WithTx(ctx, client, func(tx *ent.Tx) error {
		for _, m := range mutations {
			// fmt.Printf("applying %s %s mutation\n", m.Type(), m.Op().String())
			if _, mutErr := tx.Client().Mutate(ctx, m); mutErr != nil {
				return mutErr
			}
		}
		return nil
	})
}

// TODO: just do this in postgres
type slugTracker struct {
	prefixCount map[string]int
}

func newSlugTracker() *slugTracker {
	return &slugTracker{
		prefixCount: make(map[string]int),
	}
}

func (s *slugTracker) reset() {
	s.prefixCount = make(map[string]int)
}

func (s *slugTracker) generateUnique(title string, initialCountFn func(string) (int, error)) (string, error) {
	base := slug.MakeLang(title, "en")

	numExisting, ok := s.prefixCount[base]
	if !ok || numExisting == 0 {
		var countErr error
		numExisting, countErr = initialCountFn(base)
		if countErr != nil {
			return "", countErr
		}
	}

	count := numExisting + 1
	s.prefixCount[base] = count

	if count > 1 {
		return fmt.Sprintf("%s-%d", base, count), nil
	}
	return base, nil
}

// TODO: userTracker ?
func getUserByEmail(ctx context.Context, dbc *ent.Client, email string) (*ent.User, error) {
	// TODO: cache?
	return dbc.User.Query().Where(user.Email(email)).First(ctx)
}
