package db

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	imodel "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/incidentrole"
	ira "github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/pkg/errs"
)

type IncidentService struct {
	db             rez.Database
	msgs           rez.MessageQueue
	situations     rez.SituationService
	retrospectives rez.RetrospectiveService
}

func NewIncidentService(db rez.Database, msgs rez.MessageQueue, situations rez.SituationService, retrospectives rez.RetrospectiveService) (*IncidentService, error) {
	svc := &IncidentService{
		db:             db,
		msgs:           msgs,
		situations:     situations,
		retrospectives: retrospectives,
	}
	return svc, nil
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
		mq.Order(imodel.ByTimestamp(sql.OrderAsc()), imodel.ByID(sql.OrderAsc()))
		mq.WithUser()
	})
	q.WithVideoConferences()
	q.WithSituations()
}

func (s *IncidentService) incidentQuery(ctx context.Context, pred predicate.Incident, edgesFn func(*ent.IncidentQuery)) *ent.IncidentQuery {
	// TODO: use a view for this
	q := s.db.Client(ctx).Incident.Query().Where(pred)
	edgesFn(q)
	return q
}

func (s *IncidentService) ListIncidents(ctx context.Context, params rez.ListIncidentsParams) (*ent.ListResult[ent.Incident], error) {
	query := s.db.Client(ctx).Incident.Query()
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(incident.Or(incident.TitleContainsFold(search), incident.SummaryContainsFold(search), incident.SlugContainsFold(search)))
	}
	if params.SeverityId != uuid.Nil {
		query.Where(incident.SeverityID(params.SeverityId))
	}
	if len(params.ResponseStates) > 0 {
		query.Where(incident.ResponseStateIn(params.ResponseStates...))
	}
	if !params.OpenedAfter.IsZero() {
		query.Where(incident.OpenedAtGT(params.OpenedAfter))
	}
	if !params.OpenedBefore.IsZero() {
		query.Where(incident.OpenedAtLT(params.OpenedBefore))
	}

	if params.UserId != uuid.Nil {
		preds := []predicate.IncidentRoleAssignment{
			ira.HasRoleWith(incidentrole.ArchiveTimeIsNil()),
			ira.UserID(params.UserId),
		}
		query.Where(incident.HasRoleAssignmentsWith(preds...))
		query.WithRoleAssignments(func(q *ent.IncidentRoleAssignmentQuery) {
			q.Where(preds...).WithRole().WithUser()
		})
	}
	s.allQueryEdges(query)

	query.Order(
		incident.BySeverityField(incidentseverity.FieldRank, sql.OrderAsc()),
		incident.ByUpdatedAt(sql.OrderDesc()),
		incident.ByID(sql.OrderAsc()),
	)

	return ent.DoListQuery[ent.Incident, *ent.IncidentQuery](ctx, query, params.ListParams)
}

func (s *IncidentService) Query(ctx context.Context, p predicate.Incident, withFn func(*ent.IncidentQuery)) (*ent.Incident, error) {
	return s.incidentQuery(ctx, p, withFn).Only(ctx)
}

func (s *IncidentService) Get(ctx context.Context, p predicate.Incident) (*ent.Incident, error) {
	return s.incidentQuery(ctx, p, s.allQueryEdges).Only(ctx)
}

func (s *IncidentService) Set(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMutation)) (*ent.Incident, error) {
	isCreate := id == uuid.Nil
	updatedID, txErr := ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (uuid.UUID, error) {
		var mutator ent.EntityMutator[*ent.Incident, *ent.IncidentMutation]
		if isCreate {
			mutator = tx.Incident.Create().SetID(uuid.New())
		} else {
			mutator = tx.Incident.UpdateOneID(id)
		}
		mut := mutator.Mutation()
		setFn(mut)
		// Situation links are made by the situation service, which raises the situations it links.
		links, linksErr := s.takeSituationLinkChanges(ctx, id, mut)
		if linksErr != nil {
			return uuid.Nil, linksErr
		}
		if isCreate {
			openedAt := time.Now()
			if at, exists := mut.OpenedAt(); exists {
				openedAt = at
			}
			incSlug, slugErr := s.generateIncidentSlug(ctx, openedAt)
			if slugErr != nil {
				return uuid.Nil, fmt.Errorf("generate unique slug: %w", slugErr)
			}
			mut.SetSlug(incSlug)
		}
		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return uuid.Nil, fmt.Errorf("save incident: %w", saveErr)
		}
		if updated.ResponseState == incident.ResponseStateResolved {
			if _, createErr := s.retrospectives.CreateForIncident(ctx, updated.ID); createErr != nil {
				return uuid.Nil, fmt.Errorf("create incident retrospective: %w", createErr)
			}
		}
		// Every write is synced, so a change such as resolution reaches the incident's situations.
		if sitErr := s.situations.SyncIncidentLinks(ctx, updated.ID, links); sitErr != nil {
			return uuid.Nil, fmt.Errorf("sync incident situations: %w", sitErr)
		}
		if publishErr := s.msgs.Publish(ctx, rez.EventOnIncidentUpdated{Created: isCreate, IncidentId: updated.ID}); publishErr != nil {
			return uuid.Nil, fmt.Errorf("publish incident update: %w", publishErr)
		}
		return updated.ID, nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return s.Get(ctx, incident.ID(updatedID))
}

// takeSituationLinkChanges removes the situation link changes from the mutation before it is saved. Clearing
// the incident's situations becomes removing every situation it links now.
func (s *IncidentService) takeSituationLinkChanges(ctx context.Context, id uuid.UUID, mut *ent.IncidentMutation) (rez.IncidentSituationLinkChanges, error) {
	links := rez.IncidentSituationLinkChanges{
		Added:   mut.SituationsIDs(),
		Removed: mut.RemovedSituationsIDs(),
	}
	if mut.SituationsCleared() && id != uuid.Nil {
		queryLinked := s.db.Client(ctx).Situation.Query().
			Where(situation.HasIncidentsWith(incident.ID(id)))
		linkedIDs, queryErr := queryLinked.IDs(ctx)
		if queryErr != nil {
			return links, fmt.Errorf("query incident situations: %w", queryErr)
		}
		links.Removed = append(links.Removed, linkedIDs...)
	}
	mut.ResetSituations()
	return links, nil
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
	datePrefix := openedAt.Format("060102")
	const maxRetries = 3
	for range maxRetries {
		adj := slugAdjectives[rand.Intn(len(slugAdjectives))]
		noun := slugNouns[rand.Intn(len(slugNouns))]
		candidate := slug.Make(fmt.Sprintf("%s-%s-%s", datePrefix, adj, noun))

		querySlug := s.db.Client(ctx).Incident.Query().
			Where(incident.Slug(candidate))
		exists, queryErr := querySlug.Exist(ctx)
		if !exists {
			if queryErr != nil {
				return "", fmt.Errorf("failed to check slug uniqueness: %w", queryErr)
			}
			return candidate, nil
		}
	}

	// fallback - use uuid as suffix
	adj := slugAdjectives[rand.Intn(len(slugAdjectives))]
	noun := slugNouns[rand.Intn(len(slugNouns))]
	shortUUID := uuid.New().String()[:8]
	uuidSlug := slug.Make(fmt.Sprintf("%s-%s-%s-%s", datePrefix, adj, noun, shortUUID))
	slog.Warn("falling back to uuid incident slug", "slug", uuidSlug)
	return uuidSlug, nil
}

func (s *IncidentService) ListIncidentRoles(ctx context.Context) (ent.IncidentRoles, error) {
	return s.db.Client(ctx).IncidentRole.Query().Where(incidentrole.ArchiveTimeIsNil()).All(ctx)
}

func (s *IncidentService) ListIncidentSeverities(ctx context.Context) (ent.IncidentSeverities, error) {
	return s.db.Client(ctx).IncidentSeverity.Query().All(ctx)
}

func (s *IncidentService) GetIncidentSeverity(ctx context.Context, id uuid.UUID) (*ent.IncidentSeverity, error) {
	return s.db.Client(ctx).IncidentSeverity.Get(ctx, id)
}

func (s *IncidentService) GetIncidentRoleAssignment(ctx context.Context, id uuid.UUID) (*ent.IncidentRoleAssignment, error) {
	return s.db.Client(ctx).IncidentRoleAssignment.Query().
		Where(ira.ID(id)).
		WithRole().
		WithUser().
		Only(ctx)
}

func (s *IncidentService) SetIncidentRoleAssignment(ctx context.Context, id uuid.UUID, params rez.SetIncidentRoleAssignmentParams) (*ent.IncidentRoleAssignment, error) {
	client := s.db.Client(ctx)
	var mut ent.EntityMutator[*ent.IncidentRoleAssignment, *ent.IncidentRoleAssignmentMutation]
	if id != uuid.Nil {
		if params.UserID == uuid.Nil && params.RoleID == uuid.Nil {
			return nil, fmt.Errorf("%w: assignment requires a user or role", errs.ErrInvalidInput)
		}

		u := client.IncidentRoleAssignment.UpdateOneID(id)
		if params.UserID != uuid.Nil {
			u.SetUserID(params.UserID)
		} else if params.RoleID != uuid.Nil {
			u.SetRoleID(params.RoleID)
		}
		mut = u
	} else {
		if params.IncidentID == uuid.Nil || params.RoleID == uuid.Nil || params.UserID == uuid.Nil {
			return nil, fmt.Errorf("%w: missing ids", errs.ErrInvalidInput)
		}

		createRoleAssignment := client.IncidentRoleAssignment.Create().
			SetIncidentID(params.IncidentID).
			SetUserID(params.UserID).
			SetRoleID(params.RoleID)
		created, createErr := createRoleAssignment.Save(ctx)
		if createErr != nil {
			return nil, fmt.Errorf("create incident role assignment: %w", createErr)
		}
		return s.GetIncidentRoleAssignment(ctx, created.ID)
	}
	if updateErr := mut.Exec(ctx); updateErr != nil {
		return nil, fmt.Errorf("update incident role assignment: %w", updateErr)
	}

	return s.GetIncidentRoleAssignment(ctx, id)
}

func (s *IncidentService) DeleteIncidentRoleAssignment(ctx context.Context, id uuid.UUID) error {
	return s.db.Client(ctx).IncidentRoleAssignment.DeleteOneID(id).Exec(ctx)
}

func (s *IncidentService) GetIncidentMilestone(ctx context.Context, id uuid.UUID) (*ent.IncidentMilestone, error) {
	return s.db.Client(ctx).IncidentMilestone.Query().
		Where(im.ID(id)).
		WithUser().
		Only(ctx)
}

func (s *IncidentService) ListMilestonesForIncident(ctx context.Context, incId uuid.UUID) (ent.IncidentMilestones, error) {
	return s.db.Client(ctx).IncidentMilestone.Query().
		Where(im.IncidentID(incId)).
		Order(im.ByTimestamp(), im.ByID()).
		WithUser().
		All(ctx)
}

func (s *IncidentService) SetIncidentMilestone(ctx context.Context, id uuid.UUID, setFn func(*ent.IncidentMilestoneMutation)) (*ent.IncidentMilestone, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (*ent.IncidentMilestone, error) {
		var mutator ent.EntityMutator[*ent.IncidentMilestone, *ent.IncidentMilestoneMutation]
		if id == uuid.Nil {
			mutator = tx.IncidentMilestone.Create().SetID(uuid.New())
		} else {
			mutator = tx.IncidentMilestone.UpdateOneID(id)
		}
		mut := mutator.Mutation()
		setFn(mut)

		updated, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return nil, fmt.Errorf("save incident milestone: %w", saveErr)
		}

		updatedEvent := rez.EventOnIncidentMilestoneUpdated{
			IncidentId:  updated.IncidentID,
			MilestoneId: updated.ID,
			Created:     id == uuid.Nil,
		}
		if publishErr := s.msgs.Publish(ctx, updatedEvent); publishErr != nil {
			return nil, fmt.Errorf("publish incident milestone update: %w", publishErr)
		}
		return updated, nil
	})
}

func (s *IncidentService) ListIncidentTypes(ctx context.Context) (ent.IncidentTypes, error) {
	return s.db.Client(ctx).IncidentType.Query().All(ctx)
}

func (s *IncidentService) ListIncidentFields(ctx context.Context) (ent.IncidentFields, error) {
	return s.db.Client(ctx).IncidentField.Query().
		WithOptions().
		All(ctx)
}

func (s *IncidentService) ListIncidentTags(ctx context.Context) (ent.IncidentTags, error) {
	return s.db.Client(ctx).IncidentTag.Query().
		All(ctx)
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
