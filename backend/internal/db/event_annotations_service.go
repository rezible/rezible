package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/event"
	ea "github.com/rezible/rezible/ent/eventannotation"
	"github.com/rezible/rezible/ent/predicate"
)

type EventAnnotationsService struct {
	db     *ent.Client
	events rez.EventsService
}

func NewEventAnnotationsService(db *ent.Client, events rez.EventsService) (*EventAnnotationsService, error) {
	s := &EventAnnotationsService{
		db:     db,
		events: events,
	}

	return s, nil
}

func (s *EventAnnotationsService) ListAnnotations(ctx context.Context, params rez.ListAnnotationsParams) (*ent.ListResult[*ent.EventAnnotation], error) {
	query := s.db.EventAnnotation.Query()

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

	return ent.DoListQuery[*ent.EventAnnotation, *ent.EventAnnotationQuery](ctx, query, params.ListParams)
}

func (s *EventAnnotationsService) GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.EventAnnotation, error) {
	return s.db.EventAnnotation.Query().
		Where(ea.ID(id)).
		WithCreator().
		WithEvent().
		Only(ctx)
}

func (s *EventAnnotationsService) LookupByUserEvent(ctx context.Context, userId uuid.UUID, ev *ent.Event) (*ent.EventAnnotation, error) {
	var eventPred predicate.Event

	if ev.ID != uuid.Nil {
		eventPred = event.ID(ev.ID)
	} else if ev.ExternalID != "" {
		eventPred = event.ExternalID(ev.ExternalID)
	}

	query := s.db.EventAnnotation.Query().
		Where(ea.CreatorID(userId)).
		Where(ea.HasEventWith(eventPred)).
		WithEvent()

	return query.Only(ctx)
}

func (s *EventAnnotationsService) SetAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error) {
	_, currErr := s.GetAnnotation(ctx, anno.ID)
	if currErr != nil {
		if ent.IsNotFound(currErr) {
			return s.createAnnotation(ctx, anno)
		}
		return nil, fmt.Errorf("querying current annotation: %w", currErr)
	}
	updated, annoErr := s.db.EventAnnotation.UpdateOneID(anno.ID).
		SetMinutesOccupied(anno.MinutesOccupied).
		SetNotes(anno.Notes).
		SetTags(anno.Tags).
		Save(ctx)
	if annoErr != nil {
		return nil, fmt.Errorf("failed to update annotation: %w", annoErr)
	}
	return updated, nil
}

func (s *EventAnnotationsService) createAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error) {
	var created *ent.EventAnnotation
	eventId := anno.EventID
	if eventId == uuid.Nil && anno.Edges.Event != nil {
		eventQuery := s.db.Event.Query().
			Where(event.ExternalID(anno.Edges.Event.ExternalID))
		existingId, eventErr := eventQuery.OnlyID(ctx)
		if eventErr != nil && !ent.IsNotFound(eventErr) {
			return nil, fmt.Errorf("failed to check for existing oncall event: %w", eventErr)
		}
		eventId = existingId
	}
	createFn := func(tx *ent.Tx) error {
		if eventId == uuid.Nil {
			e := anno.Edges.Event
			if anno.Edges.Event == nil {
				return fmt.Errorf("oncall annotation event is empty")
			}
			createEvent := tx.Event.Create().
				SetExternalID(e.ExternalID).
				SetSource(e.Source).
				SetKind(e.Kind).
				SetTitle(e.Title).
				SetDescription(e.Description).
				SetTimestamp(e.Timestamp)
			createdEvent, eventErr := createEvent.Save(ctx)
			if eventErr != nil {
				return fmt.Errorf("create annotation event: %w", eventErr)
			}
			anno.EventID = createdEvent.ID
		}

		createdAnno, annoErr := tx.EventAnnotation.Create().
			SetEventID(anno.EventID).
			SetCreatorID(anno.CreatorID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetTags(anno.Tags).
			Save(ctx)
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
	if txErr := ent.WithTx(ctx, s.db, createFn); txErr != nil {
		return nil, fmt.Errorf("creating annotation: %w", txErr)
	}
	return created, nil
}

func (s *EventAnnotationsService) DeleteAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.EventAnnotation.DeleteOneID(id).Exec(ctx)
}
