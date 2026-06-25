package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	ii "github.com/rezible/rezible/ent/incidentimpact"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	imodel "github.com/rezible/rezible/ent/incidentmilestone"
	ira "github.com/rezible/rezible/ent/incidentroleassignment"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
)

type IncidentService struct {
	db        rez.Database
	msgs      rez.MessageService
	knowledge rez.KnowledgeService
}

func NewIncidentService(db rez.Database, msgs rez.MessageService, knowledge rez.KnowledgeService) (*IncidentService, error) {
	svc := &IncidentService{
		db:        db,
		msgs:      msgs,
		knowledge: knowledge,
	}

	if msgsErr := svc.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("failed registering message handlers: %w", msgsErr)
	}

	return svc, nil
}

func (s *IncidentService) registerMessageHandlers() error {
	eventsErr := s.msgs.AddEventHandlers(
		rez.NewEventHandler("db.IncidentService.OnIncidentUpdate", s.onIncidentUpdate))
	return errors.Join(eventsErr)
}

func (s *IncidentService) onIncidentUpdate(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	//msQuery := s.db.IncidentMilestone.Query().
	//	Where(incidentmilestone.IncidentID(ev.IncidentId))
	//milestones, msErr := msQuery.All(ctx)
	//if msErr != nil {
	//	return fmt.Errorf("incident milestone query: %w", msErr)
	//}
	//for _, m := range milestones {
	//	slog.Debug("Incident milestone", "milestone", m.String())
	//}
	return nil
}

func (s *IncidentService) allQueryEdges(q *ent.IncidentQuery) {
	q.WithRetrospective(func(rq *ent.RetrospectiveQuery) {
		rq.Select(retrospective.FieldID)
	})
	q.WithSeverity()
	q.WithType()
	q.WithFieldSelections(func(foq *ent.IncidentFieldOptionQuery) {
		foq.WithIncidentField()
	})
	q.WithTagAssignments()
	q.WithRoleAssignments(func(raq *ent.IncidentRoleAssignmentQuery) {
		raq.WithRole().WithUser()
	})
	q.WithMilestones(func(mq *ent.IncidentMilestoneQuery) {
		mq.Order(ent.Desc(imodel.FieldTimestamp))
		mq.WithUser()
	})
	q.WithVideoConferences()
}

func (s *IncidentService) incidentQuery(ctx context.Context, pred predicate.Incident, edgesFn func(*ent.IncidentQuery)) *ent.IncidentQuery {
	// TODO: use a view for this
	q := s.db.Client(ctx).Incident.Query().Where(pred)
	edgesFn(q)
	return q
}

func (s *IncidentService) ListIncidents(ctx context.Context, params rez.ListIncidentsParams) (*ent.ListResult[ent.Incident], error) {
	query := s.db.Client(ctx).Incident.Query()
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
	s.allQueryEdges(query)

	return ent.DoListQuery[ent.Incident, *ent.IncidentQuery](ctx, query, params.ListParams)
}

func (s *IncidentService) Query(ctx context.Context, p predicate.Incident, withFn func(*ent.IncidentQuery)) (*ent.Incident, error) {
	return s.incidentQuery(ctx, p, withFn).Only(ctx)
}

func (s *IncidentService) Get(ctx context.Context, p predicate.Incident) (*ent.Incident, error) {
	return s.incidentQuery(ctx, p, s.allQueryEdges).Only(ctx)
}

func (s *IncidentService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMutation)) (*ent.Incident, error) {
	var curr *ent.Incident
	client := s.db.Client(ctx)
	if id != uuid.Nil {
		inc, getErr := client.Incident.Get(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("fetch existing incident: %w", getErr)
		}
		curr = inc
	}
	isCreate := id == uuid.Nil && curr == nil

	var generatedUniqueSlug string
	if curr == nil {
		m := client.Incident.Create().Mutation()
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

	var mutator ent.EntityMutator[*ent.Incident, *ent.IncidentMutation]
	if curr == nil {
		mutator = client.Incident.Create().SetID(uuid.New())
	} else {
		mutator = client.Incident.UpdateOne(curr)
	}
	incidentMut := mutator.Mutation()
	setFn(incidentMut)
	if generatedUniqueSlug != "" {
		incidentMut.SetSlug(generatedUniqueSlug)
	}
	updated, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save incident: %w", saveErr)
	}

	updatedEvent := rez.EventOnIncidentUpdated{
		Created:    isCreate,
		IncidentId: updated.ID,
	}
	if pubEvErr := s.msgs.PublishEvent(ctx, updatedEvent); pubEvErr != nil {
		slog.Error("failed to publish incident update event message", "error", pubEvErr)
	}
	return s.Get(ctx, incident.ID(updated.ID))
}

func (s *IncidentService) SetIncidentMilestone(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMilestoneMutation)) (*ent.IncidentMilestone, error) {
	var mutator ent.EntityMutator[*ent.IncidentMilestone, *ent.IncidentMilestoneMutation]
	client := s.db.Client(ctx)
	if id == uuid.Nil {
		mutator = client.IncidentMilestone.Create().SetID(uuid.New())
	} else {
		mutator = client.IncidentMilestone.UpdateOneID(id)
	}
	mut := mutator.Mutation()
	setFn(mut)
	updated, saveErr := mutator.Save(ctx)
	if saveErr != nil {
		return nil, fmt.Errorf("save incident: %w", saveErr)
	}

	updatedEvent := rez.EventOnIncidentMilestoneUpdated{
		IncidentId:  updated.IncidentID,
		MilestoneId: updated.ID,
		Created:     id == uuid.Nil,
	}
	if pubEvErr := s.msgs.PublishEvent(ctx, updatedEvent); pubEvErr != nil {
		slog.Error("failed to publish incident milestone update event message", "error", pubEvErr)
	}
	return updated, nil
}

func (s *IncidentService) Archive(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).Incident.DeleteOneID(id).Exec(ctx)
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

		exists, queryErr := s.db.Client(ctx).Incident.Query().Where(incident.Slug(candidate)).Exist(ctx)
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
	slog.Warn("falling back to uuid incident slug", "slug", uuidSlug)
	return uuidSlug, nil
}

func (s *IncidentService) ListIncidentRoles(ctx context.Context) ([]*ent.IncidentRole, error) {
	return s.db.Client(ctx).IncidentRole.Query().All(ctx)
}

func (s *IncidentService) ListIncidentSeverities(ctx context.Context) ([]*ent.IncidentSeverity, error) {
	return s.db.Client(ctx).IncidentSeverity.Query().All(ctx)
}

func (s *IncidentService) GetIncidentSeverity(ctx context.Context, id uuid.UUID) (*ent.IncidentSeverity, error) {
	return s.db.Client(ctx).IncidentSeverity.Get(ctx, id)
}

func (s *IncidentService) GetIncidentMilestone(ctx context.Context, id uuid.UUID) (*ent.IncidentMilestone, error) {
	query := s.db.Client(ctx).IncidentMilestone.Query().
		Where(im.ID(id)).
		WithUser()
	return query.Only(ctx)
}

func (s *IncidentService) ListIncidentTypes(ctx context.Context) ([]*ent.IncidentType, error) {
	return s.db.Client(ctx).IncidentType.Query().All(ctx)
}

func (s *IncidentService) ListIncidentFields(ctx context.Context) ([]*ent.IncidentField, error) {
	return s.db.Client(ctx).IncidentField.Query().
		WithOptions().
		All(ctx)
}

func (s *IncidentService) ListIncidentTags(ctx context.Context) ([]*ent.IncidentTag, error) {
	return s.db.Client(ctx).IncidentTag.Query().All(ctx)
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

	md.Tags, err = s.ListIncidentTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("tags: %w", err)
	}

	return &md, nil
}

func (s *IncidentService) ListIncidentImpacts(ctx context.Context, incidentID uuid.UUID) ([]*ent.IncidentImpact, error) {
	impacts, queryErr := s.db.Client(ctx).IncidentImpact.Query().
		Where(ii.IncidentID(incidentID)).
		WithKnowledgeEntity().
		Order(ii.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query incident impacts: %w", queryErr)
	}
	return impacts, nil
}

// TODO: fix this slop
func (s *IncidentService) SetIncidentImpacts(ctx context.Context, id uuid.UUID, input []rez.IncidentImpactInput) ([]*ent.IncidentImpact, error) {
	var impacts []*ent.IncidentImpact
	txErr := s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		if _, getErr := tx.Incident.Get(txCtx, id); getErr != nil {
			return fmt.Errorf("get incident: %w", getErr)
		}

		entityIDs := make([]uuid.UUID, 0, len(input))
		for _, inputImpact := range input {
			entityID, resolveErr := s.resolveIncidentImpactEntity(txCtx, tx, inputImpact)
			if resolveErr != nil {
				return resolveErr
			}
			entityIDs = append(entityIDs, entityID)

			existing, queryErr := tx.IncidentImpact.Query().
				Where(ii.IncidentID(id), ii.KnowledgeEntityID(entityID)).
				Only(txCtx)
			if queryErr != nil && !ent.IsNotFound(queryErr) {
				return fmt.Errorf("query existing impact: %w", queryErr)
			}

			if existing == nil {
				create := tx.IncidentImpact.Create().
					SetIncidentID(id).
					SetKnowledgeEntityID(entityID)
				if inputImpact.Source != "" {
					create.SetSource(inputImpact.Source)
				}
				if inputImpact.Note != "" {
					create.SetNote(inputImpact.Note)
				}
				if _, createErr := create.Save(txCtx); createErr != nil {
					return fmt.Errorf("create incident impact: %w", createErr)
				}
				continue
			}

			update := tx.IncidentImpact.UpdateOne(existing)
			if inputImpact.Source != "" {
				update.SetSource(inputImpact.Source)
			} else {
				update.ClearSource()
			}
			if inputImpact.Note != "" {
				update.SetNote(inputImpact.Note)
			} else {
				update.ClearNote()
			}
			if _, updateErr := update.Save(txCtx); updateErr != nil {
				return fmt.Errorf("update incident impact: %w", updateErr)
			}
		}

		deleteQuery := tx.IncidentImpact.Delete().Where(ii.IncidentID(id))
		if len(entityIDs) > 0 {
			deleteQuery.Where(ii.KnowledgeEntityIDNotIn(entityIDs...))
		}
		if _, deleteErr := deleteQuery.Exec(txCtx); deleteErr != nil {
			return fmt.Errorf("delete replaced incident impacts: %w", deleteErr)
		}

		var listErr error
		impacts, listErr = tx.IncidentImpact.Query().
			Where(ii.IncidentID(id)).
			WithKnowledgeEntity().
			Order(ii.ByCreatedAt(sql.OrderAsc())).
			All(txCtx)
		if listErr != nil {
			return fmt.Errorf("reload impacts: %w", listErr)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	if pubErr := s.msgs.PublishEvent(ctx, rez.EventOnIncidentImpactsUpdated{IncidentId: id}); pubErr != nil {
		slog.WarnContext(ctx, "failed to publish incident impacts update event", "error", pubErr)
	}
	return impacts, nil
}

func (s *IncidentService) resolveIncidentImpactEntity(ctx context.Context, tx *ent.Client, input rez.IncidentImpactInput) (uuid.UUID, error) {
	if input.KnowledgeEntityID != uuid.Nil {
		if _, getErr := tx.KnowledgeEntity.Get(ctx, input.KnowledgeEntityID); getErr != nil {
			return uuid.Nil, fmt.Errorf("get impact knowledge entity: %w", getErr)
		}
		return input.KnowledgeEntityID, nil
	}
	if input.Kind == "" || input.DisplayName == "" {
		return uuid.Nil, fmt.Errorf("impact requires knowledgeEntityId or kind/displayName")
	}

	existing, queryErr := tx.KnowledgeEntity.Query().
		Where(kne.Kind(input.Kind), kne.DisplayName(input.DisplayName)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query impact knowledge entity: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	create := tx.KnowledgeEntity.Create().
		SetKind(input.Kind).
		SetDisplayName(input.DisplayName).
		SetFirstObservedAt(time.Now().UTC()).
		SetLastObservedAt(time.Now().UTC())
	if input.Description != "" {
		create.SetDescription(input.Description)
	}
	created, createErr := create.Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create impact knowledge entity: %w", createErr)
	}
	return created.ID, nil
}
