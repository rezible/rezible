package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/incidenttype"
	"github.com/rezible/rezible/projections"
)

const (
	assertionIncidentObserved = "incident_observed"
	knowledgeKindIncident     = "incident"
)

func (s *IncidentService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) error {
	if projections.SubjectKindIncident.Matches(event) {
		decoded, validationErr := projections.DecodeIncidentEvent(event)
		if validationErr != nil || decoded == nil {
			return fmt.Errorf("invalid event: %w", validationErr)
		}
		return s.handleIncidentEventProjection(ctx, decoded)
	}
	return nil
}

func (s *IncidentService) handleIncidentEventProjection(ctx context.Context, ie *projections.IncidentEvent) error {
	attrs := ie.Attributes
	openedAt := attrs.OpenedAt
	if openedAt.IsZero() {
		openedAt = ie.Event.DeriveObservedAt()
	}

	knowledgeEntity := rez.ProjectedKnowledgeEntity{
		Kind:              knowledgeKindIncident,
		DisplayName:       attrs.Title,
		EvidenceAssertion: assertionIncidentObserved,
		AliasRefs:         []ent.KnowledgeEntityAliasRef{ie.Event.MakeEntityAliasRef()},
	}
	return s.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		knowledgeEntityId, knowledgeErr := s.knowledge.ResolveProjectedEntity(ctx, ie.Event, knowledgeEntity)
		if knowledgeErr != nil {
			return fmt.Errorf("resolve incident knowledge entity: %w", knowledgeErr)
		}

		sevId, severityErr := s.saveProjectedIncidentSeverity(ctx, attrs)
		if severityErr != nil {
			return fmt.Errorf("upsert incident severity: %w", severityErr)
		}

		typeId, typeErr := s.saveProjectedIncidentType(ctx, attrs)
		if typeErr != nil {
			return fmt.Errorf("upsert incident type: %w", typeErr)
		}

		queryExisting := s.db.Client(ctx).Incident.Query().
			Where(incident.KnowledgeEntityID(knowledgeEntityId))
		existing, existingErr := queryExisting.Only(ctx)
		if existingErr != nil && !ent.IsNotFound(existingErr) {
			return fmt.Errorf("query existing incident: %w", existingErr)
		}

		id := uuid.Nil
		if existing != nil {
			// TODO: helper method on attributes struct?
			if existing.Title == attrs.Title &&
				existing.Summary == attrs.Summary &&
				existing.SeverityID == sevId &&
				existing.TypeID == typeId {
				return nil
			}
			id = existing.ID
		}

		setFn := func(m *ent.IncidentMutation) {
			m.SetKnowledgeEntityID(knowledgeEntityId)
			m.SetTitle(attrs.Title)
			m.SetSummary(attrs.Summary)
			m.SetSeverityID(sevId)
			m.SetTypeID(typeId)
			if !openedAt.IsZero() {
				m.SetOpenedAt(openedAt)
			}
		}
		if _, setErr := s.Set(ctx, id, setFn); setErr != nil {
			return fmt.Errorf("set incident: %w", setErr)
		}

		return nil
	})
}

func (s *IncidentService) saveProjectedIncidentSeverity(ctx context.Context, attrs projections.IncidentSubjectAttributes) (uuid.UUID, error) {
	existing, queryErr := s.db.Client(ctx).IncidentSeverity.Query().
		Where(incsev.Name(attrs.SeverityRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident severity: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	created, createErr := s.db.Client(ctx).IncidentSeverity.Create().
		SetName(attrs.SeverityRef).
		SetRank(0).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident severity: %w", createErr)
	}
	return created.ID, nil
}

func (s *IncidentService) saveProjectedIncidentType(ctx context.Context, attrs projections.IncidentSubjectAttributes) (uuid.UUID, error) {
	existing, queryErr := s.db.Client(ctx).IncidentType.Query().
		Where(incidenttype.Name(attrs.TypeRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident type: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}
	created, createErr := s.db.Client(ctx).IncidentType.Create().
		SetName(attrs.TypeRef).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident type: %w", createErr)
	}
	return created.ID, nil
}
