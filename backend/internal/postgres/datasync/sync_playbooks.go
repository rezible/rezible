package datasync

import (
	"context"
	"fmt"
	"iter"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/playbook"
)

func syncPlaybooks(ctx context.Context, db *ent.Client, prov rez.PlaybookDataProvider) error {
	b := &playbooksBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.Playbook](db, "playbooks", b)
	return s.Sync(ctx)
}

type playbooksBatcher struct {
	db       *ent.Client
	provider rez.PlaybookDataProvider
}

func (b *playbooksBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *playbooksBatcher) pullData(ctx context.Context) iter.Seq2[*ent.Playbook, error] {
	return b.provider.PullPlaybooks(ctx)
}

func (b *playbooksBatcher) createBatchMutations(ctx context.Context, batch []*ent.Playbook) ([]ent.Mutation, error) {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbPlaybooks, queryErr := b.db.Playbook.Query().Where(playbook.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying playbooks: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.Playbook)
	for _, pb := range dbPlaybooks {
		p := pb
		dbProvMap[p.ProviderID] = p
	}

	var muts []ent.Mutation
	for _, prov := range batch {
		curr, exists := dbProvMap[prov.ProviderID]
		if exists {
		}
		if mut := b.syncPlaybook(curr, prov); mut != nil {
			muts = append(muts, mut)
		}
	}

	return muts, nil
}

func (b *playbooksBatcher) syncPlaybook(curr, prov *ent.Playbook) *ent.PlaybookMutation {
	var m *ent.PlaybookMutation

	if curr == nil {
		m = b.db.Playbook.Create().Mutation()
	} else {
		m = b.db.Playbook.UpdateOneID(curr.ID).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := curr.Title != prov.Title
		if !needsSync {
			return nil
		}
	}

	m.SetProviderID(prov.ProviderID)
	m.SetTitle(prov.Title)
	m.SetContent(prov.Content)

	return m
}

func (b *playbooksBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}
