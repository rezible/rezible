package datasync

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallusershift"
)

func syncOncallShifts(ctx context.Context, db *ent.Client, prov rez.OncallDataProvider) error {
	b := &oncallShiftsBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.OncallUserShift](db, "oncall_shifts", b)
	return s.Sync(ctx)
}

type oncallShiftBatchParams struct {
	rosterID uuid.UUID
	from     time.Time
	to       time.Time
}

type oncallShiftsBatcher struct {
	db       *ent.Client
	provider rez.OncallDataProvider

	batchParams *oncallShiftBatchParams
}

func (b *oncallShiftsBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *oncallShiftsBatcher) pullData(ctx context.Context) iter.Seq2[*ent.OncallUserShift, error] {
	rosters, queryErr := b.db.OncallRoster.Query().All(ctx)

	return func(yield func(*ent.OncallUserShift, error) bool) {
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
			for shift, pullErr := range b.provider.PullShiftsForRoster(ctx, r.ProviderID, from, to) {
				if !yield(shift, pullErr) {
					return
				}
			}
		}
	}
}

func (b *oncallShiftsBatcher) createBatchMutations(ctx context.Context, batch []*ent.OncallUserShift) ([]ent.Mutation, error) {
	provIds := make([]string, len(batch))
	for i, s := range batch {
		provIds[i] = s.ProviderID
	}

	dbShifts, queryErr := b.db.OncallUserShift.Query().
		Where(oncallusershift.ProviderIDIn(provIds...)).
		Where(oncallusershift.RosterID(b.batchParams.rosterID)).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying roster shifts: %w", queryErr)
	}

	dbProvIds := make(map[string]*ent.OncallUserShift)
	for _, sh := range dbShifts {
		dbProvIds[sh.ProviderID] = sh
	}

	var mutations []ent.Mutation
	for _, provShift := range batch {
		prov := provShift
		prov.RosterID = b.batchParams.rosterID

		usr, usrErr := lookupProviderUser(ctx, b.db, prov.Edges.User)
		if usrErr != nil {
			log.Error().Err(usrErr).Str("email", prov.Edges.User.Email).Msg("failed to get user")
			continue
		}
		prov.UserID = usr.ID

		curr, exists := dbProvIds[prov.ProviderID]
		if exists {

		}

		m := b.syncShift(curr, prov)
		if m != nil {
			mutations = append(mutations, m)
		}
	}

	return mutations, nil
}

func (b *oncallShiftsBatcher) syncShift(curr, prov *ent.OncallUserShift) *ent.OncallUserShiftMutation {
	var m *ent.OncallUserShiftMutation

	if curr == nil {
		m = b.db.OncallUserShift.Create().Mutation()
	} else {
		m = b.db.OncallUserShift.UpdateOneID(curr.ID).Mutation()

		needsUpdate := !curr.StartAt.Equal(prov.StartAt) || !curr.EndAt.Equal(prov.EndAt)
		if !needsUpdate {
			return nil
		}
	}

	m.SetProviderID(prov.ProviderID)
	m.SetUserID(prov.UserID)
	m.SetRosterID(prov.RosterID)
	m.SetStartAt(prov.StartAt)
	m.SetEndAt(prov.EndAt)

	return m
}

func (b *oncallShiftsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}
