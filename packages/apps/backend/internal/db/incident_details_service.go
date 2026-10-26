package db

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/incidentrole"
	ira "github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/schema"
	"github.com/rezible/rezible/ent/sequencecounter"
	"github.com/rezible/rezible/ent/situation"
)

// IncidentDetailsService reads incidents and makes the writes that do not depend on incident management being
// on: details, situation links, archiving, the response channel and provider observations.
type IncidentDetailsService struct {
	db         rez.Database
	msgs       rez.MessageQueue
	situations rez.SituationService
}

func NewIncidentDetailsService(db rez.Database, msgs rez.MessageQueue, situations rez.SituationService) (*IncidentDetailsService, error) {
	return &IncidentDetailsService{db: db, msgs: msgs, situations: situations}, nil
}

func (s *IncidentDetailsService) allQueryEdges(q *ent.IncidentQuery) {
	q.WithRetrospective(func(rq *ent.RetrospectiveQuery) {
		rq.Select(retrospective.FieldID)
	})
	q.WithSeverity()
	q.WithType()
	q.WithDeclaredBy()
	q.WithFieldSelections(func(foq *ent.IncidentFieldOptionQuery) {
		foq.WithIncidentField()
	})
	q.WithTagAssignments()
	q.WithRoleAssignments(func(raq *ent.IncidentRoleAssignmentQuery) {
		raq.WithRole().WithUser()
	})
	q.WithMilestones(func(mq *ent.IncidentMilestoneQuery) {
		mq.Order(im.ByTimestamp(sql.OrderAsc()), im.ByID(sql.OrderAsc()))
		mq.WithUser()
	})
	q.WithVideoConferences()
	q.WithSituations()
}

func (s *IncidentDetailsService) ListIncidents(ctx context.Context, params rez.ListIncidentsParams) (*ent.ListResult[ent.Incident], error) {
	query := s.db.Client(ctx).Incident.Query()
	if search := strings.TrimSpace(params.Search); search != "" {
		matches := []predicate.Incident{incident.TitleContainsFold(search), incident.SummaryContainsFold(search)}
		if number, isReference := ent.ParseIncidentReference(search); isReference {
			matches = append(matches, incident.Number(number))
		}
		query.Where(incident.Or(matches...))
	}
	if params.SeverityId != uuid.Nil {
		query.Where(incident.SeverityID(params.SeverityId))
	}
	if params.Resolved != nil {
		if *params.Resolved {
			query.Where(incident.ResponseStateEQ(incident.ResponseStateResolved))
		} else {
			query.Where(incident.ResponseStateNEQ(incident.ResponseStateResolved))
		}
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

func (s *IncidentDetailsService) GetIncident(ctx context.Context, p predicate.Incident) (*ent.Incident, error) {
	query := s.db.Client(ctx).Incident.Query().Where(p)
	s.allQueryEdges(query)
	return query.Only(ctx)
}

func (s *IncidentDetailsService) UpdateIncident(ctx context.Context, id uuid.UUID, params rez.UpdateIncidentParams) (*ent.Incident, error) {
	var title *string
	if params.Title != nil {
		trimmed := strings.TrimSpace(*params.Title)
		if trimmed == "" {
			return nil, errIncidentNeedsTitle
		}
		title = &trimmed
	}
	txErr := s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		update := tx.Incident.UpdateOneID(id).
			SetNillableTitle(title).
			SetNillableSummary(params.Summary)
		if params.ClearSeverity {
			update.ClearSeverityID()
		} else if params.SeverityID != nil {
			update.SetSeverityID(*params.SeverityID)
		}
		if params.ClearType {
			update.ClearTypeID()
		} else if params.TypeID != nil {
			update.SetTypeID(*params.TypeID)
		}
		if params.FieldOptionIDs != nil {
			update.ClearFieldSelections().AddFieldSelectionIDs(*params.FieldOptionIDs...)
		}
		if params.TagIDs != nil {
			update.ClearTagAssignments().AddTagAssignmentIDs(*params.TagIDs...)
		}
		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update incident: %w", updateErr)
		}
		return s.publishUpdated(ctx, id)
	})
	if txErr != nil {
		return nil, txErr
	}
	return s.GetIncident(ctx, incident.ID(id))
}

func (s *IncidentDetailsService) LinkSituation(ctx context.Context, incidentID uuid.UUID, situationID uuid.UUID) error {
	return s.syncLinks(ctx, incidentID, rez.IncidentSituationLinkChanges{Added: []uuid.UUID{situationID}})
}

func (s *IncidentDetailsService) UnlinkSituation(ctx context.Context, incidentID uuid.UUID, situationID uuid.UUID) error {
	return s.syncLinks(ctx, incidentID, rez.IncidentSituationLinkChanges{Removed: []uuid.UUID{situationID}})
}

func (s *IncidentDetailsService) syncLinks(ctx context.Context, incidentID uuid.UUID, changes rez.IncidentSituationLinkChanges) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		if syncErr := s.situations.SyncIncidentLinks(ctx, incidentID, changes); syncErr != nil {
			return fmt.Errorf("sync incident situations: %w", syncErr)
		}
		return s.publishUpdated(ctx, incidentID)
	})
}

// Archive unlinks the incident's situations, so they re-evaluate, then soft-deletes it. Edge predicates still
// match archived incidents, so the links cannot be left in place.
func (s *IncidentDetailsService) ArchiveIncident(ctx context.Context, id uuid.UUID) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		// The incident row lock, which linking also takes, keeps a concurrent link from being left in place.
		lockIncident := tx.Incident.Query().
			Where(incident.ID(id)).
			ForUpdate()
		if _, lockErr := lockIncident.Only(ctx); lockErr != nil {
			return fmt.Errorf("lock incident: %w", lockErr)
		}
		queryLinked := tx.Situation.Query().
			Where(situation.HasIncidentsWith(incident.ID(id)))
		linkedIDs, queryErr := queryLinked.IDs(ctx)
		if queryErr != nil {
			return fmt.Errorf("query incident situations: %w", queryErr)
		}
		if syncErr := s.situations.SyncIncidentLinks(ctx, id, rez.IncidentSituationLinkChanges{Removed: linkedIDs}); syncErr != nil {
			return fmt.Errorf("unlink incident situations: %w", syncErr)
		}
		if archiveErr := tx.Incident.DeleteOneID(id).Exec(ctx); archiveErr != nil {
			return fmt.Errorf("archive incident: %w", archiveErr)
		}
		return nil
	})
}

func (s *IncidentDetailsService) SetChatChannel(ctx context.Context, id uuid.UUID, channelID string) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		update := tx.Incident.UpdateOneID(id).
			SetChatChannelID(channelID)
		if updateErr := update.Exec(ctx); updateErr != nil {
			return fmt.Errorf("set incident chat channel: %w", updateErr)
		}
		return s.publishUpdated(ctx, id)
	})
}

// SetChatDetailsMessage records the response channel's details message. It publishes nothing, so recording
// the message cannot start another channel sync.
func (s *IncidentDetailsService) SetChatDetailsMessage(ctx context.Context, id uuid.UUID, messageID string) error {
	return s.db.Client(ctx).Incident.UpdateOneID(id).
		SetChatDetailsMessageID(messageID).
		Exec(ctx)
}

// RecordObservedIncident creates or updates the incident observed through a provider, and returns its ID. An
// archived incident is left as it is. An observed incident is not run in Rezible, so nothing is published.
func (s *IncidentDetailsService) RecordObservedIncident(ctx context.Context, params rez.RecordObservedIncidentParams) (uuid.UUID, error) {
	return ent.WithTxReturning(ctx, s.db, func(ctx context.Context, tx *ent.Client) (uuid.UUID, error) {
		queryExisting := tx.Incident.Query().
			Where(incident.KnowledgeEntityID(params.KnowledgeEntityID))
		existing, queryErr := queryExisting.Only(schema.IncludeArchived(ctx))
		if queryErr != nil && !ent.IsNotFound(queryErr) {
			return uuid.Nil, fmt.Errorf("query observed incident: %w", queryErr)
		}
		if existing != nil && existing.ArchiveTime != nil {
			return existing.ID, nil
		}
		if existing != nil {
			update := existing.Update().
				SetTitle(params.Title).
				SetSummary(params.Summary).
				SetSeverityID(params.SeverityID).
				SetTypeID(params.TypeID).
				SetOpenedAt(params.OpenedAt)
			if updateErr := update.Exec(ctx); updateErr != nil {
				return uuid.Nil, fmt.Errorf("update observed incident: %w", updateErr)
			}
			return existing.ID, nil
		}

		number, numberErr := tx.SequenceCounter.Next(ctx, sequencecounter.KindIncident)
		if numberErr != nil {
			return uuid.Nil, fmt.Errorf("allocate incident number: %w", numberErr)
		}
		create := tx.Incident.Create().
			SetNumber(number).
			SetKnowledgeEntityID(params.KnowledgeEntityID).
			SetTitle(params.Title).
			SetSummary(params.Summary).
			SetSeverityID(params.SeverityID).
			SetTypeID(params.TypeID).
			SetOpenedAt(params.OpenedAt)
		created, createErr := create.Save(ctx)
		if createErr != nil {
			return uuid.Nil, fmt.Errorf("create observed incident: %w", createErr)
		}
		return created.ID, nil
	})
}

func (s *IncidentDetailsService) publishUpdated(ctx context.Context, id uuid.UUID) error {
	if publishErr := s.msgs.Publish(ctx, rez.EventOnIncidentUpdated{IncidentId: id}); publishErr != nil {
		return fmt.Errorf("publish incident update: %w", publishErr)
	}
	return nil
}

func (s *IncidentDetailsService) ListIncidentRoles(ctx context.Context) (ent.IncidentRoles, error) {
	return s.db.Client(ctx).IncidentRole.Query().Where(incidentrole.ArchiveTimeIsNil()).All(ctx)
}

func (s *IncidentDetailsService) ListIncidentSeverities(ctx context.Context) (ent.IncidentSeverities, error) {
	return s.db.Client(ctx).IncidentSeverity.Query().All(ctx)
}

func (s *IncidentDetailsService) GetMilestone(ctx context.Context, id uuid.UUID) (*ent.IncidentMilestone, error) {
	return s.db.Client(ctx).IncidentMilestone.Query().
		Where(im.ID(id)).
		WithUser().
		Only(ctx)
}

// SetMilestoneMetadata replaces the milestone's metadata. It publishes nothing.
func (s *IncidentDetailsService) SetMilestoneMetadata(ctx context.Context, id uuid.UUID, metadata map[string]string) error {
	return s.db.Client(ctx).IncidentMilestone.UpdateOneID(id).
		SetMetadata(metadata).
		Exec(ctx)
}

func (s *IncidentDetailsService) ListMilestones(ctx context.Context, incidentID uuid.UUID) (ent.IncidentMilestones, error) {
	return s.db.Client(ctx).IncidentMilestone.Query().
		Where(im.IncidentID(incidentID)).
		Order(im.ByTimestamp(), im.ByID()).
		WithUser().
		All(ctx)
}

func (s *IncidentDetailsService) ListIncidentTypes(ctx context.Context) (ent.IncidentTypes, error) {
	return s.db.Client(ctx).IncidentType.Query().All(ctx)
}

func (s *IncidentDetailsService) ListIncidentFields(ctx context.Context) (ent.IncidentFields, error) {
	return s.db.Client(ctx).IncidentField.Query().
		WithOptions().
		All(ctx)
}

func (s *IncidentDetailsService) ListIncidentTags(ctx context.Context) (ent.IncidentTags, error) {
	return s.db.Client(ctx).IncidentTag.Query().
		All(ctx)
}

func (s *IncidentDetailsService) GetIncidentMetadata(ctx context.Context) (*rez.IncidentMetadata, error) {
	md := rez.IncidentMetadata{}

	// TODO: use a view or get in parallel

	var rolesErr, typesErr, severitiesErr, fieldsErr, tagsErr error
	if md.Roles, rolesErr = s.ListIncidentRoles(ctx); rolesErr != nil {
		return nil, fmt.Errorf("roles: %w", rolesErr)
	}
	if md.Types, typesErr = s.ListIncidentTypes(ctx); typesErr != nil {
		return nil, fmt.Errorf("types: %w", typesErr)
	}
	if md.Severities, severitiesErr = s.ListIncidentSeverities(ctx); severitiesErr != nil {
		return nil, fmt.Errorf("severities: %w", severitiesErr)
	}
	if md.Fields, fieldsErr = s.ListIncidentFields(ctx); fieldsErr != nil {
		return nil, fmt.Errorf("fields: %w", fieldsErr)
	}
	if md.Tags, tagsErr = s.ListIncidentTags(ctx); tagsErr != nil {
		return nil, fmt.Errorf("tags: %w", tagsErr)
	}

	return &md, nil
}
