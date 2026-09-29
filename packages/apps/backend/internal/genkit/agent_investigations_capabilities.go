package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"

	at "github.com/rezible/rezible/ent/agentturn"
	inv "github.com/rezible/rezible/ent/investigation"
	invfvl "github.com/rezible/rezible/ent/investigationfindingversionlink"
	invhv "github.com/rezible/rezible/ent/investigationhypothesisversion"
	invui "github.com/rezible/rezible/ent/investigationuserinput"
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
	investigations rez.InvestigationService
	analyses       rez.SystemAnalysisService
	knowledge      rez.KnowledgeGraphQueryService

	investigationID uuid.UUID
	analysisID      uuid.UUID
	turnID          uuid.UUID

	hasAssignedQuestion bool
}

func (i *investigationInvocation) makeHooks(ctx context.Context, invCtx *agentInvocationContext) (*ai.Hooks, error) {
	hasQuestion, checkQuestionErr := i.checkHasAssignedQuestion(ctx)
	if checkQuestionErr != nil {
		return nil, fmt.Errorf("check assigned user question: %w", checkQuestionErr)
	}
	i.hasAssignedQuestion = hasQuestion

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
	tools := []ai.Tool{
		makeDefinedTool(rezai.ListAnalysisSubjectsTool, i.listSubjects),
		makeDefinedTool(rezai.ListAnalysisEntriesTool, i.listEntries),
		makeDefinedTool(rezai.InspectAnalysisSubjectTool, i.inspectSubject),
		makeDefinedTool(rezai.ReadAnalysisEntryTool, i.readEntry),
		makeDefinedTool(rezai.ReadAnalysisEvidenceTool, i.readEvidence),
		makeDefinedTool(rezai.PublishInvestigationReportTool, i.publishReport),
		makeDefinedTool(rezai.PublishInvestigationFindingTool, i.publishFinding),
		makeDefinedTool(rezai.PublishInvestigationHypothesisTool, i.publishHypothesis),
		makeDefinedTool(rezai.ReadInvestigationReportTool, i.readReport),
		makeDefinedTool(rezai.ListInvestigationFindingsTool, i.listFindings),
		makeDefinedTool(rezai.ListInvestigationHypothesesTool, i.listHypotheses),
		makeDefinedTool(rezai.ReadInvestigationFindingTool, i.readFinding),
		makeDefinedTool(rezai.ReadInvestigationHypothesisTool, i.readHypothesis),
	}
	if i.hasAssignedQuestion {
		tools = append(tools, makeDefinedTool(rezai.PublishInvestigationAnswerTool, i.publishAnswer))
	}
	return tools
}

func (i *investigationInvocation) checkHasAssignedQuestion(ctx context.Context) (bool, error) {
	// Answer publication requires a persisted question assigned to this exact turn.
	// The original investigation query and queued questions do not qualify.
	_, lookupErr := i.investigations.LookupInvestigation(ctx,
		inv.ID(i.investigationID),
		inv.HasUserInputsWith(invui.AgentTurnID(i.turnID)),
	)
	if lookupErr != nil && !ent.IsNotFound(lookupErr) {
		return false, fmt.Errorf("check assigned investigation question: %w", lookupErr)
	}
	return lookupErr == nil, nil
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
	subjectID, resolveErr := i.resolveRef(input.Ref)
	if resolveErr != nil {
		return nil, fmt.Errorf("ref: %w", resolveErr)
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
				saent.KnowledgeEntityID(subjectID),
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
		entries, listErr = i.entriesForSubject(ctx, "entity", subjectID, page)
		if listErr != nil {
			return nil, fmt.Errorf("list entries linked to analysis entity: %w", listErr)
		}

	case "relationship":
		params := rez.ListSystemAnalysisRelationshipsParams{
			ListParams: ent.ListParams{Page: 1, PageSize: 1},
			Predicates: []predicate.SystemAnalysisRelationship{
				sarel.AnalysisID(i.analysisID),
				sarel.KnowledgeRelationshipID(subjectID),
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
		entries, listErr = i.entriesForSubject(ctx, "relationship", subjectID, page)
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
	entryID, resolveErr := i.resolveRef(input.Ref)
	if resolveErr != nil {
		return nil, fmt.Errorf("ref: %w", resolveErr)
	}
	page, pageErr := normalizeAnalysisPage(input.Page, input.PageSize)
	if pageErr != nil {
		return nil, pageErr
	}

	entryParams := rez.ListSystemAnalysisEntriesParams{
		ListParams: ent.ListParams{Page: 1, PageSize: 1},
		Predicates: []predicate.SystemAnalysisEntry{
			sae.AnalysisID(i.analysisID),
			sae.ID(entryID),
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
		EntryID:    entryID,
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
		"entry " + i.formatRef(entry.ID),
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
		line := kind + " " + i.formatRef(id)
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
	evidenceID, resolveErr := i.resolveRef(input.Ref)
	if resolveErr != nil {
		return nil, fmt.Errorf("ref: %w", resolveErr)
	}

	attachmentPredicate := saes.KnowledgeEvidenceID(evidenceID)
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

	evidence, getEvidenceErr := i.knowledge.GetEvidence(ctx, evidenceID)
	if ent.IsNotFound(getEvidenceErr) {
		return nil, rez.ErrNotFound
	}
	if getEvidenceErr != nil {
		return nil, fmt.Errorf("read knowledge evidence: %w", getEvidenceErr)
	}

	lines := []string{
		"knowledge_evidence " + i.formatRef(evidence.ID),
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
	evidenceIDs, resolveErr := i.resolveRefs(input.EvidenceRefs)
	if resolveErr != nil {
		return nil, fmt.Errorf("evidence_refs: %w", resolveErr)
	}
	params := rez.PublishInvestigationReportParams{
		Text:        input.Text,
		EvidenceIDs: evidenceIDs,
	}
	scope := rez.InvestigationPublicationScope{InvestigationID: i.investigationID, AgentTurnID: i.turnID}
	result, publishErr := i.investigations.PublishInvestigationReport(ctx, scope, params)
	if publishErr != nil {
		return nil, fmt.Errorf("publish investigation report: %w", publishErr)
	}
	return i.reportToolResult(result), nil
}

func (i *investigationInvocation) publishFinding(ctx context.Context, input rezai.PublishInvestigationFindingToolInput) (*rezai.InvestigationFindingVersionToolResult, error) {
	evidenceIDs, resolveErr := i.resolveRefs(input.EvidenceRefs)
	if resolveErr != nil {
		return nil, fmt.Errorf("evidence_refs: %w", resolveErr)
	}
	findingReferences, referencesErr := i.findingVersionReferences(input.FindingReferences)
	if referencesErr != nil {
		return nil, referencesErr
	}
	params := rez.PublishInvestigationFindingParams{
		Key:               input.Key,
		Title:             input.Title,
		Body:              input.Body,
		EvidenceIDs:       evidenceIDs,
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
	evidenceIDs, resolveErr := i.resolveRefs(input.EvidenceRefs)
	if resolveErr != nil {
		return nil, fmt.Errorf("evidence_refs: %w", resolveErr)
	}
	findingReferences, referencesErr := i.findingVersionReferences(input.FindingReferences)
	if referencesErr != nil {
		return nil, referencesErr
	}
	params := rez.PublishInvestigationAnswerParams{
		Title:             input.Title,
		Body:              input.Body,
		EvidenceIDs:       evidenceIDs,
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
	evidenceIDs, resolveErr := i.resolveRefs(input.EvidenceRefs)
	if resolveErr != nil {
		return nil, fmt.Errorf("evidence_refs: %w", resolveErr)
	}
	status := strings.TrimSpace(input.Status)
	if status != "open" && status != "supported" && status != "disproven" && status != "inconclusive" {
		return nil, fmt.Errorf("%w: status must be open, supported, disproven, or inconclusive", rez.ErrInvalidInput)
	}
	params := rez.PublishInvestigationHypothesisParams{
		Key:           input.Key,
		Title:         input.Title,
		Justification: input.Justification,
		Status:        invhv.Status(status),
		EvidenceIDs:   evidenceIDs,
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
	versionID, resolveErr := i.resolveRef(input.VersionRef)
	if resolveErr != nil {
		return nil, fmt.Errorf("version_ref: %w", resolveErr)
	}
	result, readErr := i.investigations.GetInvestigationFindingVersion(ctx, i.investigationID, versionID)
	if readErr != nil {
		return nil, fmt.Errorf("read investigation finding version: %w", readErr)
	}
	return i.findingToolResult(result), nil
}

func (i *investigationInvocation) readHypothesis(ctx context.Context, input rezai.ReadInvestigationHypothesisToolInput) (*rezai.InvestigationHypothesisVersionToolResult, error) {
	versionID, resolveErr := i.resolveRef(input.VersionRef)
	if resolveErr != nil {
		return nil, fmt.Errorf("version_ref: %w", resolveErr)
	}
	result, readErr := i.investigations.GetInvestigationHypothesisVersion(ctx, i.investigationID, versionID)
	if readErr != nil {
		return nil, fmt.Errorf("read investigation hypothesis version: %w", readErr)
	}
	return i.hypothesisToolResult(result), nil
}

func (i *investigationInvocation) findingVersionReferences(inputs []rezai.InvestigationFindingReferenceInput) ([]rez.FindingVersionReference, error) {
	references := make([]rez.FindingVersionReference, 0, len(inputs))
	for index, input := range inputs {
		versionID, resolveErr := i.resolveRef(input.VersionRef)
		if resolveErr != nil {
			return nil, fmt.Errorf("finding_references[%d].version_ref: %w", index, resolveErr)
		}
		relation := invfvl.Relation(input.Relation)
		switch relation {
		case invfvl.RelationSupports, invfvl.RelationContradicts, invfvl.RelationInvalidates:
		default:
			return nil, fmt.Errorf("%w: relation must be supports, contradicts, or invalidates", rez.ErrInvalidInput)
		}
		references = append(references, rez.FindingVersionReference{VersionID: versionID, Relation: relation})
	}
	return references, nil
}

func (i *investigationInvocation) reportToolResult(result *rez.InvestigationReportResult) *rezai.InvestigationReportToolResult {
	if result == nil {
		return nil
	}
	return &rezai.InvestigationReportToolResult{
		Text:         result.Text,
		EvidenceRefs: i.formatRefs(result.EvidenceIDs),
		TurnStatus:   string(result.TurnStatus),
		Provisional:  result.TurnStatus == at.StatusRunning,
		CreatedAt:    result.CreatedAt,
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
			VersionRef: i.formatRef(reference.VersionID),
			Relation:   string(reference.Relation),
		})
	}
	return &rezai.InvestigationFindingVersionToolResult{
		VersionRef:               i.formatRef(result.ID),
		Key:                      key,
		IsAnswer:                 isAnswer,
		Title:                    result.Title,
		Body:                     result.Body,
		EvidenceRefs:             i.formatRefs(result.EvidenceIDs),
		FindingReferences:        findingReferences,
		InvalidatedByVersionRefs: i.formatRefs(result.InvalidatedByVersionIDs),
		TurnStatus:               string(result.TurnStatus),
		Provisional:              result.TurnStatus == at.StatusRunning,
		CreatedAt:                result.CreatedAt,
	}
}

func (i *investigationInvocation) hypothesisToolResult(result *rez.InvestigationHypothesisVersion) *rezai.InvestigationHypothesisVersionToolResult {
	if result == nil {
		return nil
	}
	return &rezai.InvestigationHypothesisVersionToolResult{
		VersionRef:    i.formatRef(result.ID),
		Key:           result.Key,
		Title:         result.Title,
		Justification: result.Justification,
		Status:        string(result.Status),
		EvidenceRefs:  i.formatRefs(result.EvidenceIDs),
		TurnStatus:    string(result.TurnStatus),
		Provisional:   result.TurnStatus == at.StatusRunning,
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
		i.formatRef(relationship.ID),
		i.entityNameOrRef(source), i.formatRef(source.ID),
		relationship.Predicate,
		i.entityNameOrRef(target), i.formatRef(target.ID),
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
	line := fmt.Sprintf("entry %s %s: %s", i.formatRef(entry.ID), entry.Kind, limitedAnalysisDescription(entry.Title))
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
	line := "entity " + i.formatRef(entity.ID)
	if strings.TrimSpace(name) != "" {
		line += " " + limitedAnalysisDescription(name)
	}
	line += " [" + entity.Kind + "]"
	if description = limitedAnalysisDescription(description); description != "" {
		line += " — " + description
	}
	return line
}

func (i *investigationInvocation) entityNameOrRef(entity *ent.KnowledgeEntity) string {
	if name := i.entityDisplayName(entity); name != "" {
		return limitedAnalysisDescription(name)
	}
	return i.formatRef(entity.ID)
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

// Refs currently use UUID strings; keep their representation at the tool boundary.
func (i *investigationInvocation) resolveRef(ref string) (uuid.UUID, error) {
	id, parseErr := uuid.Parse(ref)
	if parseErr != nil || id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: invalid reference; use a ref returned by a tool", rez.ErrInvalidInput)
	}
	return id, nil
}

func (i *investigationInvocation) resolveRefs(refs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, len(refs))
	for index, ref := range refs {
		id, resolveErr := i.resolveRef(ref)
		if resolveErr != nil {
			return nil, fmt.Errorf("reference at index %d: %w", index, resolveErr)
		}
		ids[index] = id
	}
	return ids, nil
}

func (i *investigationInvocation) formatRef(id uuid.UUID) string {
	return id.String()
}

func (i *investigationInvocation) formatRefs(ids []uuid.UUID) []string {
	refs := make([]string, len(ids))
	for index, id := range ids {
		refs[index] = i.formatRef(id)
	}
	return refs
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
