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

type EventsService struct {
	db         rez.Database
	situations rez.SituationService
}

func NewEventsService(db rez.Database, situations rez.SituationService) (*EventsService, error) {
	s := &EventsService{
		db:         db,
		situations: situations,
	}

	return s, nil
}

func (s *EventsService) GetEvent(ctx context.Context, id uuid.UUID, params rez.GetEventParams) (*ent.NormalizedEvent, error) {
	query := s.db.Client(ctx).NormalizedEvent.Query().
		Where(ne.ID(id))

	if params.WithProjection {
		query.WithProjection(func(pq *ent.NormalizedEventProjectionQuery) {
			pq.WithProjectionEntities()
		})
	}

	return query.Only(ctx)
}

func (s *EventsService) ListEvents(ctx context.Context, params rez.ListEventsParams) (*ent.ListResult[ent.NormalizedEvent], error) {
	query := s.db.Client(ctx).NormalizedEvent.Query()

	query.Order(ne.ByOccurredAt(params.GetOrder()), ne.ByID(params.GetOrder()))
	query.Where(params.Predicates...)
	if params.Kind != "" {
		query.Where(ne.KindEQ(params.Kind))
	}
	if !params.From.IsZero() {
		query.Where(ne.OccurredAtGTE(params.From))
	}
	if !params.To.IsZero() {
		query.Where(ne.OccurredAtLTE(params.To))
	}
	if params.SituationID != uuid.Nil {
		situationEventIDs, situationErr := s.situations.ListSituationEventIDs(ctx, params.SituationID)
		if situationErr != nil {
			return nil, fmt.Errorf("list situation events: %w", situationErr)
		}
		query.Where(ne.IDIn(situationEventIDs...))
	}
	if params.AnalysisID != uuid.Nil {
		// TODO
	}
	if params.WithAnnotations {
		//query.WithAnnotations(func(q *ent.EventAnnotationQuery) {
		//	if params.AnnotationRosterID != uuid.Nil {
		//		q.Where(oncallannotation.RosterID(params.AnnotationRosterID))
		//	}
		//})
	}
	if params.WithProjection {
		query.WithProjection(func(pq *ent.NormalizedEventProjectionQuery) {
			pq.WithProjectionEntities()
		})
	}

	return ent.DoListQuery[ent.NormalizedEvent, *ent.NormalizedEventQuery](ctx, query, params.ListParams)
}

func (s *EventsService) ListAnnotations(ctx context.Context, params rez.ListAnnotationsParams) (*ent.ListResult[ent.EventAnnotation], error) {
	query := s.db.Client(ctx).EventAnnotation.Query()
	query.Order(ea.ByCreatedAt(params.GetOrder()), ea.ByID(params.GetOrder()))

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

func (s *EventsService) GetAnnotation(ctx context.Context, id uuid.UUID) (*ent.EventAnnotation, error) {
	return s.db.Client(ctx).EventAnnotation.Query().
		Where(ea.ID(id)).
		WithCreator().
		WithEvent().
		Only(ctx)
}

func (s *EventsService) QueryAnnotation(ctx context.Context, pred predicate.EventAnnotation) (*ent.EventAnnotation, error) {
	query := s.db.Client(ctx).EventAnnotation.Query().
		Where(pred).
		WithEvent()
	return query.Only(ctx)
}

func (s *EventsService) SetAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error) {
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

func (s *EventsService) createAnnotation(ctx context.Context, anno *ent.EventAnnotation) (*ent.EventAnnotation, error) {
	eventId := anno.EventID
	if eventId == uuid.Nil && anno.Edges.Event != nil {
		eventQuery := s.db.Client(ctx).NormalizedEvent.Query().
			Where(ne.ProviderResourceRef(anno.Edges.Event.ProviderResourceRef))
		existingId, eventErr := eventQuery.OnlyID(ctx)
		if eventErr != nil && !ent.IsNotFound(eventErr) {
			return nil, fmt.Errorf("failed to check for existing oncall event: %w", eventErr)
		}
		eventId = existingId
	}
	createFn := func(txCtx context.Context, tx *ent.Client) (*ent.EventAnnotation, error) {
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

		createAnnotation := tx.EventAnnotation.Create().
			SetEventID(anno.EventID).
			SetCreatorID(anno.CreatorID).
			SetMinutesOccupied(anno.MinutesOccupied).
			SetNotes(anno.Notes).
			SetTags(anno.Tags)
		createdAnno, annoErr := createAnnotation.Save(txCtx)
		if annoErr != nil {
			return nil, fmt.Errorf("create annotation: %w", annoErr)
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
		return createdAnno, nil
	}
	created, txErr := ent.WithTxReturning(ctx, s.db, createFn)
	if txErr != nil {
		return nil, fmt.Errorf("creating annotation: %w", txErr)
	}
	return created, nil
}

func (s *EventsService) DeleteAnnotation(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).EventAnnotation.DeleteOneID(id).Exec(ctx)
}
