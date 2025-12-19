package db

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oe "github.com/rezible/rezible/ent/event"
)

type EventsService struct {
	db    *ent.Client
	users rez.UserService
}

func NewEventsService(db *ent.Client, users rez.UserService) (*EventsService, error) {
	s := &EventsService{
		db:    db,
		users: users,
	}

	return s, nil
}

func (s *EventsService) GetEvent(ctx context.Context, id uuid.UUID) (*ent.Event, error) {
	return s.db.Event.Get(ctx, id)
}

func (s *EventsService) ListEvents(ctx context.Context, params rez.ListEventsParams) (*ent.ListResult[*ent.Event], error) {
	query := s.db.Event.Query()

	query.Order(oe.ByTimestamp(params.GetOrder()))
	query.Where(oe.And(oe.TimestampGT(params.From), oe.TimestampLT(params.To)))

	if params.WithAnnotations {
		query.WithAnnotations(func(q *ent.EventAnnotationQuery) {
			//if params.AnnotationRosterID != uuid.Nil {
			//	q.Where(oncallannotation.RosterID(params.AnnotationRosterID))
			//}
		})
	}

	return ent.DoListQuery[*ent.Event, *ent.EventQuery](ctx, query, params.ListParams)
}
