package ai

import "time"

type ToolDefinition[I any, O any] struct {
	name        string
	description string
}

func (d ToolDefinition[I, O]) Name() string {
	return d.name
}

func (d ToolDefinition[I, O]) Description() string {
	return d.description
}

func defineTool[D ToolDefinition[I, O], I any, O any](name, description string) D {
	return D{name: name, description: description}
}

type AnalysisToolResult struct {
	Text     string `json:"text"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int    `json:"total"`
	HasMore  bool   `json:"has_more"`
}

type (
	ListAnalysisSubjectsArgs struct {
		Kind     string `json:"kind" jsonschema:"description=Membership collection to list,enum=entity,enum=relationship"`
		Page     int    `json:"page" jsonschema:"description=One-based page number; zero uses the default,minimum=0"`
		PageSize int    `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}
)

var ListAnalysisSubjectsTool = defineTool[ToolDefinition[ListAnalysisSubjectsArgs, *AnalysisToolResult]](
	"list_analysis_subjects",
	"List entities or relationships included in this investigation's supplied analysis. Results include opaque refs and are paginated.",
)

type (
	AnalysisPageArgs struct {
		Page     int `json:"page" jsonschema:"description=One-based page number; zero uses the default,minimum=0"`
		PageSize int `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}
)

var ListAnalysisEntriesTool = defineTool[ToolDefinition[AnalysisPageArgs, *AnalysisToolResult]](
	"list_analysis_entries",
	"List observation and context entries in the supplied analysis. Results include entry refs and titles, not full bodies.",
)

type (
	InspectAnalysisSubjectArgs struct {
		Kind     string `json:"kind" jsonschema:"description=Knowledge subject record type,enum=entity,enum=relationship"`
		Ref      string `json:"ref" jsonschema:"description=Opaque entity or relationship ref returned by an analysis tool"`
		Page     int    `json:"page" jsonschema:"description=One-based linked-entry page number; zero uses the default,minimum=0"`
		PageSize int    `json:"page_size" jsonschema:"description=Linked-entry page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}
)

var InspectAnalysisSubjectTool = defineTool[ToolDefinition[InspectAnalysisSubjectArgs, *AnalysisToolResult]](
	"inspect_analysis_subject",
	"Inspect an included entity or relationship by its ref and list linked observation and context entries. This does not search graph neighbors.",
)

type (
	ReadAnalysisEntryArgs struct {
		Ref      string `json:"ref" jsonschema:"description=Opaque entry ref returned by an analysis tool"`
		Page     int    `json:"page" jsonschema:"description=One-based attachment page number; zero uses the default,minimum=0"`
		PageSize int    `json:"page_size" jsonschema:"description=Attachment page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}
)

var ReadAnalysisEntryTool = defineTool[ToolDefinition[ReadAnalysisEntryArgs, *AnalysisToolResult]](
	"read_analysis_entry",
	"Read an observation or context entry by its ref and page through its attached evidence and subjects.",
)

type (
	ReadAnalysisEvidenceArgs struct {
		Ref string `json:"ref" jsonschema:"description=Opaque knowledge-evidence ref returned by an analysis tool"`
	}
)

var ReadAnalysisEvidenceTool = defineTool[ToolDefinition[ReadAnalysisEvidenceArgs, *AnalysisToolResult]](
	"read_analysis_evidence",
	"Read knowledge evidence attached to an observation or context entry in this analysis. Returns the evidence ref used for publication citations.",
)

type (
	PublishInvestigationReportToolInput struct {
		Text         string   `json:"text"`
		Summary      string   `json:"summary" jsonschema:"description=One or two plain sentences stating the current conclusion and how certain it is; no Markdown; at most 400 characters"`
		EvidenceRefs []string `json:"evidence_refs" jsonschema:"description=Knowledge-evidence refs cited by this report; may be empty"`
	}

	InvestigationReportToolResult struct {
		Text         string    `json:"text"`
		Summary      string    `json:"summary"`
		EvidenceRefs []string  `json:"evidence_refs"`
		TurnStatus   string    `json:"turn_status"`
		Provisional  bool      `json:"provisional"`
		CreatedAt    time.Time `json:"created_at"`
	}
)

var PublishInvestigationReportTool = defineTool[ToolDefinition[PublishInvestigationReportToolInput, *InvestigationReportToolResult]](
	"publish_investigation_report",
	"Publish an investigation report. text is Markdown. summary is one or two plain sentences stating the current conclusion and its certainty. Cite knowledge-evidence refs; empty evidence_refs is valid.",
)

type (
	ReadInvestigationReportToolInput struct {
		Selection string `json:"selection,omitempty" jsonschema:"description=Report to read; omitted means latest,enum=latest,enum=completed"`
	}
)

var ReadInvestigationReportTool = defineTool[ToolDefinition[ReadInvestigationReportToolInput, *InvestigationReportToolResult]](
	"read_investigation_report",
	"Read the latest eligible report or latest completed report. Returns null when none is eligible.",
)

type (
	PublishInvestigationFindingToolInput struct {
		Key               string                               `json:"key"`
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceRefs      []string                             `json:"evidence_refs" jsonschema:"description=Knowledge-evidence refs cited by this finding; may be empty"`
		FindingReferences []InvestigationFindingReferenceInput `json:"finding_references,omitempty"`
	}

	InvestigationFindingVersionToolResult struct {
		VersionRef               string                          `json:"version_ref"`
		Key                      string                          `json:"key,omitempty"`
		IsAnswer                 bool                            `json:"is_answer"`
		Title                    string                          `json:"title"`
		Body                     string                          `json:"body"`
		EvidenceRefs             []string                        `json:"evidence_refs"`
		FindingReferences        []InvestigationFindingReference `json:"finding_references"`
		InvalidatedByVersionRefs []string                        `json:"invalidated_by_version_refs"`
		TurnStatus               string                          `json:"turn_status"`
		Provisional              bool                            `json:"provisional"`
		CreatedAt                time.Time                       `json:"created_at"`
	}

	InvestigationFindingReferenceInput struct {
		VersionRef string `json:"version_ref" jsonschema:"description=Exact finding-version ref returned by a tool"`
		Relation   string `json:"relation" jsonschema:"description=How the finding version relates,enum=supports,enum=contradicts,enum=invalidates"`
	}

	InvestigationFindingReference struct {
		VersionRef string `json:"version_ref"`
		Relation   string `json:"relation"`
	}
)

var PublishInvestigationFindingTool = defineTool[ToolDefinition[PublishInvestigationFindingToolInput, *InvestigationFindingVersionToolResult]](
	"publish_investigation_finding",
	"Publish or revise a finding. Cite knowledge-evidence refs and link exact finding versions when needed.",
)

type (
	PublishInvestigationAnswerToolInput struct {
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceRefs      []string                             `json:"evidence_refs" jsonschema:"description=Knowledge-evidence refs cited by this answer; may be empty"`
		FindingReferences []InvestigationFindingReferenceInput `json:"finding_references,omitempty"`
	}
)

var PublishInvestigationAnswerTool = defineTool[ToolDefinition[PublishInvestigationAnswerToolInput, *InvestigationFindingVersionToolResult]](
	"publish_investigation_answer",
	"Publish an answer linked to the question assigned to this turn. Cite knowledge-evidence refs and link exact finding versions when needed.",
)

type (
	ReadInvestigationFindingToolInput struct {
		VersionRef string `json:"version_ref" jsonschema:"description=Exact finding-version ref returned by a tool"`
	}
)

var ReadInvestigationFindingTool = defineTool[ToolDefinition[ReadInvestigationFindingToolInput, *InvestigationFindingVersionToolResult]](
	"read_investigation_finding",
	"Read an exact finding version, including historical or failed-turn output.",
)

type (
	InvestigationOutputPage[T any] struct {
		Items    []T  `json:"items"`
		Page     int  `json:"page"`
		PageSize int  `json:"page_size"`
		Total    int  `json:"total"`
		HasMore  bool `json:"has_more"`
	}
)

var ListInvestigationFindingsTool = defineTool[ToolDefinition[AnalysisPageArgs, *InvestigationOutputPage[InvestigationFindingVersionToolResult]]](
	"list_investigation_findings",
	"List the latest eligible finding and answer versions, ordered by producing turn.",
)

type (
	PublishInvestigationHypothesisToolInput struct {
		Key           string   `json:"key"`
		Title         string   `json:"title"`
		Justification string   `json:"justification"`
		Status        string   `json:"status" jsonschema:"enum=open,enum=supported,enum=disproven,enum=inconclusive"`
		EvidenceRefs  []string `json:"evidence_refs" jsonschema:"description=Knowledge-evidence refs cited by this hypothesis; may be empty"`
	}

	InvestigationHypothesisVersionToolResult struct {
		VersionRef    string    `json:"version_ref"`
		Key           string    `json:"key"`
		Title         string    `json:"title"`
		Justification string    `json:"justification"`
		Status        string    `json:"status"`
		EvidenceRefs  []string  `json:"evidence_refs"`
		TurnStatus    string    `json:"turn_status"`
		Provisional   bool      `json:"provisional"`
		CreatedAt     time.Time `json:"created_at"`
	}
)

var PublishInvestigationHypothesisTool = defineTool[ToolDefinition[PublishInvestigationHypothesisToolInput, *InvestigationHypothesisVersionToolResult]](
	"publish_investigation_hypothesis",
	"Publish or revise a hypothesis with status open, supported, disproven, or inconclusive and optional knowledge-evidence citations.",
)

type (
	ReadInvestigationHypothesisToolInput struct {
		VersionRef string `json:"version_ref" jsonschema:"description=Exact hypothesis-version ref returned by a tool"`
	}
)

var ReadInvestigationHypothesisTool = defineTool[ToolDefinition[ReadInvestigationHypothesisToolInput, *InvestigationHypothesisVersionToolResult]](
	"read_investigation_hypothesis",
	"Read an exact hypothesis version, including historical or failed-turn output.",
)

var ListInvestigationHypothesesTool = defineTool[ToolDefinition[AnalysisPageArgs, *InvestigationOutputPage[InvestigationHypothesisVersionToolResult]]](
	"list_investigation_hypotheses",
	"List the latest eligible hypothesis versions, ordered by producing turn.",
)
