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

func (s *IncidentService) SetIncidentImpacts(ctx context.Context, id uuid.UUID, input []rez.IncidentImpactInput) ([]*ent.IncidentImpact, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("incident id is required")
	}

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

type contextEntityCandidate struct {
	entity   *ent.KnowledgeEntity
	score    float64
	signals  map[string]struct{}
	evidence []incidentContextEvidenceRef
}

type incidentContextEvidenceRef struct {
	ID             uuid.UUID
	EntityID       *uuid.UUID
	RelationshipID *uuid.UUID
	Kind           string
	Description    string
	ObservedAt     time.Time
}

func (s *IncidentService) GetIncidentContext(ctx context.Context, incidentID uuid.UUID) (*rez.AgentWorkflowContext, error) {
	inc, incErr := s.Get(ctx, incident.ID(incidentID))
	if incErr != nil {
		return nil, fmt.Errorf("get incident: %w", incErr)
	}

	now := time.Now().UTC()
	result := &rez.AgentWorkflowContext{
		GeneratedAt:  now,
		PromptSchema: "incident_context_pack_context.v1",
		Context: map[string]any{
			"incidentId":           inc.ID,
			"generatedAt":          now,
			"retrievalLimitations": []string{"Active alerts are inferred from recent alert evidence until durable alert instance state is available."},
		},
		Limitations: []string{"Active alerts are inferred from recent alert evidence until durable alert instance state is available."},
	}
	addWorkflowContextCitation(result, rez.AgentRunCitationInput{
		CitationKind:     "primary_subject",
		DomainEntityType: "incident",
		DomainEntityID:   inc.ID,
		Summary:          inc.Title,
		Snapshot: map[string]any{
			"title":     inc.Title,
			"summary":   inc.Summary,
			"slug":      inc.Slug,
			"openedAt":  inc.OpenedAt,
			"updatedAt": inc.UpdatedAt,
		},
	})

	candidates := make(map[uuid.UUID]*contextEntityCandidate)
	addCandidate := func(entity *ent.KnowledgeEntity, score float64, signal string, evidence incidentContextEvidenceRef) {
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
		ev := incidentContextEvidenceRef{
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
	alerts, alertErr := s.contextAlertsFromEvidence(ctx, result, alertEntityIDs, alertEvidence)
	if alertErr != nil {
		return nil, alertErr
	}
	result.Items = append(result.Items, alerts...)

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
			citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
				CitationKind:            "supporting_evidence",
				KnowledgeRelationshipID: rel.ID,
				Summary:                 rel.DisplayName,
				Snapshot: map[string]any{
					"kind":           rel.Kind,
					"displayName":    rel.DisplayName,
					"sourceEntityId": rel.SourceEntityID,
					"targetEntityId": rel.TargetEntityID,
				},
			})
			addCandidate(related, 50, "recent_alert_relationship", incidentContextEvidenceRef{
				Kind:        "knowledge_relationship",
				ID:          rel.ID,
				Description: rel.DisplayName,
				ObservedAt:  derefTime(rel.LastObservedAt, rel.UpdatedAt),
			})
			result.Items = append(result.Items, relationshipContextItem("related_alert", rel, related, citation))
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
			citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
				CitationKind:            "related_entity",
				KnowledgeRelationshipID: rel.ID,
				Summary:                 rel.DisplayName,
				Snapshot: map[string]any{
					"kind":           rel.Kind,
					"displayName":    rel.DisplayName,
					"sourceEntityId": rel.SourceEntityID,
					"targetEntityId": rel.TargetEntityID,
				},
			})
			addCandidate(rel.Edges.SourceEntity, 35, "functionality_dependency", incidentContextEvidenceRef{
				Kind:        "knowledge_relationship",
				ID:          rel.ID,
				Description: rel.DisplayName,
				ObservedAt:  derefTime(rel.LastObservedAt, rel.UpdatedAt),
			})
			if rel.Edges.SourceEntity != nil {
				result.Items = append(result.Items, relationshipContextItem("functionality_dependency", rel, rel.Edges.SourceEntity, citation))
			}
		}
	}

	result.Items = append(result.Items, s.contextEntitiesFromImpacts(result, explicitImpacts, candidates)...)
	result.Items = append(result.Items, contextEntitiesFromCandidates(result, candidates)...)
	if evidenceErr := s.addRecentContextEvidence(ctx, result, candidateIDs, now.Add(-contextPackEvidenceLookback)); evidenceErr != nil {
		result.Limitations = append(result.Limitations, "Recent context evidence was unavailable: "+evidenceErr.Error())
	}
	if relatedErr := s.addRelatedIncidents(ctx, result, inc, candidateIDs); relatedErr != nil {
		result.Limitations = append(result.Limitations, "Related incident search was unavailable: "+relatedErr.Error())
	}
	result.Context["limitations"] = result.Limitations
	result.Context["itemCounts"] = countContextItemsByRole(result.Items)
	return result, nil
}

func (s *IncidentService) queryRecentAlertEvidence(ctx context.Context, since time.Time) ([]uuid.UUID, map[uuid.UUID][]incidentContextEvidenceRef, error) {
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
	refsByEntityID := make(map[uuid.UUID][]incidentContextEvidenceRef)
	for _, ev := range evidence {
		entity := ev.Edges.Entity
		if entity == nil || entity.Kind != knowledgeEntityKindAlert {
			continue
		}
		if _, ok := seen[entity.ID]; !ok {
			seen[entity.ID] = struct{}{}
			ids = append(ids, entity.ID)
		}
		refsByEntityID[entity.ID] = append(refsByEntityID[entity.ID], incidentContextEvidenceRef{
			Kind:        "knowledge_evidence",
			ID:          ev.ID,
			Description: ev.Assertion,
			ObservedAt:  ev.ObservedAt,
		})
	}
	return ids, refsByEntityID, nil
}

func (s *IncidentService) contextAlertsFromEvidence(ctx context.Context, result *rez.AgentWorkflowContext, alertEntityIDs []uuid.UUID, refs map[uuid.UUID][]incidentContextEvidenceRef) ([]rez.AgentWorkflowContextItem, error) {
	if len(alertEntityIDs) == 0 {
		return []rez.AgentWorkflowContextItem{}, nil
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
	res := make([]rez.AgentWorkflowContextItem, 0, len(alerts))
	for _, a := range alerts {
		if a.KnowledgeEntityID == nil {
			continue
		}
		evidence := refs[*a.KnowledgeEntityID]
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:     "related_entity",
			DomainEntityType: "alert",
			DomainEntityID:   a.ID,
			Summary:          a.Title,
			Snapshot: map[string]any{
				"title":     a.Title,
				"summary":   a.Description,
				"source":    "incident_context",
				"openedAt":  latestEvidenceTime(evidence),
				"entityIds": relatedIDsByAlertEntityID[*a.KnowledgeEntityID],
			},
		})
		res = append(res, rez.AgentWorkflowContextItem{
			Kind:     "domain_reference",
			Role:     "active_alert",
			Name:     a.Title,
			Citation: citation,
			Payload: map[string]any{
				"id":        a.ID,
				"kind":      "alert",
				"title":     a.Title,
				"summary":   a.Description,
				"source":    "incident_context",
				"openedAt":  latestEvidenceTime(evidence),
				"entityIds": relatedIDsByAlertEntityID[*a.KnowledgeEntityID],
			},
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

func (s *IncidentService) contextEntitiesFromImpacts(result *rez.AgentWorkflowContext, impacts []*ent.IncidentImpact, candidates map[uuid.UUID]*contextEntityCandidate) []rez.AgentWorkflowContextItem {
	res := make([]rez.AgentWorkflowContextItem, 0, len(impacts))
	for _, impact := range impacts {
		candidate := candidates[impact.KnowledgeEntityID]
		if candidate == nil {
			continue
		}
		res = append(res, candidate.toContextItem(result, "explicit_impact"))
	}
	return res
}

func contextEntitiesFromCandidates(result *rez.AgentWorkflowContext, candidates map[uuid.UUID]*contextEntityCandidate) []rez.AgentWorkflowContextItem {
	res := make([]rez.AgentWorkflowContextItem, 0, len(candidates))
	for _, c := range candidates {
		res = append(res, c.toContextItem(result, "inferred_impact"))
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res
}

func (c *contextEntityCandidate) toContextItem(result *rez.AgentWorkflowContext, role string) rez.AgentWorkflowContextItem {
	signals := make([]string, 0, len(c.signals))
	for signal := range c.signals {
		signals = append(signals, signal)
	}
	sort.Strings(signals)
	citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
		CitationKind:      "related_entity",
		KnowledgeEntityID: c.entity.ID,
		Summary:           c.entity.DisplayName,
		Snapshot: map[string]any{
			"kind":        c.entity.Kind,
			"displayName": c.entity.DisplayName,
			"description": c.entity.Description,
			"score":       c.score,
			"reason":      stringsFromSignals(signals),
		},
	})
	return rez.AgentWorkflowContextItem{
		Kind:     "knowledge_entity",
		Role:     role,
		Name:     c.entity.DisplayName,
		Citation: citation,
		Payload: map[string]any{
			"entityId":    c.entity.ID,
			"kind":        c.entity.Kind,
			"displayName": c.entity.DisplayName,
			"description": c.entity.Description,
			"score":       c.score,
			"reason":      stringsFromSignals(signals),
		},
	}
}

func (s *IncidentService) addRecentContextEvidence(ctx context.Context, result *rez.AgentWorkflowContext, entityIDs []uuid.UUID, since time.Time) error {
	if len(entityIDs) == 0 {
		return nil
	}
	evidence, queryErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.ObservedAtGTE(since), knev.Or(knev.EntityIDIn(entityIDs...), knev.HasRelationshipWith(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))))).
		Order(knev.ByObservedAt(sql.OrderDesc())).
		Limit(contextPackEvidenceLimit).
		All(ctx)
	if queryErr != nil {
		return fmt.Errorf("query recent context evidence: %w", queryErr)
	}
	for _, ev := range evidence {
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:        "supporting_evidence",
			KnowledgeEvidenceID: ev.ID,
			Summary:             ev.Assertion,
			Snapshot: map[string]any{
				"eventId":        ev.EventID,
				"entityId":       ev.EntityID,
				"relationshipId": ev.RelationshipID,
				"kind":           ev.EvidenceKind.String(),
				"summary":        ev.Assertion,
				"observedAt":     ev.ObservedAt,
				"subjectType":    ev.SubjectType.String(),
			},
		})
		result.Items = append(result.Items, rez.AgentWorkflowContextItem{
			Kind:     "knowledge_evidence",
			Role:     "recent_evidence",
			Name:     ev.Assertion,
			Citation: citation,
			Payload: map[string]any{
				"id":             ev.ID,
				"eventId":        ev.EventID,
				"entityId":       ev.EntityID,
				"relationshipId": ev.RelationshipID,
				"kind":           ev.EvidenceKind.String(),
				"summary":        ev.Assertion,
				"observedAt":     ev.ObservedAt,
				"subjectType":    ev.SubjectType.String(),
			},
		})
	}
	return nil
}

func (s *IncidentService) addRelatedIncidents(ctx context.Context, result *rez.AgentWorkflowContext, inc *ent.Incident, entityIDs []uuid.UUID) error {
	tagIDs := make([]uuid.UUID, 0, len(inc.Edges.TagAssignments))
	for _, tag := range inc.Edges.TagAssignments {
		tagIDs = append(tagIDs, tag.ID)
	}
	if len(entityIDs) == 0 && len(tagIDs) == 0 {
		return nil
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
		return fmt.Errorf("query related incidents: %w", queryErr)
	}
	for _, related := range incs {
		entityIDs := make([]uuid.UUID, 0, len(related.Edges.Impacts))
		for _, impact := range related.Edges.Impacts {
			entityIDs = append(entityIDs, impact.KnowledgeEntityID)
		}
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:     "historical_example",
			DomainEntityType: "incident",
			DomainEntityID:   related.ID,
			Summary:          related.Title,
			Snapshot: map[string]any{
				"title":     related.Title,
				"summary":   "shared_impact_or_tag",
				"source":    related.Slug,
				"openedAt":  related.OpenedAt,
				"entityIds": entityIDs,
			},
		})
		result.Items = append(result.Items, rez.AgentWorkflowContextItem{
			Kind:     "domain_reference",
			Role:     "related_incident",
			Name:     related.Title,
			Citation: citation,
			Payload: map[string]any{
				"id":        related.ID,
				"kind":      "incident",
				"title":     related.Title,
				"summary":   "shared_impact_or_tag",
				"source":    related.Slug,
				"openedAt":  related.OpenedAt,
				"entityIds": entityIDs,
			},
		})
	}
	return nil
}

func latestEvidenceTime(refs []incidentContextEvidenceRef) time.Time {
	var latest time.Time
	for _, ref := range refs {
		if ref.ObservedAt.After(latest) {
			latest = ref.ObservedAt
		}
	}
	return latest
}

func stringsFromSignals(signals []string) string {
	result := ""
	for i, signal := range signals {
		if i > 0 {
			result += "\n"
		}
		result += signal
	}
	return result
}

func derefTime(v *time.Time, fallback time.Time) time.Time {
	if v == nil {
		return fallback
	}
	return *v
}
