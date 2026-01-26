package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ohot "github.com/rezible/rezible/ent/oncallhandovertemplate"
	"github.com/rezible/rezible/ent/oncallroster"
	ocs "github.com/rezible/rezible/ent/oncallshift"
	"github.com/rezible/rezible/ent/oncallshifthandover"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/jobs"
)

type OncallShiftsService struct {
	db           *ent.Client
	jobs         rez.JobsService
	integrations rez.IntegrationsService
	docs         rez.DocumentsService
	users        rez.UserService
}

func NewOncallShiftsService(db *ent.Client, jobSvc rez.JobsService, integrations rez.IntegrationsService) (*OncallShiftsService, error) {
	s := &OncallShiftsService{
		db:           db,
		jobs:         jobSvc,
		integrations: integrations,
	}

	jobs.RegisterPeriodicJob(jobs.ScanOncallShiftsPeriodicJob)
	jobs.RegisterWorkerFunc(s.periodicScanShifts)
	jobs.RegisterWorkerFunc(s.ensureShiftHandoverReminderSent)
	jobs.RegisterWorkerFunc(s.ensureShiftHandoverSent)

	return s, nil
}

func (s *OncallShiftsService) periodicScanShifts(ctx context.Context, _ jobs.ScanOncallShifts) error {
	return s.scanShifts(ctx)
}

func (s *OncallShiftsService) ensureShiftHandoverSent(ctx context.Context, args jobs.EnsureShiftHandoverSent) error {
	_, err := s.SendShiftHandover(ctx, args.ShiftId)
	return err
}

func (s *OncallShiftsService) ensureShiftHandoverReminderSent(ctx context.Context, args jobs.EnsureShiftHandoverReminderSent) error {
	return s.sendShiftHandoverReminder(ctx, args.ShiftId)
}

func (s *OncallShiftsService) GetShiftByID(ctx context.Context, id uuid.UUID) (*ent.OncallShift, error) {
	query := s.db.OncallShift.Query().
		Where(ocs.ID(id)).
		WithRoster().
		WithUser()
	return query.Only(ctx)
}

func (s *OncallShiftsService) getNextShift(ctx context.Context, shift *ent.OncallShift) (*ent.OncallShift, error) {
	return s.db.OncallShift.Query().
		Where(ocs.RosterID(shift.RosterID)).
		Where(ocs.IDNEQ(shift.ID)).
		Where(ocs.StartAtGTE(shift.StartAt)).
		Order(ocs.ByStartAt(sql.OrderAsc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallShiftsService) getPreviousShift(ctx context.Context, shift *ent.OncallShift) (*ent.OncallShift, error) {
	return s.db.OncallShift.Query().
		Where(ocs.RosterID(shift.RosterID)).
		Where(ocs.IDNEQ(shift.ID)).
		Where(ocs.EndAtLTE(shift.StartAt)).
		Order(ocs.ByEndAt(sql.OrderDesc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallShiftsService) GetAdjacentShifts(ctx context.Context, id uuid.UUID) (*ent.OncallShift, *ent.OncallShift, error) {
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

func (s *OncallShiftsService) ListShifts(ctx context.Context, params rez.ListOncallShiftsParams) (*ent.ListResult[*ent.OncallShift], error) {
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

func (s *OncallShiftsService) queryShiftsEndingWithinWindow(ctx context.Context, window time.Duration) ([]*ent.OncallShift, error) {
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

func (s *OncallShiftsService) scanShifts(ctx context.Context) error {
	shifts, shiftsErr := s.queryShiftsEndingWithinWindow(ctx, time.Hour)
	if shiftsErr != nil {
		return fmt.Errorf("failed to get shifts: %w", shiftsErr)
	}

	reminderWindow := time.Minute * 10

	var params []jobs.InsertManyParams
	for _, shift := range shifts {
		ho := shift.Edges.Handover
		reminderSent := ho != nil && ho.ReminderSent // !ho.ReminderSentAt.IsZero
		if !reminderSent {
			params = append(params, jobs.InsertManyParams{
				Args: jobs.EnsureShiftHandoverReminderSent{ShiftId: shift.ID},
				InsertOpts: &jobs.InsertOpts{
					ScheduledAt: shift.EndAt.Add(-reminderWindow),
					UniqueOpts: jobs.UniqueOpts{
						ByArgs: true,
					},
				},
			})
		}
		isSent := ho != nil && !ho.SentAt.IsZero()
		if !isSent {
			params = append(params, jobs.InsertManyParams{
				Args: jobs.EnsureShiftHandoverSent{ShiftId: shift.ID},
				InsertOpts: &jobs.InsertOpts{
					ScheduledAt: shift.EndAt,
					UniqueOpts: jobs.UniqueOpts{
						ByArgs: true,
					},
				},
			})
		}
		params = append(params, jobs.InsertManyParams{
			Args: jobs.GenerateShiftMetrics{ShiftId: shift.ID},
			InsertOpts: &jobs.InsertOpts{
				ScheduledAt: shift.EndAt,
				UniqueOpts: jobs.UniqueOpts{
					ByArgs: true,
				},
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

func (s *OncallShiftsService) GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallShiftHandover, error) {
	return s.db.OncallShiftHandover.Query().
		Where(oncallshifthandover.ID(id)).
		WithPinnedAnnotations().
		Only(ctx)
}

func (s *OncallShiftsService) getRosterForShift(ctx context.Context, id uuid.UUID) (*ent.OncallRoster, error) {
	return s.db.OncallShift.Query().
		Where(ocs.ID(id)).
		QueryRoster().
		Only(ctx)
}

func (s *OncallShiftsService) GetHandoverForShift(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error) {
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

var defaultHandoverTemplate = []byte(`[
{"kind":"regular","header":"Overview"},
{"kind":"regular","header":"Handoff Tasks"},
{"kind":"regular","header":"Things to Monitor"},
{"kind":"annotations","header":"Pinned Annotations"}
]`)

func (s *OncallShiftsService) getRosterHandoverTemplateContents(ctx context.Context, rosterId uuid.UUID) ([]byte, error) {
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

func (s *OncallShiftsService) createShiftHandover(ctx context.Context, shiftId uuid.UUID) (*ent.OncallShiftHandover, error) {
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

func (s *OncallShiftsService) UpdateShiftHandover(ctx context.Context, update *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error) {
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

func (s *OncallShiftsService) sendShiftHandoverReminder(ctx context.Context, shiftId uuid.UUID) error {
	ho, hoErr := s.GetHandoverForShift(ctx, shiftId)
	if hoErr != nil {
		return fmt.Errorf("failed to get or create shift handover: %w", hoErr)
	}

	log.Debug().
		Str("shiftId", shiftId.String()).
		Bool("reminderSent", ho.ReminderSent).
		Msg("send shift ending reminder")
	if ho.ReminderSent {
		return nil
	}

	updateErr := ho.Update().SetReminderSent(true).Exec(ctx)
	if updateErr != nil {
		return fmt.Errorf("failed to set reminder_sent: %w", updateErr)
	}
	return nil
}

func (s *OncallShiftsService) SendShiftHandover(ctx context.Context, handoverId uuid.UUID) (*ent.OncallShiftHandover, error) {
	hoQuery := s.db.OncallShiftHandover.Query().
		Where(oncallshifthandover.ID(handoverId)).
		WithShift(func(q *ent.OncallShiftQuery) {
			q.WithRoster()
		}).
		WithPinnedAnnotations(func(q *ent.EventAnnotationQuery) {
			q.WithEvent()
		})

	handover, handoverErr := hoQuery.First(ctx)
	if handover == nil || handoverErr != nil {
		return nil, fmt.Errorf("failed to get handover: %w", handoverErr)
	}
	if !handover.SentAt.IsZero() {
		return handover, nil
	}

	return s.sendShiftHandover(ctx, handover)
}

func (s *OncallShiftsService) sendShiftHandover(ctx context.Context, ho *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error) {
	var sections []rez.OncallShiftHandoverSection
	if ho.UpdatedAt.IsZero() {
		// TODO: fill in from template
	} else if jsonErr := json.Unmarshal(ho.Contents, &sections); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal content: %w", jsonErr)
	}

	shift, shiftErr := ho.Edges.ShiftOrErr()
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get handover shift: %w", shiftErr)
	}

	roster, rosterErr := shift.Edges.RosterOrErr()
	if rosterErr != nil {
		return nil, fmt.Errorf("next shift roster: %w", rosterErr)
	}
	if roster.ChatChannelID == "" {
		return nil, fmt.Errorf("no roster chat channel found")
	}

	_, annosErr := ho.Edges.PinnedAnnotationsOrErr()
	if annosErr != nil {
		return nil, fmt.Errorf("get pinned annotations: %w", annosErr)
	}

	cs, csErr := s.integrations.GetChatService(ctx)
	if csErr != nil {
		return nil, fmt.Errorf("failed to get chat service: %w", csErr)
	}

	_, msgErr := cs.SendMessage(ctx, roster.ChatChannelID, &rez.ContentNode{})
	if msgErr != nil {
		return nil, fmt.Errorf("failed to send message: %w", msgErr)
	}

	updated, updateErr := ho.Update().SetSentAt(time.Now()).Save(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("update handover sent_at time: %w", updateErr)
	}

	return updated, nil
}
