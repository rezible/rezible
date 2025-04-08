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
	"github.com/rezible/rezible/ent/oncallhandovertemplate"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallschedule"
	"github.com/rezible/rezible/ent/oncallscheduleparticipant"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rezible/rezible/ent/oncallusershiftannotation"
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

func (s *OncallService) createChatAnnotation(ctx context.Context, shiftId uuid.UUID, msg *ent.OncallEvent, setFn func(*ent.OncallUserShiftAnnotation)) error {
	annos, annosErr := s.db.OncallUserShiftAnnotation.Query().
		Where(oncallusershiftannotation.ShiftID(shiftId)).
		All(ctx)
	if annosErr != nil {
		return fmt.Errorf("failed to query: %w", annosErr)
	}
	anno := &ent.OncallUserShiftAnnotation{
		ShiftID: shiftId,
		Event:   msg,
	}
	for _, an := range annos {
		if an.Event != nil && an.Event.ID == msg.ID {
			anno = an
			break
		}
	}
	prevId := anno.ID.String()
	setFn(anno)
	if anno.ID.String() != prevId {
		return fmt.Errorf("annotation id mismatch: %s", anno.ID)
	}

	upsertQuery := s.db.OncallUserShiftAnnotation.Create().
		SetShiftID(anno.ShiftID).
		SetEvent(anno.Event).
		SetNotes(anno.Notes).
		SetPinned(anno.Pinned).
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
	shiftId := args.ShiftId

	shift, shiftErr := s.GetShiftByID(ctx, shiftId)
	if shiftErr != nil {
		return fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	ho, hoErr := s.getOrCreateShiftHandover(ctx, shift)
	if hoErr != nil {
		return fmt.Errorf("failed to get or create shift handover: %w", hoErr)
	}

	// shift has already ended, within window
	if shift.EndAt.Before(time.Now()) {
		if sendErr := s.sendFallbackShiftHandover(ctx, shift, ho); sendErr != nil {
			return fmt.Errorf("failed to send fallback handover: %w", sendErr)
		}
	} else {
		if reminderErr := s.sendShiftHandoverReminder(ctx, shift, ho); reminderErr != nil {
			return fmt.Errorf("failed to send handover reminder: %w", reminderErr)
		}
	}

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

func (s *OncallService) getOrCreateShiftHandover(ctx context.Context, shift *ent.OncallUserShift) (*ent.OncallUserShiftHandover, error) {
	ho, hoErr := s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ShiftID(shift.ID)).Only(ctx)
	if hoErr != nil && !ent.IsNotFound(hoErr) {
		return nil, fmt.Errorf("failed to get handover: %w", hoErr)
	}
	if ho != nil {
		return ho, nil
	}
	var contents []byte
	tmpl, tmplErr := s.GetRosterHandoverTemplate(ctx, shift.RosterID)
	if tmplErr != nil {
		log.Warn().Err(tmplErr).Msg("failed to get roster handover template")
	} else {
		contents = tmpl.Contents
	}
	return s.createShiftHandover(ctx, shift.ID, contents)
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

func (s *OncallService) GetShiftHandover(ctx context.Context, shiftId uuid.UUID) (*ent.OncallUserShiftHandover, error) {
	return s.db.OncallUserShiftHandover.Query().
		Where(oncallusershifthandover.ShiftID(shiftId)).
		Only(ctx)
}

func (s *OncallService) createShiftHandover(ctx context.Context, shiftId uuid.UUID, contents []byte) (*ent.OncallUserShiftHandover, error) {
	return s.db.OncallUserShiftHandover.Create().
		SetShiftID(shiftId).
		SetContents(contents).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
}

func (s *OncallService) sendFallbackShiftHandover(ctx context.Context, shift *ent.OncallUserShift, ho *ent.OncallUserShiftHandover) error {
	if !ho.SentAt.IsZero() { // already sent
		return nil
	}
	contents := []rez.OncallShiftHandoverSection{
		// TODO:
		{Header: "Annotations", Kind: "annotations"},
		{Header: "Incidents", Kind: "incidents"},
	}
	if sendErr := s.sendShiftHandoverMessage(ctx, shift, contents); sendErr != nil {
		return fmt.Errorf("sending fallback handover: %w", sendErr)
	}

	updateErr := ho.Update().SetSentAt(time.Now()).Exec(ctx)
	if updateErr != nil {
		return fmt.Errorf("failed to update handover sent_at time: %w", updateErr)
	}

	return nil
}

func (s *OncallService) SendShiftHandover(ctx context.Context, id uuid.UUID, contents []rez.OncallShiftHandoverSection) (*ent.OncallUserShiftHandover, error) {
	shift, shiftErr := s.GetShiftByID(ctx, id)
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	jsonContent, jsonErr := json.Marshal(contents)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal contents: %w", jsonErr)
	}

	handover, handoverErr := s.GetShiftHandover(ctx, id)

	// already sent
	if handover != nil && !handover.SentAt.IsZero() {
		return handover, nil
	}

	var updateContentsErr error
	if handoverErr != nil || handover == nil {
		if !ent.IsNotFound(handoverErr) {
			return nil, fmt.Errorf("failed to get handover: %w", handoverErr)
		}
		handover, updateContentsErr = s.createShiftHandover(ctx, id, jsonContent)
	} else {
		handover, updateContentsErr = handover.Update().
			SetContents(jsonContent).
			SetUpdatedAt(time.Now()).
			Save(ctx)
	}
	if updateContentsErr != nil {
		return nil, fmt.Errorf("failed to set handover contents: %w", updateContentsErr)
	}

	if sendErr := s.sendShiftHandoverMessage(ctx, shift, contents); sendErr != nil {
		return nil, fmt.Errorf("failed to send handover: %w", sendErr)
	}

	updateErr := handover.Update().SetSentAt(time.Now()).Exec(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update handover sent_at time: %w", updateErr)
	}
	handover.SentAt = time.Now()

	return handover, nil
}

func (s *OncallService) sendShiftHandoverMessage(ctx context.Context, shift *ent.OncallUserShift, content []rez.OncallShiftHandoverSection) error {
	nextShift, nextShiftErr := s.GetNextShift(ctx, shift.ID)
	if nextShiftErr != nil {
		return fmt.Errorf("failed to get next shift: %w", nextShiftErr)
	}

	var includeAnnotations, includeIncidents bool
	for _, sec := range content {
		if sec.Kind != "annotations" {
			includeAnnotations = true
		}
		if sec.Kind == "incidents" {
			includeIncidents = true
		}
	}

	var annotations []*ent.OncallUserShiftAnnotation
	if includeAnnotations {
		var listErr error
		annotations, listErr = s.ListShiftAnnotations(ctx, rez.ListOncallShiftAnnotationsParams{
			ShiftID: shift.ID,
			Pinned:  &includeAnnotations,
		})
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

func (s *OncallService) ListShiftAnnotations(ctx context.Context, params rez.ListOncallShiftAnnotationsParams) ([]*ent.OncallUserShiftAnnotation, error) {
	query := s.db.OncallUserShiftAnnotation.Query().
		Where(oncallusershiftannotation.ShiftID(params.ShiftID)).
		Limit(params.GetLimit()).
		Offset(params.Offset)

	if params.Pinned != nil {
		query.Where(oncallusershiftannotation.Pinned(*params.Pinned))
	}

	annos, annosErr := query.All(params.GetQueryContext(ctx))
	if annosErr != nil {
		return nil, fmt.Errorf("query annotations: %w", annosErr)
	}

	return annos, nil
}

func (s *OncallService) GetShiftAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallUserShiftAnnotation, error) {
	return s.db.OncallUserShiftAnnotation.Get(ctx, id)
}

func (s *OncallService) CreateShiftAnnotation(ctx context.Context, anno *ent.OncallUserShiftAnnotation) (*ent.OncallUserShiftAnnotation, error) {
	query := s.db.OncallUserShiftAnnotation.Create().
		SetID(uuid.New()).
		SetShiftID(anno.ShiftID).
		SetEvent(anno.Event).
		SetMinutesOccupied(anno.MinutesOccupied).
		SetNotes(anno.Notes).
		SetPinned(anno.Pinned).
		OnConflictColumns(oncallusershiftannotation.FieldID).
		UpdateNewValues()

	if err := query.Exec(ctx); err != nil {
		return nil, fmt.Errorf("upsert oncall annotation: %w", err)
	}
	return anno, nil
}

func (s *OncallService) ArchiveShiftAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.OncallUserShiftAnnotation.DeleteOneID(id).Exec(ctx)
}
