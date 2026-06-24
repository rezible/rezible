package db

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentcaseartifact"
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

func (s *AlertService) GetInvestigationArtifacts(ctx context.Context, alertID uuid.UUID) ([]rez.AgentCaseArtifactInput, error) {
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

	limitations := make([]string, 0)
	artifacts := []rez.AgentCaseArtifactInput{{
		Kind: agentcaseartifact.KindContext,
		Name: "retrieval_context",
		Payload: map[string]any{
			"alertId":         a.ID,
			"generatedAt":     time.Now().UTC(),
			"alertTitle":      a.Title,
			"alertSummary":    a.Description,
			"definition":      a.Definition,
			"suggestedChecks": defaultAlertInvestigationChecks(a),
			"limitations":     limitations,
		},
	}}

	alertEntity := a.Edges.KnowledgeEntity
	if alertEntity == nil {
		artifacts[0].Payload["limitations"] = append(limitations, "Alert has no knowledge entity yet.")
		return artifacts, nil
	}

	subjects, subjectIDs, subjectErr := s.resolveInvestigationSubjects(ctx, alertEntity)
	if subjectErr != nil {
		return nil, subjectErr
	}
	artifacts = append(artifacts, subjects...)
	if len(subjectIDs) == 0 {
		limitations = append(limitations, "No related component entities were found for this alert.")
		artifacts[0].Payload["limitations"] = limitations
		subjectIDs = []uuid.UUID{alertEntity.ID}
	}

	neighbors, neighborErr := s.getInvestigationNeighbors(ctx, subjectIDs)
	if neighborErr != nil {
		return nil, neighborErr
	}
	artifacts = append(artifacts, neighbors...)

	signals, signalErr := s.getInvestigationSignals(ctx, append([]uuid.UUID{alertEntity.ID}, subjectIDs...))
	if signalErr != nil {
		return nil, signalErr
	}
	artifacts = append(artifacts, signals...)

	guides, guideErr := s.getInvestigationGuides(ctx, a, subjectIDs)
	if guideErr != nil {
		return nil, guideErr
	}
	artifacts = append(artifacts, guides...)

	return artifacts, nil
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

func (s *AlertService) resolveInvestigationSubjects(ctx context.Context, alertEntity *ent.KnowledgeEntity) ([]rez.AgentCaseArtifactInput, []uuid.UUID, error) {
	rels, relErr := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityID(alertEntity.ID), knr.TargetEntityID(alertEntity.ID))).
		WithSourceEntity(func(q *ent.KnowledgeEntityQuery) { q.WithAliases() }).
		WithTargetEntity(func(q *ent.KnowledgeEntityQuery) { q.WithAliases() }).
		Limit(investigationNeighborLimit).
		All(ctx)
	if relErr != nil {
		return nil, nil, fmt.Errorf("query alert relationships: %w", relErr)
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
	}

	subjects := make([]rez.AgentCaseArtifactInput, 0, len(candidates))
	subjectIDs := make([]uuid.UUID, 0, len(candidates))
	for _, cand := range candidates {
		subjectIDs = append(subjectIDs, cand.entity.ID)
		subjects = append(subjects, rez.AgentCaseArtifactInput{
			Kind: agentcaseartifact.KindEntityRef,
			Role: "likely_subject",
			Name: cand.entity.DisplayName,
			Payload: agentArtifactPayload(rez.AgentEntityRef{
				EntityID:    cand.entity.ID,
				Kind:        cand.entity.Kind,
				DisplayName: cand.entity.DisplayName,
				Aliases:     investigationAliases(cand.entity),
				Reason:      cand.reason,
				Confidence:  "high",
			}),
		})
	}
	sort.Slice(subjects, func(i, j int) bool {
		return subjects[i].Name < subjects[j].Name
	})
	return subjects, subjectIDs, nil
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

func (s *AlertService) getInvestigationNeighbors(ctx context.Context, entityIDs []uuid.UUID) ([]rez.AgentCaseArtifactInput, error) {
	rels, relErr := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))).
		WithSourceEntity().
		WithTargetEntity().
		Limit(investigationNeighborLimit).
		All(ctx)
	if relErr != nil {
		return nil, fmt.Errorf("query subject neighbors: %w", relErr)
	}

	entitySet := uuidSet(entityIDs)
	neighbors := make([]rez.AgentCaseArtifactInput, 0, len(rels))
	for _, rel := range rels {
		if _, ok := entitySet[rel.SourceEntityID]; ok && rel.Edges.TargetEntity != nil {
			neighbors = append(neighbors, rez.AgentCaseArtifactInput{
				Kind: agentcaseartifact.KindRelationshipRef,
				Role: "neighbor",
				Name: rel.Edges.TargetEntity.DisplayName,
				Payload: agentArtifactPayload(rez.AgentRelationshipRef{
					RelationshipID:    rel.ID,
					Kind:              rel.Kind,
					Summary:           rel.Edges.TargetEntity.DisplayName,
					Direction:         "outbound",
					RelatedEntityID:   rel.TargetEntityID,
					RelatedEntity:     rel.Edges.TargetEntity.DisplayName,
					RelatedEntityKind: rel.Edges.TargetEntity.Kind,
				}),
			})
		}
		if _, ok := entitySet[rel.TargetEntityID]; ok && rel.Edges.SourceEntity != nil {
			neighbors = append(neighbors, rez.AgentCaseArtifactInput{
				Kind: agentcaseartifact.KindRelationshipRef,
				Role: "neighbor",
				Name: rel.Edges.SourceEntity.DisplayName,
				Payload: agentArtifactPayload(rez.AgentRelationshipRef{
					RelationshipID:    rel.ID,
					Kind:              rel.Kind,
					Summary:           rel.Edges.SourceEntity.DisplayName,
					Direction:         "inbound",
					RelatedEntityID:   rel.SourceEntityID,
					RelatedEntity:     rel.Edges.SourceEntity.DisplayName,
					RelatedEntityKind: rel.Edges.SourceEntity.Kind,
				}),
			})
		}
	}
	return neighbors, nil
}

func (s *AlertService) getInvestigationSignals(ctx context.Context, entityIDs []uuid.UUID) ([]rez.AgentCaseArtifactInput, error) {
	evidence, evErr := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(knev.EntityIDIn(entityIDs...)).
		WithEvent().
		Order(knev.ByObservedAt(sql.OrderDesc())).
		Limit(investigationEvidenceLimit).
		All(ctx)
	if evErr != nil {
		return nil, fmt.Errorf("query investigation signals: %w", evErr)
	}

	signals := make([]rez.AgentCaseArtifactInput, 0, len(evidence))
	for _, ev := range evidence {
		source := "knowledge"
		props := map[string]any{}
		if ev.Edges.Event != nil {
			source = ev.Edges.Event.Provider + "/" + ev.Edges.Event.ProviderSource
			props["providerEventRef"] = ev.Edges.Event.ProviderEventRef
			props["subjectKind"] = ev.Edges.Event.SubjectKind
		}
		summary := ev.Assertion + " (" + ev.EvidenceKind.String() + ")"
		signals = append(signals, rez.AgentCaseArtifactInput{
			Kind: agentcaseartifact.KindEvidenceRef,
			Role: "recent_signal",
			Name: summary,
			Payload: agentArtifactPayload(rez.AgentEvidenceRef{
				ID:         ev.ID,
				EventID:    ev.EventID,
				EntityID:   ev.EntityID,
				Source:     source,
				Kind:       ev.Assertion,
				Summary:    summary,
				ObservedAt: ev.ObservedAt,
				Properties: props,
			}),
		})
	}
	return signals, nil
}

func (s *AlertService) getInvestigationGuides(ctx context.Context, a *ent.Alert, entityIDs []uuid.UUID) ([]rez.AgentCaseArtifactInput, error) {
	guides := make([]rez.AgentCaseArtifactInput, 0, investigationGuideLimit)
	for _, playbook := range a.Edges.Playbooks {
		if len(guides) >= investigationGuideLimit {
			break
		}
		guides = append(guides, rez.AgentCaseArtifactInput{
			Kind: agentcaseartifact.KindReferenceRef,
			Role: "guide",
			Name: playbook.Title,
			Payload: agentArtifactPayload(rez.AgentReferenceRef{
				ID:      playbook.ID,
				Kind:    "playbook",
				Title:   playbook.Title,
				Summary: string(playbook.Content),
				Source:  "linked_alert",
			}),
		})
	}

	if len(guides) >= investigationGuideLimit || len(entityIDs) == 0 {
		return guides, nil
	}
	impacts, impactsErr := s.db.Client(ctx).IncidentImpact.Query().
		Where(ii.KnowledgeEntityIDIn(entityIDs...)).
		WithIncident().
		Limit(investigationGuideLimit - len(guides)).
		All(ctx)
	if impactsErr != nil {
		return nil, fmt.Errorf("query prior incident guides: %w", impactsErr)
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
		guides = append(guides, rez.AgentCaseArtifactInput{
			Kind: agentcaseartifact.KindReferenceRef,
			Role: "guide",
			Name: inc.Title,
			Payload: agentArtifactPayload(rez.AgentReferenceRef{
				ID:      inc.ID,
				Kind:    "prior_incident",
				Title:   inc.Title,
				Summary: inc.Summary,
				Source:  "shared_impacted_component",
			}),
		})
		if len(guides) >= investigationGuideLimit {
			break
		}
	}
	return guides, nil
}

func agentArtifactPayload(v any) map[string]any {
	bytes, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return map[string]any{"error": marshalErr.Error()}
	}
	var payload map[string]any
	if unmarshalErr := json.Unmarshal(bytes, &payload); unmarshalErr != nil {
		return map[string]any{"error": unmarshalErr.Error()}
	}
	return payload
}
