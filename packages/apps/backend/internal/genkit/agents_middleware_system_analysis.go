package genkit

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	rezai "github.com/rezible/rezible/pkg/ai"
)

const (
	systemAnalysisPageSize           = 20
	systemAnalysisInstructionsMarker = "system-analysis-instructions"
	systemAnalysisInstructions       = `<system-analysis>
A system analysis is attached to this session. Use its tools to gather graph context and preserve durable conclusions.
- When graph context is useful, start with summarize_system_neighborhood. Each exploration call is exactly one hop.
- Request individual neighbors by direction and relationship kind, and follow pagination only while the remaining results are relevant.
- Use inspect_knowledge_subject before relying on a subject's detailed state or evidence.
- Exploration is read-only. Use include_analysis_subjects to include only entities and relationships that are relevant to the analysis; including a relationship also includes its endpoint entities.
- Use record_analysis_finding for concise, evidence-backed conclusions. Every finding must cite at least one inspected evidence record; do not record unconfirmed hypotheses as findings.
</system-analysis>`
)

func WithSystemAnalysisAgentMiddleware(sa rez.SystemAnalysisService, kg rez.KnowledgeGraphService) AgentMiddlewareConstructorFn {
	return func(string) ai.Middleware {
		return newSystemAnalysisMiddleware(sa, kg)
	}
}

type systemAnalysisMiddleware struct {
	analyses  rez.SystemAnalysisService
	knowledge rez.KnowledgeGraphService
}

func newSystemAnalysisMiddleware(analyses rez.SystemAnalysisService, knowledge rez.KnowledgeGraphService) *systemAnalysisMiddleware {
	return &systemAnalysisMiddleware{analyses: analyses, knowledge: knowledge}
}

func (m *systemAnalysisMiddleware) Name() string {
	return "system_analysis"
}

func (m *systemAnalysisMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	if _, analysisErr := m.getSessionSystemAnalysisID(ctx); analysisErr != nil {
		return nil, analysisErr
	}
	return &ai.Hooks{
		WrapGenerate: makeSystemTextInjectorFn(systemAnalysisInstructionsMarker, systemAnalysisInstructions),
		Tools: []ai.Tool{
			makeDefinedTool(rezai.SummarizeSystemNeighborhoodTool, m.summarizeSystemNeighborhoodToolFunc),
			makeDefinedTool(rezai.ExploreSystemNeighborhoodTool, m.exploreSystemEntityNeighborhoodToolFunc),
			makeDefinedTool(rezai.InspectKnowledgeSubjectTool, m.inspectKnowledgeSubjectToolFunc),
			makeDefinedTool(rezai.IncludeAnalysisSubjectsTool, m.includeAnalysisSubjectsToolFunc),
			makeDefinedTool(rezai.RecordAnalysisFindingTool, m.recordAnalysisFindingToolFunc),
		},
	}, nil
}

func (m *systemAnalysisMiddleware) getSessionSystemAnalysisID(ctx context.Context) (uuid.UUID, error) {
	aic, ctxOk := getAgentInvocationContext(ctx)
	if !ctxOk || aic == nil || aic.Session == nil {
		return uuid.Nil, fmt.Errorf("agent session context does not exist")
	}
	if aic.Session.SystemAnalysisID == nil || *aic.Session.SystemAnalysisID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("system analysis middleware requires an invocation session with a system analysis")
	}
	return *aic.Session.SystemAnalysisID, nil
}

func (m *systemAnalysisMiddleware) resolveExplorationEntityId(ctx context.Context, entityID *string) (uuid.UUID, error) {
	analysisID, analysisIDErr := m.getSessionSystemAnalysisID(ctx)
	if analysisIDErr != nil {
		return uuid.Nil, fmt.Errorf("load analysis: %w", analysisIDErr)
	}

	analysis, analysisErr := m.analyses.GetSystemAnalysis(ctx, analysisID)
	if analysisErr != nil {
		return uuid.Nil, fmt.Errorf("get system analysis subject: %w", analysisErr)
	}

	var rootID uuid.UUID
	if entityID != nil {
		id, idErr := uuid.Parse(*entityID)
		if idErr != nil {
			return uuid.Nil, fmt.Errorf("%w: invalid entity id: %s", rez.ErrInvalidInput, idErr)
		}
		rootID = id
	} else if analysis.SubjectEntityID != nil {
		rootID = *analysis.SubjectEntityID
	}

	if rootID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: root entity ID is required", rez.ErrInvalidInput)
	} else if analysis.SubjectEntityID != nil && rootID == *analysis.SubjectEntityID {
		return rootID, nil
	}

	included, queryErr := m.analyses.HasSystemAnalysisEntity(ctx, analysis.ID, rootID)
	if !included {
		if queryErr != nil {
			return uuid.Nil, fmt.Errorf("check analysis entity: %w", queryErr)
		}
		return uuid.Nil, fmt.Errorf("%w: entity %s is neither the analysis subject nor an included entity", rez.ErrInvalidInput, rootID)
	}

	return rootID, nil
}

func (m *systemAnalysisMiddleware) summarizeSystemNeighborhoodToolFunc(ctx context.Context, input rezai.SummarizeSystemNeighborhoodToolInput) (*rezai.SummarizeSystemNeighborhoodToolOutput, error) {
	entityId, entityErr := m.resolveExplorationEntityId(ctx, input.EntityID)
	if entityErr != nil {
		return nil, entityErr
	}

	summary, summaryErr := m.knowledge.SummarizeEntityNeighborhood(ctx, entityId)
	if summaryErr != nil {
		return nil, fmt.Errorf("summarize entity neighborhood: %w", summaryErr)
	}

	output := rezai.SummarizeSystemNeighborhoodToolOutput{
		IncomingRelationships: make(map[string]rezai.SystemNeighborhoodGroupSummary, len(summary.IncomingRelationships)),
		OutgoingRelationships: make(map[string]rezai.SystemNeighborhoodGroupSummary, len(summary.OutgoingRelationships)),
	}

	for kind, group := range summary.IncomingRelationships {
		output.IncomingRelationships[kind] = rezai.SystemNeighborhoodGroupSummary{
			Count: group.Count,
		}
	}
	for kind, group := range summary.OutgoingRelationships {
		output.OutgoingRelationships[kind] = rezai.SystemNeighborhoodGroupSummary{
			Count: group.Count,
		}
	}

	return &output, nil
}

func (m *systemAnalysisMiddleware) exploreSystemEntityNeighborhoodToolFunc(ctx context.Context, input rezai.ExploreSystemNeighborhoodToolInput) (*rezai.ExploreSystemNeighborhoodToolOutput, error) {
	entityId, entityErr := m.resolveExplorationEntityId(ctx, input.EntityID)
	if entityErr != nil {
		return nil, entityErr
	}

	var offset int
	if input.Offset != nil {
		offset = *input.Offset
	}

	params := rez.QueryKnowledgeEntityNeighborhoodParams{
		EntityID:            &entityId,
		SourceEntityID:      nil,
		TargetEntityID:      nil,
		NeighborEntityKinds: nil,
		RelationshipKinds:   nil,
		Depth:               1,
		Offset:              offset,
	}
	if input.NeighborKind != nil {
		params.NeighborEntityKinds = []string{*input.NeighborKind}
	}
	if input.RelationshipKind != nil {
		params.RelationshipKinds = []string{*input.RelationshipKind}
	}

	neighborhood, neighborhoodErr := m.knowledge.QueryEntityNeighborhood(ctx, params)
	if neighborhoodErr != nil {
		return nil, fmt.Errorf("query entity neighborhood: %w", neighborhoodErr)
	}

	entitySummaryMap := make(map[uuid.UUID]rezai.KnowledgeEntitySummary)
	for _, e := range neighborhood.Entities {
		if e.ID == entityId {
			continue
		}
		if _, seen := entitySummaryMap[e.ID]; !seen {
			entitySummaryMap[e.ID] = entitySummary(e)
		}
	}

	matches := make([]rezai.SystemEntityNeighbor, 0, len(neighborhood.Relationships))
	for _, rel := range neighborhood.Relationships {
		neighborId := rel.SourceEntityID
		if neighborId == entityId {
			neighborId = rel.TargetEntityID
		}
		entSum, summaryOk := entitySummaryMap[neighborId]
		if !summaryOk {
			slog.Warn("missing entity result from neighborhood", "id", neighborId)
			continue
		}
		matches = append(matches, rezai.SystemEntityNeighbor{
			Entity:       entSum,
			Relationship: relationshipSummary(rel),
		})
	}

	output := &rezai.ExploreSystemNeighborhoodToolOutput{
		Neighbours: matches,
	}
	if offset+len(matches) < neighborhood.RelationshipCount {
		output.NextOffset = new(offset + len(matches))
	}
	return output, nil
}

func (m *systemAnalysisMiddleware) inspectKnowledgeSubjectToolFunc(ctx context.Context, input rezai.InspectKnowledgeSubjectToolInput) (*rezai.InspectKnowledgeSubjectToolOutput, error) {
	id, idErr := uuid.Parse(input.SubjectID)
	if idErr != nil {
		return nil, fmt.Errorf("%w: subject ID is required", rez.ErrInvalidInput)
	}
	var output rezai.InspectKnowledgeSubjectToolOutput
	switch strings.TrimSpace(input.SubjectKind) {
	case "entity":
		entity, getErr := m.knowledge.GetEntity(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("get entity: %w", getErr)
		}
		output.Entity = entityDetail(entity)
	case "relationship":
		relationship, getErr := m.knowledge.GetRelationship(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("get relationship: %w", getErr)
		}
		detail, detailErr := relationshipDetail(relationship)
		if detailErr != nil {
			return nil, fmt.Errorf("get relationship detail: %w", detailErr)
		}
		output.Relationship = detail
	case "evidence":
		evidence, getErr := m.knowledge.GetEvidence(ctx, id)
		if getErr != nil {
			return nil, fmt.Errorf("get evidence: %w", getErr)
		}
		detail, conversionErr := evidenceDetail(evidence)
		if conversionErr != nil {
			return nil, conversionErr
		}
		output.Evidence = detail
	default:
		return nil, fmt.Errorf("%w: subject_kind must be entity, relationship, or evidence", rez.ErrInvalidInput)
	}
	return &output, nil
}

func (m *systemAnalysisMiddleware) includeAnalysisSubjectsToolFunc(ctx context.Context, input rezai.IncludeAnalysisSubjectsToolInput) (*rezai.IncludeAnalysisSubjectsToolOutput, error) {
	if len(input.Subjects) < 1 || len(input.Subjects) > systemAnalysisPageSize {
		return nil, fmt.Errorf("%w: between 1 and %d subjects are required", rez.ErrInvalidInput, systemAnalysisPageSize)
	}
	entityIDs := mapset.NewSet[uuid.UUID]()
	relationshipIDs := mapset.NewSet[uuid.UUID]()
	for _, subject := range input.Subjects {
		kind := strings.TrimSpace(subject.SubjectKind)
		id, idErr := uuid.Parse(subject.SubjectID)
		if idErr != nil {
			return nil, fmt.Errorf("%w: subject ID is required", rez.ErrInvalidInput)
		}
		if kind == "entity" {
			if !entityIDs.Add(id) {
				return nil, fmt.Errorf("duplicate entity id %s", subject.SubjectID)
			}
		} else if kind == "relationship" {
			if !relationshipIDs.Add(id) {
				return nil, fmt.Errorf("duplicate relationship id %s", subject.SubjectID)
			}
		} else {
			return nil, fmt.Errorf("%w: subject_kind must be entity or relationship", rez.ErrInvalidInput)
		}
	}

	analysisID, analysisIDErr := m.getSessionSystemAnalysisID(ctx)
	if analysisIDErr != nil {
		return nil, fmt.Errorf("loading session analysis: %w", analysisIDErr)
	}

	includeParams := rez.IncludeSystemAnalysisSubjectsParams{
		AnalysisId:      analysisID,
		EntityIds:       entityIDs.ToSlice(),
		RelationshipIds: relationshipIDs.ToSlice(),
	}
	createErr := m.analyses.IncludeSystemAnalysisSubjects(ctx, includeParams)
	if createErr != nil {
		return nil, fmt.Errorf("include subjects: %w", createErr)
	}

	return &rezai.IncludeAnalysisSubjectsToolOutput{
		Included: entityIDs.Cardinality() + relationshipIDs.Cardinality(),
	}, nil
}

func (m *systemAnalysisMiddleware) recordAnalysisFindingToolFunc(ctx context.Context, input rezai.RecordAnalysisFindingToolInput) (*rezai.RecordAnalysisFindingToolOutput, error) {
	title, detail := strings.TrimSpace(input.Title), strings.TrimSpace(input.Detail)
	if title == "" {
		return nil, fmt.Errorf("%w: finding title is required", rez.ErrInvalidInput)
	}

	ref := strings.TrimSpace(input.Reference)
	if ref == "" {
		return nil, fmt.Errorf("%w: finding reference is required", rez.ErrInvalidInput)
	}

	setSubjects, subjectSettersErr := m.makeSubjectSetters(input.Subjects)
	if subjectSettersErr != nil {
		return nil, fmt.Errorf("include subject setters: %w", subjectSettersErr)
	}

	analysisID, analysisIDErr := m.getSessionSystemAnalysisID(ctx)
	if analysisIDErr != nil {
		return nil, fmt.Errorf("loading session analysis: %w", analysisIDErr)
	}

	existingId, existingErr := m.getExistingFindingId(ctx, analysisID, ref)
	if existingErr != nil {
		return nil, fmt.Errorf("finding existing finding: %w", existingErr)
	}

	setEntry := func(m *ent.SystemAnalysisEntryMutation) {
		m.SetAnalysisID(analysisID)
		m.SetReference(ref)
		m.SetKind(sae.KindFinding)
		m.SetTitle(title)
		m.SetBody(detail)
	}
	entry, createErr := m.analyses.SetSystemAnalysisEntry(ctx, existingId, setEntry, setSubjects...)
	if createErr != nil {
		return nil, fmt.Errorf("record analysis finding: %w", createErr)
	}
	return &rezai.RecordAnalysisFindingToolOutput{
		Reference:    entry.Reference,
		Sequence:     entry.Sequence,
		Title:        entry.Title,
		SubjectCount: len(setSubjects),
	}, nil
}

func (m *systemAnalysisMiddleware) getExistingFindingId(ctx context.Context, analysisId uuid.UUID, ref string) (uuid.UUID, error) {
	findingRefPred := sae.And(sae.AnalysisID(analysisId), sae.Reference(ref), sae.KindEQ(sae.KindFinding))
	existing, listErr := m.analyses.LookupSystemAnalysisEntry(ctx, findingRefPred)
	if existing == nil {
		if listErr != nil && !ent.IsNotFound(listErr) {
			return uuid.Nil, fmt.Errorf("list existing entries: %w", listErr)
		}
		return uuid.Nil, nil
	}
	return existing.ID, nil
}

func (m *systemAnalysisMiddleware) makeSubjectSetters(subjects []rezai.AnalysisFindingSubjectToolInputSubject) ([]func(*ent.SystemAnalysisEntrySubjectMutation), error) {
	if len(subjects) < 1 || len(subjects) > systemAnalysisPageSize {
		return nil, fmt.Errorf("%w: between 1 and %d subjects are required", rez.ErrInvalidInput, systemAnalysisPageSize)
	}

	setSubjects := make([]func(*ent.SystemAnalysisEntrySubjectMutation), len(subjects))
	seen := mapset.NewSetWithSize[string](len(subjects))
	hasEvidence := false
	for i, s := range subjects {
		kind, role := strings.TrimSpace(s.SubjectKind), strings.TrimSpace(s.Role)
		id, idErr := uuid.Parse(s.SubjectID)
		if idErr != nil || role == "" {
			return nil, fmt.Errorf("%w: each finding subject requires a valid ID and non-empty role", rez.ErrInvalidInput)
		}
		if kind == "evidence" {
			hasEvidence = true
		} else if kind != "entity" && kind != "relationship" {
			return nil, fmt.Errorf("%w: subject_kind must be entity, relationship, or evidence", rez.ErrInvalidInput)
		}

		if !seen.Add(kind + ":" + id.String() + ":" + role) {
			return nil, fmt.Errorf("%w: duplicate finding %s subject %s:%s", rez.ErrInvalidInput, kind, id, role)
		}

		setSubjects[i] = func(m *ent.SystemAnalysisEntrySubjectMutation) {
			m.SetRole(role)
			if kind == "entity" {
				m.SetKnowledgeEntityID(id)
			} else if kind == "relationship" {
				m.SetKnowledgeRelationshipID(id)
			} else if kind == "evidence" {
				m.SetKnowledgeEvidenceID(id)
			}
		}
	}
	if !hasEvidence {
		return nil, fmt.Errorf("%w: finding must cite at least one evidence record", rez.ErrInvalidInput)
	}
	return setSubjects, nil
}

func subjectDisplayName(kind, subkind string, id uuid.UUID, evidence *ent.KnowledgeEvidence) string {
	if evidence != nil && strings.TrimSpace(evidence.SubjectState.DisplayName) != "" {
		return evidence.SubjectState.DisplayName
	}
	return fmt.Sprintf("%s:%s:%s", kind, subkind, id)
}

func entitySummary(entity *ent.KnowledgeEntity) rezai.KnowledgeEntitySummary {
	return rezai.KnowledgeEntitySummary{
		ID:          entity.ID,
		Kind:        entity.Kind.String(),
		Subkind:     entity.Subkind,
		DisplayName: subjectDisplayName(entity.Kind.String(), entity.Subkind, entity.ID, entity.LatestEvidence()),
	}
}
func relationshipSummary(relationship *ent.KnowledgeRelationship) rezai.KnowledgeRelationshipSummary {
	return rezai.KnowledgeRelationshipSummary{
		ID:          relationship.ID,
		Kind:        relationship.Kind.String(),
		Subkind:     relationship.Subkind,
		DisplayName: subjectDisplayName(relationship.Kind.String(), relationship.Subkind, relationship.ID, relationship.LatestEvidence()),
	}
}
func subjectAliasSummary(alias *ent.KnowledgeSubjectAlias) rezai.KnowledgeSubjectAliasSummary {
	return rezai.KnowledgeSubjectAliasSummary{
		ID:                       alias.ID,
		Provider:                 alias.Provider,
		ProviderSource:           alias.ProviderSource,
		ProviderSubjectReference: alias.ProviderSubjectRef,
	}
}
func evidenceSummary(evidence *ent.KnowledgeEvidence) rezai.KnowledgeEvidenceSummary {
	return rezai.KnowledgeEvidenceSummary{
		ID:          evidence.ID,
		Kind:        evidence.Kind.String(),
		Assertion:   evidence.Assertion,
		EffectiveAt: evidence.EffectiveAt,
		CreatedAt:   evidence.CreatedAt,
	}
}

func aliasDetails(aliases ent.KnowledgeSubjectAliasSlice) []rezai.KnowledgeSubjectAliasDetail {
	slices.SortFunc(aliases, func(a, b *ent.KnowledgeSubjectAlias) int {
		return strings.Compare(a.ID.String(), b.ID.String())
	})
	result := make([]rezai.KnowledgeSubjectAliasDetail, 0, len(aliases))
	for _, alias := range aliases {
		detail := rezai.KnowledgeSubjectAliasDetail{
			Summary: subjectAliasSummary(alias),
		}
		aliasEvidence := ent.KnowledgeSubjectAliasSlice{alias}
		if latest := aliasEvidence.LatestEvidence(); latest != nil {
			detail.LatestEvidence = new(evidenceSummary(latest))
		}
		result = append(result, detail)
	}
	return result
}
func entityDetail(entity *ent.KnowledgeEntity) *rezai.KnowledgeEntityDetail {
	detail := &rezai.KnowledgeEntityDetail{
		Summary: entitySummary(entity),
		Aliases: aliasDetails(entity.Edges.Aliases),
	}
	return detail
}
func relationshipDetail(relationship *ent.KnowledgeRelationship) (*rezai.KnowledgeRelationshipDetail, error) {
	srcEnt, srcErr := relationship.Edges.SourceEntityOrErr()
	if srcErr != nil {
		return nil, srcErr
	}
	tgtEnt, tgtErr := relationship.Edges.TargetEntityOrErr()
	if tgtErr != nil {
		return nil, tgtErr
	}
	detail := &rezai.KnowledgeRelationshipDetail{
		Summary: relationshipSummary(relationship),
		Aliases: aliasDetails(relationship.Edges.Aliases),
		Source:  entitySummary(srcEnt),
		Target:  entitySummary(tgtEnt),
	}
	return detail, nil
}
func evidenceDetail(evidence *ent.KnowledgeEvidence) (*rezai.KnowledgeEvidenceDetail, error) {
	alias, aliasErr := evidence.Edges.SubjectAliasOrErr()
	if aliasErr != nil {
		return nil, fmt.Errorf("get evidence alias: %w", aliasErr)
	}
	return &rezai.KnowledgeEvidenceDetail{
		Summary:    evidenceSummary(evidence),
		Properties: evidence.SubjectState.Properties,
		Alias:      subjectAliasSummary(alias),
	}, nil
}
