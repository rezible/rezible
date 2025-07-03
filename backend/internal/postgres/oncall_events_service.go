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
	query := s.db.OncallEvent.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset).
		Where(oncallevent.And(oncallevent.TimestampGT(params.From), oncallevent.TimestampLT(params.To)))

	if params.RosterID != uuid.Nil {
		query.Where(oncallevent.RosterID(params.RosterID))
	}

	return query.All(params.GetQueryContext(ctx))
}

func (s *OncallEventsService) GetProviderEvent(ctx context.Context, providerId string) (*ent.OncallEvent, error) {
	return s.db.OncallEvent.Query().Where(oncallevent.ProviderID(providerId)).First(ctx)
}

func (s *OncallEventsService) ListAnnotations(ctx context.Context, params rez.ListOncallAnnotationsParams) ([]*ent.OncallAnnotation, error) {
	query := s.db.OncallAnnotation.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset)

	if params.WithCreator {
		query.WithCreator()
	}
	if params.WithRoster {
		query.WithRoster()
	}
	if params.WithAlertFeedback {
		query.WithAlertFeedback()
	}

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
		eventId := anno.EventID
		if eventId == uuid.Nil {
			event := anno.Edges.Event
			if event == nil {
				return fmt.Errorf("no event supplied")
			}
			createEvent := tx.OncallEvent.Create().
				SetProviderID(event.ProviderID).
				SetSource(event.Source).
				SetKind(event.Kind).
				SetTitle(event.Title).
				SetDescription(event.Description).
				SetTimestamp(event.Timestamp)

			// TODO: handle event existing
			//OnConflictColumns(oncallevent.FieldProviderID, oncallevent.FieldSource).
			//UpdateNewValues()

			createdEvent, eventErr := createEvent.Save(ctx)
			if eventErr != nil {
				return fmt.Errorf("create event: %w", eventErr)
			}
			eventId = createdEvent.ID
		}

		createdAnno, annoErr := tx.OncallAnnotation.Create().
			SetEventID(eventId).
			SetRosterID(anno.RosterID).
			SetCreatorID(anno.CreatorID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetTags(anno.Tags).
			Save(ctx)
		if annoErr != nil {
			return fmt.Errorf("create annotation: %w", annoErr)
		}

		alertFb := anno.Edges.AlertFeedback
		if alertFb != nil {
			createdFb, fbErr := tx.OncallAnnotationAlertFeedback.Create().
				SetDocumentationAvailable(alertFb.DocumentationAvailable).
				SetActionable(alertFb.Actionable).
				SetAccurate(alertFb.Accurate).
				SetAnnotation(createdAnno).
				Save(ctx)
			if fbErr != nil {
				return fmt.Errorf("create alert feedback: %w", fbErr)
			}
			createdAnno.Edges.AlertFeedback = createdFb
		}
		created = createdAnno
		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, createFn); txErr != nil {
		return nil, fmt.Errorf("creating annotation: %w", txErr)
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
		return nil, fmt.Errorf("querying current annotation: %w", currErr)
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
