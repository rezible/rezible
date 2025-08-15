package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/predicate"
)

type IncidentService struct {
	db    *ent.Client
	jobs  rez.JobsService
	users rez.UserService
	chat  rez.ChatService
}

func NewIncidentService(db *ent.Client, jobs rez.JobsService, lms rez.LanguageModelService, chat rez.ChatService, users rez.UserService) (*IncidentService, error) {
	svc := &IncidentService{
		db:    db,
		jobs:  jobs,
		chat:  chat,
		users: users,
	}

	//	provider.SetOnIncidentUpdatedCallback(svc.onProviderIncidentUpdated)

	return svc, nil
}

func (s *IncidentService) onProviderIncidentUpdated(providerId string, updatedAt time.Time) {
	//ctx := context.Background()
	//job := incidentDataSyncJobArgs{ProviderId: providerId}
	//insertOpts := &jobs.InsertOpts{
	//	UniqueOpts: jobs.UniqueOpts{
	//		ByArgs:  true,
	//		ByState: jobs.NonCompletedJobStates,
	//	},
	//}
	//if _, insertErr := s.jobClient.Insert(ctx, job, insertOpts); insertErr != nil {
	//	log.Error().Err(insertErr).Str("providerId", providerId).Msg("failed to insert update job")
	//}
	log.Debug().Str("id", providerId).Msg("incident updated")

	// check resolved, send debrief requests
}

func (s *IncidentService) incidentQuery(pred predicate.Incident, edges bool) *ent.IncidentQuery {
	// TODO: use a view for this
	q := s.db.Incident.Query().Where(pred)
	if edges {
		q.WithRetrospective(func(rq *ent.RetrospectiveQuery) {
			rq.Select(retrospective.FieldID)
		})
		q.WithSeverity()
		q.WithType()
		q.WithFieldSelections()
		q.WithRoleAssignments(func(raq *ent.IncidentRoleAssignmentQuery) {
			raq.WithRole().WithUser()
		})
	}
	return q
}

func (s *IncidentService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Incident, error) {
	return s.incidentQuery(incident.ID(id), true).Only(ctx)
}

func (s *IncidentService) GetIdForSlug(ctx context.Context, slug string) (uuid.UUID, error) {
	return s.incidentQuery(incident.Slug(slug), false).OnlyID(ctx)
}

func (s *IncidentService) GetBySlug(ctx context.Context, slug string) (*ent.Incident, error) {
	return s.incidentQuery(incident.Slug(slug), true).Only(ctx)
}

func (s *IncidentService) GetByProviderId(ctx context.Context, pid string) (*ent.Incident, error) {
	return s.incidentQuery(incident.ProviderID(pid), true).Only(ctx)
}

func (s *IncidentService) ListIncidents(ctx context.Context, params rez.ListIncidentsParams) ([]*ent.Incident, int, error) {
	var predicates []predicate.Incident
	if !params.OpenedAfter.IsZero() {
		predicates = append(predicates, incident.OpenedAtGT(params.OpenedAfter))
	}
	if !params.OpenedBefore.IsZero() {
		predicates = append(predicates, incident.OpenedAtLT(params.OpenedBefore))
	}

	query := s.db.Incident.Query().
		Limit(params.GetLimit()).
		Offset(params.Offset)

	if len(predicates) > 0 {
		query.Where(incident.And(predicates...))
	}

	// TODO: this is probably incorrect, should lookup role assignments first
	if params.UserId != uuid.Nil {
		query.WithRoleAssignments(func(q *ent.IncidentRoleAssignmentQuery) {
			q.Where(incidentroleassignment.UserIDEQ(params.UserId))
		})
	}

	count, queryErr := query.Count(ctx)
	if queryErr != nil {
		return nil, 0, fmt.Errorf("count: %w", queryErr)
	}

	incidents := make([]*ent.Incident, 0)
	if count > 0 {
		incidents, queryErr = query.All(params.GetQueryContext(ctx))
		if queryErr != nil {
			return nil, 0, fmt.Errorf("listing incidents: %w", queryErr)
		}
	}
	return incidents, count, nil
}

func (s *IncidentService) ListIncidentRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	return s.db.IncidentRole.Query().All(ctx)
}

func (s *IncidentService) ListIncidentSeverities(ctx context.Context) ([]*ent.IncidentSeverity, error) {
	return s.db.IncidentSeverity.Query().All(ctx)
}

func (s *IncidentService) ListIncidentTypes(ctx context.Context) ([]*ent.IncidentType, error) {
	return s.db.IncidentType.Query().All(ctx)
}
