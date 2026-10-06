package db

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"

	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ev "github.com/rezible/rezible/ent/normalizedevent"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	"github.com/rezible/rezible/pkg/situations"
)

type situationEvidenceEntry struct {
	reference string
	title     string
	body      string
	occurred  time.Time
	eventIDs  []uuid.UUID
}

func makeSetSituationEntrySubject(id uuid.UUID, kind string, role string) (string, rez.SetSystemAnalysisEntrySubjectParams) {
	params := rez.SetSystemAnalysisEntrySubjectParams{Role: role}
	if kind == "entity" {
		params.KnowledgeEntityID = &id
	} else if kind == "relationship" {
		params.KnowledgeRelationshipID = &id
	} else if kind == "evidence" {
		params.KnowledgeEvidenceID = &id
	}
	return kind + ":" + id.String() + ":" + role, params
}

const maxSituationEvidenceQueryPageSize = 100

// createInvestigation starts the situation's investigation over a system analysis prepared from its
// signals' events, and records that it saw each signal's current revision.
func (s *SituationService) createInvestigation(ctx context.Context, situationID uuid.UUID) error {
	memberIDs, membersErr := s.memberEntityIDs(ctx, situationID)
	if membersErr != nil {
		return membersErr
	}
	signals, loadErr := s.loadSignals(ctx, memberIDs, situations.LoadSignalsOptions{})
	if loadErr != nil {
		return loadErr
	}
	eventIDs := mapset.NewSet[uuid.UUID]()
	for _, signal := range signals {
		eventIDs.Append(signal.EventIDs...)
	}
	analysis, analysisErr := s.analyses.SetSystemAnalysis(ctx, uuid.Nil, func(*ent.SystemAnalysisMutation) {})
	if analysisErr != nil {
		return fmt.Errorf("create situation investigation analysis: %w", analysisErr)
	}
	if prepareErr := s.prepareOrRefreshSituationAnalysis(ctx, analysis.ID, eventIDs.ToSlice()); prepareErr != nil {
		return fmt.Errorf("prepare situation investigation analysis: %w", prepareErr)
	}
	investigationParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      defaultSituationInvestigationQuestion,
	}
	investigation, investigationErr := s.investigations.CreateInvestigation(ctx, investigationParams)
	if investigationErr != nil {
		return fmt.Errorf("create situation investigation: %w", investigationErr)
	}
	createLink := s.db.Client(ctx).SituationInvestigation.Create().
		SetSituationID(situationID).
		SetInvestigationID(investigation.ID)
	if linkErr := createLink.Exec(ctx); linkErr != nil {
		return fmt.Errorf("link situation investigation: %w", linkErr)
	}
	return s.observeRevisions(ctx, situationID, slices.Collect(maps.Values(signals)))
}

// prepareOrRefreshSituationAnalysis adds or refreshes one observation per event in the analysis, attaching
// the graph subjects linked to the event through its knowledge evidence.
func (s *SituationService) prepareOrRefreshSituationAnalysis(ctx context.Context, analysisID uuid.UUID, eventIDs []uuid.UUID) error {
	if analysisID == uuid.Nil {
		return fmt.Errorf("%w: system analysis ID is required", rez.ErrInvalidInput)
	}
	allEventIDs := mapset.NewSet(eventIDs...)
	if allEventIDs.IsEmpty() {
		return nil
	}
	queryEvents := s.db.Client(ctx).NormalizedEvent.Query().
		Where(ev.IDIn(allEventIDs.ToSlice()...)).
		Order(ev.ByOccurredAt(), ev.ByID())
	events, queryErr := queryEvents.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("load normalized events for analysis: %w", queryErr)
	}
	if len(events) != allEventIDs.Cardinality() {
		return fmt.Errorf("%w: normalized event from situation evidence is unavailable", rez.ErrNotFound)
	}

	gctx, graphErr := s.loadSituationEventGraphContext(ctx, allEventIDs)
	if graphErr != nil {
		return graphErr
	}

	entityIDs := gctx.entityIDs.ToSlice()
	relationshipIDs := gctx.relationshipIDs.ToSlice()
	if len(entityIDs)+len(relationshipIDs) > 0 {
		includeParams := rez.IncludeSystemAnalysisSubjectsParams{
			AnalysisId:      analysisID,
			EntityIds:       entityIDs,
			RelationshipIds: relationshipIDs,
		}
		if includeErr := s.analyses.IncludeSystemAnalysisSubjects(ctx, includeParams); includeErr != nil {
			return fmt.Errorf("include evidence-linked graph subjects in analysis: %w", includeErr)
		}
	}

	entries := make([]situationEvidenceEntry, 0, len(events))
	for _, event := range events {
		entries = append(entries, situationEvidenceEntry{
			reference: "normalized_event:" + event.ID.String(),
			title:     event.Kind + " event",
			body:      fmt.Sprintf("provider_event_source: %s\nprovider_event_ref: %s", event.ProviderEventSource, event.ProviderEventRef),
			occurred:  event.OccurredAt,
			eventIDs:  []uuid.UUID{event.ID},
		})
	}

	for _, entry := range entries {
		attachments := make(map[string]rez.SetSystemAnalysisEntrySubjectParams)
		for _, eventID := range entry.eventIDs {
			for _, evidence := range gctx.evidenceByEvent[eventID] {
				key, sp := makeSetSituationEntrySubject(evidence.ID, "evidence", "supports")
				attachments[key] = sp
			}

			if eventEntityIds := gctx.entitiesByEvent[eventID]; eventEntityIds != nil {
				for _, entityID := range eventEntityIds.ToSlice() {
					key, sp := makeSetSituationEntrySubject(entityID, "entity", "context")
					attachments[key] = sp
				}
			}

			if eventRelIds := gctx.relationshipsByEvent[eventID]; eventRelIds != nil {
				for _, relationshipID := range eventRelIds.ToSlice() {
					key, sp := makeSetSituationEntrySubject(relationshipID, "relationship", "context")
					attachments[key] = sp
				}
			}
		}

		queryEntry := s.db.Client(ctx).SystemAnalysisEntry.Query().
			Where(sae.AnalysisID(analysisID), sae.Reference(entry.reference))
		existingEntry, lookupErr := queryEntry.Only(ctx)
		if lookupErr != nil && !ent.IsNotFound(lookupErr) {
			return fmt.Errorf("lookup existing situation analysis observation: %w", lookupErr)
		}

		entryID := uuid.Nil
		existingAttachmentKeys := mapset.NewSet[string]()
		if existingEntry != nil {
			entryID = existingEntry.ID
			queryExistingAttachments := s.db.Client(ctx).SystemAnalysisEntrySubject.Query().
				Where(saes.EntryID(existingEntry.ID))
			storedAttachments, attachmentsErr := queryExistingAttachments.All(ctx)
			if attachmentsErr != nil {
				return fmt.Errorf("load existing situation analysis attachments: %w", attachmentsErr)
			}
			for _, attch := range storedAttachments {
				var subjKey string
				if attch.KnowledgeEntityID != nil {
					subjKey, _ = makeSetSituationEntrySubject(*attch.KnowledgeEntityID, "entity", attch.Role)
				}
				if attch.KnowledgeRelationshipID != nil {
					subjKey, _ = makeSetSituationEntrySubject(*attch.KnowledgeRelationshipID, "relationship", attch.Role)
				}
				if attch.KnowledgeEvidenceID != nil {
					subjKey, _ = makeSetSituationEntrySubject(*attch.KnowledgeEvidenceID, "evidence", attch.Role)
				}
				if subjKey != "" {
					existingAttachmentKeys.Add(subjKey)
				}
			}
		}

		setSubjects := make([]rez.SetSystemAnalysisEntrySubjectParams, 0, len(attachments))
		for key, sp := range attachments {
			if !existingAttachmentKeys.Contains(key) {
				setSubjects = append(setSubjects, sp)
			}
		}

		setEntryParams := rez.SetSystemAnalysisEntryParams{
			AnalysisID:  analysisID,
			Reference:   &entry.reference,
			Kind:        sae.KindObservation,
			Title:       entry.title,
			Body:        entry.body,
			OccurredAt:  &entry.occurred,
			SetSubjects: setSubjects,
		}
		_, saveErr := s.analyses.SetSystemAnalysisEntry(ctx, entryID, setEntryParams)
		if saveErr != nil {
			return fmt.Errorf("save situation analysis observation: %w", saveErr)
		}
	}
	return nil
}

type situationEventGraphContext struct {
	evidenceByEvent      map[uuid.UUID][]*ent.KnowledgeEvidence
	entitiesByEvent      map[uuid.UUID]mapset.Set[uuid.UUID]
	relationshipsByEvent map[uuid.UUID]mapset.Set[uuid.UUID]
	entityIDs            mapset.Set[uuid.UUID]
	relationshipIDs      mapset.Set[uuid.UUID]
}

func (c *situationEventGraphContext) ensureAdd(sets map[uuid.UUID]mapset.Set[uuid.UUID], eventID uuid.UUID, ids ...uuid.UUID) int {
	set, exists := sets[eventID]
	if !exists {
		set = mapset.NewSet[uuid.UUID]()
		sets[eventID] = set
	}
	return set.Append(ids...)
}

func (s *SituationService) loadSituationEventGraphContext(ctx context.Context, eventIDs mapset.Set[uuid.UUID]) (*situationEventGraphContext, error) {
	result := &situationEventGraphContext{
		evidenceByEvent:      make(map[uuid.UUID][]*ent.KnowledgeEvidence),
		entitiesByEvent:      make(map[uuid.UUID]mapset.Set[uuid.UUID]),
		relationshipsByEvent: make(map[uuid.UUID]mapset.Set[uuid.UUID]),
		entityIDs:            mapset.NewSet[uuid.UUID](),
		relationshipIDs:      mapset.NewSet[uuid.UUID](),
	}
	if eventIDs.IsEmpty() {
		return result, nil
	}
	var evidence []*ent.KnowledgeEvidence
	for page := 1; ; page++ {
		listEvidenceParams := rez.ListKnowledgeEvidenceParams{
			ListParams: ent.ListParams{
				Page:     page,
				PageSize: maxSituationEvidenceQueryPageSize,
				OrderAsc: true,
			},
			Predicates: []predicate.KnowledgeEvidence{knev.EventIDIn(eventIDs.ToSlice()...)},
		}
		listed, listErr := s.graph.ListEvidence(ctx, listEvidenceParams)
		if listErr != nil {
			return nil, fmt.Errorf("list event knowledge evidence for situation analysis: %w", listErr)
		}
		evidence = append(evidence, listed.Data...)
		if len(evidence) >= listed.Total || len(listed.Data) == 0 {
			break
		}
	}

	aliasIDsSet := mapset.NewSet[uuid.UUID]()
	for _, item := range evidence {
		result.evidenceByEvent[item.EventID] = append(result.evidenceByEvent[item.EventID], item)
		aliasIDsSet.Add(item.SubjectAliasID)
	}
	aliases := make(map[uuid.UUID]*ent.KnowledgeSubjectAlias, aliasIDsSet.Cardinality())
	aliasIDs := aliasIDsSet.ToSlice()
	for offset := 0; offset < len(aliasIDs); offset += maxSituationEvidenceQueryPageSize {
		end := offset + maxSituationEvidenceQueryPageSize
		if end > len(aliasIDs) {
			end = len(aliasIDs)
		}
		batch := aliasIDs[offset:end]
		listed, listErr := s.graph.ListSubjectAliases(ctx, rez.ListKnowledgeSubjectAliasesParams{
			ListParams: ent.ListParams{Page: 1, PageSize: len(batch), OrderAsc: true},
			Predicates: []predicate.KnowledgeSubjectAlias{
				ksa.IDIn(batch...),
			},
		})
		if listErr != nil {
			return nil, fmt.Errorf("list subject aliases for situation evidence: %w", listErr)
		}
		for _, alias := range listed.Data {
			aliases[alias.ID] = alias
		}
	}

	client := s.db.Client(ctx)

	entityCache := make(map[uuid.UUID]bool)
	relationshipCache := make(map[uuid.UUID]bool)
	relationshipEndpoints := make(map[uuid.UUID][2]uuid.UUID)
	for _, item := range evidence {
		alias, exists := aliases[item.SubjectAliasID]
		if !exists {
			slog.WarnContext(ctx, "situation analysis skipped missing evidence alias",
				"knowledgeEvidenceId", item.ID,
				"subjectAliasId", item.SubjectAliasID)
			continue
		}
		switch alias.SubjectKind {
		case ksa.SubjectKindEntity:
			if alias.EntityID == nil {
				slog.WarnContext(ctx, "situation analysis skipped empty entity alias",
					"knowledgeEvidenceId", item.ID,
					"subjectAliasId", alias.ID)
				continue
			}
			entityID := *alias.EntityID
			if !entityCache[entityID] {
				lookupEntity := client.KnowledgeEntity.Query().
					Where(kne.ID(entityID))
				if entityExists, lookupEntityErr := lookupEntity.Exist(ctx); lookupEntityErr != nil {
					return nil, fmt.Errorf("load evidence-linked entity: %w", lookupEntityErr)
				} else if !entityExists {
					slog.WarnContext(ctx, "situation analysis skipped inaccessible evidence entity",
						"knowledgeEvidenceId", item.ID,
						"entityId", entityID)
					entityCache[entityID] = false
					continue
				}
				entityCache[entityID] = true
			}
			if !entityCache[entityID] {
				continue
			}
			result.ensureAdd(result.entitiesByEvent, item.EventID, entityID)
			result.entityIDs.Add(entityID)
		case ksa.SubjectKindRelationship:
			if alias.RelationshipID == nil {
				slog.WarnContext(ctx, "situation analysis skipped empty relationship alias",
					"knowledgeEvidenceId", item.ID,
					"subjectAliasId", alias.ID)
				continue
			}
			relationshipID := *alias.RelationshipID
			if !relationshipCache[relationshipID] {
				queryRelationship := client.KnowledgeRelationship.Query().
					Where(knr.ID(relationshipID)).
					WithSourceEntity().
					WithTargetEntity()
				relationship, queryErr := queryRelationship.Only(ctx)
				if queryErr != nil {
					if ent.IsNotFound(queryErr) {
						slog.WarnContext(ctx, "situation analysis skipped inaccessible evidence relationship",
							"knowledgeEvidenceId", item.ID,
							"relationshipId", relationshipID)
						relationshipCache[relationshipID] = false
						continue
					}
					return nil, fmt.Errorf("load evidence-linked relationship: %w", queryErr)
				}

				source, sourceErr := relationship.Edges.SourceEntityOrErr()
				target, targetErr := relationship.Edges.TargetEntityOrErr()
				if sourceErr != nil || targetErr != nil {
					slog.WarnContext(ctx, "situation analysis skipped relationship with unavailable endpoint",
						"knowledgeEvidenceId", item.ID,
						"relationshipId", relationshipID)
					relationshipCache[relationshipID] = false
					continue
				}
				result.entityIDs.Append(source.ID, target.ID)
				relationshipEndpoints[relationshipID] = [2]uuid.UUID{source.ID, target.ID}
				relationshipCache[relationshipID] = true
			}
			if !relationshipCache[relationshipID] {
				continue
			}
			endpoints := relationshipEndpoints[relationshipID]
			result.ensureAdd(result.relationshipsByEvent, item.EventID, relationshipID)
			result.ensureAdd(result.entitiesByEvent, item.EventID, endpoints[0], endpoints[1])
			result.relationshipIDs.Add(relationshipID)
		default:
			slog.WarnContext(ctx, "situation analysis skipped unknown evidence alias kind",
				"knowledgeEvidenceId", item.ID,
				"subjectAliasId", alias.ID,
				"subjectKind", alias.SubjectKind)
		}
	}
	return result, nil
}
