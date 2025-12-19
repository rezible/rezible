package datasync

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallshift"
)

func syncOncallShifts(ctx context.Context, db *ent.Client, prov rez.OncallDataProvider) error {
	b := &oncallShiftsBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.OncallShift](db, "oncall_shifts", b)
	return s.Sync(ctx)
}

type oncallShiftBatchParams struct {
	rosterID uuid.UUID
	from     time.Time
	to       time.Time
}

type oncallShiftsBatcher struct {
	db          *ent.Client
	provider    rez.OncallDataProvider
	userTracker *providerUserTracker

	batchParams *oncallShiftBatchParams
}

func (b *oncallShiftsBatcher) setup(ctx context.Context) error {
	b.userTracker = newProviderUserTracker(b.db.User)
	return nil
}

func (b *oncallShiftsBatcher) pullData(ctx context.Context) iter.Seq2[*ent.OncallShift, error] {
	rosters, queryErr := b.db.OncallRoster.Query().All(ctx)

	return func(yield func(*ent.OncallShift, error) bool) {
		if queryErr != nil {
			yield(nil, queryErr)
			return
		}
		for _, r := range rosters {
			// TODO: load this from roster/schedule
			shiftDuration := time.Hour * 24 * 7
			syncWindow := shiftDuration * 2

			rosterTz, tzErr := time.LoadLocation("UTC")
			if tzErr != nil {
				yield(nil, fmt.Errorf("loading roster timezone: %w", tzErr))
				return
			}
			rosterNow := time.Now().In(rosterTz)

			from := rosterNow.Add(-syncWindow)
			to := rosterNow.Add(syncWindow)

			b.batchParams = &oncallShiftBatchParams{
				rosterID: r.ID,
				from:     from,
				to:       to,
			}
			for shift, pullErr := range b.provider.PullShiftsForRoster(ctx, r.ExternalID, from, to) {
				if !yield(shift, pullErr) {
					return
				}
			}
		}
	}
}

func (b *oncallShiftsBatcher) createBatchMutations(ctx context.Context, batch []*ent.OncallShift) ([]ent.Mutation, error) {
	provIds := make([]string, len(batch))
	for i, s := range batch {
		provIds[i] = s.ExternalID
	}

	dbShifts, queryErr := b.db.OncallShift.Query().
		Where(oncallshift.ExternalIDIn(provIds...)).
		Where(oncallshift.RosterID(b.batchParams.rosterID)).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying roster shifts: %w", queryErr)
	}

	dbProvIds := make(map[string]*ent.OncallShift)
	for _, sh := range dbShifts {
		dbProvIds[sh.ExternalID] = sh
	}

	//provUserMapping := b.provider.UserShiftDataMapping().Edges.User
	var provUserMapping *ent.User

	var mutations []ent.Mutation
	for _, provShift := range batch {
		prov := provShift
		prov.RosterID = b.batchParams.rosterID

		userId, userMut, userErr := b.userTracker.lookupOrCreate(ctx, prov.Edges.User, provUserMapping)
		if userErr != nil {
			return nil, fmt.Errorf("querying for provider user: %w", userErr)
		}
		if userMut != nil {
			mutations = append(mutations, userMut)
		}
		prov.UserID = userId

		curr, exists := dbProvIds[prov.ExternalID]
		if exists {

		}

		m := b.syncShift(curr, prov)
		if m != nil {
			mutations = append(mutations, m)
		}
	}

	return mutations, nil
}

func (b *oncallShiftsBatcher) syncShift(curr, prov *ent.OncallShift) *ent.OncallShiftMutation {
	var m *ent.OncallShiftMutation

	if curr == nil {
		m = b.db.OncallShift.Create().Mutation()
	} else {
		m = b.db.OncallShift.UpdateOneID(curr.ID).Mutation()

		needsUpdate := !curr.StartAt.Equal(prov.StartAt) || !curr.EndAt.Equal(prov.EndAt)
		if !needsUpdate {
			return nil
		}
	}

	m.SetExternalID(prov.ExternalID)
	m.SetUserID(prov.UserID)
	m.SetRosterID(prov.RosterID)
	m.SetStartAt(prov.StartAt)
	m.SetEndAt(prov.EndAt)

	return m
}

func (b *oncallShiftsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}
