package db

import (
	"context"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ne "github.com/rezible/rezible/ent/normalizedevent"
)

type EventsService struct {
	db    *ent.Client
	users rez.UserService
}

func NewEventsService(dbc *ent.Client, users rez.UserService) (*EventsService, error) {
	s := &EventsService{
		db:    dbc,
		users: users,
	}

	return s, nil
}

func (s *EventsService) GetEvent(ctx context.Context, id uuid.UUID) (*ent.NormalizedEvent, error) {
	return s.db.NormalizedEvent.Get(ctx, id)
}

func (s *EventsService) ListEvents(ctx context.Context, params rez.ListEventsParams) (*ent.ListResult[ent.NormalizedEvent], error) {
	query := s.db.NormalizedEvent.Query()

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
