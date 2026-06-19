package db

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/ent/incident"
	ii "github.com/rezible/rezible/ent/incidentimpact"
	itag "github.com/rezible/rezible/ent/incidenttag"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

const (
	contextPackAlertLookback    = 72 * time.Hour
	contextPackEvidenceLookback = 72 * time.Hour
	contextPackAlertLimit       = 50
	contextPackEvidenceLimit    = 50
	contextPackRelatedLimit     = 5

	relationshipFunctionalityDependsOnComponent = "functionality_depends_on_component"
)

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

func (s *IncidentService) SetIncidentImpacts(ctx context.Context, params rez.SetIncidentImpactsParams) ([]*ent.IncidentImpact, error) {
	if params.IncidentID == uuid.Nil {
		return nil, fmt.Errorf("incident id is required")
	}

	var impacts []*ent.IncidentImpact
	txErr := s.db.WithTx(ctx, func(txCtx context.Context, tx *ent.Client) error {
		if _, getErr := tx.Incident.Get(txCtx, params.IncidentID); getErr != nil {
			return fmt.Errorf("get incident: %w", getErr)
		}

		entityIDs := make([]uuid.UUID, 0, len(params.Impacts))
		for _, input := range params.Impacts {
			entityID, resolveErr := s.resolveIncidentImpactEntity(txCtx, tx, input)
			if resolveErr != nil {
				return resolveErr
			}
			entityIDs = append(entityIDs, entityID)

			existing, queryErr := tx.IncidentImpact.Query().
				Where(ii.IncidentID(params.IncidentID), ii.KnowledgeEntityID(entityID)).
				Only(txCtx)
			if queryErr != nil && !ent.IsNotFound(queryErr) {
				return fmt.Errorf("query existing impact: %w", queryErr)
			}

			if existing == nil {
				create := tx.IncidentImpact.Create().
					SetIncidentID(params.IncidentID).
					SetKnowledgeEntityID(entityID)
				if input.Source != "" {
					create.SetSource(input.Source)
				}
				if input.Note != "" {
					create.SetNote(input.Note)
				}
				if _, createErr := create.Save(txCtx); createErr != nil {
					return fmt.Errorf("create incident impact: %w", createErr)
				}
				continue
			}

			update := tx.IncidentImpact.UpdateOne(existing)
			if input.Source != "" {
				update.SetSource(input.Source)
			} else {
				update.ClearSource()
			}
			if input.Note != "" {
				update.SetNote(input.Note)
			} else {
				update.ClearNote()
			}
			if _, updateErr := update.Save(txCtx); updateErr != nil {
				return fmt.Errorf("update incident impact: %w", updateErr)
			}
		}

		deleteQuery := tx.IncidentImpact.Delete().Where(ii.IncidentID(params.IncidentID))
		if len(entityIDs) > 0 {
			deleteQuery.Where(ii.KnowledgeEntityIDNotIn(entityIDs...))
		}
		if _, deleteErr := deleteQuery.Exec(txCtx); deleteErr != nil {
			return fmt.Errorf("delete replaced incident impacts: %w", deleteErr)
		}

		var listErr error
		impacts, listErr = tx.IncidentImpact.Query().
			Where(ii.IncidentID(params.IncidentID)).
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

	if pubErr := s.msgs.PublishEvent(ctx, rez.EventOnIncidentImpactsUpdated{IncidentId: params.IncidentID}); pubErr != nil {
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

type contextEntityCandidate struct {
	entity   *ent.KnowledgeEntity
	score    float64
	signals  map[string]struct{}
	evidence []rez.IncidentContextEvidenceRef
}

func (s *IncidentService) GetIncidentContextPack(ctx context.Context, incidentID uuid.UUID) (*rez.IncidentContextPack, error) {
	inc, incErr := s.Get(ctx, incident.ID(incidentID))
	if incErr != nil {
		return nil, fmt.Errorf("get incident: %w", incErr)
	}

	now := time.Now().UTC()
	pack := &rez.IncidentContextPack{
		IncidentID:           inc.ID,
		GeneratedAt:          now,
		ExplicitImpacts:      make([]rez.IncidentContextEntity, 0),
		InferredImpacts:      make([]rez.IncidentContextEntity, 0),
		ActiveAlerts:         make([]rez.IncidentContextAlert, 0),
		RecentEvidence:       make([]rez.IncidentContextEvidence, 0),
		RelatedIncidents:     make([]rez.IncidentContextRelatedIncident, 0),
		RetrievalLimitations: []string{"Active alerts are inferred from recent alert evidence until durable alert instance state is available."},
	}

	candidates := make(map[uuid.UUID]*contextEntityCandidate)
	addCandidate := func(entity *ent.KnowledgeEntity, score float64, signal string, evidence rez.IncidentContextEvidenceRef) {
		if entity == nil {
			return
		}
		c, ok := candidates[entity.ID]
		if !ok {
			c = &contextEntityCandidate{
				entity:  entity,
				signals: make(map[string]struct{}),
			}
			candidates[entity.ID] = c
		}
		c.score += score
		if signal != "" {
			c.signals[signal] = struct{}{}
		}
		if evidence.ID != uuid.Nil {
			c.evidence = append(c.evidence, evidence)
		}
	}

	explicitImpacts, impactsErr := s.ListIncidentImpacts(ctx, inc.ID)
	if impactsErr != nil {
		return nil, impactsErr
	}
	for _, impact := range explicitImpacts {
		entity := impact.Edges.KnowledgeEntity
		ev := rez.IncidentContextEvidenceRef{
			Kind:        "incident_impact",
			ID:          impact.ID,
			Description: impact.Note,
			ObservedAt:  impact.UpdatedAt,
		}
		addCandidate(entity, 100, "explicit_incident_impact", ev)
	}

	alertEntityIDs, alertEvidence, alertEvidenceErr := s.queryRecentAlertEvidence(ctx, now.Add(-contextPackAlertLookback))
	if alertEvidenceErr != nil {
		return nil, alertEvidenceErr
	}
	alerts, alertErr := s.contextAlertsFromEvidence(ctx, alertEntityIDs, alertEvidence)
	if alertErr != nil {
		return nil, alertErr
	}
	pack.ActiveAlerts = alerts

	if len(alertEntityIDs) > 0 {
		relationships, relErr := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.Or(knr.SourceEntityIDIn(alertEntityIDs...), knr.TargetEntityIDIn(alertEntityIDs...))).
			WithSourceEntity().
			WithTargetEntity().
			All(ctx)
		if relErr != nil {
			return nil, fmt.Errorf("query alert relationships: %w", relErr)
		}
		alertEntitySet := uuidSet(alertEntityIDs)
		for _, rel := range relationships {
			var related *ent.KnowledgeEntity
			if _, isAlert := alertEntitySet[rel.SourceEntityID]; isAlert {
				related = rel.Edges.TargetEntity
			}
			if _, isAlert := alertEntitySet[rel.TargetEntityID]; isAlert {
				related = rel.Edges.SourceEntity
			}
			if related == nil || related.Kind == knowledgeEntityKindAlert {
				continue
			}
			addCandidate(related, 50, "recent_alert_relationship", rez.IncidentContextEvidenceRef{
				Kind:        "knowledge_relationship",
				ID:          rel.ID,
				Description: rel.DisplayName,
				ObservedAt:  derefTime(rel.LastObservedAt, rel.UpdatedAt),
			})
		}
	}

	candidateIDs := make([]uuid.UUID, 0, len(candidates))
	for id := range candidates {
		candidateIDs = append(candidateIDs, id)
	}
	if len(candidateIDs) > 0 {
		funcRels, relErr := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.Kind(relationshipFunctionalityDependsOnComponent), knr.TargetEntityIDIn(candidateIDs...)).
			WithSourceEntity().
			All(ctx)
		if relErr != nil {
			return nil, fmt.Errorf("query functionality dependencies: %w", relErr)
		}
		for _, rel := range funcRels {
			addCandidate(rel.Edges.SourceEntity, 35, "functionality_dependency", rez.IncidentContextEvidenceRef{
				Kind:        "knowledge_relationship",
				ID:          rel.ID,
				Description: rel.DisplayName,
				ObservedAt:  derefTime(rel.LastObservedAt, rel.UpdatedAt),
			})
		}
	}

	pack.ExplicitImpacts = s.contextEntitiesFromImpacts(explicitImpacts, candidates)
	pack.InferredImpacts = contextEntitiesFromCandidates(candidates)
	pack.RecentEvidence, _ = s.queryRecentContextEvidence(ctx, candidateIDs, now.Add(-contextPackEvidenceLookback))
	pack.RelatedIncidents, _ = s.queryRelatedIncidents(ctx, inc, candidateIDs)
	return pack, nil
}

func (s *IncidentService) queryRecentAlertEvidence(ctx context.Context, since time.Time) ([]uuid.UUID, map[uuid.UUID][]rez.IncidentContextEvidenceRef, error) {
	evidence, queryErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.Assertion(assertionAlertDefinitionObserved), knev.ObservedAtGTE(since)).
		WithEntity().
		Order(knev.ByObservedAt(sql.OrderDesc())).
		Limit(contextPackAlertLimit).
		All(ctx)
	if queryErr != nil {
		return nil, nil, fmt.Errorf("query recent alert evidence: %w", queryErr)
	}

	ids := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	refsByEntityID := make(map[uuid.UUID][]rez.IncidentContextEvidenceRef)
	for _, ev := range evidence {
		entity := ev.Edges.Entity
		if entity == nil || entity.Kind != knowledgeEntityKindAlert {
			continue
		}
		if _, ok := seen[entity.ID]; !ok {
			seen[entity.ID] = struct{}{}
			ids = append(ids, entity.ID)
		}
		refsByEntityID[entity.ID] = append(refsByEntityID[entity.ID], rez.IncidentContextEvidenceRef{
			Kind:        "knowledge_evidence",
			ID:          ev.ID,
			Description: ev.Assertion,
			ObservedAt:  ev.ObservedAt,
		})
	}
	return ids, refsByEntityID, nil
}

func (s *IncidentService) contextAlertsFromEvidence(ctx context.Context, alertEntityIDs []uuid.UUID, refs map[uuid.UUID][]rez.IncidentContextEvidenceRef) ([]rez.IncidentContextAlert, error) {
	if len(alertEntityIDs) == 0 {
		return []rez.IncidentContextAlert{}, nil
	}
	alerts, queryErr := s.db.Client(ctx).Alert.Query().
		Where(alert.KnowledgeEntityIDIn(alertEntityIDs...)).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query context alerts: %w", queryErr)
	}

	relatedIDsByAlertEntityID, relErr := s.relatedEntityIDsForAlertEntities(ctx, alertEntityIDs)
	if relErr != nil {
		return nil, relErr
	}
	res := make([]rez.IncidentContextAlert, 0, len(alerts))
	for _, a := range alerts {
		if a.KnowledgeEntityID == nil {
			continue
		}
		evidence := refs[*a.KnowledgeEntityID]
		res = append(res, rez.IncidentContextAlert{
			ID:                a.ID,
			KnowledgeEntityID: *a.KnowledgeEntityID,
			Title:             a.Title,
			Description:       a.Description,
			ObservedAt:        latestEvidenceTime(evidence),
			RelatedEntityIDs:  relatedIDsByAlertEntityID[*a.KnowledgeEntityID],
			Evidence:          evidence,
		})
	}
	return res, nil
}

func (s *IncidentService) relatedEntityIDsForAlertEntities(ctx context.Context, alertEntityIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	res := make(map[uuid.UUID][]uuid.UUID)
	if len(alertEntityIDs) == 0 {
		return res, nil
	}
	relationships, queryErr := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(alertEntityIDs...), knr.TargetEntityIDIn(alertEntityIDs...))).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query alert related entity ids: %w", queryErr)
	}
	alertSet := uuidSet(alertEntityIDs)
	for _, rel := range relationships {
		if _, ok := alertSet[rel.SourceEntityID]; ok {
			res[rel.SourceEntityID] = append(res[rel.SourceEntityID], rel.TargetEntityID)
		}
		if _, ok := alertSet[rel.TargetEntityID]; ok {
			res[rel.TargetEntityID] = append(res[rel.TargetEntityID], rel.SourceEntityID)
		}
	}
	return res, nil
}

func (s *IncidentService) contextEntitiesFromImpacts(impacts []*ent.IncidentImpact, candidates map[uuid.UUID]*contextEntityCandidate) []rez.IncidentContextEntity {
	res := make([]rez.IncidentContextEntity, 0, len(impacts))
	for _, impact := range impacts {
		candidate := candidates[impact.KnowledgeEntityID]
		if candidate == nil {
			continue
		}
		res = append(res, candidate.toContextEntity())
	}
	return res
}

func contextEntitiesFromCandidates(candidates map[uuid.UUID]*contextEntityCandidate) []rez.IncidentContextEntity {
	res := make([]rez.IncidentContextEntity, 0, len(candidates))
	for _, c := range candidates {
		res = append(res, c.toContextEntity())
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Score == res[j].Score {
			return res[i].DisplayName < res[j].DisplayName
		}
		return res[i].Score > res[j].Score
	})
	return res
}

func (c *contextEntityCandidate) toContextEntity() rez.IncidentContextEntity {
	signals := make([]string, 0, len(c.signals))
	for signal := range c.signals {
		signals = append(signals, signal)
	}
	sort.Strings(signals)
	return rez.IncidentContextEntity{
		ID:          c.entity.ID,
		Kind:        c.entity.Kind,
		DisplayName: c.entity.DisplayName,
		Description: c.entity.Description,
		Score:       c.score,
		Signals:     signals,
		Evidence:    c.evidence,
	}
}

func (s *IncidentService) queryRecentContextEvidence(ctx context.Context, entityIDs []uuid.UUID, since time.Time) ([]rez.IncidentContextEvidence, error) {
	if len(entityIDs) == 0 {
		return []rez.IncidentContextEvidence{}, nil
	}
	evidence, queryErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.ObservedAtGTE(since), knev.Or(knev.EntityIDIn(entityIDs...), knev.HasRelationshipWith(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))))).
		Order(knev.ByObservedAt(sql.OrderDesc())).
		Limit(contextPackEvidenceLimit).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query recent context evidence: %w", queryErr)
	}
	res := make([]rez.IncidentContextEvidence, len(evidence))
	for i, ev := range evidence {
		res[i] = rez.IncidentContextEvidence{
			ID:             ev.ID,
			EventID:        ev.EventID,
			EntityID:       ev.EntityID,
			RelationshipID: ev.RelationshipID,
			SubjectType:    ev.SubjectType.String(),
			Assertion:      ev.Assertion,
			EvidenceKind:   ev.EvidenceKind.String(),
			ObservedAt:     ev.ObservedAt,
			Description:    ev.Assertion,
		}
	}
	return res, nil
}

func (s *IncidentService) queryRelatedIncidents(ctx context.Context, inc *ent.Incident, entityIDs []uuid.UUID) ([]rez.IncidentContextRelatedIncident, error) {
	tagIDs := make([]uuid.UUID, 0, len(inc.Edges.TagAssignments))
	for _, tag := range inc.Edges.TagAssignments {
		tagIDs = append(tagIDs, tag.ID)
	}
	if len(entityIDs) == 0 && len(tagIDs) == 0 {
		return []rez.IncidentContextRelatedIncident{}, nil
	}
	query := s.db.Client(ctx).Incident.Query().
		Where(incident.IDNEQ(inc.ID)).
		WithImpacts().
		WithTagAssignments().
		Order(incident.ByOpenedAt(sql.OrderDesc())).
		Limit(contextPackRelatedLimit)
	if len(entityIDs) > 0 && len(tagIDs) > 0 {
		query.Where(incident.Or(
			incident.HasImpactsWith(ii.KnowledgeEntityIDIn(entityIDs...)),
			incident.HasTagAssignmentsWith(itag.IDIn(tagIDs...)),
		))
	} else if len(entityIDs) > 0 {
		query.Where(incident.HasImpactsWith(ii.KnowledgeEntityIDIn(entityIDs...)))
	} else {
		query.Where(incident.HasTagAssignmentsWith(itag.IDIn(tagIDs...)))
	}
	incs, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query related incidents: %w", queryErr)
	}
	res := make([]rez.IncidentContextRelatedIncident, 0, len(incs))
	for _, related := range incs {
		entityIDs := make([]uuid.UUID, 0, len(related.Edges.Impacts))
		for _, impact := range related.Edges.Impacts {
			entityIDs = append(entityIDs, impact.KnowledgeEntityID)
		}
		res = append(res, rez.IncidentContextRelatedIncident{
			ID:        related.ID,
			Slug:      related.Slug,
			Title:     related.Title,
			OpenedAt:  related.OpenedAt,
			Signals:   []string{"shared_impact_or_tag"},
			EntityIDs: entityIDs,
		})
	}
	return res, nil
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	res := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		res[id] = struct{}{}
	}
	return res
}

func latestEvidenceTime(refs []rez.IncidentContextEvidenceRef) time.Time {
	var latest time.Time
	for _, ref := range refs {
		if ref.ObservedAt.After(latest) {
			latest = ref.ObservedAt
		}
	}
	return latest
}

func derefTime(v *time.Time, fallback time.Time) time.Time {
	if v == nil {
		return fallback
	}
	return *v
}
