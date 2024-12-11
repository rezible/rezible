package postgres

import (
	"context"
	"encoding/json"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/texm/prosemirror-go"
	"github.com/twohundreds/rezible/ent/oncallhandovertemplate"
	"github.com/twohundreds/rezible/ent/oncallusershifthandover"
	"time"

	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/jobs"

	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/oncallroster"
	"github.com/twohundreds/rezible/ent/oncallschedule"
	"github.com/twohundreds/rezible/ent/oncallscheduleparticipant"
	"github.com/twohundreds/rezible/ent/oncallusershift"
	"github.com/twohundreds/rezible/ent/oncallusershiftannotation"
	"github.com/twohundreds/rezible/ent/predicate"
)

type OncallService struct {
	db        *ent.Client
	jobClient *jobs.BackgroundJobClient
	loader    rez.ProviderLoader
	provider  rez.OncallDataProvider
	docs      rez.DocumentsService
	chat      rez.ChatService
	users     rez.UserService
	incidents rez.IncidentService
}

func NewOncallService(ctx context.Context, db *ent.Client, jobClient *jobs.BackgroundJobClient, pl rez.ProviderLoader, docs rez.DocumentsService, chat rez.ChatService, users rez.UserService, incidents rez.IncidentService) (*OncallService, error) {
	s := &OncallService{
		db:        db,
		jobClient: jobClient,
		loader:    pl,
		docs:      docs,
		chat:      chat,
		users:     users,
		incidents: incidents,
	}

	if provErr := s.LoadDataProvider(ctx); provErr != nil {
		return nil, provErr
	}

	if jobsErr := s.RegisterJobs(); jobsErr != nil {
		return nil, fmt.Errorf("failed to register background job: %w", jobsErr)
	}

	go s.registerHandoverSchema()

	return s, nil
}

func (s *OncallService) LoadDataProvider(ctx context.Context) error {
	provider, providerErr := s.loader.LoadOncallDataProvider(ctx)
	if providerErr != nil {
		return fmt.Errorf("failed to load data provider: %w", providerErr)
	}
	s.provider = provider
	return nil
}

func (s *OncallService) SyncData(ctx context.Context) error {
	dataSyncer := newOncallDataSyncer(s.db, s.users, s.provider)
	return dataSyncer.syncProviderData(ctx)
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
	to := shift.EndAt.Add(time.Hour * 24 * 7)
	withinWindow := oncallusershift.And(oncallusershift.StartAtGTE(shift.EndAt), oncallusershift.StartAtLTE(to))

	return s.db.OncallUserShift.Query().
		Where(oncallusershift.RosterID(shift.RosterID)).
		Where(withinWindow).
		Order(oncallusershift.ByStartAt(sql.OrderAsc())).
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

func (s *OncallService) GetRosterByScheduleId(ctx context.Context, scheduleId uuid.UUID) (*ent.OncallRoster, error) {
	return s.db.OncallSchedule.Query().
		Where(oncallschedule.ID(scheduleId)).
		QueryRoster().
		Only(ctx)
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

func (s *OncallService) SendShiftHandoverReminder(ctx context.Context, shiftId uuid.UUID) error {
	args := oncallHandoverReminderJobArgs{shiftId: shiftId}
	_, jobErr := s.jobClient.Insert(ctx, args, nil)
	if jobErr != nil {
		return fmt.Errorf("insert handover reminder job failed: %w", jobErr)
	}
	return nil
}

func (s *OncallService) sendShiftHandoverReminder(ctx context.Context, shiftId uuid.UUID) error {
	shift, shiftErr := s.GetShiftByID(ctx, shiftId)
	if shiftErr != nil {
		return fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	user, userErr := s.users.GetById(ctx, shift.UserID)
	if userErr != nil {
		return fmt.Errorf("failed to get shift user: %w", userErr)
	}

	roster, rosterErr := s.GetRosterByID(ctx, shift.RosterID)
	if rosterErr != nil {
		return fmt.Errorf("failed to get roster: %w", rosterErr)
	}

	msgText := fmt.Sprintf(
		"Your shift for %s ends soon!\nPlease complete your handover",
		roster.Name,
	)
	msgLinkUrl := fmt.Sprintf("%s/oncall/shifts/%s/handover", rez.FrontendUrl, shiftId)
	msgLinkText := "Complete Handover"
	if msgErr := s.chat.SendUserLinkMessage(ctx, user, msgText, msgLinkUrl, msgLinkText); msgErr != nil {
		log.Error().Err(msgErr).Msg("Failed to send oncall shift handover reminder")
	}
	return nil
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

func (s *OncallService) SendShiftHandover(ctx context.Context, id uuid.UUID, contents []rez.OncallShiftHandoverSection) (*ent.OncallUserShiftHandover, error) {
	shift, shiftErr := s.GetShiftByID(ctx, id)
	if shiftErr != nil {
		return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
	}

	jsonContent, jsonErr := json.Marshal(contents)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal contents: %w", jsonErr)
	}

	handover, queryErr := s.GetShiftHandover(ctx, id)
	var updateContentsErr error
	if queryErr != nil {
		if !ent.IsNotFound(queryErr) {
			return nil, fmt.Errorf("failed to get handover: %w", queryErr)
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

	if sendErr := s.sendShiftHandover(ctx, shift, contents); sendErr != nil {
		return nil, fmt.Errorf("failed to send handover: %w", sendErr)
	}

	updateErr := handover.Update().SetSentAt(time.Now()).Exec(ctx)
	if updateErr != nil {
		return nil, fmt.Errorf("failed to update handover sent_at time: %w", updateErr)
	}
	handover.SentAt = time.Now()

	return handover, nil
}

func (s *OncallService) sendShiftHandover(ctx context.Context, shift *ent.OncallUserShift, content []rez.OncallShiftHandoverSection) error {
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
		SetEventID(anno.EventID).
		SetEventKind(anno.EventKind).
		SetTitle(anno.Title).
		SetOccurredAt(anno.OccurredAt).
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