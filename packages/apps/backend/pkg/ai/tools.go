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
	QueryKnowledgeGraphInput struct {
		EntityID string `json:"entity_id"`
		Depth    int    `json:"depth,omitempty"`
	}

	QueryKnowledgeGraphOutput struct {
		RootEntityID  string                           `json:"root_entity_id"`
		Entities      []KnowledgeGraphToolEntity       `json:"entities"`
		Relationships []KnowledgeGraphToolRelationship `json:"relationships"`
		Evidence      []KnowledgeGraphToolEvidence     `json:"evidence"`
		Truncated     bool                             `json:"truncated"`
	}

	KnowledgeGraphToolEntity struct {
		ID      string                                 `json:"id"`
		Kind    string                                 `json:"kind"`
		Subkind string                                 `json:"subkind"`
		State   schematypes.KnowledgeGraphSubjectState `json:"state"`
	}

	KnowledgeGraphToolRelationship struct {
		ID       string                                 `json:"id"`
		Kind     string                                 `json:"kind"`
		Subkind  string                                 `json:"subkind"`
		SourceID string                                 `json:"source_id"`
		TargetID string                                 `json:"target_id"`
		State    schematypes.KnowledgeGraphSubjectState `json:"state"`
	}

	KnowledgeGraphToolEvidence struct {
		ID             string         `json:"id"`
		Assertion      string         `json:"assertion"`
		EvidenceKind   string         `json:"evidence_kind"`
		EffectiveAt    time.Time      `json:"effective_at"`
		Properties     map[string]any `json:"properties"`
		EntityID       string         `json:"entity_id,omitempty"`
		RelationshipID string         `json:"relationship_id,omitempty"`
	}
)

var QueryKnowledgeGraphTool = defineTool[ToolDefinition[QueryKnowledgeGraphInput, QueryKnowledgeGraphOutput]](
	"query_knowledge_graph",
	"Get evidence-backed entities and relationships around a knowledge graph entity.",
)

type (
	RecordKnowledgeCitationsInput struct {
		Citations []KnowledgeCitation `json:"citations"`
	}

	KnowledgeCitation struct {
		EvidenceID string `json:"evidence_id"`
		Summary    string `json:"summary"`
	}

	RecordKnowledgeCitationsOutput struct {
		Recorded int `json:"recorded"`
	}
)

var RecordKnowledgeCitationsTool = defineTool[ToolDefinition[RecordKnowledgeCitationsInput, RecordKnowledgeCitationsOutput]](
	"record_knowledge_citations",
	"Record only the retrieved evidence that directly supports claims in the response.",
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
