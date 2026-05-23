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

func handleIncidentEventProjection(ctx context.Context, client *ent.Client, event *ent.NormalizedEvent) error {
	if !projections.SubjectKindIncident.Matches(event) {
		return nil
	}

	observed, validationErr := projections.DecodeIncidentEvent(event)
	if validationErr != nil || observed == nil {
		return fmt.Errorf("invalid event: %w", validationErr)
	}
	h := &incidentEventProjectionHandler{client: client, observed: observed}
	return h.handle(ctx)
}

type incidentEventProjectionHandler struct {
	client   *ent.Client
	observed *projections.IncidentEvent
}

func (h *incidentEventProjectionHandler) handle(ctx context.Context) error {
	sevId, severityErr := h.upsertProjectedIncidentSeverity(ctx)
	if severityErr != nil {
		return fmt.Errorf("upsert incident severity: %w", severityErr)
	}
	typeId, typeErr := h.upsertProjectedIncidentType(ctx)
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

	var mut *ent.IncidentMutation
	if existing != nil {
		mut = existing.Update().Mutation()
	} else {
		incidentSlug, slugErr := h.generateProjectedIncidentSlug(ctx, openedAt, attrs.Title)
		if slugErr != nil {
			return fmt.Errorf("generate incident slug: %w", slugErr)
		}
		create := h.client.Incident.Create().
			SetProjectedFromID(ev.ID).
			SetSlug(incidentSlug)
		mut = create.Mutation()
	}
	mut.SetTitle(attrs.Title)
	mut.SetSummary(attrs.Summary)
	mut.SetOpenedAt(openedAt)
	mut.SetSeverityID(sevId)
	mut.SetTypeID(typeId)

	if _, mutErr := h.client.Mutate(ctx, mut); mutErr != nil {
		return fmt.Errorf("incident mutation: %w", mutErr)
	}

	return nil
}

func (h *incidentEventProjectionHandler) upsertProjectedIncidentSeverity(ctx context.Context) (uuid.UUID, error) {
	ref := h.observed.Attributes.SeverityRef
	upsert := h.client.IncidentSeverity.Create().
		SetName(ref).
		SetRank(0).
		OnConflictColumns(incsev.FieldTenantID, incsev.FieldName).
		DoNothing()
	return upsert.ID(ctx)
}

func (h *incidentEventProjectionHandler) upsertProjectedIncidentType(ctx context.Context) (uuid.UUID, error) {
	ref := h.observed.Attributes.TypeRef
	upsert := h.client.IncidentType.Create().
		SetName(ref).
		OnConflictColumns(incidenttype.FieldTenantID, incidenttype.FieldName).
		DoNothing()
	return upsert.ID(ctx)
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
