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
	db    *ent.Client
	jobs  rez.JobsService
	docs  rez.DocumentsService
	users rez.UserService
}

func NewOncallShiftsService(db *ent.Client, jobs rez.JobsService) (*OncallShiftsService, error) {
	s := &OncallShiftsService{
		db:   db,
		jobs: jobs,
	}

	return s, nil
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

func (s *OncallShiftsService) HandlePeriodicScanShifts(ctx context.Context, _ jobs.ScanOncallShifts) error {
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

func (s *OncallShiftsService) HandleEnsureShiftHandoverSent(ctx context.Context, args jobs.EnsureShiftHandoverSent) error {
	shiftId := args.ShiftId

	ho, hoErr := s.GetHandoverForShift(ctx, shiftId)
	if hoErr != nil {
		return fmt.Errorf("failed to get or create shift handover: %w", hoErr)
	}
	_, sendErr := s.sendShiftHandover(ctx, ho)

	return sendErr
}

func (s *OncallShiftsService) HandleEnsureShiftHandoverReminderSent(ctx context.Context, args jobs.EnsureShiftHandoverReminderSent) error {
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

	log.Debug().
		Str("shiftId", shift.ID.String()).
		Msg("send shift ending reminder")

	update := ho.Update().SetReminderSent(true)
	if updateErr := update.Exec(ctx); updateErr != nil {
		return fmt.Errorf("failed to set reminder_sent: %w", updateErr)
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

func (s *OncallShiftsService) SendShiftHandover(ctx context.Context, handoverId uuid.UUID) (*ent.OncallShiftHandover, error) {
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

func (s *OncallShiftsService) sendShiftHandover(ctx context.Context, ho *ent.OncallShiftHandover) (*ent.OncallShiftHandover, error) {
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
	log.Debug().
		Str("shiftId", shift.ID.String()).
		Interface("params", params).
		Msg("send shift handover")
	//if sendErr := s.chat.SendOncallHandover(ctx, params); sendErr != nil {
	//	return nil, fmt.Errorf("failed to send oncall handover: %w", sendErr)
	//}

	updated, updateErr := ho.Update().SetSentAt(time.Now()).Save(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("update handover sent_at time: %w", updateErr)
	}

	return updated, nil
}
