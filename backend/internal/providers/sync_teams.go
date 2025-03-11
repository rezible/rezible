package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/team"
)

type teamDataSyncer struct {
	db       *ent.Client
	provider rez.TeamDataProvider

	slugs     *slugTracker
	mutations []ent.Mutation
}

func newTeamDataSyncer(db *ent.Client, prov rez.TeamDataProvider) *teamDataSyncer {
	ds := &teamDataSyncer{db: db, provider: prov, slugs: newSlugTracker()}
	ds.resetState()
	return ds
}

func (ds *teamDataSyncer) resetState() {
	ds.mutations = make([]ent.Mutation, 0)
	ds.slugs.reset()
}

func (ds *teamDataSyncer) saveSyncHistory(ctx context.Context, start time.Time, num int, dataType string) {
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

func (ds *teamDataSyncer) SyncProviderData(ctx context.Context) error {
	start := time.Now()

	lastSync := getLastSyncTime(ctx, ds.db, "teams")
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	if usersErr := ds.syncAllProviderTeams(ctx); usersErr != nil {
		return fmt.Errorf("teams: %w", usersErr)
	}
	log.Info().Msg("teams data sync complete")

	return nil
}

func (ds *teamDataSyncer) syncAllProviderTeams(ctx context.Context) error {
	var batch []*ent.Team

	start := time.Now()
	var numMutations int

	batchSize := 10
	for provTeam, pullErr := range ds.provider.PullTeams(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull teams: %w", pullErr)
		}
		batch = append(batch, provTeam)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncBatch(ctx, batch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			batch = make([]*ent.Team, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	ds.saveSyncHistory(ctx, start, numMutations, "teams")

	return nil
}

func (ds *teamDataSyncer) syncBatch(ctx context.Context, batch []*ent.Team) (int, error) {
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

func (ds *teamDataSyncer) createBatchSyncMutations(ctx context.Context, batch []*ent.Team) error {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbTeams, queryErr := ds.db.Team.Query().Where(team.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying users: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.Team)
	for _, tm := range dbTeams {
		t := tm
		dbProvMap[t.ProviderID] = t
	}

	for _, provTeam := range batch {
		dbTeam, exists := dbProvMap[provTeam.ProviderID]
		if exists {
			// don't delete this user
		}
		_, syncErr := ds.syncTeam(ctx, dbTeam, provTeam)
		if syncErr != nil {
			return fmt.Errorf("syncing team: %w", syncErr)
		}
	}

	return nil
}

func (ds *teamDataSyncer) syncTeam(ctx context.Context, db, prov *ent.Team) (uuid.UUID, error) {
	var m *ent.TeamMutation
	var teamId uuid.UUID
	needsSync := true
	if db == nil {
		teamId = uuid.New()
		m = ds.db.Team.Create().SetID(teamId).Mutation()
	} else {
		teamId = db.ID
		m = ds.db.Team.UpdateOneID(teamId).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Name != prov.Name || db.Timezone != prov.Timezone
	}

	slug, slugErr := ds.slugs.generateUnique(prov.Name, func(prefix string) (int, error) {
		return ds.db.Team.Query().Where(team.SlugHasPrefix(prefix)).Count(ctx)
	})
	if slugErr != nil {
		return uuid.Nil, fmt.Errorf("failed to create unique incident slug: %w", slugErr)
	}

	m.SetProviderID(prov.ProviderID)
	m.SetName(prov.Name)
	m.SetSlug(slug)
	m.SetTimezone(prov.Timezone)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return teamId, nil
}
