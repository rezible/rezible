package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
)

type userDataSyncer struct {
	db       *ent.Client
	provider rez.UserDataProvider

	mutations []ent.Mutation
}

func newUserDataSyncer(db *ent.Client, prov rez.UserDataProvider) *userDataSyncer {
	ds := &userDataSyncer{db: db, provider: prov}
	ds.resetState()
	return ds
}

func (ds *userDataSyncer) resetState() {
	ds.mutations = make([]ent.Mutation, 0)
}

func (ds *userDataSyncer) saveSyncHistory(ctx context.Context, start time.Time, num int, dataType string) {
	historyErr := ds.db.ProviderSyncHistory.Create().
		SetStartedAt(start).
		SetFinishedAt(time.Now()).
		SetNumMutations(num).
		SetDataType(dataType).
		Exec(ctx)
	if historyErr != nil {
		log.Error().Err(historyErr).Msg("failed to save sync history")
	}
}

func (ds *userDataSyncer) syncProviderData(ctx context.Context) error {
	start := time.Now()

	lastSync := getLastSyncTime(ctx, ds.db, "users")
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	if usersErr := ds.syncAllProviderUsers(ctx); usersErr != nil {
		return fmt.Errorf("users: %w", usersErr)
	}
	log.Info().
		Msg("users data sync complete")

	return nil
}

func (ds *userDataSyncer) syncAllProviderUsers(ctx context.Context) error {
	var batch []*ent.User

	start := time.Now()
	var numMutations int

	batchSize := 10
	for usr, pullErr := range ds.provider.PullUsers(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull users: %w", pullErr)
		}

		batch = append(batch, usr)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncBatch(ctx, batch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			batch = make([]*ent.User, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	ds.saveSyncHistory(ctx, start, numMutations, "users")

	return nil
}

func (ds *userDataSyncer) syncBatch(ctx context.Context, batch []*ent.User) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	ds.resetState()
	syncErr := ds.createBatchSyncMutations(ctx, batch)
	if syncErr != nil {
		return 0, fmt.Errorf("building mutations: %w", syncErr)
	}

	if applyErr := applySyncMutations(ctx, ds.db, ds.mutations); applyErr != nil {
		return 0, fmt.Errorf("applying mutations: %w", applyErr)
	}
	numMutations := len(ds.mutations)
	ds.resetState()

	return numMutations, nil
}

func (ds *userDataSyncer) createBatchSyncMutations(ctx context.Context, batch []*ent.User) error {
	emails := make([]string, len(batch))
	for i, u := range batch {
		emails[i] = u.Email
	}

	dbUsers, queryErr := ds.db.User.Query().Where(user.EmailIn(emails...)).All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying users: %w", queryErr)
	}
	dbEmailMap := make(map[string]*ent.User)
	for _, usr := range dbUsers {
		u := usr
		dbEmailMap[u.Email] = usr
	}

	for _, provUser := range batch {
		dbUser, exists := dbEmailMap[provUser.Email]
		if exists {
			// don't delete this user
		}
		_ = ds.syncUser(dbUser, provUser)
	}

	return nil
}

func (ds *userDataSyncer) syncUser(db, prov *ent.User) uuid.UUID {
	var m *ent.UserMutation
	var userId uuid.UUID
	needsSync := true
	if db == nil {
		userId = uuid.New()
		m = ds.db.User.Create().SetID(userId).Mutation()
	} else {
		userId = db.ID
		m = ds.db.User.UpdateOneID(userId).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Name != prov.Name || db.Email != prov.Email || db.Timezone != prov.Timezone || db.ChatID != prov.ChatID
	}

	m.SetName(prov.Name)
	m.SetEmail(prov.Email)
	m.SetChatID(prov.ChatID)
	m.SetTimezone(prov.Timezone)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return userId
}
