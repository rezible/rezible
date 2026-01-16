package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	ira "github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
)

type IncidentService struct {
	db    *ent.Client
	jobs  rez.JobsService
	msgs  rez.MessageService
	users rez.UserService
}

func NewIncidentService(db *ent.Client, jobs rez.JobsService, msgs rez.MessageService, users rez.UserService) (*IncidentService, error) {
	svc := &IncidentService{
		db:    db,
		jobs:  jobs,
		msgs:  msgs,
		users: users,
	}

	if msgsErr := svc.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("failed registering message handlers: %w", msgsErr)
	}

	return svc, nil
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
			q.Where(ira.UserID(params.UserId))
		})
	}

	return ent.DoListQuery[*ent.Incident, *ent.IncidentQuery](ctx, query, params.ListParams)
}

func (s *IncidentService) Get(ctx context.Context, id uuid.UUID) (*ent.Incident, error) {
	return s.incidentQuery(incident.ID(id), true).Only(ctx)
}

func (s *IncidentService) GetByChatChannelID(ctx context.Context, id string) (*ent.Incident, error) {
	return s.incidentQuery(incident.ChatChannelID(id), false).Only(ctx)
}

func (s *IncidentService) GetBySlug(ctx context.Context, slug string) (*ent.Incident, error) {
	return s.incidentQuery(incident.Slug(slug), true).Only(ctx)
}

func (s *IncidentService) getIncidentEdgeMutationUpdateEvent(incidentId uuid.UUID, m ent.Mutation, v ent.Value) any {
	op := m.Op()
	isCreate := op.Is(ent.OpCreate)
	isUpdate := op.Is(ent.OpUpdateOne)
	if !(isCreate || isUpdate) {
		return nil
	}
	if m.Type() == ent.TypeIncidentMilestone {
		ms, ok := v.(*ent.IncidentMilestone)
		if !ok {
			log.Warn().Interface("v", v).Msg("failed to cast value to ent.IncidentMilestone")
			return nil
		}
		return &rez.EventOnIncidentMilestoneUpdated{
			IncidentId:  incidentId,
			MilestoneId: ms.ID,
			Created:     m.Op().Is(ent.OpCreate),
		}
	}
	log.Debug().
		Str("op", op.String()).
		Str("type", m.Type()).
		Msg("maybe add update event")
	return nil
}

func (s *IncidentService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMutation) []ent.Mutation) (*ent.Incident, error) {
	var curr *ent.Incident
	if id != uuid.Nil {
		inc, getErr := s.db.Incident.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("fetch existing incident: %w", getErr)
		}
		curr = inc
	}
	isCreate := id == uuid.Nil && curr == nil

	var generatedUniqueSlug string
	if curr == nil {
		m := s.db.Incident.Create().Mutation()
		setFn(m)
		openedAt := time.Now()
		if at, exists := m.OpenedAt(); exists {
			openedAt = at
		}
		incSlug, slugErr := s.generateIncidentSlug(ctx, openedAt)
		if slugErr != nil {
			return nil, fmt.Errorf("generate unique slug: %w", slugErr)
		}
		generatedUniqueSlug = incSlug
	}

	var updateEvents []any
	var updated *ent.Incident
	updateTx := func(tx *ent.Tx) error {
		var mutator ent.EntityMutator[*ent.Incident, *ent.IncidentMutation]
		if curr == nil {
			mutator = tx.Incident.Create().SetID(uuid.New())
		} else {
			mutator = tx.Incident.UpdateOne(curr)
		}

		incidentMut := mutator.Mutation()
		edgeMuts := setFn(incidentMut)
		if generatedUniqueSlug != "" {
			incidentMut.SetSlug(generatedUniqueSlug)
		}

		var saveErr error
		updated, saveErr = mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save incident: %w", saveErr)
		}
		incEvent := rez.EventOnIncidentUpdated{
			Created:    isCreate,
			IncidentId: updated.ID,
		}
		updateEvents = append(updateEvents, incEvent)

		for _, edgeMut := range edgeMuts {
			v, edgeErr := tx.Client().Mutate(ctx, edgeMut)
			if edgeErr != nil {
				return fmt.Errorf("edge mutation: %w", edgeErr)
			}
			updateEvents = append(updateEvents, s.getIncidentEdgeMutationUpdateEvent(updated.ID, edgeMut, v))
		}

		return nil
	}
	if txErr := ent.WithTx(ctx, s.db, updateTx); txErr != nil {
		return nil, fmt.Errorf("update: %w", txErr)
	}

	for _, ev := range updateEvents {
		if pubEvErr := s.msgs.PublishEvent(ctx, ev); pubEvErr != nil {
			log.Error().Err(pubEvErr).Msg("failed to publish incident update event message")
		}
	}

	return updated, nil
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
	datePrefix := openedAt.Format("060102")
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
	uuidSlug := slug.Make(fmt.Sprintf("%s-%s-%s-%s", datePrefix, adj, noun, shortUUID))
	log.Warn().Str("slug", uuidSlug).Msg("falling back to uuid incident slug")
	return uuidSlug, nil
}

func (s *IncidentService) ListIncidentRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	return s.db.IncidentRole.Query().All(ctx)
}

func (s *IncidentService) ListIncidentSeverities(ctx context.Context) ([]*ent.IncidentSeverity, error) {
	return s.db.IncidentSeverity.Query().All(ctx)
}

func (s *IncidentService) GetIncidentSeverity(ctx context.Context, id uuid.UUID) (*ent.IncidentSeverity, error) {
	return s.db.IncidentSeverity.Get(ctx, id)
}

func (s *IncidentService) GetIncidentMilestone(ctx context.Context, id uuid.UUID) (*ent.IncidentMilestone, error) {
	query := s.db.IncidentMilestone.Query().
		Where(im.ID(id)).
		WithUser()
	return query.Only(ctx)
}

func (s *IncidentService) SetIncidentMilestone(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMilestoneMutation)) (*ent.IncidentMilestone, error) {
	var curr *ent.IncidentMilestone
	if id != uuid.Nil {
		inc, getErr := s.db.IncidentMilestone.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("fetch existing incident: %w", getErr)
		}
		curr = inc
	}

	var updated *ent.IncidentMilestone
	updateTx := func(tx *ent.Tx) error {
		var mutator ent.EntityMutator[*ent.IncidentMilestone, *ent.IncidentMilestoneMutation]
		if curr == nil {
			mutator = tx.IncidentMilestone.Create().SetID(uuid.New())
		} else {
			mutator = tx.IncidentMilestone.UpdateOne(curr)
		}

		mut := mutator.Mutation()
		setFn(mut)

		var saveErr error
		updated, saveErr = mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("save: %w", saveErr)
		}

		return nil
	}

	if txErr := ent.WithTx(ctx, s.db, updateTx); txErr != nil {
		return nil, fmt.Errorf("update: %w", txErr)
	}

	ev := &rez.EventOnIncidentMilestoneUpdated{
		Created:     id == uuid.Nil,
		MilestoneId: updated.ID,
		IncidentId:  updated.IncidentID,
	}
	if pubErr := s.msgs.PublishEvent(ctx, ev); pubErr != nil {
		log.Error().Err(pubErr).Msg("failed to publish incident milestone updated message")
	}

	return updated, nil
}

func (s *IncidentService) ListIncidentTypes(ctx context.Context) ([]*ent.IncidentType, error) {
	return s.db.IncidentType.Query().All(ctx)
}

func (s *IncidentService) ListIncidentFields(ctx context.Context) ([]*ent.IncidentField, error) {
	return s.db.IncidentField.Query().All(ctx)
}

func (s *IncidentService) GetIncidentMetadata(ctx context.Context) (*rez.IncidentMetadata, error) {
	md := rez.IncidentMetadata{}
	var err error

	// TODO: use a view or get in parallel

	md.Roles, err = s.ListIncidentRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("roles: %w", err)
	}

	md.Types, err = s.ListIncidentTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("types: %w", err)
	}

	md.Severities, err = s.ListIncidentSeverities(ctx)
	if err != nil {
		return nil, fmt.Errorf("severities: %w", err)
	}

	md.Fields, err = s.ListIncidentFields(ctx)
	if err != nil {
		return nil, fmt.Errorf("fields: %w", err)
	}

	return &md, nil
}
