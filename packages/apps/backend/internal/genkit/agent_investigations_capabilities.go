package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/investigationfindingversionlink"
	"github.com/rezible/rezible/ent/investigationhypothesisversion"
	"github.com/rezible/rezible/ent/predicate"
	saent "github.com/rezible/rezible/ent/systemanalysisentity"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sarel "github.com/rezible/rezible/ent/systemanalysisrelationship"
	rezai "github.com/rezible/rezible/pkg/ai"
)

const (
	analysisToolDefaultPageSize  = 25
	analysisToolMaxPageSize      = 50
	analysisToolTextLimit        = 12000
	analysisToolDescriptionLimit = 240
	investigationIntroductionTag = "investigation-introduction"
	analysisToolTruncationMarker = "\n[truncated: complete tool result is limited to 12,000 characters]"
)

type investigationInvocation struct {
	investigations  rez.InvestigationService
	analyses        rez.SystemAnalysisService
	knowledge       rez.KnowledgeGraphQueryService
	investigationID uuid.UUID
	analysisID      uuid.UUID
	turnID          uuid.UUID
}

func (i *investigationInvocation) makeHooks(ctx context.Context, invCtx *agentInvocationContext) (*ai.Hooks, error) {
	wrapInvocationPrompt, wgErr := i.makeWrapGenerate(ctx, invCtx)
	if wgErr != nil {
		return nil, wgErr
	}

	return &ai.Hooks{
		WrapGenerate: wrapInvocationPrompt,
		Tools:        i.makeTools(),
	}, nil
}

func (i *investigationInvocation) makeWrapGenerate(ctx context.Context, invCtx *agentInvocationContext) (wrapGenerateFn, error) {
	var sessionInput rezai.InvestigationAgentSessionInput
	if decodeErr := json.Unmarshal(invCtx.Session.Input, &sessionInput); decodeErr != nil {
		return nil, fmt.Errorf("decode investigation session input: %w", decodeErr)
	}
	if validateErr := sessionInput.Validate(); validateErr != nil {
		return nil, fmt.Errorf("validate investigation session input: %w", validateErr)
	}

	introduction, introductionErr := i.makeIntroduction(ctx, sessionInput, invCtx.Input, invCtx.Turn.Sequence)
	if introductionErr != nil {
		return nil, fmt.Errorf("build investigation introduction: %w", introductionErr)
	}

	return makeSystemTextInjectorFn(investigationIntroductionTag, introduction), nil
}

func (i *investigationInvocation) makeIntroduction(
	ctx context.Context,
	sessInput rezai.InvestigationAgentSessionInput,
	turnInput *rez.AiAgentTurnInput,
	turnSequence int,
) (string, error) {
	countParams := ent.ListParams{Page: 1, PageSize: 1}
	entities, entitiesErr := i.analyses.ListSystemAnalysisEntities(ctx, rez.ListSystemAnalysisEntitiesParams{
		ListParams: countParams,
		Predicates: []predicate.SystemAnalysisEntity{saent.AnalysisID(i.analysisID)},
		Order:      rez.SystemAnalysisSubjectOrderCanonicalID,
	})
	if entitiesErr != nil {
		return "", fmt.Errorf("count included analysis entities: %w", entitiesErr)
	}
	relationships, relationshipsErr := i.analyses.ListSystemAnalysisRelationships(ctx, rez.ListSystemAnalysisRelationshipsParams{
		ListParams: countParams,
		Predicates: []predicate.SystemAnalysisRelationship{sarel.AnalysisID(i.analysisID)},
		Order:      rez.SystemAnalysisSubjectOrderCanonicalID,
	})
	if relationshipsErr != nil {
		return "", fmt.Errorf("count included analysis relationships: %w", relationshipsErr)
	}
	entries, entriesErr := i.analyses.ListSystemAnalysisEntries(ctx, rez.ListSystemAnalysisEntriesParams{
		ListParams: countParams,
		Predicates: []predicate.SystemAnalysisEntry{
			sae.AnalysisID(i.analysisID),
			sae.KindIn(sae.KindObservation, sae.KindContext),
		},
		Order:       rez.SystemAnalysisEntryOrderSequence,
		SummaryOnly: true,
	})
	if entriesErr != nil {
		return "", fmt.Errorf("count observation and context entries: %w", entriesErr)
	}

	assignedWork := "Initial investigation"
	if turnSequence > 1 {
		assignedWork = "Continue the investigation"
		if turnInput != nil && turnInput.Message != nil {
			if messageText := strings.TrimSpace(turnInput.Message.Text()); messageText != "" {
				assignedWork = messageText
			}
		}
	}
	return rezai.FormatInvestigationAgentIntroduction(rezai.InvestigationAgentIntroduction{
		OriginalQuestion: strings.TrimSpace(sessInput.Query),
		AssignedWork:     assignedWork,
		AvailableContent: rezai.InvestigationAvailableContent{
			Entities:      entities.Total,
			Relationships: relationships.Total,
			Entries:       entries.Total,
		},
	})
}

func (i *investigationInvocation) makeTools() []ai.Tool {
	return []ai.Tool{
		makeDefinedTool(rezai.ListAnalysisSubjectsTool, i.listSubjects),
		makeDefinedTool(rezai.ListAnalysisEntriesTool, i.listEntries),
		makeDefinedTool(rezai.InspectAnalysisSubjectTool, i.inspectSubject),
		makeDefinedTool(rezai.ReadAnalysisEntryTool, i.readEntry),
		makeDefinedTool(rezai.ReadAnalysisEvidenceTool, i.readEvidence),
		makeDefinedTool(rezai.PublishInvestigationReportTool, i.publishReport),
		makeDefinedTool(rezai.PublishInvestigationFindingTool, i.publishFinding),
		makeDefinedTool(rezai.PublishInvestigationAnswerTool, i.publishAnswer),
		makeDefinedTool(rezai.PublishInvestigationHypothesisTool, i.publishHypothesis),
		makeDefinedTool(rezai.ReadInvestigationReportTool, i.readReport),
		makeDefinedTool(rezai.ListInvestigationFindingsTool, i.listFindings),
		makeDefinedTool(rezai.ListInvestigationHypothesesTool, i.listHypotheses),
		makeDefinedTool(rezai.ReadInvestigationFindingTool, i.readFinding),
		makeDefinedTool(rezai.ReadInvestigationHypothesisTool, i.readHypothesis),
	}
}

func (i *investigationInvocation) listSubjects(ctx context.Context, input rezai.ListAnalysisSubjectsArgs) (*rezai.AnalysisToolResult, error) {
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}

	switch strings.TrimSpace(input.Kind) {
	case "entity":
		params := rez.ListSystemAnalysisEntitiesParams{
			ListParams: page,
			Predicates: []predicate.SystemAnalysisEntity{saent.AnalysisID(i.analysisID)},
			Order:      rez.SystemAnalysisSubjectOrderCanonicalID,
		}
		entities, listErr := i.analyses.ListSystemAnalysisEntities(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("list included analysis entities: %w", listErr)
		}

		lines := make([]string, 0, len(entities.Data))
		for _, membership := range entities.Data {
			line, renderErr := i.renderEntityMembership(membership)
			if renderErr != nil {
				return nil, renderErr
			}
			lines = append(lines, line)
		}
		return i.analysisPageResult(strings.Join(lines, "\n"), page, entities.Total, len(entities.Data)), nil

	case "relationship":
		params := rez.ListSystemAnalysisRelationshipsParams{
			ListParams:    page,
			Predicates:    []predicate.SystemAnalysisRelationship{sarel.AnalysisID(i.analysisID)},
			Order:         rez.SystemAnalysisSubjectOrderCanonicalID,
			WithEndpoints: true,
		}
		relationships, listErr := i.analyses.ListSystemAnalysisRelationships(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("list included analysis relationships: %w", listErr)
		}

		lines := make([]string, 0, len(relationships.Data))
		for _, membership := range relationships.Data {
			line, renderErr := i.renderRelationshipMembership(membership)
			if renderErr != nil {
				return nil, renderErr
			}
			lines = append(lines, line)
		}
		return i.analysisPageResult(strings.Join(lines, "\n"), page, relationships.Total, len(relationships.Data)), nil
	default:
		return nil, fmt.Errorf("%w: kind must be entity or relationship", rez.ErrInvalidInput)
	}
}

func (i *investigationInvocation) listEntries(ctx context.Context, input rezai.AnalysisPageArgs) (*rezai.AnalysisToolResult, error) {
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}

	params := rez.ListSystemAnalysisEntriesParams{
		ListParams: page,
		Predicates: []predicate.SystemAnalysisEntry{
			sae.AnalysisID(i.analysisID),
			sae.KindIn(sae.KindObservation, sae.KindContext),
		},
		Order:       rez.SystemAnalysisEntryOrderSequence,
		SummaryOnly: true,
	}
	entries, listErr := i.analyses.ListSystemAnalysisEntries(ctx, params)
	if listErr != nil {
		return nil, fmt.Errorf("list observation and context entries: %w", listErr)
	}
	return i.analysisPageResult(i.formatEntryPage(entries.Data), page, entries.Total, len(entries.Data)), nil
}

func (i *investigationInvocation) inspectSubject(ctx context.Context, input rezai.InspectAnalysisSubjectArgs) (*rezai.AnalysisToolResult, error) {
	if input.ID == uuid.Nil {
		return nil, fmt.Errorf("%w: id is required", rez.ErrInvalidInput)
	}
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}

	var subjectLine string
	var entries *ent.ListResult[ent.SystemAnalysisEntry]
	switch strings.TrimSpace(input.Kind) {
	case "entity":
		params := rez.ListSystemAnalysisEntitiesParams{
			ListParams: ent.ListParams{Page: 1, PageSize: 1},
			Predicates: []predicate.SystemAnalysisEntity{
				saent.AnalysisID(i.analysisID),
				saent.KnowledgeEntityID(input.ID),
			},
			Order: rez.SystemAnalysisSubjectOrderCanonicalID,
		}
		memberships, listErr := i.analyses.ListSystemAnalysisEntities(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("verify included analysis entity: %w", listErr)
		}
		if memberships.Total == 0 {
			return nil, rez.ErrNotFound
		}
		subjectLine, listErr = i.renderEntityMembership(memberships.Data[0])
		if listErr != nil {
			return nil, listErr
		}
		entries, listErr = i.entriesForSubject(ctx, "entity", input.ID, page)
		if listErr != nil {
			return nil, fmt.Errorf("list entries linked to analysis entity: %w", listErr)
		}

	case "relationship":
		params := rez.ListSystemAnalysisRelationshipsParams{
			ListParams: ent.ListParams{Page: 1, PageSize: 1},
			Predicates: []predicate.SystemAnalysisRelationship{
				sarel.AnalysisID(i.analysisID),
				sarel.KnowledgeRelationshipID(input.ID),
			},
			Order:         rez.SystemAnalysisSubjectOrderCanonicalID,
			WithEndpoints: true,
		}
		memberships, listErr := i.analyses.ListSystemAnalysisRelationships(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("verify included analysis relationship: %w", listErr)
		}
		if memberships.Total == 0 {
			return nil, rez.ErrNotFound
		}
		subjectLine, listErr = i.renderRelationshipMembership(memberships.Data[0])
		if listErr != nil {
			return nil, listErr
		}
		entries, listErr = i.entriesForSubject(ctx, "relationship", input.ID, page)
		if listErr != nil {
			return nil, fmt.Errorf("list entries linked to analysis relationship: %w", listErr)
		}
	default:
		return nil, fmt.Errorf("%w: kind must be entity or relationship", rez.ErrInvalidInput)
	}

	text := subjectLine + "\n" + i.formatEntryPage(entries.Data)
	return i.analysisPageResult(text, page, entries.Total, len(entries.Data)), nil
}

func (i *investigationInvocation) readEntry(ctx context.Context, input rezai.ReadAnalysisEntryArgs) (*rezai.AnalysisToolResult, error) {
	if input.EntryID == uuid.Nil {
		return nil, fmt.Errorf("%w: entry_id is required", rez.ErrInvalidInput)
	}
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}

	entryParams := rez.ListSystemAnalysisEntriesParams{
		ListParams: ent.ListParams{Page: 1, PageSize: 1},
		Predicates: []predicate.SystemAnalysisEntry{
			sae.AnalysisID(i.analysisID),
			sae.ID(input.EntryID),
			sae.KindIn(sae.KindObservation, sae.KindContext),
		},
		Order:       rez.SystemAnalysisEntryOrderSequence,
		SummaryOnly: true,
	}
	entries, listErr := i.analyses.ListSystemAnalysisEntries(ctx, entryParams)
	if listErr != nil {
		return nil, fmt.Errorf("read analysis entry: %w", listErr)
	}
	if entries.Total == 0 {
		return nil, rez.ErrNotFound
	}
	entry := entries.Data[0]

	attachmentParams := rez.ListSystemAnalysisEntrySubjectsParams{
		AnalysisID: i.analysisID,
		EntryID:    input.EntryID,
		ListParams: page,
	}
	attachments, attachmentsErr := i.analyses.ListSystemAnalysisEntrySubjects(ctx, attachmentParams)
	if ent.IsNotFound(attachmentsErr) {
		return nil, rez.ErrNotFound
	}
	if attachmentsErr != nil {
		return nil, fmt.Errorf("list analysis entry attachments: %w", attachmentsErr)
	}

	lines := []string{
		"entry " + entry.ID.String(),
		"kind=" + entry.Kind.String(),
		"title=" + limitedAnalysisDescription(entry.Title),
	}
	if entry.OccurredAt != nil {
		lines = append(lines, "occurred_at="+entry.OccurredAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"))
	}
	var attachmentLines []string
	for _, attachment := range attachments.Data {
		kind, id, targetErr := i.analysisEntryAttachmentTarget(attachment)
		if targetErr != nil {
			return nil, targetErr
		}
		if kind == "normalized_event" {
			continue
		}
		line := kind + " " + id.String()
		if strings.TrimSpace(attachment.Role) != "" {
			line += " role=" + limitedAnalysisDescription(attachment.Role)
		}
		attachmentLines = append(attachmentLines, line)
	}
	if len(attachmentLines) == 0 {
		lines = append(lines, "No supported attachments on this page.")
	} else {
		lines = append(lines, attachmentLines...)
	}
	content := strings.Join(lines, "\n")
	if body := strings.TrimSpace(entry.Body); body != "" {
		content += "\nbody:\n" + body
	}
	hasMore, pageStatus := analysisPageStatus(page, attachments.Total, len(attachments.Data))
	return &rezai.AnalysisToolResult{
		Text:     limitedAnalysisResultText("", content, "\n"+pageStatus),
		Page:     page.Page,
		PageSize: page.PageSize,
		Total:    attachments.Total,
		HasMore:  hasMore,
	}, nil
}

func (i *investigationInvocation) readEvidence(ctx context.Context, input rezai.ReadAnalysisEvidenceArgs) (*rezai.AnalysisToolResult, error) {
	if input.EvidenceID == uuid.Nil {
		return nil, fmt.Errorf("%w: evidence_id is required", rez.ErrInvalidInput)
	}

	attachmentPredicate := saes.KnowledgeEvidenceID(input.EvidenceID)
	params := rez.ListSystemAnalysisEntriesParams{
		ListParams: ent.ListParams{Page: 1, PageSize: 1},
		Predicates: []predicate.SystemAnalysisEntry{
			sae.AnalysisID(i.analysisID),
			sae.KindIn(sae.KindObservation, sae.KindContext),
			sae.HasSubjectsWith(attachmentPredicate),
		},
		Order:       rez.SystemAnalysisEntryOrderSequence,
		SummaryOnly: true,
	}
	linkedEntries, linkedEntriesErr := i.analyses.ListSystemAnalysisEntries(ctx, params)
	if linkedEntriesErr != nil {
		return nil, fmt.Errorf("verify analysis evidence attachment: %w", linkedEntriesErr)
	}
	if linkedEntries.Total == 0 {
		return nil, rez.ErrNotFound
	}

	evidence, getEvidenceErr := i.knowledge.GetEvidence(ctx, input.EvidenceID)
	if ent.IsNotFound(getEvidenceErr) {
		return nil, rez.ErrNotFound
	}
	if getEvidenceErr != nil {
		return nil, fmt.Errorf("read knowledge evidence: %w", getEvidenceErr)
	}

	lines := []string{
		"knowledge_evidence " + evidence.ID.String(),
		"kind=" + evidence.Kind.String(),
		"assertion=" + evidence.Assertion,
		"effective_at=" + evidence.EffectiveAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
	}
	if displayName := strings.TrimSpace(evidence.SubjectState.DisplayName); displayName != "" {
		lines = append(lines, "subject_display_name="+limitedAnalysisDescription(displayName))
	}
	if description := limitedAnalysisDescription(evidence.SubjectState.Description); description != "" {
		lines = append(lines, "subject_description="+description)
	}
	return &rezai.AnalysisToolResult{
		Text:     limitedAnalysisResultText("", strings.Join(lines, "\n"), "\nhas_more=false"),
		Page:     1,
		PageSize: 1,
		Total:    1,
	}, nil
}

func (i *investigationInvocation) publishReport(ctx context.Context, input rezai.PublishInvestigationReportToolInput) (*rezai.InvestigationReportToolResult, error) {
	if validateErr := i.validateEvidenceIDs(input.EvidenceIDs); validateErr != nil {
		return nil, validateErr
	}
	params := rez.PublishInvestigationReportParams{
		Text:        input.Text,
		EvidenceIDs: input.EvidenceIDs,
	}
	scope := rez.InvestigationPublicationScope{InvestigationID: i.investigationID, AgentTurnID: i.turnID}
	result, publishErr := i.investigations.PublishInvestigationReport(ctx, scope, params)
	if publishErr != nil {
		return nil, fmt.Errorf("publish investigation report: %w", publishErr)
	}
	return i.reportToolResult(result), nil
}

func (i *investigationInvocation) publishFinding(ctx context.Context, input rezai.PublishInvestigationFindingToolInput) (*rezai.InvestigationFindingVersionToolResult, error) {
	if validateErr := i.validateEvidenceIDs(input.EvidenceIDs); validateErr != nil {
		return nil, validateErr
	}
	findingReferences, referencesErr := i.findingVersionReferences(input.FindingReferences)
	if referencesErr != nil {
		return nil, referencesErr
	}
	params := rez.PublishInvestigationFindingParams{
		Key:               input.Key,
		Title:             input.Title,
		Body:              input.Body,
		EvidenceIDs:       input.EvidenceIDs,
		FindingReferences: findingReferences,
	}
	scope := rez.InvestigationPublicationScope{InvestigationID: i.investigationID, AgentTurnID: i.turnID}
	result, publishErr := i.investigations.PublishInvestigationFinding(ctx, scope, params)
	if publishErr != nil {
		return nil, fmt.Errorf("publish investigation finding: %w", publishErr)
	}
	return i.findingToolResult(result), nil
}

func (i *investigationInvocation) publishAnswer(ctx context.Context, input rezai.PublishInvestigationAnswerToolInput) (*rezai.InvestigationFindingVersionToolResult, error) {
	if validateErr := i.validateEvidenceIDs(input.EvidenceIDs); validateErr != nil {
		return nil, validateErr
	}
	findingReferences, referencesErr := i.findingVersionReferences(input.FindingReferences)
	if referencesErr != nil {
		return nil, referencesErr
	}
	params := rez.PublishInvestigationAnswerParams{
		Title:             input.Title,
		Body:              input.Body,
		EvidenceIDs:       input.EvidenceIDs,
		FindingReferences: findingReferences,
	}
	scope := rez.InvestigationPublicationScope{InvestigationID: i.investigationID, AgentTurnID: i.turnID}
	result, publishErr := i.investigations.PublishInvestigationAnswer(ctx, scope, params)
	if publishErr != nil {
		return nil, fmt.Errorf("publish investigation answer: %w", publishErr)
	}
	return i.findingToolResult(result), nil
}

func (i *investigationInvocation) publishHypothesis(ctx context.Context, input rezai.PublishInvestigationHypothesisToolInput) (*rezai.InvestigationHypothesisVersionToolResult, error) {
	if validateErr := i.validateEvidenceIDs(input.EvidenceIDs); validateErr != nil {
		return nil, validateErr
	}
	status := strings.TrimSpace(input.Status)
	if status != "open" && status != "supported" && status != "disproven" && status != "inconclusive" {
		return nil, fmt.Errorf("%w: status must be open, supported, disproven, or inconclusive", rez.ErrInvalidInput)
	}
	params := rez.PublishInvestigationHypothesisParams{
		Key:           input.Key,
		Title:         input.Title,
		Justification: input.Justification,
		Status:        investigationhypothesisversion.Status(status),
		EvidenceIDs:   input.EvidenceIDs,
	}
	scope := rez.InvestigationPublicationScope{InvestigationID: i.investigationID, AgentTurnID: i.turnID}
	result, publishErr := i.investigations.PublishInvestigationHypothesis(ctx, scope, params)
	if publishErr != nil {
		return nil, fmt.Errorf("publish investigation hypothesis: %w", publishErr)
	}
	return i.hypothesisToolResult(result), nil
}

func (i *investigationInvocation) readReport(ctx context.Context, input rezai.ReadInvestigationReportToolInput) (*rezai.InvestigationReportToolResult, error) {
	selection := strings.TrimSpace(input.Selection)
	if selection != "" && selection != "latest" && selection != "completed" {
		return nil, fmt.Errorf("%w: selection must be latest or completed", rez.ErrInvalidInput)
	}
	var reportSelection rez.InvestigationReportSelection
	switch selection {
	case "", string(rez.InvestigationReportSelectionLatest):
		reportSelection = rez.InvestigationReportSelectionLatest
	case string(rez.InvestigationReportSelectionCompleted):
		reportSelection = rez.InvestigationReportSelectionCompleted
	default:
		return nil, fmt.Errorf("%w: selection must be latest or completed", rez.ErrInvalidInput)
	}
	params := rez.ReadInvestigationReportParams{Selection: reportSelection}
	result, readErr := i.investigations.ReadInvestigationReport(ctx, i.investigationID, params)
	if readErr != nil {
		return nil, fmt.Errorf("read investigation report: %w", readErr)
	}
	return i.reportToolResult(result), nil
}

func (i *investigationInvocation) listFindings(ctx context.Context, input rezai.AnalysisPageArgs) (*rezai.InvestigationOutputPage[rezai.InvestigationFindingVersionToolResult], error) {
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}
	result, listErr := i.investigations.ListInvestigationFindings(ctx, i.investigationID, page)
	if listErr != nil {
		return nil, fmt.Errorf("list investigation findings: %w", listErr)
	}
	if result == nil {
		return nil, nil
	}

	items := make([]rezai.InvestigationFindingVersionToolResult, 0, len(result.Data))
	for _, finding := range result.Data {
		items = append(items, *i.findingToolResult(finding))
	}
	return &rezai.InvestigationOutputPage[rezai.InvestigationFindingVersionToolResult]{
		Items: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total,
		HasMore: result.Page*result.PageSize < result.Total,
	}, nil
}

func (i *investigationInvocation) listHypotheses(ctx context.Context, input rezai.AnalysisPageArgs) (*rezai.InvestigationOutputPage[rezai.InvestigationHypothesisVersionToolResult], error) {
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}
	result, listErr := i.investigations.ListInvestigationHypotheses(ctx, i.investigationID, page)
	if listErr != nil {
		return nil, fmt.Errorf("list investigation hypotheses: %w", listErr)
	}
	if result == nil {
		return nil, nil
	}

	items := make([]rezai.InvestigationHypothesisVersionToolResult, 0, len(result.Data))
	for _, hypothesis := range result.Data {
		items = append(items, *i.hypothesisToolResult(hypothesis))
	}
	return &rezai.InvestigationOutputPage[rezai.InvestigationHypothesisVersionToolResult]{
		Items: items, Page: result.Page, PageSize: result.PageSize, Total: result.Total,
		HasMore: result.Page*result.PageSize < result.Total,
	}, nil
}

func (i *investigationInvocation) readFinding(ctx context.Context, input rezai.ReadInvestigationFindingToolInput) (*rezai.InvestigationFindingVersionToolResult, error) {
	if input.VersionID == uuid.Nil {
		return nil, fmt.Errorf("%w: version_id is required", rez.ErrInvalidInput)
	}
	result, readErr := i.investigations.GetInvestigationFindingVersion(ctx, i.investigationID, input.VersionID)
	if readErr != nil {
		return nil, fmt.Errorf("read investigation finding version: %w", readErr)
	}
	return i.findingToolResult(result), nil
}

func (i *investigationInvocation) readHypothesis(ctx context.Context, input rezai.ReadInvestigationHypothesisToolInput) (*rezai.InvestigationHypothesisVersionToolResult, error) {
	if input.VersionID == uuid.Nil {
		return nil, fmt.Errorf("%w: version_id is required", rez.ErrInvalidInput)
	}
	result, readErr := i.investigations.GetInvestigationHypothesisVersion(ctx, i.investigationID, input.VersionID)
	if readErr != nil {
		return nil, fmt.Errorf("read investigation hypothesis version: %w", readErr)
	}
	return i.hypothesisToolResult(result), nil
}

func (i *investigationInvocation) findingVersionReferences(inputs []rezai.InvestigationFindingReferenceInput) ([]rez.FindingVersionReference, error) {
	references := make([]rez.FindingVersionReference, 0, len(inputs))
	for _, input := range inputs {
		if input.VersionID == uuid.Nil {
			return nil, fmt.Errorf("%w: finding version ID is required", rez.ErrInvalidInput)
		}
		relation := investigationfindingversionlink.Relation(input.Relation)
		switch relation {
		case investigationfindingversionlink.RelationSupports, investigationfindingversionlink.RelationContradicts, investigationfindingversionlink.RelationInvalidates:
		default:
			return nil, fmt.Errorf("%w: relation must be supports, contradicts, or invalidates", rez.ErrInvalidInput)
		}
		references = append(references, rez.FindingVersionReference{VersionID: input.VersionID, Relation: relation})
	}
	return references, nil
}

func (i *investigationInvocation) reportToolResult(result *rez.InvestigationReportResult) *rezai.InvestigationReportToolResult {
	if result == nil {
		return nil
	}
	return &rezai.InvestigationReportToolResult{
		Text:        result.Text,
		EvidenceIDs: append([]uuid.UUID{}, result.EvidenceIDs...),
		TurnStatus:  string(result.TurnStatus),
		Provisional: result.TurnStatus == agentturn.StatusRunning,
		CreatedAt:   result.CreatedAt,
	}
}

func (i *investigationInvocation) findingToolResult(result *rez.InvestigationFindingVersion) *rezai.InvestigationFindingVersionToolResult {
	if result == nil {
		return nil
	}
	key := result.Key
	isAnswer := result.UserInputID != nil
	if isAnswer {
		key = ""
	}
	findingReferences := make([]rezai.InvestigationFindingReference, 0, len(result.FindingReferences))
	for _, reference := range result.FindingReferences {
		findingReferences = append(findingReferences, rezai.InvestigationFindingReference{
			VersionID: reference.VersionID,
			Relation:  string(reference.Relation),
		})
	}
	return &rezai.InvestigationFindingVersionToolResult{
		VersionID:               result.ID,
		Key:                     key,
		IsAnswer:                isAnswer,
		Title:                   result.Title,
		Body:                    result.Body,
		EvidenceIDs:             append([]uuid.UUID{}, result.EvidenceIDs...),
		FindingReferences:       findingReferences,
		InvalidatedByVersionIDs: append([]uuid.UUID{}, result.InvalidatedByVersionIDs...),
		TurnStatus:              string(result.TurnStatus),
		Provisional:             result.TurnStatus == agentturn.StatusRunning,
		CreatedAt:               result.CreatedAt,
	}
}

func (i *investigationInvocation) hypothesisToolResult(result *rez.InvestigationHypothesisVersion) *rezai.InvestigationHypothesisVersionToolResult {
	if result == nil {
		return nil
	}
	return &rezai.InvestigationHypothesisVersionToolResult{
		VersionID:     result.ID,
		Key:           result.Key,
		Title:         result.Title,
		Justification: result.Justification,
		Status:        string(result.Status),
		EvidenceIDs:   append([]uuid.UUID{}, result.EvidenceIDs...),
		TurnStatus:    string(result.TurnStatus),
		Provisional:   result.TurnStatus == agentturn.StatusRunning,
		CreatedAt:     result.CreatedAt,
	}
}

func (i *investigationInvocation) renderEntityMembership(membership *ent.SystemAnalysisEntity) (string, error) {
	entity, entityErr := membership.Edges.KnowledgeEntityOrErr()
	if entityErr != nil {
		return "", fmt.Errorf("load included analysis entity: %w", entityErr)
	}
	name := i.entityDisplayName(entity)
	description := i.entityDescription(entity)
	if membership.LabelOverride != nil {
		name = *membership.LabelOverride
	}
	if membership.DescriptionOverride != nil {
		description = *membership.DescriptionOverride
	}
	return i.renderEntityLine(entity, name, description), nil
}

func (i *investigationInvocation) renderRelationshipMembership(membership *ent.SystemAnalysisRelationship) (string, error) {
	relationship, relationshipErr := membership.Edges.KnowledgeRelationshipOrErr()
	if relationshipErr != nil {
		return "", fmt.Errorf("load included analysis relationship: %w", relationshipErr)
	}
	source, sourceErr := relationship.Edges.SourceEntityOrErr()
	if sourceErr != nil {
		return "", fmt.Errorf("load relationship source entity: %w", sourceErr)
	}
	target, targetErr := relationship.Edges.TargetEntityOrErr()
	if targetErr != nil {
		return "", fmt.Errorf("load relationship target entity: %w", targetErr)
	}
	line := fmt.Sprintf(
		"relationship %s %s (%s) --%s--> %s (%s)",
		relationship.ID,
		i.entityNameOrID(source), source.ID,
		relationship.Predicate,
		i.entityNameOrID(target), target.ID,
	)
	name := ""
	description := ""
	if evidence := relationship.LatestEvidence(); evidence != nil {
		name = strings.TrimSpace(evidence.SubjectState.DisplayName)
		description = evidence.SubjectState.Description
	}
	if membership.LabelOverride != nil {
		name = strings.TrimSpace(*membership.LabelOverride)
	}
	if membership.DescriptionOverride != nil {
		description = *membership.DescriptionOverride
	}
	if name != "" {
		line += " name=" + limitedAnalysisDescription(name)
	}
	if description = limitedAnalysisDescription(description); description != "" {
		line += " — " + description
	}
	return line, nil
}

func (i *investigationInvocation) entriesForSubject(ctx context.Context, kind string, subjectID uuid.UUID, page ent.ListParams) (*ent.ListResult[ent.SystemAnalysisEntry], error) {
	var subjectPredicate predicate.SystemAnalysisEntrySubject
	if kind == "entity" {
		subjectPredicate = saes.KnowledgeEntityID(subjectID)
	} else {
		subjectPredicate = saes.KnowledgeRelationshipID(subjectID)
	}
	params := rez.ListSystemAnalysisEntriesParams{
		ListParams: page,
		Predicates: []predicate.SystemAnalysisEntry{
			sae.AnalysisID(i.analysisID),
			sae.KindIn(sae.KindObservation, sae.KindContext),
			sae.HasSubjectsWith(subjectPredicate),
		},
		Order:       rez.SystemAnalysisEntryOrderSequence,
		SummaryOnly: true,
	}
	return i.analyses.ListSystemAnalysisEntries(ctx, params)
}

func (i *investigationInvocation) renderEntryLine(entry *ent.SystemAnalysisEntry) string {
	line := fmt.Sprintf("entry %s %s: %s", entry.ID, entry.Kind, limitedAnalysisDescription(entry.Title))
	if entry.OccurredAt != nil {
		line += " occurred_at=" + entry.OccurredAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	}
	return line
}

func (i *investigationInvocation) analysisPageResult(text string, page ent.ListParams, total, returned int) *rezai.AnalysisToolResult {
	if strings.TrimSpace(text) == "" {
		text = "No items found."
	}
	hasMore, status := analysisPageStatus(page, total, returned)
	return &rezai.AnalysisToolResult{
		Text:     text + "\n" + status,
		Page:     page.Page,
		PageSize: page.PageSize,
		Total:    total,
		HasMore:  hasMore,
	}
}

func analysisPageStatus(page ent.ListParams, total, returned int) (bool, string) {
	offset := (page.Page - 1) * page.PageSize
	if offset+returned < total {
		return true, fmt.Sprintf("has_more=true; request page=%d, page_size=%d to continue.", page.Page+1, page.PageSize)
	}
	return false, "has_more=false"
}

func normalizeAnalysisPage(page, pageSize int) (ent.ListParams, error) {
	if page < 0 || pageSize < 0 || pageSize > analysisToolMaxPageSize {
		return ent.ListParams{}, fmt.Errorf("%w: page and page_size must be non-negative and page_size must not exceed %d", rez.ErrInvalidInput, analysisToolMaxPageSize)
	}
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = analysisToolDefaultPageSize
	}
	maxInt := int(^uint(0) >> 1)
	if page > maxInt/pageSize {
		return ent.ListParams{}, fmt.Errorf("%w: page is too large", rez.ErrInvalidInput)
	}
	return ent.ListParams{Page: page, PageSize: pageSize}, nil
}

func (i *investigationInvocation) formatEntryPage(entries []*ent.SystemAnalysisEntry) string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, i.renderEntryLine(entry))
	}
	return strings.Join(lines, "\n")
}

func (i *investigationInvocation) renderEntityLine(entity *ent.KnowledgeEntity, name, description string) string {
	line := "entity " + entity.ID.String()
	if strings.TrimSpace(name) != "" {
		line += " " + limitedAnalysisDescription(name)
	}
	line += " [" + entity.Kind + "]"
	if description = limitedAnalysisDescription(description); description != "" {
		line += " — " + description
	}
	return line
}

func (i *investigationInvocation) entityNameOrID(entity *ent.KnowledgeEntity) string {
	if name := i.entityDisplayName(entity); name != "" {
		return limitedAnalysisDescription(name)
	}
	return entity.ID.String()
}

func (i *investigationInvocation) entityDisplayName(entity *ent.KnowledgeEntity) string {
	if evidence := entity.LatestEvidence(); evidence != nil {
		return strings.TrimSpace(evidence.SubjectState.DisplayName)
	}
	return ""
}

func (i *investigationInvocation) entityDescription(entity *ent.KnowledgeEntity) string {
	if evidence := entity.LatestEvidence(); evidence != nil {
		return evidence.SubjectState.Description
	}
	return ""
}

func (i *investigationInvocation) analysisEntryAttachmentTarget(subject *ent.SystemAnalysisEntrySubject) (string, uuid.UUID, error) {
	targetCount := 0
	targetKind := ""
	targetID := uuid.Nil
	if subject.KnowledgeEntityID != nil {
		targetCount++
		targetKind = "entity"
		targetID = *subject.KnowledgeEntityID
	}
	if subject.KnowledgeRelationshipID != nil {
		targetCount++
		targetKind = "relationship"
		targetID = *subject.KnowledgeRelationshipID
	}
	if subject.KnowledgeEvidenceID != nil {
		targetCount++
		targetKind = "knowledge_evidence"
		targetID = *subject.KnowledgeEvidenceID
	}
	if targetCount != 1 || targetID == uuid.Nil {
		return "", uuid.Nil, fmt.Errorf("analysis entry attachment %s has an invalid target", subject.ID)
	}
	return targetKind, targetID, nil
}

func (i *investigationInvocation) validateEvidenceIDs(ids []uuid.UUID) error {
	if slices.Contains(ids, uuid.Nil) {
		return fmt.Errorf("%w: evidence IDs must not be empty", rez.ErrInvalidInput)
	}
	return nil
}

func limitedAnalysisDescription(value string) string {
	cleaned := strings.TrimSpace(value)
	if len([]rune(cleaned)) <= analysisToolDescriptionLimit {
		return cleaned
	}
	return string([]rune(cleaned)[:analysisToolDescriptionLimit]) + "…"
}

func limitedAnalysisResultText(prefix, content, suffix string) string {
	full := prefix + content + suffix
	if len([]rune(full)) <= analysisToolTextLimit {
		return full
	}
	contentRunes := []rune(content)
	contentLimit := max(analysisToolTextLimit-len([]rune(prefix))-len([]rune(suffix))-len([]rune(analysisToolTruncationMarker)), 0)
	if contentLimit > len(contentRunes) {
		contentLimit = len(contentRunes)
	}
	return prefix + string(contentRunes[:contentLimit]) + analysisToolTruncationMarker + suffix
}
