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
	"github.com/rezible/rezible/integrations/projections"
)

const (
	assertionIncidentObserved = "incident_observed"
	knowledgeKindIncident     = "incident"
)

func (s *IncidentService) HandleEventProjection(ctx context.Context, event *ent.NormalizedEvent) error {
	if !projections.SubjectKindIncident.Matches(event) {
		return nil
	}
	decoded, validationErr := projections.DecodeIncidentEvent(event)
	if validationErr != nil || decoded == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	sevId, severityErr := s.saveProjectedIncidentSeverity(ctx, decoded.Attributes)
	if severityErr != nil {
		return fmt.Errorf("upsert incident severity: %w", severityErr)
	}
	typeId, typeErr := s.saveProjectedIncidentType(ctx, decoded.Attributes)
	if typeErr != nil {
		return fmt.Errorf("upsert incident type: %w", typeErr)
	}

	attrs := decoded.Attributes
	observedAt := observedAtForEvent(event)
	entityParams := rez.ResolveKnowledgeEntityParams{
		Event:             event,
		EvidenceAssertion: assertionIncidentObserved,
		Entity: &ent.KnowledgeEntity{
			Kind:        knowledgeKindIncident,
			DisplayName: attrs.Title,
		},
		Aliases: eventKnowledgeEntityAliases(event),
	}
	knowledgeEntity, knowledgeErr := s.knowledge.ResolveEntity(ctx, entityParams)
	if knowledgeErr != nil {
		return fmt.Errorf("resolve incident knowledge entity: %w", knowledgeErr)
	}

	updateCount, updateErr := s.client.Incident.Update().
		Where(incident.KnowledgeEntityID(knowledgeEntity.ID)).
		SetTitle(attrs.Title).
		SetSummary(attrs.Summary).
		SetSeverityID(sevId).
		SetTypeID(typeId).
		Save(ctx)
	if updateErr != nil {
		return fmt.Errorf("update incident: %w", updateErr)
	}
	if updateCount > 1 {
		return fmt.Errorf("expected at most one incident for knowledge entity %s, updated %d",
			knowledgeEntity.ID, updateCount)
	}
	if updateCount == 1 {
		return nil
	}

	openedAt := observedAt
	incidentSlug, slugErr := s.generateIncidentSlug(ctx, openedAt)
	if slugErr != nil {
		return fmt.Errorf("generate incident slug: %w", slugErr)
	}
	if _, createErr := s.client.Incident.Create().
		SetKnowledgeEntityID(knowledgeEntity.ID).
		SetSlug(incidentSlug).
		SetTitle(attrs.Title).
		SetSummary(attrs.Summary).
		SetOpenedAt(openedAt).
		SetSeverityID(sevId).
		SetTypeID(typeId).
		Save(ctx); createErr != nil {
		return fmt.Errorf("create incident: %w", createErr)
	}

	return nil
}

func (s *IncidentService) saveProjectedIncidentSeverity(ctx context.Context, attrs projections.IncidentSubjectAttributes) (uuid.UUID, error) {
	existing, queryErr := s.client.IncidentSeverity.Query().
		Where(incsev.Name(attrs.SeverityRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident severity: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	created, createErr := s.client.IncidentSeverity.Create().
		SetName(attrs.SeverityRef).
		SetRank(0).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident severity: %w", createErr)
	}
	return created.ID, nil
}

func (s *IncidentService) saveProjectedIncidentType(ctx context.Context, attrs projections.IncidentSubjectAttributes) (uuid.UUID, error) {
	existing, queryErr := s.client.IncidentType.Query().
		Where(incidenttype.Name(attrs.TypeRef)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident type: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}
	created, createErr := s.client.IncidentType.Create().
		SetName(attrs.TypeRef).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident type: %w", createErr)
	}
	return created.ID, nil
}
