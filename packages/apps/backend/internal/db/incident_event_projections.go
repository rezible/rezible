package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/ent/incidenttype"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/integrations/projections"
)

func handleIncidentEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	if event.Kind != ne.KindIncidentObserved {
		return nil
	}

	observed, validationErr := projections.DecodeEvent[projections.IncidentObservedAttributes](event)
	if validationErr != nil || observed == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	h := &incidentEventProjectionHandler{client: client, observed: observed}
	return h.handle(ctx)
}

type incidentEventProjectionHandler struct {
	client   *ent.Client
	observed *projections.IncidentObserved
}

func (h *incidentEventProjectionHandler) handle(ctx context.Context) error {
	severity, severityErr := h.upsertProjectedIncidentSeverity(ctx)
	if severityErr != nil {
		return fmt.Errorf("upsert incident severity: %w", severityErr)
	}
	incidentType, typeErr := h.upsertProjectedIncidentType(ctx)
	if typeErr != nil {
		return fmt.Errorf("upsert incident type: %w", typeErr)
	}

	ev := h.observed.Event
	attrs := h.observed.Attributes

	openedAt := ev.OccurredAt
	if openedAt.IsZero() {
		openedAt = ev.ReceivedAt
	}
	if openedAt.IsZero() {
		openedAt = time.Now().UTC()
	}

	existing, queryExistingErr := h.client.Incident.Query().
		Where(incident.ProjectedEventID(ev.ID)).
		Only(ctx)
	if queryExistingErr != nil && !ent.IsNotFound(queryExistingErr) {
		return fmt.Errorf("query existing incident: %w", queryExistingErr)
	}

	var mut ent.Mutation
	if existing != nil {
		update := h.client.Incident.UpdateOne(existing).
			SetTitle(attrs.Title).
			SetSummary(attrs.Summary).
			SetOpenedAt(openedAt).
			SetSeverityID(severity.ID).
			SetTypeID(incidentType.ID)
		mut = update.Mutation()
	} else {
		incidentSlug, slugErr := h.generateProjectedIncidentSlug(ctx, openedAt, attrs.Title)
		if slugErr != nil {
			return fmt.Errorf("generate incident slug: %w", slugErr)
		}
		create := h.client.Incident.Create().
			SetProjectedFromID(ev.ID).
			SetSlug(incidentSlug).
			SetTitle(attrs.Title).
			SetSummary(attrs.Summary).
			SetOpenedAt(openedAt).
			SetSeverityID(severity.ID).
			SetTypeID(incidentType.ID)
		mut = create.Mutation()
	}

	if _, mutErr := h.client.Mutate(ctx, mut); mutErr != nil {
		return fmt.Errorf("incident projection mutation: %w", mutErr)
	}

	return nil
}

func (h *incidentEventProjectionHandler) upsertProjectedIncidentSeverity(ctx context.Context) (*ent.IncidentSeverity, error) {
	sevName := h.observed.Attributes.SeverityName
	sevRank := h.observed.Attributes.SeverityRank
	existing, queryErr := h.client.IncidentSeverity.Query().
		Where(incidentseverity.Name(sevName)).
		First(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query: %w", queryErr)
	}
	if existing != nil {
		return h.client.IncidentSeverity.UpdateOne(existing).
			SetRank(sevRank).
			Save(ctx)
	}
	return h.client.IncidentSeverity.Create().
		SetName(sevName).
		SetRank(sevRank).
		SetProjectedEventID(h.observed.Event.ID).
		Save(ctx)
}

func (h *incidentEventProjectionHandler) upsertProjectedIncidentType(ctx context.Context) (*ent.IncidentType, error) {
	typeName := h.observed.Attributes.TypeName
	existing, queryErr := h.client.IncidentType.Query().
		Where(incidenttype.Name(typeName)).
		First(ctx)
	if queryErr != nil && !ent.IsNotFound(queryErr) {
		return nil, fmt.Errorf("query: %w", queryErr)
	}
	if existing != nil {
		return existing, nil
	}
	return h.client.IncidentType.Create().
		SetName(typeName).
		Save(ctx)
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
