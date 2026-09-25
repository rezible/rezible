package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"

	ale "github.com/rezible/rezible/ent/alertepisode"
	ali "github.com/rezible/rezible/ent/alertinstance"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ev "github.com/rezible/rezible/ent/normalizedevent"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
)

const maxSituationEvidenceQueryPageSize = 100

type situationEvidenceEntry struct {
	reference string
	title     string
	body      string
	occurred  time.Time
	eventIDs  []uuid.UUID
}

type situationEntrySubjectAttachment struct {
	id    uuid.UUID
	kind  string
	role  string
	setFn func(*ent.SystemAnalysisEntrySubjectMutation)
}

func (a situationEntrySubjectAttachment) key() string {
	return a.kind + ":" + a.id.String() + ":" + a.role
}

func makeSituationEntrySubjectAttachment(id uuid.UUID, kind string, role string, setFn func(*ent.SystemAnalysisEntrySubjectMutation)) *situationEntrySubjectAttachment {
	setFnWithRole := func(m *ent.SystemAnalysisEntrySubjectMutation) {
		m.SetRole(role)
		if setFn != nil {
			setFn(m)
		}
	}
	return &situationEntrySubjectAttachment{id, kind, role, setFnWithRole}
}

func (s *SituationService) prepareOrRefreshSituationAnalysis(ctx context.Context, analysisID uuid.UUID, directEventIDs, episodeIDs []uuid.UUID) error {
	if analysisID == uuid.Nil {
		return fmt.Errorf("%w: system analysis ID is required", rez.ErrInvalidInput)
	}
	directEventIDs = mapset.NewSet(directEventIDs...).ToSlice()
	episodeIDs = mapset.NewSet(episodeIDs...).ToSlice()
	if validateErr := s.validateSituationSources(ctx, directEventIDs, episodeIDs); validateErr != nil {
		return validateErr
	}

	episodes := make(map[uuid.UUID]*ent.AlertEpisode, len(episodeIDs))
	if len(episodeIDs) > 0 {
		queryEpisodes := s.db.Client(ctx).AlertEpisode.Query().
			Where(ale.IDIn(episodeIDs...)).
			WithAlertDefinition()
		loadedEpisodes, queryErr := queryEpisodes.All(ctx)
		if queryErr != nil {
			return fmt.Errorf("load alert episodes for analysis: %w", queryErr)
		}
		for _, episode := range loadedEpisodes {
			episodes[episode.ID] = episode
		}
	}

	eventsByID := make(map[uuid.UUID]*ent.NormalizedEvent)
	for _, eventID := range directEventIDs {
		eventsByID[eventID] = nil
	}
	episodeEvents := make(map[uuid.UUID][]uuid.UUID, len(episodeIDs))
	if len(episodeIDs) > 0 {
		queryInstances := s.db.Client(ctx).AlertInstance.Query().
			Where(ali.AlertEpisodeIDIn(episodeIDs...)).
			Order(ali.ByID(sql.OrderAsc()))
		instances, queryErr := queryInstances.All(ctx)
		if queryErr != nil {
			return fmt.Errorf("load alert episode instances for analysis: %w", queryErr)
		}
		for _, instance := range instances {
			episodeEvents[instance.AlertEpisodeID] = append(episodeEvents[instance.AlertEpisodeID], instance.NormalizedEventID)
			eventsByID[instance.NormalizedEventID] = nil
		}
	}
	allEventIDs := mapset.NewSetFromMapKeys(eventsByID)
	if !allEventIDs.IsEmpty() {
		queryEvents := s.db.Client(ctx).NormalizedEvent.Query().
			Where(ev.IDIn(allEventIDs.ToSlice()...))
		events, queryErr := queryEvents.All(ctx)
		if queryErr != nil {
			return fmt.Errorf("load normalized events for analysis: %w", queryErr)
		}
		for _, event := range events {
			eventsByID[event.ID] = event
		}
		if len(events) != allEventIDs.Cardinality() {
			return fmt.Errorf("%w: normalized event from situation evidence is unavailable", rez.ErrNotFound)
		}
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

	entries := make([]situationEvidenceEntry, 0, len(directEventIDs)+len(episodeIDs))
	for _, eventID := range directEventIDs {
		event := eventsByID[eventID]
		entries = append(entries, situationEvidenceEntry{
			reference: "normalized_event:" + event.ID.String(),
			title:     event.Kind + " event",
			body:      fmt.Sprintf("provider_event_source: %s\nprovider_event_ref: %s", event.ProviderEventSource, event.ProviderEventRef),
			occurred:  event.OccurredAt,
			eventIDs:  []uuid.UUID{event.ID},
		})
	}
	for _, episodeID := range episodeIDs {
		episode := episodes[episodeID]
		definition, definitionErr := episode.Edges.AlertDefinitionOrErr()
		if definitionErr != nil {
			return fmt.Errorf("load alert episode definition: %w", definitionErr)
		}
		entries = append(entries, situationEvidenceEntry{
			reference: "alert_episode:" + episode.ID.String(),
			title:     definition.Title,
			body:      fmt.Sprintf("episode_id: %s\nstatus: %s", episode.ID, episode.Status),
			occurred:  episode.StartedAt,
			eventIDs:  mapset.NewSet(episodeEvents[episodeID]...).ToSlice(),
		})
	}

	for _, entry := range entries {
		attachments := make(map[string]situationEntrySubjectAttachment)
		appendAttachment := func(id uuid.UUID, kind string, role string, setFn func(*ent.SystemAnalysisEntrySubjectMutation)) {
			a := makeSituationEntrySubjectAttachment(id, kind, role, setFn)
			if _, exists := attachments[a.key()]; !exists {
				attachments[a.key()] = *a
			}
		}
		for _, eventID := range entry.eventIDs {
			for _, evidence := range gctx.evidenceByEvent[eventID] {
				evidenceID := evidence.ID
				appendAttachment(evidenceID, "evidence", "knowledge_evidence", func(m *ent.SystemAnalysisEntrySubjectMutation) {
					m.SetKnowledgeEvidenceID(evidenceID)
				})
			}

			if eventEntityIds := gctx.entitiesByEvent[eventID]; eventEntityIds != nil {
				for _, entityID := range eventEntityIds.ToSlice() {
					subjectID := entityID
					appendAttachment(subjectID, "knowledge_entity", "subject", func(m *ent.SystemAnalysisEntrySubjectMutation) {
						m.SetKnowledgeEntityID(subjectID)
					})
				}
			}

			if eventRelIds := gctx.relationshipsByEvent[eventID]; eventRelIds != nil {
				for _, relationshipID := range eventRelIds.ToSlice() {
					subjectID := relationshipID
					appendAttachment(subjectID, "knowledge_relationship", "subject", func(m *ent.SystemAnalysisEntrySubjectMutation) {
						m.SetKnowledgeRelationshipID(subjectID)
					})
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
				var subjAtt *situationEntrySubjectAttachment
				if attch.KnowledgeEntityID != nil {
					subjAtt = makeSituationEntrySubjectAttachment(*attch.KnowledgeEntityID, "knowledge_entity", attch.Role, nil)
				}
				if attch.KnowledgeRelationshipID != nil {
					subjAtt = makeSituationEntrySubjectAttachment(*attch.KnowledgeRelationshipID, "knowledge_relationship", attch.Role, nil)
				}
				if attch.KnowledgeEvidenceID != nil {
					subjAtt = makeSituationEntrySubjectAttachment(*attch.KnowledgeEvidenceID, "knowledge_evidence", attch.Role, nil)
				}
				if subjAtt != nil {
					existingAttachmentKeys.Add(subjAtt.key())
				}
			}
		}

		newAttachments := make([]func(*ent.SystemAnalysisEntrySubjectMutation), 0, len(attachments))
		for _, att := range attachments {
			if !existingAttachmentKeys.Contains(att.key()) {
				newAttachments = append(newAttachments, att.setFn)
			}
		}

		occurredAt := entry.occurred
		setEntryFn := func(m *ent.SystemAnalysisEntryMutation) {
			if entryID == uuid.Nil {
				m.SetAnalysisID(analysisID)
				m.SetReference(entry.reference)
				m.SetKind(sae.KindObservation)
			}
			m.SetTitle(entry.title)
			m.SetBody(entry.body)
			m.SetOccurredAt(occurredAt)
		}
		_, saveErr := s.analyses.SetSystemAnalysisEntry(ctx, entryID, setEntryFn, newAttachments...)
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
