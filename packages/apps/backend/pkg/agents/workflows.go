package agents

type workflowDefinition[Input any, Output any] struct {
	name   string
	input  Input
	output Output
}

func defineWorkflow[Input any, Output any](name string) workflowDefinition[Input, Output] {
	return workflowDefinition[Input, Output]{name: name}
}

var WorkflowAlertInvestigation = defineWorkflow[AlertInvestigationInput, AlertInvestigationOutput]("alert_investigation")

type (
	AlertInvestigationInput struct {
		Objectives []string `json:"objectives"`
	}

	AlertInvestigationOutput struct {
		Limitations        []string `json:"limitations"`
		RecommendedActions []string `json:"recommendedActions"`
	}

	AlertInvestigationFindings struct {
		LikelyCause     string   `json:"likelyCause"`
		AffectedSystems []string `json:"affectedSystems"`
		SuggestedChecks []string `json:"suggestedChecks"`
		RecommendedNext string   `json:"recommendedNext"`
	}
)

var WorkflowIncidentTriage = defineWorkflow[IncidentTriageInput, IncidentTriageOutput]("incident_triage")

type (
	IncidentTriageInput struct {
		Objectives []string `json:"objectives"`
	}

	IncidentTriageOutput struct {
		LikelyImpact       []IncidentImpactFinding `json:"likelyImpact"`
		SuggestedChecks    []string                `json:"suggestedChecks"`
		Limitations        []string                `json:"limitations"`
		RecommendedActions []string                `json:"recommendedActions"`
	}

	IncidentImpactFinding struct {
		EntityID    string   `json:"entityId"`
		DisplayName string   `json:"displayName"`
		Rationale   string   `json:"rationale"`
		EvidenceIDs []string `json:"evidenceIds"`
	}
)
