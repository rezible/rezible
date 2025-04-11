package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/texm/prosemirror-go"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncalleventannotation"
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

	if chat != nil {
		chat.SetCreateAnnotationFunc(s.createChatAnnotation)
	}

	if jobsErr := s.registerBackgroundJobs(); jobsErr != nil {
		return nil, fmt.Errorf("registering job workers: %w", jobsErr)
	}

	//go s.registerHandoverSchema()

	return s, nil
}

func (s *OncallService) registerBackgroundJobs() error {
	interval := time.Hour

	job := jobs.NewPeriodicJob(
		jobs.PeriodicInterval(interval),
		func() (jobs.JobArgs, *jobs.InsertOpts) {
			return &jobs.ScanOncallHandovers{}, nil
		},
		&jobs.PeriodicJobOpts{
			RunOnStart: true,
		},
	)
	s.jobs.RegisterPeriodicJob(job)

	return errors.Join(
		jobs.RegisterWorkerFunc(s.HandleScanForShiftsNeedingHandoverJob),
		jobs.RegisterWorkerFunc(s.HandleEnsureShiftHandoverJob),
	)
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

func (s *OncallService) createChatAnnotation(ctx context.Context, rosterId uuid.UUID, msg *rez.OncallEvent, setFn func(*ent.OncallEventAnnotation)) error {
	annos, annosErr := s.db.OncallEventAnnotation.Query().
		Where(oncalleventannotation.RosterID(rosterId)).
		All(ctx)
	if annosErr != nil {
		return fmt.Errorf("failed to query: %w", annosErr)
	}
	anno := &ent.OncallEventAnnotation{
		RosterID: rosterId,
		EventID:  msg.ID,
	}
	for _, an := range annos {
		if an.EventID != "" && an.EventID == msg.ID {
			anno = an
			break
		}
	}
	prevId := anno.ID.String()
	setFn(anno)
	if anno.ID.String() != prevId {
		return fmt.Errorf("annotation id mismatch: %s", anno.ID)
	}

	upsertQuery := s.db.OncallEventAnnotation.Create().
		SetRosterID(anno.RosterID).
		SetEventID(anno.EventID).
		SetNotes(anno.Notes).
		SetMinutesOccupied(anno.MinutesOccupied)
	if upsertErr := upsertQuery.Exec(ctx); upsertErr != nil {
		return fmt.Errorf("failed to upsert char annotation: %w", upsertErr)
	}

	return nil
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

func (s *OncallService) GetNextShift(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, error) {
	shift, shiftErr := s.db.OncallUserShift.Get(ctx, id)
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	return s.db.OncallUserShift.Query().
		Where(oncallusershift.And(
			oncallusershift.RosterID(shift.RosterID),
			oncallusershift.StartAtGTE(shift.EndAt),
			oncallusershift.IDNEQ(id))).
		Order(oncallusershift.ByStartAt(sql.OrderAsc())).
		WithUser().
		WithRoster().
		First(ctx)
}

func (s *OncallService) GetPreviousShift(ctx context.Context, id uuid.UUID) (*ent.OncallUserShift, error) {
	shift, shiftErr := s.db.OncallUserShift.Get(ctx, id)
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	return s.db.OncallUserShift.Query().
		Where(oncallusershift.And(
			oncallusershift.RosterID(shift.RosterID),
			oncallusershift.EndAtLTE(shift.StartAt),
			oncallusershift.IDNEQ(id))).
		Order(oncallusershift.ByStartAt(sql.OrderDesc())).
		WithUser().
		WithRoster().
		First(ctx)
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

func (s *OncallService) HandleScanForShiftsNeedingHandoverJob(ctx context.Context, args jobs.ScanOncallHandovers) error {
	// check for shifts ending within window, that don't have reminder sent
	window := time.Hour
	shiftEndingWithinWindow := oncallusershift.And(
		oncallusershift.EndAtGT(time.Now().Add(-window)),
		oncallusershift.EndAtLT(time.Now().Add(window)))
	endingShiftsQuery := s.db.OncallUserShift.Query().Where(shiftEndingWithinWindow)

	shifts, shiftsErr := endingShiftsQuery.All(ctx)
	if shiftsErr != nil {
		return fmt.Errorf("failed to get shifts: %w", shiftsErr)
	}

	shiftIds := make([]uuid.UUID, 0, len(shifts))
	for i, shift := range shifts {
		shiftIds[i] = shift.ID
	}

	sentShiftIds := mapset.NewSet[uuid.UUID]()
	sentHandovers, handoversErr := s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ShiftIDIn(shiftIds...)).
		Where(oncallusershifthandover.ReminderSent(true)).
		All(ctx)
	if handoversErr != nil {
		return fmt.Errorf("failed to get handovers: %w", handoversErr)
	}
	for _, h := range sentHandovers {
		sentShiftIds.Add(h.ShiftID)
	}

	var endingShiftIds []uuid.UUID
	for _, sh := range shifts {
		shiftId := sh.ID
		if !sentShiftIds.Contains(shiftId) {
			endingShiftIds = append(endingShiftIds, shiftId)
		}
	}

	if len(endingShiftIds) == 0 {
		return nil
	}

	params := make([]jobs.InsertManyParams, len(endingShiftIds))
	for i, id := range endingShiftIds {
		params[i] = jobs.InsertManyParams{
			Args: jobs.EnsureShiftHandover{ShiftId: id},
			Opts: &jobs.InsertOpts{
				Uniqueness: &jobs.UniquenessOpts{
					Args: true,
				},
			},
		}
	}

	if insertErr := s.jobs.InsertMany(ctx, params); insertErr != nil {
		return fmt.Errorf("could not insert jobs: %w", insertErr)
	}

	return nil
}

func (s *OncallService) HandleEnsureShiftHandoverJob(ctx context.Context, args jobs.EnsureShiftHandover) error {
	/*
		shiftId := args.ShiftId

		shift, shiftErr := s.GetShiftByID(ctx, shiftId)
		if shiftErr != nil {
			return fmt.Errorf("failed to get shift: %w", shiftErr)
		}

		ho, hoErr := s.GetHandoverForShift(ctx, shiftId, true)
		if hoErr != nil {
			return fmt.Errorf("failed to get or create shift handover: %w", hoErr)
		}

		// shift has already ended, within window
		if shift.EndAt.Before(time.Now()) {
			if sendErr := s.sendFallbackShiftHandover(ctx, ho); sendErr != nil {
				return fmt.Errorf("failed to send fallback handover: %w", sendErr)
			}
		} else {
			if reminderErr := s.sendShiftHandoverReminder(ctx, shift, ho); reminderErr != nil {
				return fmt.Errorf("failed to send handover reminder: %w", reminderErr)
			}
		}
	*/

	return nil
}

func (s *OncallService) sendShiftHandoverReminder(ctx context.Context, shift *ent.OncallUserShift, ho *ent.OncallUserShiftHandover) error {
	if ho.ReminderSent {
		return nil
	}

	user, userErr := s.users.GetById(ctx, shift.UserID)
	if userErr != nil {
		return fmt.Errorf("failed to get shift user: %w", userErr)
	}

	roster, rosterErr := s.GetRosterByID(ctx, shift.RosterID)
	if rosterErr != nil {
		return fmt.Errorf("failed to get roster: %w", rosterErr)
	}

	msgText := fmt.Sprintf("Your shift for %s is ending in %d minutes!\nPlease complete your handover",
		roster.Name, int(shift.EndAt.Sub(time.Now()).Minutes()))
	msgLinkUrl := fmt.Sprintf("%s/oncall/shifts/%s/handover", rez.FrontendUrl, shift.ID)
	if msgErr := s.chat.SendUserLinkMessage(ctx, user, msgText, msgLinkUrl, "Complete Handover"); msgErr != nil {
		return fmt.Errorf("failed to send message: %w", msgErr)
	}

	if updateErr := ho.Update().SetReminderSent(true).Exec(ctx); updateErr != nil {
		return fmt.Errorf("failed to set reminder_sent: %w", updateErr)
	}

	return nil
}

func (s *OncallService) GetRosterHandoverTemplate(ctx context.Context, rosterId uuid.UUID) (*ent.OncallHandoverTemplate, error) {
	roster, rosterErr := s.db.OncallRoster.Get(ctx, rosterId)
	if rosterErr != nil {
		return nil, fmt.Errorf("failed to get roster: %w", rosterErr)
	}
	if roster.HandoverTemplateID != uuid.Nil {
		return roster.QueryHandoverTemplate().Only(ctx)
	}
	return s.db.OncallHandoverTemplate.Query().
		Where(oncallhandovertemplate.Not(oncallhandovertemplate.HasRoster())).
		Where(oncallhandovertemplate.IsDefault(true)).
		Only(ctx)
}

func (s *OncallService) GetShiftHandover(ctx context.Context, id uuid.UUID) (*ent.OncallUserShiftHandover, error) {
	return s.db.OncallUserShiftHandover.Query().Where(oncallusershifthandover.ID(id)).WithPinnedAnnotations().Only(ctx)
}

func (s *OncallService) GetHandoverForShift(ctx context.Context, shiftId uuid.UUID, create bool) (*ent.OncallUserShiftHandover, error) {
	handover, queryErr := s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ShiftID(shiftId)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("failed to query shift handover: %w", queryErr)
	}
	if handover != nil || !create {
		return handover, queryErr
	}
	shift, shiftErr := s.db.OncallUserShift.Get(ctx, shiftId)
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
	}
	var contents []byte
	tmpl, tmplErr := s.GetRosterHandoverTemplate(ctx, shift.RosterID)
	if tmplErr != nil {
		log.Warn().Err(tmplErr).Msg("failed to get roster handover template")
		// TODO: use default template to create contents
		contents = []byte("[]")
	} else {
		contents = tmpl.Contents
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

//
//func (s *OncallService) sendFallbackShiftHandover(ctx context.Context, ho *ent.OncallUserShiftHandover) error {
//	if ho == nil || !ho.SentAt.IsZero() { // already sent
//		return nil
//	}
//	if sendErr := s.sendShiftHandover(ctx, ho); sendErr != nil {
//		return fmt.Errorf("sending fallback handover: %w", sendErr)
//	}
//
//	updateErr := ho.Update().SetSentAt(time.Now()).Exec(ctx)
//	if updateErr != nil {
//		return fmt.Errorf("failed to update handover sent_at time: %w", updateErr)
//	}
//
//	return nil
//}

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
	handover, handoverErr := s.db.OncallUserShiftHandover.Get(ctx, handoverId)
	if handover == nil || handoverErr != nil {
		return nil, fmt.Errorf("failed to get handover: %w", handoverErr)
	}

	if !handover.SentAt.IsZero() {
		return handover, nil
	}

	if sendErr := s.sendShiftHandover(ctx, handover); sendErr != nil {
		return nil, fmt.Errorf("failed to send handover: %w", sendErr)
	}

	updateErr := handover.Update().SetSentAt(time.Now()).Exec(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update handover sent_at time: %w", updateErr)
	}
	handover.SentAt = time.Now()

	return handover, nil
}

func (s *OncallService) sendShiftHandover(ctx context.Context, ho *ent.OncallUserShiftHandover) error {
	shift, shiftErr := s.GetShiftByID(ctx, ho.ShiftID)
	if shiftErr != nil {
		return fmt.Errorf("failed to query shift: %w", shiftErr)
	}

	nextShift, nextShiftErr := s.GetNextShift(ctx, ho.ShiftID)
	if nextShiftErr != nil {
		return fmt.Errorf("failed to get next shift: %w", nextShiftErr)
	}

	var content []rez.OncallShiftHandoverSection
	if jsonErr := json.Unmarshal(ho.Contents, &content); jsonErr != nil {
		return fmt.Errorf("failed to unmarshal content: %w", jsonErr)
	}

	var includeAnnotations, includeIncidents bool
	for _, sec := range content {
		if sec.Kind == "annotations" {
			includeAnnotations = true
		}
		if sec.Kind == "incidents" {
			includeIncidents = true
		}
	}

	var annotations []*ent.OncallEventAnnotation
	if includeAnnotations {
		var listErr error
		annotations, listErr = s.db.OncallEventAnnotation.Query().
			Where(oncalleventannotation.HasHandoversWith(oncallusershifthandover.ID(ho.ID))).All(ctx)
		if listErr != nil && !ent.IsNotFound(listErr) {
			return fmt.Errorf("failed to query pinned annotations: %w", listErr)
		}
	}

	var incidents []*ent.Incident
	if includeIncidents {
		var listErr error
		incidents, listErr = s.incidents.ListIncidents(ctx, rez.ListIncidentsParams{
			UserId: shift.UserID,
			// OpenedAfter:  shift.StartAt,
			OpenedBefore: shift.EndAt,
		})
		if listErr != nil && !ent.IsNotFound(listErr) {
			return fmt.Errorf("failed to query incidents: %w", listErr)
		}
	}

	params := rez.SendOncallHandoverParams{
		Content:       content,
		EndingShift:   shift,
		StartingShift: nextShift,
		Incidents:     incidents,
		Annotations:   annotations,
	}
	if sendErr := s.chat.SendOncallHandover(ctx, params); sendErr != nil {
		return fmt.Errorf("failed to send: %w", sendErr)
	}

	return nil
}

func (s *OncallService) ListEventAnnotations(ctx context.Context, params rez.ListOncallEventAnnotationsParams) ([]*ent.OncallEventAnnotation, error) {
	query := s.db.OncallEventAnnotation.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset)

	rosterId := params.RosterID
	if params.ShiftID != uuid.Nil {
		shift, shiftErr := s.GetShiftByID(ctx, params.ShiftID)
		if shiftErr != nil {
			return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
		}
		rosterId = shift.RosterID
		query.Where(oncalleventannotation.And(
			oncalleventannotation.CreatedAtGT(shift.StartAt),
			oncalleventannotation.CreatedAtLT(shift.EndAt)))
	}
	if rosterId != uuid.Nil {
		query.Where(oncalleventannotation.HasRosterWith(oncallroster.ID(rosterId)))
	}

	annos, annosErr := query.All(params.GetQueryContext(ctx))
	if annosErr != nil {
		return nil, fmt.Errorf("query annotations: %w", annosErr)
	}

	return annos, nil
}

func (s *OncallService) GetEventAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallEventAnnotation, error) {
	return s.db.OncallEventAnnotation.Get(ctx, id)
}

func (s *OncallService) CreateEventAnnotation(ctx context.Context, anno *ent.OncallEventAnnotation) (*ent.OncallEventAnnotation, error) {
	query := s.db.OncallEventAnnotation.Create().
		SetID(uuid.New()).
		SetEventID(anno.EventID).
		SetMinutesOccupied(anno.MinutesOccupied).
		SetNotes(anno.Notes).
		SetRosterID(anno.RosterID).
		OnConflictColumns(oncalleventannotation.FieldID).
		UpdateNewValues()

	if err := query.Exec(ctx); err != nil {
		return nil, fmt.Errorf("upsert oncall annotation: %w", err)
	}
	return anno, nil
}

func (s *OncallService) DeleteEventAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.OncallEventAnnotation.DeleteOneID(id).Exec(ctx)
}
