package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rezible/rezible/ent/oncallannotationalertfeedback"
	"github.com/rezible/rezible/ent/oncallevent"
	"github.com/rezible/rezible/ent/oncallroster"
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

func (s *OncallEventsService) ListEvents(ctx context.Context, params rez.ListOncallEventsParams) ([]*ent.OncallEvent, error) {
	withinWindow := oncallevent.And(oncallevent.TimestampGT(params.Start), oncallevent.TimestampLT(params.End))
	query := s.db.OncallEvent.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset).
		Where(withinWindow)

	if params.WithAnnotations {
		query.WithAnnotations()
	}

	return query.All(params.GetQueryContext(ctx))
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

func (s *OncallEventsService) ListAnnotations(ctx context.Context, params rez.ListOncallAnnotationsParams) ([]*ent.OncallAnnotation, error) {
	query := s.db.OncallAnnotation.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset).
		WithCreator().
		WithEvent().
		WithRoster().
		WithAlertFeedback()

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
		Where(oncallannotation.ID(id)).
		WithCreator().WithRoster().WithEvent().WithAlertFeedback().
		Only(ctx)
}

func (s *OncallEventsService) CreateAnnotation(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error) {
	var created *ent.OncallAnnotation
	createFn := func(tx *ent.Tx) error {
		createdAnno, annoErr := tx.OncallAnnotation.Create().
			SetEventID(anno.EventID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetRosterID(anno.RosterID).
			Save(ctx)
		if annoErr != nil {
			return fmt.Errorf("failed to create annotation: %w", annoErr)
		}
		if alertFb := anno.Edges.AlertFeedback; alertFb != nil {
			createdFb, fbErr := tx.OncallAnnotationAlertFeedback.Create().
				SetDocumentationAvailable(alertFb.DocumentationAvailable).
				SetActionable(alertFb.Actionable).
				SetAccurate(alertFb.Accurate).
				SetAnnotation(createdAnno).
				Save(ctx)
			if fbErr != nil {
				return fmt.Errorf("failed to create alert feedback: %w", fbErr)
			}
			createdAnno.Edges.AlertFeedback = createdFb
		}
		created = createdAnno
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, createFn); txErr != nil {
		return nil, fmt.Errorf("failed to create annotation: %w", txErr)
	}
	return created, nil
}

func (s *OncallEventsService) UpdateAnnotation(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error) {
	dbAnno, currErr := s.db.OncallAnnotation.Query().
		Where(oncallannotation.ID(anno.ID)).
		WithAlertFeedback().
		WithEvent().
		Only(ctx)
	if currErr != nil {
		return nil, fmt.Errorf("failed to query current annotation: %w", currErr)
	}
	dbAlertFb := dbAnno.Edges.AlertFeedback
	updated := dbAnno
	updateFn := func(tx *ent.Tx) error {
		updatedAnno, annoErr := tx.OncallAnnotation.UpdateOneID(anno.ID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetTags(anno.Tags).
			Save(ctx)
		if annoErr != nil {
			return fmt.Errorf("failed to update annotation: %w", annoErr)
		}

		if fb := anno.Edges.AlertFeedback; fb != nil {
			upsert := tx.OncallAnnotationAlertFeedback.Create()
			if dbAlertFb != nil {
				upsert.SetID(dbAlertFb.ID)
			}
			updateFb := upsert.
				SetAccurate(fb.Accurate).
				SetActionable(fb.Actionable).
				SetDocumentationAvailable(fb.DocumentationAvailable).
				SetAnnotationID(updatedAnno.ID).
				OnConflictColumns(oncallannotationalertfeedback.FieldID).
				UpdateNewValues()
			fbId, updateErr := updateFb.ID(ctx)
			if updateErr != nil {
				return fmt.Errorf("failed to update alert feedback: %w", updateErr)
			}
			fb.ID = fbId
			updatedAnno.Edges.AlertFeedback = fb
		}

		updated = updatedAnno
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, updateFn); txErr != nil {
		return nil, fmt.Errorf("failed to update annotation: %w", txErr)
	}
	return updated, nil
}

func (s *OncallEventsService) DeleteAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.OncallAnnotation.DeleteOneID(id).Exec(ctx)
}
