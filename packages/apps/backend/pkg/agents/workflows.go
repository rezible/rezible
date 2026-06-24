package agents

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkflowKind string

func (k WorkflowKind) String() string {
	return string(k)
}

const (
	WorkflowKindIncidentContextPack WorkflowKind = "incident_context_pack"
	WorkflowKindAlertInvestigation  WorkflowKind = "alert_investigation"
)

type SubjectRef struct {
	Type string    `json:"type" validate:"required"`
	ID   uuid.UUID `json:"id" validate:"required"`
}

type IncidentContextPackInput struct {
	Subjects   []SubjectRef `json:"subjects" validate:"required,dive"`
	Objectives []string     `json:"objectives"`
}

type IncidentContextPackOutput struct {
	Findings           IncidentTriageFindings `json:"findings"`
	Limitations        []string               `json:"limitations"`
	RecommendedActions []string               `json:"recommendedActions"`
}

type IncidentImpactFinding struct {
	EntityID    string   `json:"entityId"`
	DisplayName string   `json:"displayName"`
	Rationale   string   `json:"rationale"`
	EvidenceIDs []string `json:"evidenceIds"`
}

type IncidentTriageFindings struct {
	LikelyImpact    []IncidentImpactFinding `json:"likelyImpact"`
	SuggestedChecks []string                `json:"suggestedChecks"`
}

type AlertInvestigationInput struct {
	Subjects   []SubjectRef `json:"subjects" validate:"required,dive"`
	Objectives []string     `json:"objectives"`
}

type AlertInvestigationOutput struct {
	Findings           AlertInvestigationFindings `json:"findings"`
	Limitations        []string                   `json:"limitations"`
	RecommendedActions []string                   `json:"recommendedActions"`
}

type AlertInvestigationFindings struct {
	LikelyCause     string   `json:"likelyCause"`
	AffectedSystems []string `json:"affectedSystems"`
	SuggestedChecks []string `json:"suggestedChecks"`
	RecommendedNext string   `json:"recommendedNext"`
}

type WorkflowContextItem struct {
	Kind     string         `json:"kind"`
	Role     string         `json:"role,omitempty"`
	Name     string         `json:"name"`
	Payload  map[string]any `json:"payload,omitempty"`
	Citation int            `json:"citation,omitempty"`
}

type WorkflowContext struct {
	GeneratedAt time.Time
	Context     map[string]any
	Items       []WorkflowContextItem
	Citations   []RunCitationInput
	Limitations []string
	Suggested   []string
}

func ValidateWorkflowInput(kind WorkflowKind, input map[string]any) error {
	if kind == "" {
		return fmt.Errorf("missing agent workflow kind")
	}
	switch kind {
	case WorkflowKindIncidentContextPack:
		params, err := DecodeInput[IncidentContextPackInput](input)
		if err != nil {
			return fmt.Errorf("validate incident context pack input: %w", err)
		}
		if FindSubjectRefID(params.Subjects, "incident") == uuid.Nil {
			return fmt.Errorf("%s requires incident subject", kind)
		}
	case WorkflowKindAlertInvestigation:
		params, err := DecodeInput[AlertInvestigationInput](input)
		if err != nil {
			return fmt.Errorf("validate alert investigation input: %w", err)
		}
		if FindSubjectRefID(params.Subjects, "alert") == uuid.Nil {
			return fmt.Errorf("%s requires alert subject", kind)
		}
	default:
		return fmt.Errorf("unknown agent workflow kind %q", kind)
	}
	return nil
}
