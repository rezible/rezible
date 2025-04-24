package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rezible/rezible/ent/oncallevent"
	"github.com/rezible/rezible/ent/oncallroster"
	"math/rand"
	"time"
)

type OncallEventsService struct {
	db        *ent.Client
	users     rez.UserService
	oncall    rez.OncallService
	incidents rez.IncidentService
}

func NewOncallEventsService(ctx context.Context, db *ent.Client, users rez.UserService, oncall rez.OncallService, incidents rez.IncidentService) (*OncallEventsService, error) {
	s := &OncallEventsService{
		db:        db,
		users:     users,
		oncall:    oncall,
		incidents: incidents,
	}

	return s, nil
}

func makeFakeShiftEvent(date time.Time) *ent.OncallEvent {
	isAlert := rand.Float64() > 0.25
	eventKind := "incident"
	if isAlert {
		eventKind = "alert"
	}

	hour := rand.Intn(24)
	minute := rand.Intn(60)

	timestamp := time.Date(
		date.Year(), date.Month(), date.Day(),
		hour, minute, 0, 0, date.Location(),
	)

	id := uuid.New()

	return &ent.OncallEvent{
		ID:          id,
		ProviderID:  id.String(),
		Timestamp:   timestamp,
		Source:      "fake",
		Kind:        eventKind,
		Title:       "title",
		Description: "fake description",
	}
}

func makeFakeOncallEvents(start, end time.Time) []*ent.OncallEvent {
	numHours := end.Sub(start).Hours()
	if numHours <= 0 {
		return nil
	}
	numDays := int(numHours / 24)
	events := make([]*ent.OncallEvent, 0, numDays*10)

	for day := 0; day < numDays; day++ {
		dayDate := start.AddDate(0, 0, day)
		numDayEvents := rand.Intn(10)

		for i := 0; i < numDayEvents; i++ {
			events = append(events, makeFakeShiftEvent(dayDate))
		}
	}

	return events
}

func (s *OncallEventsService) ListEvents(ctx context.Context, params rez.ListOncallEventsParams) ([]*ent.OncallEvent, error) {
	events := makeFakeOncallEvents(params.Start, params.End)

	return events, nil
}

func (s *OncallEventsService) CreateEventAnnotation(ctx context.Context, evAnno rez.OncallEventAnnotation) error {
	return nil
}

func (s *OncallEventsService) QueryUserChatMessageEventDetails(ctx context.Context, userChatId string, msgId string) ([]*ent.OncallRoster, *ent.OncallEvent, error) {
	user, userErr := s.users.GetByChatId(ctx, userChatId)
	if userErr != nil {
		return nil, nil, userErr
	}

	rosters, rostersErr := user.QueryOncallSchedules().QuerySchedule().QueryRoster().All(ctx)
	if rostersErr != nil && !ent.IsNotFound(rostersErr) {
		return nil, nil, fmt.Errorf("failed to query oncall rosters for user: %w", rostersErr)
	}

	rosterIds := make([]uuid.UUID, len(rosters))
	for i, r := range rosters {
		rosterIds[i] = r.ID
	}

	// Get event by message id
	event, eventErr := s.db.OncallEvent.Query().Where(oncallevent.ProviderID(msgId)).Only(ctx)
	if eventErr != nil && !ent.IsNotFound(eventErr) {
		return nil, nil, fmt.Errorf("failed to query oncall event for msg: %w", eventErr)
	}

	return rosters, event, nil
}

/*
func (s *OncallService) CreateAnnotation(ctx context.Context, rosterId uuid.UUID, ev *rez.OncallEvent, anno *ent.OncallAnnotation) error {
	annos, annosErr := s.db.OncallAnnotation.Query().
		Where(oncallannotation.RosterID(rosterId)).
		All(ctx)
	if annosErr != nil {
		return fmt.Errorf("failed to query: %w", annosErr)
	}

	for _, an := range annos {
		if an.EventID != "" && an.EventID == msg.ID {
			anno = an
			break
		}
	}
	prevId := anno.ID.String()
	anno := &ent.OncallAnnotation{
		RosterID: rosterId,
		EventID:  msg.ID,
	}
	setFn(anno)
	if anno.ID.String() != prevId {
		return fmt.Errorf("annotation id mismatch: %s", anno.ID)
	}

	upsertQuery := s.db.OncallAnnotation.Create().
		SetRosterID(anno.RosterID).
		SetEventID(anno.EventID).
		SetNotes(anno.Notes).
		SetMinutesOccupied(anno.MinutesOccupied)
	if upsertErr := upsertQuery.Exec(ctx); upsertErr != nil {
		return fmt.Errorf("failed to upsert char annotation: %w", upsertErr)
	}

	return nil
}
*/

func (s *OncallEventsService) ListAnnotations(ctx context.Context, params rez.ListOncallAnnotationsParams) ([]*ent.OncallAnnotation, error) {
	query := s.db.OncallAnnotation.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset).
		WithCreator().
		WithEvent().
		WithRoster()

	rosterId := params.RosterID
	if params.ShiftID != uuid.Nil {
		shift, shiftErr := s.oncall.GetShiftByID(ctx, params.ShiftID)
		if shiftErr != nil {
			return nil, fmt.Errorf("failed to get shift: %w", shiftErr)
		}
		rosterId = shift.RosterID
		query.Where(oncallannotation.And(
			oncallannotation.CreatedAtGT(shift.StartAt),
			oncallannotation.CreatedAtLT(shift.EndAt)))
	}
	if rosterId != uuid.Nil {
		query.Where(oncallannotation.HasRosterWith(oncallroster.ID(rosterId)))
	}

	annos, annosErr := query.All(params.GetQueryContext(ctx))
	if annosErr != nil {
		return nil, fmt.Errorf("query annotations: %w", annosErr)
	}

	return annos, nil
}

func (s *OncallEventsService) GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallAnnotation, error) {
	return s.db.OncallAnnotation.Query().
		WithCreator().WithRoster().WithEvent().
		Where(oncallannotation.ID(id)).
		Only(ctx)
}

func (s *OncallEventsService) CreateAnnotation(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error) {
	query := s.db.OncallAnnotation.Create().
		SetID(uuid.New()).
		SetEventID(anno.EventID).
		SetMinutesOccupied(anno.MinutesOccupied).
		SetNotes(anno.Notes).
		SetRosterID(anno.RosterID).
		OnConflictColumns(oncallannotation.FieldID).
		UpdateNewValues()

	if err := query.Exec(ctx); err != nil {
		return nil, fmt.Errorf("upsert oncall annotation: %w", err)
	}
	return anno, nil
}

func (s *OncallEventsService) DeleteAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.OncallAnnotation.DeleteOneID(id).Exec(ctx)
}
