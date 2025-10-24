package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
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

// TODO: load these from somewhere
var (
	slugAdjectives = []string{
		"quick", "bright", "calm", "wise", "bold", "clear", "fair", "grand", "kind", "noble",
		"quiet", "swift", "warm", "young", "crisp", "fresh", "light", "solid", "steady", "vital",
		"active", "clever", "direct", "eager", "gentle", "honest", "lively", "modest", "polite", "prompt",
		"secure", "simple", "smooth", "stable", "strong", "subtle", "tender", "upbeat", "useful", "valid",
		"aware", "brief", "civil", "exact", "frank", "happy", "ideal", "joint", "loose", "lucky",
	}
	slugNouns = []string{
		"cloud", "river", "mountain", "forest", "ocean", "valley", "meadow", "harbor", "prairie", "canyon",
		"desert", "glacier", "island", "plateau", "summit", "delta", "fjord", "lagoon", "marsh", "oasis",
		"ridge", "stream", "tundra", "basin", "beacon", "bridge", "castle", "garden", "haven", "portal",
		"quest", "refuge", "signal", "tower", "voyage", "anchor", "compass", "horizon", "journey", "path",
		"storm", "sunrise", "tide", "wave", "wind", "crystal", "ember", "flame", "prism", "spark",
	}
)

func (s *IncidentService) generateIncidentSlug(ctx context.Context, openedAt time.Time) (string, error) {
	randgen := rand.New(rand.NewSource(openedAt.UnixNano()))
	datePrefix := openedAt.Format("20060102")
	const maxRetries = 5
	for attempt := 0; attempt < maxRetries; attempt++ {
		adj := slugAdjectives[randgen.Intn(len(slugAdjectives))]
		noun := slugNouns[randgen.Intn(len(slugNouns))]
		candidate := slug.Make(fmt.Sprintf("%s-%s-%s", datePrefix, adj, noun))

		exists, queryErr := s.db.Incident.Query().Where(incident.Slug(candidate)).Exist(ctx)
		if queryErr != nil {
			return "", fmt.Errorf("failed to check slug uniqueness: %w", queryErr)
		}
		if !exists {
			return candidate, nil
		}
	}

	// fallback - use uuid as suffix
	adj := slugAdjectives[randgen.Intn(len(slugAdjectives))]
	noun := slugNouns[randgen.Intn(len(slugNouns))]
	shortUUID := uuid.New().String()[:8]
	return slug.Make(fmt.Sprintf("%s-%s-%s-%s", datePrefix, adj, noun, shortUUID)), nil
}

type incidentMutator interface {
	Save(ctx context.Context) (*ent.Incident, error)
	Mutation() *ent.IncidentMutation
}

func (s *IncidentService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMutation)) (*ent.Incident, error) {
	var mutator incidentMutator
	isNew := id == uuid.Nil
	if isNew {
		mutator = s.db.Incident.Create().SetID(uuid.New())
	} else {
		curr, getErr := s.db.Incident.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("fetch existing incident: %w", getErr)
		}
		mutator = s.db.Incident.UpdateOne(curr)
	}

	mut := mutator.Mutation()
	setFn(mut)

	if isNew {
		openedAt := time.Now()
		if at, exists := mut.OpenedAt(); exists {
			openedAt = at
		}
		generatedSlug, slugErr := s.generateIncidentSlug(ctx, openedAt)
		if slugErr != nil {
			return nil, fmt.Errorf("generate slug: %w", slugErr)
		}
		mut.SetSlug(generatedSlug)
	}

	updated, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save incident: %w", saveErr)
	}

	// TODO: notify services on incident update

	return updated, nil
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
