package datasync

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	"github.com/rezible/rezible/ent/oncallscheduleparticipant"
	"iter"
)

func syncOncallRosters(ctx context.Context, db *ent.Client, prov rez.OncallDataProvider) error {
	b := &oncallRosterBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.OncallRoster](db, "oncall_rosters", b)
	return s.Sync(ctx)
}

type oncallRosterBatcher struct {
	db       *ent.Client
	provider rez.OncallDataProvider

	deletedRosterIds      []uuid.UUID
	deletedScheduleIds    []uuid.UUID
	deletedParticipantIds []uuid.UUID
}

func newOncallRosterSyncer(db *ent.Client, prov rez.OncallDataProvider) *batchedDataSyncer[*ent.OncallRoster] {
	b := &oncallRosterBatcher{db: db, provider: prov}
	return newBatchedDataSyncer[*ent.OncallRoster](db, "oncall_rosters", b)
}

func (b *oncallRosterBatcher) setup(ctx context.Context) error {
	b.deletedRosterIds = make([]uuid.UUID, 0)
	b.deletedScheduleIds = make([]uuid.UUID, 0)
	b.deletedParticipantIds = make([]uuid.UUID, 0)
	return nil
}

func (b *oncallRosterBatcher) pullData(ctx context.Context) iter.Seq2[*ent.OncallRoster, error] {
	return b.provider.PullRosters(ctx)
}

func (b *oncallRosterBatcher) queryRosters(ctx context.Context, ids []string) ([]*ent.OncallRoster, error) {
	return b.db.OncallRoster.Query().
		Where(oncallroster.ProviderIDIn(ids...)).
		WithSchedules(func(q *ent.OncallScheduleQuery) {
			q.WithParticipants()
		}).All(ctx)
}

func makeOncallRosterSlug(name string) string {
	return slug.MakeLang(name, "en")
}

// One user can appear multiple times in a schedule participation rotation
func participantUniqueId(p *ent.OncallScheduleParticipant) string {
	return fmt.Sprintf("%s_%s_%d", p.ScheduleID.String(), p.UserID.String(), p.Index)
}

func (b *oncallRosterBatcher) createBatchMutations(ctx context.Context, batch []*ent.OncallRoster) ([]ent.Mutation, error) {
	providerIds := make([]string, len(batch))
	for i, p := range batch {
		providerIds[i] = p.ProviderID
	}

	dbRosters, rostersErr := b.queryRosters(ctx, providerIds)
	if rostersErr != nil {
		return nil, fmt.Errorf("get current rosters: %w", rostersErr)
	}

	deletedRosterIds := mapset.NewSet[uuid.UUID]()
	rosterProviderIdMap := make(map[string]*ent.OncallRoster)

	deletedScheduleIds := mapset.NewSet[uuid.UUID]()
	scheduleProviderIdMap := make(map[string]*ent.OncallSchedule)

	deletedParticipantIds := mapset.NewSet[uuid.UUID]()
	participantIdMap := make(map[string]*ent.OncallScheduleParticipant)
	for _, roster := range dbRosters {
		rosterProviderIdMap[roster.ProviderID] = roster
		deletedRosterIds.Add(roster.ID)
		for _, sched := range roster.Edges.Schedules {
			scheduleProviderIdMap[sched.ProviderID] = sched
			deletedScheduleIds.Add(sched.ID)
			for _, part := range sched.Edges.Participants {
				participantIdMap[participantUniqueId(part)] = part
				deletedParticipantIds.Add(part.ID)
			}
		}
	}

	var mutations []ent.Mutation
	for _, roster := range batch {
		provRoster := roster
		provRoster.Slug = makeOncallRosterSlug(provRoster.Name)

		currRoster, rosterExists := rosterProviderIdMap[provRoster.ProviderID]
		if rosterExists {
			deletedRosterIds.Remove(currRoster.ID)
		}
		rosterId, rosterMut := b.syncRoster(currRoster, provRoster)
		provRoster.ID = rosterId
		if rosterMut != nil {
			mutations = append(mutations, rosterMut)
		}

		for _, sched := range provRoster.Edges.Schedules {
			provSched := sched
			provSched.RosterID = provRoster.ID

			currSched, schedExists := scheduleProviderIdMap[provSched.ProviderID]
			if schedExists {
				deletedScheduleIds.Remove(currSched.ID)
			}
			schedId, schedMut := b.syncSchedule(currSched, provSched)
			provSched.ID = schedId
			if schedMut != nil {
				mutations = append(mutations, schedMut)
			}

			for _, part := range provSched.Edges.Participants {
				provPart := part
				provPart.ScheduleID = provSched.ID

				currUser, userErr := lookupProviderUser(ctx, b.db, provPart.Edges.User)
				if userErr != nil {
					return nil, fmt.Errorf("querying for provider user: %w", userErr)
				}
				provPart.UserID = currUser.ID

				currPart, exists := participantIdMap[participantUniqueId(provPart)]
				if exists {
					deletedParticipantIds.Remove(currPart.ID)
					provPart.ID = currPart.ID
				}

				pMut := b.syncScheduleParticipant(currPart, provPart)
				if pMut != nil {
					mutations = append(mutations, pMut)
				}
			}
		}
	}

	return mutations, nil
}

func (b *oncallRosterBatcher) syncRoster(curr, prov *ent.OncallRoster) (uuid.UUID, *ent.OncallRosterMutation) {
	var mut *ent.OncallRosterMutation
	var id uuid.UUID
	if curr == nil {
		id = uuid.New()
		mut = b.db.OncallRoster.Create().SetID(id).Mutation()
	} else {
		id = curr.ID
		mut = b.db.OncallRoster.UpdateOneID(id).Mutation()

		needsUpdate := curr.Name != prov.Name || curr.ChatChannelID != prov.ChatChannelID || curr.ChatHandle != prov.ChatHandle
		if !needsUpdate {
			return id, nil
		}
	}
	mut.SetProviderID(prov.ProviderID)
	mut.SetName(prov.Name)
	mut.SetSlug(prov.Slug)
	mut.SetTimezone(prov.Timezone)
	if prov.ChatChannelID != "" {
		mut.SetChatChannelID(prov.ChatChannelID)
	}
	if prov.ChatHandle != "" {
		mut.SetChatHandle(prov.ChatHandle)
	}
	return id, mut
}

func (b *oncallRosterBatcher) syncSchedule(curr, prov *ent.OncallSchedule) (uuid.UUID, *ent.OncallScheduleMutation) {
	var mut *ent.OncallScheduleMutation

	var id uuid.UUID
	if curr == nil {
		id = uuid.New()
		mut = b.db.OncallSchedule.Create().SetID(id).Mutation()
	} else {
		id = curr.ID
		mut = b.db.OncallSchedule.UpdateOneID(curr.ID).Mutation()
		needsUpdate := curr.Timezone != prov.Timezone || curr.Name != prov.Name
		if !needsUpdate {
			return id, nil
		}
	}

	mut.SetProviderID(prov.ProviderID)
	mut.SetRosterID(prov.RosterID)
	mut.SetName(prov.Name)
	if prov.Timezone != "" {
		mut.SetTimezone(prov.Timezone)
	}

	return id, mut
}

func (b *oncallRosterBatcher) syncScheduleParticipant(curr, prov *ent.OncallScheduleParticipant) *ent.OncallScheduleParticipantMutation {
	var mut *ent.OncallScheduleParticipantMutation

	if curr == nil {
		mut = b.db.OncallScheduleParticipant.Create().Mutation()
	} else {
		mut = b.db.OncallScheduleParticipant.UpdateOneID(curr.ID).Mutation()

		needsUpdate := curr.Index != prov.Index
		if !needsUpdate {
			return nil
		}
	}
	mut.SetIndex(prov.Index)
	mut.SetUserID(prov.UserID)
	mut.SetScheduleID(prov.ScheduleID)

	return mut
}

func (b *oncallRosterBatcher) getDeletionMutations() []ent.Mutation {
	var muts []ent.Mutation
	if len(b.deletedRosterIds) > 0 {
		var mut ent.OncallRosterMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(oncallroster.IDIn(b.deletedRosterIds...))
		muts = append(muts, &mut)
	}

	if len(b.deletedScheduleIds) > 0 {
		var mut ent.OncallScheduleMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(oncallschedule.IDIn(b.deletedScheduleIds...))
		muts = append(muts, &mut)
	}

	if len(b.deletedParticipantIds) > 0 {
		var mut ent.OncallScheduleParticipantMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(oncallscheduleparticipant.IDIn(b.deletedParticipantIds...))
		muts = append(muts, &mut)
	}

	return muts
}
