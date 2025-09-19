package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ohot "github.com/rezible/rezible/ent/oncallhandovertemplate"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	ocsp "github.com/rezible/rezible/ent/oncallscheduleparticipant"
	ocs "github.com/rezible/rezible/ent/oncallshift"
	"github.com/rezible/rezible/ent/oncallshifthandover"
	"github.com/rezible/rezible/ent/oncallshiftmetrics"
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

func NewOncallService(db *ent.Client, jobs rez.JobsService, docs rez.DocumentsService, chat rez.ChatService, users rez.UserService, incidents rez.IncidentService) (*OncallService, error) {
	s := &OncallService{
		db:        db,
		jobs:      jobs,
		docs:      docs,
		chat:      chat,
		users:     users,
		incidents: incidents,
	}

	return s, nil
}

func (s *OncallService) ListSchedules(ctx context.Context, params rez.ListOncallSchedulesParams) (*ent.ListResult[*ent.OncallSchedule], error) {
	var query *ent.OncallScheduleQuery
	if params.UserID != uuid.Nil {
		query = s.db.OncallScheduleParticipant.Query().
			Where(ocsp.UserID(params.UserID)).
			QuerySchedule()
	} else {
		query = s.db.OncallSchedule.Query()
	}

	return ent.DoListQuery[*ent.OncallSchedule, *ent.OncallScheduleQuery](ctx, query, params.ListParams)
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

func (s *OncallService) ListRosters(ctx context.Context, params rez.ListOncallRostersParams) (*ent.ListResult[*ent.OncallRoster], error) {
	query := s.db.OncallRoster.Query()

	if params.UserID != uuid.Nil {
		schedList, schedulesErr := s.ListSchedules(ctx, rez.ListOncallSchedulesParams{
			UserID: params.UserID,
		})
		if schedulesErr != nil {
			return nil, fmt.Errorf("failed to list oncall schedules: %w", schedulesErr)
		}
		var rosterIds []uuid.UUID
		for _, sched := range schedList.Data {
			rosterIds = append(rosterIds, sched.RosterID)
		}
		query = query.Where(oncallroster.IDIn(rosterIds...))
	}

	return ent.DoListQuery[*ent.OncallRoster, *ent.OncallRosterQuery](ctx, query, params.ListParams)
}

func (s *OncallService) GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallShift, error) {
	query := s.db.OncallShift.Query().
		Where(ocs.ID(id)).
		WithRoster().
		WithUser()
	return query.Only(ctx)
}

func (s *OncallService) getNextShift(ctx context.Context, shift *ent.OncallShift) (*ent.OncallShift, error) {
	return s.db.OncallShift.Query().
		Where(ocs.RosterID(shift.RosterID)).
		Where(ocs.IDNEQ(shift.ID)).
		Where(ocs.StartAtGTE(shift.StartAt)).
		Order(ocs.ByStartAt(sql.OrderAsc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallService) getPreviousShift(ctx context.Context, shift *ent.OncallShift) (*ent.OncallShift, error) {
	return s.db.OncallShift.Query().
		Where(ocs.RosterID(shift.RosterID)).
		Where(ocs.IDNEQ(shift.ID)).
		Where(ocs.EndAtLTE(shift.StartAt)).
		Order(ocs.ByEndAt(sql.OrderDesc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallService) GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallShift, *ent.OncallShift, error) {
	shift, shiftErr := s.db.OncallShift.Get(ctx, id)
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

func (s *OncallService) ListShifts(ctx context.Context, params rez.ListOncallShiftsParams) (*ent.ListResult[*ent.OncallShift], error) {
	query := s.db.OncallShift.Query().
		Order(ocs.ByEndAt(sql.OrderDesc())).
		WithRoster().
		WithUser()

	var predicates []predicate.OncallShift
	if !params.Anchor.IsZero() && !(params.Window.Milliseconds() == 0) {
		from := params.Anchor.Add(-params.Window)
		to := params.Anchor.Add(params.Window)
		withinWindow := ocs.And(ocs.EndAtGTE(from), ocs.StartAtLTE(to))
		predicates = append(predicates, withinWindow)
	}
	if params.UserID != uuid.Nil {
		predicates = append(predicates, ocs.UserID(params.UserID))
	}
	if len(predicates) > 0 {
		query = query.Where(predicates...)
	}

	return ent.DoListQuery[*ent.OncallShift, *ent.OncallShiftQuery](ctx, query, params.ListParams)
}

func (s *OncallService) queryShiftsEndingWithinWindow(ctx context.Context, window time.Duration) ([]*ent.OncallShift, error) {
	windowStart := time.Now().Add(-window)
	windowEnd := time.Now().Add(window)
	shiftEndingWithinWindow := ocs.And(ocs.EndAtGTE(windowStart), ocs.EndAtLTE(windowEnd))

	query := s.db.OncallShift.Query().
		Where(shiftEndingWithinWindow).
		WithHandover()

	shifts, shiftsErr := query.All(ctx)
	if shiftsErr != nil {
		return nil, fmt.Errorf("failed to get shifts: %w", shiftsErr)
	}

	return shifts, nil
}

func (s *OncallService) MakeScanShiftsPeriodicJob() jobs.PeriodicJob {
	interval := time.Hour
	return jobs.PeriodicJob{
		ConstructorFunc: func() jobs.InsertJobParams {
			return jobs.InsertJobParams{Args: &jobs.ScanOncallShifts{}}
		},
		Interval: interval,
		Opts:     &jobs.PeriodicJobOpts{RunOnStart: true},
	}
}

func (s *OncallService) HandlePeriodicScanShifts(ctx context.Context, _ jobs.ScanOncallShifts) error {
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

	ho, hoErr := s.GetHandoverForShift(ctx, shiftId)
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

	ho, hoErr := s.GetHandoverForShift(ctx, shiftId)
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
	isRosterOrDefault := ohot.Or(ohot.IsDefault(true), ohot.HasRosterWith(oncallroster.ID(rosterId)))
	tmpl, tmplErr := s.db.OncallHandoverTemplate.Query().
		Where(isRosterOrDefault).
		Order(ohot.ByUpdatedAt()).
		First(ctx)
	if tmplErr != nil {
		if ent.IsNotFound(tmplErr) {
			return defaultHandoverTemplate, nil
		}
		return nil, fmt.Errorf("failed to get roster handover template: %w", tmplErr)
	}
	return tmpl.Contents, nil
}

func (s *OncallService) GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error) {
	return s.db.OncallShiftHandover.Query().
		Where(oncallshifthandover.ID(id)).
		WithPinnedAnnotations().
		Only(ctx)
}

func (s *OncallService) getRosterForShift(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error) {
	return s.db.OncallShift.Query().
		Where(ocs.ID(id)).
		QueryRoster().
		Only(ctx)
}

func (s *OncallService) GetHandoverForShift(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error) {
	handover, queryErr := s.db.OncallShiftHandover.Query().
		Where(oncallshifthandover.ShiftID(shiftId)).
		WithPinnedAnnotations().
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("failed to query shift handover: %w", queryErr)
	}
	if handover != nil {
		return handover, nil
	}

	return s.createShiftHandover(ctx, shiftId)
}

func (s *OncallService) createShiftHandover(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error) {
	roster, rosterErr := s.getRosterForShift(ctx, shiftId)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to get roster: %w", rosterErr)
	}

	contents, contentsErr := s.getRosterHandoverTemplateContents(ctx, roster.ID)
	if contentsErr != nil {
		return nil, fmt.Errorf("failed to get roster handover template contents: %w", contentsErr)
	}

	return s.db.OncallShiftHandover.Create().
		SetShiftID(shiftId).
		SetContents(contents).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
}

func (s *OncallService) UpdateShiftHandover(ctx context.Context, update *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error) {
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

func (s *OncallService) SendShiftHandover(ctx context.Context, handoverId uuid.UUID) (*ent.OncallShiftHandover, error) {
	query := s.db.OncallShiftHandover.Query().
		Where(oncallshifthandover.ID(handoverId)).
		WithShift().
		WithPinnedAnnotations(func(q *ent.EventAnnotationQuery) {
			q.WithEvent()
		})

	handover, handoverErr := query.First(ctx)
	if handover == nil || handoverErr != nil {
		return nil, fmt.Errorf("failed to get handover: %w", handoverErr)
	}
	return s.sendShiftHandover(ctx, handover)
}

func (s *OncallService) sendShiftHandover(ctx context.Context, ho *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error) {
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

func (s *OncallService) queryShiftMetrics(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftMetrics, error) {
	return s.db.OncallShiftMetrics.Query().Where(oncallshiftmetrics.ShiftID(shiftId)).Only(ctx)
}

func (s *OncallService) getShiftIncidents(ctx context.Context, shift *ent.OncallShift) ([]*ent.Incident, error) {
	return nil, nil
}

func (s *OncallService) upsertShiftMetrics(ctx context.Context, m *ent.OncallShiftMetrics) (*ent.OncallShiftMetrics, error) {
	create := s.db.OncallShiftMetrics.Create().
		SetShiftID(m.ShiftID).
		SetBurdenScore(m.BurdenScore).
		SetEventFrequency(m.EventFrequency).
		SetLifeImpact(m.LifeImpact).
		SetTimeImpact(m.TimeImpact).
		SetResponseRequirements(m.ResponseRequirements).
		SetIsolation(m.Isolation).
		SetEventsTotal(m.EventsTotal).
		SetIncidentsTotal(m.IncidentsTotal).
		SetIncidentResponseTime(m.IncidentResponseTime).
		SetAlertsTotal(m.AlertsTotal).
		SetInterruptsTotal(m.InterruptsTotal).
		SetInterruptsBusinessHours(m.InterruptsBusinessHours).
		SetInterruptsNight(m.InterruptsNight)

	upsert := create.OnConflict(sql.ConflictColumns(oncallshiftmetrics.ShiftColumn)).
		UpdateNewValues()

	metricsId, upsertErr := upsert.ID(ctx)
	if upsertErr != nil {
		return nil, fmt.Errorf("create or update shift metrics: %w", upsertErr)
	}
	m.ID = metricsId
	return m, nil
}

func (s *OncallService) generateMetricsForShift(ctx context.Context, id uuid.UUID) (*ent.OncallShiftMetrics, error) {
	shift, shiftErr := s.GetShiftByID(ctx, id)
	if shiftErr != nil {
		return nil, shiftErr
	}

	incidents, incErr := s.getShiftIncidents(ctx, shift)
	if incErr != nil {
		return nil, fmt.Errorf("shift incidents: %w", incErr)
	}

	m := &ent.OncallShiftMetrics{
		ShiftID: id,

		BurdenScore:          0,
		EventFrequency:       0,
		LifeImpact:           0,
		TimeImpact:           0,
		ResponseRequirements: 0,
		Isolation:            0,

		EventsTotal:          0,
		IncidentsTotal:       float32(len(incidents)),
		IncidentResponseTime: 0,

		AlertsTotal:             0,
		InterruptsTotal:         0,
		InterruptsNight:         0,
		InterruptsBusinessHours: 0,
	}
	return s.upsertShiftMetrics(ctx, m)
}

func (s *OncallService) GetShiftMetrics(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftMetrics, error) {
	metrics, metErr := s.queryShiftMetrics(ctx, shiftId)
	if metErr == nil {
		return metrics, nil
	}
	if !ent.IsNotFound(metErr) {
		return nil, fmt.Errorf("querying metrics: %w", metErr)
	}
	generated, genErr := s.generateMetricsForShift(ctx, shiftId)
	if genErr != nil {
		return nil, fmt.Errorf("generating metrics: %w", genErr)
	}
	return generated, nil
}

func (s *OncallService) GetComparisonShiftMetrics(ctx context.Context, from, to time.Time) (*ent.OncallShiftMetrics, error) {
	// TODO
	return &ent.OncallShiftMetrics{
		BurdenScore:          5.9,
		EventFrequency:       4.3,
		LifeImpact:           4.5,
		TimeImpact:           4.2,
		ResponseRequirements: 3.0,
		Isolation:            3.4,

		IncidentsTotal:       1.1,
		IncidentResponseTime: 33,

		InterruptsTotal:         19,
		AlertsTotal:             15,
		InterruptsNight:         4,
		InterruptsBusinessHours: 8,
	}, nil
}

func (s *OncallService) HandleGenerateShiftMetrics(ctx context.Context, args jobs.GenerateShiftMetrics) error {
	_, genErr := s.generateMetricsForShift(ctx, args.ShiftId)
	return genErr
}
