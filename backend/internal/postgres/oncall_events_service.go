package postgres

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rezible/rezible/ent/oncallannotationalertfeedback"
	"github.com/rezible/rezible/ent/oncallevent"
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

func (s *OncallEventsService) GetEvent(ctx context.Context, id uuid.UUID) (*ent.OncallEvent, error) {
	return s.db.OncallEvent.Get(ctx, id)
}

func (s *OncallEventsService) ListEvents(ctx context.Context, params rez.ListOncallEventsParams) ([]*ent.OncallEvent, int, error) {
	query := s.db.OncallEvent.Query().
		Where(oncallevent.And(oncallevent.TimestampGT(params.From), oncallevent.TimestampLT(params.To)))
	if params.RosterID != uuid.Nil {
		query.Where(oncallevent.RosterID(params.RosterID))
	}
	order := sql.OrderDesc()
	if params.OrderAsc {
		order = sql.OrderAsc()
	}
	query.Order(oncallevent.ByTimestamp(order))
	query.Offset(params.Offset)
	query.Limit(params.GetLimit())
	if params.WithAnnotations {
		query.WithAnnotations(func(q *ent.OncallAnnotationQuery) {
			if params.RosterID != uuid.Nil {
				q.Where(oncallannotation.RosterID(params.RosterID))
			}
		})
	}

	return ent.RunCountableQuery[*ent.OncallEvent](params.GetQueryContext(ctx), query, params.Count)
}

func (s *OncallEventsService) GetProviderEvent(ctx context.Context, providerId string) (*ent.OncallEvent, error) {
	return s.db.OncallEvent.Query().Where(oncallevent.ProviderID(providerId)).First(ctx)
}

func (s *OncallEventsService) ListAnnotations(ctx context.Context, params rez.ListOncallAnnotationsParams) ([]*ent.OncallAnnotation, int, error) {
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
	if params.WithEvent {
		query.WithEvent()
	}

	rosterId := params.RosterID
	if params.Shift != nil {
		rosterId = params.Shift.RosterID
		query.Where(oncallannotation.And(
			oncallannotation.CreatedAtGT(params.Shift.StartAt),
			oncallannotation.CreatedAtLT(params.Shift.EndAt)))
	} else {
		if !params.From.IsZero() {
			query.Where(oncallannotation.CreatedAtGT(params.From))
		}
		if !params.To.IsZero() {
			query.Where(oncallannotation.CreatedAtLT(params.To))
		}
	}
	if rosterId != uuid.Nil {
		query.Where(oncallannotation.RosterID(rosterId))
	}

	return ent.RunCountableQuery[*ent.OncallAnnotation](params.GetQueryContext(ctx), query, params.Count)
}

func (s *OncallEventsService) GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.OncallAnnotation, error) {
	return s.db.OncallAnnotation.Query().
		Where(oncallannotation.ID(id)).
		WithCreator().WithRoster().WithEvent().WithAlertFeedback().
		Only(ctx)
}

func (s *OncallEventsService) createAnnotation(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error) {
	var created *ent.OncallAnnotation
	eventId := anno.EventID
	if eventId == uuid.Nil && anno.Edges.Event != nil {
		e := anno.Edges.Event
		pred := oncallevent.And(oncallevent.ProviderID(e.ProviderID))
		existingId, eventErr := s.db.OncallEvent.Query().Where(pred).OnlyID(ctx)
		if eventErr != nil && !ent.IsNotFound(eventErr) {
			return nil, fmt.Errorf("failed to check for existing oncall event: %w", eventErr)
		}
		eventId = existingId
	}
	createFn := func(tx *ent.Tx) error {
		if eventId == uuid.Nil {
			event := anno.Edges.Event
			if anno.Edges.Event == nil {
				return fmt.Errorf("oncall annotation event is empty")
			}
			createEvent := tx.OncallEvent.Create().
				SetProviderID(event.ProviderID).
				SetSource(event.Source).
				SetKind(event.Kind).
				SetTitle(event.Title).
				SetDescription(event.Description).
				SetTimestamp(event.Timestamp)
			var eventErr error
			event, eventErr = createEvent.Save(ctx)
			if eventErr != nil {
				return fmt.Errorf("upsert event: %w", eventErr)
			}
			anno.EventID = event.ID
		}

		createdAnno, annoErr := tx.OncallAnnotation.Create().
			SetEventID(anno.EventID).
			SetRosterID(anno.RosterID).
			SetCreatorID(anno.CreatorID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetTags(anno.Tags).
			Save(ctx)
		if annoErr != nil {
			return fmt.Errorf("create annotation: %w", annoErr)
		}

		if alertFb := anno.Edges.AlertFeedback; alertFb != nil {
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
		if ent.IsNotFound(currErr) {
			return s.createAnnotation(ctx, anno)
		}
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
