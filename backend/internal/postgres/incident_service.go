package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/predicate"
)

type IncidentService struct {
	db       *ent.Client
	jobs     rez.BackgroundJobService
	loader   rez.ProviderLoader
	provider rez.IncidentDataProvider
	chat     rez.ChatService
	users    rez.UserService
}

func NewIncidentService(ctx context.Context, db *ent.Client, jobs rez.BackgroundJobService, pl rez.ProviderLoader, ai rez.AiService, chat rez.ChatService, users rez.UserService) (*IncidentService, error) {
	svc := &IncidentService{
		db:     db,
		jobs:   jobs,
		loader: pl,
		chat:   chat,
		users:  users,
	}

	if dataErr := svc.LoadDataProvider(ctx); dataErr != nil {
		return nil, dataErr
	}

	return svc, nil
}

func (s *IncidentService) LoadDataProvider(ctx context.Context) error {
	provider, providerErr := s.loader.LoadIncidentDataProvider(ctx)
	if providerErr != nil {
		return fmt.Errorf("failed to load incident data provider: %w", providerErr)
	}
	s.provider = provider
	provider.SetOnIncidentUpdatedCallback(s.onProviderIncidentUpdated)
	return nil
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

func (s *IncidentService) SyncData(ctx context.Context) error {
	syncer := newIncidentDataSyncer(s.db, s.users, s.provider)
	return syncer.syncProviderData(ctx)
}

func (s *IncidentService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Incident, error) {
	return s.db.Incident.Get(ctx, id)
}

func (s *IncidentService) GetIdForSlug(ctx context.Context, slug string) (uuid.UUID, error) {
	return s.db.Incident.Query().Where(incident.Slug(slug)).OnlyID(ctx)
}

func (s *IncidentService) GetBySlug(ctx context.Context, slug string) (*ent.Incident, error) {
	return s.db.Incident.Query().Where(incident.Slug(slug)).Only(ctx)
}

func (s *IncidentService) GetByProviderId(ctx context.Context, id string) (*ent.Incident, error) {
	return s.db.Incident.Query().Where(incident.ProviderID(id)).Only(ctx)
}

func (s *IncidentService) ListIncidents(ctx context.Context, params rez.ListIncidentsParams) ([]*ent.Incident, error) {
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
	if params.UserId != uuid.Nil {
		query.WithRoleAssignments(func(q *ent.IncidentRoleAssignmentQuery) {
			q.Where(incidentroleassignment.UserIDEQ(params.UserId))
		})
	}
	return query.All(params.GetQueryContext(ctx))
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
