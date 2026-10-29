package ai

import (
	"errors"
	"time"

	"github.com/rezible/rezible/ent/schema/schematypes"
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

// ErrToolRefused is a tool call refused for a reason the agent can act on, such as a provider read outside its
// window or budget, or a provider that failed. The agent gets the reason as an error result and the turn
// continues.
var ErrToolRefused = errors.New("refused")

const (
	// MaxProviderReadsPerTurn is how many provider reads one agent turn may make, across all its attempts.
	MaxProviderReadsPerTurn = 12
	// MaxProviderReadHighlight is how many items of a read an output may cite, and its highlight keeps.
	MaxProviderReadHighlight = 3
	// MaxCountLogsPoints is the most points a count_logs result returns; the step is chosen for it.
	MaxCountLogsPoints = 60
	// ProviderReadLeadIn is how long before a situation opened its investigation may read, for a baseline.
	ProviderReadLeadIn = 24 * time.Hour

	// MaxNextChecks is how many next checks one report may suggest.
	MaxNextChecks = 3
	// MaxNextCheckLength and MaxNextCheckReasonLength bound a next check's text and its reason, in characters.
	MaxNextCheckLength       = 200
	MaxNextCheckReasonLength = 300
)

// ProviderReadRecord describes the read that ran: what the agent and a person see of it.
type ProviderReadRecord struct {
	Query  string    `json:"query"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
	Step   string    `json:"step,omitempty"`
	Limits []string  `json:"limits"`
	// SourceURL opens the same query and range in the provider.
	SourceURL string `json:"source_url,omitempty"`
}

// ProviderReadItem is one citable item a provider read returned. It is also what a read's highlight stores.
type ProviderReadItem = schematypes.ProviderReadItem

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
		Page     int    `json:"page" jsonschema:"description=One-based relationship page number; zero uses the default"`
		PageSize int    `json:"page_size" jsonschema:"description=Relationship page size from 1 to 50; zero uses the default"`
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
		Page     int    `json:"page" jsonschema:"description=One-based page number; zero uses the default"`
		PageSize int    `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default"`
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
		Text         string                   `json:"text"`
		Summary      string                   `json:"summary" jsonschema:"description=One or two plain sentences stating the current conclusion and how certain it is; no Markdown; at most 400 characters"`
		EvidenceRefs []string                 `json:"evidence_refs" jsonschema:"description=Knowledge-evidence and provider read refs cited by this report; may be empty"`
		NextChecks   []InvestigationNextCheck `json:"next_checks,omitempty" jsonschema:"description=Up to 3 things a responder can look at next, most useful first; may be empty"`
	}

	InvestigationNextCheck struct {
		Check  string `json:"check" jsonschema:"description=Something a responder can look at that Rezible could not read or did not settle; at most 200 characters"`
		Reason string `json:"reason" jsonschema:"description=Why it is worth checking; at most 300 characters"`
	}

	InvestigationReportToolResult struct {
		VersionRef   string                   `json:"version_ref"`
		Text         string                   `json:"text"`
		Summary      string                   `json:"summary"`
		EvidenceRefs []string                 `json:"evidence_refs"`
		NextChecks   []InvestigationNextCheck `json:"next_checks,omitempty"`
		TurnStatus   string                   `json:"turn_status"`
		Provisional  bool                     `json:"provisional"`
		CreatedAt    time.Time                `json:"created_at"`
	}
)

var PublishInvestigationReportTool = defineTool[ToolDefinition[PublishInvestigationReportToolInput, *InvestigationReportToolResult]](
	"publish_investigation_report",
	"Publish an investigation report. text is Markdown. summary is one or two plain sentences stating the current conclusion and its certainty. Cite knowledge-evidence and provider read refs; empty evidence_refs is valid. next_checks are up to 3 checks for a responder, each with why.",
)

type (
	PublishInvestigationFindingToolInput struct {
		Key               string                               `json:"key"`
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceRefs      []string                             `json:"evidence_refs" jsonschema:"description=Knowledge-evidence and provider read refs cited by this finding; may be empty"`
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
	"Publish or revise a finding. Cite knowledge-evidence and provider read refs and link exact finding versions when needed.",
)

type (
	PublishInvestigationAnswerToolInput struct {
		Title             string                               `json:"title"`
		Body              string                               `json:"body"`
		EvidenceRefs      []string                             `json:"evidence_refs" jsonschema:"description=Knowledge-evidence and provider read refs cited by this answer; may be empty"`
		FindingReferences []InvestigationFindingReferenceInput `json:"finding_references,omitempty"`
	}
)

var PublishInvestigationAnswerTool = defineTool[ToolDefinition[PublishInvestigationAnswerToolInput, *InvestigationFindingVersionToolResult]](
	"publish_investigation_answer",
	"Publish an answer linked to the question assigned to this turn. Cite knowledge-evidence and provider read refs and link exact finding versions when needed.",
)

type (
	PublishInvestigationHypothesisToolInput struct {
		Key           string   `json:"key"`
		Title         string   `json:"title"`
		Justification string   `json:"justification"`
		Status        string   `json:"status" jsonschema:"enum=open,enum=supported,enum=disproven,enum=inconclusive"`
		EvidenceRefs  []string `json:"evidence_refs" jsonschema:"description=Knowledge-evidence and provider read refs cited by this hypothesis; may be empty"`
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
	"Publish or revise a hypothesis with status open, supported, disproven, or inconclusive and optional knowledge-evidence and provider read citations.",
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

type (
	ListServicesArgs struct {
		Name     string `json:"name,omitempty" jsonschema:"description=Optional text the service name must contain, compared after normalizing case and separators, for example a name from a log line; empty lists every service"`
		Page     int    `json:"page" jsonschema:"description=One-based page number; zero uses the default"`
		PageSize int    `json:"page_size" jsonschema:"description=Page size from 1 to 50; zero uses the default"`
	}
)

var ListServicesTool = defineTool[ToolDefinition[ListServicesArgs, *TextToolResult]](
	"list_services",
	"List the organization's services with their refs for other tools, optionally only those whose name contains some text. Use it to see which services exist.",
)

type (
	ListChangesArgs struct {
		Service     string `json:"service" jsonschema:"description=A service's ref from the introduction or a tool result\\, or its name\\, for example one named in a log line\\, matched like list_services and required to match one service"`
		Start       string `json:"start" jsonschema:"description=RFC 3339 start of the window; at most 7 days before end"`
		End         string `json:"end" jsonschema:"description=RFC 3339 end of the window"`
		Environment string `json:"environment,omitempty" jsonschema:"description=Optional environment; empty lists every environment"`
	}
)

var ListChangesTool = defineTool[ToolDefinition[ListChangesArgs, *TextToolResult]](
	"list_changes",
	"List a service's deployments in a window, most recent first, with the merged changes linked to each and how. Reads Rezible's data; it is not a provider read.",
)

type (
	SearchLogsArgs struct {
		Service  string `json:"service" jsonschema:"description=A service's ref from the introduction or a tool result\\, or its name\\, for example one named in a log line\\, matched like list_services and required to match one service"`
		Start    string `json:"start" jsonschema:"description=RFC 3339 start; within the provider read window and at most 6 hours before end"`
		End      string `json:"end" jsonschema:"description=RFC 3339 end; within the provider read window"`
		Contains string `json:"contains,omitempty" jsonschema:"description=Optional literal text each line must contain"`
		Limit    int    `json:"limit,omitempty" jsonschema:"description=Lines to return from 1 to 200; zero uses 50"`
	}

	SearchLogsToolResult struct {
		ProviderReadRef string `json:"provider_read_ref"`
		ServiceRef      string `json:"service_ref"`
		// ServiceName is the name the service's logs were read by.
		ServiceName string `json:"service_name"`
		// Labels are the labels every returned line carries with the same value.
		Labels map[string]string `json:"labels,omitempty"`
		// Items are the lines, newest first, lines identical or differing only in identifiers grouped into one.
		Items []ProviderReadItem `json:"items"`
		// Limits are the read's limits that applied, such as a truncated result.
		Limits          []string `json:"limits,omitempty"`
		Note            string   `json:"note,omitempty"`
		BudgetRemaining int      `json:"budget_remaining"`
	}
)

var SearchLogsTool = defineTool[ToolDefinition[SearchLogsArgs, *SearchLogsToolResult]](
	"search_logs",
	"Provider read: search a service's log lines, newest first, optionally only those containing some text; any service can be read, by ref or by name. Counts against the turn's provider read budget; cite its provider_read_ref, or up to three of its items as <provider_read_ref>#<item ref>.",
)

type (
	CountLogsArgs struct {
		Service  string `json:"service" jsonschema:"description=A service's ref from the introduction or a tool result\\, or its name\\, for example one named in a log line\\, matched like list_services and required to match one service"`
		Start    string `json:"start" jsonschema:"description=RFC 3339 start; within the provider read window and at most 6 hours before end"`
		End      string `json:"end" jsonschema:"description=RFC 3339 end; within the provider read window"`
		Contains string `json:"contains,omitempty" jsonschema:"description=Optional literal text each counted line must contain"`
	}

	CountLogsToolResult struct {
		ProviderReadRef string `json:"provider_read_ref"`
		ServiceRef      string `json:"service_ref"`
		ServiceName     string `json:"service_name"`
		// Items are the counted points, each with its value.
		Items []ProviderReadItem `json:"items"`
		// Limits are the read's limits that applied, such as a raised step.
		Limits          []string `json:"limits,omitempty"`
		Note            string   `json:"note,omitempty"`
		BudgetRemaining int      `json:"budget_remaining"`
	}
)

var CountLogsTool = defineTool[ToolDefinition[CountLogsArgs, *CountLogsToolResult]](
	"count_logs",
	"Provider read: count a service's log lines over time, optionally only those containing some text; any service can be read, by ref or by name. The step is chosen for the range. Counts against the turn's provider read budget; cite its provider_read_ref, or up to three of its items as <provider_read_ref>#<item ref>.",
)
