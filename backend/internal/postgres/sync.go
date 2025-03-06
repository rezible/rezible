package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/gosimple/slug"

	"entgo.io/ent/dialect/sql"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providersynchistory"
)

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
