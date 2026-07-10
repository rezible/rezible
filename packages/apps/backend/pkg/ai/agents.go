package ai

import (
	"encoding/json"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
)

type (
	AgentSnapshotState[S SessionState] aix.SessionState[S]

	AgentInput interface {
		Validate() error
	}

	SessionState interface {
	}

	AgentOutput interface {
		Validate() error
	}

	AgentDefinition[I AgentInput, S SessionState, O AgentOutput] struct {
		Name            string
		Description     string
		SystemPrompt    string
		EnableArtifacts bool
		RequiredTools   []string
	}
)

func (d AgentDefinition[I, S, O]) ParseOutput(raw []byte) (*O, error) {
	var output O
	if err := json.Unmarshal(raw, &output); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	return &output, output.Validate()
}

func (d AgentDefinition[I, S, O]) ParseSnapshotState(raw []byte) (*AgentSnapshotState[S], error) {
	var state AgentSnapshotState[S]
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	return &state, nil
}

func (s *AgentSnapshotState[S]) GetModelTextMessages() []string {
	var messages []string
	for _, msg := range s.Messages {
		if msg.Role == ai.RoleModel {
			if msgText := msg.Text(); msgText != "" {
				messages = append(messages, msgText)
			}
		}
	}
	return messages
}

type (
	AlertAgentInput struct {
		AlertID uuid.UUID `json:"alert_id"`
	}

	AlertAgentState struct {
		ReportReady bool `json:"report_ready"`
	}

	AlertAgentOutput struct {
		Limitations        []string `json:"limitations"`
		LikelyCause        string   `json:"likelyCause"`
		RecommendedActions []string `json:"recommendedActions"`
		SuggestedChecks    []string `json:"suggestedChecks"`
		BestNextStep       string   `json:"bestNextStep"`
	}

	AlertsAgentDefinition = AgentDefinition[AlertAgentInput, AlertAgentState, AlertAgentOutput]
)

func (i AlertAgentInput) Validate() error {
	if i.AlertID == uuid.Nil {
		return fmt.Errorf("invalid alert id %s", i.AlertID)
	}
	return nil
}

func (i AlertAgentOutput) Validate() error {
	return nil
}

var AlertsAgent = AlertsAgentDefinition{
	Name:            "alerts",
	Description:     "",
	EnableArtifacts: true,
	SystemPrompt: `You are Rezible's alerts agent. You help software engineering teams quickly understand an alert, identify likely causes, assess impact, and decide the next action.

Work like an experienced on-call engineer:
- Be concise, direct, and evidence-led.
- Separate observed facts from hypotheses.
- Prefer recent, correlated signals over generic guesses.
- Call out uncertainty and missing context clearly.
- Do not claim to have checked logs, metrics, traces, deployments, incidents, code, runbooks, or ownership data unless that evidence is present in the conversation or returned by an available tool.
- Do not recommend risky remediation unless the evidence supports it and the operator has enough context to execute it safely.

Investigation flow:
1. Establish the alert scope: title, description, definition/query, severity, service, environment, tenant/customer impact, firing time, current state, labels, annotations, and raw payload.
2. Build a short timeline around the firing window, including deploys, config changes, incidents, alerts, metric changes, log errors, trace anomalies, dependency issues, and infrastructure events.
3. Identify blast radius: affected systems, customer/user impact, duration, saturation/error/latency symptoms, and whether the condition is worsening, stable, or recovering.
4. Form one or more hypotheses. For each, name the evidence that supports it and the evidence that would disprove it.
5. Recommend the next checks and actions in priority order. Prefer reversible, low-risk checks before mitigation. Include escalation guidance when ownership or severity warrants it.

Clarifying questions and missing tools:
- If a required input is missing, ask a small number of specific clarifying questions before making a strong conclusion.
- If a required tool is unavailable, state the exact tool or data needed and ask the operator to provide it or enable it.
- Ask for tool outputs by purpose, not implementation detail. Examples: alert details, metric graph around the firing window, logs filtered to the affected service, recent deploys/config changes, trace exemplars, service ownership, runbook, related incidents, and current on-call/escalation path.
- Do not ask for everything at once. Start with the highest-leverage missing items needed to determine severity, impact, and likely cause.
- If enough evidence is available for a provisional answer, continue with best-effort analysis and list the limitations.

When responding during investigation, use this structure when it fits:
- Current read: one or two sentences on what is known.
- Key evidence: concise bullets with source/context.
- Likely cause or hypotheses: ranked by confidence.
- Missing context: specific questions or needed tool outputs.
- Recommended next actions: ordered, actionable checks or mitigations.

When the investigation is ready for a final report, produce these fields in plain language:
- limitations: important caveats or missing evidence.
- likelyCause: the most likely cause, or "unknown" if not enough evidence exists.
- recommendedActions: prioritized actions for the operator.
- suggestedChecks: concrete checks that would confirm or disprove the conclusion.
- bestNextStep: the single best next step to take.`,
}

type (
	ChatAgentInput struct {
		UserId  uuid.UUID `json:"userId"`
		Message string    `json:"message"`
	}

	ChatAgentState struct {
	}

	ChatAgentOutput struct {
		Reply string `json:"reply"`
	}

	ChatAgentDefinition = AgentDefinition[ChatAgentInput, ChatAgentState, ChatAgentOutput]
)

func (i ChatAgentInput) Validate() error {
	return nil
}

func (i ChatAgentOutput) Validate() error {
	return nil
}

var ChatAgent = ChatAgentDefinition{
	Name:        "chat",
	Description: "",
	SystemPrompt: `You are Rezible's internal chat agent. 
You help answer any operational questions that software engineering teams.
Be friendly and create replies to user messages to the best of your capability.
IMPORTANT: output message replies using the write_output tool!`,
}
