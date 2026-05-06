package datasync

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"reflect"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/providersynchistory"
)

type dataBatcher[T any] interface {
	setup(context.Context) error
	pullData(context.Context) iter.Seq2[*T, error]
	createBatchMutations(context.Context, []*T) ([]ent.Mutation, error)
	getDeletionMutations() []ent.Mutation
}

type batchedDataSyncer[T any] struct {
	db           *ent.Client
	dataType     string
	providerType string
	batcher      dataBatcher[T]
	syncInterval time.Duration
	opts         SyncOptions
	metrics      *metrics
}

func newBatchedDataSyncer[T any](db *ent.Client, dataType string, batcher dataBatcher[T], opts SyncOptions, met *metrics) *batchedDataSyncer[T] {
	return &batchedDataSyncer[T]{
		db:           db,
		dataType:     dataType,
		providerType: reflect.TypeOf(batcher).String(),
		batcher:      batcher,
		syncInterval: time.Minute * 30,
		opts:         opts,
		metrics:      met,
	}
}

func (ds *batchedDataSyncer[T]) setSyncInterval(duration time.Duration) {
	ds.syncInterval = duration
}

func (ds *batchedDataSyncer[T]) Sync(ctx context.Context) error {
	res, err := ds.sync(ctx)
	ds.metrics.recordRun(ctx, ds.dataType, ds.providerType, res, err)
	return err
}

type syncResult struct {
	skipped       bool
	setupTime     time.Duration
	syncTime      time.Duration
	recordsCount  int64
	mutationCount int64
}

func (ds *batchedDataSyncer[T]) sync(ctx context.Context) (*syncResult, error) {
	res := &syncResult{}

	start := time.Now()
	nextSyncTime := ds.getLastSyncTime(ctx).Add(ds.syncInterval)
	if time.Now().Before(nextSyncTime) {
		res.setupTime = time.Since(start)
		res.skipped = true
		return res, nil
	}

	setupErr := ds.batcher.setup(ctx)
	res.setupTime = time.Since(start)
	if setupErr != nil {
		return res, fmt.Errorf("setup: %w", setupErr)
	}

	var batch []*T

	syncBatch := func() error {
		if len(batch) == 0 {
			return nil
		}
		res.recordsCount += int64(len(batch))

		batchMutations, batchSyncErr := ds.batcher.createBatchMutations(ctx, batch)
		if batchSyncErr != nil {
			return fmt.Errorf("building mutations: %w", batchSyncErr)
		}
		numBatchMutations := len(batchMutations)

		if applyErr := ds.applySyncMutations(ctx, batchMutations); applyErr != nil {
			return fmt.Errorf("applying mutations: %w", applyErr)
		}

		res.mutationCount += int64(numBatchMutations)
		return nil
	}

	syncStart := time.Now()
	batchSyncSize := 10
	for prov, pullErr := range ds.batcher.pullData(ctx) {
		if pullErr != nil {
			return res, fmt.Errorf("pullData: %w", pullErr)
		}
		batch = append(batch, prov)

		if len(batch) >= batchSyncSize {
			if batchErr := syncBatch(); batchErr != nil {
				return res, fmt.Errorf("batch: %w", batchErr)
			}
			batch = make([]*T, 0)
		}
	}
	if batchErr := syncBatch(); batchErr != nil {
		return res, fmt.Errorf("batch: %w", batchErr)
	}
	res.syncTime = time.Since(syncStart)

	if delMuts := ds.batcher.getDeletionMutations(); len(delMuts) > 0 {
		slog.Debug("deletion mutations", "numMutations", res.mutationCount)
	}

	if saveErr := ds.saveSyncHistory(ctx, start, res.mutationCount); saveErr != nil {
		slog.Error("failed to save data sync history", "error", saveErr, "dataType", ds.dataType)
	}

	return res, nil
}

func (ds *batchedDataSyncer[T]) getLastSyncTime(ctx context.Context) time.Time {
	if ds.opts.IgnoreHistory {
		return time.Time{}
	}
	last, queryErr := ds.db.ProviderSyncHistory.Query().
		Where(providersynchistory.DataType(ds.dataType)).
		Order(providersynchistory.ByFinishedAt(sql.OrderDesc())).
		Select(providersynchistory.FieldFinishedAt).
		First(ctx)
	if queryErr != nil {
		return time.Time{}
	}
	return last.FinishedAt
}

func (ds *batchedDataSyncer[T]) saveSyncHistory(ctx context.Context, start time.Time, num int64) error {
	return ds.db.ProviderSyncHistory.Create().
		SetStartedAt(start).
		SetFinishedAt(time.Now()).
		SetNumMutations(int(num)).
		SetDataType(ds.dataType).
		Exec(ctx)
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
