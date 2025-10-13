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
}

func NewIncidentService(db *ent.Client, jobs rez.JobsService, users rez.UserService) (*IncidentService, error) {
	svc := &IncidentService{
		db:    db,
		jobs:  jobs,
		users: users,
	}

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

func (s *IncidentService) Get(ctx context.Context, id uuid.UUID) (*ent.Incident, error) {
	return s.incidentQuery(incident.ID(id), true).Only(ctx)
}

func (s *IncidentService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMutation)) (*ent.Incident, error) {
	if id != uuid.Nil {
		curr, getErr := s.db.Incident.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("failed to fetch incident: %w", getErr)
		}
		update := s.db.Incident.UpdateOne(curr)
		setFn(update.Mutation())
		return update.Save(ctx)
	}

	create := s.db.Incident.Create().SetID(uuid.New())
	mut := create.Mutation()
	setFn(mut)

	if title, ok := mut.Title(); ok {
		// TODO: generate slug
		create.SetSlug(title)
	}

	return create.Save(ctx)
}

func (s *IncidentService) GetByChatChannelID(ctx context.Context, id string) (*ent.Incident, error) {
	return s.incidentQuery(incident.ChatChannelID(id), false).Only(ctx)
}

func (s *IncidentService) GetBySlug(ctx context.Context, slug string) (*ent.Incident, error) {
	return s.incidentQuery(incident.Slug(slug), true).Only(ctx)
}

func (s *IncidentService) GetByProviderId(ctx context.Context, pid string) (*ent.Incident, error) {
	return s.incidentQuery(incident.ProviderID(pid), true).Only(ctx)
}

func (s *IncidentService) ListIncidents(ctx context.Context, params rez.ListIncidentsParams) (*ent.ListResult[*ent.Incident], error) {
	query := s.db.Incident.Query()
	query.Order(incident.ByOpenedAt(params.GetOrder()))
	if !params.OpenedAfter.IsZero() {
		query.Where(incident.OpenedAtGT(params.OpenedAfter))
	}
	if !params.OpenedBefore.IsZero() {
		query.Where(incident.OpenedAtLT(params.OpenedBefore))
	}

	// TODO: this is probably incorrect, should lookup role assignments first
	if params.UserId != uuid.Nil {
		query.WithRoleAssignments(func(q *ent.IncidentRoleAssignmentQuery) {
			q.Where(incidentroleassignment.UserIDEQ(params.UserId))
		})
	}

	return ent.DoListQuery[*ent.Incident, *ent.IncidentQuery](ctx, query, params.ListParams)
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

func (s *IncidentService) ListIncidentFields(context.Context) (ent.IncidentEdges, error) {
	return ent.IncidentEdges{}, nil
}
