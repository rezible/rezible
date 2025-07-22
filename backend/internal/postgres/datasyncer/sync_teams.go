package datasyncer

import (
	"context"
	"fmt"
	"iter"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/team"
)

func syncTeams(ctx context.Context, db *ent.Client, prov rez.TeamDataProvider) error {
	b := &teamsBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.Team](db, "teams", b)
	return s.Sync(ctx)
}

type teamsBatcher struct {
	db       *ent.Client
	provider rez.TeamDataProvider

	slugs *slugTracker
}

func (b *teamsBatcher) setup(ctx context.Context) error {
	b.slugs = newSlugTracker()
	return nil
}

func (b *teamsBatcher) pullData(ctx context.Context) iter.Seq2[*ent.Team, error] {
	return b.provider.PullTeams(ctx)
}

func (b *teamsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}

func (b *teamsBatcher) createBatchMutations(ctx context.Context, batch []*ent.Team) ([]ent.Mutation, error) {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbTeams, queryErr := b.db.Team.Query().Where(team.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying users: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.Team)
	for _, tm := range dbTeams {
		t := tm
		dbProvMap[t.ProviderID] = t
	}

	var mutations []ent.Mutation
	for _, provTeam := range batch {
		dbTeam, exists := dbProvMap[provTeam.ProviderID]
		if exists {

		}
		mut, syncErr := b.syncTeam(ctx, dbTeam, provTeam)
		if syncErr != nil {
			return nil, fmt.Errorf("syncing team: %w", syncErr)
		} else if mut != nil {
			mutations = append(mutations, mut)
		}
	}

	return mutations, nil
}

func (b *teamsBatcher) makeTeamSlugCountFn(ctx context.Context) func(prefix string) (int, error) {
	return func(prefix string) (int, error) {
		return b.db.Team.Query().Where(team.SlugHasPrefix(prefix)).Count(ctx)
	}
}

func (b *teamsBatcher) syncTeam(ctx context.Context, db, prov *ent.Team) (*ent.TeamMutation, error) {
	var m *ent.TeamMutation

	if db == nil {
		m = b.db.Team.Create().Mutation()
	} else {
		m = b.db.Team.UpdateOneID(db.ID).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := db.Name != prov.Name || db.Timezone != prov.Timezone
		if !needsSync {
			return nil, nil
		}
	}

	slug, slugErr := b.slugs.generateUnique(prov.Name, b.makeTeamSlugCountFn(ctx))
	if slugErr != nil {
		return nil, fmt.Errorf("failed to create unique incident slug: %w", slugErr)
	}

	m.SetProviderID(prov.ProviderID)
	m.SetName(prov.Name)
	m.SetSlug(slug)
	m.SetTimezone(prov.Timezone)

	return m, nil
}
