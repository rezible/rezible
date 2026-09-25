package ai

import (
	"time"

	"github.com/google/uuid"
	"github.com/invopop/jsonschema"
)

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

type (
	AnalysisPageArgs struct {
		Page     int `json:"page" jsonschema:"description=One-based page number; zero uses the default,minimum=0"`
		PageSize int `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}

	ListAnalysisSubjectsArgs struct {
		Kind     string `json:"kind" jsonschema:"description=Membership collection to list,enum=entity,enum=relationship"`
		Page     int    `json:"page" jsonschema:"description=One-based page number; zero uses the default,minimum=0"`
		PageSize int    `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}

	InspectAnalysisSubjectArgs struct {
		Kind     string    `json:"kind" jsonschema:"description=Knowledge subject record type,enum=entity,enum=relationship"`
		ID       uuid.UUID `json:"id" jsonschema:"description=Canonical knowledge entity or relationship ID,type=string,format=uuid"`
		Page     int       `json:"page" jsonschema:"description=One-based linked-entry page number; zero uses the default,minimum=0"`
		PageSize int       `json:"page_size" jsonschema:"description=Linked-entry page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}

	ReadAnalysisEntryArgs struct {
		EntryID  uuid.UUID `json:"entry_id" jsonschema:"description=Canonical analysis-entry ID returned by an analysis tool,type=string,format=uuid"`
		Page     int       `json:"page" jsonschema:"description=One-based attachment page number; zero uses the default,minimum=0"`
		PageSize int       `json:"page_size" jsonschema:"description=Attachment page size from 1 to 50; zero uses the default,minimum=0,maximum=50"`
	}

	ReadAnalysisEvidenceArgs struct {
		EvidenceID uuid.UUID `json:"evidence_id" jsonschema:"description=Canonical knowledge-evidence ID attached to an analysis entry,type=string,format=uuid"`
	}

	AnalysisToolResult struct {
		Text     string `json:"text"`
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
		Total    int    `json:"total"`
		HasMore  bool   `json:"has_more"`
	}
)

var ListAnalysisSubjectsTool = defineTool[ToolDefinition[ListAnalysisSubjectsArgs, *AnalysisToolResult]](
	"list_analysis_subjects",
	"List entities or relationships included in this investigation's supplied analysis. Results include canonical IDs and are paginated.",
)

var ListAnalysisEntriesTool = defineTool[ToolDefinition[AnalysisPageArgs, *AnalysisToolResult]](
	"list_analysis_entries",
	"List observation and context entries in the supplied analysis. Results include canonical entry IDs and titles, not full bodies.",
)

var InspectAnalysisSubjectTool = defineTool[ToolDefinition[InspectAnalysisSubjectArgs, *AnalysisToolResult]](
	"inspect_analysis_subject",
	"Inspect an included entity or relationship by its canonical ID and list linked observation and context entries. This does not search graph neighbors.",
)

var ReadAnalysisEntryTool = defineTool[ToolDefinition[ReadAnalysisEntryArgs, *AnalysisToolResult]](
	"read_analysis_entry",
	"Read an observation or context entry by its canonical ID and page through its attached evidence and subjects.",
)

var ReadAnalysisEvidenceTool = defineTool[ToolDefinition[ReadAnalysisEvidenceArgs, *AnalysisToolResult]](
	"read_analysis_evidence",
	"Read knowledge evidence attached to an observation or context entry in this analysis. Returns the evidence ID used for publication citations.",
)

type (
	InvestigationFindingReferenceInput struct {
		VersionID uuid.UUID `json:"version_id" jsonschema:"description=Exact finding-version ID,type=string,format=uuid"`
		Relation  string    `json:"relation" jsonschema:"description=How the finding version relates,enum=supports,enum=contradicts,enum=invalidates"`
	}

	InvestigationFindingReference struct {
		VersionID uuid.UUID `json:"version_id" jsonschema:"type=string,format=uuid"`
		Relation  string    `json:"relation"`
	}

	PublishInvestigationReportToolInput struct {
		Text        string      `json:"text"`
		EvidenceIDs []uuid.UUID `json:"evidence_ids" jsonschema:"description=Knowledge-evidence IDs cited by this report; may be empty,format=uuid"`
	}

	PublishInvestigationFindingToolInput struct {
		Key               string                               `json:"key"`
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceIDs       []uuid.UUID                          `json:"evidence_ids" jsonschema:"description=Knowledge-evidence IDs cited by this finding; may be empty,format=uuid"`
		FindingReferences []InvestigationFindingReferenceInput `json:"finding_references,omitempty"`
	}

	PublishInvestigationAnswerToolInput struct {
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceIDs       []uuid.UUID                          `json:"evidence_ids" jsonschema:"description=Knowledge-evidence IDs cited by this answer; may be empty,format=uuid"`
		FindingReferences []InvestigationFindingReferenceInput `json:"finding_references,omitempty"`
	}

	PublishInvestigationHypothesisToolInput struct {
		Key           string      `json:"key"`
		Title         string      `json:"title"`
		Justification string      `json:"justification"`
		Status        string      `json:"status" jsonschema:"enum=open,enum=supported,enum=disproven,enum=inconclusive"`
		EvidenceIDs   []uuid.UUID `json:"evidence_ids" jsonschema:"description=Knowledge-evidence IDs cited by this hypothesis; may be empty,format=uuid"`
	}

	ReadInvestigationReportToolInput struct {
		Selection string `json:"selection,omitempty" jsonschema:"description=Report to read; omitted means latest,enum=latest,enum=completed"`
	}

	ReadInvestigationFindingToolInput struct {
		VersionID uuid.UUID `json:"version_id" jsonschema:"description=Exact finding-version ID,type=string,format=uuid"`
	}

	ReadInvestigationHypothesisToolInput struct {
		VersionID uuid.UUID `json:"version_id" jsonschema:"description=Exact hypothesis-version ID,type=string,format=uuid"`
	}

	InvestigationReportToolResult struct {
		Text        string      `json:"text"`
		EvidenceIDs []uuid.UUID `json:"evidence_ids" jsonschema:"format=uuid"`
		TurnStatus  string      `json:"turn_status"`
		Provisional bool        `json:"provisional"`
		CreatedAt   time.Time   `json:"created_at"`
	}

	InvestigationFindingVersionToolResult struct {
		VersionID               uuid.UUID                       `json:"version_id" jsonschema:"type=string,format=uuid"`
		Key                     string                          `json:"key,omitempty"`
		IsAnswer                bool                            `json:"is_answer"`
		Title                   string                          `json:"title"`
		Body                    string                          `json:"body"`
		EvidenceIDs             []uuid.UUID                     `json:"evidence_ids" jsonschema:"format=uuid"`
		FindingReferences       []InvestigationFindingReference `json:"finding_references"`
		InvalidatedByVersionIDs []uuid.UUID                     `json:"invalidated_by_version_ids" jsonschema:"format=uuid"`
		TurnStatus              string                          `json:"turn_status"`
		Provisional             bool                            `json:"provisional"`
		CreatedAt               time.Time                       `json:"created_at"`
	}

	InvestigationHypothesisVersionToolResult struct {
		VersionID     uuid.UUID   `json:"version_id" jsonschema:"type=string,format=uuid"`
		Key           string      `json:"key"`
		Title         string      `json:"title"`
		Justification string      `json:"justification"`
		Status        string      `json:"status"`
		EvidenceIDs   []uuid.UUID `json:"evidence_ids" jsonschema:"format=uuid"`
		TurnStatus    string      `json:"turn_status"`
		Provisional   bool        `json:"provisional"`
		CreatedAt     time.Time   `json:"created_at"`
	}

	InvestigationOutputPage[T any] struct {
		Items    []T  `json:"items"`
		Page     int  `json:"page"`
		PageSize int  `json:"page_size"`
		Total    int  `json:"total"`
		HasMore  bool `json:"has_more"`
	}
)

var PublishInvestigationReportTool = defineTool[ToolDefinition[PublishInvestigationReportToolInput, *InvestigationReportToolResult]](
	"publish_investigation_report",
	"Publish an explicit plain-text investigation report with knowledge-evidence citations. Empty evidence_ids is valid.",
)

var PublishInvestigationFindingTool = defineTool[ToolDefinition[PublishInvestigationFindingToolInput, *InvestigationFindingVersionToolResult]](
	"publish_investigation_finding",
	"Publish or revise a finding. Cite knowledge-evidence IDs and link exact finding versions when needed.",
)

var PublishInvestigationAnswerTool = defineTool[ToolDefinition[PublishInvestigationAnswerToolInput, *InvestigationFindingVersionToolResult]](
	"publish_investigation_answer",
	"Publish an answer linked to the question assigned to this turn. Cite knowledge-evidence IDs and link exact finding versions when needed.",
)

var PublishInvestigationHypothesisTool = defineTool[ToolDefinition[PublishInvestigationHypothesisToolInput, *InvestigationHypothesisVersionToolResult]](
	"publish_investigation_hypothesis",
	"Publish or revise a hypothesis with status open, supported, disproven, or inconclusive and optional knowledge-evidence citations.",
)

var ReadInvestigationReportTool = defineTool[ToolDefinition[ReadInvestigationReportToolInput, *InvestigationReportToolResult]](
	"read_investigation_report",
	"Read the latest eligible report or latest completed report. Returns null when none is eligible.",
)

var ListInvestigationFindingsTool = defineTool[ToolDefinition[AnalysisPageArgs, *InvestigationOutputPage[InvestigationFindingVersionToolResult]]](
	"list_investigation_findings",
	"List the latest eligible finding and answer versions, ordered by producing turn.",
)

var ListInvestigationHypothesesTool = defineTool[ToolDefinition[AnalysisPageArgs, *InvestigationOutputPage[InvestigationHypothesisVersionToolResult]]](
	"list_investigation_hypotheses",
	"List the latest eligible hypothesis versions, ordered by producing turn.",
)

var ReadInvestigationFindingTool = defineTool[ToolDefinition[ReadInvestigationFindingToolInput, *InvestigationFindingVersionToolResult]](
	"read_investigation_finding",
	"Read an exact finding version, including historical or failed-turn output.",
)

var ReadInvestigationHypothesisTool = defineTool[ToolDefinition[ReadInvestigationHypothesisToolInput, *InvestigationHypothesisVersionToolResult]](
	"read_investigation_hypothesis",
	"Read an exact hypothesis version, including historical or failed-turn output.",
)

type uuidStringArray []string

func (uuidStringArray) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Items.Format = "uuid"
}

// uuidStringArrayProperty keeps UUID slices typed as []uuid.UUID in Go while
// describing their JSON representation as arrays of canonical UUID strings.
func uuidStringArrayProperty(name string, fields ...string) any {
	for _, field := range fields {
		if name == field {
			return uuidStringArray{}
		}
	}
	return nil
}

func (PublishInvestigationReportToolInput) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids")
}

func (PublishInvestigationFindingToolInput) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids")
}

func (PublishInvestigationAnswerToolInput) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids")
}

func (PublishInvestigationHypothesisToolInput) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids")
}

func (InvestigationReportToolResult) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids")
}

func (InvestigationFindingVersionToolResult) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids", "invalidated_by_version_ids")
}

func (InvestigationHypothesisVersionToolResult) JSONSchemaProperty(name string) any {
	return uuidStringArrayProperty(name, "evidence_ids")
}
