package ai

import (
	"time"

	"github.com/google/uuid"
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
	KnowledgeEntitySummary struct {
		ID          uuid.UUID `json:"id"`
		Category    string    `json:"category"`
		Kind        string    `json:"kind"`
		DisplayName string    `json:"display_name"`
	}

	KnowledgeRelationshipSummary struct {
		ID          uuid.UUID `json:"id"`
		Predicate   string    `json:"predicate"`
		DisplayName string    `json:"display_name"`
	}

	KnowledgeSubjectAliasSummary struct {
		ID                uuid.UUID `json:"id"`
		Provider          string    `json:"provider"`
		ProviderNamespace string    `json:"provider_namespace"`
		ResourceRef       string    `json:"resource_ref"`
	}

	KnowledgeEvidenceSummary struct {
		ID          uuid.UUID `json:"id"`
		Kind        string    `json:"kind"`
		Assertion   string    `json:"assertion"`
		EffectiveAt time.Time `json:"effective_at"`
		CreatedAt   time.Time `json:"created_at"`
	}
)

type (
	SummarizeSystemNeighborhoodToolInput struct {
		EntityID *string `json:"entity_id,omitempty" jsonschema:"description=Knowledge entity to summarize. Omit to use the default analysis subject"`
	}
	SummarizeSystemNeighborhoodToolOutput struct {
		IncomingRelationships map[string]SystemNeighborhoodGroupSummary `json:"incoming_relationships"`
		OutgoingRelationships map[string]SystemNeighborhoodGroupSummary `json:"outgoing_relationships"`
	}
	SystemNeighborhoodGroupSummary struct {
		Count int `json:"count"`
	}
)

var SummarizeSystemNeighborhoodTool = defineTool[ToolDefinition[SummarizeSystemNeighborhoodToolInput, SummarizeSystemNeighborhoodToolOutput]](
	"summarize_system_neighborhood",
	"Summarize the one-hop knowledge graph neighborhood of the system analysis subject or an included entity, grouped by direction and relationship predicate.",
)

type (
	ExploreSystemNeighborhoodToolInput struct {
		EntityID              *string `json:"entity_id,omitempty" jsonschema:"description=Knowledge entity to explore. Omit to use the default analysis subject"`
		RelationshipPredicate *string `json:"relationship_predicate,omitempty" jsonschema:"description=Optional exact relationship predicate,minLength=1"`
		NeighborCategory      *string `json:"neighbor_category,omitempty" jsonschema:"description=Optional category of the entity at the opposite endpoint,minLength=1"`
		Offset                *int    `json:"offset,omitempty" jsonschema:"description=Zero-based result offset,minimum=0"`
	}

	ExploreSystemNeighborhoodToolOutput struct {
		Neighbours []SystemEntityNeighbor `json:"neighbours"`
		NextOffset *int                   `json:"next_offset,omitempty"`
	}
	SystemEntityNeighbor struct {
		Relationship KnowledgeRelationshipSummary `json:"relationship"`
		Entity       KnowledgeEntitySummary       `json:"entity"`
	}
)

var ExploreSystemNeighborhoodTool = defineTool[ToolDefinition[ExploreSystemNeighborhoodToolInput, ExploreSystemNeighborhoodToolOutput]](
	"explore_system_neighborhood",
	"Query neighboring related entities. Exploration is read-only.",
)

type (
	InspectKnowledgeSubjectToolInput struct {
		SubjectKind string `json:"subject_kind" jsonschema:"description=Kind of knowledge subject to inspect,enum=entity,enum=relationship,enum=evidence"`
		SubjectID   string `json:"subject_id" jsonschema:"description=Knowledge subject id"`
	}
	InspectKnowledgeSubjectToolOutput struct {
		Entity       *KnowledgeEntityDetail       `json:"entity,omitempty"`
		Relationship *KnowledgeRelationshipDetail `json:"relationship,omitempty"`
		Evidence     *KnowledgeEvidenceDetail     `json:"evidence,omitempty"`
	}
	KnowledgeEntityDetail struct {
		Summary     KnowledgeEntitySummary        `json:"summary"`
		Description string                        `json:"description,omitempty"`
		Properties  map[string]any                `json:"properties"`
		Aliases     []KnowledgeSubjectAliasDetail `json:"aliases"`
	}
	KnowledgeRelationshipDetail struct {
		Summary     KnowledgeRelationshipSummary  `json:"summary"`
		Description string                        `json:"description,omitempty"`
		Properties  map[string]any                `json:"properties"`
		Aliases     []KnowledgeSubjectAliasDetail `json:"aliases"`
		Source      KnowledgeEntitySummary        `json:"source"`
		Target      KnowledgeEntitySummary        `json:"target"`
	}
	KnowledgeSubjectAliasDetail struct {
		Summary        KnowledgeSubjectAliasSummary `json:"summary"`
		LatestEvidence *KnowledgeEvidenceSummary    `json:"latest_evidence,omitempty"`
	}
	KnowledgeEvidenceDetail struct {
		Summary     KnowledgeEvidenceSummary     `json:"summary"`
		Description string                       `json:"description,omitempty"`
		Properties  map[string]any               `json:"properties"`
		Alias       KnowledgeSubjectAliasSummary `json:"alias"`
	}
)

var InspectKnowledgeSubjectTool = defineTool[ToolDefinition[InspectKnowledgeSubjectToolInput, InspectKnowledgeSubjectToolOutput]](
	"inspect_knowledge_subject",
	`Get detailed current information and evidence for one knowledge entity, relationship, or evidence record.
	Use IDs returned by neighborhood exploration or another trusted tool.`,
)

type (
	IncludeAnalysisSubjectsToolInput struct {
		Subjects []IncludeAnalysisSubjectToolInputSubject `json:"subjects" jsonschema:"description=Knowledge subjects to include,minItems=1,maxItems=20"`
	}
	IncludeAnalysisSubjectToolInputSubject struct {
		SubjectKind string `json:"subject_kind" jsonschema:"description=Kind of knowledge subject to include,enum=entity,enum=relationship"`
		SubjectID   string `json:"subject_id" jsonschema:"description=Knowledge subject id"`
	}

	IncludeAnalysisSubjectsToolOutput struct {
		Included int `json:"included"`
	}
)

var IncludeAnalysisSubjectsTool = defineTool[ToolDefinition[IncludeAnalysisSubjectsToolInput, IncludeAnalysisSubjectsToolOutput]](
	"include_analysis_subjects",
	`Include chosen knowledge entities and relationships in the current system analysis subgraph.
	Including a relationship also includes both related entities. This does not record a finding.`,
)

type (
	RecordAnalysisFindingToolInput struct {
		Reference string                                   `json:"reference" jsonschema:"description=A unique finding reference to create or update,minLength=1"`
		Title     string                                   `json:"title" jsonschema:"description=Concise finding title,minLength=1"`
		Detail    string                                   `json:"body,omitempty" jsonschema:"description=Optional supporting detail for the finding"`
		Subjects  []AnalysisFindingSubjectToolInputSubject `json:"subjects" jsonschema:"description=Knowledge subjects supporting the finding; at least one must be evidence,minItems=1,maxItems=20"`
	}
	AnalysisFindingSubjectToolInputSubject struct {
		SubjectKind string `json:"subject_kind" jsonschema:"description=Kind of cited knowledge subject,enum=entity,enum=relationship,enum=evidence"`
		SubjectID   string `json:"subject_id" jsonschema:"description=Knowledge subject id"`
		Role        string `json:"role" jsonschema:"description=Concise role such as primary or affected or contributing or evidence_for,minLength=1"`
	}

	RecordAnalysisFindingToolOutput struct {
		Reference    string `json:"reference"`
		Sequence     int    `json:"sequence"`
		Title        string `json:"title"`
		SubjectCount int    `json:"subject_count"`
	}
)

var RecordAnalysisFindingTool = defineTool[ToolDefinition[RecordAnalysisFindingToolInput, RecordAnalysisFindingToolOutput]](
	"record_analysis_finding",
	"Record an evidence-backed finding in the current system analysis. Cite at least one inspected evidence record and any relevant knowledge entities or relationships as subjects. Use concise roles such as primary, affected, contributing, and evidence_for.",
)

type (
	SaveSituationInvestigationReportToolInput struct {
		Report schematypes.SituationInvestigationReport `json:"report" jsonschema:"description=Completed situation investigation report"`
	}

	SaveSituationInvestigationReportToolOutput struct {
		Saved bool `json:"saved"`
	}
)

var SaveSituationInvestigationReportTool = defineTool[ToolDefinition[SaveSituationInvestigationReportToolInput, SaveSituationInvestigationReportToolOutput]](
	"save_situation_investigation_report",
	"Save the situation investigation report for this agent session.",
)
