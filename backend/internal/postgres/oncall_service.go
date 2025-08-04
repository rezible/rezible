package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/texm/prosemirror-go"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallhandovertemplate"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	"github.com/rezible/rezible/ent/oncallscheduleparticipant"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rezible/rezible/ent/oncallusershifthandover"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/jobs"
)

type OncallService struct {
	db        *ent.Client
	jobs      rez.JobsService
	docs      rez.DocumentsService
	chat      rez.ChatService
	users     rez.UserService
	incidents rez.IncidentService
}

func NewOncallService(ctx context.Context, db *ent.Client, jobs rez.JobsService, docs rez.DocumentsService, chat rez.ChatService, users rez.UserService, incidents rez.IncidentService) (*OncallService, error) {
	s := &OncallService{
		db:        db,
		jobs:      jobs,
		docs:      docs,
		chat:      chat,
		users:     users,
		incidents: incidents,
	}

	//go s.registerHandoverSchema()

	return s, nil
}

func (s *OncallService) registerHandoverSchema() {
	if s.docs == nil {
		log.Warn().Msg("no docs service for oncall service, not registering schema")
		return
	}
	ctx := context.Background()
	spec, specErr := s.docs.GetDocumentSchemaSpec(ctx, "handover")
	if specErr != nil || spec == nil {
		log.Error().Err(specErr).Msg("Failed to get handover schema spec")
		return
	}
	schema, schemaErr := prosemirror.NewSchema(*spec)
	if schemaErr != nil {
		log.Error().Err(schemaErr).Msg("Failed to create handover schema")
		return
	}
	prosemirror.RegisterSchema(schema)
}

func (s *OncallService) ListSchedules(ctx context.Context, params rez.ListOncallSchedulesParams) ([]*ent.OncallSchedule, error) {
	var query *ent.OncallScheduleQuery
	if params.UserID != uuid.Nil {
		query = s.db.OncallScheduleParticipant.Query().
			Where(oncallscheduleparticipant.UserID(params.UserID)).
			QuerySchedule()
	} else {
		query = s.db.OncallSchedule.Query()
	}
	query = query.Limit(params.GetLimit()).Offset(params.Offset)

	userSchedules, schedulesErr := query.All(params.GetQueryContext(ctx))
	if schedulesErr != nil {
		return nil, fmt.Errorf("failed to query oncall schedules: %w", schedulesErr)
	}
	return userSchedules, nil
}

func (s *OncallService) GetRosterByID(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error) {
	roster, rosterErr := s.db.OncallRoster.Get(ctx, id)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to query roster: %w", rosterErr)
	}
	return roster, nil
}

func (s *OncallService) GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error) {
	return s.db.OncallSchedule.Query().
		Where(oncallschedule.ID(scheduleId)).
		QueryRoster().
		Only(ctx)
}

func (s *OncallService) GetRosterBySlug(ctx context.Context, slug string) (*ent.OncallRoster, error) {
	query := s.db.OncallRoster.Query().
		Where(oncallroster.Slug(slug))

	roster, rosterErr := query.Only(ctx)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to query roster: %w", rosterErr)
	}
	return roster, nil
}

func (s *OncallService) ListRosters(ctx context.Context, params rez.ListOncallRostersParams) ([]*ent.OncallRoster, error) {
	query := s.db.OncallRoster.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset)

	if params.UserID != uuid.Nil {
		schedules, schedulesErr := s.ListSchedules(ctx, rez.ListOncallSchedulesParams{
			UserID: params.UserID,
		})
		if schedulesErr != nil {
			return nil, fmt.Errorf("failed to list oncall schedules: %w", schedulesErr)
		}
		var rosterIds []uuid.UUID
		for _, schedule := range schedules {
			sch := schedule
			rosterIds = append(rosterIds, sch.RosterID)
		}
		query = query.Where(oncallroster.IDIn(rosterIds...))
	}

	rosters, queryErr := query.All(params.GetQueryContext(ctx))
	if queryErr != nil {
		return nil, fmt.Errorf("failed to query rosters: %w", queryErr)
	}

	return rosters, queryErr
}

func (s *OncallService) GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, error) {
	query := s.db.OncallUserShift.Query().
		Where(oncallusershift.ID(id)).
		WithRoster().
		WithUser()
	return query.Only(ctx)
}

func (s *OncallService) getNextShift(ctx context.Context, shift *ent.OncallUserShift) (*ent.OncallUserShift, error) {
	return s.db.OncallUserShift.Query().
		Where(oncallusershift.RosterID(shift.RosterID)).
		Where(oncallusershift.IDNEQ(shift.ID)).
		Where(oncallusershift.StartAtGTE(shift.StartAt)).
		Order(oncallusershift.ByStartAt(sql.OrderAsc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallService) getPreviousShift(ctx context.Context, shift *ent.OncallUserShift) (*ent.OncallUserShift, error) {
	return s.db.OncallUserShift.Query().
		Where(oncallusershift.RosterID(shift.RosterID)).
		Where(oncallusershift.IDNEQ(shift.ID)).
		Where(oncallusershift.EndAtLTE(shift.StartAt)).
		Order(oncallusershift.ByEndAt(sql.OrderDesc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallService) GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, *ent.OncallUserShift, error) {
	shift, shiftErr := s.db.OncallUserShift.Get(ctx, id)
	if shiftErr != nil {
		return nil, nil, fmt.Errorf("lookup shift: %w", shiftErr)
	}

	nextShift, nextShiftErr := s.getNextShift(ctx, shift)
	if nextShiftErr != nil && !ent.IsNotFound(nextShiftErr) {
		return nil, nil, fmt.Errorf("next shift: %w", nextShiftErr)
	}

	prevShift, prevShiftErr := s.getPreviousShift(ctx, shift)
	if prevShiftErr != nil && !ent.IsNotFound(prevShiftErr) {
		return nil, nil, fmt.Errorf("previous shift: %w", prevShiftErr)
	}

	return prevShift, nextShift, nil
}

func (s *OncallService) ListShifts(ctx context.Context, params rez.ListOncallShiftsParams) ([]*ent.OncallUserShift, error) {
	query := s.db.OncallUserShift.Query().
		Order(oncallusershift.ByEndAt(sql.OrderDesc())).
		Limit(params.GetLimit()).
		Offset(params.Offset).
		WithRoster().
		WithUser()

	var predicates []predicate.OncallUserShift
	if !params.Anchor.IsZero() && !(params.Window.Milliseconds() == 0) {
		from := params.Anchor.Add(-params.Window)
		to := params.Anchor.Add(params.Window)
		withinWindow := oncallusershift.And(oncallusershift.EndAtGTE(from), oncallusershift.StartAtLTE(to))
		predicates = append(predicates, withinWindow)
	}
	if params.UserID != uuid.Nil {
		predicates = append(predicates, oncallusershift.UserID(params.UserID))
	}
	if len(predicates) > 0 {
		query = query.Where(predicates...)
	}

	return query.All(params.GetQueryContext(ctx))
}

func (s *OncallService) queryShiftsEndingWithinWindow(ctx context.Context, window time.Duration) ([]*ent.OncallUserShift, error) {
	shiftEndingWithinWindow := oncallusershift.And(
		oncallusershift.EndAtGT(time.Now().Add(-window)),
		oncallusershift.EndAtLT(time.Now().Add(window)))

	query := s.db.OncallUserShift.Query().
		Where(shiftEndingWithinWindow).
		WithHandover()

	shifts, shiftsErr := query.All(ctx)
	if shiftsErr != nil {
		return nil, fmt.Errorf("failed to get shifts: %w", shiftsErr)
	}

	return shifts, nil
}

func (s *OncallService) MakeScanShiftsPeriodicJob(ctx context.Context) (*jobs.PeriodicJob, error) {
	// TODO: check shift schedule duration, make interval less than it
	interval := time.Hour
	job := &jobs.PeriodicJob{
		ConstructorFunc: func() jobs.InsertJobParams {
			return jobs.InsertJobParams{Args: &jobs.ScanOncallShifts{}}
		},
		Interval: interval,
		Opts: &jobs.PeriodicJobOpts{
			RunOnStart: true,
		},
	}
	return job, nil
}

func (s *OncallService) HandlePeriodicScanShifts(ctx context.Context, _ jobs.ScanOncallShifts) error {
	// check for shifts ending within window, that don't have reminder sent
	shifts, shiftsErr := s.queryShiftsEndingWithinWindow(ctx, time.Hour)
	if shiftsErr != nil {
		return fmt.Errorf("failed to get shifts: %w", shiftsErr)
	}

	reminderWindow := time.Minute * 10

	var params []jobs.InsertJobParams
	for _, shift := range shifts {
		ho := shift.Edges.Handover
		reminderSent := ho != nil && ho.ReminderSent // !ho.ReminderSentAt.IsZero
		if !reminderSent {
			params = append(params, jobs.InsertJobParams{
				Args:        jobs.EnsureShiftHandoverReminderSent{ShiftId: shift.ID},
				ScheduledAt: shift.EndAt.Add(-reminderWindow),
				Uniqueness: &jobs.JobUniquenessOpts{
					Args: true,
				},
			})
		}
		isSent := ho != nil && !ho.SentAt.IsZero()
		if !isSent {
			params = append(params, jobs.InsertJobParams{
				Args:        jobs.EnsureShiftHandoverSent{ShiftId: shift.ID},
				ScheduledAt: shift.EndAt,
				Uniqueness: &jobs.JobUniquenessOpts{
					Args: true,
				},
			})
		}
		params = append(params, jobs.InsertJobParams{
			Args:        jobs.GenerateShiftMetrics{ShiftId: shift.ID},
			ScheduledAt: shift.EndAt,
			Uniqueness: &jobs.JobUniquenessOpts{
				Args: true,
			},
		})
	}

	if len(params) > 0 {
		if insertErr := s.jobs.InsertMany(ctx, params); insertErr != nil {
			return fmt.Errorf("could not insert jobs: %w", insertErr)
		}
	}

	return nil
}

func (s *OncallService) HandleEnsureShiftHandoverSent(ctx context.Context, args jobs.EnsureShiftHandoverSent) error {
	shiftId := args.ShiftId

	ho, hoErr := s.GetHandoverForShift(ctx, shiftId, true)
	if hoErr != nil {
		return fmt.Errorf("failed to get or create shift handover: %w", hoErr)
	}
	_, sendErr := s.sendShiftHandover(ctx, ho)

	return sendErr
}

func (s *OncallService) HandleEnsureShiftHandoverReminderSent(ctx context.Context, args jobs.EnsureShiftHandoverReminderSent) error {
	shiftId := args.ShiftId

	shift, shiftErr := s.GetShiftByID(ctx, shiftId)
	if shiftErr != nil {
		return fmt.Errorf("querying shift: %w", shiftErr)
	}

	ho, hoErr := s.GetHandoverForShift(ctx, shiftId, true)
	if hoErr != nil {
		return fmt.Errorf("failed to get or create shift handover: %w", hoErr)
	}

	if ho.ReminderSent {
		return nil
	}

	if msgErr := s.chat.SendOncallHandoverReminder(ctx, shift); msgErr != nil {
		return fmt.Errorf("sending reminder: %w", msgErr)
	}

	update := ho.Update().SetReminderSent(true)
	if updateErr := update.Exec(ctx); updateErr != nil {
		return fmt.Errorf("failed to set reminder_sent: %w", updateErr)
	}

	return nil
}

var defaultHandoverTemplate = []byte(`[
{"kind":"regular","header":"Overview"},
{"kind":"regular","header":"Handoff Tasks"},
{"kind":"regular","header":"Things to Monitor"},
{"kind":"annotations","header":"Pinned Annotations"}
]`)

func (s *OncallService) getRosterHandoverTemplateContents(ctx context.Context, rosterId uuid.UUID) ([]byte, error) {
	isRosterOrDefault := oncallhandovertemplate.Or(
		oncallhandovertemplate.HasRosterWith(oncallroster.ID(rosterId)),
		oncallhandovertemplate.IsDefault(true))
	tmpl, tmplErr := s.db.OncallHandoverTemplate.Query().
		Where(isRosterOrDefault).
		Order(oncallhandovertemplate.ByUpdatedAt()).
		First(ctx)
	if tmplErr != nil {
		if ent.IsNotFound(tmplErr) {
			return defaultHandoverTemplate, nil
		}
		return nil, fmt.Errorf("failed to get roster handover template: %w", tmplErr)
	}
	return tmpl.Contents, nil
}

func (s *OncallService) GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallUserShiftHandover, error) {
	return s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ID(id)).
		WithPinnedAnnotations().
		Only(ctx)
}

func (s *OncallService) GetHandoverForShift(ctx context.Context, shiftId uuid.UUID, create bool) (*ent.OncallUserShiftHandover, error) {
	handover, queryErr := s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ShiftID(shiftId)).
		WithPinnedAnnotations().
		Only(ctx)
	if queryErr != nil {
		if !create || !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("failed to query shift handover: %w", queryErr)
		}
	}
	if handover != nil {
		return handover, nil
	}

	shift, shiftErr := s.db.OncallUserShift.Get(ctx, shiftId)
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	contents, contentsErr := s.getRosterHandoverTemplateContents(ctx, shift.RosterID)
	if contentsErr != nil {
		return nil, fmt.Errorf("failed to get roster handover template contents: %w", contentsErr)
	}

	return s.createShiftHandover(ctx, shift.ID, contents)
}

func (s *OncallService) createShiftHandover(ctx context.Context, shiftId uuid.UUID, contents []byte) (*ent.OncallUserShiftHandover, error) {
	return s.db.OncallUserShiftHandover.Create().
		SetShiftID(shiftId).
		SetContents(contents).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
}

func (s *OncallService) UpdateShiftHandover(ctx context.Context, update *ent.OncallUserShiftHandover) (*ent.OncallUserShiftHandover, error) {
	curr, getErr := s.GetShiftHandover(ctx, update.ID)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get handover: %w", getErr)
	}
	query := curr.Update()
	if update.Contents != nil {
		query.SetContents(update.Contents)
	}
	if update.Edges.PinnedAnnotations != nil {
		currIds := mapset.NewSet[uuid.UUID]()
		for _, a := range curr.Edges.PinnedAnnotations {
			currIds.Add(a.ID)
		}
		updatedIds := mapset.NewSet[uuid.UUID]()
		for _, a := range update.Edges.PinnedAnnotations {
			updatedIds.Add(a.ID)
		}
		if addIds := updatedIds.Difference(currIds); addIds.Cardinality() > 0 {
			query.AddPinnedAnnotationIDs(addIds.ToSlice()...)
		}
		if deleteIds := currIds.Difference(updatedIds); deleteIds.Cardinality() > 0 {
			query.RemovePinnedAnnotationIDs(deleteIds.ToSlice()...)
		}
	}
	return query.Save(ctx)
}

func (s *OncallService) SendShiftHandover(ctx context.Context, handoverId uuid.UUID) (*ent.OncallUserShiftHandover, error) {
	query := s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ID(handoverId)).
		WithShift().
		WithPinnedAnnotations(func(q *ent.OncallAnnotationQuery) {
			q.WithEvent()
		})

	handover, handoverErr := query.First(ctx)
	if handover == nil || handoverErr != nil {
		return nil, fmt.Errorf("failed to get handover: %w", handoverErr)
	}
	return s.sendShiftHandover(ctx, handover)
}

func (s *OncallService) sendShiftHandover(ctx context.Context, ho *ent.OncallUserShiftHandover) (*ent.OncallUserShiftHandover, error) {
	if !ho.SentAt.IsZero() {
		return ho, nil
	}

	if ho.UpdatedAt.IsZero() {
		// TODO: fill in template
	}

	var sections []rez.OncallShiftHandoverSection
	if jsonErr := json.Unmarshal(ho.Contents, &sections); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", jsonErr)
	}

	shift, shiftErr := ho.Edges.ShiftOrErr()
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get handover shift: %w", shiftErr)
	}

	nextShift, nextShiftErr := s.getNextShift(ctx, shift)
	if nextShiftErr != nil {
		return nil, fmt.Errorf("get next shift: %w", nextShiftErr)
	}

	roster, rosterErr := nextShift.Edges.RosterOrErr()
	if rosterErr != nil {
		return nil, fmt.Errorf("next shift roster: %w", rosterErr)
	}
	if roster.ChatChannelID == "" {
		return nil, fmt.Errorf("no roster chat channel found")
	}

	annos, annosErr := ho.Edges.PinnedAnnotationsOrErr()
	if annosErr != nil {
		return nil, fmt.Errorf("get pinned annotations: %w", annosErr)
	}

	params := rez.SendOncallHandoverParams{
		Content:           sections,
		EndingShift:       shift,
		StartingShift:     nextShift,
		PinnedAnnotations: annos,
	}
	if sendErr := s.chat.SendOncallHandover(ctx, params); sendErr != nil {
		return nil, fmt.Errorf("failed to send oncall handover: %w", sendErr)
	}

	updated, updateErr := ho.Update().SetSentAt(time.Now()).Save(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("update handover sent_at time: %w", updateErr)
	}

	return updated, nil
}

func (s *OncallService) HandleGenerateShiftMetrics(ctx context.Context, args jobs.GenerateShiftMetrics) error {
	log.Debug().Msg("generate shift metrics")
	return nil
}
