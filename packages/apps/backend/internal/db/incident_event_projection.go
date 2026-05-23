package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
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

func handleIncidentEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	if projections.SubjectKindIncident.Matches(event) {
		observed, validationErr := projections.DecodeIncidentEvent(event)
		if validationErr != nil || observed == nil {
			return fmt.Errorf("invalid event: %w", validationErr)
		}
		h := &incidentEventProjectionHandler{
			client:    client,
			observed:  observed,
			knowledge: newKnowledgeService(client),
		}
		return h.handle(ctx)
	}

	return nil
}

type incidentEventProjectionHandler struct {
	client    *ent.Client
	observed  *projections.IncidentEvent
	knowledge *KnowledgeService
}

func (h *incidentEventProjectionHandler) handle(ctx context.Context) error {
	sevId, severityErr := h.saveProjectedIncidentSeverity(ctx)
	if severityErr != nil {
		return fmt.Errorf("upsert incident severity: %w", severityErr)
	}
	typeId, typeErr := h.saveProjectedIncidentType(ctx)
	if typeErr != nil {
		return fmt.Errorf("upsert incident type: %w", typeErr)
	}

	ev := h.observed.Event

	attrs := h.observed.Attributes
	observedAt := observedAtForEvent(ev)
	entityParams := ResolveKnowledgeEntityParams{
		Event:       ev,
		Kind:        knowledgeKindIncident,
		DisplayName: attrs.Title,
		Aliases:     []EntityAliasRef{entityAliasRefForEvent(ev, "")},
	}
	savedKnowledge, knowledgeErr := h.knowledge.ResolveEntityWithAssertion(ctx, entityParams, assertionIncidentObserved)
	if knowledgeErr != nil {
		return fmt.Errorf("resolve incident knowledge entity: %w", knowledgeErr)
	}

	updateCount, updateErr := h.client.Incident.Update().
		Where(incident.KnowledgeEntityID(savedKnowledge.Entity.ID)).
		SetTitle(attrs.Title).
		SetSummary(attrs.Summary).
		SetSeverityID(sevId).
		SetTypeID(typeId).
		Save(ctx)
	if updateErr != nil {
		return fmt.Errorf("update incident: %w", updateErr)
	}
	if updateCount > 1 {
		return fmt.Errorf("expected at most one incident for knowledge entity %s, updated %d", savedKnowledge.Entity.ID, updateCount)
	}
	if updateCount == 1 {
		return nil
	}

	openedAt := observedAt
	incidentSlug, slugErr := h.generateProjectedIncidentSlug(ctx, openedAt, attrs.Title)
	if slugErr != nil {
		return fmt.Errorf("generate incident slug: %w", slugErr)
	}
	if _, createErr := h.client.Incident.Create().
		SetKnowledgeEntityID(savedKnowledge.Entity.ID).
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

func (h *incidentEventProjectionHandler) saveProjectedIncidentSeverity(ctx context.Context) (uuid.UUID, error) {
	ref := h.observed.Attributes.SeverityRef
	existing, queryErr := h.client.IncidentSeverity.Query().
		Where(incsev.Name(ref)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident severity: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	created, createErr := h.client.IncidentSeverity.Create().
		SetName(ref).
		SetRank(0).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident severity: %w", createErr)
	}
	return created.ID, nil
}

func (h *incidentEventProjectionHandler) saveProjectedIncidentType(ctx context.Context) (uuid.UUID, error) {
	ref := h.observed.Attributes.TypeRef
	existing, queryErr := h.client.IncidentType.Query().
		Where(incidenttype.Name(ref)).
		Only(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return uuid.Nil, fmt.Errorf("query incident type: %w", queryErr)
	}
	if existing != nil {
		return existing.ID, nil
	}

	created, createErr := h.client.IncidentType.Create().
		SetName(ref).
		Save(ctx)
	if createErr != nil {
		return uuid.Nil, fmt.Errorf("create incident type: %w", createErr)
	}
	return created.ID, nil
}

func (h *incidentEventProjectionHandler) generateProjectedIncidentSlug(ctx context.Context, openedAt time.Time, title string) (string, error) {
	datePrefix := openedAt.Format("060102")
	base := slug.Make(fmt.Sprintf("%s-%s", datePrefix, title))
	if base == "" {
		base = slug.Make(fmt.Sprintf("%s-incident", datePrefix))
	}

	const maxRetries = 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		candidate := base
		if attempt > 0 {
			candidate = fmt.Sprintf("%s-%d", base, attempt+1)
		}
		exists, queryErr := h.client.Incident.Query().Where(incident.Slug(candidate)).Exist(ctx)
		if queryErr != nil {
			return "", fmt.Errorf("check uniqueness: %w", queryErr)
		}
		if !exists {
			return candidate, nil
		}
	}

	return fmt.Sprintf("%s-%s", base, uuid.NewString()[:8]), nil
}
