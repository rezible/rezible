package providers

import (
	"context"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	"github.com/rezible/rezible/ent/oncallscheduleparticipant"
	"github.com/rezible/rezible/ent/oncallusershift"
)

type oncallDataSyncer struct {
	db       *ent.Client
	provider rez.OncallDataProvider

	mutations []ent.Mutation
}

func newOncallDataSyncer(db *ent.Client, provider rez.OncallDataProvider) *oncallDataSyncer {
	return &oncallDataSyncer{db: db, provider: provider}
}

func (ds *oncallDataSyncer) SyncProviderData(ctx context.Context) error {
	start := time.Now()

	lastRostersSync := getLastSyncTime(ctx, ds.db, "oncall_rosters")
	skipRosters := lastRostersSync.Add(time.Minute * 30).After(start)

	lastShiftsSync := getLastSyncTime(ctx, ds.db, "oncall_shifts")
	skipShifts := lastShiftsSync.Add(time.Minute * 30).After(start)

	if !skipRosters {
		if rostersErr := ds.syncAllRosters(ctx); rostersErr != nil {
			return fmt.Errorf("failed to deep sync roster data: %w", rostersErr)
		}
	}

	if !skipShifts {
		if shiftsErr := ds.syncAllOncallShifts(ctx); shiftsErr != nil {
			return fmt.Errorf("failed to sync oncall shifts: %w", shiftsErr)
		}
	}

	return nil
}

func makeOncallRosterSlug(name string) string {
	return slug.MakeLang(name, "en")
}

// One user can appear multiple times in a schedule participation rotation
func participantUniqueId(p *ent.OncallScheduleParticipant) string {
	return fmt.Sprintf("%s_%s_%d", p.ScheduleID.String(), p.UserID.String(), p.Index)
}

func (ds *oncallDataSyncer) syncAllRosters(ctx context.Context) error {
	start := time.Now()
	var numMutations int

	batchSize := 10
	var batch []*ent.OncallRoster
	for roster, pullErr := range ds.provider.PullRosters(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull: %w", pullErr)
		}
		ros := roster
		batch = append(batch, ros)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncProviderOncallRosters(ctx, batch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			batch = make([]*ent.OncallRoster, 0)
		}
	}

	batchMuts, syncErr := ds.syncProviderOncallRosters(ctx, batch)
	if syncErr != nil {
		return syncErr
	}
	numMutations += batchMuts

	if saveErr := saveSyncHistory(ctx, ds.db, start, numMutations, "oncall_rosters"); saveErr != nil {
		log.Error().Err(saveErr).Msg("failed to save oncall rosters data sync history")
	}

	return nil
}

func (ds *oncallDataSyncer) syncProviderOncallRosters(ctx context.Context, rosters []*ent.OncallRoster) (int, error) {
	ds.mutations = make([]ent.Mutation, 0)
	if syncErr := ds.buildOncallRostersSyncMutations(ctx, rosters); syncErr != nil {
		return 0, fmt.Errorf("syncing rosters: %w", syncErr)
	}

	if applyErr := applySyncMutations(ctx, ds.db, ds.mutations); applyErr != nil {
		return 0, fmt.Errorf("error applying sync mutations: %w", applyErr)
	}
	numMutations := len(ds.mutations)
	ds.mutations = nil

	return numMutations, nil
}

func (ds *oncallDataSyncer) buildOncallRostersSyncMutations(ctx context.Context, provRosters []*ent.OncallRoster) error {
	if len(provRosters) == 0 {
		return nil
	}

	providerIds := make([]string, len(provRosters))
	for i, p := range provRosters {
		providerIds[i] = p.ProviderID
	}

	dbRosters, rostersErr := ds.db.OncallRoster.Query().
		Where(oncallroster.ProviderIDIn(providerIds...)).
		WithSchedules(func(q *ent.OncallScheduleQuery) {
			q.WithParticipants()
		}).All(ctx)
	if rostersErr != nil {
		return fmt.Errorf("get current rosters: %w", rostersErr)
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

	for _, roster := range provRosters {
		provRoster := roster
		provRoster.Slug = makeOncallRosterSlug(provRoster.Name)

		currRoster, rosterExists := rosterProviderIdMap[provRoster.ProviderID]
		if rosterExists {
			deletedRosterIds.Remove(currRoster.ID)
			provRoster.ID = currRoster.ID
		}
		provRoster.ID = ds.syncRoster(currRoster, provRoster)

		for _, sched := range provRoster.Edges.Schedules {
			provSched := sched
			provSched.RosterID = provRoster.ID

			currSched, schedExists := scheduleProviderIdMap[provSched.ProviderID]
			if schedExists {
				deletedScheduleIds.Remove(currSched.ID)
				provSched.ID = currSched.ID
			}
			provSched.ID = ds.syncSchedule(currSched, provSched)

			for _, part := range sched.Edges.Participants {
				provPart := part
				provPart.ScheduleID = provSched.ID

				provUser := provPart.Edges.User
				currUser, userErr := lookupProviderUser(ctx, ds.db, provUser)
				if userErr != nil && !ent.IsNotFound(userErr) {
					return fmt.Errorf("querying for existing schedule user: %w", userErr)
				}
				provPart.UserID = ds.syncOncallUser(currUser, provUser)

				uid := participantUniqueId(provPart)
				currPart, exists := participantIdMap[uid]
				if exists {
					deletedParticipantIds.Remove(currPart.ID)
					provPart.ID = currPart.ID
				}

				ds.syncScheduleParticipant(currPart, provPart)
			}
		}
	}

	if !deletedRosterIds.IsEmpty() {
		var mut ent.OncallRosterMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(oncallroster.IDIn(deletedRosterIds.ToSlice()...))
		ds.mutations = append(ds.mutations, &mut)
	}

	if !deletedScheduleIds.IsEmpty() {
		var mut ent.OncallScheduleMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(oncallschedule.IDIn(deletedScheduleIds.ToSlice()...))
		ds.mutations = append(ds.mutations, &mut)
	}

	if !deletedParticipantIds.IsEmpty() {
		var mut ent.OncallScheduleParticipantMutation
		mut.SetOp(ent.OpDelete)
		mut.Where(oncallscheduleparticipant.IDIn(deletedParticipantIds.ToSlice()...))
		ds.mutations = append(ds.mutations, &mut)
	}

	return nil
}

func (ds *oncallDataSyncer) syncRoster(curr *ent.OncallRoster, prov *ent.OncallRoster) uuid.UUID {
	var mut *ent.OncallRosterMutation
	var rosterId uuid.UUID
	needsUpdate := true
	if curr == nil {
		rosterId = uuid.New()
		mut = ds.db.OncallRoster.Create().SetID(rosterId).Mutation()
	} else {
		rosterId = curr.ID
		needsUpdate = curr.Name != prov.Name || curr.ChatChannelID != prov.ChatChannelID || curr.ChatHandle != prov.ChatHandle
		mut = ds.db.OncallRoster.UpdateOneID(rosterId).Mutation()
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
	if needsUpdate {
		ds.mutations = append(ds.mutations, mut)
	}
	return rosterId
}

func (ds *oncallDataSyncer) syncSchedule(curr *ent.OncallSchedule, prov *ent.OncallSchedule) uuid.UUID {
	var mut *ent.OncallScheduleMutation

	var scheduleId uuid.UUID
	needsUpdate := true
	if curr == nil {
		scheduleId = uuid.New()
		mut = ds.db.OncallSchedule.Create().SetID(scheduleId).Mutation()
	} else {
		scheduleId = curr.ID
		mut = ds.db.OncallSchedule.UpdateOneID(scheduleId).Mutation()
		needsUpdate = curr.Timezone != prov.Timezone || curr.Name != prov.Name
	}

	mut.SetProviderID(prov.ProviderID)
	mut.SetRosterID(prov.RosterID)
	mut.SetName(prov.Name)
	if prov.Timezone != "" {
		mut.SetTimezone(prov.Timezone)
	}

	if needsUpdate {
		ds.mutations = append(ds.mutations, mut)
	}

	return scheduleId
}

func (ds *oncallDataSyncer) syncScheduleParticipant(
	curr *ent.OncallScheduleParticipant, prov *ent.OncallScheduleParticipant,
) uuid.UUID {
	var mut *ent.OncallScheduleParticipantMutation

	var partId uuid.UUID
	needsUpdate := true
	if curr == nil {
		partId = uuid.New()
		mut = ds.db.OncallScheduleParticipant.Create().SetID(partId).Mutation()
	} else {
		partId = curr.ID
		mut = ds.db.OncallScheduleParticipant.UpdateOneID(partId).Mutation()
		needsUpdate = curr.Index != prov.Index
	}
	mut.SetIndex(prov.Index)
	mut.SetUserID(prov.UserID)
	mut.SetScheduleID(prov.ScheduleID)

	if needsUpdate {
		ds.mutations = append(ds.mutations, mut)
	}

	return partId
}

func (ds *oncallDataSyncer) syncOncallUser(curr *ent.User, prov *ent.User) uuid.UUID {
	var mut *ent.UserMutation
	// TODO: check if we should sync

	var userId uuid.UUID
	needsUpdate := true
	if curr == nil {
		userId = uuid.New()
		mut = ds.db.User.Create().SetID(userId).Mutation()
		// TODO: add to lookup struct so we don't create twice?
	} else {
		userId = curr.ID
		mut = ds.db.User.UpdateOneID(userId).Mutation()

		// needsUpdate = curr.ID != userId && curr.Name != prov.Name
		needsUpdate = false
	}
	mut.SetName(prov.Name)
	mut.SetEmail(prov.Email)

	if needsUpdate {
		ds.mutations = append(ds.mutations, mut)
	}

	return userId
}

func (ds *oncallDataSyncer) syncAllOncallShifts(ctx context.Context) error {
	start := time.Now()
	rosters, rostersErr := ds.db.OncallRoster.Query().All(ctx)
	if rostersErr != nil {
		return fmt.Errorf("failed to query rosters: %w", rostersErr)
	}

	var numMutations int
	for _, r := range rosters {
		roster := r

		// TODO: check last sync time for roster

		ds.mutations = make([]ent.Mutation, 0)
		if syncErr := ds.buildRosterShiftsSyncMutations(ctx, roster); syncErr != nil {
			return fmt.Errorf("failed to build roster shifts sync mutations: %w", syncErr)
		}
		numMutations += len(ds.mutations)

		if applyErr := applySyncMutations(ctx, ds.db, ds.mutations); applyErr != nil {
			return fmt.Errorf("failed to apply shifts sync mutations: %w", applyErr)
		}
		ds.mutations = nil
	}

	if saveErr := saveSyncHistory(ctx, ds.db, start, numMutations, "oncall_shifts"); saveErr != nil {
		log.Error().Err(saveErr).Msg("failed to save oncall shifts data sync history")
	}

	return nil
}

func (ds *oncallDataSyncer) buildRosterShiftsSyncMutations(ctx context.Context, roster *ent.OncallRoster) error {
	rosterTz, tzErr := time.LoadLocation("UTC")
	if tzErr != nil {
		return fmt.Errorf("loading roster timezone: %w", tzErr)
	}

	formatTime := func(t time.Time) string {
		timeFmt := "2006-01-02T15:04"
		return t.In(rosterTz).Format(timeFmt)
	}

	anchor := time.Now().In(rosterTz)
	// TODO: load this from schedule
	shiftDuration := time.Hour * 24 * 7
	syncWindow := shiftDuration * 2

	shiftsFrom := anchor.Add(-syncWindow)
	shiftsTo := anchor.Add(syncWindow)

	formattedFrom := formatTime(shiftsFrom)
	formattedTo := formatTime(shiftsTo)
	isIncompleteShift := func(from, to time.Time) bool {
		return formatTime(from) == formattedFrom || formatTime(to) == formattedTo
	}

	shiftsBetween := oncallusershift.And(oncallusershift.StartAtGTE(shiftsFrom), oncallusershift.EndAtLTE(shiftsTo))
	dbShifts, queryErr := ds.db.OncallUserShift.Query().
		Where(shiftsBetween).
		Where(oncallusershift.RosterID(roster.ID)).
		All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying roster shifts: %w", queryErr)
	}

	makeOncallShiftKey := func(sh *ent.OncallUserShift) string {
		return fmt.Sprintf("%s_%s:%s-%s", sh.RosterID, sh.UserID, formatTime(sh.StartAt), formatTime(sh.EndAt))
	}

	currShifts := make(map[string]*ent.OncallUserShift)
	for _, sh := range dbShifts {
		currShifts[makeOncallShiftKey(sh)] = sh
	}

	for provShift, pullErr := range ds.provider.PullShiftsForRoster(ctx, roster.ProviderID, shiftsFrom, shiftsTo) {
		if pullErr != nil {
			return fmt.Errorf("pulling provider shifts: %w", pullErr)
		}

		shift := provShift
		if isIncompleteShift(shift.StartAt, shift.EndAt) {
			// log.Debug().Msg("incomplete shift, skipping")
			continue
		}
		shift.RosterID = roster.ID

		usr, usrErr := lookupProviderUser(ctx, ds.db, shift.Edges.User)
		if usrErr != nil {
			log.Error().Err(usrErr).Str("email", shift.Edges.User.Email).Msg("failed to get user")
			continue
		}
		shift.UserID = usr.ID

		key := makeOncallShiftKey(shift)
		_, exists := currShifts[key]
		if exists {
			delete(currShifts, key)
			continue
		}

		shiftId := uuid.New()
		create := ds.db.OncallUserShift.Create().
			SetID(shiftId).
			SetProviderID(shift.ProviderID).
			SetUserID(usr.ID).
			SetRosterID(roster.ID).
			SetStartAt(shift.StartAt).
			SetEndAt(shift.EndAt)
		ds.mutations = append(ds.mutations, create.Mutation())
	}

	for _, sh := range currShifts {
		// don't delete shifts before anchor
		if sh.StartAt.Before(anchor) {
			continue
		}
		log.Debug().Str("id", sh.ID.String()).Msg("delete shift")
		// deleteShiftIds = append(deleteShiftIds, sh.ID)
	}

	//log.Debug().
	//	Str("roster", roster.Name).
	//	Int("mutations", len(ds.mutations)).
	//	Time("from", shiftsFrom).
	//	Time("to", shiftsTo).
	//	Msg("oncall shift sync")

	return nil
}
