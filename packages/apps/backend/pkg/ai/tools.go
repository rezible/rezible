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

type TextToolResult struct {
	Text     string `json:"text"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Total    int    `json:"total"`
	HasMore  bool   `json:"has_more"`
}

type (
	ReadEntityArgs struct {
		Ref      string `json:"ref" jsonschema:"description=Opaque entity ref from the introduction or a tool result"`
		Page     int    `json:"page" jsonschema:"description=One-based relationship page number; zero uses the default,minimum=0"`
		PageSize int    `json:"page_size" jsonschema:"description=Relationship page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}
)

var ReadEntityTool = defineTool[ToolDefinition[ReadEntityArgs, *TextToolResult]](
	"read_entity",
	"Read a knowledge entity: its state, linking attributes, aliases and one page of its relationships with their latest evidence, newest first. Relationships to events are counted; list_entity_evidence covers them.",
)

type (
	ListEntityEvidenceArgs struct {
		Ref      string `json:"ref" jsonschema:"description=Opaque entity ref"`
		Start    string `json:"start,omitempty" jsonschema:"description=Optional RFC 3339 earliest effective time"`
		End      string `json:"end,omitempty" jsonschema:"description=Optional RFC 3339 latest effective time"`
		Page     int    `json:"page" jsonschema:"description=One-based page number; zero uses the default,minimum=0"`
		PageSize int    `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}
)

var ListEntityEvidenceTool = defineTool[ToolDefinition[ListEntityEvidenceArgs, *TextToolResult]](
	"list_entity_evidence",
	"List what Rezible has recorded about an entity and its relationships, one line per subject, assertion and source with a count, newest first.",
)

type (
	ReadEvidenceArgs struct {
		Ref string `json:"ref" jsonschema:"description=Opaque knowledge-evidence ref"`
	}
)

var ReadEvidenceTool = defineTool[ToolDefinition[ReadEvidenceArgs, *TextToolResult]](
	"read_evidence",
	"Read one piece of knowledge evidence: its subject, its source and the state it stated. Its ref is what publications cite.",
)

type (
	PublishInvestigationReportToolInput struct {
		Text         string   `json:"text"`
		Summary      string   `json:"summary" jsonschema:"description=One or two plain sentences stating the current conclusion and how certain it is; no Markdown; at most 400 characters"`
		EvidenceRefs []string `json:"evidence_refs" jsonschema:"description=Knowledge-evidence refs cited by this report; may be empty"`
	}

	InvestigationReportToolResult struct {
		VersionRef   string    `json:"version_ref"`
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
	ReadInvestigationOutputToolInput struct {
		VersionRef string `json:"version_ref" jsonschema:"description=Exact report, finding or hypothesis version ref"`
	}

	// InvestigationOutputToolResult sets the output named by Kind: report, finding or hypothesis.
	InvestigationOutputToolResult struct {
		Kind       string                                    `json:"kind"`
		Report     *InvestigationReportToolResult            `json:"report,omitempty"`
		Finding    *InvestigationFindingVersionToolResult    `json:"finding,omitempty"`
		Hypothesis *InvestigationHypothesisVersionToolResult `json:"hypothesis,omitempty"`
	}
)

var ReadInvestigationOutputTool = defineTool[ToolDefinition[ReadInvestigationOutputToolInput, *InvestigationOutputToolResult]](
	"read_investigation_output",
	"Read one exact report, finding or hypothesis version of this investigation, including historical or failed-turn output.",
)
