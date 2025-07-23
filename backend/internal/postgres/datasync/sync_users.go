package datasync

import (
	"context"
	"fmt"
	"iter"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
)

func syncUsers(ctx context.Context, db *ent.Client, prov rez.UserDataProvider) error {
	b := &usersBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.User](db, "users", b)
	return s.Sync(ctx)
}

type usersBatcher struct {
	db       *ent.Client
	provider rez.UserDataProvider
}

func (b *usersBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *usersBatcher) pullData(ctx context.Context) iter.Seq2[*ent.User, error] {
	return b.provider.PullUsers(ctx)
}

func (b *usersBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}

func (b *usersBatcher) createBatchMutations(ctx context.Context, batch []*ent.User) ([]ent.Mutation, error) {
	emails := make([]string, len(batch))
	for i, u := range batch {
		emails[i] = u.Email
	}

	dbUsers, queryErr := b.db.User.Query().Where(user.EmailIn(emails...)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying users: %w", queryErr)
	}
	dbEmailMap := make(map[string]*ent.User)
	for _, usr := range dbUsers {
		u := usr
		dbEmailMap[u.Email] = usr
	}

	var mutations []ent.Mutation
	for _, provUser := range batch {
		dbUser, exists := dbEmailMap[provUser.Email]
		if !exists {
		}
		userMut, syncErr := b.syncUser(dbUser, provUser)
		if syncErr != nil {
			return nil, fmt.Errorf("syncing user: %w", syncErr)
		} else if userMut != nil {
			mutations = append(mutations, userMut)
		}
	}

	return mutations, nil
}

func (b *usersBatcher) syncUser(db, prov *ent.User) (*ent.UserMutation, error) {
	var m *ent.UserMutation
	var userId uuid.UUID
	if db == nil {
		m = b.db.User.Create().Mutation()
	} else {
		userId = db.ID
		m = b.db.User.UpdateOneID(userId).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := db.Name != prov.Name || db.Email != prov.Email || db.Timezone != prov.Timezone || db.ChatID != prov.ChatID
		if !needsSync {
			return nil, nil
		}
	}

	m.SetName(prov.Name)
	m.SetEmail(prov.Email)
	m.SetChatID(prov.ChatID)
	m.SetTimezone(prov.Timezone)

	return m, nil
}
