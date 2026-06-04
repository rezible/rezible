package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ea "github.com/rezible/rezible/ent/eventannotation"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/predicate"
)

type EventService struct {
	db rez.Database
}

func NewEventService(db rez.Database) (*EventService, error) {
	s := &EventService{
		db: db,
	}

	return s, nil
}

func (s *EventService) GetEvent(ctx context.Context, id uuid.UUID) (*ent.NormalizedEvent, error) {
	return s.db.Client(ctx).NormalizedEvent.Get(ctx, id)
}

func (s *EventService) ListEvents(ctx context.Context, params rez.ListEventsParams) (*ent.ListResult[ent.NormalizedEvent], error) {
	query := s.db.Client(ctx).NormalizedEvent.Query()

	query.Order(ne.ByOccurredAt(params.GetOrder()))
	query.Where(params.Predicates...)

	if params.WithAnnotations {
		//query.WithAnnotations(func(q *ent.EventAnnotationQuery) {
		//	if params.AnnotationRosterID != uuid.Nil {
		//		q.Where(oncallannotation.RosterID(params.AnnotationRosterID))
		//	}
		//})
	}

	return ent.DoListQuery[ent.NormalizedEvent, *ent.NormalizedEventQuery](ctx, query, params.ListParams)
}

func (s *EventService) ListAnnotations(ctx context.Context, params rez.ListAnnotationsParams) (*ent.ListResult[ent.EventAnnotation], error) {
	query := s.db.Client(ctx).EventAnnotation.Query()

	if !params.From.IsZero() {
		query.Where(ea.CreatedAtGTE(params.From))
	}
	if !params.To.IsZero() {
		query.Where(ea.CreatedAtLTE(params.To))
	}

	if params.Expand.WithCreator {
		query.WithCreator()
	}
	if params.Expand.WithEvent {
		query.WithEvent()
	}

	return ent.DoListQuery[ent.EventAnnotation, *ent.EventAnnotationQuery](ctx, query, params.ListParams)
}

func (s *EventService) GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.EventAnnotation, error) {
	return s.db.Client(ctx).EventAnnotation.Query().
		Where(ea.ID(id)).
		WithCreator().
		WithEvent().
		Only(ctx)
}

func (s *EventService) QueryAnnotation(ctx context.Context, pred predicate.EventAnnotation) (*ent.EventAnnotation, error) {
	query := s.db.Client(ctx).EventAnnotation.Query().
		Where(pred).
		WithEvent()
	return query.Only(ctx)
}

func (s *EventService) SetAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error) {
	_, currErr := s.GetAnnotation(ctx, anno.ID)
	if currErr != nil {
		if ent.IsNotFound(currErr) {
			return s.createAnnotation(ctx, anno)
		}
		return nil, fmt.Errorf("querying current annotation: %w", currErr)
	}
	updated, annoErr := s.db.Client(ctx).EventAnnotation.UpdateOneID(anno.ID).
		SetMinutesOccupied(anno.MinutesOccupied).
		SetNotes(anno.Notes).
		SetTags(anno.Tags).
		Save(ctx)
	if annoErr != nil {
		return nil, fmt.Errorf("failed to update annotation: %w", annoErr)
	}
	return updated, nil
}

func (s *EventService) createAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error) {
	var created *ent.EventAnnotation
	eventId := anno.EventID
	if eventId == uuid.Nil && anno.Edges.Event != nil {
		eventQuery := s.db.Client(ctx).NormalizedEvent.Query().
			Where(ne.ProviderSubjectRef(anno.Edges.Event.ProviderSubjectRef))
		existingId, eventErr := eventQuery.OnlyID(ctx)
		if eventErr != nil && !ent.IsNotFound(eventErr) {
			return nil, fmt.Errorf("failed to check for existing oncall event: %w", eventErr)
		}
		eventId = existingId
	}
	createFn := func(txCtx context.Context, tx *ent.Client) error {
		//if eventId == uuid.Nil {
		//	e := anno.Edges.Event
		//	if anno.Edges.Event == nil {
		//		return fmt.Errorf("oncall annotation event is empty")
		//	}
		//	createEvent := tx.NormalizedEvent.Create().
		//		SetExternalID(e.ExternalID).
		//		SetSource(e.Source).
		//		SetKind(e.Kind).
		//		SetTitle(e.Title).
		//		SetDescription(e.Description).
		//		SetTimestamp(e.Timestamp)
		//	createdEvent, eventErr := createEvent.Save(ctx)
		//	if eventErr != nil {
		//		return fmt.Errorf("create annotation event: %w", eventErr)
		//	}
		//	anno.EventID = createdEvent.ID
		//}

		createdAnno, annoErr := tx.EventAnnotation.Create().
			SetEventID(anno.EventID).
			SetCreatorID(anno.CreatorID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetTags(anno.Tags).
			Save(txCtx)
		if annoErr != nil {
			return fmt.Errorf("create annotation: %w", annoErr)
		}

		//if alertFb := anno.Edges.AlertFeedback; alertFb != nil {
		//	createdFb, fbErr := tx.AlertFeedback.Create().
		//		SetDocumentationAvailable(alertFb.DocumentationAvailable).
		//		SetActionable(alertFb.Actionable).
		//		SetAccurate(alertFb.Accurate).
		//		SetAnnotation(createdAnno).
		//		Save(ctx)
		//	if fbErr != nil {
		//		return fmt.Errorf("create alert feedback: %w", fbErr)
		//	}
		//	createdAnno.Edges.AlertFeedback = createdFb
		//}
		created = createdAnno
		return nil
	}
	if txErr := s.db.WithTx(ctx, createFn); txErr != nil {
		return nil, fmt.Errorf("creating annotation: %w", txErr)
	}
	return created, nil
}

func (s *EventService) DeleteAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).EventAnnotation.DeleteOneID(id).Exec(ctx)
}
