package datasyncer

import (
	"context"
	"fmt"
	"iter"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providersynchistory"
)

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

	if setupErr := ds.batcher.setup(ctx); setupErr != nil {
		return fmt.Errorf("setup: %w", setupErr)
	}

	lastSync := ds.getLastSyncTime(ctx)
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	var batch []T
	var numMutations int

	batchSize := 10
	for prov, pullErr := range ds.batcher.pullData(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pullData: %w", pullErr)
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

	if len(batch) > 0 {
		lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
		if batchErr != nil {
			return batchErr
		}
		numMutations += lastBatchMuts
	}

	if delMuts := ds.batcher.getDeletionMutations(); len(delMuts) > 0 {
		log.Debug().Int("numMutations", numMutations).Msg("deletion mutations")
	}

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
