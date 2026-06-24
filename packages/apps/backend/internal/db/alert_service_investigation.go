package db

import (
	"context"
	"fmt"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	ii "github.com/rezible/rezible/ent/incidentimpact"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

const (
	investigationEvidenceLimit = 20
	investigationNeighborLimit = 30
	investigationGuideLimit    = 5
)

func (s *AlertService) GetInvestigationContext(ctx context.Context, alertID uuid.UUID) (*rez.AgentWorkflowContext, error) {
	a, alertErr := s.db.Client(ctx).Alert.Query().
		Where(alert.ID(alertID)).
		WithKnowledgeEntity(func(q *ent.KnowledgeEntityQuery) {
			q.WithAliases()
		}).
		WithPlaybooks().
		Only(ctx)
	if alertErr != nil {
		return nil, fmt.Errorf("get alert: %w", alertErr)
	}

	result := &rez.AgentWorkflowContext{
		GeneratedAt: time.Now().UTC(),
		Context: map[string]any{
			"alertId":      a.ID,
			"alertTitle":   a.Title,
			"alertSummary": a.Description,
			"definition":   a.Definition,
		},
		PromptSchema: "alert_investigation_context.v1",
		Suggested:    defaultAlertInvestigationChecks(a),
	}
	result.Context["suggestedChecks"] = result.Suggested
	addWorkflowContextCitation(result, rez.AgentRunCitationInput{
		CitationKind:     "primary_subject",
		DomainEntityType: "alert",
		DomainEntityID:   a.ID,
		Summary:          a.Title,
		Snapshot: map[string]any{
			"title":       a.Title,
			"description": a.Description,
			"definition":  a.Definition,
		},
	})

	alertEntity := a.Edges.KnowledgeEntity
	if alertEntity == nil {
		result.Limitations = append(result.Limitations, "Alert has no knowledge entity yet.")
		result.Context["limitations"] = result.Limitations
		return result, nil
	}

	subjectIDs, subjectErr := s.resolveInvestigationSubjects(ctx, result, alertEntity)
	if subjectErr != nil {
		return nil, subjectErr
	}
	if len(subjectIDs) == 0 {
		result.Limitations = append(result.Limitations, "No related component entities were found for this alert.")
		subjectIDs = []uuid.UUID{alertEntity.ID}
	}

	if err := s.addInvestigationNeighbors(ctx, result, subjectIDs); err != nil {
		return nil, err
	}
	if err := s.addInvestigationSignals(ctx, result, append([]uuid.UUID{alertEntity.ID}, subjectIDs...)); err != nil {
		return nil, err
	}
	if err := s.addInvestigationGuides(ctx, result, a, subjectIDs); err != nil {
		return nil, err
	}

	result.Context["limitations"] = result.Limitations
	result.Context["itemCounts"] = countContextItemsByRole(result.Items)
	return result, nil
}

func defaultAlertInvestigationChecks(a *ent.Alert) []string {
	checks := []string{"Confirm whether the alert is still firing and whether customer-facing paths are affected."}
	if a.Definition != "" {
		checks = append(checks, "Inspect the metric behind "+a.Definition+".")
	}
	checks = append(checks,
		"Review recent changes touching the likely affected component.",
		"Check prior incidents and playbooks before declaring mitigation complete.",
	)
	return checks
}

func (s *AlertService) resolveInvestigationSubjects(ctx context.Context, result *rez.AgentWorkflowContext, alertEntity *ent.KnowledgeEntity) ([]uuid.UUID, error) {
	rels, relErr := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityID(alertEntity.ID), knr.TargetEntityID(alertEntity.ID))).
		WithSourceEntity(func(q *ent.KnowledgeEntityQuery) { q.WithAliases() }).
		WithTargetEntity(func(q *ent.KnowledgeEntityQuery) { q.WithAliases() }).
		Limit(investigationNeighborLimit).
		All(ctx)
	if relErr != nil {
		return nil, fmt.Errorf("query alert relationships: %w", relErr)
	}

	type candidate struct {
		entity *ent.KnowledgeEntity
		reason string
	}
	candidates := make(map[uuid.UUID]candidate)
	for _, rel := range rels {
		var entity *ent.KnowledgeEntity
		if rel.SourceEntityID == alertEntity.ID {
			entity = rel.Edges.TargetEntity
		} else if rel.TargetEntityID == alertEntity.ID {
			entity = rel.Edges.SourceEntity
		}
		if entity == nil || entity.Kind == knowledgeEntityKindAlert {
			continue
		}
		candidates[entity.ID] = candidate{
			entity: entity,
			reason: "direct alert relationship: " + rel.Kind,
		}
		addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:            "supporting_evidence",
			KnowledgeRelationshipID: rel.ID,
			Summary:                 "Direct alert relationship: " + rel.Kind,
			Snapshot: map[string]any{
				"kind":             rel.Kind,
				"sourceEntityId":   rel.SourceEntityID,
				"targetEntityId":   rel.TargetEntityID,
				"lastObservedAt":   rel.LastObservedAt,
				"relationshipName": rel.DisplayName,
			},
		})
	}

	ids := make([]uuid.UUID, 0, len(candidates))
	items := make([]rez.AgentWorkflowContextItem, 0, len(candidates))
	for _, cand := range candidates {
		ids = append(ids, cand.entity.ID)
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:      "related_entity",
			KnowledgeEntityID: cand.entity.ID,
			Summary:           cand.entity.DisplayName,
			Snapshot: map[string]any{
				"kind":        cand.entity.Kind,
				"displayName": cand.entity.DisplayName,
				"description": cand.entity.Description,
				"aliases":     investigationAliases(cand.entity),
			},
		})
		items = append(items, rez.AgentWorkflowContextItem{
			Kind:     "knowledge_entity",
			Role:     "likely_subject",
			Name:     cand.entity.DisplayName,
			Citation: citation,
			Payload: map[string]any{
				"entityId":    cand.entity.ID,
				"kind":        cand.entity.Kind,
				"displayName": cand.entity.DisplayName,
				"aliases":     investigationAliases(cand.entity),
				"reason":      cand.reason,
			},
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	result.Items = append(result.Items, items...)
	return ids, nil
}

func investigationAliases(entity *ent.KnowledgeEntity) []string {
	if entity == nil {
		return []string{}
	}
	aliases := make([]string, 0, len(entity.Edges.Aliases))
	for _, alias := range entity.Edges.Aliases {
		aliases = append(aliases, alias.Provider+":"+alias.ProviderSubjectRef)
	}
	sort.Strings(aliases)
	return aliases
}

func (s *AlertService) addInvestigationNeighbors(ctx context.Context, result *rez.AgentWorkflowContext, entityIDs []uuid.UUID) error {
	rels, relErr := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))).
		WithSourceEntity().
		WithTargetEntity().
		Limit(investigationNeighborLimit).
		All(ctx)
	if relErr != nil {
		return fmt.Errorf("query subject neighbors: %w", relErr)
	}

	entitySet := uuidSet(entityIDs)
	for _, rel := range rels {
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:            "related_entity",
			KnowledgeRelationshipID: rel.ID,
			Summary:                 rel.Kind,
			Snapshot: map[string]any{
				"kind":           rel.Kind,
				"sourceEntityId": rel.SourceEntityID,
				"targetEntityId": rel.TargetEntityID,
			},
		})
		if _, ok := entitySet[rel.SourceEntityID]; ok && rel.Edges.TargetEntity != nil {
			result.Items = append(result.Items, relationshipContextItem("outbound", rel, rel.Edges.TargetEntity, citation))
		}
		if _, ok := entitySet[rel.TargetEntityID]; ok && rel.Edges.SourceEntity != nil {
			result.Items = append(result.Items, relationshipContextItem("inbound", rel, rel.Edges.SourceEntity, citation))
		}
	}
	return nil
}

func relationshipContextItem(direction string, rel *ent.KnowledgeRelationship, related *ent.KnowledgeEntity, citation int) rez.AgentWorkflowContextItem {
	return rez.AgentWorkflowContextItem{
		Kind:     "knowledge_relationship",
		Role:     "neighbor",
		Name:     related.DisplayName,
		Citation: citation,
		Payload: map[string]any{
			"relationshipId":    rel.ID,
			"kind":              rel.Kind,
			"direction":         direction,
			"relatedEntityId":   related.ID,
			"relatedEntity":     related.DisplayName,
			"relatedEntityKind": related.Kind,
		},
	}
}

func (s *AlertService) addInvestigationSignals(ctx context.Context, result *rez.AgentWorkflowContext, entityIDs []uuid.UUID) error {
	evidence, evErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.EntityIDIn(entityIDs...)).
		WithEvent().
		Order(knev.ByObservedAt(sql.OrderDesc())).
		Limit(investigationEvidenceLimit).
		All(ctx)
	if evErr != nil {
		return fmt.Errorf("query investigation signals: %w", evErr)
	}

	for _, ev := range evidence {
		source := "knowledge"
		props := map[string]any{}
		if ev.Edges.Event != nil {
			source = ev.Edges.Event.Provider + "/" + ev.Edges.Event.ProviderSource
			props["providerEventRef"] = ev.Edges.Event.ProviderEventRef
			props["subjectKind"] = ev.Edges.Event.SubjectKind
		}
		summary := ev.Assertion + " (" + ev.EvidenceKind.String() + ")"
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:        "supporting_evidence",
			KnowledgeEvidenceID: ev.ID,
			Summary:             summary,
			Snapshot: map[string]any{
				"eventId":    ev.EventID,
				"entityId":   ev.EntityID,
				"source":     source,
				"assertion":  ev.Assertion,
				"observedAt": ev.ObservedAt,
				"properties": props,
			},
		})
		result.Items = append(result.Items, rez.AgentWorkflowContextItem{
			Kind:     "knowledge_evidence",
			Role:     "recent_signal",
			Name:     summary,
			Citation: citation,
			Payload: map[string]any{
				"id":         ev.ID,
				"eventId":    ev.EventID,
				"entityId":   ev.EntityID,
				"source":     source,
				"kind":       ev.Assertion,
				"summary":    summary,
				"observedAt": ev.ObservedAt,
				"properties": props,
			},
		})
	}
	return nil
}

func (s *AlertService) addInvestigationGuides(ctx context.Context, result *rez.AgentWorkflowContext, a *ent.Alert, entityIDs []uuid.UUID) error {
	guides := 0
	for _, playbook := range a.Edges.Playbooks {
		if guides >= investigationGuideLimit {
			break
		}
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:     "operational_guide",
			DomainEntityType: "playbook",
			DomainEntityID:   playbook.ID,
			Summary:          playbook.Title,
			Snapshot: map[string]any{
				"title":   playbook.Title,
				"summary": string(playbook.Content),
				"source":  "linked_alert",
			},
		})
		result.Items = append(result.Items, referenceContextItem("guide", "playbook", playbook.ID, playbook.Title, string(playbook.Content), "linked_alert", citation))
		guides++
	}

	if guides >= investigationGuideLimit || len(entityIDs) == 0 {
		return nil
	}
	impacts, impactsErr := s.db.Client(ctx).IncidentImpact.Query().
		Where(ii.KnowledgeEntityIDIn(entityIDs...)).
		WithIncident().
		Limit(investigationGuideLimit - guides).
		All(ctx)
	if impactsErr != nil {
		return fmt.Errorf("query prior incident guides: %w", impactsErr)
	}
	seenIncidents := make(map[uuid.UUID]struct{})
	for _, impact := range impacts {
		inc := impact.Edges.Incident
		if inc == nil {
			continue
		}
		if _, seen := seenIncidents[inc.ID]; seen {
			continue
		}
		seenIncidents[inc.ID] = struct{}{}
		citation := addWorkflowContextCitation(result, rez.AgentRunCitationInput{
			CitationKind:     "historical_example",
			DomainEntityType: "incident",
			DomainEntityID:   inc.ID,
			Summary:          inc.Title,
			Snapshot: map[string]any{
				"title":   inc.Title,
				"summary": inc.Summary,
				"source":  "shared_impacted_component",
			},
		})
		result.Items = append(result.Items, referenceContextItem("guide", "prior_incident", inc.ID, inc.Title, inc.Summary, "shared_impacted_component", citation))
		guides++
		if guides >= investigationGuideLimit {
			break
		}
	}
	return nil
}

func referenceContextItem(role, kind string, id uuid.UUID, title, summary, source string, citation int) rez.AgentWorkflowContextItem {
	return rez.AgentWorkflowContextItem{
		Kind:     "domain_reference",
		Role:     role,
		Name:     title,
		Citation: citation,
		Payload: map[string]any{
			"id":      id,
			"kind":    kind,
			"title":   title,
			"summary": summary,
			"source":  source,
		},
	}
}

func addWorkflowContextCitation(ctx *rez.AgentWorkflowContext, input rez.AgentRunCitationInput) int {
	if input.Snapshot == nil {
		input.Snapshot = map[string]any{}
	}
	ctx.Citations = append(ctx.Citations, input)
	return len(ctx.Citations)
}

func countContextItemsByRole(items []rez.AgentWorkflowContextItem) map[string]int {
	counts := make(map[string]int)
	for _, item := range items {
		if item.Role != "" {
			counts[item.Role]++
		}
	}
	return counts
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	res := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		res[id] = struct{}{}
	}
	return res
}
