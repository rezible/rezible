package postgres

import (
	"context"
	"time"

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
