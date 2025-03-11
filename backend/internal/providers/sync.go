package providers

import (
	"context"
	"fmt"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/jobs"
	"time"

	"github.com/gosimple/slug"

	"entgo.io/ent/dialect/sql"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providersynchistory"
)

type DataSyncer interface {
	GetSyncMutations(ctx context.Context) error
}

func SyncData(ctx context.Context, args *jobs.SyncProviderData, dbc *ent.Client, l *Loader) error {
	if args.Hard {
		// TODO: maybe just pass a flag?
		dbc.ProviderSyncHistory.Delete().ExecX(ctx)
	}

	if args.Teams {
		teamsProv, provErr := l.LoadTeamDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load teams data provider: %w", provErr)
		}
		syncer := newTeamDataSyncer(dbc, teamsProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("teams sync failed: %w", syncErr)
		}
	}

	if args.Users {
		usersProv, provErr := l.LoadUserDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load users data provider: %w", provErr)
		}
		syncer := newUserDataSyncer(dbc, usersProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("users sync failed: %w", syncErr)
		}
	}

	if args.Oncall {
		oncallProv, provErr := l.LoadOncallDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		syncer := newOncallDataSyncer(dbc, oncallProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("oncall sync failed: %w", syncErr)
		}
	}

	if args.Incidents {
		incProv, provErr := l.LoadIncidentDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		syncer := newIncidentDataSyncer(dbc, incProv)
		if syncErr := syncer.SyncProviderData(ctx); syncErr != nil {
			return fmt.Errorf("incidents sync failed: %w", syncErr)
		}
	}

	if args.SystemComponents {
		cmpProv, provErr := l.LoadSystemComponentsDataProvider(ctx)
		if provErr != nil {
			return fmt.Errorf("failed to load oncall data provider: %w", provErr)
		}
		syncer := newSystemComponentsDataSyncer(dbc, cmpProv)
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
	existingSlugs map[string]int
	newSlugs      map[string]int
}

func newSlugTracker() *slugTracker {
	return &slugTracker{
		newSlugs:      make(map[string]int),
		existingSlugs: make(map[string]int),
	}
}

func (s *slugTracker) reset() {
	s.newSlugs = make(map[string]int)
	s.existingSlugs = make(map[string]int)
}

func (s *slugTracker) generateUnique(base string, countFn func(string) (int, error)) (string, error) {
	tmp := slug.MakeLang(base, "en")

	numExisting := s.existingSlugs[tmp]
	if numExisting == 0 {
		var countErr error
		numExisting, countErr = countFn(tmp)
		if countErr != nil {
			return "", countErr
		}
		s.existingSlugs[tmp] = numExisting
	}
	numNew := s.newSlugs[tmp]

	slugCount := numExisting + numNew + 1
	if slugCount > 1 {
		tmp = fmt.Sprintf("%s-%d", tmp, slugCount)
	}
	s.newSlugs[tmp] = slugCount

	return tmp, nil
}

// TODO: userTracker ?
func getUserByEmail(ctx context.Context, dbc *ent.Client, email string) (*ent.User, error) {
	// TODO: cache?
	return dbc.User.Query().Where(user.Email(email)).First(ctx)
}
