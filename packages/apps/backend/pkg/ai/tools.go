package ai

import (
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

type (
	ExploreSystemAnalysisInput struct {
		Search string `json:"search,omitempty"`
	}

	ExploreSystemAnalysisOutput struct {
		Analysis string `json:"analysis"`
	}

	UpdateSystemAnalysisInput struct {
		Search  string                         `json:"search,omitempty"`
		Include []string                       `json:"include,omitempty"`
		Prune   []string                       `json:"prune,omitempty"`
		Entries []SystemAnalysisEntryToolInput `json:"entries,omitempty"`
	}

	UpdateSystemAnalysisOutput struct {
		Analysis string                         `json:"analysis"`
		Counts   SystemAnalysisToolUpdateCounts `json:"counts"`
	}

	SystemAnalysisEntryToolInput struct {
		Kind       string                                `json:"kind"`
		Title      string                                `json:"title"`
		Body       string                                `json:"body,omitempty"`
		OccurredAt *time.Time                            `json:"occurred_at,omitempty"`
		Subjects   []SystemAnalysisEntrySubjectToolInput `json:"subjects,omitempty"`
	}

	SystemAnalysisEntrySubjectToolInput struct {
		Role   string `json:"role"`
		Handle string `json:"handle"`
	}

	SystemAnalysisToolUpdateCounts struct {
		IncludedEntities      int `json:"included_entities"`
		IncludedRelationships int `json:"included_relationships"`
		PrunedEntities        int `json:"pruned_entities"`
		PrunedRelationships   int `json:"pruned_relationships"`
		Entries               int `json:"entries"`
		EntrySubjects         int `json:"entry_subjects"`
	}
)

var ExploreSystemAnalysisTool = defineTool[ToolDefinition[ExploreSystemAnalysisInput, ExploreSystemAnalysisOutput]](
	"explore_system_analysis",
	"Inspect the current system analysis, nearby knowledge graph candidates, and search results as compact text with handles.",
)

var UpdateSystemAnalysisTool = defineTool[ToolDefinition[UpdateSystemAnalysisInput, UpdateSystemAnalysisOutput]](
	"update_system_analysis",
	"Update the current system analysis using handles from the rendered analysis.",
)

type (
	SaveAlertInvestigationReportInput struct {
		Report schematypes.AlertInvestigationReport `json:"report"`
	}

	SaveAlertInvestigationReportOutput struct {
		Saved bool `json:"saved"`
	}
)

var SaveAlertInvestigationReportTool = defineTool[ToolDefinition[SaveAlertInvestigationReportInput, SaveAlertInvestigationReportOutput]](
	"save_alert_investigation_report",
	"Save the alert investigation report for this agent session.",
)
